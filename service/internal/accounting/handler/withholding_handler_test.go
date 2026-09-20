package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func withholdingTestApp(t *testing.T, withTenant bool, svc accounting.WithholdingService) *fiber.App {
	t.Helper()
	h := NewWithholdingTaxHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newWithholdingService(withholdings accounting.WithholdingTaxDAOMock) accounting.WithholdingService {
	return accounting.NewWithholdingService(withholdings, accounting.PosterMock{}, accounting.TransactionerMock{})
}

func TestWithholdingTaxHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.WithholdingTax], error) {
			return &query.Page[accounting.WithholdingTax]{Items: []*accounting.WithholdingTax{sampleWithholdingTax()}, Count: 1}, nil
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodGet, "/withholding-taxes/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.WithholdingTax], error) {
			return &query.Page[accounting.WithholdingTax]{Items: []*accounting.WithholdingTax{sampleWithholdingTax()}, Count: 1}, nil
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodGet, "/withholding-taxes/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := withholdingTestApp(t, true, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/withholding-taxes/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := withholdingTestApp(t, false, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/withholding-taxes/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.WithholdingTax], error) {
			return nil, errors.New("boom")
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodGet, "/withholding-taxes/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWithholdingTaxHandlerCreate(t *testing.T) {
	body := `{"name":"WHT","rate_pct":2,"account_id":500,"scope":"purchase"}`

	t.Run("success", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.CreateFunc = func(_ context.Context, _ *accounting.WithholdingTax) (*accounting.WithholdingTax, error) {
			return sampleWithholdingTax(), nil
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := withholdingTestApp(t, true, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := withholdingTestApp(t, false, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.CreateFunc = func(_ context.Context, _ *accounting.WithholdingTax) (*accounting.WithholdingTax, error) {
			return nil, errors.New("boom")
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWithholdingTaxHandlerApply(t *testing.T) {
	body := `{"journal_id":1,"contact_id":5,"amount":1000,"date":"2026-01-15","scope":"purchase","withholding_tax_id":1,"bank_account_id":300,"payable_account_id":400}`

	t.Run("success", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.FindFunc = func(_ context.Context, _ uint64) (*accounting.WithholdingTax, error) {
			return &accounting.WithholdingTax{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Name: ptrString("WHT"), RatePct: 2, AccountID: ptrUint64(500), Scope: "purchase", Active: true}, nil
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := withholdingTestApp(t, true, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := withholdingTestApp(t, false, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		app := withholdingTestApp(t, true, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", `{"journal_id":1,"contact_id":5,"amount":1000,"date":"bogus","scope":"purchase","withholding_tax_id":1,"bank_account_id":300,"payable_account_id":400}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("missing date", func(t *testing.T) {
		app := withholdingTestApp(t, true, newWithholdingService(accounting.WithholdingTaxDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", `{"journal_id":1,"contact_id":5,"amount":1000,"scope":"purchase","withholding_tax_id":1,"bank_account_id":300,"payable_account_id":400}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.FindFunc = func(_ context.Context, _ uint64) (*accounting.WithholdingTax, error) {
			return nil, nil
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("no account", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.FindFunc = func(_ context.Context, _ uint64) (*accounting.WithholdingTax, error) {
			return &accounting.WithholdingTax{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Name: ptrString("WHT"), RatePct: 2, Scope: "purchase", Active: true}, nil
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("scope mismatch", func(t *testing.T) {
		withholdings := accounting.WithholdingTaxDAOMock{}
		withholdings.FindFunc = func(_ context.Context, _ uint64) (*accounting.WithholdingTax, error) {
			return &accounting.WithholdingTax{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Name: ptrString("WHT"), RatePct: 2, AccountID: ptrUint64(500), Scope: "sale", Active: true}, nil
		}
		app := withholdingTestApp(t, true, newWithholdingService(withholdings))

		resp, err := doRequest(app, http.MethodPost, "/withholding-taxes/apply", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestWriteWithholdingTaxError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrWithholdingNotFound, status: http.StatusNotFound},
		{name: "no account", err: accounting.ErrWithholdingNoAccount, status: http.StatusUnprocessableEntity},
		{name: "scope", err: accounting.ErrWithholdingScope, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeWithholdingTaxError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
