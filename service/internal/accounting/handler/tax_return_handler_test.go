package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func taxReturnTestApp(t *testing.T, withTenant bool, svc accounting.TaxReturnService) *fiber.App {
	t.Helper()
	h := NewTaxReturnHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newTaxReturnService(returns accounting.TaxReturnDAOMock, periods accounting.TaxPeriodDAOMock, invoiceTaxes accounting.InvoiceTaxDAOMock) accounting.TaxReturnService {
	return accounting.NewTaxReturnService(returns, periods, invoiceTaxes, accounting.TransactionerMock{})
}

func taxReturnServiceMocks() (accounting.TaxReturnDAOMock, accounting.TaxPeriodDAOMock, accounting.InvoiceTaxDAOMock) {
	returns := accounting.TaxReturnDAOMock{}
	periods := accounting.TaxPeriodDAOMock{}
	periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
		start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		return &accounting.TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, DateStart: &start, DateEnd: &end, State: accounting.TaxPeriodStateClosed}, nil
	}
	invoiceTaxes := accounting.InvoiceTaxDAOMock{}
	invoiceTaxes.SumTaxByPeriodFunc = func(_ context.Context, _ uint64, _ string, _, _ time.Time) (float64, error) {
		return 100, nil
	}
	return returns, periods, invoiceTaxes
}

func TestTaxReturnHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxReturn], error) {
			return &query.Page[accounting.TaxReturn]{Items: []*accounting.TaxReturn{sampleTaxReturn()}, Count: 1}, nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxReturn], error) {
			return &query.Page[accounting.TaxReturn]{Items: []*accounting.TaxReturn{sampleTaxReturn()}, Count: 1}, nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxReturn], error) {
			return nil, errors.New("boom")
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTaxReturnHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return nil, nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return &accounting.TaxReturn{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(99)}, nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return nil, errors.New("boom")
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/tax-returns/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTaxReturnHandlerCreate(t *testing.T) {
	body := `{"period_id":1,"type":"vat"}`

	t.Run("success", func(t *testing.T) {
		returns, periods, invoiceTaxes := taxReturnServiceMocks()
		returns.CreateFunc = func(_ context.Context, _ *accounting.TaxReturn) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		svc := newTaxReturnService(returns, periods, invoiceTaxes)
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		returns, periods, invoiceTaxes := taxReturnServiceMocks()
		svc := newTaxReturnService(returns, periods, invoiceTaxes)
		app := taxReturnTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("period not found", func(t *testing.T) {
		periods := accounting.TaxPeriodDAOMock{}
		periods.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		}
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, periods, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		returns, periods, invoiceTaxes := taxReturnServiceMocks()
		returns.FindByPeriodFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		svc := newTaxReturnService(returns, periods, invoiceTaxes)
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestTaxReturnHandlerFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		returns.UpdateFunc = func(_ context.Context, _ *accounting.TaxReturn) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/1/file", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/abc/file", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return nil, nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/1/file", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("not draft", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return &accounting.TaxReturn{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), State: accounting.TaxReturnStateFiled}, nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/1/file", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestTaxReturnHandlerPay(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return &accounting.TaxReturn{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), State: accounting.TaxReturnStateFiled}, nil
		}
		returns.UpdateFunc = func(_ context.Context, _ *accounting.TaxReturn) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/1/pay", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/abc/pay", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not filed", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/1/pay", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestTaxReturnHandlerOpen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return &accounting.TaxReturn{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), State: accounting.TaxReturnStateFiled}, nil
		}
		returns.UpdateFunc = func(_ context.Context, _ *accounting.TaxReturn) (*accounting.TaxReturn, error) {
			return sampleTaxReturn(), nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/1/open", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newTaxReturnService(accounting.TaxReturnDAOMock{}, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/abc/open", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("already paid", func(t *testing.T) {
		returns := accounting.TaxReturnDAOMock{}
		returns.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxReturn, error) {
			return &accounting.TaxReturn{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), State: accounting.TaxReturnStatePaid}, nil
		}
		svc := newTaxReturnService(returns, accounting.TaxPeriodDAOMock{}, accounting.InvoiceTaxDAOMock{})
		app := taxReturnTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/tax-returns/1/open", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestWriteTaxReturnError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrTaxReturnNotFound, status: http.StatusNotFound},
		{name: "period not found", err: accounting.ErrPeriodNotFound, status: http.StatusNotFound},
		{name: "exists", err: accounting.ErrTaxReturnExists, status: http.StatusConflict},
		{name: "not draft", err: accounting.ErrTaxReturnNotDraft, status: http.StatusConflict},
		{name: "not filed", err: accounting.ErrTaxReturnNotFiled, status: http.StatusConflict},
		{name: "paid", err: accounting.ErrTaxReturnPaid, status: http.StatusConflict},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeTaxReturnError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
