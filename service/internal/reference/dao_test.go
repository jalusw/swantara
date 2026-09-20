package reference

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestPaymentTermDAO_ListLines_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	d := NewPaymentTermDAO(db)

	_, err := d.ListLines(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermDAO_ReplaceLines_PropagatesListError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	d := NewPaymentTermDAO(db)

	err := d.ReplaceLines(ctx, 1, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermDAO_ReplaceLines_PropagatesHardDeleteError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_term_id"}).AddRow(7, 1))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "payment_term_lines"`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	d := NewPaymentTermDAO(db)

	err := d.ReplaceLines(ctx, 1, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermDAO_ReplaceLines_PropagatesCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_term_id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_term_lines"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	d := NewPaymentTermDAO(db)

	err := d.ReplaceLines(ctx, 1, []*PaymentTermLine{{Sequence: 10, ValueType: "percent", Value: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermDAO_DeleteWithLines_PropagatesListError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))

	d := NewPaymentTermDAO(db)

	err := d.DeleteWithLines(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentTermDAO_DeleteWithLines_PropagatesHardDeleteError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_term_lines" WHERE payment_term_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "payment_term_id"}).AddRow(7, 1))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "payment_term_lines"`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	d := NewPaymentTermDAO(db)

	err := d.DeleteWithLines(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
