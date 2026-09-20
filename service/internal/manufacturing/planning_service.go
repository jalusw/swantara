package manufacturing

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

type PurchasePlanner interface {
	Create(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error)
}

type ManufacturePlanner interface {
	Create(ctx context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error)
	Confirm(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error)
}

type TransferPlanner interface {
	Create(ctx context.Context, transfer *inventory.WarehouseTransfer, lines []inventory.TransferLine) (*inventory.WarehouseTransfer, error)
}

type PlanningService struct {
	runs         PlanningRunDAO
	demands      PlanningNeedDAO
	planned      PlannedSupplyDAO
	forecasts    DemandPlanDAO
	soOrders     sales.SaleOrderDAO
	soLines      sales.SaleOrderLineDAO
	poOrders     procurement.PurchaseOrderDAO
	poLines      procurement.PurchaseOrderLineDAO
	orders       ProductionOrderDAO
	reorder      inventory.ReorderService
	ledger       inventory.LedgerService
	recipes      RecipeDAO
	recipeLines  RecipeLineDAO
	recipeSvc    RecipeService
	variants     products.ItemVariantDAO
	templates    products.ItemDAO
	locations    inventory.StockLocationDAO
	purchaser    PurchasePlanner
	manufacturer ManufacturePlanner
	transfers    TransferPlanner
}

func NewPlanningService(
	runs PlanningRunDAO,
	demands PlanningNeedDAO,
	planned PlannedSupplyDAO,
	forecasts DemandPlanDAO,
	soOrders sales.SaleOrderDAO,
	soLines sales.SaleOrderLineDAO,
	poOrders procurement.PurchaseOrderDAO,
	poLines procurement.PurchaseOrderLineDAO,
	orders ProductionOrderDAO,
	reorder inventory.ReorderService,
	ledger inventory.LedgerService,
	recipes RecipeDAO,
	recipeLines RecipeLineDAO,
	recipeSvc RecipeService,
	variants products.ItemVariantDAO,
	templates products.ItemDAO,
	locations inventory.StockLocationDAO,
	purchaser PurchasePlanner,
	manufacturer ManufacturePlanner,
	transfers TransferPlanner,
) PlanningService {
	return PlanningService{
		runs:         runs,
		demands:      demands,
		planned:      planned,
		forecasts:    forecasts,
		soOrders:     soOrders,
		soLines:      soLines,
		poOrders:     poOrders,
		poLines:      poLines,
		orders:       orders,
		reorder:      reorder,
		ledger:       ledger,
		recipes:      recipes,
		recipeLines:  recipeLines,
		recipeSvc:    recipeSvc,
		variants:     variants,
		templates:    templates,
		locations:    locations,
		purchaser:    purchaser,
		manufacturer: manufacturer,
		transfers:    transfers,
	}
}

func (s PlanningService) Run(ctx context.Context, orgID uint64, horizonDays int, runDate time.Time) (*PlanningRun, error) {
	if horizonDays <= 0 {
		return nil, ErrPlanningRunNotRequired
	}
	run, err := s.runs.Create(ctx, &PlanningRun{
		OrganizationID: helper.Ptr(orgID),
		RunDate:        helper.Ptr(runDate),
		HorizonDays:    horizonDays,
		State:          PlanningRunStateRunning,
	})
	if err != nil {
		return nil, err
	}

	horizonEnd := runDate.AddDate(0, 0, horizonDays)
	demands, err := s.collectDemands(ctx, run, orgID, horizonEnd)
	if err != nil {
		return nil, err
	}

	byProduct := map[uint64][]*PlanningNeed{}
	for _, demand := range demands {
		byProduct[demand.ItemID] = append(byProduct[demand.ItemID], demand)
	}

	incoming := make(map[uint64]float64)
	if err := s.confirmedIncoming(ctx, incoming); err != nil {
		return nil, err
	}

	for itemID, productDemands := range byProduct {
		total := amount.Zero()
		for _, demand := range productDemands {
			total = total.Add(amount.FromFloat64(demand.Qty))
		}
		onHand, err := s.ledger.OnHandByProduct(ctx, itemID)
		if err != nil {
			return nil, err
		}
		available := amount.FromFloat64(onHand).Add(amount.FromFloat64(incoming[itemID]))
		net := total.Sub(available)
		if !net.IsPositive() {
			continue
		}
		pegged := productDemands[0]
		if pegged.WarehouseID != nil {
			onHandByWarehouse, err := s.ledger.OnHandByWarehouse(ctx, itemID)
			if err != nil {
				return nil, err
			}
			transferQty, srcWarehouseID := s.surplusWarehouse(onHandByWarehouse, *pegged.WarehouseID, net.Float64())
			if transferQty > 0 {
				if err := s.planTransfer(ctx, run, itemID, pegged, transferQty, srcWarehouseID, horizonEnd); err != nil {
					return nil, err
				}
				net = net.Sub(amount.FromFloat64(transferQty))
			}
		}
		if !net.IsPositive() {
			continue
		}
		if err := s.planProduct(ctx, run, itemID, net.Float64(), pegged, horizonEnd); err != nil {
			return nil, err
		}
	}

	run.State = PlanningRunStateDone
	return s.runs.Update(ctx, run)
}

