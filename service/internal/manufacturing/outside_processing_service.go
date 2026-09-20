package manufacturing

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type PurchaseOrderCreator interface {
	Create(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error)
}

type OutsideProcessingService struct {
	orders           OutsideProcessingOrderDAO
	productionOrders ProductionOrderDAO
	components       ConsumedMaterialDAO
	recipes          RecipeDAO
	purchases        PurchaseOrderCreator
	poDAO            procurement.PurchaseOrderDAO
	poLines          procurement.PurchaseOrderLineDAO
	locations        inventory.StockLocationDAO
	ledger           inventory.LedgerService
	valuation        inventory.ValuationService
	resolver         inventory.ItemResolver
	poster           accounting.Poster
}

func NewOutsideProcessingService(
	orders OutsideProcessingOrderDAO,
	productionOrders ProductionOrderDAO,
	components ConsumedMaterialDAO,
	recipes RecipeDAO,
	purchases PurchaseOrderCreator,
	poDAO procurement.PurchaseOrderDAO,
	poLines procurement.PurchaseOrderLineDAO,
	locations inventory.StockLocationDAO,
	ledger inventory.LedgerService,
	valuation inventory.ValuationService,
	resolver inventory.ItemResolver,
	poster accounting.Poster,
) OutsideProcessingService {
	return OutsideProcessingService{
		orders:           orders,
		productionOrders: productionOrders,
		components:       components,
		recipes:          recipes,
		purchases:        purchases,
		poDAO:            poDAO,
		poLines:          poLines,
		locations:        locations,
		ledger:           ledger,
		valuation:        valuation,
		resolver:         resolver,
		poster:           poster,
	}
}

func (s OutsideProcessingService) Create(ctx context.Context, productionOrderID, supplierID uint64) (*OutsideProcessingOrder, error) {
	productionOrder, err := s.productionOrders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State == ProductionOrderStateDraft {
		return nil, ErrOutsideProcessingOrderState
	}
	if productionOrder.RecipeID == nil {
		return nil, ErrProductionOrderRecipe
	}
	recipe, err := s.recipes.Find(ctx, *productionOrder.RecipeID)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, ErrProductionOrderRecipe
	}
	if recipe.Type != RecipeTypeSubcontract {
		return nil, ErrOutsideProcessingNotSubcontracted
	}

	existingPage, err := s.orders.List(ctx, &query.Query{Filters: []query.Filter{{Field: "production_order_id", Operator: query.Equal, Value: productionOrderID}}})
	if err != nil {
		return nil, err
	}
	if len(existingPage.Items) > 0 {
		return nil, ErrOutsideProcessingDuplicate
	}

	return s.orders.Create(ctx, &OutsideProcessingOrder{
		ProductionOrderID: productionOrderID,
		SupplierID:        supplierID,
		State:             OutsideProcessingStateDraft,
	})
}

func (s OutsideProcessingService) Send(ctx context.Context, orderID, journalID, wipAccountID uint64, date time.Time) (*OutsideProcessingOrder, error) {
	if wipAccountID == 0 {
		return nil, ErrWIPAccount
	}

	order, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOutsideProcessingNotFound
	}
	if order.State != OutsideProcessingStateDraft {
		return nil, ErrOutsideProcessingState
	}
	if order.PurchaseOrderID != nil {
		return nil, ErrOutsideProcessingState
	}

	productionOrder, err := s.productionOrders.Find(ctx, order.ProductionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateConfirmed && productionOrder.State != ProductionOrderStatePlanned {
		return nil, ErrOutsideProcessingOrderState
	}
	if productionOrder.SrcLocationID == nil || productionOrder.OrganizationID == nil {
		return nil, ErrProductionOrderLocation
	}

	components, err := s.components.ListByProductionOrder(ctx, order.ProductionOrderID)
	if err != nil {
		return nil, err
	}
	if len(components) == 0 {
		return nil, ErrOutsideProcessingComponents
	}

	supplierLoc, err := s.findSupplierLocation(ctx, *productionOrder.OrganizationID)
	if err != nil {
		return nil, err
	}
	if supplierLoc == 0 {
		return nil, ErrOutsideProcessingSupplierLocation
	}

	po, err := s.purchases.Create(ctx, &procurement.PurchaseOrder{
		OrganizationID: productionOrder.OrganizationID,
		SupplierID:     order.SupplierID,
		State:          procurement.PurchaseOrderStateDraft,
	}, []*procurement.PurchaseOrderLine{{
		ItemID:     helper.Ptr(productionOrder.ItemID),
		QtyOrdered: productionOrder.QtyToProduce,
		UnitID:     productionOrder.UnitID,
		UnitPrice:  0,
	}})
	if err != nil {
		return nil, err
	}
	if po.AmountUntaxed <= 0 {
		return nil, ErrOutsideProcessingOperation
	}

	for _, component := range components {
		movement, err := s.ledger.CreateMovement(ctx, &inventory.StockMovement{
			OrganizationID: productionOrder.OrganizationID,
			ItemID:         component.ItemID,
			Qty:            component.QtyPlanned,
			UnitID:         component.UnitID,
			SrcLocationID:  *productionOrder.SrcLocationID,
			DstLocationID:  supplierLoc,
			OriginType:     helper.Ptr(accounting.OriginTypeOutsideProcessingOrder),
			OriginID:       helper.Ptr(order.ID),
		})
		if err != nil {
			return nil, err
		}
		if _, err := s.valuation.Consume(ctx, movement.ID, journalID, wipAccountID, date); err != nil {
			return nil, err
		}
	}

	order.PurchaseOrderID = helper.Ptr(po.ID)
	order.State = OutsideProcessingStateSent
	return s.orders.Update(ctx, order)
}

