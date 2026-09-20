package manufacturing

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
)

type RecipeDAOMock struct {
	dao.CRUDMock[Recipe]
	CreateWithLinesFunc func(ctx context.Context, recipe *Recipe, lines []*RecipeLine) (*Recipe, error)
	ListByItemFunc      func(ctx context.Context, itemID uint64) ([]*Recipe, error)
}

func (m RecipeDAOMock) CreateWithLines(ctx context.Context, recipe *Recipe, lines []*RecipeLine) (*Recipe, error) {
	if m.CreateWithLinesFunc != nil {
		return m.CreateWithLinesFunc(ctx, recipe, lines)
	}
	return recipe, nil
}

func (m RecipeDAOMock) ListByItem(ctx context.Context, itemID uint64) ([]*Recipe, error) {
	if m.ListByItemFunc != nil {
		return m.ListByItemFunc(ctx, itemID)
	}
	return []*Recipe{}, nil
}

type RecipeLineDAOMock struct {
	dao.CRUDMock[RecipeLine]
	ListByRecipeFunc func(ctx context.Context, recipeID uint64) ([]*RecipeLine, error)
}

func (m RecipeLineDAOMock) ListByRecipe(ctx context.Context, recipeID uint64) ([]*RecipeLine, error) {
	if m.ListByRecipeFunc != nil {
		return m.ListByRecipeFunc(ctx, recipeID)
	}
	return []*RecipeLine{}, nil
}

type ProductionOrderDAOMock struct {
	dao.CRUDMock[ProductionOrder]
	CreateWithComponentsFunc func(ctx context.Context, productionOrder *ProductionOrder, components []*ConsumedMaterial) (*ProductionOrder, error)
}

func (m ProductionOrderDAOMock) CreateWithComponents(ctx context.Context, productionOrder *ProductionOrder, components []*ConsumedMaterial) (*ProductionOrder, error) {
	if m.CreateWithComponentsFunc != nil {
		return m.CreateWithComponentsFunc(ctx, productionOrder, components)
	}
	return productionOrder, nil
}

type ConsumedMaterialDAOMock struct {
	dao.CRUDMock[ConsumedMaterial]
	ListByProductionOrderFunc func(ctx context.Context, productionOrderID uint64) ([]*ConsumedMaterial, error)
}

func (m ConsumedMaterialDAOMock) ListByProductionOrder(ctx context.Context, productionOrderID uint64) ([]*ConsumedMaterial, error) {
	if m.ListByProductionOrderFunc != nil {
		return m.ListByProductionOrderFunc(ctx, productionOrderID)
	}
	return []*ConsumedMaterial{}, nil
}

type ProductionStepDAOMock struct {
	dao.CRUDMock[ProductionStep]
	ListByRecipeFunc func(ctx context.Context, recipeID uint64) ([]*ProductionStep, error)
}

func (m ProductionStepDAOMock) ListByRecipe(ctx context.Context, recipeID uint64) ([]*ProductionStep, error) {
	if m.ListByRecipeFunc != nil {
		return m.ListByRecipeFunc(ctx, recipeID)
	}
	return []*ProductionStep{}, nil
}

type ShopTaskDAOMock struct {
	dao.CRUDMock[ShopTask]
	ListByProductionOrderFunc func(ctx context.Context, productionOrderID uint64) ([]*ShopTask, error)
}

func (m ShopTaskDAOMock) ListByProductionOrder(ctx context.Context, productionOrderID uint64) ([]*ShopTask, error) {
	if m.ListByProductionOrderFunc != nil {
		return m.ListByProductionOrderFunc(ctx, productionOrderID)
	}
	return []*ShopTask{}, nil
}

type PlanningRunDAOMock struct {
	dao.CRUDMock[PlanningRun]
}

type PlanningNeedDAOMock struct {
	dao.CRUDMock[PlanningNeed]
	ListByRunFunc func(ctx context.Context, runID uint64) ([]*PlanningNeed, error)
}

func (m PlanningNeedDAOMock) ListByRun(ctx context.Context, runID uint64) ([]*PlanningNeed, error) {
	if m.ListByRunFunc != nil {
		return m.ListByRunFunc(ctx, runID)
	}
	return []*PlanningNeed{}, nil
}

type PlannedSupplyDAOMock struct {
	dao.CRUDMock[PlannedSupply]
	ListByRunFunc func(ctx context.Context, runID uint64) ([]*PlannedSupply, error)
}

func (m PlannedSupplyDAOMock) ListByRun(ctx context.Context, runID uint64) ([]*PlannedSupply, error) {
	if m.ListByRunFunc != nil {
		return m.ListByRunFunc(ctx, runID)
	}
	return []*PlannedSupply{}, nil
}

type DemandPlanDAOMock struct {
	dao.CRUDMock[DemandPlan]
}

type OutsideProcessingOrderDAOMock struct {
	dao.CRUDMock[OutsideProcessingOrder]
}
