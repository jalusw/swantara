package audit

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func callbackDB(t *testing.T, rv any) (*gorm.DB, *gorm.Statement) {
	t.Helper()
	db, _ := query.NewMockDB(t)
	stmt := db.Model(&saleOrder{}).Statement
	if err := stmt.Parse(&saleOrder{}); err != nil {
		t.Fatalf("parse error = %v", err)
	}
	holder := &gorm.DB{Statement: stmt, Config: db.Config}
	holder.Statement.ReflectValue = reflect.ValueOf(rv)
	return holder, stmt
}

func TestGormPlugin_CallbackGuards(t *testing.T) {
	plugin := NewGormPlugin()
	empty := &gorm.DB{Statement: &gorm.Statement{}}

	plugin.beforeUpdate(empty)
	plugin.afterUpdate(empty)
	plugin.beforeDelete(empty)
	plugin.afterDelete(empty)
	plugin.afterCreate(empty)

	db, stmt := callbackDB(t, nil)
	stmt.Schema.Table = "audit_logs"
	holder := &gorm.DB{Statement: stmt, Config: db.Config}
	plugin.beforeUpdate(holder)
	plugin.afterUpdate(holder)
	plugin.beforeDelete(holder)
	plugin.afterDelete(holder)
	plugin.afterCreate(holder)

	db2, stmt2 := callbackDB(t, nil)
	holder2 := &gorm.DB{Statement: stmt2, Config: db2.Config}
	plugin.afterUpdate(holder2)
	plugin.afterDelete(holder2)
}

func TestGormPlugin_InitializeDuplicate(t *testing.T) {
	db, _ := query.NewMockDB(t)
	plugin := NewGormPlugin()
	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("first initialize = %v", err)
	}
	if err := plugin.Initialize(db); err != nil {
		t.Fatalf("second initialize = %v", err)
	}
}

func TestGormPlugin_SnapshotFailure(t *testing.T) {
	db, mock := query.NewMockDB(t)
	db.SkipDefaultTransaction = true
	plugin := NewGormPlugin()

	mock.ExpectQuery(`SELECT .* FROM "sale_orders"`).WillReturnError(errors.New("db down"))
	stmt := db.Model(&saleOrder{}).Statement
	if err := stmt.Parse(&saleOrder{}); err != nil {
		t.Fatal(err)
	}
	holder := &gorm.DB{Statement: stmt, Config: db.Config}
	if _, ok := plugin.snapshot(holder); ok {
		t.Error("expected snapshot failure")
	}
	query.AssertDBMockDone(t, mock)
}

func TestGormPlugin_RecordSavepointFailures(t *testing.T) {
	plugin := NewGormPlugin()
	ctx := context.Background()

	t.Run("savepoint fails", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		mock.ExpectBegin()
		tx := db.Begin()
		mock.ExpectExec(`SAVEPOINT swantara_audit`).WillReturnError(errors.New("no savepoint"))
		mock.ExpectRollback()
		plugin.record(ctx, tx, "sale_orders", 5, ActionInsert, map[string]any{"state": "draft"})
		_ = tx.Rollback()
		query.AssertDBMockDone(t, mock)
	})

	t.Run("release fails", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		mock.ExpectBegin()
		tx := db.Begin()
		mock.ExpectExec(`SAVEPOINT swantara_audit`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`RELEASE SAVEPOINT swantara_audit`).WillReturnError(errors.New("no release"))
		mock.ExpectCommit()
		plugin.record(ctx, tx, "sale_orders", 5, ActionInsert, map[string]any{"state": "draft"})
		if err := tx.Commit().Error; err != nil {
			t.Fatal(err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("write and rollback fail", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		db.SkipDefaultTransaction = true
		mock.ExpectBegin()
		tx := db.Begin()
		mock.ExpectExec(`SAVEPOINT swantara_audit`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectQuery(`INSERT INTO "audit_logs" .* RETURNING "id"`).
			WillReturnError(errors.New("db down"))
		mock.ExpectExec(`ROLLBACK TO SAVEPOINT swantara_audit`).WillReturnError(errors.New("no rollback"))
		mock.ExpectExec(`RELEASE SAVEPOINT swantara_audit`).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectRollback()
		plugin.record(ctx, tx, "sale_orders", 5, ActionInsert, map[string]any{"state": "draft"})
		_ = tx.Rollback()
		query.AssertDBMockDone(t, mock)
	})
}
