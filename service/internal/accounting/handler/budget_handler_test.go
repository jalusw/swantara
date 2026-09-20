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

func budgetTestApp(t *testing.T, withTenant bool, svc accounting.BudgetService) *fiber.App {
	t.Helper()
	h := NewBudgetHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newBudgetService(budgets accounting.BudgetDAOMock, lines accounting.BudgetLineDAOMock, queryMock accounting.BudgetQueryDAOMock) accounting.BudgetService {
	return accounting.NewBudgetService(budgets, lines, queryMock, accounting.TransactionerMock{})
}

func TestBudgetHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Budget], error) {
			return &query.Page[accounting.Budget]{Items: []*accounting.Budget{sampleBudget()}, Count: 1}, nil
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Budget], error) {
			return &query.Page[accounting.Budget]{Items: []*accounting.Budget{sampleBudget()}, Count: 1}, nil
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := budgetTestApp(t, true, newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := budgetTestApp(t, false, newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Budget], error) {
			return nil, errors.New("boom")
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestBudgetHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.FindFunc = func(_ context.Context, _ uint64) (*accounting.Budget, error) {
			return sampleBudget(), nil
		}
		lines := accounting.BudgetLineDAOMock{}
		lines.ListByBudgetFunc = func(_ context.Context, _ uint64) ([]*accounting.BudgetLine, error) {
			return []*accounting.BudgetLine{{Base: model.Base{ID: 1}, BudgetID: 1, AccountID: 100}}, nil
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, lines, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := budgetTestApp(t, true, newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.FindFunc = func(_ context.Context, _ uint64) (*accounting.Budget, error) {
			return nil, nil
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.FindFunc = func(_ context.Context, _ uint64) (*accounting.Budget, error) {
			return &accounting.Budget{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(99)}, nil
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.FindFunc = func(_ context.Context, _ uint64) (*accounting.Budget, error) {
			return nil, errors.New("boom")
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("lines error", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.FindFunc = func(_ context.Context, _ uint64) (*accounting.Budget, error) {
			return sampleBudget(), nil
		}
		lines := accounting.BudgetLineDAOMock{}
		lines.ListByBudgetFunc = func(_ context.Context, _ uint64) ([]*accounting.BudgetLine, error) {
			return nil, errors.New("boom")
		}
		app := budgetTestApp(t, true, newBudgetService(budgets, lines, accounting.BudgetQueryDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/budgets/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestBudgetHandlerCreate(t *testing.T) {
	body := `{"name":"FY2026","date_start":"2026-01-01","date_end":"2026-12-31","lines":[{"account_id":100,"planned_amount":1000}]}`

	t.Run("success", func(t *testing.T) {
		svc := newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/budgets", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/budgets", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/budgets", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid start date", func(t *testing.T) {
		svc := newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/budgets", `{"name":"FY2026","date_start":"bogus","date_end":"2026-12-31","lines":[{"account_id":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid end date", func(t *testing.T) {
		svc := newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/budgets", `{"name":"FY2026","date_start":"2026-01-01","date_end":"bogus","lines":[{"account_id":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date range", func(t *testing.T) {
		svc := newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/budgets", `{"name":"FY2026","date_start":"2026-12-01","date_end":"2026-01-01","lines":[{"account_id":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestBudgetHandlerVariance(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		budgets := accounting.BudgetDAOMock{}
		budgets.FindFunc = func(_ context.Context, _ uint64) (*accounting.Budget, error) {
			return &accounting.Budget{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Name: ptrString("FY2026"), DateStart: &start, DateEnd: &end, State: accounting.BudgetStateDraft}, nil
		}
		lines := accounting.BudgetLineDAOMock{}
		lines.ListByBudgetFunc = func(_ context.Context, _ uint64) ([]*accounting.BudgetLine, error) {
			return []*accounting.BudgetLine{{Base: model.Base{ID: 1}, BudgetID: 1, AccountID: 100, PlannedAmount: 1000}}, nil
		}
		svc := newBudgetService(budgets, lines, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/budgets/1/variance", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newBudgetService(accounting.BudgetDAOMock{}, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/budgets/abc/variance", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		budgets := accounting.BudgetDAOMock{}
		budgets.FindFunc = func(_ context.Context, _ uint64) (*accounting.Budget, error) {
			return nil, nil
		}
		svc := newBudgetService(budgets, accounting.BudgetLineDAOMock{}, accounting.BudgetQueryDAOMock{})
		app := budgetTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/budgets/1/variance", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestWriteBudgetError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrBudgetNotFound, status: http.StatusNotFound},
		{name: "no lines", err: accounting.ErrBudgetNoLines, status: http.StatusUnprocessableEntity},
		{name: "invalid dates", err: accounting.ErrBudgetInvalidDates, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeBudgetError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
