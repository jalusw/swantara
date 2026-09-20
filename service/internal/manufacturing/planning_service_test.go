package manufacturing

import (
	"context"
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

func testPlanningService(
	runs PlanningRunDAOMock,
	demands PlanningNeedDAOMock,
	planned PlannedSupplyDAOMock,
	forecasts DemandPlanDAOMock,
	soOrders sales.SaleOrderDAOMock,
	soLines sales.SaleOrderLineDAOMock,
	poOrders procurement.PurchaseOrderDAOMock,
	poLines procurement.PurchaseOrderLineDAOMock,
	orders ProductionOrderDAOMock,
	reorder inventory.ReorderService,
	ledger inventory.LedgerService,
	recipes RecipeDAOMock,
	recipeLines RecipeLineDAOMock,
	variants products.ItemVariantDAOMock,
	templates products.ItemDAOMock,
	locations inventory.StockLocationDAOMock,
	purchaser PurchasePlanner,
	manufacturer ManufacturePlanner,
	transfers TransferPlanner,
) PlanningService {
	return NewTestPlanningService(runs, demands, planned, forecasts, soOrders, soLines, poOrders, poLines, orders, reorder, ledger, recipes, recipeLines, variants, templates, locations, purchaser, manufacturer, transfers)
}

func TestPlanningService_Run_NetsAgainstOnHandAndIncoming(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expectedDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[sales.SaleOrder], error) {
				return &query.Page[sales.SaleOrder]{Items: []*sales.SaleOrder{{Base: model.Base{ID: 1}, State: sales.OrderStateConfirmed, ExpectedDate: helper.Ptr(expectedDate)}}}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(ctx context.Context, orderID uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 10, QtyDelivered: 0}}, nil
		},
	}
	demands := PlanningNeedDAOMock{
		CRUDMock: dao.CRUDMock[PlanningNeed]{
			CreateFunc: func(ctx context.Context, demand *PlanningNeed) (*PlanningNeed, error) {
				if demand.ID == 0 {
					demand.ID = 1
				}
				return demand, nil
			},
		},
	}
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			CreateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				if order.ID == 0 {
					order.ID = 1
				}
				return order, nil
			},
			UpdateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				return order, nil
			},
		},
	}
	ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{
		ListByItemFunc: func(ctx context.Context, itemID uint64) ([]*inventory.StockBalance, error) {
			return []*inventory.StockBalance{{Quantity: 4}}, nil
		},
	}, inventory.StockLocationDAOMock{}, inventory.TransactionerMock{})
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[procurement.PurchaseOrder], error) {
				return &query.Page[procurement.PurchaseOrder]{Items: []*procurement.PurchaseOrder{{Base: model.Base{ID: 5}, State: procurement.PurchaseOrderStateConfirmed}}}, nil
			},
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(ctx context.Context, orderID uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 3, QtyReceived: 0}}, nil
		},
	}
	productionOrders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[ProductionOrder], error) {
				if len(q.Filters) == 0 || q.Filters[0].Field != "state" || q.Filters[0].Value != ProductionOrderStateConfirmed {
					return &query.Page[ProductionOrder]{Items: []*ProductionOrder{}}, nil
				}
				return &query.Page[ProductionOrder]{Items: []*ProductionOrder{{ItemID: 100, QtyToProduce: 1, QtyProduced: 0, State: ProductionOrderStateConfirmed}}}, nil
			},
		},
	}
	reorder := inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, ledger, inventory.ItemResolverMock{})

	created := []*PlannedSupply{}
	plannedOrders.CreateFunc = func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
		if order.ID == 0 {
			order.ID = 1
		}
		created = append(created, order)
		return order, nil
	}

	svc := testPlanningService(PlanningRunDAOMock{}, demands, plannedOrders, DemandPlanDAOMock{}, soOrders, soLines, poOrders, poLines, productionOrders, reorder, ledger, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	run, err := svc.Run(ctx, 7, 30, runDate)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if run.State != PlanningRunStateDone {
		t.Fatalf("expected run state %q, got %q", PlanningRunStateDone, run.State)
	}
	if len(created) != 1 {
		t.Fatalf("expected 1 planned order, got %d", len(created))
	}
	order := created[0]
	if order.Type != PlannedSupplyTypePurchase {
		t.Fatalf("expected purchase planned order, got %q", order.Type)
	}
	if order.Qty != 2 {
		t.Fatalf("expected net quantity 2 (10 - 4 on-hand - 3 po - 1 productionOrder), got %v", order.Qty)
	}
}

