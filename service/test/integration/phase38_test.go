//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type phase38Fixture struct {
	orgID         uint64
	warehouseID   uint64
	internalLocID uint64
	componentID   uint64
	finishedID    uint64
	supplierID    uint64

	mrpSvc      manufacturing.PlanningService
	runDAO      manufacturing.PlanningRunDAO
	demandDAO   manufacturing.PlanningNeedDAO
	plannedDAO  manufacturing.PlannedSupplyDAO
	forecastDAO manufacturing.DemandPlanDAO
	soDAO       sales.SaleOrderDAO
	poDAO       procurement.PurchaseOrderDAO
	poLineDAO   procurement.PurchaseOrderLineDAO
	moDAO       manufacturing.ProductionOrderDAO
}

func seedPhase38Fixture(t *testing.T) phase38Fixture {
	t.Helper()
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	supplier, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: helper.Ptr(org.ID), Name: gofakeit.Company(),
	})
	if err != nil {
		t.Fatalf("create supplier contact failed: %v", err)
	}
	if _, err := contacts.NewSupplierProfileDAO(testDB).Create(ctx, &contacts.SupplierProfile{
		ContactID: supplier.ID, Active: true,
	}); err != nil {
		t.Fatalf("create supplier supplier failed: %v", err)
	}

	warehouse, err := dao.NewBase[reference.Warehouse](testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(org.ID), Name: "Main",
	})
	if err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}
	internalLoc, err := dao.NewBase[reference.StockLocation](testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), WarehouseID: helper.Ptr(warehouse.ID),
		Name: "Stock", Usage: "internal",
	})
	if err != nil {
		t.Fatalf("create internal location failed: %v", err)
	}
	if _, err := dao.NewBase[reference.StockLocation](testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), Name: "Production", Usage: "production",
	}); err != nil {
		t.Fatalf("create production location failed: %v", err)
	}

	componentTemplate, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID), Name: "Component", CategoryID: nil,
		Type: "stockable", StandardCost: 50, IsPurchasable: true, IsSellable: true, Tracking: "none",
	})
	if err != nil {
		t.Fatalf("create component template failed: %v", err)
	}
	component, err := products.NewItemVariantDAO(testDB).Create(ctx, &products.ItemVariant{
		ItemID: componentTemplate.ID, Active: true,
	})
	if err != nil {
		t.Fatalf("create component variant failed: %v", err)
	}
	finishedTemplate, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID), Name: "Finished", CategoryID: nil,
		Type: "stockable", StandardCost: 200, IsPurchasable: true, IsSellable: true, Tracking: "none",
	})
	if err != nil {
		t.Fatalf("create finished template failed: %v", err)
	}
	finished, err := products.NewItemVariantDAO(testDB).Create(ctx, &products.ItemVariant{
		ItemID: finishedTemplate.ID, Active: true,
	})
	if err != nil {
		t.Fatalf("create finished variant failed: %v", err)
	}

	bomSvc := manufacturing.NewRecipeService(
		products.NewItemVariantDAO(testDB),
		manufacturing.NewRecipeDAO(testDB),
		manufacturing.NewRecipeLineDAO(testDB),
	)
	if _, err := bomSvc.CreateRecipe(ctx, &manufacturing.Recipe{
		OrganizationID: helper.Ptr(org.ID), ItemID: finished.ID,
		Type: manufacturing.RecipeTypeManufacture, Qty: 1,
	}, []*manufacturing.RecipeLine{
		{ComponentID: component.ID, Qty: 2},
	}); err != nil {
		t.Fatalf("create recipe failed: %v", err)
	}

	if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
		OrganizationID: org.ID, Code: procurement.SequencePurchaseOrderCode, NextNumber: 1, Padding: 5,
	}).Error; err != nil {
		t.Fatalf("create purchase order sequence failed: %v", err)
	}
	if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
		OrganizationID: org.ID, Code: manufacturing.SequenceProductionOrderCode, NextNumber: 1, Padding: 5,
	}).Error; err != nil {
		t.Fatalf("create manufacturing order sequence failed: %v", err)
	}

	productResolver := inventory.NewProductResolver(
		products.NewItemVariantDAO(testDB),
		products.NewItemDAO(testDB),
		dao.NewBase[reference.ItemCategory](testDB),
	)
	purchaseOrderSvc := phasePurchaseOrderService(t, dao.NewBase[reference.SystemConfig](testDB), dao.NewBase[reference.Tax](testDB), productResolver)

	ledgerSvc := inventory.NewLedgerService(
		inventory.NewStockMovementDAO(testDB),
		inventory.NewStockBalanceDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		db.NewDBTransactioner(testDB),
	)
	reorderSvc := inventory.NewReorderService(inventory.NewReorderRuleDAO(testDB), ledgerSvc, productResolver)
	reservationSvc := inventory.NewHoldService(
		inventory.NewStockHoldDAO(testDB),
		inventory.NewStockBalanceDAO(testDB),
	)

	moDAO := manufacturing.NewProductionOrderDAO(testDB)
	consumedMaterialDAO := manufacturing.NewConsumedMaterialDAO(testDB)
	recipeDAO := manufacturing.NewRecipeDAO(testDB)
	recipeLineDAO := manufacturing.NewRecipeLineDAO(testDB)
	moSvc := manufacturing.NewProductionOrderService(
		moDAO,
		consumedMaterialDAO,
		recipeDAO,
		bomSvc,
		products.NewItemVariantDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		reservationSvc,
		sequence.NewSequenceService(sequence.NewDAO(testDB)),
	)
	transferSvc := inventory.NewTransferService(
		inventory.NewWarehouseTransferDAO(testDB),
		inventory.NewStockMovementDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		inventory.NewWarehouseDAO(testDB),
		inventory.NewCostLayerDAO(testDB),
		ledgerSvc,
		productResolver,
		inventory.PosterMock{},
		db.NewDBTransactioner(testDB),
	)

	mrpSvc := manufacturing.NewPlanningService(
		manufacturing.NewPlanningRunDAO(testDB),
		manufacturing.NewPlanningNeedDAO(testDB),
		manufacturing.NewPlannedSupplyDAO(testDB),
		manufacturing.NewDemandPlanDAO(testDB),
		sales.NewSaleOrderDAO(testDB),
		sales.NewSaleOrderLineDAO(testDB),
		procurement.NewPurchaseOrderDAO(testDB),
		procurement.NewPurchaseOrderLineDAO(testDB),
		moDAO,
		reorderSvc,
		ledgerSvc,
		recipeDAO,
		recipeLineDAO,
		bomSvc,
		products.NewItemVariantDAO(testDB),
		products.NewItemDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		purchaseOrderSvc,
		moSvc,
		transferSvc,
	)

	return phase38Fixture{
		orgID:         org.ID,
		warehouseID:   warehouse.ID,
		internalLocID: internalLoc.ID,
		componentID:   component.ID,
		finishedID:    finished.ID,
		supplierID:    supplier.ID,
		mrpSvc:        mrpSvc,
		runDAO:        manufacturing.NewPlanningRunDAO(testDB),
		demandDAO:     manufacturing.NewPlanningNeedDAO(testDB),
		plannedDAO:    manufacturing.NewPlannedSupplyDAO(testDB),
		forecastDAO:   manufacturing.NewDemandPlanDAO(testDB),
		soDAO:         sales.NewSaleOrderDAO(testDB),
		poDAO:         procurement.NewPurchaseOrderDAO(testDB),
		poLineDAO:     procurement.NewPurchaseOrderLineDAO(testDB),
		moDAO:         moDAO,
	}
}

