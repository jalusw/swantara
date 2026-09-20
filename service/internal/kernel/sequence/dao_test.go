package sequence

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestReserve_AllocatesNextNumber(t *testing.T) {
	db, mock := query.NewMockDB(t)
	now := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

	columns := []string{"id", "organization_id", "code", "prefix", "suffix", "next_number", "padding", "reset_period", "created_at", "updated_at", "deleted_at"}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "doc_sequences" WHERE organization_id = $1 AND code = $2 ORDER BY "doc_sequences"."id" LIMIT $3 FOR UPDATE`)).
		WithArgs(1, "SALE_ORDER", 1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(9, 1, "SALE_ORDER", "SO/", "", 42, 5, "never", now, now, nil))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "doc_sequences" SET "next_number"=$1,"updated_at"=$2 WHERE id = $3`)).
		WithArgs(int64(43), now, int64(9)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dao := NewDAO(db)
	reservation, err := dao.Reserve(context.Background(), 1, "SALE_ORDER", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reservation.Value != 42 {
		t.Errorf("reserved value = %d, want 42", reservation.Value)
	}
	if reservation.Number != "SO/00042" {
		t.Errorf("document number = %q, want SO/00042", reservation.Number)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReserve_ResetsOnNewPeriod(t *testing.T) {
	db, mock := query.NewMockDB(t)
	january := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	february := time.Date(2026, time.February, 1, 9, 0, 0, 0, time.UTC)

	columns := []string{"id", "organization_id", "code", "prefix", "suffix", "next_number", "padding", "reset_period", "created_at", "updated_at", "deleted_at"}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "doc_sequences" WHERE organization_id = $1 AND code = $2 ORDER BY "doc_sequences"."id" LIMIT $3 FOR UPDATE`)).
		WithArgs(1, "INVOICE", 1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(9, 1, "INVOICE", "INV/", "", 7, 4, "monthly", january, january, nil))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "doc_sequences" SET "next_number"=$1,"updated_at"=$2 WHERE id = $3`)).
		WithArgs(int64(2), february, int64(9)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dao := NewDAO(db)
	reservation, err := dao.Reserve(context.Background(), 1, "INVOICE", february)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reservation.Value != 1 {
		t.Errorf("reserved value after reset = %d, want 1", reservation.Value)
	}
	if reservation.Number != "INV/0001" {
		t.Errorf("document number = %q, want INV/0001", reservation.Number)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReserve_UnknownSequence(t *testing.T) {
	db, mock := query.NewMockDB(t)
	now := time.Now()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "doc_sequences" WHERE organization_id = $1 AND code = $2 ORDER BY "doc_sequences"."id" LIMIT $3 FOR UPDATE`)).
		WithArgs(1, "MISSING", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectRollback()

	dao := NewDAO(db)
	_, err := dao.Reserve(context.Background(), 1, "MISSING", now)
	if helper.AssertError(t, err, true, ErrSequenceNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
