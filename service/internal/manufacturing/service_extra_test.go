package manufacturing

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

func TestRecipeService_CreateRecipe_SetsDefaultsAndPersists(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
			},
		},
	}
	svc := NewRecipeService(variants, RecipeDAOMock{}, RecipeLineDAOMock{})

	recipe, err := svc.CreateRecipe(ctx, &Recipe{ItemID: 200, Type: RecipeTypeManufacture}, []*RecipeLine{{ComponentID: 300, Qty: 2, ScrapPct: 5}})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recipe.Qty != 1 || recipe.Version != 1 || !recipe.Active {
		t.Errorf("recipe = qty %v version %d active %v, want defaults 1/1/true", recipe.Qty, recipe.Version, recipe.Active)
	}
}

func TestRecipeService_CreateRecipe_PropagatesPersistError(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
			},
		},
	}
	recipes := RecipeDAOMock{
		CreateWithLinesFunc: func(_ context.Context, _ *Recipe, _ []*RecipeLine) (*Recipe, error) {
			return nil, errors.New("db down")
		},
	}
	svc := NewRecipeService(variants, recipes, RecipeLineDAOMock{})

	_, err := svc.CreateRecipe(ctx, &Recipe{ItemID: 200, Type: RecipeTypeManufacture}, []*RecipeLine{{ComponentID: 300, Qty: 2}})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestRecipeService_CreateRecipe_RejectsNegativeScrap(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
			},
		},
	}
	svc := NewRecipeService(variants, RecipeDAOMock{}, RecipeLineDAOMock{})

	_, err := svc.CreateRecipe(ctx, &Recipe{ItemID: 200, Type: RecipeTypeManufacture}, []*RecipeLine{{ComponentID: 300, Qty: 2, ScrapPct: -1}})

	if helper.AssertError(t, err, true, ErrRecipeLineScrap) {
		return
	}
}

func TestRecipeService_UpdateRecipe_ValidatesAndUpdates(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
			},
		},
	}
	var updated *Recipe
	recipes := RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			UpdateFunc: func(_ context.Context, recipe *Recipe) (*Recipe, error) {
				updated = recipe
				return recipe, nil
			},
		},
	}
	svc := NewRecipeService(variants, recipes, RecipeLineDAOMock{})
	recipe := &Recipe{Base: model.Base{ID: 1}, ItemID: 200, Type: RecipeTypeManufacture, Qty: 4}

	result, err := svc.UpdateRecipe(ctx, recipe)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil || updated != recipe {
		t.Errorf("UpdateRecipe did not persist the recipe")
	}
	if result.Qty != 4 {
		t.Errorf("result qty = %v, want 4", result.Qty)
	}
}

func TestRecipeService_UpdateRecipe_RejectsInvalidType(t *testing.T) {
	ctx := context.Background()
	svc := NewRecipeService(products.ItemVariantDAOMock{}, RecipeDAOMock{}, RecipeLineDAOMock{})

	_, err := svc.UpdateRecipe(ctx, &Recipe{ItemID: 200, Type: "outsource"})

	if helper.AssertError(t, err, true, ErrInvalidRecipeType) {
		return
	}
}

func TestRecipeService_UpdateRecipe_PropagatesPersistError(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
			},
		},
	}
	recipes := RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			UpdateFunc: func(_ context.Context, _ *Recipe) (*Recipe, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := NewRecipeService(variants, recipes, RecipeLineDAOMock{})

	_, err := svc.UpdateRecipe(ctx, &Recipe{Base: model.Base{ID: 1}, ItemID: 200, Type: RecipeTypeManufacture, Qty: 4})

	if helper.AssertError(t, err, true, nil) {
		return
	}
}
