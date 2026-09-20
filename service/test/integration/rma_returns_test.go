//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/returns"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type rmaReturnsFixture struct {
	orgID        uint64
	journalID    uint64
	customerID   uint64
	supplierID   uint64
	variantID    uint64
	internalLoc  uint64
	customerLoc  uint64
	supplierLoc  uint64
	date         time.Time
	warehouseID  uint64
	stockValuID  uint64
	cogsID       uint64
	stockInputID uint64

	soDAO            sales.SaleOrderDAO
	soLineDAO        sales.SaleOrderLineDAO
	poDAO            procurement.PurchaseOrderDAO
	poLineDAO        procurement.PurchaseOrderLineDAO
	stockMovementDAO inventory.StockMovementDAO
	layerDAO         inventory.CostLayerDAO

	sequenceSvc      sequence.Service
	saleOrderSvc     sales.SaleOrderService
	purchaseOrderSvc procurement.PurchaseOrderService
	rmaSvc           returns.RMAService
}

func seedRMAReturnsFixture(t *testing.T) rmaReturnsFixture {
	t.Helper()
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	accounts := map[string]*reference.Account{}
	for _, spec := range []struct {
		key  string
		code string
		name string
		typ  string
	}{
		{"receivable", "1200", "Receivable", "receivable"},
		{"income", "4000", "Income", "income"},
		{"bank", "1001", "Bank", "bank"},
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
		accounts[spec.key] = account
	}

	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "RMA Returns Journal", Code: helper.Ptr("RMAJ"),
		Type: "general", DefaultAccountID: helper.Ptr(accounts["bank"].ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	customer, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(org.ID), Name: gofakeit.Name(),
	})
	if err != nil {
		t.Fatalf("create customer failed: %v", err)
	}
	supplier, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(org.ID), Name: gofakeit.Company(),
	})
	if err != nil {
		t.Fatalf("create supplier failed: %v", err)
	}
	if _, err := contacts.NewSupplierProfileDAO(testDB).Create(ctx, &contacts.SupplierProfile{
		ContactID: supplier.ID, Active: true,
	}); err != nil {
		t.Fatalf("create supplier supplier failed: %v", err)
	}

	category, err := dao.NewBase[reference.ItemCategory](testDB).Create(ctx, &reference.ItemCategory{
		Name:                    "RMA Returns Category",
		CostMethod:              helper.Ptr("standard"),
		IncomeAccountID:         helper.Ptr(accounts["income"].ID),
		StockValuationAccountID: helper.Ptr(accounts["stock_cost"].ID),
		StockInputAccountID:     helper.Ptr(accounts["stock_input"].ID),
		CogsAccountID:           helper.Ptr(accounts["cogs"].ID),
	})
	if err != nil {
		t.Fatalf("create item category failed: %v", err)
	}
	template, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID), Name: "RMA Widget", CategoryID: helper.Ptr(category.ID),
		Type: "stockable", StandardCost: 50, IsPurchasable: true, IsSellable: true, Tracking: "none",
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

	warehouse, err := inventory.NewWarehouseDAO(testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(org.ID), Name: "Main", Code: helper.Ptr("WH"),
	})
	if err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}
	stockLocationDAO := inventory.NewStockLocationDAO(testDB)
	internalLoc, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), WarehouseID: helper.Ptr(warehouse.ID), Name: "Stock", Usage: "internal",
	})
	if err != nil {
		t.Fatalf("create internal location failed: %v", err)
	}
	supplierLoc, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), Name: "Suppliers", Usage: "supplier",
	})
	if err != nil {
		t.Fatalf("create supplier location failed: %v", err)
	}
	customerLoc, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), Name: "Customers", Usage: "customer",
	})
	if err != nil {
		t.Fatalf("create customer location failed: %v", err)
	}

	for _, spec := range []struct {
		code string
	}{
		{sales.SequenceSaleOrderCode},
		{procurement.SequencePurchaseOrderCode},
		{returns.SequenceRMACode},
		{accounting.SequenceInvoiceCode},
	} {
		if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
			OrganizationID: org.ID, Code: spec.code, NextNumber: 1, Padding: 5,
		}).Error; err != nil {
			t.Fatalf("create %s sequence failed: %v", spec.code, err)
		}
	}

	date := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	year, err := dao.NewBase[reference.TaxYear](testDB).Create(ctx, &reference.TaxYear{
		OrganizationID: helper.Ptr(org.ID), Name: "2026", State: helper.Ptr("open"),
	})
	if err != nil {
		t.Fatalf("create tax year failed: %v", err)
	}
	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(testDB), dao.NewBase[reference.TaxYear](testDB))
	if _, err := taxPeriodSvc.Create(ctx, &accounting.TaxPeriod{
		OrganizationID: org.ID, TaxYearID: year.ID, Name: "Feb 2026",
		DateStart: helper.Ptr(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)),
		DateEnd:   helper.Ptr(time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)),
	}); err != nil {
		t.Fatalf("create tax period failed: %v", err)
	}

	postingSvc := accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB)).SetPeriods(taxPeriodSvc)
	sequenceSvc := sequence.NewSequenceService(sequence.NewDAO(testDB))

	stockMovementDAO := inventory.NewStockMovementDAO(testDB)
	stockCostLayerDAO := inventory.NewCostLayerDAO(testDB)
	productResolver := inventory.NewProductResolver(
		products.NewItemVariantDAO(testDB),
		products.NewItemDAO(testDB),
		dao.NewBase[reference.ItemCategory](testDB),
	)
	valuationSvc := inventory.NewValuationService(
		stockMovementDAO,
		stockCostLayerDAO,
		stockLocationDAO,
		productResolver,
		postingSvc,
		db.NewDBTransactioner(testDB),
	)

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
	paymentSvc := accounting.NewPaymentService(
		accounting.NewPaymentDAO(testDB),
		invoiceDAO,
		postingSvc,
		dao.NewBase[reference.Account](testDB),
		dao.NewBase[reference.Journal](testDB),
		sequenceSvc,
		db.NewDBTransactioner(testDB),
	)

	productSvc := products.NewProductService(
		products.NewItemDAO(testDB),
		products.NewItemVariantDAO(testDB),
		dao.NewBase[reference.ItemCategory](testDB),
		products.NewPriceBookDAO(testDB),
		products.NewPriceRuleDAO(testDB),
	)
	reservationSvc := inventory.NewHoldService(
		inventory.NewStockHoldDAO(testDB),
		inventory.NewStockBalanceDAO(testDB),
	)

	soDAO := sales.NewSaleOrderDAO(testDB)
	soLineDAO := sales.NewSaleOrderLineDAO(testDB)
	saleOrderSvc := sales.NewSaleOrderService(
		soDAO,
		soLineDAO,
		sequenceSvc,
		productSvc,
		products.NewPriceBookDAO(testDB),
		dao.NewBase[reference.Tax](testDB),
		contacts.NewContactDAO(testDB),
		crm.NewProspectDAO(testDB),
		dao.NewBase[reference.PipelineStage](testDB),
		inventory.NewWarehouseDAO(testDB),
		stockLocationDAO,
		inventory.NewStockBalanceDAO(testDB),
		inventory.NewShipmentDAO(testDB),
		stockMovementDAO,
		inventory.NewStockHoldDAO(testDB),
		reservationSvc,
		valuationSvc,
		invoiceSvc,
		invoiceDAO,
		paymentSvc,
	)

	purchaseOrderSvc := phasePurchaseOrderService(t, dao.NewBase[reference.SystemConfig](testDB), dao.NewBase[reference.Tax](testDB), productResolver)

	rmaSvc := returns.NewRMAService(
		returns.NewRMADAO(testDB),
		returns.NewRMALineDAO(testDB),
		returns.NewOriginOrderLookup(soDAO, procurement.NewPurchaseOrderDAO(testDB)),
		stockMovementDAO,
		stockLocationDAO,
		stockCostLayerDAO,
		valuationSvc,
		invoiceDAO,
		invoiceSvc,
		saleOrderSvc,
		purchaseOrderSvc,
		sequenceSvc,
		db.NewDBTransactioner(testDB),
	)

	return rmaReturnsFixture{
		orgID:            org.ID,
		journalID:        journal.ID,
		customerID:       customer.ID,
		supplierID:       supplier.ID,
		variantID:        variant.ID,
		internalLoc:      internalLoc.ID,
		customerLoc:      customerLoc.ID,
		supplierLoc:      supplierLoc.ID,
		warehouseID:      warehouse.ID,
		stockValuID:      accounts["stock_cost"].ID,
		cogsID:           accounts["cogs"].ID,
		stockInputID:     accounts["stock_input"].ID,
		date:             date,
		soDAO:            soDAO,
		soLineDAO:        soLineDAO,
		poDAO:            procurement.NewPurchaseOrderDAO(testDB),
		poLineDAO:        procurement.NewPurchaseOrderLineDAO(testDB),
		stockMovementDAO: stockMovementDAO,
		layerDAO:         stockCostLayerDAO,
		sequenceSvc:      sequenceSvc,
		saleOrderSvc:     saleOrderSvc,
		purchaseOrderSvc: purchaseOrderSvc,
		rmaSvc:           rmaSvc,
	}
}

