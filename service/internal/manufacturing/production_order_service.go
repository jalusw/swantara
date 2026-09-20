package manufacturing

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

type ProductionOrderService struct {
	orders       ProductionOrderDAO
	components   ConsumedMaterialDAO
	recipes      RecipeDAO
	recipeSvc    RecipeService
	variants     products.ItemVariantDAO
	locations    inventory.StockLocationDAO
	reservations inventory.HoldService
	sequences    sequence.Service
}

func NewProductionOrderService(
	orders ProductionOrderDAO,
	components ConsumedMaterialDAO,
	recipes RecipeDAO,
	recipeSvc RecipeService,
	variants products.ItemVariantDAO,
	locations inventory.StockLocationDAO,
	reservations inventory.HoldService,
	sequences sequence.Service,
) ProductionOrderService {
	return ProductionOrderService{
		orders:       orders,
		components:   components,
		recipes:      recipes,
		recipeSvc:    recipeSvc,
		variants:     variants,
		locations:    locations,
		reservations: reservations,
		sequences:    sequences,
	}
}

func (s ProductionOrderService) Create(ctx context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
	if productionOrder.OrganizationID == nil {
		return nil, ErrProductionOrderOrganization
	}
	if !amount.FromFloat64(productionOrder.QtyToProduce).GreaterThan(amount.Zero()) {
		return nil, ErrProductionOrderQty
	}
	if productionOrder.RecipeID == nil {
		return nil, ErrProductionOrderRecipe
	}
	if productionOrder.SrcLocationID == nil || productionOrder.DstLocationID == nil {
		return nil, ErrProductionOrderLocation
	}

	variant, err := s.variants.Find(ctx, productionOrder.ItemID)
	if err != nil {
		return nil, err
	}
	if variant == nil {
		return nil, ErrProductionOrderItem
	}

	recipe, err := s.recipes.Find(ctx, *productionOrder.RecipeID)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, ErrProductionOrderRecipe
	}
	if recipe.ItemID != productionOrder.ItemID {
		return nil, ErrProductionOrderRecipe
	}

	if err := s.validateLocations(ctx, *productionOrder.OrganizationID, productionOrder.SrcLocationID, productionOrder.DstLocationID); err != nil {
		return nil, err
	}

	requirements, err := s.recipeSvc.Explode(ctx, *productionOrder.RecipeID, amount.FromFloat64(productionOrder.QtyToProduce))
	if err != nil {
		return nil, err
	}
	if len(requirements) == 0 {
		return nil, ErrConsumedMaterial
	}

	components := make([]*ConsumedMaterial, len(requirements))
	for i, requirement := range requirements {
		components[i] = &ConsumedMaterial{
			ItemID:          requirement.ComponentID,
			QtyPlanned:      requirement.Qty.Float64(),
			UnitID:          requirement.UnitID,
			StockMovementID: nil,
		}
	}

	name, err := s.sequences.Next(ctx, *productionOrder.OrganizationID, SequenceProductionOrderCode)
	if err != nil {
		return nil, err
	}
	productionOrder.Name = helper.Ptr(name)
	productionOrder.State = ProductionOrderStateDraft

	return s.orders.CreateWithComponents(ctx, productionOrder, components)
}

func (s ProductionOrderService) Confirm(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error) {
	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateDraft {
		return nil, ErrProductionOrderState
	}
	productionOrder.State = ProductionOrderStateConfirmed
	return s.orders.Update(ctx, productionOrder)
}

func (s ProductionOrderService) Plan(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error) {
	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State != ProductionOrderStateConfirmed {
		return nil, ErrProductionOrderState
	}
	if productionOrder.SrcLocationID == nil {
		return nil, ErrProductionOrderLocation
	}

	components, err := s.components.ListByProductionOrder(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if len(components) == 0 {
		return nil, ErrConsumedMaterial
	}

	for _, component := range components {
		if _, err := s.reservations.Reserve(ctx, *productionOrder.OrganizationID, component.ItemID, *productionOrder.SrcLocationID, nil, amount.FromFloat64(component.QtyPlanned), nil); err != nil {
			return nil, err
		}
	}

	productionOrder.State = ProductionOrderStatePlanned
	return s.orders.Update(ctx, productionOrder)
}

func (s ProductionOrderService) Cancel(ctx context.Context, productionOrderID uint64) (*ProductionOrder, error) {
	productionOrder, err := s.orders.Find(ctx, productionOrderID)
	if err != nil {
		return nil, err
	}
	if productionOrder == nil {
		return nil, ErrProductionOrderNotFound
	}
	if productionOrder.State == ProductionOrderStateDone || productionOrder.State == ProductionOrderStateCancelled {
		return nil, ErrProductionOrderState
	}
	productionOrder.State = ProductionOrderStateCancelled
	return s.orders.Update(ctx, productionOrder)
}

func (s ProductionOrderService) validateLocations(ctx context.Context, organizationID uint64, src, dst *uint64) error {
	for _, locationID := range []uint64{*src, *dst} {
		location, err := s.locations.Find(ctx, locationID)
		if err != nil {
			return err
		}
		if location == nil || !helper.OwnedByOrg(location.OrganizationID, &organizationID) {
			return ErrProductionOrderLocation
		}
	}
	return nil
}

func (s ProductionOrderService) List(ctx context.Context, q *query.Query) (*query.Page[ProductionOrder], error) {
	return s.orders.List(ctx, q)
}

func (s ProductionOrderService) Find(ctx context.Context, id uint64) (*ProductionOrder, error) {
	return s.orders.Find(ctx, id)
}

func (s ProductionOrderService) ListComponents(ctx context.Context, productionOrderID uint64) ([]*ConsumedMaterial, error) {
	return s.components.ListByProductionOrder(ctx, productionOrderID)
}
