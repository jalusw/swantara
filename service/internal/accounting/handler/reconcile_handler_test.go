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
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func reconcileTestApp(t *testing.T, withTenant bool, svc accounting.ReconcileService) *fiber.App {
	t.Helper()
	h := NewReconcileHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func TestReconcileHandlerReconcile(t *testing.T) {
	body := `{"debit_line_id":1,"credit_line_id":2,"amount":50}`

	t.Run("success", func(t *testing.T) {
		lines := accounting.JournalLineDAOMock{}
		lines.FindFunc = func(_ context.Context, id uint64) (*accounting.JournalLine, error) {
			line := &accounting.JournalLine{Base: model.Base{ID: id}, EntryID: 1, AccountID: 100}
			if id == 1 {
				line.Debit = amount.FromInt64(100)
			} else {
				line.Credit = amount.FromInt64(100)
			}
			return line, nil
		}
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
		}
		partials := accounting.AccountPartialReconcileDAOMock{}
		partials.CreateTxFunc = func(_ context.Context, _ *gorm.DB, _ *accounting.AccountPartialReconcile) (*accounting.AccountPartialReconcile, error) {
			return &accounting.AccountPartialReconcile{Base: model.Base{ID: 1}, DebitLineID: 1, CreditLineID: 2, Amount: 50}, nil
		}
		svc := accounting.NewReconcileService(lines, partials, accounting.AccountFullReconcileDAOMock{}, movements, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("full reconcile", func(t *testing.T) {
		lines := accounting.JournalLineDAOMock{}
		lines.FindFunc = func(_ context.Context, id uint64) (*accounting.JournalLine, error) {
			line := &accounting.JournalLine{Base: model.Base{ID: id}, EntryID: 1, AccountID: 100}
			if id == 1 {
				line.Debit = amount.FromInt64(100)
			} else {
				line.Credit = amount.FromInt64(100)
			}
			return line, nil
		}
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
		}
		partials := accounting.AccountPartialReconcileDAOMock{}
		partials.CreateTxFunc = func(_ context.Context, _ *gorm.DB, _ *accounting.AccountPartialReconcile) (*accounting.AccountPartialReconcile, error) {
			return &accounting.AccountPartialReconcile{Base: model.Base{ID: 1}, DebitLineID: 1, CreditLineID: 2, Amount: 100}, nil
		}
		fulls := accounting.AccountFullReconcileDAOMock{}
		fulls.CreateTxFunc = func(_ context.Context, _ *gorm.DB, _ *accounting.AccountFullReconcile) (*accounting.AccountFullReconcile, error) {
			return &accounting.AccountFullReconcile{Base: model.Base{ID: 1}}, nil
		}
		svc := accounting.NewReconcileService(lines, partials, fulls, movements, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", `{"debit_line_id":1,"credit_line_id":2,"amount":100}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := accounting.NewReconcileService(accounting.JournalLineDAOMock{}, accounting.AccountPartialReconcileDAOMock{}, accounting.AccountFullReconcileDAOMock{}, accounting.JournalEntryDAOMock{}, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := accounting.NewReconcileService(accounting.JournalLineDAOMock{}, accounting.AccountPartialReconcileDAOMock{}, accounting.AccountFullReconcileDAOMock{}, accounting.JournalEntryDAOMock{}, accounting.TransactionerMock{})
		app := reconcileTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("amount zero", func(t *testing.T) {
		svc := accounting.NewReconcileService(accounting.JournalLineDAOMock{}, accounting.AccountPartialReconcileDAOMock{}, accounting.AccountFullReconcileDAOMock{}, accounting.JournalEntryDAOMock{}, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", `{"debit_line_id":1,"credit_line_id":2,"amount":0}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("line not found", func(t *testing.T) {
		lines := accounting.JournalLineDAOMock{}
		lines.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalLine, error) {
			return nil, nil
		}
		svc := accounting.NewReconcileService(lines, accounting.AccountPartialReconcileDAOMock{}, accounting.AccountFullReconcileDAOMock{}, accounting.JournalEntryDAOMock{}, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("not in organization", func(t *testing.T) {
		lines := accounting.JournalLineDAOMock{}
		lines.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalLine, error) {
			return &accounting.JournalLine{Base: model.Base{ID: 1}, EntryID: 1, AccountID: 100, Debit: amount.FromInt64(100)}, nil
		}
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 99}, nil
		}
		svc := accounting.NewReconcileService(lines, accounting.AccountPartialReconcileDAOMock{}, accounting.AccountFullReconcileDAOMock{}, movements, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("different account", func(t *testing.T) {
		lines := accounting.JournalLineDAOMock{}
		lines.FindFunc = func(_ context.Context, id uint64) (*accounting.JournalLine, error) {
			line := &accounting.JournalLine{Base: model.Base{ID: id}, EntryID: 1}
			if id == 1 {
				line.AccountID = 100
				line.Debit = amount.FromInt64(100)
			} else {
				line.AccountID = 200
				line.Credit = amount.FromInt64(100)
			}
			return line, nil
		}
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
		}
		svc := accounting.NewReconcileService(lines, accounting.AccountPartialReconcileDAOMock{}, accounting.AccountFullReconcileDAOMock{}, movements, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("exceeds balance", func(t *testing.T) {
		lines := accounting.JournalLineDAOMock{}
		lines.FindFunc = func(_ context.Context, id uint64) (*accounting.JournalLine, error) {
			line := &accounting.JournalLine{Base: model.Base{ID: id}, EntryID: 1, AccountID: 100}
			if id == 1 {
				line.Debit = amount.FromInt64(100)
			} else {
				line.Credit = amount.FromInt64(100)
			}
			return line, nil
		}
		movements := accounting.JournalEntryDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
		}
		svc := accounting.NewReconcileService(lines, accounting.AccountPartialReconcileDAOMock{}, accounting.AccountFullReconcileDAOMock{}, movements, accounting.TransactionerMock{})
		app := reconcileTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reconciliations", `{"debit_line_id":1,"credit_line_id":2,"amount":150}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestWriteReconcileError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "line not found", err: accounting.ErrLineNotFound, status: http.StatusNotFound},
		{name: "different account", err: accounting.ErrLinesDifferentAccount, status: http.StatusUnprocessableEntity},
		{name: "amount", err: accounting.ErrReconcileAmount, status: http.StatusUnprocessableEntity},
		{name: "exceeds balance", err: accounting.ErrReconcileExceedsBalance, status: http.StatusUnprocessableEntity},
		{name: "invalid line", err: accounting.ErrInvalidLine, status: http.StatusUnprocessableEntity},
		{name: "not in organization", err: accounting.ErrLineNotInOrganization, status: http.StatusNotFound},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeReconcileError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
