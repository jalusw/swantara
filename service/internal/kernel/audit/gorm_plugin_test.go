package audit

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestGormPlugin_InitializeAndName(t *testing.T) {
	plugin := NewGormPlugin()
	if plugin.Name() != "swantara:audit" {
		t.Errorf("Name() = %q, want swantara:audit", plugin.Name())
	}
	if plugin.skipped("audit_logs") != true || plugin.skipped("sale_orders") != false {
		t.Error("skipped() produced unexpected result")
	}

	db, _ := query.NewMockDB(t)
	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}
}

func TestGormPlugin_AfterCreate_RecordsInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()
	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	mock.ExpectQuery(`INSERT INTO "sale_orders" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	result := db.Create(&saleOrder{Code: "SO/00042", State: "draft", AmountTotal: "1000.00"})
	if result.Error != nil {
		t.Fatalf("create error = %v", result.Error)
	}
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_AfterCreate_EmptyDiffSkipsAudit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()
	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	mock.ExpectExec(`INSERT INTO "plain_records" .*`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result := db.Create(&plainRecord{Name: ""})
	if result.Error != nil {
		t.Fatalf("create error = %v", result.Error)
	}
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_AfterUpdate_RecordsDiff(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()
	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	snapshotPattern := `SELECT \* FROM "sale_orders" WHERE .*`
	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "code", "state", "amount_total"}).
		AddRow(9, nil, nil, nil, "SO/00009", "draft", "100.00")
	mock.ExpectQuery(snapshotPattern).
		WillReturnRows(rows)
	mock.ExpectExec(`UPDATE "sale_orders" SET .* WHERE .*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(snapshotPattern).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "code", "state", "amount_total"}).
			AddRow(9, nil, nil, nil, "SO/00009", "confirmed", "100.00"))
	mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	result := db.Model(&saleOrder{Base: model.Base{ID: 9}}).Update("state", "confirmed")
	if result.Error != nil {
		t.Fatalf("update error = %v", result.Error)
	}
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_AfterDelete_RecordsDiff(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()
	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	snapshotPattern := `SELECT \* FROM "sale_orders" WHERE .*`
	mock.ExpectQuery(snapshotPattern).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "code", "state", "amount_total"}).
			AddRow(3, nil, nil, nil, "SO/00003", "cancelled", "50.00"))
	mock.ExpectExec(`DELETE FROM "sale_orders" WHERE .*`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	result := db.Delete(&saleOrder{Base: model.Base{ID: 3}})
	if result.Error != nil {
		t.Fatalf("delete error = %v", result.Error)
	}
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_pkFromReflect(t *testing.T) {
	db, _ := query.NewMockDB(t)
	plugin := NewGormPlugin()

	withSchema := func(rv reflect.Value) *gorm.Statement {
		stmt := db.Model(&saleOrder{}).Statement
		if err := stmt.Parse(&saleOrder{}); err != nil {
			t.Fatalf("parse error = %v", err)
		}
		stmt.ReflectValue = rv
		return stmt
	}

	if got := plugin.pkFromReflect(withSchema(reflect.ValueOf(&saleOrder{Base: model.Base{ID: 7}}))); got == nil || got.(uint64) != 7 {
		t.Errorf("pkFromReflect(struct with id) = %v, want 7", got)
	}
	if got := plugin.pkFromReflect(withSchema(reflect.ValueOf(&saleOrder{}))); got != nil {
		t.Errorf("pkFromReflect(zero id) = %v, want nil", got)
	}
	if got := plugin.pkFromReflect(withSchema(reflect.ValueOf(&[]saleOrder{{Base: model.Base{ID: 1}}}))); got != nil {
		t.Errorf("pkFromReflect(slice) = %v, want nil", got)
	}
	if got := plugin.pkFromReflect(withSchema(reflect.ValueOf((*saleOrder)(nil)))); got != nil {
		t.Errorf("pkFromReflect(nil ptr) = %v, want nil", got)
	}
	if got := plugin.pkFromReflect(withSchema(reflect.ValueOf(map[string]string{"a": "b"}))); got != nil {
		t.Errorf("pkFromReflect(non-struct) = %v, want nil", got)
	}
	if got := plugin.pkFromReflect(&gorm.Statement{}); got != nil {
		t.Errorf("pkFromReflect(no schema) = %v, want nil", got)
	}
}

func TestGormPlugin_Record_EmptyDiffSkipsWrite(t *testing.T) {
	db, mock := query.NewMockDB(t)
	plugin := NewGormPlugin()

	plugin.record(context.Background(), db, "sale_orders", 5, ActionInsert, map[string]any{})
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_Record_WritesAudit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()

	mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	plugin.record(context.Background(), db, "sale_orders", 5, ActionUpdate, map[string]any{"state": "confirmed"})
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_Record_WarnsOnWriteFailure(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()

	mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
		WillReturnError(errors.New("db down"))

	plugin.record(context.Background(), db, "sale_orders", 5, ActionInsert, map[string]any{"state": "draft"})
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_Record_UsesSavepointInsideTx(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()

	mock.ExpectBegin()
	tx := db.Begin()
	mock.ExpectExec(`SAVEPOINT swantara_audit`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec(`RELEASE SAVEPOINT swantara_audit`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	plugin.record(context.Background(), tx, "sale_orders", 5, ActionInsert, map[string]any{"state": "draft"})
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit error = %v", err)
	}
	query.AssertDBMockDone(t, mock)
}

var _ = context.Background

func TestGormPlugin_Helpers(t *testing.T) {
	plugin := NewGormPlugin()

	if _, ok := plugin.idOf(&saleOrder{Base: model.Base{ID: 5}}); !ok {
		t.Error("idOf() should return true for valid entity")
	}
	if _, ok := plugin.idOf(&plainRecord{Name: "x"}); ok {
		t.Error("idOf() should fail for struct without ID field")
	}
	if _, ok := plugin.idOf(nil); ok {
		t.Error("idOf() should fail for nil")
	}
	if _, ok := plugin.idOf(5); ok {
		t.Error("idOf() should fail for non-struct")
	}

	rows := rowsOf(reflect.ValueOf(&saleOrder{Base: model.Base{ID: 5}}))
	if len(rows) != 1 {
		t.Errorf("rowsOf(struct) = %d rows, want 1", len(rows))
	}
	rows = rowsOf(reflect.ValueOf([]*saleOrder{{Base: model.Base{ID: 1}}, {Base: model.Base{ID: 2}}}))
	if len(rows) != 2 {
		t.Errorf("rowsOf(slice) = %d rows, want 2", len(rows))
	}
	rows = rowsOf(reflect.ValueOf([]saleOrder{{Base: model.Base{ID: 1}}}))
	if len(rows) != 1 {
		t.Errorf("rowsOf(value slice) = %d rows, want 1", len(rows))
	}
	if rows := rowsOf(reflect.ValueOf(nil)); rows != nil {
		t.Errorf("rowsOf(nil) = %v, want nil", rows)
	}
	if rows := rowsOf(reflect.ValueOf("not-a-struct")); rows != nil {
		t.Errorf("rowsOf(string) = %v, want nil", rows)
	}
}
