package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	accountingHandler "github.com/jalusw/swantara/apps/service/internal/accounting/handler"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	assetHandler "github.com/jalusw/swantara/apps/service/internal/asset/handler"
	"github.com/jalusw/swantara/apps/service/internal/commission"
	commissionHandler "github.com/jalusw/swantara/apps/service/internal/commission/handler"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	contactsHandler "github.com/jalusw/swantara/apps/service/internal/contacts/handler"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	crmHandler "github.com/jalusw/swantara/apps/service/internal/crm/handler"
	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	crosscuttingHandler "github.com/jalusw/swantara/apps/service/internal/crosscutting/handler"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/expense"
	expenseHandler "github.com/jalusw/swantara/apps/service/internal/expense/handler"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	giftcardHandler "github.com/jalusw/swantara/apps/service/internal/giftcard/handler"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/iam/handler"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	interorganizationHandler "github.com/jalusw/swantara/apps/service/internal/interorganization/handler"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	inventoryHandler "github.com/jalusw/swantara/apps/service/internal/inventory/handler"
	"github.com/jalusw/swantara/apps/service/internal/kernel/audit"
	auditHandler "github.com/jalusw/swantara/apps/service/internal/kernel/audit/handler"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/logger"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	manufacturingHandler "github.com/jalusw/swantara/apps/service/internal/manufacturing/handler"
	"github.com/jalusw/swantara/apps/service/internal/organization"
	organizationHandler "github.com/jalusw/swantara/apps/service/internal/organization/handler"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	payrollHandler "github.com/jalusw/swantara/apps/service/internal/payroll/handler"
	"github.com/jalusw/swantara/apps/service/internal/pos"
	posHandler "github.com/jalusw/swantara/apps/service/internal/pos/handler"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	procurementHandler "github.com/jalusw/swantara/apps/service/internal/procurement/handler"
	"github.com/jalusw/swantara/apps/service/internal/products"
	productsHandler "github.com/jalusw/swantara/apps/service/internal/products/handler"
	"github.com/jalusw/swantara/apps/service/internal/project"
	projectHandler "github.com/jalusw/swantara/apps/service/internal/project/handler"
	"github.com/jalusw/swantara/apps/service/internal/quality"
	qualityHandler "github.com/jalusw/swantara/apps/service/internal/quality/handler"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	referenceHandler "github.com/jalusw/swantara/apps/service/internal/reference/handler"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
	reportingHandler "github.com/jalusw/swantara/apps/service/internal/reporting/handler"
	"github.com/jalusw/swantara/apps/service/internal/returns"
	returnsHandler "github.com/jalusw/swantara/apps/service/internal/returns/handler"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	salesHandler "github.com/jalusw/swantara/apps/service/internal/sales/handler"
	"github.com/jalusw/swantara/apps/service/internal/service"
	serviceHandler "github.com/jalusw/swantara/apps/service/internal/service/handler"
	"github.com/jalusw/swantara/apps/service/internal/storage"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	subscriptionHandler "github.com/jalusw/swantara/apps/service/internal/subscription/handler"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
	xtradataHandler "github.com/jalusw/swantara/apps/service/internal/xtradata/handler"
)

