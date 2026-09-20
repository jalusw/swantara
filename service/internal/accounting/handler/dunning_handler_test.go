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
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func reminderTestApp(t *testing.T, withTenant bool, svc accounting.ReminderService) *fiber.App {
	t.Helper()
	h := NewReminderHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func reminderServiceMocks() (accounting.InvoiceDAOMock, accounting.ReminderLevelDAOMock, accounting.ReminderActionDAOMock) {
	invoices := accounting.InvoiceDAOMock{}
	invoices.ListOverdueFunc = func(_ context.Context, _ *time.Time) ([]*accounting.Invoice, error) {
		due := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		return []*accounting.Invoice{{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), ContactID: 5, DueDate: &due}}, nil
	}
	levels := accounting.ReminderLevelDAOMock{}
	levels.ListSortedFunc = func(_ context.Context) ([]*reference.ReminderLevel, error) {
		return []*reference.ReminderLevel{{Base: model.Base{ID: 1}, DaysOverdue: 0}}, nil
	}
	return invoices, levels, accounting.ReminderActionDAOMock{}
}

func TestReminderHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		actions := accounting.ReminderActionDAOMock{}
		actions.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.ReminderAction], error) {
			return &query.Page[accounting.ReminderAction]{Items: []*accounting.ReminderAction{{Base: model.Base{ID: 1}, ContactID: 5, InvoiceID: 1, LevelID: 1}}, Count: 1}, nil
		}
		app := reminderTestApp(t, true, accounting.NewReminderService(accounting.InvoiceDAOMock{}, accounting.ReminderLevelDAOMock{}, actions, accounting.TransactionerMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reminder/actions/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := reminderTestApp(t, true, accounting.NewReminderService(accounting.InvoiceDAOMock{}, accounting.ReminderLevelDAOMock{}, accounting.ReminderActionDAOMock{}, accounting.TransactionerMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reminder/actions/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := reminderTestApp(t, false, accounting.NewReminderService(accounting.InvoiceDAOMock{}, accounting.ReminderLevelDAOMock{}, accounting.ReminderActionDAOMock{}, accounting.TransactionerMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reminder/actions/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		actions := accounting.ReminderActionDAOMock{}
		actions.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.ReminderAction], error) {
			return nil, errors.New("boom")
		}
		app := reminderTestApp(t, true, accounting.NewReminderService(accounting.InvoiceDAOMock{}, accounting.ReminderLevelDAOMock{}, actions, accounting.TransactionerMock{}))

		resp, err := doRequest(app, http.MethodGet, "/reminder/actions/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestReminderHandlerGenerate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		invoices, levels, _ := reminderServiceMocks()
		svc := accounting.NewReminderService(invoices, levels, accounting.ReminderActionDAOMock{}, accounting.TransactionerMock{})
		app := reminderTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reminder/actions/generate", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("success with as of", func(t *testing.T) {
		invoices, levels, _ := reminderServiceMocks()
		svc := accounting.NewReminderService(invoices, levels, accounting.ReminderActionDAOMock{}, accounting.TransactionerMock{})
		app := reminderTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reminder/actions/generate", `{"as_of":"2026-03-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		invoices, levels, _ := reminderServiceMocks()
		svc := accounting.NewReminderService(invoices, levels, accounting.ReminderActionDAOMock{}, accounting.TransactionerMock{})
		app := reminderTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/reminder/actions/generate", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid as of", func(t *testing.T) {
		invoices, levels, _ := reminderServiceMocks()
		svc := accounting.NewReminderService(invoices, levels, accounting.ReminderActionDAOMock{}, accounting.TransactionerMock{})
		app := reminderTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reminder/actions/generate", `{"as_of":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no levels", func(t *testing.T) {
		invoices, _, _ := reminderServiceMocks()
		levels := accounting.ReminderLevelDAOMock{}
		levels.ListSortedFunc = func(_ context.Context) ([]*reference.ReminderLevel, error) {
			return nil, nil
		}
		svc := accounting.NewReminderService(invoices, levels, accounting.ReminderActionDAOMock{}, accounting.TransactionerMock{})
		app := reminderTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/reminder/actions/generate", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestWriteReminderError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "no level", err: accounting.ErrReminderLevelNotFound, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeReminderError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
