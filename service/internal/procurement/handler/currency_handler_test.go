package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

func currencyTestApp(t *testing.T, rates procurement.CurrencyRateDAOMock) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		NewCurrencyRateHandler(procurement.NewCurrencyRateService(rates)).Register(api, guards)
	})
}

func sampleCurrencyRate(orgID uint64) *procurement.CurrencyRate {
	return &procurement.CurrencyRate{
		Base:           model.Base{ID: 1},
		OrganizationID: &orgID,
		FromCurrency:   "USD",
		ToCurrency:     "IDR",
		Rate:           16000,
	}
}

func TestCurrencyRateHandler_CRUD(t *testing.T) {
	t.Run("lists rates", func(t *testing.T) {
		mock := procurement.CurrencyRateDAOMock{}
		mock.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.CurrencyRate], error) {
			return &query.Page[procurement.CurrencyRate]{Items: []*procurement.CurrencyRate{sampleCurrencyRate(10)}, Count: 1}, nil
		}
		resp, err := doRequest(currencyTestApp(t, mock), http.MethodGet, "/currency-rates/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid query and missing tenant and reports failure", func(t *testing.T) {
		resp, err := doRequest(currencyTestApp(t, procurement.CurrencyRateDAOMock{}), http.MethodGet, "/currency-rates/?size=abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
			NewCurrencyRateHandler(procurement.NewCurrencyRateService(procurement.CurrencyRateDAOMock{})).Register(api, guards)
		})
		resp, err = doRequest(app, http.MethodGet, "/currency-rates/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

		failing := procurement.CurrencyRateDAOMock{}
		failing.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.CurrencyRate], error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(currencyTestApp(t, failing), http.MethodGet, "/currency-rates/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
			resp, err := doRequest(currencyTestApp(t, procurement.CurrencyRateDAOMock{}), method, "/currency-rates/abc", "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("returns rate and handles missing foreign and failure", func(t *testing.T) {
		found := procurement.CurrencyRateDAOMock{}
		found.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return sampleCurrencyRate(10), nil
		}
		resp, err := doRequest(currencyTestApp(t, found), http.MethodGet, "/currency-rates/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)

		for _, orgID := range []uint64{0, 11} {
			var rate *procurement.CurrencyRate
			if orgID != 0 {
				rate = sampleCurrencyRate(orgID)
			}
			missing := procurement.CurrencyRateDAOMock{}
			missing.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
				return rate, nil
			}
			resp, err = doRequest(currencyTestApp(t, missing), http.MethodGet, "/currency-rates/1", "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
		}

		failing := procurement.CurrencyRateDAOMock{}
		failing.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(currencyTestApp(t, failing), http.MethodGet, "/currency-rates/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("creates rate", func(t *testing.T) {
		mock := procurement.CurrencyRateDAOMock{}
		mock.CreateFunc = func(_ context.Context, rate *procurement.CurrencyRate) (*procurement.CurrencyRate, error) {
			rate.ID = 5
			return rate, nil
		}
		resp, err := doRequest(currencyTestApp(t, mock), http.MethodPost, "/currency-rates/", `{"from_currency":"USD","to_currency":"IDR","rate":16000,"rate_date":"2026-01-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)
	})

	t.Run("rejects invalid create", func(t *testing.T) {
		resp, err := doRequest(currencyTestApp(t, procurement.CurrencyRateDAOMock{}), http.MethodPost, "/currency-rates/", `{"from_currency":"US"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(currencyTestApp(t, procurement.CurrencyRateDAOMock{}), http.MethodPost, "/currency-rates/", `{"from_currency":"USD","to_currency":"IDR","rate":16000,"rate_date":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("maps create errors", func(t *testing.T) {
		conflict := procurement.CurrencyRateDAOMock{}
		conflict.CreateFunc = func(_ context.Context, _ *procurement.CurrencyRate) (*procurement.CurrencyRate, error) {
			return nil, procurement.ErrCurrencyRateExists
		}
		resp, err := doRequest(currencyTestApp(t, conflict), http.MethodPost, "/currency-rates/", `{"from_currency":"USD","to_currency":"IDR","rate":16000}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusConflict)

		failing := procurement.CurrencyRateDAOMock{}
		failing.CreateFunc = func(_ context.Context, _ *procurement.CurrencyRate) (*procurement.CurrencyRate, error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(currencyTestApp(t, failing), http.MethodPost, "/currency-rates/", `{"from_currency":"USD","to_currency":"IDR","rate":16000}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("updates rate and maps errors", func(t *testing.T) {
		mock := procurement.CurrencyRateDAOMock{}
		mock.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return sampleCurrencyRate(10), nil
		}
		mock.UpdateFunc = func(_ context.Context, rate *procurement.CurrencyRate) (*procurement.CurrencyRate, error) {
			return rate, nil
		}
		resp, err := doRequest(currencyTestApp(t, mock), http.MethodPut, "/currency-rates/1", `{"rate":16500}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)

		missing := procurement.CurrencyRateDAOMock{}
		missing.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return nil, nil
		}
		resp, err = doRequest(currencyTestApp(t, missing), http.MethodPut, "/currency-rates/1", `{"rate":16500}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		conflict := procurement.CurrencyRateDAOMock{}
		conflict.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return sampleCurrencyRate(10), nil
		}
		conflict.UpdateFunc = func(_ context.Context, _ *procurement.CurrencyRate) (*procurement.CurrencyRate, error) {
			return nil, procurement.ErrCurrencyRateExists
		}
		resp, err = doRequest(currencyTestApp(t, conflict), http.MethodPut, "/currency-rates/1", `{"rate":16500}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusConflict)

		failing := procurement.CurrencyRateDAOMock{}
		failing.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return sampleCurrencyRate(10), nil
		}
		failing.UpdateFunc = func(_ context.Context, _ *procurement.CurrencyRate) (*procurement.CurrencyRate, error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(currencyTestApp(t, failing), http.MethodPut, "/currency-rates/1", `{"rate":16500}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("deletes rate", func(t *testing.T) {
		mock := procurement.CurrencyRateDAOMock{}
		mock.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return sampleCurrencyRate(10), nil
		}
		mock.DeleteFunc = func(_ context.Context, _ uint64) error { return nil }
		resp, err := doRequest(currencyTestApp(t, mock), http.MethodDelete, "/currency-rates/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects delete for foreign and reports failure", func(t *testing.T) {
		foreign := procurement.CurrencyRateDAOMock{}
		foreign.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return sampleCurrencyRate(11), nil
		}
		resp, err := doRequest(currencyTestApp(t, foreign), http.MethodDelete, "/currency-rates/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		failing := procurement.CurrencyRateDAOMock{}
		failing.FindFunc = func(_ context.Context, _ uint64) (*procurement.CurrencyRate, error) {
			return sampleCurrencyRate(10), nil
		}
		failing.DeleteFunc = func(_ context.Context, _ uint64) error { return errors.New("db down") }
		resp, err = doRequest(currencyTestApp(t, failing), http.MethodDelete, "/currency-rates/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