func phase38PlannedByProduct(t *testing.T, fx phase38Fixture, runID, itemID uint64) []*manufacturing.PlannedSupply {
	t.Helper()
	planned, err := fx.plannedDAO.ListByRun(testutil.SystemContext(), runID)
	if err != nil {
		t.Fatalf("list planned orders failed: %v", err)
	}
	var matches []*manufacturing.PlannedSupply
	for _, order := range planned {
		if order.ItemID == itemID {
			matches = append(matches, order)
		}
	}
	return matches
}

func TestMrpForecastNettingAndConfirmIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase38Fixture(t)
	ctx := testutil.SystemContext()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	if _, err := inventory.NewStockBalanceDAO(testDB).Create(ctx, &inventory.StockBalance{
		OrganizationID: helper.Ptr(fx.orgID), ItemID: fx.componentID, LocationID: fx.internalLocID, Quantity: 4,
	}); err != nil {
		t.Fatalf("create on-hand quant failed: %v", err)
	}
	forecast, err := fx.forecastDAO.Create(ctx, &manufacturing.DemandPlan{
		OrganizationID: helper.Ptr(fx.orgID), ItemID: fx.componentID, WarehouseID: helper.Ptr(fx.warehouseID),
		PeriodStart: helper.Ptr(time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)),
		PeriodEnd:   helper.Ptr(time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)),
		ForecastQty: 10,
	})
	if err != nil {
		t.Fatalf("create forecast failed: %v", err)
	}

	run, err := fx.mrpSvc.Run(ctx, fx.orgID, 30, runDate)
	if err != nil {
		t.Fatalf("planning run failed: %v", err)
	}
	if run.State != manufacturing.PlanningRunStateDone {
		t.Errorf("run = %+v, want done", run)
	}

	demands, err := fx.demandDAO.ListByRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("list demands failed: %v", err)
	}
	if len(demands) != 1 || demands[0].SourceType != manufacturing.PlanningNeedSourceForecast || demands[0].SourceID != forecast.ID || demands[0].Qty != 10 {
		t.Errorf("demands = %+v, want forecast demand 10", demands)
	}

	planned := phase38PlannedByProduct(t, fx, run.ID, fx.componentID)
	if len(planned) != 1 {
		t.Fatalf("planned orders for component = %d, want 1", len(planned))
	}
	if planned[0].Type != manufacturing.PlannedSupplyTypePurchase || planned[0].Qty != 6 {
		t.Errorf("planned = %+v, want purchase qty 6 (10 - 4 on hand)", planned[0])
	}
	if planned[0].PeggedDemandID == nil || *planned[0].PeggedDemandID != demands[0].ID {
		t.Errorf("planned pegged demand = %v, want %v", planned[0].PeggedDemandID, demands[0].ID)
	}

	confirmed, err := fx.mrpSvc.Confirm(ctx, planned[0].ID, fx.supplierID)
	if err != nil {
		t.Fatalf("confirm planned order failed: %v", err)
	}
	if !confirmed.Confirmed || confirmed.GeneratedDocType == nil || *confirmed.GeneratedDocType != "purchase_order" || confirmed.GeneratedDocID == nil {
		t.Errorf("planned = %+v, want confirmed purchase order", confirmed)
	}
	po, err := fx.poDAO.Find(ctx, *confirmed.GeneratedDocID)
	if err != nil {
		t.Fatalf("find generated po failed: %v", err)
	}
	if po == nil || po.SupplierID != fx.supplierID || po.State != procurement.PurchaseOrderStateDraft {
		t.Errorf("po = %+v, want draft for supplier %v", po, fx.supplierID)
	}
	poLines, err := fx.poLineDAO.ListByOrder(ctx, po.ID)
	if err != nil {
		t.Fatalf("list po lines failed: %v", err)
	}
	if len(poLines) != 1 || poLines[0].QtyOrdered != 6 || poLines[0].UnitPrice != 50 {
		t.Errorf("po lines = %+v, want 6 x 50", poLines)
	}
}

func TestMrpNettingAgainstIncomingIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase38Fixture(t)
	ctx := testutil.SystemContext()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expectedDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	if _, err := inventory.NewStockBalanceDAO(testDB).Create(ctx, &inventory.StockBalance{
		OrganizationID: helper.Ptr(fx.orgID), ItemID: fx.componentID, LocationID: fx.internalLocID, Quantity: 4,
	}); err != nil {
		t.Fatalf("create on-hand quant failed: %v", err)
	}
	if _, err := fx.poDAO.CreateWithLines(ctx, &procurement.PurchaseOrder{
		OrganizationID: helper.Ptr(fx.orgID), SupplierID: fx.supplierID, WarehouseID: helper.Ptr(fx.warehouseID),
		CurrencyCode: helper.Ptr("IDR"), State: procurement.PurchaseOrderStateConfirmed,
		InvoiceStatus: "no", ReceiptStatus: "pending", OrderDate: &runDate,
	}, []*procurement.PurchaseOrderLine{
		{ItemID: helper.Ptr(fx.componentID), QtyOrdered: 3, QtyReceived: 0, UnitPrice: 50},
	}); err != nil {
		t.Fatalf("create confirmed po failed: %v", err)
	}

	order, err := fx.soDAO.CreateWithLines(ctx, &sales.SaleOrder{
		OrganizationID: helper.Ptr(fx.orgID), Name: helper.Ptr("SO/Planning"),
		ContactID: fx.supplierID, CurrencyCode: helper.Ptr("IDR"),
		WarehouseID: helper.Ptr(fx.warehouseID), State: sales.OrderStateConfirmed,
		OrderDate: &runDate, ExpectedDate: &expectedDate,
		DeliveryStatus: sales.DeliveryStatusPending, InvoiceStatus: "no",
		AmountUntaxed: 500, AmountTotal: 500,
	}, []*sales.SaleOrderLine{
		{ItemID: helper.Ptr(fx.componentID), QtyOrdered: 10, QtyDelivered: 0, UnitPrice: 50, PriceSubtotal: 500},
	})
	if err != nil {
		t.Fatalf("create sale order failed: %v", err)
	}

	run, err := fx.mrpSvc.Run(ctx, fx.orgID, 30, runDate)
	if err != nil {
		t.Fatalf("planning run failed: %v", err)
	}

	demands, err := fx.demandDAO.ListByRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("list demands failed: %v", err)
	}
	if len(demands) != 1 || demands[0].SourceType != manufacturing.PlanningNeedSourceSaleOrder || demands[0].SourceID != order.ID {
		t.Errorf("demands = %+v, want sale order demand", demands)
	}

	planned := phase38PlannedByProduct(t, fx, run.ID, fx.componentID)
	if len(planned) != 1 || planned[0].Type != manufacturing.PlannedSupplyTypePurchase || planned[0].Qty != 3 {
		t.Errorf("planned = %+v, want purchase qty 3 (10 - 4 on hand - 3 incoming)", planned)
	}
	if planned[0].PeggedDemandID == nil || *planned[0].PeggedDemandID != demands[0].ID {
		t.Errorf("planned pegged demand = %v, want %v", planned[0].PeggedDemandID, demands[0].ID)
	}
}

