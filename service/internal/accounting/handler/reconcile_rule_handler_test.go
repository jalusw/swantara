package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type reconcileRuleDAOFake struct {
	dao.CRUDMock[accounting.ReconcileRule]
	listByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*accounting.ReconcileRule, error)
}

func (f reconcileRuleDAOFake) ListByOrganization(ctx context.Context, organizationID uint64) ([]*accounting.ReconcileRule, error) {
	if f.listByOrganizationFunc != nil {
		return f.listByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}

func reconcileRuleTestApp(t *testing.T, withTenant bool, engine accounting.ReconcileRuleEngine) *fiber.App {
	t.Helper()
	h := NewReconcileRuleHandler(engine)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newReconcileRuleEngine(rules accounting.ReconcileRuleDAO, lines accounting.JournalLineDAOMock) accounting.ReconcileRuleEngine {
	return accounting.NewReconcileRuleEngine(rules, dao.CRUDMock[accounting.ReconcileRuleMatch]{}, lines, accounting.AccountPartialReconcileDAOMock{}, nil)
}

func sampleReconcileRule() *accounting.ReconcileRule {
	return &accounting.ReconcileRule{
		Base:           model.Base{ID: 1},
		OrganizationID: 10,
		Name:           ptrString("Contact match"),
		AccountID:      ptrUint64(1100),
		MatchContact:   true,
		MatchAmount:    true,
		Active:         true,
	}
}

func matchingRuleLines() accounting.JournalLineDAOMock {
	debit := &accounting.JournalLine{Base: model.Base{ID: 11}, EntryID: 1, AccountID: 1100, ContactID: ptrUint64(5), Debit: amount.FromInt64(100)}
	credit := &accounting.JournalLine{Base: model.Base{ID: 22}, EntryID: 2, AccountID: 1100, ContactID: ptrUint64(5), Credit: amount.FromInt64(100)}
	return accounting.JournalLineDAOMock{
		ListUnreconciledByAccountFunc: func(_ context.Context, _ uint64) ([]*accounting.JournalLine, error) {
			return []*accounting.JournalLine{debit, credit}, nil
		},
		CRUDMock: dao.CRUDMock[accounting.JournalLine]{
			FindFunc: func(_ context.Context, id uint64) (*accounting.JournalLine, error) {
				if id == 11 {
					return debit, nil
				}
				return credit, nil
			},
		},
	}
}

func TestReconcileRuleHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := reconcileRuleDAOFake{}
		rules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.ReconcileRule], error) {
			return &query.Page[accounting.ReconcileRule]{Items: []*accounting.ReconcileRule{sampleReconcileRule()}, Count: 1}, nil
		}
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(rules, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reconcile-rules/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reconcile-rules/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := reconcileRuleTestApp(t, false, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reconcile-rules/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		rules := reconcileRuleDAOFake{}
		rules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.ReconcileRule], error) {
			return nil, errors.New("boom")
		}
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(rules, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reconcile-rules/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestReconcileRuleHandlerCreate(t *testing.T) {
	body := `{"organization_id":10,"name":"Contact match","account_id":1100,"match_contact":true,"match_amount":true,"active":true}`

	t.Run("success", func(t *testing.T) {
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reconcileRuleTestApp(t, false, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules", `{"name":"Contact match","account_id":1100,"active":true}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		rules := reconcileRuleDAOFake{}
		rules.CreateFunc = func(_ context.Context, _ *accounting.ReconcileRule) (*accounting.ReconcileRule, error) {
			return nil, errors.New("boom")
		}
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(rules, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestReconcileRuleHandlerApply(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := reconcileRuleDAOFake{
			listByOrganizationFunc: func(_ context.Context, _ uint64) ([]*accounting.ReconcileRule, error) {
				return []*accounting.ReconcileRule{sampleReconcileRule()}, nil
			},
		}
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(rules, matchingRuleLines()))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules/apply", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reconcileRuleTestApp(t, false, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules/apply", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid min score", func(t *testing.T) {
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules/apply?min_score=bogus", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestReconcileRuleHandlerSuggest(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := reconcileRuleDAOFake{
			listByOrganizationFunc: func(_ context.Context, _ uint64) ([]*accounting.ReconcileRule, error) {
				return []*accounting.ReconcileRule{sampleReconcileRule()}, nil
			},
		}
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(rules, matchingRuleLines()))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules/suggest", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid min score", func(t *testing.T) {
		app := reconcileRuleTestApp(t, true, newReconcileRuleEngine(reconcileRuleDAOFake{}, accounting.JournalLineDAOMock{}))

		resp, err := doRequest(app, http.MethodPost, "/reconcile-rules/suggest?min_score=bogus", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}
