package handler

import (
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

var currencyColumns = []string{"id", "created_at", "updated_at", "deleted_at", "code", "name", "symbol", "decimal_places", "rounding"}

func currencyHandlerTest(t *testing.T, currencies dao.Base[reference.Currency]) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewCurrencyHandler(reference.NewFxRateService(dao.NewBase[reference.FxRate](nil), currencies))
		h.Register(api, guards)
	})
}

func TestCurrencyHandler_List_ReturnsCurrencies(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "currencies"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(currencyColumns).
			AddRow(1, timeNow(), timeNow(), nil, "IDR", "Indonesian Rupiah", "Rp", 2, 0.01))
	app := currencyHandlerTest(t, dao.NewBase[reference.Currency](db))

	resp, err := doRequest(app, http.MethodGet, "/currencies/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCurrencyHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := currencyHandlerTest(t, dao.NewBase[reference.Currency](db))

	resp, err := doRequest(app, http.MethodGet, "/currencies/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCurrencyHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "currencies"`)).
		WillReturnError(errors.New("db down"))
	app := currencyHandlerTest(t, dao.NewBase[reference.Currency](db))

	resp, err := doRequest(app, http.MethodGet, "/currencies/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCurrencyHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "currencies"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(currencyColumns).
			AddRow(1, timeNow(), timeNow(), nil, "IDR", "Indonesian Rupiah", "Rp", 2, 0.01))
	app := currencyHandlerTest(t, dao.NewBase[reference.Currency](db))

	resp, err := doRequest(app, http.MethodGet, "/currencies/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCurrencyHandler_Get_ReturnsCurrency(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("IDR", 1).
		WillReturnRows(sqlmock.NewRows(currencyColumns).
			AddRow(1, timeNow(), timeNow(), nil, "IDR", "Indonesian Rupiah", "Rp", 2, 0.01))
	app := currencyHandlerTest(t, dao.NewBase[reference.Currency](db))

	resp, err := doRequest(app, http.MethodGet, "/currencies/idr", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCurrencyHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("XXX", 1).
		WillReturnRows(sqlmock.NewRows(currencyColumns))
	app := currencyHandlerTest(t, dao.NewBase[reference.Currency](db))

	resp, err := doRequest(app, http.MethodGet, "/currencies/xxx", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCurrencyHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "currencies" WHERE code = $1 LIMIT $2`)).
		WithArgs("IDR", 1).
		WillReturnError(errors.New("db down"))
	app := currencyHandlerTest(t, dao.NewBase[reference.Currency](db))

	resp, err := doRequest(app, http.MethodGet, "/currencies/IDR", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}
