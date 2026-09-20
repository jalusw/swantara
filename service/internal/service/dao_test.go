package service

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestNewEquipmentDAO_Find_ReturnsNilWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "equipments" WHERE id = $1`)).
		WillReturnError(gorm.ErrRecordNotFound)

	equipments := NewEquipmentDAO(db)
	equipment, err := equipments.Find(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if equipment != nil {
		t.Errorf("equipment = %+v, want nil when missing", equipment)
	}

	query.AssertDBMockDone(t, mock)
}

func TestServiceContractDAO_UpdateTx_SavesContract(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "service_contracts" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	contracts := NewServiceContractDAO(db)
	tx := db.Begin()

	contract := &ServiceContract{Base: model.Base{ID: 1}, Name: "Gold SLA", State: ContractStateActive}
	updated, err := contracts.UpdateTx(ctx, tx, contract)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != contract {
		t.Errorf("UpdateTx returned a different contract")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestServiceContractDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "service_contracts" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	contracts := NewServiceContractDAO(db)
	_, err := contracts.UpdateTx(ctx, db, &ServiceContract{Base: model.Base{ID: 1}})
	helper.AssertError(t, err, true, nil)

	query.AssertDBMockDone(t, mock)
}

func TestServiceOrderDAO_UpdateTx_SavesOrder(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "service_orders" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	orders := NewServiceOrderDAO(db)
	tx := db.Begin()

	order := &ServiceOrder{Base: model.Base{ID: 1}, Name: "Fix printer", State: OrderStateDone}
	updated, err := orders.UpdateTx(ctx, tx, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != order {
		t.Errorf("UpdateTx returned a different order")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestServiceOrderDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "service_orders" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	orders := NewServiceOrderDAO(db)
	_, err := orders.UpdateTx(ctx, db, &ServiceOrder{Base: model.Base{ID: 1}})
	helper.AssertError(t, err, true, nil)

	query.AssertDBMockDone(t, mock)
}

func TestServiceOrderLineDAO_CreateTx_CreatesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "service_order_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	lines := NewServiceOrderLineDAO(db)
	tx := db.Begin()

	line := &ServiceOrderLine{ServiceOrderID: 8, Type: LineTypePart, Qty: 2}
	created, err := lines.CreateTx(ctx, tx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 11 {
		t.Errorf("line id = %d, want 11", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestServiceOrderLineDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "service_order_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	lines := NewServiceOrderLineDAO(db)
	_, err := lines.CreateTx(ctx, db, &ServiceOrderLine{Type: LineTypePart})
	helper.AssertError(t, err, true, nil)

	query.AssertDBMockDone(t, mock)
}

func TestServiceOrderLineDAO_ListByOrder_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "service_order_lines" WHERE service_order_id = $1`)).
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "service_order_lines" WHERE service_order_id = $1`)).
		WithArgs(8).
		WillReturnRows(sqlmock.NewRows([]string{"id", "service_order_id", "type", "qty"}).
			AddRow(11, 8, LineTypePart, 2).
			AddRow(12, 8, LineTypeLabor, 1))

	lines := NewServiceOrderLineDAO(db)
	items, err := lines.ListByOrder(ctx, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Type != LineTypePart || items[1].Qty != 1 {
		t.Errorf("items = %+v, want two lines", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestServiceOrderLineDAO_ListByOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "service_order_lines" WHERE service_order_id = $1`)).
		WithArgs(8).
		WillReturnError(errors.New("db down"))

	lines := NewServiceOrderLineDAO(db)
	_, err := lines.ListByOrder(ctx, 8)
	helper.AssertError(t, err, true, nil)

	query.AssertDBMockDone(t, mock)
}

func TestMaintenancePlanDAO_ListDue_ReturnsDuePlans(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	due := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "maintenance_plans" JOIN equipments ON equipments.id = maintenance_plans.equipment_id WHERE equipments.organization_id = $1`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "equipment_id", "name", "interval_days", "next_due", "active"}).
			AddRow(1, 5, "Quarterly", 30, due, true))

	plans := NewMaintenancePlanDAO(db)
	items, err := plans.ListDue(ctx, 10, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Name != "Quarterly" || !items[0].NextDue.Equal(due) {
		t.Errorf("items = %+v, want one due plan", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestMaintenancePlanDAO_ListDue_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "maintenance_plans" JOIN equipments ON equipments.id = maintenance_plans.equipment_id WHERE equipments.organization_id = $1`)).
		WillReturnError(errors.New("db down"))

	plans := NewMaintenancePlanDAO(db)
	_, err := plans.ListDue(ctx, 10, asOf)
	helper.AssertError(t, err, true, nil)

	query.AssertDBMockDone(t, mock)
}

func TestMaintenancePlanDAO_UpdateTx_SavesPlan(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "maintenance_plans" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	plans := NewMaintenancePlanDAO(db)
	tx := db.Begin()

	plan := &MaintenancePlan{Base: model.Base{ID: 1}, Name: "Quarterly", IntervalDays: 30}
	updated, err := plans.UpdateTx(ctx, tx, plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != plan {
		t.Errorf("UpdateTx returned a different plan")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestMaintenancePlanDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "maintenance_plans" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	plans := NewMaintenancePlanDAO(db)
	_, err := plans.UpdateTx(ctx, db, &MaintenancePlan{Base: model.Base{ID: 1}})
	helper.AssertError(t, err, true, nil)

	query.AssertDBMockDone(t, mock)
}
