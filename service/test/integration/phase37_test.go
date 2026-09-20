//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type phase37Fixture struct {
	groupOrgID     uint64
	memberOrgID    uint64
	org1JournalID  uint64
	org2JournalID  uint64
	customerLocID  uint64
	supplierLocID  uint64
	supplierID     uint64
	org2SupplierID uint64
	customerID     uint64
	variantID      uint64
	org1Receivable uint64
	org1Payable    uint64
	org1Income     uint64
	org1Expense    uint64
	org2Receivable uint64
	org2Payable    uint64
	org2Income     uint64
	org2Expense    uint64
	periodID       uint64

	postingSvc    accounting.PostingService
	dropshipSvc   interorganization.DropShipService
	interorgSvc   interorganization.InterorganizationService
	consolidation interorganization.ConsolidationService
	soDAO         sales.SaleOrderDAO
	soLineDAO     sales.SaleOrderLineDAO
	poLineDAO     procurement.PurchaseOrderLineDAO
	linkDAO       interorganization.DropshipLinkDAO
	transDAO      interorganization.InterorganizationTransactionDAO
	moveLineDAO   accounting.JournalLineDAO
}

type phase37ExpenseEngine struct{}

func (phase37ExpenseEngine) ResolveExpenseAccount(_ context.Context, _ uint64) (uint64, error) {
	return 0, nil
}

type phase37ReceiveEngine struct{}

func (phase37ReceiveEngine) Receive(ctx context.Context, moveID uint64, _ amount.Amount, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
	return &inventory.CostLayer{MovementID: &moveID}, nil
}

