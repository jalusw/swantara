package manufacturing

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type RecipeDAO interface {
	dao.CRUD[Recipe]
	CreateWithLines(ctx context.Context, recipe *Recipe, lines []*RecipeLine) (*Recipe, error)
	ListByItem(ctx context.Context, itemID uint64) ([]*Recipe, error)
}

type recipeDAO struct {
	dao.Base[Recipe]
	db *gorm.DB
}

func NewRecipeDAO(db *gorm.DB) RecipeDAO {
	return recipeDAO{Base: dao.NewBase[Recipe](db), db: db}
}

func (d recipeDAO) ListByItem(ctx context.Context, itemID uint64) ([]*Recipe, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "item_id", Operator: query.Equal, Value: itemID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d recipeDAO) CreateWithLines(ctx context.Context, recipe *Recipe, lines []*RecipeLine) (*Recipe, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(recipe).Error; err != nil {
			return err
		}
		for _, line := range lines {
			line.RecipeID = recipe.ID
			if err := tx.Create(line).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return recipe, nil
}

type RecipeLineDAO interface {
	dao.CRUD[RecipeLine]
	ListByRecipe(ctx context.Context, recipeID uint64) ([]*RecipeLine, error)
}

type recipeLineDAO struct {
	dao.Base[RecipeLine]
}

func NewRecipeLineDAO(db *gorm.DB) RecipeLineDAO {
	return recipeLineDAO{Base: dao.NewBase[RecipeLine](db)}
}

func (d recipeLineDAO) ListByRecipe(ctx context.Context, recipeID uint64) ([]*RecipeLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "recipe_id", Operator: query.Equal, Value: recipeID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type ProductionOrderDAO interface {
	dao.CRUD[ProductionOrder]
	CreateWithComponents(ctx context.Context, productionOrder *ProductionOrder, components []*ConsumedMaterial) (*ProductionOrder, error)
}

type manufacturingOrderDAO struct {
	dao.Base[ProductionOrder]
	db *gorm.DB
}

func NewProductionOrderDAO(db *gorm.DB) ProductionOrderDAO {
	return manufacturingOrderDAO{Base: dao.NewBase[ProductionOrder](db), db: db}
}

func (d manufacturingOrderDAO) CreateWithComponents(ctx context.Context, productionOrder *ProductionOrder, components []*ConsumedMaterial) (*ProductionOrder, error) {
	err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(productionOrder).Error; err != nil {
			return err
		}
		for _, component := range components {
			component.ProductionOrderID = productionOrder.ID
			if err := tx.Create(component).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return productionOrder, nil
}

type ConsumedMaterialDAO interface {
	dao.CRUD[ConsumedMaterial]
	ListByProductionOrder(ctx context.Context, productionOrderID uint64) ([]*ConsumedMaterial, error)
}

type consumedMaterialDAO struct {
	dao.Base[ConsumedMaterial]
}

func NewConsumedMaterialDAO(db *gorm.DB) ConsumedMaterialDAO {
	return consumedMaterialDAO{Base: dao.NewBase[ConsumedMaterial](db)}
}

func (d consumedMaterialDAO) ListByProductionOrder(ctx context.Context, productionOrderID uint64) ([]*ConsumedMaterial, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "production_order_id", Operator: query.Equal, Value: productionOrderID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type ProductionStepDAO interface {
	dao.CRUD[ProductionStep]
	ListByRecipe(ctx context.Context, recipeID uint64) ([]*ProductionStep, error)
}

type routingOperationDAO struct {
	dao.Base[ProductionStep]
}

func NewProductionStepDAO(db *gorm.DB) ProductionStepDAO {
	return routingOperationDAO{Base: dao.NewBase[ProductionStep](db)}
}

func (d routingOperationDAO) ListByRecipe(ctx context.Context, recipeID uint64) ([]*ProductionStep, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "recipe_id", Operator: query.Equal, Value: recipeID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type ShopTaskDAO interface {
	dao.CRUD[ShopTask]
	ListByProductionOrder(ctx context.Context, productionOrderID uint64) ([]*ShopTask, error)
}

type workOrderDAO struct {
	dao.Base[ShopTask]
}

func NewShopTaskDAO(db *gorm.DB) ShopTaskDAO {
	return workOrderDAO{Base: dao.NewBase[ShopTask](db)}
}

func (d workOrderDAO) ListByProductionOrder(ctx context.Context, productionOrderID uint64) ([]*ShopTask, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "production_order_id", Operator: query.Equal, Value: productionOrderID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type PlanningRunDAO interface {
	dao.CRUD[PlanningRun]
}

type mrpRunDAO struct {
	dao.Base[PlanningRun]
}

func NewPlanningRunDAO(db *gorm.DB) PlanningRunDAO {
	return mrpRunDAO{Base: dao.NewBase[PlanningRun](db)}
}

type PlanningNeedDAO interface {
	dao.CRUD[PlanningNeed]
	ListByRun(ctx context.Context, runID uint64) ([]*PlanningNeed, error)
}

type mrpDemandDAO struct {
	dao.Base[PlanningNeed]
}

func NewPlanningNeedDAO(db *gorm.DB) PlanningNeedDAO {
	return mrpDemandDAO{Base: dao.NewBase[PlanningNeed](db)}
}

func (d mrpDemandDAO) ListByRun(ctx context.Context, runID uint64) ([]*PlanningNeed, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "planning_run_id", Operator: query.Equal, Value: runID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type PlannedSupplyDAO interface {
	dao.CRUD[PlannedSupply]
	ListByRun(ctx context.Context, runID uint64) ([]*PlannedSupply, error)
}

type mrpPlannedOrderDAO struct {
	dao.Base[PlannedSupply]
}

func NewPlannedSupplyDAO(db *gorm.DB) PlannedSupplyDAO {
	return mrpPlannedOrderDAO{Base: dao.NewBase[PlannedSupply](db)}
}

func (d mrpPlannedOrderDAO) ListByRun(ctx context.Context, runID uint64) ([]*PlannedSupply, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "planning_run_id", Operator: query.Equal, Value: runID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type DemandPlanDAO interface {
	dao.CRUD[DemandPlan]
}

type demandForecastDAO struct {
	dao.Base[DemandPlan]
}

func NewDemandPlanDAO(db *gorm.DB) DemandPlanDAO {
	return demandForecastDAO{Base: dao.NewBase[DemandPlan](db)}
}

type OutsideProcessingOrderDAO interface {
	dao.CRUD[OutsideProcessingOrder]
}

type subcontractOrderDAO struct {
	dao.Base[OutsideProcessingOrder]
}

func NewOutsideProcessingOrderDAO(db *gorm.DB) OutsideProcessingOrderDAO {
	return subcontractOrderDAO{Base: dao.NewBase[OutsideProcessingOrder](db)}
}
