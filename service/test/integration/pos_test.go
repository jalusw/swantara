//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/pos"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type posFixture struct {
	orgID           uint64
	cashierID       uint64
	configID        uint64
	journalID       uint64
	taxID           uint64
	variantID       uint64
	contactID       uint64
	stockLocationID uint64
	sessionID       uint64

	posSvc           pos.POSService
	invoiceDAO       accounting.InvoiceDAO
	journalEntryDAO  accounting.JournalEntryDAO
	stockMovementDAO inventory.StockMovementDAO
}

func TestPOSSellWithContactIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPOSFixture(t)
	ctx := testutil.SystemContext()

	session, err := fx.posSvc.OpenSession(ctx, fx.configID, fx.cashierID, 0)
	if err != nil {
		t.Fatalf("open session failed: %v", err)
	}
	fx.sessionID = session.ID

	contactID := fx.contactID
	order, err := fx.posSvc.Sell(ctx, pos.SellRequest{
		SessionID: session.ID,
		ContactID: &contactID,
		Lines: []pos.SellLineRequest{
			{ItemID: fx.variantID, Qty: 2, TaxIDs: helper.Int64Array{int64(fx.taxID)}},
		},
		Payments: []pos.SellPaymentRequest{
			{Method: "cash", Amount: 220},
		},
	})
	if err != nil {
		t.Fatalf("sell failed: %v", err)
	}

	if order.State != pos.OrderStateDone {
		t.Errorf("expected order state %q, got %q", pos.OrderStateDone, order.State)
	}
	if order.InvoiceID == nil {
		t.Fatal("expected order to reference an invoice")
	}

	invoice, err := fx.invoiceDAO.Find(ctx, *order.InvoiceID)
	if err != nil {
		t.Fatalf("find invoice failed: %v", err)
	}
	if invoice == nil {
		t.Fatal("expected invoice to exist")
	}
	if invoice.Type != accounting.InvoiceTypeCustomerInvoice {
		t.Errorf("expected invoice type %q, got %q", accounting.InvoiceTypeCustomerInvoice, invoice.Type)
	}
	if invoice.State != accounting.InvoiceStatePosted {
		t.Errorf("expected invoice state %q, got %q", accounting.InvoiceStatePosted, invoice.State)
	}

	shifts, err := fx.journalEntryDAO.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "origin_type", Operator: query.Equal, Value: "pos_order_invoice_shift"},
	}})
	if err != nil {
		t.Fatalf("list shift movements failed: %v", err)
	}
	if len(shifts.Items) != 1 {
		t.Errorf("expected 1 AR-shift movement, got %d", len(shifts.Items))
	}

	movements, err := fx.stockMovementDAO.ListByOrigin(ctx, "pos_order", order.ID)
	if err != nil {
		t.Fatalf("list stock movements failed: %v", err)
	}
	if len(movements) != 1 {
		t.Fatalf("expected 1 outgoing stock movement, got %d", len(movements))
	}
	if movements[0].State != inventory.MovementStateDone {
		t.Errorf("expected stock movement state %q, got %q", inventory.MovementStateDone, movements[0].State)
	}
}

func TestPOSRefundIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPOSFixture(t)
	ctx := testutil.SystemContext()

	session, err := fx.posSvc.OpenSession(ctx, fx.configID, fx.cashierID, 0)
	if err != nil {
		t.Fatalf("open session failed: %v", err)
	}
	fx.sessionID = session.ID

	contactID := fx.contactID
	order, err := fx.posSvc.Sell(ctx, pos.SellRequest{
		SessionID: session.ID,
		ContactID: &contactID,
		Lines: []pos.SellLineRequest{
			{ItemID: fx.variantID, Qty: 2, TaxIDs: helper.Int64Array{int64(fx.taxID)}},
		},
		Payments: []pos.SellPaymentRequest{
			{Method: "cash", Amount: 220},
		},
	})
	if err != nil {
		t.Fatalf("sell failed: %v", err)
	}

	refunded, err := fx.posSvc.Refund(ctx, order.ID, fx.journalID, time.Now().UTC())
	if err != nil {
		t.Fatalf("refund failed: %v", err)
	}
	if refunded.State != pos.OrderStateRefunded {
		t.Errorf("expected order state %q, got %q", pos.OrderStateRefunded, refunded.State)
	}

	creditNote, err := fx.invoiceDAO.FindByOrigin(ctx, *order.InvoiceID)
	if err != nil {
		t.Fatalf("find credit note failed: %v", err)
	}
	if creditNote == nil {
		t.Fatal("expected credit note to exist")
	}
	if creditNote.Type != accounting.InvoiceTypeCustomerCredit {
		t.Errorf("expected credit note type %q, got %q", accounting.InvoiceTypeCustomerCredit, creditNote.Type)
	}

	restocks, err := fx.stockMovementDAO.ListByOrigin(ctx, "pos_refund", order.ID)
	if err != nil {
		t.Fatalf("list restock movements failed: %v", err)
	}
	if len(restocks) != 1 {
		t.Fatalf("expected 1 restock movement, got %d", len(restocks))
	}
	if restocks[0].State != inventory.MovementStateDone {
		t.Errorf("expected restock movement state %q, got %q", inventory.MovementStateDone, restocks[0].State)
	}

	refunds, err := fx.journalEntryDAO.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "origin_type", Operator: query.Equal, Value: "pos_refund"},
	}})
	if err != nil {
		t.Fatalf("list refund movements failed: %v", err)
	}
	if len(refunds.Items) != 1 {
		t.Errorf("expected 1 refund GL movement, got %d", len(refunds.Items))
	}

	reversals, err := fx.journalEntryDAO.List(ctx, &query.Query{Filters: []query.Filter{
		{Field: "origin_type", Operator: query.Equal, Value: "reversal"},
	}})
	if err != nil {
		t.Fatalf("list reversal movements failed: %v", err)
	}
	if len(reversals.Items) != 1 {
		t.Errorf("expected 1 reversal movement, got %d", len(reversals.Items))
	}
}

