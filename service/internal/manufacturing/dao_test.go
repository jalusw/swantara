package manufacturing

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestRecipeDAO_ListByItem_ReturnsRecipes(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "recipes" WHERE item_id = $1`)).
		WithArgs(200).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "recipes" WHERE item_id = $1`)).
		WithArgs(200).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "qty", "type", "active"}).
			AddRow(1, 200, 1, RecipeTypeManufacture, true).
			AddRow(2, 200, 2, RecipeTypeKit, false))

	recipes := NewRecipeDAO(db)

	items, err := recipes.ListByItem(ctx, 200)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Qty != 1 || items[1].Type != RecipeTypeKit {
		t.Errorf("items = %+v, want two recipes for item 200", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRecipeDAO_ListByItem_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "recipes" WHERE item_id = $1`)).
		WithArgs(200).
		WillReturnError(errors.New("db down"))

	recipes := NewRecipeDAO(db)

	_, err := recipes.ListByItem(ctx, 200)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestRecipeDAO_CreateWithLines_CreatesRecipeAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipe_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipe_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))
	mock.ExpectCommit()

	recipes := NewRecipeDAO(db)
	recipe := &Recipe{ItemID: 200, Qty: 1, Type: RecipeTypeManufacture, Active: true}
	lines := []*RecipeLine{{ComponentID: 300, Qty: 2}, {ComponentID: 301, Qty: 0.5}}

	created, err := recipes.CreateWithLines(ctx, recipe, lines)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("recipe id = %d, want 1", created.ID)
	}
	if lines[0].RecipeID != 1 || lines[1].RecipeID != 1 {
		t.Errorf("lines recipe_id = %d/%d, want both 1", lines[0].RecipeID, lines[1].RecipeID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRecipeDAO_CreateWithLines_RollsBackOnRecipeError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipes"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	recipes := NewRecipeDAO(db)

	_, err := recipes.CreateWithLines(ctx, &Recipe{}, []*RecipeLine{{ComponentID: 300, Qty: 2}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestRecipeDAO_CreateWithLines_RollsBackOnLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipe_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	recipes := NewRecipeDAO(db)

	_, err := recipes.CreateWithLines(ctx, &Recipe{}, []*RecipeLine{{ComponentID: 300, Qty: 2}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestRecipeLineDAO_ListByRecipe_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "recipe_lines" WHERE recipe_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "recipe_lines" WHERE recipe_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "component_id", "qty", "scrap_pct"}).
			AddRow(11, 1, 300, 2, 10))

	lines := NewRecipeLineDAO(db)

	items, err := lines.ListByRecipe(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ComponentID != 300 || items[0].Qty != 2 {
		t.Errorf("items = %+v, want one line for recipe 1", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRecipeLineDAO_ListByRecipe_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "recipe_lines" WHERE recipe_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	lines := NewRecipeLineDAO(db)

	_, err := lines.ListByRecipe(ctx, 1)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestProductionOrderDAO_CreateWithComponents_CreatesMOAndComponents(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "production_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consumed_materials"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
	mock.ExpectCommit()

	orders := NewProductionOrderDAO(db)
	productionOrder := &ProductionOrder{ItemID: 200, QtyToProduce: 10, State: ProductionOrderStateDraft}
	components := []*ConsumedMaterial{{ItemID: 300, QtyPlanned: 22}}

	created, err := orders.CreateWithComponents(ctx, productionOrder, components)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("productionOrder id = %d, want 1", created.ID)
	}
	if components[0].ProductionOrderID != 1 {
		t.Errorf("component production_order_id = %d, want 1", components[0].ProductionOrderID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProductionOrderDAO_CreateWithComponents_RollsBackOnError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "production_orders"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	orders := NewProductionOrderDAO(db)

	_, err := orders.CreateWithComponents(ctx, &ProductionOrder{}, []*ConsumedMaterial{{ItemID: 300, QtyPlanned: 22}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestConsumedMaterialDAO_ListByProductionOrder_ReturnsComponents(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "consumed_materials" WHERE production_order_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consumed_materials" WHERE production_order_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "production_order_id", "item_id", "qty_planned"}).
			AddRow(21, 1, 300, 22))

	components := NewConsumedMaterialDAO(db)

	items, err := components.ListByProductionOrder(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != 300 || items[0].QtyPlanned != 22 {
		t.Errorf("items = %+v, want one component for productionOrder 1", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProductionStepDAO_ListByRecipe_ReturnsOperations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "production_steps" WHERE recipe_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "production_steps" WHERE recipe_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "recipe_id", "sequence", "setup_minutes", "time_minutes"}).
			AddRow(11, 1, 10, 15, 45))

	operations := NewProductionStepDAO(db)

	items, err := operations.ListByRecipe(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Sequence != 10 || items[0].TimeMinutes != 45 {
		t.Errorf("items = %+v, want one operation for recipe 1", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestShopTaskDAO_ListByProductionOrder_ReturnsShopTasks(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "shop_tasks" WHERE production_order_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "shop_tasks" WHERE production_order_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "production_order_id", "work_center_id", "state", "planned_minutes"}).
			AddRow(31, 1, 7, ShopTaskStatePlanned, 60))

	workOrders := NewShopTaskDAO(db)

	items, err := workOrders.ListByProductionOrder(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].WorkCenterID != 7 || items[0].PlannedMinutes != 60 {
		t.Errorf("items = %+v, want one shop task for productionOrder 1", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPlanningNeedDAO_ListByRun_ReturnsDemands(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "planning_needs" WHERE planning_run_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "planning_needs" WHERE planning_run_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "planning_run_id", "item_id", "source_type", "qty"}).
			AddRow(41, 1, 100, PlanningNeedSourceSaleOrder, 10))

	demands := NewPlanningNeedDAO(db)

	items, err := demands.ListByRun(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != 100 || items[0].Qty != 10 {
		t.Errorf("items = %+v, want one demand for run 1", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPlannedSupplyDAO_ListByRun_ReturnsOrders(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "planned_supplies" WHERE planning_run_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "planned_supplies" WHERE planning_run_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "planning_run_id", "item_id", "type", "qty"}).
			AddRow(51, 1, 100, PlannedSupplyTypePurchase, 5))

	planned := NewPlannedSupplyDAO(db)

	items, err := planned.ListByRun(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Type != PlannedSupplyTypePurchase || items[0].Qty != 5 {
		t.Errorf("items = %+v, want one planned order for run 1", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPlanningRunDAO_Create_PersistsRun(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "planning_runs"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	runs := NewPlanningRunDAO(db)
	run := &PlanningRun{OrganizationID: helper.Ptr(uint64(7)), HorizonDays: 30, State: PlanningRunStateRunning}

	created, err := runs.Create(ctx, run)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 || created.State != PlanningRunStateRunning {
		t.Errorf("run = %+v, want created running run", created)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDemandPlanDAO_Create_PersistsForecast(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "demand_plans"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	forecasts := NewDemandPlanDAO(db)
	forecast := &DemandPlan{ItemID: 100, ForecastQty: 50}

	created, err := forecasts.Create(ctx, forecast)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 || created.ForecastQty != 50 {
		t.Errorf("forecast = %+v, want created forecast", created)
	}

	query.AssertDBMockDone(t, mock)
}

func TestOutsideProcessingOrderDAO_Create_PersistsOrder(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "outside_processing_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	orders := NewOutsideProcessingOrderDAO(db)
	order := &OutsideProcessingOrder{ProductionOrderID: 1, SupplierID: 44, State: OutsideProcessingStateDraft}

	created, err := orders.Create(ctx, order)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 || created.ProductionOrderID != 1 || created.SupplierID != 44 {
		t.Errorf("order = %+v, want created outside processing order", created)
	}

	query.AssertDBMockDone(t, mock)
}
