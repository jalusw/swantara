package reporting

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func expectRawQuery(mock sqlmock.Sqlmock, raw string, args ...any) *sqlmock.ExpectedQuery {
	expected := strings.Join(strings.Fields(raw), " ")
	var buf strings.Builder
	placeholder := 0
	for _, r := range expected {
		if r == '?' {
			placeholder++
			fmt.Fprintf(&buf, "$%d", placeholder)
			continue
		}
		buf.WriteRune(r)
	}
	values := make([]driver.Value, len(args))
	for i, arg := range args {
		values[i] = arg
	}
	return mock.ExpectQuery(regexp.QuoteMeta(buf.String())).WithArgs(values...)
}

func TestFxRevaluationDAO_ListByOrganization_ReturnsRevaluations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_revaluations" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_revaluations" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "period_id", "name", "date", "state", "total_gain_loss"}).
			AddRow(1, 10, 7, "2026-08", time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC), FxRevaluationStatePosted, 100))

	revaluations := NewFxRevaluationDAO(db)

	items, err := revaluations.ListByOrganization(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 || items[0].TotalGainLoss != 100 {
		t.Errorf("items = %+v, want one posted revaluation", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRevaluationDAO_ListByOrganization_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_revaluations"`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))

	revaluations := NewFxRevaluationDAO(db)

	_, err := revaluations.ListByOrganization(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRevaluationLineDAO_ListByRevaluation_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_revaluation_lines" WHERE revaluation_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_revaluation_lines" WHERE revaluation_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "revaluation_id", "account_id", "currency_code", "foreign_balance", "base_balance", "closing_rate", "gain_loss"}).
			AddRow(11, 3, 22, "USD", 100, 80, 0.8, 20))

	lines := NewFxRevaluationLineDAO(db)

	items, err := lines.ListByRevaluation(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].AccountID != 22 || items[0].CurrencyCode != "USD" {
		t.Errorf("items = %+v, want one USD line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRevaluationLineDAO_ListByRevaluation_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_revaluation_lines"`)).
		WithArgs(3).
		WillReturnError(errors.New("db down"))

	lines := NewFxRevaluationLineDAO(db)

	_, err := lines.ListByRevaluation(ctx, 3)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRevaluationLineDAO_ListOpenByOrg_ReturnsOpenLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "fx_revaluation_lines"."id","fx_revaluation_lines"."created_at","fx_revaluation_lines"."updated_at","fx_revaluation_lines"."deleted_at","fx_revaluation_lines"."revaluation_id","fx_revaluation_lines"."account_id","fx_revaluation_lines"."currency_code","fx_revaluation_lines"."foreign_balance","fx_revaluation_lines"."base_balance","fx_revaluation_lines"."closing_rate","fx_revaluation_lines"."gain_loss","fx_revaluation_lines"."entry_id","fx_revaluation_lines"."reversed" FROM "fx_revaluation_lines" JOIN fx_revaluations ON fx_revaluations.id = fx_revaluation_lines.revaluation_id WHERE fx_revaluations.organization_id = $1 AND fx_revaluation_lines.reversed = $2 AND fx_revaluations.deleted_at IS NULL`)).
		WithArgs(10, false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "revaluation_id", "account_id", "currency_code", "foreign_balance", "base_balance", "closing_rate", "gain_loss", "reversed"}).
			AddRow(11, 3, 22, "USD", 100, 80, 0.8, 20, false))

	lines := NewFxRevaluationLineDAO(db)

	items, err := lines.ListOpenByOrg(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].RevaluationID != 3 || items[0].Reversed {
		t.Errorf("items = %+v, want one open line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRevaluationLineDAO_ListOpenByOrg_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "fx_revaluation_lines"."id","fx_revaluation_lines"."created_at","fx_revaluation_lines"."updated_at","fx_revaluation_lines"."deleted_at","fx_revaluation_lines"."revaluation_id","fx_revaluation_lines"."account_id","fx_revaluation_lines"."currency_code","fx_revaluation_lines"."foreign_balance","fx_revaluation_lines"."base_balance","fx_revaluation_lines"."closing_rate","fx_revaluation_lines"."gain_loss","fx_revaluation_lines"."entry_id","fx_revaluation_lines"."reversed" FROM "fx_revaluation_lines" JOIN fx_revaluations ON fx_revaluations.id = fx_revaluation_lines.revaluation_id WHERE fx_revaluations.organization_id = $1 AND fx_revaluation_lines.reversed = $2 AND fx_revaluations.deleted_at IS NULL`)).
		WithArgs(10, false).
		WillReturnError(errors.New("db down"))

	lines := NewFxRevaluationLineDAO(db)

	_, err := lines.ListOpenByOrg(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccrualDAO_ListByOrganization_ReturnsAccruals(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "accruals" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accruals" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "period_id", "name", "state"}).
			AddRow(1, 10, 7, "Bonus", AccrualStatePosted))

	accruals := NewAccrualDAO(db)

	items, err := accruals.ListByOrganization(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Name == nil || *items[0].Name != "Bonus" {
		t.Errorf("items = %+v, want one posted accrual", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccrualDAO_ListByOrganization_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "accruals"`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))

	accruals := NewAccrualDAO(db)

	_, err := accruals.ListByOrganization(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccrualLineDAO_ListByAccrual_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "accrual_lines" WHERE accrual_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accrual_lines" WHERE accrual_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "accrual_id", "account_id", "name", "debit", "credit"}).
			AddRow(11, 3, 22, "Rent", 100, 0))

	lines := NewAccrualLineDAO(db)

	items, err := lines.ListByAccrual(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].AccountID != 22 || items[0].Debit != 100 {
		t.Errorf("items = %+v, want one debit line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccrualLineDAO_ListByAccrual_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "accrual_lines"`)).
		WithArgs(3).
		WillReturnError(errors.New("db down"))

	lines := NewAccrualLineDAO(db)

	_, err := lines.ListByAccrual(ctx, 3)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
