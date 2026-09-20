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

func taxRuleTestApp(t *testing.T, withTenant bool, resolver accounting.TaxRuleResolver) *fiber.App {
	t.Helper()
	h := NewTaxRuleHandler(resolver)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newTaxRuleResolver(positions accounting.TaxRuleDAOMock, taxMaps accounting.TaxRuleTaxMapDAOMock, accountMaps accounting.TaxRuleAccountMapDAOMock) accounting.TaxRuleResolver {
	return accounting.NewTaxRuleResolver(positions, taxMaps, accountMaps, accounting.TransactionerMock{})
}

func TestTaxRuleHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxRule], error) {
			return &query.Page[accounting.TaxRule]{Items: []*accounting.TaxRule{sampleTaxRule()}, Count: 1}, nil
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxRule], error) {
			return &query.Page[accounting.TaxRule]{Items: []*accounting.TaxRule{sampleTaxRule()}, Count: 1}, nil
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, false, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.TaxRule], error) {
			return nil, errors.New("boom")
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTaxRuleHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return sampleTaxRule(), nil
		}
		taxMaps := accounting.TaxRuleTaxMapDAOMock{}
		taxMaps.ListByPositionFunc = func(_ context.Context, _ uint64) ([]*accounting.TaxRuleTaxMap, error) {
			return []*accounting.TaxRuleTaxMap{{Base: model.Base{ID: 1}, TaxRuleID: 1, SrcTaxID: 1}}, nil
		}
		accountMaps := accounting.TaxRuleAccountMapDAOMock{}
		accountMaps.ListByPositionFunc = func(_ context.Context, _ uint64) ([]*accounting.TaxRuleAccountMap, error) {
			return []*accounting.TaxRuleAccountMap{{Base: model.Base{ID: 1}, TaxRuleID: 1, SrcAccountID: 100, DestAccountID: 200}}, nil
		}
		resolver := newTaxRuleResolver(positions, taxMaps, accountMaps)
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return nil, nil
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return &accounting.TaxRule{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(99)}, nil
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return nil, errors.New("boom")
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("tax maps error", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return sampleTaxRule(), nil
		}
		taxMaps := accounting.TaxRuleTaxMapDAOMock{}
		taxMaps.ListByPositionFunc = func(_ context.Context, _ uint64) ([]*accounting.TaxRuleTaxMap, error) {
			return nil, errors.New("boom")
		}
		resolver := newTaxRuleResolver(positions, taxMaps, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("account maps error", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return sampleTaxRule(), nil
		}
		accountMaps := accounting.TaxRuleAccountMapDAOMock{}
		accountMaps.ListByPositionFunc = func(_ context.Context, _ uint64) ([]*accounting.TaxRuleAccountMap, error) {
			return nil, errors.New("boom")
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accountMaps)
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodGet, "/tax-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTaxRuleHandlerCreate(t *testing.T) {
	body := `{"name":"Local","tax_maps":[{"src_tax_id":1,"dest_tax_id":2}],"account_maps":[{"src_account_id":100,"dest_account_id":200}]}`

	t.Run("success", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, false, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestTaxRuleHandlerResolve(t *testing.T) {
	body := `{"account_id":100}`

	t.Run("success", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return sampleTaxRule(), nil
		}
		accountMaps := accounting.TaxRuleAccountMapDAOMock{}
		accountMaps.ListByPositionFunc = func(_ context.Context, _ uint64) ([]*accounting.TaxRuleAccountMap, error) {
			return []*accounting.TaxRuleAccountMap{{Base: model.Base{ID: 1}, TaxRuleID: 1, SrcAccountID: 100, DestAccountID: 200}}, nil
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accountMaps)
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules/1/resolve", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("success with tax", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return sampleTaxRule(), nil
		}
		taxMaps := accounting.TaxRuleTaxMapDAOMock{}
		taxMaps.ListByPositionFunc = func(_ context.Context, _ uint64) ([]*accounting.TaxRuleTaxMap, error) {
			return []*accounting.TaxRuleTaxMap{{Base: model.Base{ID: 1}, TaxRuleID: 1, SrcTaxID: 1, DestTaxID: ptrUint64(2)}}, nil
		}
		resolver := newTaxRuleResolver(positions, taxMaps, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules/1/resolve", `{"tax_id":1,"account_id":100}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules/abc/resolve", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		resolver := newTaxRuleResolver(accounting.TaxRuleDAOMock{}, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules/1/resolve", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		positions := accounting.TaxRuleDAOMock{}
		positions.FindFunc = func(_ context.Context, _ uint64) (*accounting.TaxRule, error) {
			return nil, nil
		}
		resolver := newTaxRuleResolver(positions, accounting.TaxRuleTaxMapDAOMock{}, accounting.TaxRuleAccountMapDAOMock{})
		app := taxRuleTestApp(t, true, resolver)

		resp, err := doRequest(app, http.MethodPost, "/tax-rules/1/resolve", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestWriteTaxRuleError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrTaxRuleNotFound, status: http.StatusNotFound},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeTaxRuleError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
