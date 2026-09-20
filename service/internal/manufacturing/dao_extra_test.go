package manufacturing

import (
	"context"
	"database/sql/driver"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func listPair(mock sqlmock.Sqlmock, table string, id uint64) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "` + table + `"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "` + table + `"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
}

func listPairError(mock sqlmock.Sqlmock, table string, dbErr error) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "` + table + `"`)).WillReturnError(dbErr)
}

func TestManufacturingDAO_Constructors(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewRecipeDAO(db) == nil {
		t.Error("recipes = nil")
	}
	if NewRecipeLineDAO(db) == nil {
		t.Error("lines = nil")
	}
	if NewProductionOrderDAO(db) == nil {
		t.Error("orders = nil")
	}
	if NewConsumedMaterialDAO(db) == nil {
		t.Error("components = nil")
	}
	if NewProductionStepDAO(db) == nil {
		t.Error("operations = nil")
	}
	if NewShopTaskDAO(db) == nil {
		t.Error("shop tasks = nil")
	}
	if NewPlanningRunDAO(db) == nil {
		t.Error("runs = nil")
	}
	if NewPlanningNeedDAO(db) == nil {
		t.Error("demands = nil")
	}
	if NewPlannedSupplyDAO(db) == nil {
		t.Error("planned = nil")
	}
	if NewDemandPlanDAO(db) == nil {
		t.Error("forecasts = nil")
	}
	if NewOutsideProcessingOrderDAO(db) == nil {
		t.Error("subcontracts = nil")
	}
}

func TestManufacturingDAO_ListQueries(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("lists by item", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		listPair(mock, "recipes", 1)
		items, err := NewRecipeDAO(db).ListByItem(ctx, 100)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists recipe lines", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		listPair(mock, "recipe_lines", 1)
		items, err := NewRecipeLineDAO(db).ListByRecipe(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists productionOrder components", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		listPair(mock, "consumed_materials", 1)
		items, err := NewConsumedMaterialDAO(db).ListByProductionOrder(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists routing operations", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		listPair(mock, "production_steps", 1)
		items, err := NewProductionStepDAO(db).ListByRecipe(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists shop tasks", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		listPair(mock, "shop_tasks", 1)
		items, err := NewShopTaskDAO(db).ListByProductionOrder(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("lists demands and planned", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		listPair(mock, "planning_needs", 1)
		items, err := NewPlanningNeedDAO(db).ListByRun(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)

		db, mock = query.NewMockDB(t)
		listPair(mock, "planned_supplies", 1)
		planned, err := NewPlannedSupplyDAO(db).ListByRun(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(planned) != 1 {
			t.Errorf("len = %d", len(planned))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("propagates list errors", func(t *testing.T) {
		tables := []struct {
			table string
			call  func() error
		}{
			{table: "recipes", call: func() error {
				db, mock := query.NewMockDB(t)
				listPairError(mock, "recipes", dbErr)
				_, err := NewRecipeDAO(db).ListByItem(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "recipe_lines", call: func() error {
				db, mock := query.NewMockDB(t)
				listPairError(mock, "recipe_lines", dbErr)
				_, err := NewRecipeLineDAO(db).ListByRecipe(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "consumed_materials", call: func() error {
				db, mock := query.NewMockDB(t)
				listPairError(mock, "consumed_materials", dbErr)
				_, err := NewConsumedMaterialDAO(db).ListByProductionOrder(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "production_steps", call: func() error {
				db, mock := query.NewMockDB(t)
				listPairError(mock, "production_steps", dbErr)
				_, err := NewProductionStepDAO(db).ListByRecipe(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "shop_tasks", call: func() error {
				db, mock := query.NewMockDB(t)
				listPairError(mock, "shop_tasks", dbErr)
				_, err := NewShopTaskDAO(db).ListByProductionOrder(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "planning_needs", call: func() error {
				db, mock := query.NewMockDB(t)
				listPairError(mock, "planning_needs", dbErr)
				_, err := NewPlanningNeedDAO(db).ListByRun(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
			{table: "planned_supplies", call: func() error {
				db, mock := query.NewMockDB(t)
				listPairError(mock, "planned_supplies", dbErr)
				_, err := NewPlannedSupplyDAO(db).ListByRun(ctx, 1)
				query.AssertDBMockDone(t, mock)
				return err
			}},
		}
		for _, tt := range tables {
			t.Run(tt.table, func(t *testing.T) {
				helper.AssertError(t, tt.call(), true, dbErr)
			})
		}
	})
}

func TestManufacturingDAO_CreateWith(t *testing.T) {
	ctx := context.Background()

	t.Run("creates recipe with lines", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipes"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipe_lines"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
		mock.ExpectCommit()

		recipe := &Recipe{}
		lines := []*RecipeLine{{Qty: 2}}
		got, err := NewRecipeDAO(db).CreateWithLines(ctx, recipe, lines)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 || lines[0].RecipeID != 1 {
			t.Errorf("recipe = %+v line = %+v", got, lines[0])
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("rolls back recipe on error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "recipes"`)).WillReturnError(errors.New("recipe down"))
		mock.ExpectRollback()

		_, err := NewRecipeDAO(db).CreateWithLines(ctx, &Recipe{}, []*RecipeLine{{}})
		if err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("creates productionOrder with components", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "production_orders"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consumed_materials"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
		mock.ExpectCommit()

		productionOrder := &ProductionOrder{}
		components := []*ConsumedMaterial{{QtyPlanned: 5}}
		got, err := NewProductionOrderDAO(db).CreateWithComponents(ctx, productionOrder, components)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 || components[0].ProductionOrderID != 1 {
			t.Errorf("productionOrder = %+v component = %+v", got, components[0])
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("rolls back productionOrder on component error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "production_orders"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consumed_materials"`)).WillReturnError(errors.New("comp down"))
		mock.ExpectRollback()

		_, err := NewProductionOrderDAO(db).CreateWithComponents(ctx, &ProductionOrder{}, []*ConsumedMaterial{{}})
		if err == nil {
			t.Error("expected error, got nil")
		}
		query.AssertDBMockDone(t, mock)
	})
}

var _ = driver.Value(nil)