func TestRMAServiceReceiveIntegrationRecordsSaleReturn(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedRMAReturnsFixture(t)
	ctx := testutil.SystemContext()

	order, err := fx.soDAO.CreateWithLines(ctx, &sales.SaleOrder{
		OrganizationID: helper.Ptr(fx.orgID),
		Name:           helper.Ptr("SO/00001"),
		ContactID:      fx.customerID,
		CurrencyCode:   helper.Ptr("IDR"),
		State:          sales.OrderStateConfirmed,
		OrderDate:      &fx.date,
		DeliveryStatus: sales.DeliveryStatusDone,
		InvoiceStatus:  sales.InvoiceStatusInvoiced,
		AmountUntaxed:  1000,
		AmountTax:      0,
		AmountTotal:    1000,
	}, []*sales.SaleOrderLine{
		{ItemID: helper.Ptr(fx.variantID), Description: helper.Ptr("RMA Widget"), QtyOrdered: 10, QtyDelivered: 10, QtyInvoiced: 10, UnitPrice: 100, DiscountPct: 0, PriceSubtotal: 1000},
	})
	if err != nil {
		t.Fatalf("create sale order failed: %v", err)
	}

	origin, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            10,
		SrcLocationID:  fx.internalLoc,
		DstLocationID:  fx.customerLoc,
		State:          inventory.MovementStateDone,
		OriginType:     helper.Ptr("sale_order"),
		OriginID:       &order.ID,
		ScheduledDate:  &fx.date,
	})
	if err != nil {
		t.Fatalf("create origin movement failed: %v", err)
	}
	if _, err := fx.layerDAO.Create(ctx, &inventory.CostLayer{
		MovementID:     &origin.ID,
		ItemID:         fx.variantID,
		Quantity:       -10,
		UnitCost:       helper.Ptr(50.0),
		Value:          -500,
		RemainingQty:   0,
		RemainingValue: 0,
		Description:    helper.Ptr("Goods shipment"),
	}); err != nil {
		t.Fatalf("create origin layer failed: %v", err)
	}

	rma, err := fx.rmaSvc.Create(ctx, returns.CreateRMARequest{
		OrganizationID:  fx.orgID,
		Type:            returns.TypeCustomerReturn,
		ContactID:       fx.customerID,
		OriginOrderType: returns.OriginSaleOrder,
		OriginOrderID:   order.ID,
		Reason:          "damaged",
		Lines: []returns.RMALineRequest{
			{ItemID: fx.variantID, Qty: 4, Disposition: returns.DispositionRestock},
		},
	})
	if err != nil {
		t.Fatalf("create rma failed: %v", err)
	}
	if rma.State != returns.StateDraft {
		t.Fatalf("rma state = %v, want draft", rma.State)
	}

	if _, err := fx.rmaSvc.Confirm(ctx, rma.ID); err != nil {
		t.Fatalf("confirm rma failed: %v", err)
	}
	if _, err := fx.rmaSvc.Receive(ctx, rma.ID, fx.journalID, fx.date); err != nil {
		t.Fatalf("receive rma failed: %v", err)
	}

	lines, err := fx.soLineDAO.ListByOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("list so lines failed: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("so lines = %d, want 1", len(lines))
	}
	if lines[0].QtyReturns != 4 {
		t.Errorf("so line qty_returns = %v, want 4", lines[0].QtyReturns)
	}

	updated, err := fx.soDAO.Find(ctx, order.ID)
	if err != nil {
		t.Fatalf("find so failed: %v", err)
	}
	if updated.DeliveryStatus != sales.DeliveryStatusPartial {
		t.Errorf("so delivery status = %v, want partial", updated.DeliveryStatus)
	}
	if updated.InvoiceStatus != sales.InvoiceStatusToInvoice {
		t.Errorf("so invoice status = %v, want to_invoice", updated.InvoiceStatus)
	}
}