func TestPlanningService_Run_ExplodesRecipeIntoManufactureAndComponents(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expectedDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[sales.SaleOrder], error) {
				return &query.Page[sales.SaleOrder]{Items: []*sales.SaleOrder{{Base: model.Base{ID: 1}, State: sales.OrderStateConfirmed, ExpectedDate: helper.Ptr(expectedDate)}}}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(ctx context.Context, orderID uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(200)), QtyOrdered: 10, QtyDelivered: 0}}, nil
		},
	}
	recipes := RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			FindFunc: func(ctx context.Context, id uint64) (*Recipe, error) {
				return &Recipe{Base: model.Base{ID: 1}, ItemID: 200, Type: RecipeTypeManufacture, Qty: 1, Active: true}, nil
			},
		},
		ListByItemFunc: func(ctx context.Context, itemID uint64) ([]*Recipe, error) {
			if itemID != 200 {
				return []*Recipe{}, nil
			}
			return []*Recipe{{Base: model.Base{ID: 1}, ItemID: 200, Type: RecipeTypeManufacture, Qty: 1, Active: true}}, nil
		},
	}
	recipeLines := RecipeLineDAOMock{
		ListByRecipeFunc: func(ctx context.Context, recipeID uint64) ([]*RecipeLine, error) {
			return []*RecipeLine{{ComponentID: 201, Qty: 2}}, nil
		},
	}
	demands := PlanningNeedDAOMock{
		CRUDMock: dao.CRUDMock[PlanningNeed]{
			CreateFunc: func(ctx context.Context, demand *PlanningNeed) (*PlanningNeed, error) {
				if demand.ID == 0 {
					demand.ID = 1
				}
				return demand, nil
			},
		},
	}
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			CreateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				if order.ID == 0 {
					order.ID = 1
				}
				return order, nil
			},
			UpdateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				return order, nil
			},
		},
	}
	ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.TransactionerMock{})
	reorder := inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, ledger, inventory.ItemResolverMock{})

	created := []*PlannedSupply{}
	plannedOrders.CreateFunc = func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
		if order.ID == 0 {
			order.ID = 1
		}
		created = append(created, order)
		return order, nil
	}

	svc := testPlanningService(PlanningRunDAOMock{}, demands, plannedOrders, DemandPlanDAOMock{}, soOrders, soLines, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, reorder, ledger, recipes, recipeLines, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.Run(ctx, 7, 30, runDate)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 planned orders (manufacture + component purchase), got %d", len(created))
	}
	if created[0].Type != PlannedSupplyTypeManufacture {
		t.Fatalf("expected manufacture planned order first, got %q", created[0].Type)
	}
	if created[1].Type != PlannedSupplyTypePurchase {
		t.Fatalf("expected component purchase planned order, got %q", created[1].Type)
	}
	if created[1].Qty != 20 {
		t.Fatalf("expected component quantity 20 (10 x 2), got %v", created[1].Qty)
	}
}

func TestPlanningService_Run_PegsPlannedOrdersToDemands(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expectedDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[sales.SaleOrder], error) {
				return &query.Page[sales.SaleOrder]{Items: []*sales.SaleOrder{{Base: model.Base{ID: 1}, State: sales.OrderStateConfirmed, ExpectedDate: helper.Ptr(expectedDate)}}}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(ctx context.Context, orderID uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 10, QtyDelivered: 0}}, nil
		},
	}
	demands := PlanningNeedDAOMock{
		CRUDMock: dao.CRUDMock[PlanningNeed]{
			CreateFunc: func(ctx context.Context, demand *PlanningNeed) (*PlanningNeed, error) {
				if demand.ID == 0 {
					demand.ID = 1
				}
				return demand, nil
			},
		},
	}
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			CreateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				if order.ID == 0 {
					order.ID = 1
				}
				return order, nil
			},
			UpdateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				return order, nil
			},
		},
	}
	ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.TransactionerMock{})
	reorder := inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, ledger, inventory.ItemResolverMock{})

	created := []*PlannedSupply{}
	plannedOrders.CreateFunc = func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
		if order.ID == 0 {
			order.ID = 1
		}
		created = append(created, order)
		return order, nil
	}

	svc := testPlanningService(PlanningRunDAOMock{}, demands, plannedOrders, DemandPlanDAOMock{}, soOrders, soLines, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, reorder, ledger, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.Run(ctx, 7, 30, runDate)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(created) != 1 {
		t.Fatalf("expected 1 planned order, got %d", len(created))
	}
	if created[0].PeggedDemandID == nil || *created[0].PeggedDemandID != 1 {
		t.Fatalf("expected planned order pegged to demand 1, got %v", created[0].PeggedDemandID)
	}
}