func (s PlanningService) collectDemands(ctx context.Context, run *PlanningRun, orgID uint64, horizonEnd time.Time) ([]*PlanningNeed, error) {
	demands := []*PlanningNeed{}

	confirmed, err := s.soOrders.List(ctx, &query.Query{Filters: []query.Filter{{Field: "state", Operator: query.Equal, Value: sales.OrderStateConfirmed}}})
	if err != nil {
		return nil, err
	}
	for _, order := range confirmed.Items {
		if order.ExpectedDate != nil && order.ExpectedDate.After(horizonEnd) {
			continue
		}
		lines, err := s.soLines.ListByOrder(ctx, order.ID)
		if err != nil {
			return nil, err
		}
		for _, line := range lines {
			if line.ItemID == nil {
				continue
			}
			open := amount.FromFloat64(line.QtyOrdered).Sub(amount.FromFloat64(line.QtyDelivered))
			if !open.IsPositive() {
				continue
			}
			demand := &PlanningNeed{
				PlanningRunID: run.ID,
				ItemID:        *line.ItemID,
				WarehouseID:   order.WarehouseID,
				SourceType:    PlanningNeedSourceSaleOrder,
				SourceID:      line.ID,
				Qty:           open.Float64(),
				RequiredDate:  order.ExpectedDate,
			}
			demands = append(demands, demand)
		}
	}

	forecasts, err := s.forecasts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: orgID}}})
	if err != nil {
		return nil, err
	}
	for _, forecast := range forecasts.Items {
		if forecast.PeriodStart != nil && forecast.PeriodStart.After(horizonEnd) {
			continue
		}
		if forecast.PeriodEnd != nil && forecast.PeriodEnd.Before(*run.RunDate) {
			continue
		}
		if !amount.FromFloat64(forecast.ForecastQty).IsPositive() {
			continue
		}
		demand := &PlanningNeed{
			PlanningRunID: run.ID,
			ItemID:        forecast.ItemID,
			WarehouseID:   forecast.WarehouseID,
			SourceType:    PlanningNeedSourceForecast,
			SourceID:      forecast.ID,
			Qty:           forecast.ForecastQty,
			RequiredDate:  forecast.PeriodStart,
		}
		demands = append(demands, demand)
	}

	candidates, err := s.reorder.Candidates(ctx, orgID)
	if err != nil {
		return nil, err
	}
	for _, candidate := range candidates {
		if !amount.FromFloat64(candidate.RecommendedQty).IsPositive() {
			continue
		}
		demand := &PlanningNeed{
			PlanningRunID: run.ID,
			ItemID:        candidate.ItemID,
			SourceType:    PlanningNeedSourceReorder,
			SourceID:      candidate.RuleID,
			Qty:           candidate.RecommendedQty,
			RequiredDate:  helper.Ptr(time.Now().AddDate(0, 0, 7)),
		}
		demands = append(demands, demand)
	}

	for _, demand := range demands {
		if _, err := s.demands.Create(ctx, demand); err != nil {
			return nil, err
		}
	}
	return demands, nil
}