func TestRMAServiceReceiveIntegrationRecordsVendorReturn(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedRMAReturnsFixture(t)
	ctx := testutil.SystemContext()

	po, err := fx.poDAO.CreateWithLines(ctx, &procurement.PurchaseOrder{
		OrganizationID: helper.Ptr(fx.orgID),
		Name:           helper.Ptr("PO/00001"),
		SupplierID:     fx.supplierID,
		CurrencyCode:   helper.Ptr("IDR"),
		State:          procurement.PurchaseOrderStateConfirmed,
		OrderDate:      &fx.date,
		ReceiptStatus:  procurement.PurchaseOrderReceiptStatusDone,
		InvoiceStatus:  procurement.PurchaseOrderInvoiceStatusInvoiced,
		AmountUntaxed:  500,
		AmountTax:      0,
		AmountTotal:    500,
	}, []*procurement.PurchaseOrderLine{
		{ItemID: helper.Ptr(fx.variantID), Description: helper.Ptr("RMA Widget"), QtyOrdered: 10, QtyReceived: 10, QtyBilled: 10, UnitPrice: 50, DiscountPct: 0, PriceSubtotal: 500},
	})
	if err != nil {
		t.Fatalf("create purchase order failed: %v", err)
	}

	origin, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            10,
		SrcLocationID:  fx.supplierLoc,
		DstLocationID:  fx.internalLoc,
		State:          inventory.MovementStateDone,
		OriginType:     helper.Ptr("purchase_order"),
		OriginID:       &po.ID,
		ScheduledDate:  &fx.date,
	})
	if err != nil {
		t.Fatalf("create origin movement failed: %v", err)
	}
	if _, err := fx.layerDAO.Create(ctx, &inventory.CostLayer{
		MovementID:     &origin.ID,
		ItemID:         fx.variantID,
		Quantity:       10,
		UnitCost:       helper.Ptr(50.0),
		Value:          500,
		RemainingQty:   10,
		RemainingValue: 500,
		Description:    helper.Ptr("Goods receipt"),
	}); err != nil {
		t.Fatalf("create origin layer failed: %v", err)
	}

	rma, err := fx.rmaSvc.Create(ctx, returns.CreateRMARequest{
		OrganizationID:  fx.orgID,
		Type:            returns.TypeVendorReturn,
		ContactID:       fx.supplierID,
		OriginOrderType: returns.OriginPurchaseOrder,
		OriginOrderID:   po.ID,
		Reason:          "overdelivery",
		Lines: []returns.RMALineRequest{
			{ItemID: fx.variantID, Qty: 3, Disposition: returns.DispositionRestock},
		},
	})
	if err != nil {
		t.Fatalf("create rma failed: %v", err)
	}

	if _, err := fx.rmaSvc.Confirm(ctx, rma.ID); err != nil {
		t.Fatalf("confirm rma failed: %v", err)
	}
	if _, err := fx.rmaSvc.Receive(ctx, rma.ID, fx.journalID, fx.date); err != nil {
		t.Fatalf("receive rma failed: %v", err)
	}

	lines, err := fx.poLineDAO.ListByOrder(ctx, po.ID)
	if err != nil {
		t.Fatalf("list po lines failed: %v", err)
	}
	if len(lines) != 1 {
		t.Fatalf("po lines = %d, want 1", len(lines))
	}
	if lines[0].QtyReturns != 3 {
		t.Errorf("po line qty_returns = %v, want 3", lines[0].QtyReturns)
	}

	updated, err := fx.poDAO.Find(ctx, po.ID)
	if err != nil {
		t.Fatalf("find po failed: %v", err)
	}
	if updated.ReceiptStatus != procurement.PurchaseOrderReceiptStatusPartial {
		t.Errorf("po receipt status = %v, want partial", updated.ReceiptStatus)
	}
	if updated.InvoiceStatus != procurement.PurchaseOrderInvoiceStatusToInvoice {
		t.Errorf("po invoice status = %v, want to_invoice", updated.InvoiceStatus)
	}
}
