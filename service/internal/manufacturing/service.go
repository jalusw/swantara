package manufacturing

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

type RecipeService struct {
	variants products.ItemVariantDAO
	recipes  RecipeDAO
	lines    RecipeLineDAO
}

func NewRecipeService(variants products.ItemVariantDAO, recipes RecipeDAO, lines RecipeLineDAO) RecipeService {
	return RecipeService{
		variants: variants,
		recipes:  recipes,
		lines:    lines,
	}
}

func (s RecipeService) CreateRecipe(ctx context.Context, recipe *Recipe, lines []*RecipeLine) (*Recipe, error) {
	if err := s.validateRecipe(ctx, recipe); err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrRecipeNoLines
	}
	for _, line := range lines {
		if err := s.validateLine(ctx, line); err != nil {
			return nil, err
		}
	}
	if recipe.Qty == 0 {
		recipe.Qty = 1
	}
	if recipe.Version == 0 {
		recipe.Version = 1
	}
	recipe.Active = true
	return s.recipes.CreateWithLines(ctx, recipe, lines)
}

func (s RecipeService) UpdateRecipe(ctx context.Context, recipe *Recipe) (*Recipe, error) {
	if err := s.validateRecipe(ctx, recipe); err != nil {
		return nil, err
	}
	return s.recipes.Update(ctx, recipe)
}

func (s RecipeService) Explode(ctx context.Context, recipeID uint64, outputQty amount.Amount) ([]RecipeComponentRequirement, error) {
	if !outputQty.IsPositive() {
		return nil, ErrRecipeOutputQty
	}
	recipe, err := s.recipes.Find(ctx, recipeID)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return nil, ErrRecipeNotFound
	}
	lines, err := s.lines.ListByRecipe(ctx, recipeID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, ErrRecipeNoLines
	}

	batch := amount.FromFloat64(recipe.Qty)
	if batch.IsZero() {
		batch = amount.FromFloat64(1)
	}
	ratio, err := outputQty.Div(batch)
	if err != nil {
		return nil, err
	}

	requirements := make([]RecipeComponentRequirement, 0, len(lines))
	for _, line := range lines {
		qty := amount.FromFloat64(line.Qty).
			Mul(ratio).
			Mul(amount.FromFloat64(1 + line.ScrapPct/100)).
			Round(4)
		requirements = append(requirements, RecipeComponentRequirement{
			ComponentID: line.ComponentID,
			Qty:         qty,
			UnitID:      line.UnitID,
			ScrapPct:    line.ScrapPct,
		})
	}
	return requirements, nil
}

func (s RecipeService) validateRecipe(ctx context.Context, recipe *Recipe) error {
	if !validRecipeType[recipe.Type] {
		return ErrInvalidRecipeType
	}
	if recipe.Qty < 0 {
		return ErrRecipeQty
	}
	variant, err := s.variants.Find(ctx, recipe.ItemID)
	if err != nil {
		return err
	}
	if variant == nil {
		return ErrRecipeItem
	}
	return nil
}

func (s RecipeService) validateLine(ctx context.Context, line *RecipeLine) error {
	if line.Qty <= 0 {
		return ErrRecipeLineQty
	}
	if line.ScrapPct < 0 {
		return ErrRecipeLineScrap
	}
	variant, err := s.variants.Find(ctx, line.ComponentID)
	if err != nil {
		return err
	}
	if variant == nil {
		return ErrRecipeComponent
	}
	return nil
}

type RecipeComponentRequirement struct {
	ComponentID uint64        `json:"component_id"`
	Qty         amount.Amount `json:"qty"`
	UnitID      *uint64       `json:"unit_id,omitempty"`
	ScrapPct    float64       `json:"scrap_pct"`
}

func (s RecipeService) List(ctx context.Context, q *query.Query) (*query.Page[Recipe], error) {
	return s.recipes.List(ctx, q)
}

func (s RecipeService) Find(ctx context.Context, id uint64) (*Recipe, error) {
	return s.recipes.Find(ctx, id)
}

func (s RecipeService) Delete(ctx context.Context, id uint64) error {
	return s.recipes.Delete(ctx, id)
}

func (s RecipeService) ListLines(ctx context.Context, recipeID uint64) ([]*RecipeLine, error) {
	return s.lines.ListByRecipe(ctx, recipeID)
}