func (s PlanningService) confirmedIncoming(ctx context.Context, incoming map[uint64]float64) error {
	confirmedPOs, err := s.poOrders.List(ctx, &query.Query{Filters: []query.Filter{{Field: "state", Operator: query.Equal, Value: procurement.PurchaseOrderStateConfirmed}}})
	if err != nil {
		return err
	}
	for _, po := range confirmedPOs.Items {
		lines, err := s.poLines.ListByOrder(ctx, po.ID)
		if err != nil {
			return err
		}
		for _, line := range lines {
			if line.ItemID == nil {
				continue
			}
			open := amount.FromFloat64(line.QtyOrdered).Sub(amount.FromFloat64(line.QtyReceived))
			if open.IsPositive() {
				incoming[*line.ItemID] += open.Float64()
			}
		}
	}

	for _, state := range []string{ProductionOrderStateConfirmed, ProductionOrderStatePlanned, ProductionOrderStateInProgress} {
		orders, err := s.orders.List(ctx, &query.Query{Filters: []query.Filter{{Field: "state", Operator: query.Equal, Value: state}}})
		if err != nil {
			return err
		}
		for _, productionOrder := range orders.Items {
			open := amount.FromFloat64(productionOrder.QtyToProduce).Sub(amount.FromFloat64(productionOrder.QtyProduced))
			if open.IsPositive() {
				incoming[productionOrder.ItemID] += open.Float64()
			}
		}
	}
	return nil
}

func (s PlanningService) planProduct(ctx context.Context, run *PlanningRun, itemID uint64, qty float64, pegged *PlanningNeed, dueDate time.Time) error {
	recipe, err := s.manufactureRecipe(ctx, itemID)
	if err != nil {
		return err
	}
	orderDate := run.RunDate
	if orderDate == nil {
		now := time.Now()
		orderDate = &now
	}
	if recipe == nil {
		_, err := s.planned.Create(ctx, &PlannedSupply{
			PlanningRunID:  run.ID,
			ItemID:         itemID,
			WarehouseID:    pegged.WarehouseID,
			Type:           PlannedSupplyTypePurchase,
			Qty:            qty,
			OrderDate:      orderDate,
			DueDate:        helper.Ptr(dueDate),
			PeggedDemandID: helper.Ptr(pegged.ID),
		})
		return err
	}

	plannedOrder, err := s.planned.Create(ctx, &PlannedSupply{
		PlanningRunID:  run.ID,
		ItemID:         itemID,
		WarehouseID:    pegged.WarehouseID,
		Type:           PlannedSupplyTypeManufacture,
		Qty:            qty,
		OrderDate:      orderDate,
		DueDate:        helper.Ptr(dueDate),
		PeggedDemandID: helper.Ptr(pegged.ID),
	})
	if err != nil {
		return err
	}

	requirements, err := s.recipeSvc.Explode(ctx, recipe.ID, amount.FromFloat64(qty))
	if err != nil {
		return err
	}
	for _, requirement := range requirements {
		componentDemand := &PlanningNeed{
			PlanningRunID: run.ID,
			ItemID:        requirement.ComponentID,
			WarehouseID:   pegged.WarehouseID,
			SourceType:    PlanningNeedSourceRecipe,
			SourceID:      plannedOrder.ID,
			Qty:           requirement.Qty.Float64(),
			RequiredDate:  helper.Ptr(dueDate),
		}
		if _, err := s.demands.Create(ctx, componentDemand); err != nil {
			return err
		}
		if err := s.planProduct(ctx, run, requirement.ComponentID, requirement.Qty.Float64(), componentDemand, dueDate); err != nil {
			return err
		}
	}
	return nil
}

func (s PlanningService) planTransfer(ctx context.Context, run *PlanningRun, itemID uint64, pegged *PlanningNeed, qty float64, srcWarehouseID uint64, dueDate time.Time) error {
	orderDate := run.RunDate
	if orderDate == nil {
		now := time.Now()
		orderDate = &now
	}
	_, err := s.planned.Create(ctx, &PlannedSupply{
		PlanningRunID:  run.ID,
		ItemID:         itemID,
		WarehouseID:    pegged.WarehouseID,
		SrcWarehouseID: helper.Ptr(srcWarehouseID),
		Type:           PlannedSupplyTypeTransfer,
		Qty:            qty,
		OrderDate:      orderDate,
		DueDate:        helper.Ptr(dueDate),
		PeggedDemandID: helper.Ptr(pegged.ID),
	})
	return err
}

func (s PlanningService) surplusWarehouse(onHandByWarehouse map[uint64]float64, destWarehouseID uint64, qty float64) (float64, uint64) {
	best, bestQty := uint64(0), 0.0
	for warehouseID, onHand := range onHandByWarehouse {
		if warehouseID == destWarehouseID {
			continue
		}
		transferable := onHand
		if transferable > qty {
			transferable = qty
		}
		if transferable > bestQty {
			best, bestQty = warehouseID, transferable
		}
	}
	return bestQty, best
}

