//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/commission"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/service"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type phase36Fixture struct {
	orgID             uint64
	journalID         uint64
	cashAcc           uint64
	giftLiabilityAcc  uint64
	revenueAcc        uint64
	refundAcc         uint64
	commExpenseAcc    uint64
	commPayableAcc    uint64
	cogsAcc           uint64
	stockValuationAcc uint64
	contactID         uint64

	commissionSvc commission.CommissionService
	giftSvc       giftcard.GiftCardService
	serviceSvc    service.ServiceService
	orderDAO      service.ServiceOrderDAO
	moveLineDAO   accounting.JournalLineDAO
	invoiceDAO    accounting.InvoiceDAO
	invoiceSvc    accounting.InvoiceService
}

func seedPhase36Fixture(t *testing.T) phase36Fixture {
	t.Helper()
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	contact, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(org.ID), Name: gofakeit.Name(),
	})
	if err != nil {
		t.Fatalf("create contact failed: %v", err)
	}

	cashAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1001", Name: "Cash", Type: "bank", Active: true,
	})
	if err != nil {
		t.Fatalf("create cash account failed: %v", err)
	}
	giftLiabilityAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2300", Name: "Gift Card Liability", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create gift liability account failed: %v", err)
	}
	revenueAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Service Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create revenue account failed: %v", err)
	}
	refundAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4100", Name: "Sales Refund", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create refund account failed: %v", err)
	}
	commExpenseAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "6300", Name: "Commission Expense", Type: "expense", Active: true,
	})
	if err != nil {
		t.Fatalf("create commission expense account failed: %v", err)
	}
	commPayableAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2110", Name: "Commission Payable", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create commission payable account failed: %v", err)
	}
	receivableAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1200", Name: "Receivable", Type: "receivable", Active: true,
	})
	if err != nil {
		t.Fatalf("create receivable account failed: %v", err)
	}
	_ = receivableAcc
	cogsAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "5100", Name: "COGS", Type: "expense", Active: true,
	})
	if err != nil {
		t.Fatalf("create cogs account failed: %v", err)
	}
	stockValuationAcc, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1100", Name: "Stock Valuation", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create stock valuation account failed: %v", err)
	}

	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Phase 36 Journal", Code: helper.Ptr("P36"),
		Type: "general", DefaultAccountID: helper.Ptr(cashAcc.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
		OrganizationID: org.ID, Code: accounting.SequenceInvoiceCode, NextNumber: 1, Padding: 5,
	}).Error; err != nil {
		t.Fatalf("create invoice sequence failed: %v", err)
	}

	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	journalEntryLineDAO := accounting.NewJournalLineDAO(testDB)
	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(testDB), dao.NewBase[reference.TaxYear](testDB))
	reversalEngine := accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO)
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine)

	invoiceSvc := accounting.NewInvoiceService(
		accounting.NewInvoiceDAO(testDB),
		accounting.NewInvoiceLineDAO(testDB),
		accounting.NewInvoiceTaxDAO(testDB),
		postingSvc,
		dao.NewBase[reference.Account](testDB),
		dao.NewBase[reference.Tax](testDB),
		sequence.NewSequenceService(sequence.NewDAO(testDB)),
		db.NewDBTransactioner(testDB),
	)

	commissionSvc := commission.NewCommissionService(
		commission.NewCommissionPlanDAO(testDB),
		commission.NewCommissionRuleDAO(testDB),
		commission.NewCommissionAssignmentDAO(testDB),
		commission.NewCommissionEntryDAO(testDB),
		postingSvc,
		accounting.NewInvoiceDAO(testDB),
		accounting.NewInvoiceLineDAO(testDB),
		accounting.NewPaymentAllocationDAO(testDB),
		inventory.NewProductResolver(
			products.NewItemVariantDAO(testDB),
			products.NewItemDAO(testDB),
			dao.NewBase[reference.ItemCategory](testDB),
		),
		db.NewDBTransactioner(testDB),
	)

	giftSvc := giftcard.NewGiftCardService(
		giftcard.NewGiftCardDAO(testDB),
		giftcard.NewGiftCardTransactionDAO(testDB),
		postingSvc,
		db.NewDBTransactioner(testDB),
	)

	serviceSvc := service.NewServiceService(
		service.NewEquipmentDAO(testDB),
		service.NewServiceContractDAO(testDB),
		service.NewServiceOrderDAO(testDB),
		service.NewServiceOrderLineDAO(testDB),
		postingSvc,
		invoiceSvc,
		db.NewDBTransactioner(testDB),
	)

	return phase36Fixture{
		orgID:             org.ID,
		journalID:         journal.ID,
		cashAcc:           cashAcc.ID,
		giftLiabilityAcc:  giftLiabilityAcc.ID,
		revenueAcc:        revenueAcc.ID,
		refundAcc:         refundAcc.ID,
		commExpenseAcc:    commExpenseAcc.ID,
		commPayableAcc:    commPayableAcc.ID,
		cogsAcc:           cogsAcc.ID,
		stockValuationAcc: stockValuationAcc.ID,
		contactID:         contact.ID,
		commissionSvc:     commissionSvc,
		giftSvc:           giftSvc,
		serviceSvc:        serviceSvc,
		orderDAO:          service.NewServiceOrderDAO(testDB),
		moveLineDAO:       journalEntryLineDAO,
		invoiceDAO:        accounting.NewInvoiceDAO(testDB),
		invoiceSvc:        invoiceSvc,
	}
}

func TestCommissionAccrueIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase36Fixture(t)
	ctx := testutil.SystemContext()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	plan, err := fx.commissionSvc.CreatePlan(ctx, commission.CreatePlanRequest{
		OrganizationID: fx.orgID, Name: "Sales Plan", Basis: commission.BasisRevenue,
	})
	if err != nil {
		t.Fatalf("create plan failed: %v", err)
	}
	if _, err := fx.commissionSvc.CreateRule(ctx, commission.CreateRuleRequest{PlanID: plan.ID, RatePct: 5}); err != nil {
		t.Fatalf("create rule failed: %v", err)
	}
	if _, err := fx.commissionSvc.Assign(ctx, commission.AssignRequest{PlanID: plan.ID, SalespersonID: fx.contactID, DateStart: start}); err != nil {
		t.Fatalf("assign failed: %v", err)
	}

	entry, err := fx.commissionSvc.Accrue(ctx, commission.AccrueRequest{
		SalespersonID:    fx.contactID,
		PlanID:           plan.ID,
		SourceType:       "invoice",
		SourceID:         7,
		BaseAmount:       1000,
		JournalID:        fx.journalID,
		ExpenseAccountID: fx.commExpenseAcc,
		PayableAccountID: fx.commPayableAcc,
		Date:             time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("accrue failed: %v", err)
	}
	if entry.State != commission.EntryStateConfirmed || entry.CommissionAmount != 50 {
		t.Errorf("entry = %+v, want confirmed of 50", entry)
	}

	paid, err := fx.commissionSvc.Pay(ctx, commission.PayRequest{EntryID: entry.ID})
	if err != nil {
		t.Fatalf("pay failed: %v", err)
	}
	if paid.State != commission.EntryStatePaid {
		t.Errorf("entry = %+v, want paid", paid)
	}
}

func TestCommissionAccrueFromInvoiceRevenueIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase36Fixture(t)
	ctx := testutil.SystemContext()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	plan, err := fx.commissionSvc.CreatePlan(ctx, commission.CreatePlanRequest{
		OrganizationID: fx.orgID, Name: "Revenue Plan", Basis: commission.BasisRevenue,
	})
	if err != nil {
		t.Fatalf("create plan failed: %v", err)
	}
	if _, err := fx.commissionSvc.CreateRule(ctx, commission.CreateRuleRequest{PlanID: plan.ID, RatePct: 5}); err != nil {
		t.Fatalf("create rule failed: %v", err)
	}
	if _, err := fx.commissionSvc.Assign(ctx, commission.AssignRequest{PlanID: plan.ID, SalespersonID: fx.contactID, DateStart: start}); err != nil {
		t.Fatalf("assign failed: %v", err)
	}

	invoice, err := fx.invoiceSvc.Create(ctx, accounting.CreateInvoiceRequest{
		OrganizationID: fx.orgID,
		JournalID:      fx.journalID,
		ContactID:      fx.contactID,
		Date:           time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		Lines: []accounting.InvoiceLineRequest{
			{Description: "Consulting", Qty: 2, UnitPrice: 500, AccountID: fx.revenueAcc},
		},
	})
	if err != nil {
		t.Fatalf("create invoice failed: %v", err)
	}

	entry, err := fx.commissionSvc.AccrueFromInvoice(ctx, commission.AccrueFromInvoiceRequest{
		InvoiceID:        invoice.ID,
		SalespersonID:    fx.contactID,
		JournalID:        fx.journalID,
		ExpenseAccountID: fx.commExpenseAcc,
		PayableAccountID: fx.commPayableAcc,
		Date:             time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("accrue from invoice failed: %v", err)
	}
	if entry.BaseAmount != 1000 || entry.CommissionAmount != 50 || entry.State != commission.EntryStateConfirmed {
		t.Errorf("entry = %+v, want base 1000 and commission 50 confirmed", entry)
	}

	duplicate, err := fx.commissionSvc.AccrueFromInvoice(ctx, commission.AccrueFromInvoiceRequest{
		InvoiceID: invoice.ID, SalespersonID: fx.contactID, JournalID: fx.journalID,
		ExpenseAccountID: fx.commExpenseAcc, PayableAccountID: fx.commPayableAcc,
		Date: time.Date(2026, 2, 2, 0, 0, 0, 0, time.UTC),
	})
	if duplicate != nil || err == nil {
		t.Errorf("duplicate accrual = %+v, want rejected", duplicate)
	}
}

func TestGiftCardIssueRedeemRefundIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase36Fixture(t)
	ctx := testutil.SystemContext()

	card, err := fx.giftSvc.Issue(ctx, giftcard.IssueRequest{
		OrganizationID: fx.orgID, Code: "GC-001", Amount: 100, CurrencyCode: "USD",
		JournalID: fx.journalID, CashAccountID: fx.cashAcc, LiabilityAccountID: fx.giftLiabilityAcc,
		Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("issue failed: %v", err)
	}
	if card.Balance != 100 || card.State != giftcard.GiftCardStateActive {
		t.Errorf("card = %+v, want active balance 100", card)
	}

	redeemed, err := fx.giftSvc.Redeem(ctx, giftcard.RedeemRequest{
		GiftCardID: card.ID, Amount: 30, JournalID: fx.journalID,
		RevenueAccountID: fx.revenueAcc, LiabilityAccountID: fx.giftLiabilityAcc,
		Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("redeem failed: %v", err)
	}
	if redeemed.Balance != 70 {
		t.Errorf("balance = %v, want 70", redeemed.Balance)
	}

	refunded, err := fx.giftSvc.Refund(ctx, giftcard.RefundRequest{
		GiftCardID: card.ID, Amount: 10, JournalID: fx.journalID,
		RefundAccountID: fx.refundAcc, LiabilityAccountID: fx.giftLiabilityAcc,
		Date: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}
	if refunded.Balance != 80 || refunded.State != giftcard.GiftCardStateActive {
		t.Errorf("card = %+v, want active balance 80", refunded)
	}

	transactions, err := fx.giftSvc.ListTransactions(ctx, card.ID)
	if err != nil {
		t.Fatalf("list transactions failed: %v", err)
	}
	if len(transactions) != 3 {
		t.Errorf("transactions = %d, want 3 (issue/redeem/refund)", len(transactions))
	}
}

func TestServiceOrderCompleteAndBillIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase36Fixture(t)
	ctx := testutil.SystemContext()
	date := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	order, err := fx.serviceSvc.CreateOrder(ctx, service.CreateOrderRequest{
		OrganizationID: fx.orgID,
		Name:           "Fix printer",
		ContactID:      &fx.contactID,
		Type:           service.OrderTypeRepair,
		Lines: []service.LineRequest{
			{Type: service.LineTypePart, Description: "Toner", Qty: 2, UnitCost: 25, UnitPrice: 80, Billable: true},
			{Type: service.LineTypeLabor, Description: "Labor", Qty: 2, UnitPrice: 50, Billable: true},
			{Type: service.LineTypePart, Description: "Warranty gasket", Qty: 1, UnitCost: 5, UnitPrice: 10, Billable: false, CoveredByWarranty: true},
		},
	})
	if err != nil {
		t.Fatalf("create order failed: %v", err)
	}
	if order.State != service.OrderStateNew {
		t.Errorf("order = %+v, want new", order)
	}

	scheduled, err := fx.serviceSvc.Schedule(ctx, service.ScheduleRequest{OrderID: order.ID, ScheduledDate: date})
	if err != nil {
		t.Fatalf("schedule failed: %v", err)
	}
	if scheduled.State != service.OrderStateScheduled {
		t.Errorf("order = %+v, want scheduled", scheduled)
	}

	started, err := fx.serviceSvc.Start(ctx, order.ID)
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if started.State != service.OrderStateInProgress {
		t.Errorf("order = %+v, want in_progress", started)
	}

	completed, err := fx.serviceSvc.Complete(ctx, service.CompleteRequest{
		OrderID: order.ID, JournalID: fx.journalID, Date: date,
		COGSAccountID: fx.cogsAcc, StockValuationAccountID: fx.stockValuationAcc, Resolution: "Replaced toner",
	})
	if err != nil {
		t.Fatalf("complete failed: %v", err)
	}
	if completed.State != service.OrderStateDone {
		t.Fatalf("order = %+v, want done", completed)
	}

	invoiced, err := fx.serviceSvc.Bill(ctx, service.BillRequest{
		OrderID: order.ID, JournalID: fx.journalID, Date: date, RevenueAccountID: fx.revenueAcc,
	})
	if err != nil {
		t.Fatalf("bill failed: %v", err)
	}
	if invoiced.State != service.OrderStateInvoiced {
		t.Fatalf("order = %+v, want invoiced", invoiced)
	}
	if invoiced.InvoiceID == nil {
		t.Fatalf("order has no invoice link")
	}
	invoice, err := fx.invoiceDAO.Find(ctx, *invoiced.InvoiceID)
	if err != nil {
		t.Fatalf("find invoice failed: %v", err)
	}
	if invoice == nil {
		t.Fatalf("invoice not found")
	}
	if invoice.AmountUntaxed.Float64() != 260 {
		t.Errorf("invoice untaxed = %v, want 260 (2*80 + 2*50)", invoice.AmountUntaxed)
	}
}