func seedPhase37Fixture(t *testing.T) phase37Fixture {
	t.Helper()
	ctx := testutil.SystemContext()

	groupOrg, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create group organization failed: %v", err)
	}
	memberOrg, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC", ParentID: helper.Ptr(groupOrg.ID),
	})
	if err != nil {
		t.Fatalf("create member organization failed: %v", err)
	}

	accounts := map[uint64]map[string]*reference.Account{}
	for _, org := range []*reference.Organization{groupOrg, memberOrg} {
		orgAccounts := map[string]*reference.Account{}
		for _, spec := range []struct {
			key  string
			code string
			name string
			typ  string
		}{
			{"receivable", "1200", "Receivable", "receivable"},
			{"payable", "2100", "Payable", "payable"},
			{"income", "4000", "Income", "income"},
			{"expense", "6000", "Expense", "expense"},
			{"bank", "1001", "Bank", "bank"},
		} {
			account, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
				OrganizationID: org.ID, Code: spec.code, Name: spec.name, Type: spec.typ, Active: true,
			})
			if err != nil {
				t.Fatalf("create account %s failed: %v", spec.key, err)
			}
			orgAccounts[spec.key] = account
		}
		if org.ID == groupOrg.ID {
			for _, spec := range []struct {
				key  string
				code string
				name string
				typ  string
			}{
				{"stock_input", "5200", "Stock Input", "expense"},
				{"stock_cost", "1100", "Stock Valuation", "asset"},
				{"cogs", "5100", "COGS", "expense"},
			} {
				account, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
					OrganizationID: org.ID, Code: spec.code, Name: spec.name, Type: spec.typ, Active: true,
				})
				if err != nil {
					t.Fatalf("create account %s failed: %v", spec.key, err)
				}
				orgAccounts[spec.key] = account
			}
		}
		accounts[org.ID] = orgAccounts
	}

	org1Journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: groupOrg.ID, Name: "Phase 37 Journal", Code: helper.Ptr("P37"),
		Type: "general", DefaultAccountID: helper.Ptr(accounts[groupOrg.ID]["bank"].ID),
	})
	if err != nil {
		t.Fatalf("create org1 journal failed: %v", err)
	}
	org2Journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: memberOrg.ID, Name: "Phase 37 Member Journal", Code: helper.Ptr("P37M"),
		Type: "general", DefaultAccountID: helper.Ptr(accounts[memberOrg.ID]["bank"].ID),
	})
	if err != nil {
		t.Fatalf("create org2 journal failed: %v", err)
	}

	customer, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(groupOrg.ID), Name: gofakeit.Name(),
	})
	if err != nil {
		t.Fatalf("create customer contact failed: %v", err)
	}
	supplier, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(groupOrg.ID), Name: gofakeit.Company(),
	})
	if err != nil {
		t.Fatalf("create supplier contact failed: %v", err)
	}
	if _, err := contacts.NewSupplierProfileDAO(testDB).Create(ctx, &contacts.SupplierProfile{
		ContactID: supplier.ID, Active: true,
	}); err != nil {
		t.Fatalf("create supplier supplier failed: %v", err)
	}
	org2Vendor, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(memberOrg.ID), Name: gofakeit.Company(),
	})
	if err != nil {
		t.Fatalf("create org2 supplier contact failed: %v", err)
	}
	if _, err := contacts.NewSupplierProfileDAO(testDB).Create(ctx, &contacts.SupplierProfile{
		ContactID: org2Vendor.ID, Active: true,
	}); err != nil {
		t.Fatalf("create org2 supplier supplier failed: %v", err)
	}

	category, err := dao.NewBase[reference.ItemCategory](testDB).Create(ctx, &reference.ItemCategory{
		Name:                    "Phase 37 Category",
		StockValuationAccountID: helper.Ptr(accounts[groupOrg.ID]["stock_cost"].ID),
		StockInputAccountID:     helper.Ptr(accounts[groupOrg.ID]["stock_input"].ID),
		CogsAccountID:           helper.Ptr(accounts[groupOrg.ID]["cogs"].ID),
	})
	if err != nil {
		t.Fatalf("create item category failed: %v", err)
	}
	template, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(groupOrg.ID), Name: "Widget", CategoryID: helper.Ptr(category.ID),
		Type: "stockable", StandardCost: 100, IsPurchasable: true, IsSellable: true, Tracking: "none",
	})
	if err != nil {
		t.Fatalf("create item template failed: %v", err)
	}
	variant, err := products.NewItemVariantDAO(testDB).Create(ctx, &products.ItemVariant{
		ItemID: template.ID, Active: true,
	})
	if err != nil {
		t.Fatalf("create item variant failed: %v", err)
	}

	customerLoc, err := inventory.NewStockLocationDAO(testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(groupOrg.ID), Name: "Customer", Usage: "customer",
	})
	if err != nil {
		t.Fatalf("create customer location failed: %v", err)
	}
	supplierLoc, err := inventory.NewStockLocationDAO(testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(groupOrg.ID), Name: "Supplier", Usage: "supplier",
	})
	if err != nil {
		t.Fatalf("create supplier location failed: %v", err)
	}

	for _, org := range []*reference.Organization{groupOrg, memberOrg} {
		if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
			OrganizationID: org.ID, Code: procurement.SequencePurchaseOrderCode, NextNumber: 1, Padding: 5,
		}).Error; err != nil {
			t.Fatalf("create purchase order sequence failed: %v", err)
		}
	}

	year, err := dao.NewBase[reference.TaxYear](testDB).Create(ctx, &reference.TaxYear{
		OrganizationID: helper.Ptr(groupOrg.ID), Name: "2026", State: helper.Ptr("open"),
	})
	if err != nil {
		t.Fatalf("create tax year failed: %v", err)
	}
	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(testDB), dao.NewBase[reference.TaxYear](testDB))
	period, err := taxPeriodSvc.Create(ctx, &accounting.TaxPeriod{
		OrganizationID: groupOrg.ID, TaxYearID: year.ID, Name: "Feb 2026",
		DateStart: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)),
		DateEnd:   helper.Ptr(time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)),
	})
	if err != nil {
		t.Fatalf("create tax period failed: %v", err)
	}

	postingSvc := accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB)).SetPeriods(taxPeriodSvc)
	productResolver := inventory.NewProductResolver(
		products.NewItemVariantDAO(testDB),
		products.NewItemDAO(testDB),
		dao.NewBase[reference.ItemCategory](testDB),
	)

	purchaseOrderSvc := phasePurchaseOrderService(t, dao.NewBase[reference.SystemConfig](testDB), dao.NewBase[reference.Tax](testDB), productResolver)

	dropshipSvc := interorganization.NewDropShipService(
		interorganization.NewDropshipLinkDAO(testDB),
		purchaseOrderSvc,
		procurement.NewPurchaseOrderDAO(testDB),
		procurement.NewPurchaseOrderLineDAO(testDB),
		sales.NewSaleOrderDAO(testDB),
		sales.NewSaleOrderLineDAO(testDB),
		inventory.NewStockMovementDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		productResolver,
		postingSvc,
		db.NewDBTransactioner(testDB),
	)

	interorgSvc := interorganization.NewInterorganizationService(
		interorganization.NewInterorganizationRuleDAO(testDB),
		interorganization.NewInterorganizationTransactionDAO(testDB),
		purchaseOrderSvc,
		sales.NewSaleOrderDAO(testDB),
		sales.NewSaleOrderLineDAO(testDB),
	)

	consolidationSvc := interorganization.NewConsolidationService(
		interorganization.NewConsolidationRunDAO(testDB),
		interorganization.NewConsolidationEliminationDAO(testDB),
		interorganization.NewInterorganizationTransactionDAO(testDB),
		dao.NewBase[reference.Organization](testDB),
		accounting.NewTaxPeriodDAO(testDB),
		dao.NewBase[reference.Account](testDB),
		accounting.NewAccountBalanceDAO(testDB),
		procurement.NewPurchaseOrderLineDAO(testDB),
		productResolver,
		inventory.NewCostLayerDAO(testDB),
		reference.NewFxRateSource(dao.NewBase[reference.FxRate](testDB)),
		db.NewDBTransactioner(testDB),
	)

	return phase37Fixture{
		groupOrgID:     groupOrg.ID,
		memberOrgID:    memberOrg.ID,
		org1JournalID:  org1Journal.ID,
		org2JournalID:  org2Journal.ID,
		customerLocID:  customerLoc.ID,
		supplierLocID:  supplierLoc.ID,
		supplierID:     supplier.ID,
		org2SupplierID: org2Vendor.ID,
		customerID:     customer.ID,
		variantID:      variant.ID,
		org1Receivable: accounts[groupOrg.ID]["receivable"].ID,
		org1Payable:    accounts[groupOrg.ID]["payable"].ID,
		org1Income:     accounts[groupOrg.ID]["income"].ID,
		org1Expense:    accounts[groupOrg.ID]["expense"].ID,
		org2Receivable: accounts[memberOrg.ID]["receivable"].ID,
		org2Payable:    accounts[memberOrg.ID]["payable"].ID,
		org2Income:     accounts[memberOrg.ID]["income"].ID,
		org2Expense:    accounts[memberOrg.ID]["expense"].ID,
		periodID:       period.ID,

		postingSvc:    postingSvc,
		dropshipSvc:   dropshipSvc,
		interorgSvc:   interorgSvc,
		consolidation: consolidationSvc,
		soDAO:         sales.NewSaleOrderDAO(testDB),
		soLineDAO:     sales.NewSaleOrderLineDAO(testDB),
		poLineDAO:     procurement.NewPurchaseOrderLineDAO(testDB),
		linkDAO:       interorganization.NewDropshipLinkDAO(testDB),
		transDAO:      interorganization.NewInterorganizationTransactionDAO(testDB),
		moveLineDAO:   accounting.NewJournalLineDAO(testDB),
	}
}