func TestPlanningService_Confirm_PurchaseCreatesPurchaseOrder(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(ctx context.Context, id uint64) (*PlannedSupply, error) {
				return &PlannedSupply{
					Base:          model.Base{ID: 1},
					PlanningRunID: 1,
					ItemID:        100,
					Type:          PlannedSupplyTypePurchase,
					Qty:           5,
					OrderDate:     helper.Ptr(runDate),
					WarehouseID:   helper.Ptr(uint64(9)),
				}, nil
			},
			UpdateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				return order, nil
			},
		},
	}
	runs := PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[PlanningRun]{
			FindFunc: func(ctx context.Context, id uint64) (*PlanningRun, error) {
				return &PlanningRun{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(7)), RunDate: helper.Ptr(runDate)}, nil
			},
		},
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "warehouse_id" && q.Filters[0].Value == uint64(9) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 50}, OrganizationID: helper.Ptr(uint64(7)), Usage: "internal"}}}, nil
				}
				if q.Filters[0].Field == "usage" && q.Filters[0].Value == "production" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 51}, OrganizationID: helper.Ptr(uint64(7)), Usage: "production"}}}, nil
				}
				return &query.Page[reference.StockLocation]{}, nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(ctx context.Context, id uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 100}, ItemID: 2}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(ctx context.Context, id uint64) (*products.Item, error) {
				return &products.Item{Base: model.Base{ID: 2}, StandardCost: 25}, nil
			},
		},
	}
	var createdOrder *procurement.PurchaseOrder
	var createdLines []*procurement.PurchaseOrderLine
	purchaser := PurchasePlannerMock{
		CreateFunc: func(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			createdOrder = order
			createdLines = lines
			return &procurement.PurchaseOrder{Base: model.Base{ID: 77}, SupplierID: order.SupplierID}, nil
		},
	}

	svc := testPlanningService(runs, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, variants, templates, locations, purchaser, ManufacturePlannerMock{}, TransferPlannerMock{})

	confirmed, err := svc.Confirm(ctx, 1, 44)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if !confirmed.Confirmed {
		t.Fatal("expected planned order to be confirmed")
	}
	if confirmed.GeneratedDocType == nil || *confirmed.GeneratedDocType != "purchase_order" {
		t.Fatalf("expected generated doc type purchase_order, got %v", confirmed.GeneratedDocType)
	}
	if confirmed.GeneratedDocID == nil || *confirmed.GeneratedDocID != 77 {
		t.Fatalf("expected generated doc id 77, got %v", confirmed.GeneratedDocID)
	}
	if createdOrder == nil {
		t.Fatal("expected purchase order to be created")
	}
	if createdOrder.SupplierID != 44 {
		t.Fatalf("expected supplier 44, got %d", createdOrder.SupplierID)
	}
	if len(createdLines) != 1 || createdLines[0].UnitPrice != 25 {
		t.Fatalf("expected purchase line priced at standard cost 25, got %v", createdLines)
	}
}

func TestPlanningService_Confirm_PurchaseRequiresVendor(t *testing.T) {
	ctx := context.Background()
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(ctx context.Context, id uint64) (*PlannedSupply, error) {
				return &PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 100, Type: PlannedSupplyTypePurchase, Qty: 5}, nil
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.Confirm(ctx, 1, 0)
	helper.AssertError(t, err, true, ErrPlanningSupplier)
}

