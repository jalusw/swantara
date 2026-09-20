package handler

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

var fxRateColumns = []string{"id", "created_at", "updated_at", "deleted_at", "currency_code", "organization_id", "rate", "rate_type", "valid_from"}

type fxRateSourceStub struct {
	rateFunc func(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error)
}

func (m fxRateSourceStub) Rate(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error) {
	return m.rateFunc(ctx, currencyCode, organizationID, rateType, date)
}

func fxRateHandlerTest(t *testing.T, rates dao.Base[reference.FxRate], svc reference.FxRateService, source amount.RateSource) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewFxRateHandler(svc, source)
		h.Register(api, guards)
	})
}

func fxRateHandlerTestNoTenant(t *testing.T, rates dao.Base[reference.FxRate], svc reference.FxRateService, source amount.RateSource) *fiber.App {
	t.Helper()
	return referenceTestApp(t, false, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewFxRateHandler(svc, source)
		h.Register(api, guards)
	})
}

func fxRateTestSvc(rates dao.Base[reference.FxRate], currencies dao.Base[reference.Currency]) reference.FxRateService {
	return reference.NewFxRateService(rates, currencies)
}

func TestFxRateHandler_List_ReturnsRates(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTestNoTenant(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "fx_rates" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Get_ReturnsRate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Get_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(99), 15000.0, "spot", timeNow()))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodGet, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Create_CreatesRate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "USD"))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "fx_rates"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"USD","rate":15000,"rate_type":"spot","valid_from":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"US","rate":0,"valid_from":""}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Create_RejectsInvalidDate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"USD","rate":15000,"valid_from":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Create_MapsUnknownCurrency(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("XXX", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"XXX","rate":15000,"valid_from":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Create_MapsInvalidRateType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"USD","rate":15000,"rate_type":"monthly","valid_from":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Create_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnError(errors.New("db down"))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"USD","rate":15000,"valid_from":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Create_ReturnsServerErrorOnInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "USD"))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "fx_rates"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"USD","rate":15000,"valid_from":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Resolve_ResolvesRate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	source := fxRateSourceStub{
		rateFunc: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.FromFloat64(15000), nil
		},
	}
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), source)

	body := `{"currency_code":"USD","rate_type":"spot","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Resolve_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"US","date":""}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Resolve_RejectsInvalidDate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"currency_code":"USD","date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Resolve_MapsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	source := fxRateSourceStub{
		rateFunc: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.Amount{}, reference.ErrRateNotFound
		},
	}
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), source)

	body := `{"currency_code":"USD","rate_type":"spot","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Resolve_MapsInvalidRateType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	source := fxRateSourceStub{
		rateFunc: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.Amount{}, amount.ErrInvalidRate
		},
	}
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), source)

	body := `{"currency_code":"USD","rate_type":"bogus","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Resolve_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	source := fxRateSourceStub{
		rateFunc: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.Amount{}, errors.New("rate source down")
		},
	}
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), source)

	body := `{"currency_code":"USD","rate_type":"spot","date":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/fx-rates/resolve", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_UpdatesRate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "USD"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fx_rates" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"rate_type":"spot","valid_from":"2026-01-02"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"valid_from":"2026-01-02"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":0,"valid_from":""}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_RejectsInvalidDate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"valid_from":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"valid_from":"2026-01-02"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(99), 15000.0, "spot", timeNow()))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"valid_from":"2026-01-02"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"valid_from":"2026-01-02"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_MapsUnknownCurrency(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"valid_from":"2026-01-02"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("USD", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(1, "USD"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "fx_rates" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	body := `{"rate":15200,"valid_from":"2026-01-02"}`
	resp, err := doRequest(app, http.MethodPut, "/fx-rates/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Delete_DeletesRate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "fx_rates" WHERE "fx_rates"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodDelete, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodDelete, "/fx-rates/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodDelete, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodDelete, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "fx_rates" WHERE id = $1 ORDER BY "fx_rates"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(fxRateColumns).
			AddRow(1, timeNow(), timeNow(), nil, "USD", uint64(10), 15000.0, "spot", timeNow()))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "fx_rates" WHERE "fx_rates"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := fxRateHandlerTest(t, dao.NewBase[reference.FxRate](db), fxRateTestSvc(dao.NewBase[reference.FxRate](db), dao.NewBase[reference.Currency](db)), fxRateSourceStub{})

	resp, err := doRequest(app, http.MethodDelete, "/fx-rates/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestFxRateHandler_WriteFxRateError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "unknown currency", err: reference.ErrCurrencyNotFound, want: http.StatusUnprocessableEntity},
		{name: "missing valid from", err: reference.ErrRateValidFromMissing, want: http.StatusUnprocessableEntity},
		{name: "invalid rate", err: amount.ErrInvalidRate, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeFxRateError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
