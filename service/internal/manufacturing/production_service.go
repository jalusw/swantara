package manufacturing

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ProductionService struct {
	orders      ProductionOrderDAO
	components  ConsumedMaterialDAO
	workOrders  ShopTaskDAO
	operations  ProductionStepDAO
	workCenters dao.CRUD[reference.WorkCenter]
	locations   inventory.StockLocationDAO
	ledger      inventory.LedgerService
	valuation   inventory.ValuationService
	resolver    inventory.ItemResolver
	poster      accounting.Poster
	lines       accounting.JournalLineDAO
	sequences   sequence.Service
}

func NewProductionService(
	orders ProductionOrderDAO,
	components ConsumedMaterialDAO,
	workOrders ShopTaskDAO,
	operations ProductionStepDAO,
	workCenters dao.CRUD[reference.WorkCenter],
	locations inventory.StockLocationDAO,
	ledger inventory.LedgerService,
	valuation inventory.ValuationService,
	resolver inventory.ItemResolver,
	poster accounting.Poster,
	lines accounting.JournalLineDAO,
	sequences sequence.Service,
) ProductionService {
	return ProductionService{
		orders:      orders,
		components:  components,
		workOrders:  workOrders,
		operations:  operations,
		workCenters: workCenters,
		locations:   locations,
		ledger:      ledger,
		valuation:   valuation,
		resolver:    resolver,
		poster:      poster,
		lines:       lines,
		sequences:   sequences,
	}
}

func (s ProductionService) Start(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error) {
	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStatePlanned {
		return nil, ErrProductionOrderState
	}
	now := time.Now().UTC()
	productionOrder.State = ProductionOrderStateInProgress
	productionOrder.DateStart = &now
	return s.orders.Update(ctx, productionOrder)
}

func (s ProductionService) GenerateShopTasks(ctx context.Context, productionOrderID uint64) ([]*ShopTask, error) {
	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStatePlanned && productionOrder.State != ProductionOrderStateInProgress {
		return nil, ErrProductionOrderState
	}
	if productionOrder.RecipeID == nil {
		return nil, ErrProductionOrderRecipe
	}
	if productionOrder.OrganizationID == nil {
		return nil, ErrProductionOrderOrganization
	}

	operations, err := s.operations.ListByRecipe(ctx, *productionOrder.RecipeID)
	if err != nil {
		return nil, err
	}
	created := make([]*ShopTask, 0, len(operations))
	for _, operation := range operations {
		if operation.WorkCenterID == nil {
			return nil, ErrShopTaskWorkCenter
		}
		workCenter, err := s.workCenters.Find(ctx, *operation.WorkCenterID)
		if err != nil {
			return nil, err
		}
		if workCenter == nil {
			return nil, ErrShopTaskWorkCenter
		}

		name, err := s.sequences.Next(ctx, *productionOrder.OrganizationID, SequenceShopTaskCode)
		if err != nil {
			return nil, err
		}

		plannedMinutes := operation.SetupMinutes + operation.TimeMinutes
		workOrder := &ShopTask{
			OrganizationID:    productionOrder.OrganizationID,
			ProductionOrderID: productionOrder.ID,
			ProductionStepID:  &operation.ID,
			WorkCenterID:      *operation.WorkCenterID,
			Name:              helper.Ptr(name),
			State:             ShopTaskStatePlanned,
			Sequence:          operation.Sequence,
			PlannedStart:      productionOrder.DatePlannedStart,
			PlannedFinish:     productionOrder.DatePlannedFinish,
			PlannedMinutes:    plannedMinutes,
		}
		workOrder, err = s.workOrders.Create(ctx, workOrder)
		if err != nil {
			return nil, err
		}
		created = append(created, workOrder)
	}
	return created, nil
}

