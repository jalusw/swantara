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

func scorecardTestApp(t *testing.T, scorecards procurement.SupplierScorecardDAOMock) *fiber.App {
	t.Helper()
	return procurementTestApp(t, func(api fiber.Router, guards httpx.RouteGuards) {
		NewSupplierScorecardHandler(procurement.NewSupplierScorecardService(scorecards)).Register(api, guards)
	})
}

func sampleScorecard() *procurement.SupplierScorecard {
	return &procurement.SupplierScorecard{
		Base:             model.Base{ID: 1},
		SupplierID:       20,
		OverallScore:     90,
		TotalOrders:      10,
		OnTimeDeliveries: 9,
	}
}

func TestSupplierScorecardHandler_Read(t *testing.T) {
	t.Run("lists scorecards", func(t *testing.T) {
		mock := procurement.SupplierScorecardDAOMock{}
		mock.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierScorecard], error) {
			return &query.Page[procurement.SupplierScorecard]{Items: []*procurement.SupplierScorecard{sampleScorecard()}, Count: 1}, nil
		}
		resp, err := doRequest(scorecardTestApp(t, mock), http.MethodGet, "/supplier-scorecards/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid query and missing tenant and reports failure", func(t *testing.T) {
		resp, err := doRequest(scorecardTestApp(t, procurement.SupplierScorecardDAOMock{}), http.MethodGet, "/supplier-scorecards/?size=abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		app := procurementTestAppNoTenant(t, func(api fiber.Router, guards httpx.RouteGuards) {
			NewSupplierScorecardHandler(procurement.NewSupplierScorecardService(procurement.SupplierScorecardDAOMock{})).Register(api, guards)
		})
		resp, err = doRequest(app, http.MethodGet, "/supplier-scorecards/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnauthorized)

		failing := procurement.SupplierScorecardDAOMock{}
		failing.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierScorecard], error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(scorecardTestApp(t, failing), http.MethodGet, "/supplier-scorecards/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		for _, path := range []string{"/supplier-scorecards/abc", "/supplier-scorecards/supplier/abc"} {
			resp, err := doRequest(scorecardTestApp(t, procurement.SupplierScorecardDAOMock{}), http.MethodGet, path, "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("returns scorecard and maps errors", func(t *testing.T) {
		found := procurement.SupplierScorecardDAOMock{}
		found.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplierScorecard, error) {
			return sampleScorecard(), nil
		}
		resp, err := doRequest(scorecardTestApp(t, found), http.MethodGet, "/supplier-scorecards/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)

		missing := procurement.SupplierScorecardDAOMock{}
		missing.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplierScorecard, error) {
			return nil, nil
		}
		resp, err = doRequest(scorecardTestApp(t, missing), http.MethodGet, "/supplier-scorecards/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		failing := procurement.SupplierScorecardDAOMock{}
		failing.FindFunc = func(_ context.Context, _ uint64) (*procurement.SupplierScorecard, error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(scorecardTestApp(t, failing), http.MethodGet, "/supplier-scorecards/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("lists by supplier and reports failure", func(t *testing.T) {
		mock := procurement.SupplierScorecardDAOMock{}
		mock.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierScorecard], error) {
			return &query.Page[procurement.SupplierScorecard]{Items: []*procurement.SupplierScorecard{sampleScorecard()}, Count: 1}, nil
		}
		resp, err := doRequest(scorecardTestApp(t, mock), http.MethodGet, "/supplier-scorecards/supplier/20", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)

		failing := procurement.SupplierScorecardDAOMock{}
		failing.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[procurement.SupplierScorecard], error) {
			return nil, errors.New("db down")
		}
		resp, err = doRequest(scorecardTestApp(t, failing), http.MethodGet, "/supplier-scorecards/supplier/20", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