func seedPOSFixture(t *testing.T) posFixture {
	t.Helper()
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name:         gofakeit.Company(),
		BaseCurrency: "USD",
		Timezone:     "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	userSvc := iam.NewUserService(iam.NewUserDAO(testDB), iam.NewPasswordService(testCfg))
	user, err := userSvc.Create(ctx, &iam.User{
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		Email:     gofakeit.Email(),
		Password:  "test-password",
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	if _, err := iam.NewMemberDAO(testDB).Create(ctx, &iam.Member{
		UserID:         user.ID,
		OrganizationID: org.ID,
	}); err != nil {
		t.Fatalf("create member failed: %v", err)
	}

	cash, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1000", Name: "Cash", Type: "cash", Active: true,
	})
	if err != nil {
		t.Fatalf("create cash account failed: %v", err)
	}
	income, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Sales Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create income account failed: %v", err)
	}
	taxOut, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2200", Name: "Output Tax", Type: "tax", Active: true,
	})
	if err != nil {
		t.Fatalf("create tax account failed: %v", err)
	}
	if _, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1100", Name: "Accounts Receivable", Type: "receivable", Active: true,
	}); err != nil {
		t.Fatalf("create receivable account failed: %v", err)
	}
	stockVal, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1500", Name: "Inventory", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create inventory account failed: %v", err)
	}
	stockIn, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1501", Name: "Stock Input", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create stock input account failed: %v", err)
	}
	stockOut, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1502", Name: "Stock Output", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create stock output account failed: %v", err)
	}
	cogs, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "5000", Name: "Cost of Goods Sold", Type: "cogs", Active: true,
	})
	if err != nil {
		t.Fatalf("create cogs account failed: %v", err)
	}

	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID:   org.ID,
		Name:             "Sales Journal",
		Code:             helper.Ptr("SAL"),
		Type:             "sale",
		DefaultAccountID: helper.Ptr(cash.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	tax, err := dao.NewBase[reference.Tax](testDB).Create(ctx, &reference.Tax{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Sales Tax",
		Amount:         helper.Ptr(10.0),
		Type:           reference.TaxTypePercent,
		Scope:          reference.TaxScopeSale,
		TaxAccountID:   helper.Ptr(taxOut.ID),
		Active:         true,
	})
	if err != nil {
		t.Fatalf("create tax failed: %v", err)
	}

	category, err := dao.NewBase[reference.ItemCategory](testDB).Create(ctx, &reference.ItemCategory{
		Name:                    "General",
		IncomeAccountID:         helper.Ptr(income.ID),
		StockValuationAccountID: helper.Ptr(stockVal.ID),
		StockInputAccountID:     helper.Ptr(stockIn.ID),
		StockOutputAccountID:    helper.Ptr(stockOut.ID),
		CogsAccountID:           helper.Ptr(cogs.ID),
	})
	if err != nil {
		t.Fatalf("create item category failed: %v", err)
	}

	itemDAO := products.NewItemDAO(testDB)
	itemVariantDAO := products.NewItemVariantDAO(testDB)
	template, err := itemDAO.CreateWithVariants(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Test Item",
		CategoryID:     helper.Ptr(category.ID),
		Type:           "stockable",
		ListPrice:      100,
		StandardCost:   50,
		Tracking:       "none",
		IsSellable:     true,
		Active:         true,
	}, []*products.ItemVariant{{Active: true}})
	if err != nil {
		t.Fatalf("create item template failed: %v", err)
	}
	variants, err := itemVariantDAO.ListByTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("list variants failed: %v", err)
	}
	if len(variants) == 0 {
		t.Fatal("expected at least one variant")
	}

	price_book, err := products.NewPriceBookDAO(testDB).Create(ctx, &products.PriceBook{
		Name:           "Retail",
		OrganizationID: helper.Ptr(org.ID),
		Active:         true,
	})
	if err != nil {
		t.Fatalf("create price_book failed: %v", err)
	}
	if _, err := products.NewPriceRuleDAO(testDB).Create(ctx, &products.PriceRule{
		PriceBookID: price_book.ID,
		AppliesTo:   products.AppliesToAll,
		ComputeType: products.ComputeFixed,
		FixedPrice:  helper.Ptr(100.0),
	}); err != nil {
		t.Fatalf("create price_book rule failed: %v", err)
	}

	warehouse, err := inventory.NewWarehouseDAO(testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Main Warehouse",
		Code:           helper.Ptr("WH"),
	})
	if err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}

	stockLocationDAO := inventory.NewStockLocationDAO(testDB)
	internal, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID),
		WarehouseID:    helper.Ptr(warehouse.ID),
		Name:           "Stock",
		Code:           helper.Ptr("STK"),
		Usage:          "internal",
	})
	if err != nil {
		t.Fatalf("create internal location failed: %v", err)
	}
	if _, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Customers",
		Usage:          "customer",
	}); err != nil {
		t.Fatalf("create customer location failed: %v", err)
	}
	supplier, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Suppliers",
		Usage:          "supplier",
	})
	if err != nil {
		t.Fatalf("create supplier location failed: %v", err)
	}

	contact, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(org.ID),
		Name:           gofakeit.Company(),
		Active:         true,
	})
	if err != nil {
		t.Fatalf("create contact failed: %v", err)
	}

	config, err := dao.NewBase[reference.POSConfig](testDB).Create(ctx, &reference.POSConfig{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Main POS",
		WarehouseID:    helper.Ptr(warehouse.ID),
		JournalID:      helper.Ptr(journal.ID),
		PriceBookID:    helper.Ptr(price_book.ID),
	})
	if err != nil {
		t.Fatalf("create pos config failed: %v", err)
	}

	date := time.Now().UTC()
	for _, code := range []string{pos.SequencePOSOrderCode, accounting.SequenceInvoiceCode} {
		if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
			OrganizationID: org.ID,
			Code:           code,
			NextNumber:     1,
			Padding:        5,
		}).Error; err != nil {
			t.Fatalf("create document sequence %q failed: %v", code, err)
		}
	}

	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	journalEntryLineDAO := accounting.NewJournalLineDAO(testDB)
	taxPeriodDAO := accounting.NewTaxPeriodDAO(testDB)
	taxPeriodSvc := accounting.NewTaxPeriodService(taxPeriodDAO, dao.NewBase[reference.TaxYear](testDB))
	reversalEngine := accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO)
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine)
	sequenceSvc := sequence.NewSequenceService(sequence.NewDAO(testDB))
	invoiceDAO := accounting.NewInvoiceDAO(testDB)
	invoiceSvc := accounting.NewInvoiceService(
		invoiceDAO,
		accounting.NewInvoiceLineDAO(testDB),
		accounting.NewInvoiceTaxDAO(testDB),
		postingSvc,
		dao.NewBase[reference.Account](testDB),
		dao.NewBase[reference.Tax](testDB),
		sequenceSvc,
		db.NewDBTransactioner(testDB),
	)

	stockMovementDAO := inventory.NewStockMovementDAO(testDB)
	stockCostLayerDAO := inventory.NewCostLayerDAO(testDB)
	productResolver := inventory.NewProductResolver(itemVariantDAO, itemDAO, dao.NewBase[reference.ItemCategory](testDB))
	valuationSvc := inventory.NewValuationService(
		stockMovementDAO,
		stockCostLayerDAO,
		stockLocationDAO,
		productResolver,
		postingSvc,
		db.NewDBTransactioner(testDB),
	)

	recvMove, err := stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(org.ID),
		ItemID:         variants[0].ID,
		Qty:            10,
		SrcLocationID:  supplier.ID,
		DstLocationID:  internal.ID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(date),
	})
	if err != nil {
		t.Fatalf("create receive movement failed: %v", err)
	}
	if _, err := valuationSvc.Receive(ctx, recvMove.ID, amount.FromFloat64(50), journal.ID, date); err != nil {
		t.Fatalf("receive stock failed: %v", err)
	}

	posSvc := pos.NewPOSService(
		dao.NewBase[reference.POSConfig](testDB),
		pos.NewPOSSessionDAO(testDB),
		pos.NewPOSOrderDAO(testDB),
		pos.NewPOSOrderLineDAO(testDB),
		pos.NewPOSPaymentDAO(testDB),
		products.NewProductService(
			itemDAO,
			itemVariantDAO,
			dao.NewBase[reference.ItemCategory](testDB),
			products.NewPriceBookDAO(testDB),
			products.NewPriceRuleDAO(testDB),
		),
		dao.NewBase[reference.Tax](testDB),
		dao.NewBase[reference.Journal](testDB),
		dao.NewBase[reference.Account](testDB),
		dao.NewBase[reference.POSPaymentAccount](testDB),
		stockLocationDAO,
		inventory.NewStockBalanceDAO(testDB),
		stockMovementDAO,
		stockCostLayerDAO,
		inventory.NewShipmentDAO(testDB),
		valuationSvc,
		invoiceSvc,
		postingSvc,
		iam.NewMemberDAO(testDB),
		sequenceSvc,
		db.NewDBTransactioner(testDB),
	)

	return posFixture{
		orgID:            org.ID,
		cashierID:        user.ID,
		configID:         config.ID,
		journalID:        journal.ID,
		taxID:            tax.ID,
		variantID:        variants[0].ID,
		contactID:        contact.ID,
		stockLocationID:  internal.ID,
		posSvc:           posSvc,
		invoiceDAO:       invoiceDAO,
		journalEntryDAO:  journalEntryDAO,
		stockMovementDAO: stockMovementDAO,
	}
}
