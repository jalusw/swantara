package returns

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestRMADAO_CreateWithLinesTx_CreatesRMAAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "rmas"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "rma_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	rmas := NewRMADAO(db)
	tx := db.Begin()

	rma := &RMA{OrganizationID: helper.Ptr(uint64(1)), Type: TypeCustomerReturn}
	lines := []*RMALine{{ItemID: 100, Qty: 2, Disposition: DispositionRestock}}

	created, err := rmas.CreateWithLinesTx(ctx, tx, rma, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("rma id = %d, want 1", created.ID)
	}
	if lines[0].RMAID != 1 {
		t.Errorf("line rma_id = %d, want 1", lines[0].RMAID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMADAO_CreateWithLinesTx_PropagatesCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "rmas"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	rmas := NewRMADAO(db)
	tx := db.Begin()

	_, err := rmas.CreateWithLinesTx(ctx, tx, &RMA{}, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMADAO_CreateWithLinesTx_PropagatesLineCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "rmas"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "rma_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	rmas := NewRMADAO(db)
	tx := db.Begin()

	_, err := rmas.CreateWithLinesTx(ctx, tx, &RMA{}, []*RMALine{{ItemID: 100, Qty: 2}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMADAO_UpdateTx_SavesRMA(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "rmas" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rmas := NewRMADAO(db)
	tx := db.Begin()

	rma := &RMA{Base: model.Base{ID: 1}, State: StateConfirmed}
	updated, err := rmas.UpdateTx(ctx, tx, rma)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != rma {
		t.Errorf("UpdateTx returned a different rma")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMADAO_UpdateTx_PropagatesSaveError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "rmas" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	rmas := NewRMADAO(db)
	tx := db.Begin()

	_, err := rmas.UpdateTx(ctx, tx, &RMA{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMALineDAO_ListByRMA_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "rma_lines" WHERE rma_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "rma_lines" WHERE rma_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "rma_id", "item_id", "qty"}).
			AddRow(11, 7, 100, 2).
			AddRow(12, 7, 200, 1))

	lines := NewRMALineDAO(db)

	items, err := lines.ListByRMA(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].ItemID != 100 || items[1].Qty != 1 {
		t.Errorf("items = %+v, want two lines", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMALineDAO_ListByRMA_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "rma_lines" WHERE rma_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	lines := NewRMALineDAO(db)

	_, err := lines.ListByRMA(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMALineDAO_UpdateTx_SavesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "rma_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	lines := NewRMALineDAO(db)
	tx := db.Begin()

	line := &RMALine{Base: model.Base{ID: 11}, StockMovementID: helper.Ptr(uint64(9))}
	updated, err := lines.UpdateTx(ctx, tx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != line {
		t.Errorf("UpdateTx returned a different line")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRMALineDAO_UpdateTx_PropagatesSaveError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "rma_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	lines := NewRMALineDAO(db)
	tx := db.Begin()

	_, err := lines.UpdateTx(ctx, tx, &RMALine{Base: model.Base{ID: 11}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