func TestPlanningService_Confirm_ManufactureCreatesMO(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(ctx context.Context, id uint64) (*PlannedSupply, error) {
				return &PlannedSupply{
					Base:          model.Base{ID: 1},
					PlanningRunID: 1,
					ItemID:        200,
					Type:          PlannedSupplyTypeManufacture,
					Qty:           4,
					OrderDate:     helper.Ptr(runDate),
					WarehouseID:   helper.Ptr(uint64(9)),
				}, nil
			},
			UpdateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				return order, nil
			},
		},
	}
	runs := PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[PlanningRun]{
			FindFunc: func(ctx context.Context, id uint64) (*PlanningRun, error) {
				return &PlanningRun{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(7))}, nil
			},
		},
	}
	recipes := RecipeDAOMock{
		ListByItemFunc: func(ctx context.Context, itemID uint64) ([]*Recipe, error) {
			return []*Recipe{{Base: model.Base{ID: 3}, ItemID: 200, Type: RecipeTypeManufacture, Active: true}}, nil
		},
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(ctx context.Context, id uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: id}, Usage: "production"}, nil
			},
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field == "warehouse_id" && q.Filters[0].Value == uint64(9) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 50}, OrganizationID: helper.Ptr(uint64(7)), Usage: "internal"}}}, nil
				}
				if q.Filters[0].Field == "usage" && q.Filters[0].Value == "production" {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 51}, OrganizationID: helper.Ptr(uint64(7)), Usage: "production"}}}, nil
				}
				return &query.Page[reference.StockLocation]{}, nil
			},
		},
	}
	var createdMO *ProductionOrder
	manufacturer := ManufacturePlannerMock{
		CreateFunc: func(ctx context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
			createdMO = productionOrder
			productionOrder.ID = 55
			productionOrder.State = ProductionOrderStateDraft
			return productionOrder, nil
		},
		ConfirmFunc: func(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error) {
			return &ProductionOrder{Base: model.Base{ID: productionOrderID}, State: ProductionOrderStateConfirmed}, nil
		},
	}

	svc := testPlanningService(runs, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, recipes, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, locations, PurchasePlannerMock{}, manufacturer, TransferPlannerMock{})

	confirmed, err := svc.Confirm(ctx, 1, 0)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if !confirmed.Confirmed {
		t.Fatal("expected planned order to be confirmed")
	}
	if confirmed.GeneratedDocType == nil || *confirmed.GeneratedDocType != "production_order" {
		t.Fatalf("expected generated doc type manufacturing_order, got %v", confirmed.GeneratedDocType)
	}
	if confirmed.GeneratedDocID == nil || *confirmed.GeneratedDocID != 55 {
		t.Fatalf("expected generated doc id 55, got %v", confirmed.GeneratedDocID)
	}
	if createdMO == nil || createdMO.QtyToProduce != 4 {
		t.Fatalf("expected MO qty 4, got %+v", createdMO)
	}
	if createdMO == nil || createdMO.RecipeID == nil || *createdMO.RecipeID != 3 {
		t.Fatalf("expected MO recipe 3, got %+v", createdMO)
	}
}

func TestPlanningService_Confirm_RejectsConfirmedPlannedOrder(t *testing.T) {
	ctx := context.Background()
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(ctx context.Context, id uint64) (*PlannedSupply, error) {
				return &PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 100, Type: PlannedSupplyTypePurchase, Qty: 5, Confirmed: true}, nil
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.Confirm(ctx, 1, 0)
	helper.AssertError(t, err, true, ErrPlanningPlannedState)
}

func TestPlanningService_Run_PlansTransferFromSurplusWarehouse(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	expectedDate := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			ListFunc: func(ctx context.Context, q *query.Query) (*query.Page[sales.SaleOrder], error) {
				return &query.Page[sales.SaleOrder]{Items: []*sales.SaleOrder{{Base: model.Base{ID: 1}, State: sales.OrderStateConfirmed, WarehouseID: helper.Ptr(uint64(9)), ExpectedDate: helper.Ptr(expectedDate)}}}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(ctx context.Context, orderID uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 11}, ItemID: helper.Ptr(uint64(100)), QtyOrdered: 10, QtyDelivered: 0}}, nil
		},
	}
	demands := PlanningNeedDAOMock{
		CRUDMock: dao.CRUDMock[PlanningNeed]{
			CreateFunc: func(ctx context.Context, demand *PlanningNeed) (*PlanningNeed, error) {
				if demand.ID == 0 {
					demand.ID = 1
				}
				return demand, nil
			},
		},
	}
	ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{
		ListByItemFunc: func(ctx context.Context, itemID uint64) ([]*inventory.StockBalance, error) {
			return []*inventory.StockBalance{{Quantity: 8, LocationID: 5}}, nil
		},
	}, inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(ctx context.Context, id uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: id}, WarehouseID: helper.Ptr(uint64(8))}, nil
			},
		},
	}, inventory.TransactionerMock{})
	reorder := inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, ledger, inventory.ItemResolverMock{})

	created := []*PlannedSupply{}
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			CreateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				if order.ID == 0 {
					order.ID = 1
				}
				created = append(created, order)
				return order, nil
			},
		},
	}

	svc := testPlanningService(PlanningRunDAOMock{}, demands, plannedOrders, DemandPlanDAOMock{}, soOrders, soLines, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, reorder, ledger, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	run, err := svc.Run(ctx, 7, 30, runDate)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if run.State != PlanningRunStateDone {
		t.Fatalf("expected run state %q, got %q", PlanningRunStateDone, run.State)
	}
	if len(created) != 1 {
		t.Fatalf("expected 1 planned order, got %d", len(created))
	}
	order := created[0]
	if order.Type != PlannedSupplyTypeTransfer {
		t.Fatalf("expected transfer planned order, got %q", order.Type)
	}
	if order.Qty != 2 {
		t.Fatalf("expected transfer qty 2, got %v", order.Qty)
	}
	if order.SrcWarehouseID == nil || *order.SrcWarehouseID != 8 {
		t.Fatalf("expected src warehouse 8, got %v", order.SrcWarehouseID)
	}
	if order.WarehouseID == nil || *order.WarehouseID != 9 {
		t.Fatalf("expected dst warehouse 9, got %v", order.WarehouseID)
	}
	if order.PeggedDemandID == nil {
		t.Fatalf("expected pegged demand, got %v", order.PeggedDemandID)
	}
}