func (s PlanningService) manufactureRecipe(ctx context.Context, itemID uint64) (*Recipe, error) {
	recipes, err := s.recipes.ListByItem(ctx, itemID)
	if err != nil {
		return nil, err
	}
	for _, recipe := range recipes {
		if recipe.Active && recipe.Type == RecipeTypeManufacture {
			return recipe, nil
		}
	}
	return nil, nil
}

func (s PlanningService) Confirm(ctx context.Context, plannedOrderID, supplierID uint64) (*PlannedSupply, error) {
	plannedOrder, err := s.planned.Find(ctx, plannedOrderID)
	if err != nil {
		return nil, err
	}
	if plannedOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if plannedOrder.Confirmed {
		return nil, ErrPlanningPlannedState
	}

	switch plannedOrder.Type {
	case PlannedSupplyTypePurchase:
		return s.confirmPurchase(ctx, plannedOrder, supplierID)
	case PlannedSupplyTypeManufacture:
		return s.confirmManufacture(ctx, plannedOrder)
	case PlannedSupplyTypeTransfer:
		return s.confirmTransfer(ctx, plannedOrder)
	default:
		return nil, ErrPlanningPlannedType
	}
}

func (s PlanningService) confirmPurchase(ctx context.Context, plannedOrder *PlannedSupply, supplierID uint64) (*PlannedSupply, error) {
	if supplierID == 0 {
		return nil, ErrPlanningSupplier
	}
	run, err := s.runs.Find(ctx, plannedOrder.PlanningRunID)
	if err != nil {
		return nil, err
	}
	if run == nil {
		return nil, ErrProductionOrderNotFound
	}
	if run.OrganizationID == nil {
		return nil, ErrProductionOrderOrganization
	}
	orderDate := plannedOrder.OrderDate
	if orderDate == nil {
		now := time.Now()
		orderDate = &now
	}
	destLocationID, err := s.destinationLocation(ctx, *run.OrganizationID, plannedOrder.WarehouseID)
	if err != nil {
		return nil, err
	}
	line := &procurement.PurchaseOrderLine{
		ItemID:     helper.Ptr(plannedOrder.ItemID),
		QtyOrdered: plannedOrder.Qty,
	}
	if err := s.pricePurchaseLine(ctx, line); err != nil {
		return nil, err
	}
	po, err := s.purchaser.Create(ctx, &procurement.PurchaseOrder{
		OrganizationID: run.OrganizationID,
		SupplierID:     supplierID,
		WarehouseID:    plannedOrder.WarehouseID,
		DestLocationID: destLocationID,
		OrderDate:      orderDate,
		ExpectedDate:   plannedOrder.DueDate,
	}, []*procurement.PurchaseOrderLine{line})
	if err != nil {
		return nil, err
	}
	plannedOrder.Confirmed = true
	plannedOrder.GeneratedDocType = helper.Ptr("purchase_order")
	plannedOrder.GeneratedDocID = helper.Ptr(po.ID)
	return s.planned.Update(ctx, plannedOrder)
}

func (s PlanningService) confirmManufacture(ctx context.Context, plannedOrder *PlannedSupply) (*PlannedSupply, error) {
	recipe, err := s.manufactureRecipe(ctx, plannedOrder.ItemID)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, ErrPlanningNoRecipe
	}
	run, err := s.runs.Find(ctx, plannedOrder.PlanningRunID)
	if err != nil {
		return nil, err
	}
	if run == nil || run.OrganizationID == nil {
		return nil, ErrProductionOrderOrganization
	}
	organizationID := *run.OrganizationID
	srcLocationID, err := s.productionLocation(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	dstLocationID, err := s.destinationLocation(ctx, organizationID, plannedOrder.WarehouseID)
	if err != nil {
		return nil, err
	}
	productionOrder, err := s.manufacturer.Create(ctx, &ProductionOrder{
		OrganizationID: helper.Ptr(organizationID),
		ItemID:         plannedOrder.ItemID,
		RecipeID:       helper.Ptr(recipe.ID),
		QtyToProduce:   plannedOrder.Qty,
		SrcLocationID:  srcLocationID,
		DstLocationID:  dstLocationID,
	})
	if err != nil {
		return nil, err
	}
	if _, err := s.manufacturer.Confirm(ctx, productionOrder.ID); err != nil {
		return nil, err
	}
	plannedOrder.Confirmed = true
	plannedOrder.GeneratedDocType = helper.Ptr("production_order")
	plannedOrder.GeneratedDocID = helper.Ptr(productionOrder.ID)
	return s.planned.Update(ctx, plannedOrder)
}

