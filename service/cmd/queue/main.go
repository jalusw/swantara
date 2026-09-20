package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/jobs"
	"github.com/jalusw/swantara/apps/service/internal/kernel/audit"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/logger"
	"github.com/jalusw/swantara/apps/service/internal/mail"
	"github.com/jalusw/swantara/apps/service/internal/organization"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/queue/handlers"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.New("", config.ConfigFilePath())
	if err != nil {
		log.Fatalf("Failed to load configuration %v", err)
	}

	logger.Setup(cfg.ApplicationDebug, cfg.LogLevel, logger.Options{
		File:       cfg.LogFile,
		Format:     cfg.LogFormat,
		MaxSize:    cfg.LogMaxSize,
		MaxAge:     cfg.LogMaxAge,
		MaxBackups: cfg.LogMaxBackups,
		Compress:   cfg.LogCompress,
	})

	queueServer := queue.NewQueueServer(cfg)
	mux := asynq.NewServeMux()

	d, err := db.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to load database configuration %v", err)
	}
	if err := d.Use(audit.NewGormPlugin()); err != nil {
		logger.Fatalf("Failed to register audit plugin %v", err)
	}

	mailer, err := mail.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to initialize mailer %v", err)
	}

	emailVerificationHandler := handlers.NewSendEmailHandler(
		mailer,
		iam.NewUserDAO(d),
		mail.EmailVerificationTemplate,
		cfg.MailUserVerificationRedirection,
		"VerificationURL",
		"Verify your email address",
	)
	mux.HandleFunc(tasks.TypeSendEmailVerification, emailVerificationHandler.Handle)

	passwordResetHandler := handlers.NewSendEmailHandler(
		mailer,
		iam.NewUserDAO(d),
		mail.PasswordResetTemplate,
		cfg.MailPasswordResetRedirection,
		"ResetURL",
		"Reset your password",
	)
	mux.HandleFunc(tasks.TypeSendPasswordReset, passwordResetHandler.Handle)

	reminderHandler := handlers.NewSendReminderHandler(
		mailer,
		accounting.NewReminderActionDAO(d),
		accounting.NewInvoiceDAO(d),
		accounting.NewReminderLevelDAO(d),
		contacts.NewContactDAO(d),
		mail.ReminderReminderTemplate,
		"Payment reminder",
	)
	mux.HandleFunc(tasks.TypeSendReminderEmail, reminderHandler.Handle)

	subscriptionSvc, deferralSvc := buildSubscriptionService(d)
	jobRunService := jobs.NewService(jobs.NewJobRunDAO(d))

	webhookDeliverer := xtradata.NewWebhookDeliverer(
		xtradata.NewWebhookSubscriptionDAO(d),
		xtradata.NewWebhookDeliveryDAO(d),
		&http.Client{Timeout: 10 * time.Second},
	)
	eventService := xtradata.NewEventService(
		xtradata.NewIntegrationEventDAO(d),
		queue.NewQueueClient(cfg),
	).WithDeliver(webhookDeliverer.Deliver)

	subscriptionBillingHandler := handlers.NewSubscriptionBillingHandler(
		subscriptionSvc,
		deferralSvc,
		jobRunService,
		eventService,
	)
	mux.HandleFunc(tasks.TypeSubscriptionBilling, subscriptionBillingHandler.Handle)

	deferralRecognitionHandler := handlers.NewDeferralRecognitionHandler(deferralSvc, jobRunService, eventService)
	mux.HandleFunc(tasks.TypeDeferralRecognition, deferralRecognitionHandler.Handle)

	integrationEventHandler := handlers.NewIntegrationEventDispatchHandler(eventService)
	mux.HandleFunc(tasks.TypeIntegrationEventDispatch, integrationEventHandler.Handle)

	provisioningHandler := buildOrganizationProvisioningHandler(d)
	mux.HandleFunc(tasks.TypeOrganizationProvisioning, provisioningHandler.Handle)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	runErr := make(chan error, 1)
	go func() {
		runErr <- queueServer.Run(mux)
	}()

	select {
	case err := <-runErr:
		logger.Fatalf("Failed to run queue server %v", err)
	case sig := <-stop:
		slog.Info("received signal, shutting down queue worker", "signal", sig.String())
		queueServer.Shutdown()
		if err := <-runErr; err != nil {
			logger.Fatalf("Failed to shutdown queue worker %v", err)
		}
	}

	slog.Info("queue worker stopped")
}

