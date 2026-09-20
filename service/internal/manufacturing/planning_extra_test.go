package manufacturing

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

func TestPlanningService_CollectDemands_CollectsForecastAndReorder(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	periodStart := time.Date(2026, 8, 5, 0, 0, 0, 0, time.UTC)
	run := &PlanningRun{Base: model.Base{ID: 1}, RunDate: helper.Ptr(runDate), HorizonDays: 30}
	horizonEnd := runDate.AddDate(0, 0, 30)

	forecasts := DemandPlanDAOMock{
		CRUDMock: dao.CRUDMock[DemandPlan]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[DemandPlan], error) {
				return &query.Page[DemandPlan]{Items: []*DemandPlan{
					{Base: model.Base{ID: 1}, ItemID: 100, ForecastQty: 50, PeriodStart: helper.Ptr(periodStart), PeriodEnd: helper.Ptr(periodStart.AddDate(0, 0, 30))},
				}}, nil
			},
		},
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(7)), Usage: "internal"}, nil
			},
		},
	}
	quants := inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(7)), Quantity: 5}, nil
		},
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*inventory.StockBalance, error) {
			return []*inventory.StockBalance{}, nil
		},
	}
	ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, quants, locations, inventory.TransactionerMock{})
	rules := inventory.ReorderRuleDAOMock{
		ListActiveFunc: func(_ context.Context) ([]*inventory.ReorderRule, error) {
			return []*inventory.ReorderRule{{Base: model.Base{ID: 2}, ItemID: 200, LocationID: helper.Ptr(uint64(10)), MinQty: 20, MaxQty: 100, QtyMultiple: 10}}, nil
		},
	}
	reorder := inventory.NewReorderService(rules, ledger, inventory.ItemResolverMock{})

	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, forecasts, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, reorder, ledger, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	demands, err := svc.collectDemands(ctx, run, 7, horizonEnd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(demands) != 2 {
		t.Fatalf("demands = %d, want 2 (forecast + reorder)", len(demands))
	}
	if demands[0].SourceType != PlanningNeedSourceForecast || demands[0].Qty != 50 {
		t.Errorf("demand[0] = %+v, want forecast demand of 50", demands[0])
	}
	if demands[1].SourceType != PlanningNeedSourceReorder || demands[1].Qty != 100 {
		t.Errorf("demand[1] = %+v, want reorder demand of 100", demands[1])
	}
}

func TestPlanningService_CollectDemands_SkipsOutOfScope(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	run := &PlanningRun{Base: model.Base{ID: 1}, RunDate: helper.Ptr(runDate), HorizonDays: 30}
	horizonEnd := runDate.AddDate(0, 0, 30)
	afterHorizon := runDate.AddDate(0, 0, 45)

	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[sales.SaleOrder], error) {
				return &query.Page[sales.SaleOrder]{Items: []*sales.SaleOrder{{Base: model.Base{ID: 1}, ExpectedDate: helper.Ptr(afterHorizon)}}}, nil
			},
		},
	}
	lines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{
				{Base: model.Base{ID: 11}, ItemID: nil, QtyOrdered: 10, QtyDelivered: 0},
				{Base: model.Base{ID: 12}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 10, QtyDelivered: 10},
			}, nil
		},
	}
	forecasts := DemandPlanDAOMock{
		CRUDMock: dao.CRUDMock[DemandPlan]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[DemandPlan], error) {
				return &query.Page[DemandPlan]{Items: []*DemandPlan{
					{Base: model.Base{ID: 1}, ItemID: 100, ForecastQty: 50, PeriodStart: helper.Ptr(afterHorizon)},
					{Base: model.Base{ID: 2}, ItemID: 100, ForecastQty: 50, PeriodEnd: helper.Ptr(runDate.AddDate(0, 0, -1))},
					{Base: model.Base{ID: 3}, ItemID: 100, ForecastQty: 0, PeriodStart: helper.Ptr(runDate.AddDate(0, 0, 5))},
				}}, nil
			},
		},
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(7))}, nil
			},
		},
	}
	quants := inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(7)), Quantity: 5}, nil
		},
	}
	ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, quants, locations, inventory.TransactionerMock{})
	rules := inventory.ReorderRuleDAOMock{
		ListActiveFunc: func(_ context.Context) ([]*inventory.ReorderRule, error) {
			return []*inventory.ReorderRule{{Base: model.Base{ID: 2}, ItemID: 200, LocationID: helper.Ptr(uint64(10)), MinQty: 10, MaxQty: 5, QtyMultiple: 1}}, nil
		},
	}
	reorder := inventory.NewReorderService(rules, ledger, inventory.ItemResolverMock{})

	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, forecasts, orders, lines, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, reorder, ledger, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	demands, err := svc.collectDemands(ctx, run, 7, horizonEnd)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(demands) != 0 {
		t.Errorf("demands = %d, want 0 (all out of scope)", len(demands))
	}
}

