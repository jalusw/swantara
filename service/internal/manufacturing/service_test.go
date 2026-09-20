package manufacturing

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

func TestRecipeService_Explode_ComputesScrapAdjustedQuantities(t *testing.T) {
	ctx := context.Background()

	recipes := RecipeDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Recipe, error) {
			return &Recipe{Base: model.Base{ID: 1}, ItemID: 200, Qty: 1, Type: RecipeTypeManufacture}, nil
		}),
	}
	lines := RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*RecipeLine, error) {
			return []*RecipeLine{
				{Base: model.Base{ID: 1}, RecipeID: 1, ComponentID: 300, Qty: 2, ScrapPct: 10},
				{Base: model.Base{ID: 2}, RecipeID: 1, ComponentID: 301, Qty: 0.5, ScrapPct: 0},
			}, nil
		},
	}
	svc := NewRecipeService(products.ItemVariantDAOMock{}, recipes, lines)

	requirements, err := svc.Explode(ctx, 1, amount.FromFloat64(10))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(requirements) != 2 {
		t.Fatalf("requirements = %d, want 2", len(requirements))
	}
	if got := requirements[0].Qty.String(); got != "22" {
		t.Errorf("component 300 qty = %s, want 22 (2 * 10 * 1.1 scrap)", got)
	}
	if got := requirements[1].Qty.String(); got != "5" {
		t.Errorf("component 301 qty = %s, want 5 (0.5 * 10)", got)
	}
}

func TestRecipeService_Explode_HonorsBatchQty(t *testing.T) {
	ctx := context.Background()

	recipes := RecipeDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Recipe, error) {
			return &Recipe{Base: model.Base{ID: 1}, ItemID: 200, Qty: 4, Type: RecipeTypeManufacture}, nil
		}),
	}
	lines := RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*RecipeLine, error) {
			return []*RecipeLine{
				{Base: model.Base{ID: 1}, RecipeID: 1, ComponentID: 300, Qty: 2, ScrapPct: 0},
			}, nil
		},
	}
	svc := NewRecipeService(products.ItemVariantDAOMock{}, recipes, lines)

	requirements, err := svc.Explode(ctx, 1, amount.FromFloat64(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := requirements[0].Qty.String(); got != "1" {
		t.Errorf("component 300 qty = %s, want 1 (2 * 2/4)", got)
	}
}

func TestRecipeService_Explode_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	foundRecipe := RecipeDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*Recipe, error) {
			return &Recipe{Base: model.Base{ID: 1}, ItemID: 200, Qty: 1, Type: RecipeTypeManufacture}, nil
		}),
	}

	tests := []struct {
		name    string
		recipes RecipeDAOMock
		lines   RecipeLineDAOMock
		qty     amount.Amount
		wantErr error
	}{
		{name: "non positive output", qty: amount.Zero(), wantErr: ErrRecipeOutputQty},
		{name: "missing recipe", qty: amount.FromFloat64(10), wantErr: ErrRecipeNotFound},
		{name: "recipe without lines", recipes: foundRecipe, qty: amount.FromFloat64(10), wantErr: ErrRecipeNoLines},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewRecipeService(products.ItemVariantDAOMock{}, tt.recipes, tt.lines)

			_, err := svc.Explode(ctx, 1, tt.qty)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestRecipeService_CreateRecipe_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	variants := products.ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
			return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
		}),
	}
	variantsWithMissingComponent := products.ItemVariantDAOMock{
		CRUDMock: daoCRUDFind(func(_ context.Context, id uint64) (*products.ItemVariant, error) {
			if id == 200 {
				return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
			}
			return nil, nil
		}),
	}

	tests := []struct {
		name     string
		variants products.ItemVariantDAOMock
		recipe   *Recipe
		lines    []*RecipeLine
		wantErr  error
	}{
		{name: "invalid type", recipe: &Recipe{ItemID: 200, Type: "outsource"}, wantErr: ErrInvalidRecipeType},
		{name: "negative qty", recipe: &Recipe{ItemID: 200, Type: RecipeTypeManufacture, Qty: -1}, wantErr: ErrRecipeQty},
		{name: "unknown item", recipe: &Recipe{ItemID: 200, Type: RecipeTypeManufacture}, wantErr: ErrRecipeItem},
		{name: "recipe without lines", variants: variants, recipe: &Recipe{ItemID: 200, Type: RecipeTypeManufacture}, wantErr: ErrRecipeNoLines},
		{
			name:     "unknown component",
			variants: variantsWithMissingComponent,
			recipe:   &Recipe{ItemID: 200, Type: RecipeTypeManufacture},
			lines:    []*RecipeLine{{ComponentID: 300, Qty: 2}},
			wantErr:  ErrRecipeComponent,
		},
		{
			name:     "non positive line qty",
			variants: variants,
			recipe:   &Recipe{ItemID: 200, Type: RecipeTypeManufacture},
			lines:    []*RecipeLine{{ComponentID: 300, Qty: 0}},
			wantErr:  ErrRecipeLineQty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewRecipeService(tt.variants, RecipeDAOMock{}, RecipeLineDAOMock{})

			_, err := svc.CreateRecipe(ctx, tt.recipe, tt.lines)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func daoCRUDFind[E any](fn func(ctx context.Context, id uint64) (*E, error)) dao.CRUDMock[E] {
	return dao.CRUDMock[E]{FindFunc: fn}
}
