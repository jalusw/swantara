package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func taxPeriodTestApp(t *testing.T, withTenant bool, svc accounting.TaxPeriodService) *fiber.App {
	t.Helper()
	h := NewTaxPeriodHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newTaxPeriodService(periods accounting.TaxPeriodDAOMock, years dao.CRUDMock[reference.TaxYear]) accounting.TaxPeriodService {
	return accounting.NewTaxPeriodService(periods, years)
}

func taxYearMock() dao.CRUDMock[reference.TaxYear] {
	years := dao.CRUDMock[reference.TaxYear]{}
	years.FindFunc = func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
		return &reference.TaxYear{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10)}, nil
	}
	return years
}

func TestTaxPeriodHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxPeriod], error) {
			return &query.Page[accounting.TaxPeriod]{Items: []*accounting.TaxPeriod{sampleTaxPeriod()}, Count: 1}, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxPeriod], error) {
			return &query.Page[accounting.TaxPeriod]{Items: []*accounting.TaxPeriod{sampleTaxPeriod()}, Count: 1}, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := taxPeriodTestApp(t, false, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxPeriod], error) {
			return nil, errors.New("boom")
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTaxPeriodHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 99}, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, errors.New("boom")
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodGet, "/tax-periods/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTaxPeriodHandlerCreate(t *testing.T) {
	body := `{"tax_year_id":1,"name":"January","date_start":"2026-01-01","date_end":"2026-01-31"}`

	t.Run("success", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.CreateFunc = func(_ context.Context, _ *accounting.TaxPeriod) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("success with state", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.CreateFunc = func(_ context.Context, _ *accounting.TaxPeriod) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", `{"tax_year_id":1,"name":"January","date_start":"2026-01-01","date_end":"2026-01-31","state":"closed"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := taxPeriodTestApp(t, false, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid start date", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", `{"tax_year_id":1,"name":"January","date_start":"bogus","date_end":"2026-01-31"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid end date", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", `{"tax_year_id":1,"name":"January","date_start":"2026-01-01","date_end":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid period state", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", `{"tax_year_id":1,"name":"January","date_start":"2026-01-01","date_end":"2026-01-31","state":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date range", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", `{"tax_year_id":1,"name":"January","date_start":"2026-02-01","date_end":"2026-01-31"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("tax year not found", func(t *testing.T) {
		years := dao.CRUDMock[reference.TaxYear]{}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, years))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestTaxPeriodHandlerClose(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		periods.UpdateFunc = func(_ context.Context, _ *accounting.TaxPeriod) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/close", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/abc/close", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/close", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("locked", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, State: accounting.TaxPeriodStateLocked}, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/close", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestTaxPeriodHandlerLock(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, State: accounting.TaxPeriodStateClosed}, nil
		}
		periods.UpdateFunc = func(_ context.Context, _ *accounting.TaxPeriod) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/lock", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/abc/lock", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not closed", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/lock", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestTaxPeriodHandlerOpen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, State: accounting.TaxPeriodStateClosed}, nil
		}
		periods.UpdateFunc = func(_ context.Context, _ *accounting.TaxPeriod) (*accounting.TaxPeriod, error) {
			return sampleTaxPeriod(), nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/open", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := taxPeriodTestApp(t, true, newTaxPeriodService(accounting.TaxPeriodDAOMock{}, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/abc/open", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/open", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("locked", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, State: accounting.TaxPeriodStateLocked}, nil
		}
		app := taxPeriodTestApp(t, true, newTaxPeriodService(periods, taxYearMock()))

		resp, err := doRequest(app, http.MethodPost, "/tax-periods/1/open", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestWriteTaxPeriodError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrPeriodNotFound, status: http.StatusNotFound},
		{name: "invalid period", err: accounting.ErrInvalidPeriod, status: http.StatusUnprocessableEntity},
		{name: "invalid state", err: accounting.ErrInvalidPeriodState, status: http.StatusUnprocessableEntity},
		{name: "not closed", err: accounting.ErrPeriodNotClosed, status: http.StatusConflict},
		{name: "locked", err: accounting.ErrPeriodLocked, status: http.StatusConflict},
		{name: "year not found", err: reference.ErrTaxYearNotFound, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeTaxPeriodError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