func (s OutsideProcessingService) Receive(ctx context.Context, orderID, journalID, wipAccountID, apPayableAccountID uint64, date time.Time) (*OutsideProcessingOrder, error) {
	if wipAccountID == 0 {
		return nil, ErrWIPAccount
	}
	if apPayableAccountID == 0 {
		return nil, ErrOutsideProcessingOperation
	}

	order, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOutsideProcessingNotFound
	}
	if order.State != OutsideProcessingStateSent {
		return nil, ErrOutsideProcessingState
	}
	if order.PurchaseOrderID == nil {
		return nil, ErrOutsideProcessingPurchaseOrder
	}

	productionOrder, err := s.productionOrders.Find(ctx, order.ProductionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateConfirmed && productionOrder.State != ProductionOrderStatePlanned && productionOrder.State != ProductionOrderStateInProgress {
		return nil, ErrOutsideProcessingOrderState
	}
	if productionOrder.DstLocationID == nil || productionOrder.OrganizationID == nil {
		return nil, ErrProductionOrderLocation
	}

	po, err := s.poDAO.Find(ctx, *order.PurchaseOrderID)
	if err != nil {
		return nil, err
	}
	if po == nil {
		return nil, ErrOutsideProcessingPurchaseOrder
	}

	supplierLoc, err := s.findSupplierLocation(ctx, *productionOrder.OrganizationID)
	if err != nil {
		return nil, err
	}
	if supplierLoc == 0 {
		return nil, ErrOutsideProcessingSupplierLocation
	}

	poLines, err := s.poLines.ListByOrder(ctx, po.ID)
	if err != nil {
		return nil, err
	}
	for _, line := range poLines {
		line.QtyReceived = line.QtyOrdered
		if _, err := s.poLines.Update(ctx, line); err != nil {
			return nil, err
		}
	}

	ref := fmt.Sprintf("SUB/%d", order.ID)
	if _, err := s.poster.Post(ctx, accounting.PostRequest{
		OrganizationID: *productionOrder.OrganizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            ref,
		OriginType:     accounting.OriginTypeOutsideProcessingOrder,
		OriginID:       order.ID,
		Description:    "Subcontract operation",
		Lines: []accounting.PostingLine{
			{AccountID: wipAccountID, Name: "Work in Progress", Debit: amount.FromFloat64(po.AmountUntaxed)},
			{AccountID: apPayableAccountID, Name: "Accounts Payable", Credit: amount.FromFloat64(po.AmountUntaxed)},
		},
	}); err != nil {
		return nil, err
	}

	components, err := s.components.ListByProductionOrder(ctx, order.ProductionOrderID)
	if err != nil {
		return nil, err
	}
	componentCost := amount.Zero()
	for _, component := range components {
		resolved, err := s.resolver.Resolve(ctx, component.ItemID)
		if err != nil {
			return nil, err
		}
		componentCost = componentCost.Add(amount.FromFloat64(resolved.StandardCost).Mul(amount.FromFloat64(component.QtyPlanned)))
	}
	produceValue := componentCost.Add(amount.FromFloat64(po.AmountUntaxed))
	unitCost, err := produceValue.Div(amount.FromFloat64(productionOrder.QtyToProduce))
	if err != nil {
		return nil, err
	}

	movement, err := s.ledger.CreateMovement(ctx, &inventory.StockMovement{
		OrganizationID: productionOrder.OrganizationID,
		ItemID:         productionOrder.ItemID,
		Qty:            productionOrder.QtyToProduce,
		UnitID:         productionOrder.UnitID,
		SrcLocationID:  supplierLoc,
		DstLocationID:  *productionOrder.DstLocationID,
		OriginType:     helper.Ptr(accounting.OriginTypeOutsideProcessingOrder),
		OriginID:       helper.Ptr(order.ID),
	})
	if err != nil {
		return nil, err
	}
	if _, err := s.valuation.Produce(ctx, movement.ID, unitCost, journalID, wipAccountID, date); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	productionOrder.QtyProduced = productionOrder.QtyToProduce
	productionOrder.State = ProductionOrderStateDone
	productionOrder.DateFinished = &now
	if _, err := s.productionOrders.Update(ctx, productionOrder); err != nil {
		return nil, err
	}

	order.State = OutsideProcessingStateReceived
	return s.orders.Update(ctx, order)
}

func (s OutsideProcessingService) Done(ctx context.Context, orderID uint64) (*OutsideProcessingOrder, error) {
	order, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOutsideProcessingNotFound
	}
	if order.State != OutsideProcessingStateReceived {
		return nil, ErrOutsideProcessingState
	}
	order.State = OutsideProcessingStateDone
	return s.orders.Update(ctx, order)
}

func (s OutsideProcessingService) Cancel(ctx context.Context, orderID uint64) (*OutsideProcessingOrder, error) {
	order, err := s.orders.Find(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOutsideProcessingNotFound
	}
	if order.State == OutsideProcessingStateDone || order.State == OutsideProcessingStateCancelled {
		return nil, ErrOutsideProcessingState
	}
	order.State = OutsideProcessingStateCancelled
	return s.orders.Update(ctx, order)
}

func (s OutsideProcessingService) findSupplierLocation(ctx context.Context, organizationID uint64) (uint64, error) {
	locations, err := s.locations.List(ctx, &query.Query{Filters: []query.Filter{{Field: "usage", Operator: query.Equal, Value: "supplier"}}})
	if err != nil {
		return 0, err
	}
	for _, location := range locations.Items {
		if helper.Deref(location.OrganizationID, 0) == 0 || helper.Deref(location.OrganizationID, 0) == organizationID {
			return location.ID, nil
		}
	}
	return 0, nil
}