func (s ProductionService) Consume(ctx context.Context, productionOrderID, componentID uint64, qty float64, journalID, wipAccountID uint64, date time.Time) (*ConsumedMaterial, error) {
	if !amount.FromFloat64(qty).GreaterThan(amount.Zero()) {
		return nil, ErrConsumeQuantity
	}
	if wipAccountID == 0 {
		return nil, ErrWIPAccount
	}

	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateInProgress {
		return nil, ErrProductionOrderState
	}
	if productionOrder.SrcLocationID == nil || productionOrder.OrganizationID == nil {
		return nil, ErrProductionOrderLocation
	}

	component, err := s.components.Find(ctx, componentID)
	if err != nil {
		return nil, err
	}
	if component == nil || component.ProductionOrderID != productionOrderID {
		return nil, ErrConsumeComponent
	}

	remaining := amount.FromFloat64(component.QtyPlanned).Sub(amount.FromFloat64(component.QtyConsumed))
	if amount.FromFloat64(qty).GreaterThan(remaining) {
		return nil, ErrConsumeQuantity
	}

	production, err := s.findProductionLocation(ctx, *productionOrder.OrganizationID)
	if err != nil {
		return nil, err
	}
	if production == nil {
		return nil, ErrProductionLocation
	}

	movement, err := s.ledger.CreateMovement(ctx, &inventory.StockMovement{
		OrganizationID: productionOrder.OrganizationID,
		ItemID:         component.ItemID,
		Qty:            qty,
		UnitID:         component.UnitID,
		SrcLocationID:  *productionOrder.SrcLocationID,
		DstLocationID:  production.ID,
		OriginType:     helper.Ptr(accounting.OriginTypeProductionOrder),
		OriginID:       helper.Ptr(productionOrder.ID),
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.valuation.Consume(ctx, movement.ID, journalID, wipAccountID, date); err != nil {
		return nil, err
	}

	component.QtyConsumed = amount.FromFloat64(component.QtyConsumed).Add(amount.FromFloat64(qty)).Float64()
	component.StockMovementID = helper.Ptr(movement.ID)
	return s.components.Update(ctx, component)
}

func (s ProductionService) Produce(ctx context.Context, productionOrderID uint64, qty float64, journalID, wipAccountID uint64, date time.Time) (*ProductionOrder, error) {
	if !amount.FromFloat64(qty).GreaterThan(amount.Zero()) {
		return nil, ErrProduceQuantity
	}
	if wipAccountID == 0 {
		return nil, ErrWIPAccount
	}

	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateInProgress {
		return nil, ErrProductionOrderState
	}
	if productionOrder.DstLocationID == nil || productionOrder.OrganizationID == nil {
		return nil, ErrProductionOrderLocation
	}

	remaining := amount.FromFloat64(productionOrder.QtyToProduce).Sub(amount.FromFloat64(productionOrder.QtyProduced))
	if amount.FromFloat64(qty).GreaterThan(remaining) {
		return nil, ErrProduceQuantity
	}

	resolved, err := s.resolver.Resolve(ctx, productionOrder.ItemID)
	if err != nil {
		return nil, err
	}
	if resolved.StockValuationAccountID == 0 {
		return nil, inventory.ErrValuationAccount
	}

	production, err := s.findProductionLocation(ctx, *productionOrder.OrganizationID)
	if err != nil {
		return nil, err
	}
	if production == nil {
		return nil, ErrProductionLocation
	}

	movement, err := s.ledger.CreateMovement(ctx, &inventory.StockMovement{
		OrganizationID: productionOrder.OrganizationID,
		ItemID:         productionOrder.ItemID,
		Qty:            qty,
		UnitID:         productionOrder.UnitID,
		SrcLocationID:  production.ID,
		DstLocationID:  *productionOrder.DstLocationID,
		OriginType:     helper.Ptr(accounting.OriginTypeProductionOrder),
		OriginID:       helper.Ptr(productionOrder.ID),
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.valuation.Produce(ctx, movement.ID, amount.FromFloat64(resolved.StandardCost), journalID, wipAccountID, date); err != nil {
		return nil, err
	}

	productionOrder.QtyProduced = amount.FromFloat64(productionOrder.QtyProduced).Add(amount.FromFloat64(qty)).Float64()
	return s.orders.Update(ctx, productionOrder)
}

func (s ProductionService) RecordLabor(ctx context.Context, productionOrderID, workOrderID uint64, minutes float64, journalID, wipAccountID, appliedLaborAccountID uint64, date time.Time) (*ShopTask, error) {
	if !amount.FromFloat64(minutes).GreaterThan(amount.Zero()) {
		return nil, ErrShopTaskLabor
	}
	if wipAccountID == 0 {
		return nil, ErrWIPAccount
	}
	if appliedLaborAccountID == 0 {
		return nil, ErrLaborAccount
	}

	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateInProgress {
		return nil, ErrProductionOrderState
	}

	workOrder, err := s.workOrders.Find(ctx, workOrderID)
	if err != nil {
		return nil, err
	}
	if workOrder == nil {
		return nil, ErrShopTaskNotFound
	}
	if workOrder.ProductionOrderID != productionOrderID {
		return nil, ErrShopTaskMismatch
	}

	workCenter, err := s.workCenters.Find(ctx, workOrder.WorkCenterID)
	if err != nil {
		return nil, err
	}
	if workCenter == nil {
		return nil, ErrShopTaskWorkCenter
	}

	costPerHour := amount.FromFloat64(helper.Deref(workCenter.CostPerHour, 0))
	hourly, err := costPerHour.Mul(amount.FromFloat64(minutes)).Div(amount.FromFloat64(60))
	if err != nil {
		return nil, err
	}
	cost := hourly.Round(4)

	ref := fmt.Sprintf("LAB/%d", workOrder.ID)
	if _, err := s.poster.Post(ctx, accounting.PostRequest{
		OrganizationID: *productionOrder.OrganizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            ref,
		OriginType:     accounting.OriginTypeProductionOrder,
		OriginID:       productionOrder.ID,
		Description:    "Manufacturing labor",
		Lines: []accounting.PostingLine{
			{AccountID: wipAccountID, Name: "Work in Progress", Debit: cost},
			{AccountID: appliedLaborAccountID, Name: "Applied Labor/OH", Credit: cost},
		},
	}); err != nil {
		return nil, err
	}

	workOrder.ActualMinutes = amount.FromFloat64(workOrder.ActualMinutes).Add(amount.FromFloat64(minutes)).Float64()
	now := time.Now().UTC()
	workOrder.DateFinished = &now
	workOrder.State = ShopTaskStateDone
	return s.workOrders.Update(ctx, workOrder)
}

func (s ProductionService) SettleVariance(ctx context.Context, productionOrderID, journalID, wipAccountID, varianceAccountID uint64, date time.Time) (*ProductionOrder, error) {
	if wipAccountID == 0 {
		return nil, ErrWIPAccount
	}
	if varianceAccountID == 0 {
		return nil, ErrVarianceAccount
	}

	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateInProgress {
		return nil, ErrProductionOrderState
	}
	if productionOrder.OrganizationID == nil {
		return nil, ErrProductionOrderOrganization
	}

	balance, err := s.lines.BalanceByOriginAndAccount(ctx, "production_order", productionOrderID, wipAccountID)
	if err != nil {
		return nil, err
	}

	if balance != 0 {
		var lines []accounting.PostingLine
		if balance > 0 {
			lines = []accounting.PostingLine{
				{AccountID: varianceAccountID, Name: "Manufacturing Variance", Debit: amount.FromFloat64(balance)},
				{AccountID: wipAccountID, Name: "Work in Progress", Credit: amount.FromFloat64(balance)},
			}
		} else {
			lines = []accounting.PostingLine{
				{AccountID: wipAccountID, Name: "Work in Progress", Debit: amount.FromFloat64(-balance)},
				{AccountID: varianceAccountID, Name: "Manufacturing Variance", Credit: amount.FromFloat64(-balance)},
			}
		}
		if _, err := s.poster.Post(ctx, accounting.PostRequest{
			OrganizationID: *productionOrder.OrganizationID,
			JournalID:      journalID,
			Date:           date,
			Ref:            fmt.Sprintf("VAR/%d", productionOrder.ID),
			OriginType:     accounting.OriginTypeProductionOrder,
			OriginID:       productionOrder.ID,
			Description:    "Manufacturing variance settlement",
			Lines:          lines,
		}); err != nil {
			return nil, err
		}
	}

	productionOrder.State = ProductionOrderStateDone
	now := time.Now().UTC()
	productionOrder.DateFinished = &now
	return s.orders.Update(ctx, productionOrder)
}

func (s ProductionService) findProductionLocation(ctx context.Context, organizationID uint64) (*reference.StockLocation, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: "production"}}})
	if err != nil {
		return nil, err
	}
	for _, location := range locations.Items {
		if helper.Deref(location.OrganizationID, 0) == organizationID {
			return location, nil
		}
	}
	return nil, nil
}

func (s ProductionService) ListShopTasksByMO(ctx context.Context, productionOrderID uint64) ([]*ShopTask, error) {
	return s.workOrders.ListByProductionOrder(ctx, productionOrderID)
}