func TestMrpRecipeExplosionAndConfirmManufactureIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase38Fixture(t)
	ctx := testutil.SystemContext()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expectedDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	order, err := fx.soDAO.CreateWithLines(ctx, &sales.SaleOrder{
		OrganizationID: helper.Ptr(fx.orgID), Name: helper.Ptr("SO/BOM"),
		ContactID: fx.supplierID, CurrencyCode: helper.Ptr("IDR"),
		WarehouseID: helper.Ptr(fx.warehouseID), State: sales.OrderStateConfirmed,
		OrderDate: &runDate, ExpectedDate: &expectedDate,
		DeliveryStatus: sales.DeliveryStatusPending, InvoiceStatus: "no",
		AmountUntaxed: 1000, AmountTotal: 1000,
	}, []*sales.SaleOrderLine{
		{ItemID: helper.Ptr(fx.finishedID), QtyOrdered: 5, QtyDelivered: 0, UnitPrice: 200, PriceSubtotal: 1000},
	})
	if err != nil {
		t.Fatalf("create sale order failed: %v", err)
	}

	run, err := fx.mrpSvc.Run(ctx, fx.orgID, 30, runDate)
	if err != nil {
		t.Fatalf("planning run failed: %v", err)
	}

	demands, err := fx.demandDAO.ListByRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("list demands failed: %v", err)
	}
	if len(demands) != 2 {
		t.Fatalf("demands = %d, want 2 (sale + recipe)", len(demands))
	}

	finishedPlanned := phase38PlannedByProduct(t, fx, run.ID, fx.finishedID)
	if len(finishedPlanned) != 1 || finishedPlanned[0].Type != manufacturing.PlannedSupplyTypeManufacture || finishedPlanned[0].Qty != 5 {
		t.Fatalf("finished planned = %+v, want manufacture qty 5", finishedPlanned)
	}
	componentPlanned := phase38PlannedByProduct(t, fx, run.ID, fx.componentID)
	if len(componentPlanned) != 1 || componentPlanned[0].Type != manufacturing.PlannedSupplyTypePurchase || componentPlanned[0].Qty != 10 {
		t.Fatalf("component planned = %+v, want purchase qty 10 (5 x recipe 2)", componentPlanned)
	}

	confirmedMO, err := fx.mrpSvc.Confirm(ctx, finishedPlanned[0].ID, fx.supplierID)
	if err != nil {
		t.Fatalf("confirm manufacture failed: %v", err)
	}
	if !confirmedMO.Confirmed || confirmedMO.GeneratedDocType == nil || *confirmedMO.GeneratedDocType != "production_order" || confirmedMO.GeneratedDocID == nil {
		t.Errorf("planned = %+v, want confirmed manufacturing order", confirmedMO)
	}
	mo, err := fx.moDAO.Find(ctx, *confirmedMO.GeneratedDocID)
	if err != nil {
		t.Fatalf("find generated mo failed: %v", err)
	}
	if mo == nil || mo.ItemID != fx.finishedID || mo.QtyToProduce != 5 || mo.State != manufacturing.ProductionOrderStateConfirmed {
		t.Errorf("mo = %+v, want confirmed mo for finished qty 5", mo)
	}

	confirmedPO, err := fx.mrpSvc.Confirm(ctx, componentPlanned[0].ID, fx.supplierID)
	if err != nil {
		t.Fatalf("confirm component purchase failed: %v", err)
	}
	if confirmedPO.GeneratedDocType == nil || *confirmedPO.GeneratedDocType != "purchase_order" || confirmedPO.GeneratedDocID == nil {
		t.Errorf("planned = %+v, want confirmed purchase order", confirmedPO)
	}
	if !confirmedPO.Confirmed {
		t.Errorf("planned confirmed = false, want true")
	}
	if order.ID == 0 {
		t.Errorf("sale order id = 0")
	}
}

func TestMrpTransferPlanningAndConfirmIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedPhase38Fixture(t)
	ctx := testutil.SystemContext()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expectedDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	warehouseB, err := dao.NewBase[reference.Warehouse](testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(fx.orgID), Name: "Branch",
	})
	if err != nil {
		t.Fatalf("create warehouse b failed: %v", err)
	}
	internalLocB, err := dao.NewBase[reference.StockLocation](testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(fx.orgID), WarehouseID: helper.Ptr(warehouseB.ID),
		Name: "Stock B", Usage: "internal",
	})
	if err != nil {
		t.Fatalf("create warehouse b location failed: %v", err)
	}
	if _, err := dao.NewBase[reference.StockLocation](testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(fx.orgID), Name: "Transit", Usage: "transit",
	}); err != nil {
		t.Fatalf("create transit location failed: %v", err)
	}

	if _, err := inventory.NewStockBalanceDAO(testDB).Create(ctx, &inventory.StockBalance{
		OrganizationID: helper.Ptr(fx.orgID), ItemID: fx.componentID, LocationID: fx.internalLocID, Quantity: 4,
	}); err != nil {
		t.Fatalf("create warehouse a quant failed: %v", err)
	}

	order, err := fx.soDAO.CreateWithLines(ctx, &sales.SaleOrder{
		OrganizationID: helper.Ptr(fx.orgID), Name: helper.Ptr("SO/TRF"),
		ContactID: fx.supplierID, CurrencyCode: helper.Ptr("IDR"),
		WarehouseID: helper.Ptr(warehouseB.ID), State: sales.OrderStateConfirmed,
		OrderDate: &runDate, ExpectedDate: &expectedDate,
		DeliveryStatus: sales.DeliveryStatusPending, InvoiceStatus: "no",
		AmountUntaxed: 500, AmountTotal: 500,
	}, []*sales.SaleOrderLine{
		{ItemID: helper.Ptr(fx.componentID), QtyOrdered: 10, QtyDelivered: 0, UnitPrice: 50, PriceSubtotal: 500},
	})
	if err != nil {
		t.Fatalf("create sale order failed: %v", err)
	}

	run, err := fx.mrpSvc.Run(ctx, fx.orgID, 30, runDate)
	if err != nil {
		t.Fatalf("planning run failed: %v", err)
	}

	demands, err := fx.demandDAO.ListByRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("list demands failed: %v", err)
	}
	if len(demands) != 1 || demands[0].SourceType != manufacturing.PlanningNeedSourceSaleOrder || demands[0].SourceID != order.ID {
		t.Fatalf("demands = %+v, want sale order demand", demands)
	}

	planned := phase38PlannedByProduct(t, fx, run.ID, fx.componentID)
	if len(planned) != 2 {
		t.Fatalf("planned orders for component = %d, want 2 (transfer + purchase)", len(planned))
	}
	var transferPlanned *manufacturing.PlannedSupply
	for _, candidate := range planned {
		if candidate.Type == manufacturing.PlannedSupplyTypeTransfer {
			transferPlanned = candidate
		}
	}
	if transferPlanned == nil {
		t.Fatalf("expected transfer planned order, got %+v", planned)
	}
	if transferPlanned.Qty != 4 {
		t.Errorf("transfer planned qty = %v, want 4 (surplus in warehouse a)", transferPlanned.Qty)
	}
	if transferPlanned.SrcWarehouseID == nil || *transferPlanned.SrcWarehouseID != fx.warehouseID {
		t.Errorf("transfer planned src warehouse = %v, want %v", transferPlanned.SrcWarehouseID, fx.warehouseID)
	}
	if transferPlanned.WarehouseID == nil || *transferPlanned.WarehouseID != warehouseB.ID {
		t.Errorf("transfer planned dst warehouse = %v, want %v", transferPlanned.WarehouseID, warehouseB.ID)
	}
	if transferPlanned.PeggedDemandID == nil || *transferPlanned.PeggedDemandID != demands[0].ID {
		t.Errorf("transfer planned pegged demand = %v, want %v", transferPlanned.PeggedDemandID, demands[0].ID)
	}
	var purchasePlanned *manufacturing.PlannedSupply
	for _, candidate := range planned {
		if candidate.Type == manufacturing.PlannedSupplyTypePurchase {
			purchasePlanned = candidate
		}
	}
	if purchasePlanned == nil || purchasePlanned.Qty != 2 {
		t.Errorf("purchase planned = %+v, want purchase qty 2 (10 - 4 on hand - 4 transfer)", purchasePlanned)
	}

	confirmed, err := fx.mrpSvc.Confirm(ctx, transferPlanned.ID, fx.supplierID)
	if err != nil {
		t.Fatalf("confirm transfer failed: %v", err)
	}
	if !confirmed.Confirmed || confirmed.GeneratedDocType == nil || *confirmed.GeneratedDocType != "warehouse_transfer" || confirmed.GeneratedDocID == nil {
		t.Errorf("planned = %+v, want confirmed transfer order", confirmed)
	}
	transfer, err := inventory.NewWarehouseTransferDAO(testDB).Find(ctx, *confirmed.GeneratedDocID)
	if err != nil {
		t.Fatalf("find generated transfer failed: %v", err)
	}
	if transfer == nil || transfer.SrcWarehouseID != fx.warehouseID || transfer.DstWarehouseID != warehouseB.ID {
		t.Errorf("transfer = %+v, want %v -> %v", transfer, fx.warehouseID, warehouseB.ID)
	}
	if transfer == nil || transfer.State != inventory.TransferStateDraft {
		t.Errorf("transfer state = %+v, want draft", transfer)
	}
	if _, err := inventory.NewStockMovementDAO(testDB).ListByShipment(ctx, *transfer.OutShipmentID); err != nil {
		t.Fatalf("expected out movements for transfer: %v", err)
	}
	if internalLocB.ID == 0 {
		t.Errorf("warehouse b location id = 0")
	}
}
