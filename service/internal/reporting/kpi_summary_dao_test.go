package reporting

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestKpiSummaryDAO_OpeningBalances_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	before := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT l.account_id, COALESCE(SUM(l.debit), 0) AS debit, COALESCE(SUM(l.credit), 0) AS credit
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date < ? AND m.deleted_at IS NULL
		GROUP BY l.account_id`, 1, before).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "debit", "credit"}).
			AddRow(11, 100, 0).
			AddRow(12, 0, 100))

	summaries := NewKpiSummaryDAO(db)

	rows, err := summaries.OpeningBalances(ctx, 1, before)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 || rows[0].AccountID != 11 || rows[1].Credit != 100 {
		t.Errorf("rows = %+v, want opening balances", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_OpeningBalances_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	summaries := NewKpiSummaryDAO(db)

	_, err := summaries.OpeningBalances(ctx, 1, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_PeriodActivity_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT l.account_id, COALESCE(SUM(l.debit), 0) AS debit, COALESCE(SUM(l.credit), 0) AS credit
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND m.deleted_at IS NULL
		GROUP BY l.account_id`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "debit", "credit"}).
			AddRow(11, 50, 0))

	summaries := NewKpiSummaryDAO(db)

	rows, err := summaries.PeriodActivity(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].Debit != 50 {
		t.Errorf("rows = %+v, want one activity row", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_PeriodActivity_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	summaries := NewKpiSummaryDAO(db)

	_, err := summaries.PeriodActivity(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_ReplacePeriod_ReplacesRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "kpi_account_summaries"`)).
		WithArgs(1, 7).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "kpi_account_summaries"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	summaries := NewKpiSummaryDAO(db)
	rows := []KpiAccountSummary{{OrganizationID: 1, TaxPeriodID: 7, AccountID: 11}}

	err := summaries.ReplacePeriod(ctx, 1, 7, rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_ReplacePeriod_AllowsEmptyRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "kpi_account_summaries"`)).
		WithArgs(1, 7).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	summaries := NewKpiSummaryDAO(db)

	err := summaries.ReplacePeriod(ctx, 1, 7, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_ReplacePeriod_PropagatesBeginError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin().WillReturnError(errors.New("begin failed"))

	summaries := NewKpiSummaryDAO(db)

	err := summaries.ReplacePeriod(ctx, 1, 7, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_ReplacePeriod_PropagatesDeleteError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "kpi_account_summaries"`)).
		WithArgs(1, 7).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	summaries := NewKpiSummaryDAO(db)

	err := summaries.ReplacePeriod(ctx, 1, 7, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_ReplacePeriod_PropagatesCreateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "kpi_account_summaries"`)).
		WithArgs(1, 7).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "kpi_account_summaries"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	summaries := NewKpiSummaryDAO(db)
	rows := []KpiAccountSummary{{OrganizationID: 1, TaxPeriodID: 7, AccountID: 11}}

	err := summaries.ReplacePeriod(ctx, 1, 7, rows)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_HasPeriod_ReturnsTrue(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "kpi_account_summaries" WHERE organization_id = $1 AND tax_period_id = $2`)).
		WithArgs(1, 7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	summaries := NewKpiSummaryDAO(db)

	has, err := summaries.HasPeriod(ctx, 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Errorf("has = false, want true")
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_HasPeriod_ReturnsFalse(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "kpi_account_summaries"`)).
		WithArgs(1, 7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	summaries := NewKpiSummaryDAO(db)

	has, err := summaries.HasPeriod(ctx, 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Errorf("has = true, want false")
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_HasPeriod_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "kpi_account_summaries"`)).
		WithArgs(1, 7).
		WillReturnError(errors.New("db down"))

	summaries := NewKpiSummaryDAO(db)

	_, err := summaries.HasPeriod(ctx, 1, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_TrialBalance_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	expectRawQuery(mock, `
		SELECT s.account_id,
		       a.code,
		       a.name,
		       a.type AS account_type,
		       s.opening_debit,
		       s.opening_credit,
		       s.period_debit,
		       s.period_credit,
		       s.opening_debit + s.period_debit AS closing_debit,
		       s.opening_credit + s.period_credit AS closing_credit
		FROM kpi_account_summaries s
		JOIN accounts a ON a.id = s.account_id AND a.deleted_at IS NULL
		WHERE s.organization_id = ? AND s.tax_period_id = ?
		ORDER BY a.code`, 1, 7).
		WillReturnRows(sqlmock.NewRows([]string{
			"account_id", "code", "name", "account_type", "opening_debit", "opening_credit", "period_debit", "period_credit", "closing_debit", "closing_credit",
		}).AddRow(11, "1100", "Cash", "asset", 100, 0, 50, 0, 150, 0))

	summaries := NewKpiSummaryDAO(db)

	rows, err := summaries.TrialBalance(ctx, 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].ClosingDebit != 150 {
		t.Errorf("rows = %+v, want one closing debit row", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_TrialBalance_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	summaries := NewKpiSummaryDAO(db)

	_, err := summaries.TrialBalance(ctx, 1, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_ListPeriod_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "kpi_account_summaries" WHERE organization_id = $1 AND tax_period_id = $2 ORDER BY account_id`)).
		WithArgs(1, 7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "tax_period_id", "account_id", "opening_debit", "period_credit"}).
			AddRow(1, 1, 7, 11, 100, 50))

	summaries := NewKpiSummaryDAO(db)

	rows, err := summaries.ListPeriod(ctx, 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].AccountID != 11 || rows[0].OpeningDebit != 100 {
		t.Errorf("rows = %+v, want one summary row", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestKpiSummaryDAO_ListPeriod_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "kpi_account_summaries"`)).
		WithArgs(1, 7).
		WillReturnError(errors.New("db down"))

	summaries := NewKpiSummaryDAO(db)

	_, err := summaries.ListPeriod(ctx, 1, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
