package manufacturing

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

type PurchasePlannerMock struct {
	CreateFunc func(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error)
}

func (m PurchasePlannerMock) Create(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order, lines)
	}
	return order, nil
}

type ManufacturePlannerMock struct {
	CreateFunc  func(ctx context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error)
	ConfirmFunc func(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error)
}

func (m ManufacturePlannerMock) Create(ctx context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, productionOrder)
	}
	return productionOrder, nil
}

func (m ManufacturePlannerMock) Confirm(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error) {
	if m.ConfirmFunc != nil {
		return m.ConfirmFunc(ctx, productionOrderID)
	}
	return &ProductionOrder{Base: model.Base{ID: productionOrderID}, State: ProductionOrderStateConfirmed}, nil
}

type TransferPlannerMock struct {
	CreateFunc func(ctx context.Context, transfer *inventory.WarehouseTransfer, lines []inventory.TransferLine) (*inventory.WarehouseTransfer, error)
}

func (m TransferPlannerMock) Create(ctx context.Context, transfer *inventory.WarehouseTransfer, lines []inventory.TransferLine) (*inventory.WarehouseTransfer, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, transfer, lines)
	}
	return transfer, nil
}

type PurchaseOrderCreatorMock struct {
	CreateFunc func(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error)
}

func (m PurchaseOrderCreatorMock) Create(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order, lines)
	}
	return order, nil
}

func NewTestProductionOrderService(
	orders ProductionOrderDAOMock,
	components ConsumedMaterialDAOMock,
	recipes RecipeDAOMock,
	lines RecipeLineDAOMock,
	variants products.ItemVariantDAOMock,
	locations inventory.StockLocationDAOMock,
	reservations inventory.StockHoldDAOMock,
	quants inventory.StockBalanceDAOMock,
	sequences sequence.DAOMock,
) ProductionOrderService {
	recipeSvc := NewRecipeService(variants, recipes, lines)
	return NewProductionOrderService(orders, components, recipes, recipeSvc, variants, locations, inventory.NewHoldService(reservations, quants), sequence.NewSequenceService(sequences))
}

func NewTestProductionService(
	orders ProductionOrderDAOMock,
	components ConsumedMaterialDAOMock,
	workOrders ShopTaskDAOMock,
	operations ProductionStepDAOMock,
	workCenters dao.CRUDMock[reference.WorkCenter],
	locations inventory.StockLocationDAOMock,
	movements inventory.StockMovementDAOMock,
	layers inventory.CostLayerDAOMock,
	resolver inventory.ItemResolverMock,
	poster inventory.PosterMock,
	lines accounting.JournalLineDAOMock,
	sequences sequence.DAOMock,
) ProductionService {
	ledger := inventory.NewLedgerService(movements, inventory.StockBalanceDAOMock{}, locations, inventory.TransactionerMock{})
	valuation := inventory.NewValuationService(movements, layers, locations, resolver, poster, inventory.TransactionerMock{})
	return NewProductionService(orders, components, workOrders, operations, workCenters, locations, ledger, valuation, resolver, poster, lines, sequence.NewSequenceService(sequences))
}

func NewTestPlanningService(
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
	recipeSvc := NewRecipeService(variants, recipes, recipeLines)
	return NewPlanningService(runs, demands, planned, forecasts, soOrders, soLines, poOrders, poLines, orders, reorder, ledger, recipes, recipeLines, recipeSvc, variants, templates, locations, purchaser, manufacturer, transfers)
}

func NewTestOutsideProcessingService(
	outsideOrders OutsideProcessingOrderDAOMock,
	productionOrders ProductionOrderDAOMock,
	components ConsumedMaterialDAOMock,
	recipes RecipeDAOMock,
	purchases PurchaseOrderCreator,
	poDAO procurement.PurchaseOrderDAOMock,
	poLines procurement.PurchaseOrderLineDAOMock,
	locations inventory.StockLocationDAOMock,
	movements inventory.StockMovementDAOMock,
	layers inventory.CostLayerDAOMock,
	resolver inventory.ItemResolverMock,
	poster inventory.PosterMock,
) OutsideProcessingService {
	ledger := inventory.NewLedgerService(movements, inventory.StockBalanceDAOMock{}, locations, inventory.TransactionerMock{})
	valuation := inventory.NewValuationService(movements, layers, locations, resolver, poster, inventory.TransactionerMock{})
	return NewOutsideProcessingService(outsideOrders, productionOrders, components, recipes, purchases, poDAO, poLines, locations, ledger, valuation, resolver, poster)
}

func DefaultOrderSequenceMock() sequence.DAOMock {
	return sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "SEQ/00001"}, nil
		},
	}
}

func DefaultShopTaskSequenceMock() sequence.DAOMock {
	return sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "WO/00001"}, nil
		},
	}
}