func buildSubscriptionService(d *gorm.DB) (subscription.SubscriptionService, accounting.DeferralService) {
	itemCategoryDAO := dao.NewBase[reference.ItemCategory](d)
	accountDAO := dao.NewBase[reference.Account](d)
	taxDAO := dao.NewBase[reference.Tax](d)
	taxYearDAO := dao.NewBase[reference.TaxYear](d)
	systemConfigDAO := dao.NewBase[reference.SystemConfig](d)
	subscriptionPlanDAO := dao.NewBase[reference.SubscriptionPlan](d)

	productSvc := products.NewProductService(
		products.NewItemDAO(d),
		products.NewItemVariantDAO(d),
		itemCategoryDAO,
		products.NewPriceBookDAO(d),
		products.NewPriceRuleDAO(d),
	)

	journalEntryDAO := accounting.NewJournalEntryDAO(d)
	journalEntryLineDAO := accounting.NewJournalLineDAO(d)
	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(d), taxYearDAO).WithCloseGuard(accounting.NewPeriodCloseDAO(d))
	reversalEngine := accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO).SetPeriods(taxPeriodSvc)
	sequenceSvc := sequence.NewSequenceService(sequence.NewDAO(d))
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine).WithAccounts(accountDAO).WithSequences(sequenceSvc)

	itemVariantDAO := products.NewItemVariantDAO(d)
	itemDAO := products.NewItemDAO(d)
	organizationDAO := dao.NewBase[reference.Organization](d)
	fxRateDAO := dao.NewBase[reference.FxRate](d)
	invoiceSvc := accounting.NewInvoiceService(
		accounting.NewInvoiceDAO(d),
		accounting.NewInvoiceLineDAO(d),
		accounting.NewInvoiceTaxDAO(d),
		postingSvc,
		accountDAO,
		taxDAO,
		sequenceSvc,
		db.NewDBTransactioner(d),
	).WithProducts(accounting.NewProductAdapter(itemVariantDAO, itemDAO), itemCategoryDAO).WithCurrencyResolvers(organizationDAO, reference.NewFxRateSource(fxRateDAO)).WithPaymentTerms(accounting.NewPaymentTermAdapter(reference.NewPaymentTermDAO(d))).WithTaxRules(accounting.NewTaxRuleResolver(accounting.NewTaxRuleDAO(d), accounting.NewTaxRuleTaxMapDAO(d), accounting.NewTaxRuleAccountMapDAO(d), db.NewDBTransactioner(d))).WithInstallments(accounting.NewInvoiceInstallmentDAO(d)).WithSettlements(accounting.NewInvoiceCreditApplicationDAO(d), accounting.NewInvoiceContraSettlementDAO(d)).WithDownPayments(accounting.NewDownPaymentLinkDAO(d)).WithCreditLimits(accounting.NewContactCreditAdapter(contacts.NewCustomerProfileDAO(d)))

	subscriptionConfigSource := subscription.NewSubscriptionConfigSource(systemConfigDAO)
	deferralSvc := accounting.NewDeferralService(
		accounting.NewDeferredScheduleDAO(d),
		accounting.NewDeferredScheduleLineDAO(d),
		postingSvc,
		subscriptionConfigSource,
		db.NewDBTransactioner(d),
	)
	return subscription.NewSubscriptionService(
		subscription.NewSubscriptionDAO(d),
		subscription.NewSubscriptionLineDAO(d),
		subscriptionPlanDAO,
		contacts.NewContactDAO(d),
		products.NewPriceBookDAO(d),
		productSvc,
		invoiceSvc,
		deferralSvc,
		subscriptionConfigSource,
		db.NewDBTransactioner(d),
	), deferralSvc
}

func buildOrganizationProvisioningHandler(d *gorm.DB) handlers.OrganizationProvisioningHandler {
	organizationDAO := dao.NewBase[reference.Organization](d)

	organizationSvc := organization.NewOrganizationServiceWithPaymentTerms(
		organizationDAO,
		dao.NewBase[reference.Currency](d),
		iam.NewMemberService(iam.NewMemberDAO(d), iam.NewMemberRoleDAO(d), iam.NewPermissionDAO(d)),
		organization.NewPaymentTermSeeder(reference.NewPaymentTermDAO(d)),
	)
	organizationSvc.SetModuleStore(dao.NewBase[reference.OrganizationModule](d))
	return handlers.NewOrganizationProvisioningHandler(organizationSvc)
}