func TestPlanningService_Confirm_TransferCreatesWarehouseTransfer(t *testing.T) {
	ctx := context.Background()
	runDate := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(ctx context.Context, id uint64) (*PlannedSupply, error) {
				return &PlannedSupply{
					Base:           model.Base{ID: 1},
					PlanningRunID:  1,
					ItemID:         100,
					Type:           PlannedSupplyTypeTransfer,
					Qty:            5,
					WarehouseID:    helper.Ptr(uint64(9)),
					SrcWarehouseID: helper.Ptr(uint64(8)),
					DueDate:        helper.Ptr(runDate),
				}, nil
			},
			UpdateFunc: func(ctx context.Context, order *PlannedSupply) (*PlannedSupply, error) {
				return order, nil
			},
		},
	}
	runs := PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[PlanningRun]{
			FindFunc: func(ctx context.Context, id uint64) (*PlanningRun, error) {
				return &PlanningRun{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(7))}, nil
			},
		},
	}
	var createdTransfer *inventory.WarehouseTransfer
	var createdLines []inventory.TransferLine
	transfers := TransferPlannerMock{
		CreateFunc: func(ctx context.Context, transfer *inventory.WarehouseTransfer, lines []inventory.TransferLine) (*inventory.WarehouseTransfer, error) {
			createdTransfer = transfer
			createdLines = lines
			return &inventory.WarehouseTransfer{Base: model.Base{ID: 66}, SrcWarehouseID: transfer.SrcWarehouseID, DstWarehouseID: transfer.DstWarehouseID}, nil
		},
	}

	svc := testPlanningService(runs, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, transfers)

	confirmed, err := svc.Confirm(ctx, 1, 0)
	if err != nil {
		t.Fatalf("Confirm returned error: %v", err)
	}
	if !confirmed.Confirmed {
		t.Fatal("expected planned order to be confirmed")
	}
	if confirmed.GeneratedDocType == nil || *confirmed.GeneratedDocType != "warehouse_transfer" {
		t.Fatalf("expected generated doc type warehouse_transfer, got %v", confirmed.GeneratedDocType)
	}
	if confirmed.GeneratedDocID == nil || *confirmed.GeneratedDocID != 66 {
		t.Fatalf("expected generated doc id 66, got %v", confirmed.GeneratedDocID)
	}
	if createdTransfer == nil || createdTransfer.SrcWarehouseID != 8 || createdTransfer.DstWarehouseID != 9 {
		t.Fatalf("expected transfer 8 -> 9, got %+v", createdTransfer)
	}
	if len(createdLines) != 1 || createdLines[0].ItemID != 100 || createdLines[0].Qty != 5 {
		t.Fatalf("expected transfer line item 100 qty 5, got %v", createdLines)
	}
}

func TestPlanningService_Confirm_TransferRequiresSourceWarehouse(t *testing.T) {
	ctx := context.Background()
	plannedOrders := PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[PlannedSupply]{
			FindFunc: func(ctx context.Context, id uint64) (*PlannedSupply, error) {
				return &PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 100, Type: PlannedSupplyTypeTransfer, Qty: 5}, nil
			},
		},
	}
	svc := testPlanningService(PlanningRunDAOMock{}, PlanningNeedDAOMock{}, plannedOrders, DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, ProductionOrderDAOMock{}, inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, inventory.LedgerService{}, inventory.ItemResolverMock{}), inventory.LedgerService{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{}, PurchasePlannerMock{}, ManufacturePlannerMock{}, TransferPlannerMock{})

	_, err := svc.Confirm(ctx, 1, 0)
	helper.AssertError(t, err, true, ErrPlanningSrcWarehouse)
}
