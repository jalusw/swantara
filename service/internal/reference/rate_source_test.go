package reference

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestFxRateSource_OrganizationSpecific(t *testing.T) {
	db, mock := query.NewMockDB(t)
	date := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	now := date

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id = $4`)).
		WithArgs("USD", "spot", date, 5).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id = $4 ORDER BY valid_from DESC LIMIT $5`)).
		WithArgs("USD", "spot", date, 5, 1).
		WillReturnRows(fxRateRows().AddRow(1, "USD", 5, 1.5, "spot", date, now, now, nil))

	source := NewFxRateSource(dao.NewBase[FxRate](db))

	rate, err := source.Rate(context.Background(), "USD", 5, amount.RateSpot, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Equal(amount.FromFloat64(1.5)) {
		t.Errorf("rate = %s, want 1.5", rate)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateSource_FallsBackToGlobalRate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	date := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)
	now := date

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id = $4`)).
		WithArgs("USD", "spot", date, 5).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id = $4 ORDER BY valid_from DESC LIMIT $5`)).
		WithArgs("USD", "spot", date, 5, 1).
		WillReturnRows(fxRateRows())

	expectedGlobal := `SELECT count(*) FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id IS NULL`
	mock.ExpectQuery(regexp.QuoteMeta(expectedGlobal)).
		WithArgs("USD", "spot", date).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id IS NULL ORDER BY valid_from DESC LIMIT $4`)).
		WithArgs("USD", "spot", date, 1).
		WillReturnRows(fxRateRows().AddRow(2, "USD", nil, 1.4, "spot", date, now, now, nil))

	source := NewFxRateSource(dao.NewBase[FxRate](db))

	rate, err := source.Rate(context.Background(), "USD", 5, amount.RateSpot, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !rate.Equal(amount.FromFloat64(1.4)) {
		t.Errorf("rate = %s, want 1.4", rate)
	}

	query.AssertDBMockDone(t, mock)
}

func TestFxRateSource_NotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	date := time.Date(2026, time.January, 15, 12, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id = $4`)).
		WithArgs("USD", "spot", date, 5).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id = $4 ORDER BY valid_from DESC LIMIT $5`)).
		WithArgs("USD", "spot", date, 5, 1).
		WillReturnRows(fxRateRows())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id IS NULL`)).
		WithArgs("USD", "spot", date).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE currency_code = $1 AND rate_type = $2 AND valid_from <= $3 AND organization_id IS NULL ORDER BY valid_from DESC LIMIT $4`)).
		WithArgs("USD", "spot", date, 1).
		WillReturnRows(fxRateRows())

	source := NewFxRateSource(dao.NewBase[FxRate](db))

	_, err := source.Rate(context.Background(), "USD", 5, amount.RateSpot, date)

	if !helper.AssertError(t, err, true, ErrRateNotFound) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func fxRateRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "currency_code", "organization_id", "rate", "rate_type", "valid_from", "created_at", "updated_at", "deleted_at"})
}
