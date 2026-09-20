package manufacturing

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func productionOrderExtraService(variantErr error, variant *products.ItemVariant, recipe *Recipe, recipeErr error, lineErr error, lines []*RecipeLine) ProductionOrderService {
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) { return variant, variantErr },
		},
	}
	recipes := RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*Recipe, error) { return recipe, recipeErr },
		},
	}
	lineMock := RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*RecipeLine, error) { return lines, lineErr },
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: id}, OrganizationID: helper.Ptr(uint64(10))}, nil
			},
		},
	}
	return NewTestProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, recipes, lineMock, variants, locations, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, sequence.DAOMock{})
}

func validMOInput() *ProductionOrder {
	orgID := uint64(10)
	src := uint64(30)
	dst := uint64(40)
	return &ProductionOrder{OrganizationID: &orgID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst}
}

func TestProductionOrderService_Create_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	variant := &products.ItemVariant{Base: model.Base{ID: 200}}
	recipe := &Recipe{Base: model.Base{ID: 1}, ItemID: 200}
	lines := []*RecipeLine{{ComponentID: 300, Qty: 2}}

	t.Run("variant lookup error", func(t *testing.T) {
		svc := productionOrderExtraService(dbErr, nil, recipe, nil, nil, lines)
		_, err := svc.Create(ctx, validMOInput())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("recipe lookup error and missing", func(t *testing.T) {
		svc := productionOrderExtraService(nil, variant, nil, dbErr, nil, lines)
		_, err := svc.Create(ctx, validMOInput())
		helper.AssertError(t, err, true, dbErr)

		svc = productionOrderExtraService(nil, variant, nil, nil, nil, lines)
		_, err = svc.Create(ctx, validMOInput())
		helper.AssertError(t, err, true, ErrProductionOrderRecipe)
	})

	t.Run("explode error and empty", func(t *testing.T) {
		svc := productionOrderExtraService(nil, variant, recipe, nil, dbErr, nil)
		_, err := svc.Create(ctx, validMOInput())
		helper.AssertError(t, err, true, dbErr)

		svc = productionOrderExtraService(nil, variant, recipe, nil, nil, nil)
		_, err = svc.Create(ctx, validMOInput())
		helper.AssertError(t, err, true, ErrRecipeNoLines)
	})
}

func TestRecipeService_Explode(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(recipe *Recipe, recipeErr error, lines []*RecipeLine, lineErr error) RecipeService {
		return NewRecipeService(products.ItemVariantDAOMock{}, RecipeDAOMock{
			CRUDMock: dao.CRUDMock[Recipe]{
				FindFunc: func(_ context.Context, _ uint64) (*Recipe, error) { return recipe, recipeErr },
			},
		}, RecipeLineDAOMock{
			ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*RecipeLine, error) { return lines, lineErr },
		})
	}
	recipe := &Recipe{Base: model.Base{ID: 1}, Qty: 1}
	lines := []*RecipeLine{{ComponentID: 300, Qty: 2}}

	t.Run("explodes requirements", func(t *testing.T) {
		svc := newSvc(recipe, nil, lines, nil)
		got, err := svc.Explode(ctx, 1, amount.FromFloat64(10))
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 1 || got[0].Qty.Float64() != 20 {
			t.Errorf("requirements = %+v", got)
		}
	})

	t.Run("explode failures", func(t *testing.T) {
		svc := newSvc(recipe, nil, lines, nil)
		_, err := svc.Explode(ctx, 1, amount.FromFloat64(0))
		helper.AssertError(t, err, true, ErrRecipeOutputQty)

		svc = newSvc(nil, dbErr, lines, nil)
		_, err = svc.Explode(ctx, 1, amount.FromFloat64(1))
		helper.AssertError(t, err, true, dbErr)

		svc = newSvc(nil, nil, lines, nil)
		_, err = svc.Explode(ctx, 1, amount.FromFloat64(1))
		helper.AssertError(t, err, true, ErrRecipeNotFound)

		svc = newSvc(recipe, nil, nil, dbErr)
		_, err = svc.Explode(ctx, 1, amount.FromFloat64(1))
		helper.AssertError(t, err, true, dbErr)

		svc = newSvc(recipe, nil, nil, nil)
		_, err = svc.Explode(ctx, 1, amount.FromFloat64(1))
		helper.AssertError(t, err, true, ErrRecipeNoLines)
	})
}

func TestRecipeService_Validate(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(variant *products.ItemVariant, variantErr error) RecipeService {
		return NewRecipeService(products.ItemVariantDAOMock{
			CRUDMock: dao.CRUDMock[products.ItemVariant]{
				FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) { return variant, variantErr },
			},
		}, RecipeDAOMock{}, RecipeLineDAOMock{})
	}

	t.Run("validate recipe failures", func(t *testing.T) {
		svc := newSvc(&products.ItemVariant{}, nil)
		helper.AssertError(t, svc.validateRecipe(ctx, &Recipe{Type: "mystery"}), true, ErrInvalidRecipeType)
		helper.AssertError(t, svc.validateRecipe(ctx, &Recipe{Type: RecipeTypeManufacture, Qty: -1}), true, ErrRecipeQty)

		svc = newSvc(nil, dbErr)
		helper.AssertError(t, svc.validateRecipe(ctx, &Recipe{Type: RecipeTypeManufacture}), true, dbErr)

		svc = newSvc(nil, nil)
		helper.AssertError(t, svc.validateRecipe(ctx, &Recipe{Type: RecipeTypeManufacture}), true, ErrRecipeItem)
	})

	t.Run("validate line failures", func(t *testing.T) {
		svc := newSvc(&products.ItemVariant{}, nil)
		helper.AssertError(t, svc.validateLine(ctx, &RecipeLine{Qty: 0}), true, ErrRecipeLineQty)
		helper.AssertError(t, svc.validateLine(ctx, &RecipeLine{Qty: 1, ScrapPct: -1}), true, ErrRecipeLineScrap)

		svc = newSvc(nil, dbErr)
		helper.AssertError(t, svc.validateLine(ctx, &RecipeLine{Qty: 1}), true, dbErr)

		svc = newSvc(nil, nil)
		helper.AssertError(t, svc.validateLine(ctx, &RecipeLine{Qty: 1}), true, ErrRecipeComponent)
	})
}

func TestProductionOrderService_Transitions_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(productionOrder *ProductionOrder, moErr error, compErr error) ProductionOrderService {
		svc := NewTestProductionOrderService(ProductionOrderDAOMock{
			CRUDMock: dao.CRUDMock[ProductionOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return productionOrder, moErr },
			},
		}, ConsumedMaterialDAOMock{
			ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) { return nil, compErr },
		}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, sequence.DAOMock{})
		return svc
	}

	t.Run("confirm lookup error", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		_, err := svc.Confirm(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("plan lookup and components errors", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		_, err := svc.Plan(ctx, 1)
		helper.AssertError(t, err, true, dbErr)

		confirmed := &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateConfirmed, SrcLocationID: helper.Ptr(uint64(30))}
		svc = newSvc(confirmed, nil, dbErr)
		_, err = svc.Plan(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("cancel lookup error", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		_, err := svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("locations lookup error", func(t *testing.T) {
		svc := NewTestProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, RecipeDAOMock{}, RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) { return nil, dbErr },
			},
		}, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, sequence.DAOMock{})
		err := svc.validateLocations(ctx, 10, helper.Ptr(uint64(30)), helper.Ptr(uint64(40)))
		helper.AssertError(t, err, true, dbErr)
	})
}