func TestPlanningService_CollectDemands_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	run := &PlanningRun{Base: model.Base{ID: 1}, RunDate: helper.Ptr(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)), HorizonDays: 30}
	horizonEnd := run.RunDate.AddDate(0, 0, 30)
	orders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[sales.SaleOrder], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, DemandPlanDAOMock{}, orders, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.collectDemands(ctx, run, 7, horizonEnd)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPlanningService_ProductionLocation_FallsBackToGlobalLocation(t *testing.T) {
	ctx := context.Background()
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
					{Base: model.Base{ID: 51}, OrganizationID: helper.Ptr(uint64(7)), Usage: "production"},
					{Base: model.Base{ID: 52}, Usage: "production"},
				}}, nil
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, locations, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	matched, err := svc.productionLocation(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matched == nil || *matched != 51 {
		t.Errorf("location = %v, want 51 for matching organization", matched)
	}

	global, err := svc.productionLocation(ctx, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if global == nil || *global != 52 {
		t.Errorf("location = %v, want global 52", global)
	}
}

func TestPlanningService_ProductionLocation_ReturnsErrorWhenMissing(t *testing.T) {
	ctx := context.Background()
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
					{Base: model.Base{ID: 51}, OrganizationID: helper.Ptr(uint64(7)), Usage: "production"},
				}}, nil
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, locations, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.productionLocation(ctx, 8)

	if helper.AssertError(t, err, true, ErrProductionLocation) {
		return
	}
}

func TestPlanningService_ProductionLocation_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, locations, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.productionLocation(ctx, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestPlanningService_DestinationLocation_RequiresWarehouse(t *testing.T) {
	ctx := context.Background()
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.destinationLocation(ctx, 7, nil)

	if helper.AssertError(t, err, true, ErrProductionOrderLocation) {
		return
	}
}

func TestPlanningService_PricePurchaseLine_RequiresProduct(t *testing.T) {
	ctx := context.Background()
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, PlannedSupplyDAOMock{}, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	err := svc.pricePurchaseLine(ctx, &procurement.PurchaseOrderLine{ItemID: helper.Ptr(uint64(100))})

	if helper.AssertError(t, err, true, ErrPlanningItem) {
		return
	}
}

func TestPlanningService_Confirm_RejectsUnknownType(t *testing.T) {
	ctx := context.Background()
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*PlannedSupply, error) {
				return &PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 100, Type: "service"}, nil
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.Confirm(ctx, 1, 0)

	if helper.AssertError(t, err, true, ErrPlanningPlannedType) {
		return
	}
}

func TestPlanningService_Confirm_ManufactureRequiresRecipe(t *testing.T) {
	ctx := context.Background()
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*PlannedSupply, error) {
				return &PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 200, Type: PlannedSupplyTypeManufacture, Qty: 4}, nil
			},
		},
	}
	recipes := RecipeDAOMock{
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*Recipe, error) {
			return []*Recipe{}, nil
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, recipes, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.Confirm(ctx, 1, 0)

	if helper.AssertError(t, err, true, ErrPlanningNoRecipe) {
		return
	}
}