func phase37SaleOrder(t *testing.T, fx phase37Fixture, organizationID uint64, quantity, unitPrice float64) *sales.SaleOrder {
	t.Helper()
	ctx := testutil.SystemContext()
	orderDate := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	order, err := fx.soDAO.CreateWithLines(ctx, &sales.SaleOrder{
		OrganizationID: helper.Ptr(organizationID),
		Name:           helper.Ptr("SO/00001"),
		ContactID:      fx.customerID,
		CurrencyCode:   helper.Ptr("IDR"),
		State:          sales.OrderStateConfirmed,
		OrderDate:      &orderDate,
		DeliveryStatus: sales.DeliveryStatusPending,
		InvoiceStatus:  "no",
		AmountUntaxed:  quantity * unitPrice,
		AmountTax:      0,
		AmountTotal:    quantity * unitPrice,
	}, []*sales.SaleOrderLine{
		{ItemID: helper.Ptr(fx.variantID), Description: helper.Ptr("Widget"), QtyOrdered: quantity, QtyDelivered: 0, UnitPrice: unitPrice, DiscountPct: 0, PriceSubtotal: quantity * unitPrice},
	})
	if err != nil {
		t.Fatalf("create sale order failed: %v", err)
	}
	return order
}

func phasePurchaseOrderService(t *testing.T, systemConfigs dao.Base[reference.SystemConfig], taxes dao.Base[reference.Tax], productResolver inventory.ItemResolver) procurement.PurchaseOrderService {
	t.Helper()
	return procurement.NewPurchaseOrderService(
		procurement.NewPurchaseOrderDAO(testDB),
		procurement.NewPurchaseOrderLineDAO(testDB),
		procurement.NewPurchaseRequestDAO(testDB),
		procurement.NewPurchaseRequestLineDAO(testDB),
		sequence.NewSequenceService(sequence.NewDAO(testDB)),
		procurement.NewSystemConfigSource(systemConfigs),
		procurement.OfferEngineMock{},
		contacts.NewContactDAO(testDB),
		contacts.NewSupplierProfileDAO(testDB),
		inventory.NewWarehouseDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		inventory.NewShipmentDAO(testDB),
		inventory.NewStockMovementDAO(testDB),
		productResolver,
		phase37ExpenseEngine{},
		taxes,
		phase37ReceiveEngine{},
		procurement.ApprovalEngineMock{},
		procurement.BillEngineMock{},
		procurement.OpenInvoiceLookupMock{},
		procurement.OutboundPaymentEngineMock{},
		procurement.QualityEngineMock{},
	)
}

func TestDropShipCreateReceiveIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase37Fixture(t)
	ctx := testutil.SystemContext()
	date := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)

	order := phase37SaleOrder(t, fx, fx.groupOrgID, 2, 250)

	po, links, err := fx.dropshipSvc.Create(ctx, interorganization.CreateDropshipOrderRequest{
		OrganizationID: fx.groupOrgID, SaleOrderID: order.ID, SupplierID: fx.supplierID, Date: date,
	})
	if err != nil {
		t.Fatalf("dropship create failed: %v", err)
	}
	if po.OrganizationID == nil || *po.OrganizationID != fx.groupOrgID {
		t.Errorf("po = %+v, want group org", po)
	}
	if po.SupplierID != fx.supplierID {
		t.Errorf("po supplier = %v, want %v", po.SupplierID, fx.supplierID)
	}
	if po.DestLocationID == nil || *po.DestLocationID != fx.customerLocID {
		t.Errorf("po dest location = %v, want customer %v", po.DestLocationID, fx.customerLocID)
	}
	if po.State != procurement.PurchaseOrderStateDraft {
		t.Errorf("po state = %v, want draft", po.State)
	}
	if len(links) != 1 {
		t.Fatalf("links = %d, want 1", len(links))
	}

	received, err := fx.dropshipSvc.Receive(ctx, interorganization.ReceiveDropshipRequest{
		OrganizationID: fx.groupOrgID, PurchaseOrderID: po.ID, JournalID: fx.org1JournalID, Date: date,
	})
	if err != nil {
		t.Fatalf("dropship receive failed: %v", err)
	}
	if received.State != procurement.PurchaseOrderStateDone || received.ReceiptStatus != procurement.PurchaseOrderReceiptStatusDone {
		t.Errorf("po = %+v, want done", received)
	}

	poLines, err := fx.poLineDAO.ListByOrder(ctx, po.ID)
	if err != nil {
		t.Fatalf("list po lines failed: %v", err)
	}
	if len(poLines) != 1 || poLines[0].QtyReceived != 2 {
		t.Errorf("po lines = %+v, want one line received 2", poLines)
	}

	soLines, err := fx.soLineDAO.ListByOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("list so lines failed: %v", err)
	}
	if len(soLines) != 1 || soLines[0].QtyDelivered != 2 {
		t.Errorf("so lines = %+v, want one line delivered 2", soLines)
	}

	so, err := fx.soDAO.Find(ctx, order.ID)
	if err != nil {
		t.Fatalf("find so failed: %v", err)
	}
	if so.DeliveryStatus != sales.DeliveryStatusDone {
		t.Errorf("so delivery status = %v, want done", so.DeliveryStatus)
	}

	links, err = fx.linkDAO.ListByPurchaseOrder(ctx, po.ID)
	if err != nil {
		t.Fatalf("list links failed: %v", err)
	}
	if len(links) != 1 || links[0].StockMovementID == nil {
		t.Fatalf("links = %+v, want one linked movement", links)
	}

	movement, err := dao.NewBase[accounting.JournalEntry](testDB).Search(ctx, "origin_id", *links[0].StockMovementID)
	if err != nil {
		t.Fatalf("find dropship movement failed: %v", err)
	}
	if movement == nil || movement.OriginType == nil || *movement.OriginType != "dropship" {
		t.Fatalf("movement = %+v, want dropship origin", movement)
	}
	lines, err := fx.moveLineDAO.ListByMovement(ctx, movement.ID)
	if err != nil {
		t.Fatalf("list movement lines failed: %v", err)
	}
	var debit, credit float64
	for _, line := range lines {
		debit += line.Debit.Float64()
		credit += line.Credit.Float64()
	}
	if debit != 500 || credit != 500 {
		t.Errorf("movement lines debit %v credit %v, want 500/500 (2 x 250)", debit, credit)
	}
}

func TestInterorganizationMirrorIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase37Fixture(t)
	ctx := testutil.SystemContext()

	rule, err := fx.interorgSvc.CreateRule(ctx, interorganization.UpsertInterorganizationRuleRequest{
		FromOrganizationID: helper.Ptr(fx.groupOrgID),
		ToOrganizationID:   helper.Ptr(fx.memberOrgID),
		AutoMirror:         true,
		SupplierContactID:  helper.Ptr(fx.org2SupplierID),
		CustomerContactID:  helper.Ptr(fx.customerID),
	})
	if err != nil {
		t.Fatalf("create rule failed: %v", err)
	}
	if !rule.AutoMirror {
		t.Errorf("rule = %+v, want auto mirror", rule)
	}

	order := phase37SaleOrder(t, fx, fx.groupOrgID, 2, 250)

	po, transaction, err := fx.interorgSvc.MirrorSaleOrder(ctx, interorganization.MirrorSaleOrderRequest{
		SaleOrderID: order.ID, ToOrganizationID: fx.memberOrgID,
	})
	if err != nil {
		t.Fatalf("mirror failed: %v", err)
	}
	if po.OrganizationID == nil || *po.OrganizationID != fx.memberOrgID {
		t.Errorf("po = %+v, want member org", po)
	}
	if po.SupplierID != fx.org2SupplierID {
		t.Errorf("po supplier = %v, want org2 supplier %v", po.SupplierID, fx.org2SupplierID)
	}
	if po.State != procurement.PurchaseOrderStateDraft {
		t.Errorf("po state = %v, want draft", po.State)
	}
	if transaction.MirrorType != "purchase_order" || transaction.State != interorganization.TransactionStateDone {
		t.Errorf("transaction = %+v, want done purchase order", transaction)
	}
	if transaction.Amount != 500 {
		t.Errorf("transaction amount = %v, want 500", transaction.Amount)
	}

	transactions, err := fx.interorgSvc.ListTransactions(ctx, fx.groupOrgID)
	if err != nil {
		t.Fatalf("list transactions failed: %v", err)
	}
	if len(transactions) != 1 {
		t.Errorf("transactions = %d, want 1", len(transactions))
	}
}

func TestConsolidationRunIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase37Fixture(t)
	ctx := testutil.SystemContext()

	saleDate := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	billDate := time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC)

	if _, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.groupOrgID, JournalID: fx.org1JournalID, Date: saleDate,
		Ref: "SALE-001", OriginType: "sale", OriginID: 1, Description: "Group revenue",
		Lines: []accounting.PostingLine{
			{AccountID: fx.org1Receivable, Name: "AR", Debit: amount.FromFloat64(100)},
			{AccountID: fx.org1Income, Name: "Revenue", Credit: amount.FromFloat64(100)},
		},
	}); err != nil {
		t.Fatalf("post group revenue failed: %v", err)
	}

	if _, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.memberOrgID, JournalID: fx.org2JournalID, Date: billDate,
		Ref: "BILL-001", OriginType: "bill", OriginID: 1, Description: "Member cost",
		Lines: []accounting.PostingLine{
			{AccountID: fx.org2Expense, Name: "Expense", Debit: amount.FromFloat64(100)},
			{AccountID: fx.org2Payable, Name: "Payable", Credit: amount.FromFloat64(100)},
		},
	}); err != nil {
		t.Fatalf("post member cost failed: %v", err)
	}

	if _, err := fx.transDAO.Create(ctx, &interorganization.InterorganizationTransaction{
		SourceOrganizationID: helper.Ptr(fx.groupOrgID),
		SourceType:           "sale_order",
		SourceID:             helper.Ptr(uint64(1)),
		MirrorOrganizationID: helper.Ptr(fx.memberOrgID),
		MirrorType:           "purchase_order",
		MirrorID:             helper.Ptr(uint64(1)),
		Amount:               100,
		State:                interorganization.TransactionStateDone,
	}); err != nil {
		t.Fatalf("create transaction failed: %v", err)
	}

	run, err := fx.consolidation.CreateRun(ctx, interorganization.CreateConsolidationRunRequest{
		GroupOrganizationID: fx.groupOrgID, PeriodID: fx.periodID, ReportingCurrency: "IDR",
	})
	if err != nil {
		t.Fatalf("create run failed: %v", err)
	}

	result, err := fx.consolidation.Run(ctx, run.ID, fx.groupOrgID)
	if err != nil {
		t.Fatalf("consolidation run failed: %v", err)
	}
	if result.Run.State != interorganization.RunStateDone {
		t.Errorf("run = %+v, want done", result.Run)
	}

	balanceByAccount := func(organizationID, accountID uint64) (float64, bool) {
		for _, balance := range result.MemberBalances {
			if balance.OrganizationID == organizationID && balance.AccountID == accountID {
				return balance.Amount, true
			}
		}
		return 0, false
	}

	for _, want := range []struct {
		organizationID, accountID uint64
		amount                    float64
	}{
		{fx.groupOrgID, fx.org1Receivable, 100},
		{fx.groupOrgID, fx.org1Income, -100},
		{fx.memberOrgID, fx.org2Expense, 100},
		{fx.memberOrgID, fx.org2Payable, -100},
	} {
		got, ok := balanceByAccount(want.organizationID, want.accountID)
		if !ok || got != want.amount {
			t.Errorf("balance org %d account %d = %v (ok %v), want %v", want.organizationID, want.accountID, got, ok, want.amount)
		}
	}

	if len(result.Eliminations) != 4 {
		t.Fatalf("eliminations = %d, want 4", len(result.Eliminations))
	}
	for _, want := range []struct {
		accountID uint64
		amount    float64
	}{
		{fx.org1Receivable, -100},
		{fx.org2Payable, 100},
		{fx.org1Income, 100},
		{fx.org2Expense, -100},
	} {
		found := false
		for _, elimination := range result.Eliminations {
			if elimination.AccountID == want.accountID && elimination.Amount == want.amount {
				found = true
			}
		}
		if !found {
			t.Errorf("missing elimination for account %d amount %v", want.accountID, want.amount)
		}
	}

	persisted, err := interorganization.NewConsolidationEliminationDAO(testDB).ListByRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("list persisted eliminations failed: %v", err)
	}
	if len(persisted) != 4 {
		t.Errorf("persisted eliminations = %d, want 4", len(persisted))
	}
}
