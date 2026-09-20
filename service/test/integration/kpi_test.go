//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestKpis_DerivableFromLiveData(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()
	kpiSvc := reporting.NewKpiService(
		fx.reportDAO,
		nil, nil, nil, nil,
	)

	date := fx.date
	if _, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID, JournalID: fx.journalID, Date: date, Ref: "KPI-001",
		Description: "revenue",
		Lines: []accounting.PostingLine{
			{AccountID: fx.bankID, Debit: amount.FromFloat64(1000)},
			{AccountID: fx.incomeID, Credit: amount.FromFloat64(1000)},
		},
	}); err != nil {
		t.Fatalf("post revenue failed: %v", err)
	}

	cash, err := kpiSvc.CashKPI(ctx, fx.orgID, date)
	if err != nil {
		t.Fatalf("cash kpi failed: %v", err)
	}
	if cash.Position != 1000 || cash.Forecast != 1000 {
		t.Errorf("cash = %+v, want position 1000 forecast 1000", cash)
	}

	arAp, err := kpiSvc.ArApKPI(ctx, fx.orgID, date)
	if err != nil {
		t.Fatalf("ar/ap kpi failed: %v", err)
	}
	if arAp.DSO != 0 || arAp.DPO != 0 {
		t.Errorf("ar/ap = %+v, want DSO/DPO 0 with no invoices", arAp)
	}

	contact, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: &fx.orgID, Name: "Supplier Co", Active: true,
	})
	if err != nil {
		t.Fatalf("create contact failed: %v", err)
	}

	expected := date.AddDate(0, 0, 7)
	po, err := procurement.NewPurchaseOrderDAO(testDB).Create(ctx, &procurement.PurchaseOrder{
		OrganizationID: &fx.orgID, SupplierID: contact.ID, Name: &[]string{"PO-001"}[0],
		CurrencyCode: &[]string{"USD"}[0], State: "confirmed",
		OrderDate: &date, ExpectedDate: &expected, AmountUntaxed: 500, AmountTotal: 500,
		InvoiceStatus: "no", ReceiptStatus: "done",
	})
	if err != nil {
		t.Fatalf("create purchase order failed: %v", err)
	}
	if _, err := procurement.NewPurchaseOrderLineDAO(testDB).Create(ctx, &procurement.PurchaseOrderLine{
		OrderID: po.ID, ItemID: &fx.variantID, QtyOrdered: 10, QtyReceived: 10,
		UnitPrice: 50, PriceSubtotal: 500,
	}); err != nil {
		t.Fatalf("create purchase order line failed: %v", err)
	}
	if _, err := products.NewSupplierProductDAO(testDB).Create(ctx, &products.SupplierProduct{
		ItemID: fx.variantID, SupplierID: contact.ID, Price: &[]float64{40}[0],
		CurrencyCode: &[]string{"USD"}[0], Priority: 10,
	}); err != nil {
		t.Fatalf("create supplier item failed: %v", err)
	}
	if _, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: &fx.orgID, ItemID: fx.variantID, Qty: 10,
		SrcLocationID: fx.supplierLocID, DstLocationID: fx.internalLocID,
		State: "done", OriginType: &[]string{"purchase_order"}[0], OriginID: &po.ID,
		ScheduledDate: &expected, DateDone: &date,
	}); err != nil {
		t.Fatalf("create receipt stock movement failed: %v", err)
	}

	proc, err := kpiSvc.ProcurementKPI(ctx, fx.orgID, time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC), date)
	if err != nil {
		t.Fatalf("procurement kpi failed: %v", err)
	}
	if proc.PurchaseCount != 1 {
		t.Errorf("purchase count = %d, want 1", proc.PurchaseCount)
	}
	if proc.OnTimeDeliveryPct != 1 {
		t.Errorf("on-time = %v, want 1", proc.OnTimeDeliveryPct)
	}
	if proc.PriceVariancePct != 0.25 {
		t.Errorf("price variance = %v, want 0.25", proc.PriceVariancePct)
	}

	if _, err := manufacturing.NewProductionOrderDAO(testDB).Create(ctx, &manufacturing.ProductionOrder{
		OrganizationID: &fx.orgID, Name: &[]string{"MO-001"}[0], ItemID: fx.variantID,
		QtyToProduce: 10, QtyProduced: 9, State: "done",
		DatePlannedStart: &date, DatePlannedFinish: &date, DateStart: &date, DateFinished: &date,
	}); err != nil {
		t.Fatalf("create manufacturing order failed: %v", err)
	}
	moPage, err := manufacturing.NewProductionOrderDAO(testDB).List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: fx.orgID}},
	})
	if err != nil {
		t.Fatalf("list manufacturing orders failed: %v", err)
	}
	if len(moPage.Items) != 1 {
		t.Fatalf("expected 1 manufacturing order, got %d", len(moPage.Items))
	}
	mo := moPage.Items[0]
	workCenter, err := dao.NewBase[reference.WorkCenter](testDB).Create(ctx, &reference.WorkCenter{
		OrganizationID: &fx.orgID, Name: "Assembly Line", Code: &[]string{"WC-001"}[0],
	})
	if err != nil {
		t.Fatalf("create work center failed: %v", err)
	}
	if _, err := manufacturing.NewShopTaskDAO(testDB).Create(ctx, &manufacturing.ShopTask{
		OrganizationID: &fx.orgID, ProductionOrderID: mo.ID, WorkCenterID: workCenter.ID,
		Name: &[]string{"WO-001"}[0], State: "done", PlannedStart: &date, PlannedFinish: &date,
		DateStart: &date, DateFinished: &date, PlannedMinutes: 100, ActualMinutes: 80,
	}); err != nil {
		t.Fatalf("create work order failed: %v", err)
	}

	man, err := kpiSvc.ManufacturingKPI(ctx, fx.orgID, time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC), date)
	if err != nil {
		t.Fatalf("manufacturing kpi failed: %v", err)
	}
	if man.OrderCount != 1 || man.OEEPct != 1.25 {
		t.Errorf("manufacturing = %+v, want 1 order, OEE 1.25", man)
	}
	if man.YieldPct != 0.9 {
		t.Errorf("yield = %v, want 0.9", man.YieldPct)
	}

	reorderDAO := inventory.NewReorderRuleDAO(testDB)
	rule, err := reorderDAO.Create(ctx, &inventory.ReorderRule{
		ItemID: fx.variantID, MinQty: 5, MaxQty: 50, QtyMultiple: 1, Active: true,
	})
	if err != nil {
		t.Fatalf("create reorder rule failed: %v", err)
	}
	_ = rule
	quantDAO := inventory.NewStockBalanceDAO(testDB)
	if _, err := quantDAO.Create(ctx, &inventory.StockBalance{
		OrganizationID: &fx.orgID, ItemID: fx.variantID, LocationID: fx.internalLocID, Quantity: 0,
	}); err != nil {
		t.Fatalf("create stock quant failed: %v", err)
	}

	inv, err := kpiSvc.InventoryRatioKPI(ctx, fx.orgID, time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC), date)
	if err != nil {
		t.Fatalf("inventory ratio kpi failed: %v", err)
	}
	if inv.StockoutCount != 1 {
		t.Errorf("stockout count = %d, want 1", inv.StockoutCount)
	}
}
