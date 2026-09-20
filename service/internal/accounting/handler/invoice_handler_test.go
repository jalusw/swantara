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
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func newInvoiceService(invoices accounting.InvoiceDAOMock, lines accounting.InvoiceLineDAOMock, taxes accounting.InvoiceTaxDAOMock, accounts accounting.AccountLookupMock) accounting.InvoiceService {
	return accounting.NewInvoiceService(invoices, lines, taxes, accounting.PosterMock{}, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(accounting.SequenceDAOMock{}), accounting.TransactionerMock{})
}

func invoiceTestApp(t *testing.T, withTenant bool, svc accounting.InvoiceService) *fiber.App {
	t.Helper()
	h := NewInvoiceHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
		h.RegisterSupplierBills(api, guards)
	})
}

func receivableAccountsMock() accounting.AccountLookupMock {
	accounts := accounting.AccountLookupMock{}
	accounts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
		return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 200}, OrganizationID: 10, Type: "receivable", Active: true}}}, nil
	}
	return accounts
}

func TestInvoiceHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Invoice], error) {
			return &query.Page[accounting.Invoice]{Items: []*accounting.Invoice{sampleInvoice()}, Count: 1}, nil
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Invoice], error) {
			return &query.Page[accounting.Invoice]{Items: []*accounting.Invoice{sampleInvoice()}, Count: 1}, nil
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := invoiceTestApp(t, true, newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := invoiceTestApp(t, false, newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Invoice], error) {
			return nil, errors.New("boom")
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestInvoiceHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return sampleInvoice(), nil
		}
		lines := accounting.InvoiceLineDAOMock{}
		lines.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
			return []*accounting.InvoiceLine{{Base: model.Base{ID: 1}, InvoiceID: 1}}, nil
		}
		taxes := accounting.InvoiceTaxDAOMock{}
		taxes.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceTax, error) {
			return []*accounting.InvoiceTax{{Base: model.Base{ID: 1}, InvoiceID: 1}}, nil
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, lines, taxes, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := invoiceTestApp(t, true, newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return nil, nil
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(99)}, nil
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return nil, errors.New("boom")
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("lines error", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return sampleInvoice(), nil
		}
		lines := accounting.InvoiceLineDAOMock{}
		lines.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
			return nil, errors.New("boom")
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, lines, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("taxes error", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return sampleInvoice(), nil
		}
		taxes := accounting.InvoiceTaxDAOMock{}
		taxes.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceTax, error) {
			return nil, errors.New("boom")
		}
		app := invoiceTestApp(t, true, newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, taxes, receivableAccountsMock()))

		resp, err := doRequest(app, http.MethodGet, "/invoices/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func invoiceCreateBody() string {
	return `{"journal_id":1,"contact_id":5,"lines":[{"qty":1,"unit_price":100,"account_id":500}]}`
}

func TestInvoiceHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices", invoiceCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("success with date", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		body := `{"journal_id":1,"contact_id":5,"date":"2026-01-15","due_date":"2026-02-15","lines":[{"qty":2,"unit_price":50,"account_id":500}]}`
		resp, err := doRequest(app, http.MethodPost, "/invoices", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices", `{"journal_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		body := `{"journal_id":1,"contact_id":5,"date":"bogus","lines":[{"qty":1,"unit_price":100,"account_id":500}]}`
		resp, err := doRequest(app, http.MethodPost, "/invoices", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices", invoiceCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid line qty", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		body := `{"journal_id":1,"contact_id":5,"lines":[{"qty":0,"unit_price":100,"account_id":500}]}`
		resp, err := doRequest(app, http.MethodPost, "/invoices", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no receivable account", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, accounting.AccountLookupMock{})
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices", invoiceCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("sequence error", func(t *testing.T) {
		sequences := accounting.SequenceDAOMock{}
		sequences.ReserveFunc = func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return nil, errors.New("boom")
		}
		svc := accounting.NewInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, accounting.PosterMock{}, receivableAccountsMock(), dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(sequences), accounting.TransactionerMock{})
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices", invoiceCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestInvoiceHandlerCreateSupplierBill(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		accounts := accounting.AccountLookupMock{}
		accounts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 200}, OrganizationID: 10, Type: "payable", Active: true}}}, nil
		}
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, accounts)
		app := invoiceTestApp(t, true, svc)

		body := `{"journal_id":1,"contact_id":5,"lines":[{"qty":1,"unit_price":100,"account_id":500}]}`
		resp, err := doRequest(app, http.MethodPost, "/supplier-bills", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("no payable account", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, accounting.AccountLookupMock{})
		app := invoiceTestApp(t, true, svc)

		body := `{"journal_id":1,"contact_id":5,"lines":[{"qty":1,"unit_price":100,"account_id":500}]}`
		resp, err := doRequest(app, http.MethodPost, "/supplier-bills", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestInvoiceHandlerCreateCreditNote(t *testing.T) {
	body := `{"journal_id":2}`

	t.Run("success", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Type: accounting.InvoiceTypeCustomerInvoice, State: accounting.InvoiceStatePosted, EntryID: ptrUint64(9)}, nil
		}
		lines := accounting.InvoiceLineDAOMock{}
		lines.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
			return []*accounting.InvoiceLine{{Base: model.Base{ID: 1}, InvoiceID: 1}}, nil
		}
		taxes := accounting.InvoiceTaxDAOMock{}
		taxes.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceTax, error) {
			return []*accounting.InvoiceTax{{Base: model.Base{ID: 1}, InvoiceID: 1}}, nil
		}
		svc := newInvoiceService(invoices, lines, taxes, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices/1/credit-note", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices/abc/credit-note", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices/1/credit-note", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return nil, nil
		}
		svc := newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices/1/credit-note", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("already reversed", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Type: accounting.InvoiceTypeCustomerInvoice, State: accounting.InvoiceStatePosted, EntryID: ptrUint64(9)}, nil
		}
		invoices.FindByOriginFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return sampleInvoice(), nil
		}
		svc := newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices/1/credit-note", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestInvoiceHandlerCreateVendorCreditNote(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStatePosted, EntryID: ptrUint64(9)}, nil
		}
		lines := accounting.InvoiceLineDAOMock{}
		lines.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
			return []*accounting.InvoiceLine{{Base: model.Base{ID: 1}, InvoiceID: 1}}, nil
		}
		taxes := accounting.InvoiceTaxDAOMock{}
		taxes.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceTax, error) {
			return []*accounting.InvoiceTax{{Base: model.Base{ID: 1}, InvoiceID: 1}}, nil
		}
		svc := newInvoiceService(invoices, lines, taxes, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/supplier-bills/1/credit-note", `{"journal_id":2}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("not posted", func(t *testing.T) {
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStateDraft}, nil
		}
		svc := newInvoiceService(invoices, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/supplier-bills/1/credit-note", `{"journal_id":2}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := newInvoiceService(accounting.InvoiceDAOMock{}, accounting.InvoiceLineDAOMock{}, accounting.InvoiceTaxDAOMock{}, receivableAccountsMock())
		app := invoiceTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/invoices/1/credit-note", `{"journal_id":2,"date":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestWriteInvoiceError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: accounting.ErrInvoiceNotFound, status: http.StatusNotFound},
		{name: "not posted", err: accounting.ErrInvoiceNotPosted, status: http.StatusConflict},
		{name: "already reversed", err: accounting.ErrInvoiceReversed, status: http.StatusConflict},
		{name: "no lines", err: accounting.ErrInvoiceNoLines, status: http.StatusUnprocessableEntity},
		{name: "line qty", err: accounting.ErrInvoiceLineQty, status: http.StatusUnprocessableEntity},
		{name: "tax invalid", err: accounting.ErrInvoiceTaxInvalid, status: http.StatusUnprocessableEntity},
		{name: "no receivable", err: accounting.ErrNoReceivableAccount, status: http.StatusUnprocessableEntity},
		{name: "no payable", err: accounting.ErrNoPayableAccount, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeInvoiceError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
