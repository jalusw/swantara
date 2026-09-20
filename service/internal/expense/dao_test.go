package expense

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

func TestNewExpenseReportDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewExpenseReportDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewExpenseLineDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewExpenseLineDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestExpenseReportDAO_CreateTx_InsertsReport(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "expense_reports"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	reports := NewExpenseReportDAO(db)
	report := &ExpenseReport{OrganizationID: helper.Ptr(uint64(10)), Name: "Travel", EmployeeID: 3, State: ExpenseStateDraft}

	created, err := reports.CreateTx(ctx, db, report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("report id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestExpenseReportDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "expense_reports"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	reports := NewExpenseReportDAO(db)
	_, err := reports.CreateTx(ctx, db, &ExpenseReport{OrganizationID: helper.Ptr(uint64(10))})

	if helper.AssertError(t, err, true, nil) {
		return
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseReportDAO_UpdateTx_SavesReport(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "expense_reports" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	reports := NewExpenseReportDAO(db)
	report := &ExpenseReport{Base: model.Base{ID: 1}, State: ExpenseStateSubmitted}

	updated, err := reports.UpdateTx(ctx, db, report)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != report {
		t.Errorf("UpdateTx returned a different report")
	}

	query.AssertDBMockDone(t, mock)
}

func TestExpenseReportDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "expense_reports" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	reports := NewExpenseReportDAO(db)
	_, err := reports.UpdateTx(ctx, db, &ExpenseReport{Base: model.Base{ID: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseLineDAO_CreateTx_InsertsLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "expense_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	lines := NewExpenseLineDAO(db)
	line := &ExpenseLine{ReportID: 1, Quantity: 2, UnitPrice: 25}

	created, err := lines.CreateTx(ctx, db, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 11 {
		t.Errorf("line id = %d, want 11", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestExpenseLineDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "expense_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	lines := NewExpenseLineDAO(db)
	_, err := lines.CreateTx(ctx, db, &ExpenseLine{ReportID: 1})

	if helper.AssertError(t, err, true, nil) {
		return
	}
	query.AssertDBMockDone(t, mock)
}

func TestExpenseLineDAO_ListByReport_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "expense_lines" WHERE report_id = $1`)).
		WithArgs(uint64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "expense_lines" WHERE report_id = $1`)).
		WithArgs(uint64(7)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "report_id", "quantity", "amount"}).
			AddRow(1, 7, 3, 300).
			AddRow(2, 7, 4, 400))

	lines := NewExpenseLineDAO(db)

	items, err := lines.ListByReport(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Quantity != 3 || items[1].Quantity != 4 {
		t.Errorf("items = %+v, want two lines qty 3/4", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestExpenseLineDAO_ListByReport_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "expense_lines"`)).
		WithArgs(uint64(7)).
		WillReturnError(errors.New("db down"))

	lines := NewExpenseLineDAO(db)
	_, err := lines.ListByReport(ctx, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}
	query.AssertDBMockDone(t, mock)
}