func (s PlanningService) confirmTransfer(ctx context.Context, plannedOrder *PlannedSupply) (*PlannedSupply, error) {
	if plannedOrder.SrcWarehouseID == nil || plannedOrder.WarehouseID == nil {
		return nil, ErrPlanningSrcWarehouse
	}
	run, err := s.runs.Find(ctx, plannedOrder.PlanningRunID)
	if err != nil {
		return nil, err
	}
	if run == nil || run.OrganizationID == nil {
		return nil, ErrProductionOrderOrganization
	}
	transfer, err := s.transfers.Create(ctx, &inventory.WarehouseTransfer{
		OrganizationID: run.OrganizationID,
		SrcWarehouseID: *plannedOrder.SrcWarehouseID,
		DstWarehouseID: *plannedOrder.WarehouseID,
		State:          inventory.TransferStateDraft,
		ScheduledDate:  plannedOrder.DueDate,
	}, []inventory.TransferLine{
		{ItemID: plannedOrder.ItemID, Qty: plannedOrder.Qty},
	})
	if err != nil {
		return nil, err
	}
	plannedOrder.Confirmed = true
	plannedOrder.GeneratedDocType = helper.Ptr("warehouse_transfer")
	plannedOrder.GeneratedDocID = helper.Ptr(transfer.ID)
	return s.planned.Update(ctx, plannedOrder)
}

func (s PlanningService) productionLocation(ctx context.Context, organizationID uint64) (*uint64, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: "production"}}})
	if err != nil {
		return nil, err
	}
	for _, location := range locations.Items {
		if location.OrganizationID != nil && *location.OrganizationID == organizationID {
			return helper.Ptr(location.ID), nil
		}
	}
	for _, location := range locations.Items {
		if location.OrganizationID == nil {
			return helper.Ptr(location.ID), nil
		}
	}
	return nil, ErrProductionLocation
}

func (s PlanningService) destinationLocation(ctx context.Context, organizationID uint64, warehouseID *uint64) (*uint64, error) {
	if warehouseID == nil {
		return nil, ErrProductionOrderLocation
	}
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "warehouse_id", Operator: query.Equal, Value: *warehouseID}}})
	if err != nil {
		return nil, err
	}
	for _, location := range locations.Items {
		if location.Usage == "internal" && location.OrganizationID != nil && *location.OrganizationID == organizationID {
			return helper.Ptr(location.ID), nil
		}
	}
	return nil, ErrProductionOrderLocation
}

func (s PlanningService) pricePurchaseLine(ctx context.Context, line *procurement.PurchaseOrderLine) error {
	variant, err := s.variants.Find(ctx, *line.ItemID)
	if err != nil {
		return err
	}
	if variant == nil {
		return ErrPlanningItem
	}
	template, err := s.templates.Find(ctx, variant.ItemID)
	if err != nil {
		return err
	}
	if template == nil {
		return ErrPlanningItem
	}
	line.UnitPrice = template.StandardCost
	return nil
}

func (s PlanningService) ListRuns(ctx context.Context, q *query.Query) (*query.Page[PlanningRun], error) {
	return s.runs.List(ctx, q)
}

func (s PlanningService) FindRun(ctx context.Context, id uint64) (*PlanningRun, error) {
	return s.runs.Find(ctx, id)
}

func (s PlanningService) ListDemands(ctx context.Context, runID uint64) ([]*PlanningNeed, error) {
	return s.demands.ListByRun(ctx, runID)
}

func (s PlanningService) ListPlanned(ctx context.Context, runID uint64) ([]*PlannedSupply, error) {
	return s.planned.ListByRun(ctx, runID)
}

func (s PlanningService) CreateForecast(ctx context.Context, forecast *DemandPlan) (*DemandPlan, error) {
	return s.forecasts.Create(ctx, forecast)
}

func (s PlanningService) ListForecasts(ctx context.Context, q *query.Query) (*query.Page[DemandPlan], error) {
	return s.forecasts.List(ctx, q)
}
