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
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func bankStatementTestApp(t *testing.T, withTenant bool, svc accounting.BankStatementService) *fiber.App {
	t.Helper()
	h := NewBankStatementHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func newBankStatementService(statements accounting.BankStatementDAOMock, lines accounting.BankStatementLineDAOMock, payments accounting.PaymentDAOMock, journals dao.CRUDMock[reference.Journal]) accounting.BankStatementService {
	return accounting.NewBankStatementService(statements, lines, payments, journals, accounting.TransactionerMock{})
}

func TestBankStatementHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.BankStatement], error) {
			return &query.Page[accounting.BankStatement]{Items: []*accounting.BankStatement{sampleBankStatement()}, Count: 1}, nil
		}
		svc := newBankStatementService(statements, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.BankStatement], error) {
			return &query.Page[accounting.BankStatement]{Items: []*accounting.BankStatement{sampleBankStatement()}, Count: 1}, nil
		}
		svc := newBankStatementService(statements, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.BankStatement], error) {
			return nil, errors.New("boom")
		}
		svc := newBankStatementService(statements, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func sampleTenantBankStatement() *accounting.BankStatement {
	return &accounting.BankStatement{Base: model.Base{ID: 1}, JournalID: ptrUint64(10), State: accounting.BankStatementStateOpen}
}

func TestBankStatementHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.FindFunc = func(_ context.Context, _ uint64) (*accounting.BankStatement, error) {
			return sampleTenantBankStatement(), nil
		}
		lines := accounting.BankStatementLineDAOMock{}
		lines.ListByStatementFunc = func(_ context.Context, _ uint64) ([]*accounting.BankStatementLine, error) {
			return []*accounting.BankStatementLine{{Base: model.Base{ID: 1}, StatementID: 1, Amount: amount.FromFloat64(100)}}, nil
		}
		svc := newBankStatementService(statements, lines, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.FindFunc = func(_ context.Context, _ uint64) (*accounting.BankStatement, error) {
			return nil, nil
		}
		svc := newBankStatementService(statements, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.FindFunc = func(_ context.Context, _ uint64) (*accounting.BankStatement, error) {
			return nil, errors.New("boom")
		}
		svc := newBankStatementService(statements, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("lines error", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.FindFunc = func(_ context.Context, _ uint64) (*accounting.BankStatement, error) {
			return sampleTenantBankStatement(), nil
		}
		lines := accounting.BankStatementLineDAOMock{}
		lines.ListByStatementFunc = func(_ context.Context, _ uint64) ([]*accounting.BankStatementLine, error) {
			return nil, errors.New("boom")
		}
		svc := newBankStatementService(statements, lines, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/bank-statements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestBankStatementHandlerCreate(t *testing.T) {
	body := `{"journal_id":1,"lines":[{"amount":100,"contact_id":5,"ref":"INV/001"}]}`

	t.Run("success", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{}
		journals.FindFunc = func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
		}
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, journals)
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("success with date", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{}
		journals.FindFunc = func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
		}
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, journals)
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements", `{"journal_id":1,"date":"2026-01-15","lines":[{"date":"2026-01-15","amount":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements", `{"journal_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements", `{"journal_id":1,"date":"bogus","lines":[{"amount":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid line date", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements", `{"journal_id":1,"lines":[{"date":"bogus","amount":100}]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("journal error", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{}
		journals.FindFunc = func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return nil, errors.New("boom")
		}
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, journals)
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestBankStatementHandlerMatch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.FindFunc = func(_ context.Context, _ uint64) (*accounting.BankStatement, error) {
			return sampleBankStatement(), nil
		}
		lines := accounting.BankStatementLineDAOMock{}
		lines.ListUnreconciledByStatementFunc = func(_ context.Context, _ uint64) ([]*accounting.BankStatementLine, error) {
			return []*accounting.BankStatementLine{{Base: model.Base{ID: 1}, StatementID: 1, Amount: amount.FromFloat64(100), ContactID: ptrUint64(5)}}, nil
		}
		payments := accounting.PaymentDAOMock{}
		payments.ListPostedByContactFunc = func(_ context.Context, _ uint64) ([]*accounting.Payment, error) {
			return []*accounting.Payment{{Base: model.Base{ID: 1}, Amount: 100, EntryID: ptrUint64(9)}}, nil
		}
		svc := newBankStatementService(statements, lines, payments, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements/1/match", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newBankStatementService(accounting.BankStatementDAOMock{}, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements/abc/match", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.FindFunc = func(_ context.Context, _ uint64) (*accounting.BankStatement, error) {
			return nil, nil
		}
		svc := newBankStatementService(statements, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements/1/match", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("cancelled", func(t *testing.T) {
		statements := accounting.BankStatementDAOMock{}
		statements.FindFunc = func(_ context.Context, _ uint64) (*accounting.BankStatement, error) {
			return &accounting.BankStatement{Base: model.Base{ID: 1}, JournalID: ptrUint64(1), State: accounting.BankStatementStateCancelled}, nil
		}
		svc := newBankStatementService(statements, accounting.BankStatementLineDAOMock{}, accounting.PaymentDAOMock{}, dao.CRUDMock[reference.Journal]{})
		app := bankStatementTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/bank-statements/1/match", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestWriteBankStatementError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrStatementNotFound, status: http.StatusNotFound},
		{name: "no lines", err: accounting.ErrStatementNoLines, status: http.StatusUnprocessableEntity},
		{name: "cancelled", err: accounting.ErrStatementCancelled, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeBankStatementError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
