package manufacturing

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

func TestManufacturingFixtures(t *testing.T) {
	if RecipeFixture() == nil {
		t.Error("recipe = nil")
	}
	if RecipeLineFixture() == nil {
		t.Error("line = nil")
	}
	if ProductionOrderFixture() == nil {
		t.Error("productionOrder = nil")
	}
	if ConsumedMaterialFixture() == nil {
		t.Error("component = nil")
	}
	if ShopTaskFixture() == nil {
		t.Error("shop task = nil")
	}
	if OutsideProcessingOrderFixture() == nil {
		t.Error("subcontract = nil")
	}
	if ProductionStepFixture() == nil {
		t.Error("operation = nil")
	}
	if RecipeFixture(func(b *Recipe) *Recipe { return b }) == nil {
		t.Error("recipe opt = nil")
	}
}

func TestManufacturingMock_Fallbacks(t *testing.T) {
	ctx := context.Background()

	t.Run("dao fallbacks", func(t *testing.T) {
		if _, err := (RecipeDAOMock{}).ListByItem(ctx, 1); err != nil {
			t.Errorf("ListByItem = %v", err)
		}
		if _, err := (RecipeLineDAOMock{}).ListByRecipe(ctx, 1); err != nil {
			t.Errorf("ListByRecipe = %v", err)
		}
		if _, err := (ConsumedMaterialDAOMock{}).ListByProductionOrder(ctx, 1); err != nil {
			t.Errorf("ListByProductionOrder = %v", err)
		}
		if _, err := (ProductionStepDAOMock{}).ListByRecipe(ctx, 1); err != nil {
			t.Errorf("Routing ListByRecipe = %v", err)
		}
		if _, err := (ShopTaskDAOMock{}).ListByProductionOrder(ctx, 1); err != nil {
			t.Errorf("ShopTask ListByProductionOrder = %v", err)
		}
		if _, err := (PlanningNeedDAOMock{}).ListByRun(ctx, 1); err != nil {
			t.Errorf("Demand ListByRun = %v", err)
		}
		if _, err := (PlannedSupplyDAOMock{}).ListByRun(ctx, 1); err != nil {
			t.Errorf("Planned ListByRun = %v", err)
		}
	})

	t.Run("dao with func", func(t *testing.T) {
		recipes := RecipeDAOMock{
			CreateWithLinesFunc: func(_ context.Context, b *Recipe, _ []*RecipeLine) (*Recipe, error) { return b, nil },
		}
		if _, err := recipes.CreateWithLines(ctx, &Recipe{}, nil); err != nil {
			t.Errorf("CreateWithLines = %v", err)
		}
		orders := ProductionOrderDAOMock{
			CreateWithComponentsFunc: func(_ context.Context, m *ProductionOrder, _ []*ConsumedMaterial) (*ProductionOrder, error) {
				return m, nil
			},
		}
		if _, err := orders.CreateWithComponents(ctx, &ProductionOrder{}, nil); err != nil {
			t.Errorf("CreateWithComponents = %v", err)
		}
		byRecipe := RecipeLineDAOMock{
			ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*RecipeLine, error) { return []*RecipeLine{{}}, nil },
		}
		if _, err := byRecipe.ListByRecipe(ctx, 1); err != nil {
			t.Errorf("ListByRecipe = %v", err)
		}
	})

	t.Run("service mocks", func(t *testing.T) {
		if _, err := (PurchasePlannerMock{}).Create(ctx, &procurement.PurchaseOrder{}, nil); err != nil {
			t.Errorf("purchase planner = %v", err)
		}
		if _, err := (ManufacturePlannerMock{}).Create(ctx, &ProductionOrder{}); err != nil {
			t.Errorf("manufacture create = %v", err)
		}
		if _, err := (ManufacturePlannerMock{}).Confirm(ctx, 1); err != nil {
			t.Errorf("manufacture confirm = %v", err)
		}
		if _, err := (TransferPlannerMock{}).Create(ctx, &inventory.WarehouseTransfer{}, nil); err != nil {
			t.Errorf("transfer planner = %v", err)
		}
		if _, err := (PurchaseOrderCreatorMock{}).Create(ctx, &procurement.PurchaseOrder{}, nil); err != nil {
			t.Errorf("order creator = %v", err)
		}
	})

	t.Run("service mocks with func", func(t *testing.T) {
		planner := PurchasePlannerMock{
			CreateFunc: func(_ context.Context, o *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
				return o, nil
			},
		}
		if _, err := planner.Create(ctx, &procurement.PurchaseOrder{}, nil); err != nil {
			t.Errorf("purchase planner = %v", err)
		}
		confirm := ManufacturePlannerMock{
			CreateFunc: func(_ context.Context, m *ProductionOrder) (*ProductionOrder, error) { return m, nil },
			ConfirmFunc: func(_ context.Context, id uint64) (*ProductionOrder, error) {
				return &ProductionOrder{State: ProductionOrderStateConfirmed}, nil
			},
		}
		if _, err := confirm.Create(ctx, &ProductionOrder{}); err != nil {
			t.Errorf("manufacture create = %v", err)
		}
		if _, err := confirm.Confirm(ctx, 1); err != nil {
			t.Errorf("manufacture confirm = %v", err)
		}
	})
}

var _ = dao.CRUDMock[Recipe]{}
var _ = query.Query{}