// @title Swantara API Service
// @version 1.0.0
// @description REST API for the Swantara ERP platform.
// @description
// @description Authentication is via Bearer JWT — obtain a token from `POST /api/v1/auth/login` and send `Authorization: Bearer <token>` on subsequent requests.
// @description Most endpoints are multi-tenant and live under `/api/v1/organizations/{organization_id}`; the caller must have an active membership in the organization with the required permission.
// @description
// @description Conventions: paginated lists accept `page`, `size`, `sort`, `filter` and return `{data, meta}`; responses use a shared envelope `{success, message, data}` with a consistent error shape; mutating requests support `Idempotency-Key` for safe retries. See `GET /api/v1/health` for liveness and `/docs/index.html` for interactive documentation.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT access token. Format: "Bearer <token>".
func main() {
	cfg, err := config.New("", config.ConfigFilePath())
	if err != nil {
		logger.Fatalf("Failed to load configuration %v", err)
	}

	logger.Setup(cfg.ApplicationDebug, cfg.LogLevel, logger.Options{
		File:       cfg.LogFile,
		Format:     cfg.LogFormat,
		MaxSize:    cfg.LogMaxSize,
		MaxAge:     cfg.LogMaxAge,
		MaxBackups: cfg.LogMaxBackups,
		Compress:   cfg.LogCompress,
	})

	d, err := db.New(cfg)
	if err != nil {
		logger.Fatalf("Failed to load database %v", err)
	}
	if err := d.Use(audit.NewGormPlugin()); err != nil {
		logger.Fatalf("Failed to register audit plugin %v", err)
	}

	limiterStorage, err := httpx.NewRateLimiterStorage(cfg)
	if err != nil {
		slog.Warn("rate limiter storage unavailable, using in-memory storage", "error", err)
	}

	app := httpx.NewApp(cfg, limiterStorage)
	httpx.RegisterDocs(app, cfg)
	httpx.SetResponseVersion(cfg.ApplicationVersion)
	httpx.SetLinkBaseURL(cfg.ApplicationURL)

	userDAO := iam.NewUserDAO(d)
	userSessionDAO := iam.NewUserSessionDAO(d)
	userEmailVerificationDAO := iam.NewUserEmailVerificationDAO(d)
	userPasswordResetDAO := iam.NewUserPasswordResetDAO(d)
	permissionDAO := iam.NewPermissionDAO(d)
	memberDAO := iam.NewMemberDAO(d)
	memberRoleDAO := iam.NewMemberRoleDAO(d)

	organizationDAO := dao.NewBase[reference.Organization](d)
	currencyDAO := dao.NewBase[reference.Currency](d)
	fxRateDAO := dao.NewBase[reference.FxRate](d)
	unitGroupDAO := dao.NewBase[reference.UnitGroup](d)
	uomDAO := dao.NewBase[reference.Unit](d)
	paymentTermDAO := reference.NewPaymentTermDAO(d)
	accountDAO := dao.NewBase[reference.Account](d)
	journalDAO := dao.NewBase[reference.Journal](d)
	dimensionDAO := dao.NewBase[reference.Dimension](d)
	taxDAO := dao.NewBase[reference.Tax](d)
	taxYearDAO := dao.NewBase[reference.TaxYear](d)
	itemCategoryDAO := dao.NewBase[reference.ItemCategory](d)
	crmStageDAO := dao.NewBase[reference.PipelineStage](d)
	salesTeamDAO := dao.NewBase[reference.SalesGroup](d)
	carrierDAO := dao.NewBase[reference.Carrier](d)
	departmentDAO := dao.NewBase[reference.Department](d)
	jobPositionDAO := dao.NewBase[reference.JobPosition](d)
	workCenterDAO := dao.NewBase[reference.WorkCenter](d)
	leaveTypeDAO := dao.NewBase[reference.LeaveType](d)
	salaryRuleDAO := dao.NewBase[reference.SalaryRule](d)
	assetCategoryDAO := dao.NewBase[reference.AssetCategory](d)
	posConfigDAO := dao.NewBase[reference.POSConfig](d)
	expenseCategoryDAO := dao.NewBase[reference.ExpenseCategory](d)
	subscriptionPlanDAO := dao.NewBase[reference.SubscriptionPlan](d)
	systemConfigDAO := dao.NewBase[reference.SystemConfig](d)

	journalEntryDAO := accounting.NewJournalEntryDAO(d)

	organizationSvc := organization.NewOrganizationServiceWithPaymentTerms(
		organizationDAO,
		currencyDAO,
		iam.NewMemberService(memberDAO, memberRoleDAO, permissionDAO),
		organization.NewPaymentTermSeeder(paymentTermDAO),
	)
	organizationSvc.SetModuleStore(dao.NewBase[reference.OrganizationModule](d))

	fxRateSvc := reference.NewFxRateService(fxRateDAO, currencyDAO)
	uomSvc := reference.NewUnitService(unitGroupDAO, uomDAO)
	paymentTermSvc := reference.NewPaymentTermService(paymentTermDAO)
	dimensionSvc := reference.NewDimensionService(dimensionDAO)
	accountSvc := reference.NewAccountService(accountDAO)
	journalSvc := reference.NewJournalService(journalDAO, accountDAO)
	taxSvc := reference.NewTaxService(taxDAO, accountDAO)
	taxYearSvc := reference.NewTaxYearService(taxYearDAO)

	contactSvc := contacts.NewContactService(
		contacts.NewContactDAO(d),
		contacts.NewContactAddressDAO(d),
		contacts.NewContactBankAccountDAO(d),
		contacts.NewCustomerProfileDAO(d),
		contacts.NewSupplierProfileDAO(d),
	)

	tokenSvc := iam.NewTokenService(cfg)
	passwordSvc := iam.NewPasswordService(cfg)
	queueClient := queue.NewQueueClient(cfg)
	organizationSvc.SetProvisionEnqueuer(queueClient)

	configSvc := xtradata.NewConfigService(xtradata.NewSystemConfigDAO(d))
	idemSvc := xtradata.NewIdempotencyService(xtradata.NewIdempotencyKeyDAO(d))

	userSvc := iam.NewUserService(userDAO, passwordSvc)

	authSvc := iam.NewAuthService(
		userDAO,
		userSvc,
		userSessionDAO,
		passwordSvc,
		tokenSvc,
		userEmailVerificationDAO,
		userPasswordResetDAO,
		queueClient,
		cfg.AuthEmailVerifyTTLDuration,
		cfg.AuthPasswordResetTTLDuration,
	)
	authzSvc := iam.NewAuthzService(permissionDAO)

	localStore := storage.NewLocalStore(cfg.StoragePath)
	avatarSvc := iam.NewAvatarService(userDAO, localStore)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc, authzSvc, iam.NewMemberService(memberDAO, memberRoleDAO, permissionDAO), avatarSvc)

	authn := httpx.NewAuthenticationMiddleware(tokenSvc, userDAO, userSessionDAO)
	authz := httpx.NewAuthorizationMiddleware(permissionDAO)

	guards := httpx.RouteGuards{
		AuthN:       authn.AuthN,
		Guard:       authz.Guard,
		Idempotency: idemSvc,
	}

	api := app.Group("/api/v1")
	orgAPI := api.Group("/organizations/:organization_id")
	authHandler.Register(api, guards)
	userHandler.Register(api, guards)

	fxRateHandler := referenceHandler.NewFxRateHandler(fxRateSvc, reference.NewFxRateSource(fxRateDAO))
	unitGroupHandler := referenceHandler.NewUnitGroupHandler(uomSvc)
	uomHandler := referenceHandler.NewUnitHandler(uomSvc)
	dimensionHandler := referenceHandler.NewDimensionHandler(dimensionSvc)
	paymentTermHandler := referenceHandler.NewPaymentTermHandler(paymentTermSvc)
	accountHandler := referenceHandler.NewAccountHandler(accountSvc)
	journalHandler := referenceHandler.NewJournalHandler(journalSvc)
	taxHandler := referenceHandler.NewTaxHandler(taxSvc)
	taxYearHandler := referenceHandler.NewTaxYearHandler(taxYearSvc)

	currencyHandler := referenceHandler.NewCurrencyHandler(fxRateSvc)

	unitGroupHandler.Register(api, guards)
	uomHandler.Register(api, guards)
	paymentTermHandler.Register(api, guards)
	paymentTermHandler.RegisterOrg(orgAPI, guards)
	currencyHandler.Register(api, guards)

	fxRateHandler.Register(orgAPI, guards)
	dimensionHandler.Register(orgAPI, guards)
	accountHandler.Register(orgAPI, guards)
	journalHandler.Register(orgAPI, guards)
	taxHandler.Register(orgAPI, guards)
	taxYearHandler.Register(orgAPI, guards)

	systemConfigHandler := xtradataHandler.NewSystemConfigHandler(configSvc)
	integrationEventHandler := xtradataHandler.NewIntegrationEventHandler(xtradata.NewEventService(xtradata.NewIntegrationEventDAO(d), nil))
	systemConfigHandler.Register(orgAPI, guards)
	integrationEventHandler.Register(orgAPI, guards)

	auditLogHandler := auditHandler.NewAuditLogHandler(audit.NewLogDAO(d))
	auditLogHandler.Register(orgAPI, guards)

	orgHandler := organizationHandler.NewOrganizationHandler(
		organizationSvc,
	)
	orgHandler.Register(api, guards)

	memberHandler := handler.NewMemberHandler(iam.NewMemberService(memberDAO, memberRoleDAO, permissionDAO))
	memberHandler.Register(orgAPI, guards)

	contactHandler := contactsHandler.NewContactHandler(contactSvc)
	contactRelationHandler := contactsHandler.NewContactRelationHandler(
		contactSvc,
	)
	contactHandler.Register(orgAPI, guards)
	contactRelationHandler.Register(orgAPI, guards)

	itemDAO := products.NewItemDAO(d)
	itemVariantDAO := products.NewItemVariantDAO(d)
	priceBookDAO := products.NewPriceBookDAO(d)
	priceBookRuleDAO := products.NewPriceRuleDAO(d)
	productSvc := products.NewProductService(
		itemDAO,
		itemVariantDAO,
		itemCategoryDAO,
		priceBookDAO,
		priceBookRuleDAO,
	)

	itemCategoryHandler := productsHandler.NewItemCategoryHandler(products.NewItemCategoryService(itemCategoryDAO))
	productHandler := productsHandler.NewProductHandler(productSvc)
	priceBookHandler := productsHandler.NewPriceBookHandler(productSvc)

	carrierHandler := referenceHandler.NewCarrierHandler(reference.NewCarrierService(carrierDAO, itemVariantDAO))
	carrierHandler.Register(api, guards)

	supplierProductDAO := products.NewSupplierProductDAO(d)
	supplierProductSvc := products.NewSupplierProductService(
		itemVariantDAO,
		itemDAO,
		supplierProductDAO,
		contacts.NewContactDAO(d),
		contacts.NewSupplierProfileDAO(d),
	)
	supplierProductHandler := productsHandler.NewSupplierProductHandler(supplierProductSvc)

	recipeDAO := manufacturing.NewRecipeDAO(d)
	recipeLineDAO := manufacturing.NewRecipeLineDAO(d)
	bomSvc := manufacturing.NewRecipeService(itemVariantDAO, recipeDAO, recipeLineDAO)
	bomHandler := manufacturingHandler.NewRecipeHandler(bomSvc)

	itemCategoryHandler.Register(orgAPI, guards)
	supplierProductHandler.Register(orgAPI, guards)

	productHandler.Register(orgAPI, guards)
	priceBookHandler.Register(orgAPI, guards)
	bomHandler.Register(orgAPI, guards)

	journalEntryLineDAO := accounting.NewJournalLineDAO(d)

	taxPeriodDAO := accounting.NewTaxPeriodDAO(d)
	taxPeriodSvc := accounting.NewTaxPeriodService(taxPeriodDAO, taxYearDAO).WithCloseGuard(accounting.NewPeriodCloseDAO(d))
	reversalEngine := accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO).SetPeriods(taxPeriodSvc)
	sequenceSvc := sequence.NewSequenceService(sequence.NewDAO(d))
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine).WithLines(journalEntryLineDAO).WithAccounts(accountDAO).WithSequences(sequenceSvc)
	invoiceDAO := accounting.NewInvoiceDAO(d)
	invoiceLineDAO := accounting.NewInvoiceLineDAO(d)
	invoiceTaxDAO := accounting.NewInvoiceTaxDAO(d)
	paymentDAO := accounting.NewPaymentDAO(d)
	paymentAllocationDAO := accounting.NewPaymentAllocationDAO(d)
	invoiceSvc := accounting.NewInvoiceService(invoiceDAO, invoiceLineDAO, invoiceTaxDAO, postingSvc, accountDAO, taxDAO, sequenceSvc, db.NewDBTransactioner(d)).WithProducts(accounting.NewProductAdapter(itemVariantDAO, itemDAO), itemCategoryDAO).WithCurrencyResolvers(organizationDAO, reference.NewFxRateSource(fxRateDAO)).WithPaymentTerms(accounting.NewPaymentTermAdapter(paymentTermDAO))
	paymentSvc := accounting.NewPaymentService(paymentDAO, invoiceDAO, postingSvc, accountDAO, journalDAO, sequenceSvc, db.NewDBTransactioner(d)).WithFxResolvers(reference.NewFxRateSource(fxRateDAO), accounting.NewFxAccountResolver(systemConfigDAO, accountDAO)).WithOrganizations(organizationDAO).WithAdvanceAccounts(accounting.NewAdvanceAccountResolver(systemConfigDAO, accountDAO)).WithAllocations(paymentAllocationDAO)
	bankStatementDAO := accounting.NewBankStatementDAO(d)
	bankStatementLineDAO := accounting.NewBankStatementLineDAO(d)
	bankStatementSvc := accounting.NewBankStatementService(bankStatementDAO, bankStatementLineDAO, paymentDAO, journalDAO, db.NewDBTransactioner(d)).WithPosting(postingSvc, accounting.NewBankChargeResolver(systemConfigDAO, accountDAO))
	partialReconcileDAO := accounting.NewAccountPartialReconcileDAO(d)
	fullReconcileDAO := accounting.NewAccountFullReconcileDAO(d)
	reconcileSvc := accounting.NewReconcileService(journalEntryLineDAO, partialReconcileDAO, fullReconcileDAO, journalEntryDAO, db.NewDBTransactioner(d))
	reminderLevelDAO := accounting.NewReminderLevelDAO(d)
	reminderActionDAO := accounting.NewReminderActionDAO(d)
	reminderSvc := accounting.NewReminderService(invoiceDAO, reminderLevelDAO, reminderActionDAO, db.NewDBTransactioner(d)).WithSender(queueClient)
	budgetDAO := accounting.NewBudgetDAO(d)
	budgetLineDAO := accounting.NewBudgetLineDAO(d)
	budgetQueryDAO := accounting.NewBudgetQueryDAO(d)
	budgetSvc := accounting.NewBudgetService(budgetDAO, budgetLineDAO, budgetQueryDAO, db.NewDBTransactioner(d))
	taxRuleDAO := accounting.NewTaxRuleDAO(d)
	taxRuleTaxMapDAO := accounting.NewTaxRuleTaxMapDAO(d)
	taxRuleAccountMapDAO := accounting.NewTaxRuleAccountMapDAO(d)
	taxRuleResolver := accounting.NewTaxRuleResolver(taxRuleDAO, taxRuleTaxMapDAO, taxRuleAccountMapDAO, db.NewDBTransactioner(d))
	invoiceSvc = invoiceSvc.WithTaxRules(taxRuleResolver).WithInstallments(accounting.NewInvoiceInstallmentDAO(d)).WithSettlements(accounting.NewInvoiceCreditApplicationDAO(d), accounting.NewInvoiceContraSettlementDAO(d)).WithDownPayments(accounting.NewDownPaymentLinkDAO(d)).WithCreditLimits(accounting.NewContactCreditAdapter(contacts.NewCustomerProfileDAO(d)))
	withholdingTaxDAO := accounting.NewWithholdingTaxDAO(d)
	withholdingSvc := accounting.NewWithholdingService(withholdingTaxDAO, postingSvc, db.NewDBTransactioner(d))
	taxReturnDAO := accounting.NewTaxReturnDAO(d)
	taxReturnSvc := accounting.NewTaxReturnService(taxReturnDAO, taxPeriodDAO, invoiceTaxDAO, db.NewDBTransactioner(d)).WithSettlement(postingSvc, journalDAO, accountDAO)

	trialBalanceSvc := accounting.NewTrialBalanceService(accounting.NewTrialBalanceDAO(d), taxPeriodDAO).WithAccounts(accountDAO)
	trialBalanceHandler := accountingHandler.NewTrialBalanceHandler(trialBalanceSvc)
	cashFlowSvc := accounting.NewCashFlowService(accounting.NewCashFlowDAO(d))
	cashFlowHandler := accountingHandler.NewCashFlowHandler(cashFlowSvc)
	equitySvc := accounting.NewEquityService(accounting.NewEquityDAO(d), taxPeriodDAO)
	equityHandler := accountingHandler.NewEquityHandler(equitySvc)
	integritySvc := accounting.NewReportIntegrityService(accounting.NewTrialBalanceDAO(d), accounting.NewCashFlowDAO(d), accounting.NewEquityDAO(d), taxPeriodDAO, accountDAO)
	integrityHandler := accountingHandler.NewIntegrityHandler(integritySvc)
	generalLedgerHandler := accountingHandler.NewGeneralLedgerHandler(accounting.NewGeneralLedgerService(accounting.NewGeneralLedgerDAO(d)))

	accountMoveHandler := accountingHandler.NewJournalEntryHandler(postingSvc)
	invoiceHandler := accountingHandler.NewInvoiceHandler(invoiceSvc)
	paymentHandler := accountingHandler.NewPaymentHandler(paymentSvc)
	pdcHandler := accountingHandler.NewPdcHandler(accounting.NewPdcService(accounting.NewPdcInstrumentDAO(d), paymentSvc, db.NewDBTransactioner(d)))
	taxPeriodHandler := accountingHandler.NewTaxPeriodHandler(taxPeriodSvc)
	bankStatementHandler := accountingHandler.NewBankStatementHandler(bankStatementSvc)
	reconcileHandler := accountingHandler.NewReconcileHandler(reconcileSvc)
	reminderHandler := accountingHandler.NewReminderHandler(reminderSvc)
	budgetHandler := accountingHandler.NewBudgetHandler(budgetSvc)
	taxRuleHandler := accountingHandler.NewTaxRuleHandler(taxRuleResolver)
	withholdingHandler := accountingHandler.NewWithholdingTaxHandler(withholdingSvc)
	taxReturnHandler := accountingHandler.NewTaxReturnHandler(taxReturnSvc)
	reconcileRuleDAO := accounting.NewReconcileRuleDAO(d)
	reconcileRuleMatchDAO := accounting.NewReconcileRuleMatchDAO(d)
	reconcileRuleEngine := accounting.NewReconcileRuleEngine(reconcileRuleDAO, reconcileRuleMatchDAO, journalEntryLineDAO, partialReconcileDAO, d)
	reconcileRuleHandler := accountingHandler.NewReconcileRuleHandler(reconcileRuleEngine)

	accountMoveHandler.Register(orgAPI, guards)
	invoiceHandler.Register(orgAPI, guards)
	invoiceHandler.RegisterSupplierBills(orgAPI, guards)
	paymentHandler.Register(orgAPI, guards)
	pdcHandler.Register(orgAPI, guards)
	taxPeriodHandler.Register(orgAPI, guards)
	bankStatementHandler.Register(orgAPI, guards)
	reconcileHandler.Register(orgAPI, guards)
	reminderHandler.Register(orgAPI, guards)
	budgetHandler.Register(orgAPI, guards)
	taxRuleHandler.Register(orgAPI, guards)
	withholdingHandler.Register(orgAPI, guards)
	taxReturnHandler.Register(orgAPI, guards)
	reconcileRuleHandler.Register(orgAPI, guards)
	trialBalanceHandler.Register(orgAPI, guards)
	cashFlowHandler.Register(orgAPI, guards)
	equityHandler.Register(orgAPI, guards)
	integrityHandler.Register(orgAPI, guards)
	generalLedgerHandler.Register(orgAPI, guards)

	warehouseDAO := inventory.NewWarehouseDAO(d)
	stockLocationDAO := inventory.NewStockLocationDAO(d)
	stockMovementDAO := inventory.NewStockMovementDAO(d)
	stockShipmentDAO := inventory.NewShipmentDAO(d)
	stockQuantDAO := inventory.NewStockBalanceDAO(d)
	stockLotDAO := inventory.NewBatchDAO(d)
	stockReservationDAO := inventory.NewStockHoldDAO(d)
	stockCostLayerDAO := inventory.NewCostLayerDAO(d)
	reorderRuleDAO := inventory.NewReorderRuleDAO(d)
	inventoryCountDAO := inventory.NewStockCountDAO(d)
	inventoryCountLineDAO := inventory.NewStockCountLineDAO(d)
	transferOrderDAO := inventory.NewWarehouseTransferDAO(d)

	productResolver := inventory.NewItemResolver(itemVariantDAO, itemDAO, itemCategoryDAO)

	warehouseSvc := inventory.NewWarehouseService(warehouseDAO, stockLocationDAO)
	ledgerSvc := inventory.NewLedgerService(stockMovementDAO, stockQuantDAO, stockLocationDAO, db.NewDBTransactioner(d))
	valuationSvc := inventory.NewValuationService(stockMovementDAO, stockCostLayerDAO, stockLocationDAO, productResolver, postingSvc, db.NewDBTransactioner(d))
	reservationSvc := inventory.NewHoldService(stockReservationDAO, stockQuantDAO)
	lotSvc := inventory.NewBatchService(stockLotDAO, productResolver)
	reorderSvc := inventory.NewReorderService(reorderRuleDAO, ledgerSvc, productResolver)
	countSvc := inventory.NewStockCountService(inventoryCountDAO, inventoryCountLineDAO, stockQuantDAO, stockCostLayerDAO, ledgerSvc, productResolver, postingSvc, db.NewDBTransactioner(d))
	transferSvc := inventory.NewTransferService(transferOrderDAO, stockMovementDAO, stockLocationDAO, warehouseDAO, stockCostLayerDAO, ledgerSvc, productResolver, postingSvc, db.NewDBTransactioner(d))

	warehouseHandler := inventoryHandler.NewWarehouseHandler(warehouseSvc)
	stockHandler := inventoryHandler.NewStockHandler(inventory.NewStockService(stockMovementDAO, stockShipmentDAO, stockQuantDAO), ledgerSvc, valuationSvc)
	reservationHandler := inventoryHandler.NewHoldHandler(reservationSvc)
	lotHandler := inventoryHandler.NewBatchHandler(lotSvc)
	reorderHandler := inventoryHandler.NewReorderHandler(reorderSvc)
	countHandler := inventoryHandler.NewStockCountHandler(countSvc)
	transferHandler := inventoryHandler.NewTransferHandler(transferSvc)

	moDAO := manufacturing.NewProductionOrderDAO(d)
	consumedMaterialDAO := manufacturing.NewConsumedMaterialDAO(d)
	moSvc := manufacturing.NewProductionOrderService(moDAO, consumedMaterialDAO, recipeDAO, bomSvc, itemVariantDAO, stockLocationDAO, reservationSvc, sequenceSvc)
	moHandler := manufacturingHandler.NewProductionOrderHandler(moSvc)
	moHandler.Register(orgAPI, guards)

	workOrderDAO := manufacturing.NewShopTaskDAO(d)
	routingOperationDAO := manufacturing.NewProductionStepDAO(d)
	productionSvc := manufacturing.NewProductionService(moDAO, consumedMaterialDAO, workOrderDAO, routingOperationDAO, workCenterDAO, stockLocationDAO, ledgerSvc, valuationSvc, productResolver, postingSvc, journalEntryLineDAO, sequenceSvc)
	productionHandler := manufacturingHandler.NewProductionHandler(productionSvc)
	productionHandler.Register(orgAPI, guards)

	reservationHandler.Register(orgAPI, guards)
	lotHandler.Register(orgAPI, guards)
	reorderHandler.Register(orgAPI, guards)

	warehouseHandler.Register(orgAPI, guards)
	stockHandler.Register(orgAPI, guards)
	countHandler.Register(orgAPI, guards)
	transferHandler.Register(orgAPI, guards)

	crmLeadDAO := crm.NewProspectDAO(d)
	crmActivityDAO := crm.NewProspectActivityDAO(d)
	leadSvc := crm.NewProspectService(crmLeadDAO, crmStageDAO, salesTeamDAO, contacts.NewContactDAO(d))
	activitySvc := crm.NewProspectActivityService(crmActivityDAO, crmLeadDAO, contacts.NewContactDAO(d))
	pipelineSvc := crm.NewPipelineService(crmLeadDAO, crmStageDAO)

	leadHandler := crmHandler.NewProspectHandler(crm.NewPipelineStageService(crmStageDAO), leadSvc)
	opportunityHandler := crmHandler.NewOpportunityHandler(crm.NewPipelineStageService(crmStageDAO), leadSvc)
	activityHandler := crmHandler.NewProspectActivityHandler(activitySvc)
	pipelineHandler := crmHandler.NewPipelineHandler(pipelineSvc)
	stageHandler := crmHandler.NewPipelineStageHandler(crm.NewPipelineStageService(crmStageDAO))
	teamHandler := crmHandler.NewSalesGroupHandler(crm.NewSalesGroupService(salesTeamDAO))

	activityHandler.Register(orgAPI, guards)
	stageHandler.Register(orgAPI, guards)

	leadHandler.Register(orgAPI, guards)
	opportunityHandler.Register(orgAPI, guards)
	pipelineHandler.Register(orgAPI, guards)
	teamHandler.Register(orgAPI, guards)

	saleOrderDAO := sales.NewSaleOrderDAO(d)
	saleOrderLineDAO := sales.NewSaleOrderLineDAO(d)
	saleOrderSvc := sales.NewSaleOrderService(
		saleOrderDAO,
		saleOrderLineDAO,
		sequenceSvc,
		productSvc,
		priceBookDAO,
		taxDAO,
		contacts.NewContactDAO(d),
		crmLeadDAO,
		crmStageDAO,
		warehouseDAO,
		stockLocationDAO,
		stockQuantDAO,
		stockShipmentDAO,
		stockMovementDAO,
		stockReservationDAO,
		reservationSvc,
		valuationSvc,
		invoiceSvc,
		invoiceDAO,
		paymentSvc,
	)
	saleOrderHandler := salesHandler.NewSaleOrderHandler(saleOrderSvc)

	saleOrderHandler.Register(orgAPI, guards)

	posSessionDAO := pos.NewPOSSessionDAO(d)
	posOrderDAO := pos.NewPOSOrderDAO(d)
	posOrderLineDAO := pos.NewPOSOrderLineDAO(d)
	posPaymentDAO := pos.NewPOSPaymentDAO(d)
	posPaymentAccountDAO := dao.NewBase[reference.POSPaymentAccount](d)
	posSvc := pos.NewPOSService(
		posConfigDAO,
		posSessionDAO,
		posOrderDAO,
		posOrderLineDAO,
		posPaymentDAO,
		productSvc,
		taxDAO,
		journalDAO,
		accountDAO,
		posPaymentAccountDAO,
		stockLocationDAO,
		stockQuantDAO,
		stockMovementDAO,
		stockCostLayerDAO,
		stockShipmentDAO,
		valuationSvc,
		invoiceSvc,
		postingSvc,
		memberDAO,
		sequenceSvc,
		db.NewDBTransactioner(d),
	)
	posConfigHandler := posHandler.NewPOSConfigHandler(pos.NewPOSConfigService(posConfigDAO))
	posSessionHandler := posHandler.NewPOSSessionHandler(posSvc)
	posOrderHandler := posHandler.NewPOSOrderHandler(posSvc)
	posPaymentAccountSvc := pos.NewPaymentAccountService(posPaymentAccountDAO)
	posPaymentAccountHandler := posHandler.NewPaymentAccountHandler(posPaymentAccountSvc)

	posConfigHandler.Register(orgAPI, guards)
	posSessionHandler.Register(orgAPI, guards)
	posOrderHandler.Register(orgAPI, guards)
	posPaymentAccountHandler.Register(orgAPI, guards)

	approvalRequestDAO := crosscutting.NewApprovalRequestDAO(d)
	approvalStepDAO := crosscutting.NewApprovalStepDAO(d)
	approvalSvc := crosscutting.NewApprovalService(approvalRequestDAO, approvalStepDAO, db.NewDBTransactioner(d))
	approvalHandler := crosscuttingHandler.NewApprovalRequestHandler(approvalSvc)
	approvalHandler.Register(orgAPI, guards)

	attachmentDAO := crosscutting.NewAttachmentDAO(d)
	messageDAO := crosscutting.NewMessageDAO(d)
	attachmentSvc := crosscutting.NewAttachmentService(attachmentDAO, localStore)
	messageSvc := crosscutting.NewMessageService(messageDAO)
	attachmentHandler := crosscuttingHandler.NewAttachmentHandler(attachmentSvc)
	messageHandler := crosscuttingHandler.NewMessageHandler(messageSvc)
	attachmentHandler.Register(orgAPI, guards)
	messageHandler.Register(orgAPI, guards)

	qualityPointDAO := quality.NewQualityPointDAO(d)
	qualityCheckDAO := quality.NewQualityCheckDAO(d)
	qualityAlertDAO := quality.NewQualityAlertDAO(d)
	qualitySvc := quality.NewQualityCheckService(qualityPointDAO, qualityCheckDAO, qualityAlertDAO)
	scrapRouter := inventory.NewScrapRouter(stockShipmentDAO, stockMovementDAO, stockLocationDAO, valuationSvc)
	qualitySvc.SetScrapRouter(scrapRouter)

	purchaseConfigSource := procurement.NewSystemConfigSource(systemConfigDAO)
	purchaseRequestDAO := procurement.NewPurchaseRequestDAO(d)
	purchaseRequestLineDAO := procurement.NewPurchaseRequestLineDAO(d)
	contactDAO := contacts.NewContactDAO(d)
	purchaseRequestSvc := procurement.NewPurchaseRequestService(
		purchaseRequestDAO,
		purchaseRequestLineDAO,
		sequenceSvc,
		contactDAO,
	)
	purchaseOrderDAO := procurement.NewPurchaseOrderDAO(d)
	purchaseOrderLineDAO := procurement.NewPurchaseOrderLineDAO(d)
	purchaseRFQDAO := procurement.NewSupplierQuoteRequestDAO(d)
	purchaseRFQLineDAO := procurement.NewSupplierQuoteRequestLineDAO(d)
	purchaseRFQQuoteDAO := procurement.NewSupplierQuoteDAO(d)
	purchaseRFQQuoteLineDAO := procurement.NewSupplierQuoteLineDAO(d)
	currencyRateDAO := procurement.NewCurrencyRateDAO(d)
	currencyConverter := procurement.NewCurrencyConverter(currencyRateDAO)
	supplyAgreementDAO := procurement.NewSupplyAgreementDAO(d)
	supplyAgreementLineDAO := procurement.NewSupplyAgreementLineDAO(d)
	purchaseCreditMemoDAO := procurement.NewPurchaseCreditMemoDAO(d)
	purchaseDebitMemoDAO := procurement.NewPurchaseDebitMemoDAO(d)
	purchasePaymentBatchDAO := procurement.NewPaymentBatchDAO(d)
	purchasePaymentBatchLineDAO := procurement.NewPaymentBatchLineDAO(d)
	purchaseOrderSvc := procurement.NewPurchaseOrderService(
		purchaseOrderDAO,
		purchaseOrderLineDAO,
		purchaseRequestDAO,
		purchaseRequestLineDAO,
		supplyAgreementDAO,
		supplyAgreementLineDAO,
		purchaseCreditMemoDAO,
		purchaseDebitMemoDAO,
		purchasePaymentBatchDAO,
		purchasePaymentBatchLineDAO,
		sequenceSvc,
		purchaseConfigSource,
		supplierProductSvc,
		contactDAO,
		contacts.NewSupplierProfileDAO(d),
		warehouseDAO,
		stockLocationDAO,
		stockShipmentDAO,
		stockMovementDAO,
		productResolver,
		productSvc,
		taxDAO,
		valuationSvc,
		approvalSvc,
		invoiceSvc,
		invoiceDAO,
		paymentSvc,
		qualitySvc,
		currencyConverter,
	)
	purchaseRequestHandler := procurementHandler.NewPurchaseRequestHandler(purchaseRequestSvc)
	purchaseOrderHandler := procurementHandler.NewPurchaseOrderHandler(purchaseOrderSvc)
	purchaseRFQSvc := procurement.NewSupplierQuoteRequestService(
		purchaseRFQDAO,
		purchaseRFQLineDAO,
		purchaseRFQQuoteDAO,
		purchaseRFQQuoteLineDAO,
		purchaseRequestDAO,
		purchaseRequestLineDAO,
		sequenceSvc,
		contactDAO,
		purchaseOrderSvc,
	)
	purchaseRFQHandler := procurementHandler.NewSupplierQuoteRequestHandler(purchaseRFQSvc)
	supplyAgreementSvc := procurement.NewSupplyAgreementService(supplyAgreementDAO, supplyAgreementLineDAO, sequenceSvc, contactDAO)
	supplyAgreementHandler := procurementHandler.NewSupplyAgreementHandler(supplyAgreementSvc, purchaseOrderSvc)
	currencyRateHandler := procurementHandler.NewCurrencyRateHandler(procurement.NewCurrencyRateService(currencyRateDAO))
	vendorScorecardDAO := procurement.NewSupplierScorecardDAO(d)
	vendorScorecardSvc := procurement.NewSupplierScorecardService(vendorScorecardDAO)
	vendorScorecardHandler := procurementHandler.NewSupplierScorecardHandler(vendorScorecardSvc)
	costCenterDAO := procurement.NewCostCenterDAO(d)
	costCenterHandler := procurementHandler.NewCostCenterHandler(procurement.NewCostCenterService(costCenterDAO))
	purchasePaymentBatchHandler := procurementHandler.NewPaymentBatchHandler(purchaseOrderSvc)

	purchaseRequestHandler.Register(orgAPI, guards)
	purchaseOrderHandler.Register(orgAPI, guards)
	purchaseRFQHandler.Register(orgAPI, guards)
	supplyAgreementHandler.Register(orgAPI, guards)
	currencyRateHandler.Register(orgAPI, guards)
	vendorScorecardHandler.Register(orgAPI, guards)
	costCenterHandler.Register(orgAPI, guards)
	purchasePaymentBatchHandler.Register(orgAPI, guards)

	mrpRunDAO := manufacturing.NewPlanningRunDAO(d)
	mrpDemandDAO := manufacturing.NewPlanningNeedDAO(d)
	mrpPlannedOrderDAO := manufacturing.NewPlannedSupplyDAO(d)
	demandForecastDAO := manufacturing.NewDemandPlanDAO(d)
	mrpSvc := manufacturing.NewPlanningService(mrpRunDAO, mrpDemandDAO, mrpPlannedOrderDAO, demandForecastDAO, saleOrderDAO, saleOrderLineDAO, purchaseOrderDAO, purchaseOrderLineDAO, moDAO, reorderSvc, ledgerSvc, recipeDAO, recipeLineDAO, bomSvc, itemVariantDAO, itemDAO, stockLocationDAO, purchaseOrderSvc, moSvc, transferSvc)
	mrpHandler := manufacturingHandler.NewPlanningHandler(mrpSvc)
	mrpHandler.Register(orgAPI, guards)

	subcontractOrderDAO := manufacturing.NewOutsideProcessingOrderDAO(d)
	subcontractSvc := manufacturing.NewOutsideProcessingService(subcontractOrderDAO, moDAO, consumedMaterialDAO, recipeDAO, purchaseOrderSvc, purchaseOrderDAO, purchaseOrderLineDAO, stockLocationDAO, ledgerSvc, valuationSvc, productResolver, postingSvc)
	subcontractHandler := manufacturingHandler.NewOutsideProcessingHandler(subcontractSvc)
	subcontractHandler.Register(orgAPI, guards)

	rmaDAO := returns.NewRMADAO(d)
	rmaLineDAO := returns.NewRMALineDAO(d)
	rmaSvc := returns.NewRMAService(
		rmaDAO,
		rmaLineDAO,
		returns.NewOriginOrderLookup(saleOrderDAO, purchaseOrderDAO),
		stockMovementDAO,
		stockLocationDAO,
		stockCostLayerDAO,
		valuationSvc,
		invoiceDAO,
		invoiceSvc,
		saleOrderSvc,
		purchaseOrderSvc,
		sequenceSvc,
		db.NewDBTransactioner(d),
	)
	rmaHandler := returnsHandler.NewRMAHandler(rmaSvc)

	rmaHandler.Register(orgAPI, guards)

	qualityPointHandler := qualityHandler.NewQualityPointHandler(quality.NewQualityPointService(qualityPointDAO))
	qualityCheckHandler := qualityHandler.NewQualityCheckHandler(qualitySvc)
	qualityAlertHandler := qualityHandler.NewQualityAlertHandler(qualitySvc)

	qualityPointHandler.Register(orgAPI, guards)
	qualityCheckHandler.Register(orgAPI, guards)
	qualityAlertHandler.Register(orgAPI, guards)

	employeeDAO := payroll.NewEmployeeDAO(d)
	contractDAO := payroll.NewEmploymentContractDAO(d)
	leaveDAO := payroll.NewLeaveRequestDAO(d)
	attendanceDAO := payroll.NewAttendanceDAO(d)
	timesheetDAO := payroll.NewTimesheetDAO(d)
	shiftDAO := payroll.NewShiftDAO(d)
	shiftAssignmentDAO := payroll.NewShiftAssignmentDAO(d)
	hrSvc := payroll.NewHRService(
		employeeDAO,
		contractDAO,
		leaveDAO,
		leaveTypeDAO,
		departmentDAO,
		jobPositionDAO,
		organizationDAO,
		contacts.NewContactDAO(d),
		dimensionDAO,
		userDAO,
		attendanceDAO,
		timesheetDAO,
		shiftDAO,
		shiftAssignmentDAO,
		db.NewDBTransactioner(d),
	)
	payrollRunDAO := payroll.NewPayrollRunDAO(d)
	payslipDAO := payroll.NewPayslipDAO(d)
	payslipLineDAO := payroll.NewPayslipLineDAO(d)
	salaryRuleSvc := payroll.NewSalaryRuleService(salaryRuleDAO, accountDAO)
	payrollSvc := payroll.NewPayrollService(
		payrollRunDAO,
		payslipDAO,
		payslipLineDAO,
		employeeDAO,
		contractDAO,
		attendanceDAO,
		salaryRuleDAO,
		journalDAO,
		postingSvc,
		sequenceSvc,
		db.NewDBTransactioner(d),
	)

	departmentHandler := payrollHandler.NewDepartmentHandler(hrSvc)
	jobPositionHandler := payrollHandler.NewJobPositionHandler(hrSvc)
	leaveTypeHandler := payrollHandler.NewLeaveTypeHandler(hrSvc)
	employeeHandler := payrollHandler.NewEmployeeHandler(hrSvc)
	contractHandler := payrollHandler.NewContractHandler(hrSvc)
	leaveRequestHandler := payrollHandler.NewLeaveRequestHandler(hrSvc)
	attendanceHandler := payrollHandler.NewAttendanceHandler(hrSvc)
	timesheetHandler := payrollHandler.NewTimesheetHandler(hrSvc)
	salaryRuleHandler := payrollHandler.NewSalaryRuleHandler(salaryRuleSvc)
	payrollRunHandler := payrollHandler.NewPayrollRunHandler(payrollSvc)
	payslipHandler := payrollHandler.NewPayslipHandler(payrollSvc)

	jobPositionHandler.Register(orgAPI, guards)
	leaveTypeHandler.Register(orgAPI, guards)
	contractHandler.Register(orgAPI, guards)
	salaryRuleHandler.Register(orgAPI, guards)

	departmentHandler.Register(orgAPI, guards)
	employeeHandler.Register(orgAPI, guards)
	leaveRequestHandler.Register(orgAPI, guards)
	attendanceHandler.Register(orgAPI, guards)
	timesheetHandler.Register(orgAPI, guards)
	payrollRunHandler.Register(orgAPI, guards)
	payslipHandler.Register(orgAPI, guards)

	projectDAO := project.NewProjectDAO(d)
	projectTaskDAO := project.NewProjectTaskDAO(d)
	projectMilestoneDAO := project.NewProjectMilestoneDAO(d)
	projectInvoiceLineDAO := project.NewProjectInvoiceLineDAO(d)
	projectSvc := project.NewProjectService(
		projectDAO,
		projectTaskDAO,
		projectMilestoneDAO,
		projectInvoiceLineDAO,
		timesheetDAO,
		contractDAO,
		contactDAO,
		accountDAO,
		dimensionDAO,
		invoiceSvc,
		invoiceLineDAO,
		db.NewDBTransactioner(d),
	)
	projectHandler := projectHandler.NewProjectHandler(projectSvc)

	projectHandler.Register(orgAPI, guards)

	expenseConfigSource := expense.NewExpenseConfigSource(systemConfigDAO)
	expenseReportDAO := expense.NewExpenseReportDAO(d)
	expenseLineDAO := expense.NewExpenseLineDAO(d)
	expenseSvc := expense.NewExpenseService(
		expenseReportDAO,
		expenseLineDAO,
		expenseCategoryDAO,
		taxDAO,
		postingSvc,
		expenseConfigSource,
		db.NewDBTransactioner(d),
	).SetBilling(invoiceSvc, productSvc, projectDAO)
	expenseCategoryHandler := expenseHandler.NewExpenseCategoryHandler(expense.NewExpenseCategoryService(expenseCategoryDAO))
	expenseReportHandler := expenseHandler.NewExpenseReportHandler(expenseSvc)

	expenseCategoryHandler.Register(orgAPI, guards)
	expenseReportHandler.Register(orgAPI, guards)

	inboundCostConfigSource := inventory.NewInboundCostConfigSource(systemConfigDAO)
	inboundCostDAO := inventory.NewInboundCostDAO(d)
	inboundCostLineDAO := inventory.NewInboundCostLineDAO(d)
	inboundCostAdjustmentDAO := inventory.NewInboundCostAdjustmentDAO(d)
	inboundCostSvc := inventory.NewInboundCostService(
		inboundCostDAO,
		inboundCostLineDAO,
		inboundCostAdjustmentDAO,
		stockMovementDAO,
		stockCostLayerDAO,
		productResolver,
		postingSvc,
		inboundCostConfigSource,
		accounting.NewInvoiceLineDAO(d),
		db.NewDBTransactioner(d),
	)
	inboundCostHandler := inventoryHandler.NewInboundCostHandler(inboundCostSvc)

	inboundCostHandler.Register(orgAPI, guards)

	assetDAO := asset.NewFixedAssetDAO(d)
	assetDepreciationLineDAO := asset.NewAssetDepreciationLineDAO(d)
	assetSvc := asset.NewAssetService(
		assetDAO,
		assetDepreciationLineDAO,
		assetCategoryDAO,
		invoiceLineDAO,
		invoiceDAO,
		postingSvc,
		db.NewDBTransactioner(d),
	)
	assetHandler := assetHandler.NewAssetCategoryHandler(assetCategoryDAO, assetSvc)

	assetHandler.Register(orgAPI, guards)

	subscriptionConfigSource := subscription.NewSubscriptionConfigSource(systemConfigDAO)
	deferralSvc := accounting.NewDeferralService(
		accounting.NewDeferredScheduleDAO(d),
		accounting.NewDeferredScheduleLineDAO(d),
		postingSvc,
		subscriptionConfigSource,
		db.NewDBTransactioner(d),
	)
	paymentBatchDAO := accounting.NewPaymentBatchDAO(d)
	paymentBatchLineDAO := accounting.NewPaymentBatchLineDAO(d)
	paymentBatchSvc := accounting.NewPaymentBatchService(paymentBatchDAO, paymentBatchLineDAO, paymentDAO, subscriptionConfigSource, db.NewDBTransactioner(d))
	paymentBatchHandler := accountingHandler.NewPaymentBatchHandler(paymentBatchSvc)
	subscriptionSvc := subscription.NewSubscriptionService(
		subscription.NewSubscriptionDAO(d),
		subscription.NewSubscriptionLineDAO(d),
		subscriptionPlanDAO,
		contacts.NewContactDAO(d),
		priceBookDAO,
		productSvc,
		invoiceSvc,
		deferralSvc,
		subscriptionConfigSource,
		db.NewDBTransactioner(d),
	)
	subscriptionPlanHandler := subscriptionHandler.NewSubscriptionPlanHandler(subscription.NewSubscriptionPlanService(subscriptionPlanDAO))
	subscriptionHandler := subscriptionHandler.NewSubscriptionHandler(subscriptionSvc)
	deferralHandler := accountingHandler.NewDeferralHandler(deferralSvc)

	subscriptionPlanHandler.Register(orgAPI, guards)
	subscriptionHandler.Register(orgAPI, guards)
	deferralHandler.Register(orgAPI, guards)
	paymentBatchHandler.Register(orgAPI, guards)

	commissionSvc := commission.NewCommissionService(
		commission.NewCommissionPlanDAO(d),
		commission.NewCommissionRuleDAO(d),
		commission.NewCommissionAssignmentDAO(d),
		commission.NewCommissionEntryDAO(d),
		postingSvc,
		accounting.NewInvoiceDAO(d),
		accounting.NewInvoiceLineDAO(d),
		accounting.NewPaymentAllocationDAO(d),
		productResolver,
		db.NewDBTransactioner(d),
	)
	commissionHandler := commissionHandler.NewCommissionHandler(
		commissionSvc,
	)
	commissionHandler.Register(orgAPI, guards)

	giftCardSvc := giftcard.NewGiftCardService(
		giftcard.NewGiftCardDAO(d),
		giftcard.NewGiftCardTransactionDAO(d),
		postingSvc,
		db.NewDBTransactioner(d),
	)
	giftCardHandler := giftcardHandler.NewGiftCardHandler(giftCardSvc)
	giftCardHandler.Register(orgAPI, guards)

	couponSvc := giftcard.NewCouponService(giftcard.NewCouponDAO(d), db.NewDBTransactioner(d))
	couponHandler := giftcardHandler.NewCouponHandler(couponSvc)
	couponHandler.Register(orgAPI, guards)

	serviceSvc := service.NewServiceService(
		service.NewEquipmentDAO(d),
		service.NewServiceContractDAO(d),
		service.NewServiceOrderDAO(d),
		service.NewServiceOrderLineDAO(d),
		postingSvc,
		invoiceSvc,
		db.NewDBTransactioner(d),
	)
	maintenanceSvc := service.NewMaintenanceService(
		service.NewMaintenancePlanDAO(d),
		service.NewEquipmentDAO(d),
		service.NewServiceOrderDAO(d),
		db.NewDBTransactioner(d),
	)
	serviceHandler := serviceHandler.NewServiceHandler(
		serviceSvc,
		maintenanceSvc,
	)
	serviceHandler.Register(orgAPI, guards)

	dropshipSvc := interorganization.NewDropShipService(
		interorganization.NewDropshipLinkDAO(d),
		purchaseOrderSvc,
		purchaseOrderDAO,
		purchaseOrderLineDAO,
		saleOrderDAO,
		saleOrderLineDAO,
		stockMovementDAO,
		stockLocationDAO,
		productResolver,
		postingSvc,
		db.NewDBTransactioner(d),
	)
	dropshipHandler := interorganizationHandler.NewDropShipHandler(dropshipSvc)

	interorgSvc := interorganization.NewInterorganizationService(
		interorganization.NewInterorganizationRuleDAO(d),
		interorganization.NewInterorganizationTransactionDAO(d),
		purchaseOrderSvc,
		saleOrderDAO,
		saleOrderLineDAO,
	).WithInvoiceMirror(invoiceDAO, invoiceLineDAO, invoiceSvc)
	interorgHandler := interorganizationHandler.NewInterorganizationHandler(
		interorgSvc,
	)

	consolidationSvc := interorganization.NewConsolidationService(
		interorganization.NewConsolidationRunDAO(d),
		interorganization.NewConsolidationEliminationDAO(d),
		interorganization.NewInterorganizationTransactionDAO(d),
		organizationDAO,
		taxPeriodDAO,
		accountDAO,
		accounting.NewAccountBalanceDAO(d),
		purchaseOrderLineDAO,
		productResolver,
		stockCostLayerDAO,
		reference.NewFxRateSource(fxRateDAO),
		db.NewDBTransactioner(d),
	)
	consolidationHandler := interorganizationHandler.NewConsolidationHandler(consolidationSvc)

	dropshipHandler.Register(orgAPI, guards)
	interorgHandler.Register(orgAPI, guards)
	consolidationHandler.Register(orgAPI, guards)

	reportConfigSource := reporting.NewConfigSource(systemConfigDAO)
	reportDAO := reporting.NewReportDAO(d)
	fxRevaluationDAO := reporting.NewFxRevaluationDAO(d)
	fxRevaluationLineDAO := reporting.NewFxRevaluationLineDAO(d)
	accrualDAO := reporting.NewAccrualDAO(d)
	accrualLineDAO := reporting.NewAccrualLineDAO(d)
	kpiSummaryDAO := reporting.NewKpiSummaryDAO(d)
	reportSvc := reporting.NewReportService(reportDAO, taxPeriodDAO, kpiSummaryDAO)
	kpiSummarySvc := reporting.NewKpiSummaryService(kpiSummaryDAO, taxPeriodDAO)
	kpiSvc := reporting.NewKpiService(
		reportDAO,
		pipelineSvc,
		subscriptionSvc,
		project.NewProjectDAO(d),
		projectSvc,
	)
	fxRevaluationSvc := reporting.NewFxRevaluationService(
		reportDAO,
		fxRevaluationDAO,
		fxRevaluationLineDAO,
		postingSvc,
		reportConfigSource,
		reference.NewFxRateSource(fxRateDAO),
		organizationDAO,
	)
	accrualSvc := reporting.NewAccrualService(accrualDAO, accrualLineDAO, postingSvc, reportConfigSource)
	periodCloseSvc := reporting.NewPeriodCloseService(
		taxPeriodSvc,
		taxPeriodDAO,
		reportConfigSource,
		assetDAO,
		assetSvc,
		fxRevaluationSvc,
		accrualSvc,
		deferralSvc,
		kpiSummarySvc,
	)

	reportHandler := reportingHandler.NewReportHandler(reportSvc)
	kpiHandler := reportingHandler.NewKpiHandler(kpiSvc)
	fxRevaluationHandler := reportingHandler.NewFxRevaluationHandler(fxRevaluationSvc)
	accrualHandler := reportingHandler.NewAccrualHandler(accrualSvc)
	periodCloseHandler := reportingHandler.NewPeriodCloseHandler(periodCloseSvc)
	statementSvc := reporting.NewStatementService(reportDAO, taxPeriodDAO, taxPeriodDAO, taxYearDAO, postingSvc)
	statementHandler := reportingHandler.NewStatementHandler(statementSvc)

	reportHandler.Register(orgAPI, guards)
	kpiHandler.Register(orgAPI, guards)
	fxRevaluationHandler.Register(orgAPI, guards)
	accrualHandler.Register(orgAPI, guards)
	periodCloseHandler.Register(orgAPI, guards)
	statementHandler.Register(orgAPI, guards)

	queueHealth := queue.NewHealthChecker(cfg)
	healthHandler := httpx.NewHealthHandler(cfg, d, queueHealth)
	api.Get("/health", healthHandler.Health)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		listenConfig := fiber.ListenConfig{
			EnablePrefork:         cfg.ApplicationPrefork,
			DisableStartupMessage: true,
		}
		errCh <- app.Listen(cfg.HTTPAddress(), listenConfig)
	}()

	select {
	case err := <-errCh:
		logger.Fatalf("Failed to run http server %v", err)
	case <-ctx.Done():
		slog.Info("Shutting down gracefully")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Fatalf("Failed to shutdown http server gracefully %v", err)
	}

	if limiterStorage != nil {
		_ = limiterStorage.Close()
	}
}
