package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func newPaymentService(payments accounting.PaymentDAOMock, invoices accounting.InvoiceDAOMock, accounts accounting.AccountLookupMock, journals dao.CRUDMock[reference.Journal], allocations accounting.PaymentAllocationDAOMock) accounting.PaymentService {
	return accounting.NewPaymentService(payments, invoices, accounting.PosterMock{}, accounts, journals, sequence.NewSequenceService(accounting.SequenceDAOMock{}), accounting.TransactionerMock{}).WithAllocations(allocations)
}

func paymentTestApp(t *testing.T, withTenant bool, svc accounting.PaymentService) *fiber.App {
	t.Helper()
	h := NewPaymentHandler(svc)
	return accountingTestApp(t, withTenant, func(api fiber.Router, guards httpx.RouteGuards) {
		h.Register(api, guards)
	})
}

func paymentServiceMocks() (accounting.PaymentDAOMock, accounting.InvoiceDAOMock, accounting.AccountLookupMock, dao.CRUDMock[reference.Journal]) {
	payments := accounting.PaymentDAOMock{}
	invoices := accounting.InvoiceDAOMock{}
	invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
		return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Type: accounting.InvoiceTypeCustomerInvoice, State: accounting.InvoiceStatePosted, AmountResidual: amount.FromFloat64(100), EntryID: ptrUint64(9)}, nil
	}
	accounts := accounting.AccountLookupMock{}
	accounts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
		return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 200}, OrganizationID: 10, Type: "receivable", Active: true}}}, nil
	}
	journals := dao.CRUDMock[reference.Journal]{}
	journals.FindFunc = func(_ context.Context, _ uint64) (*reference.Journal, error) {
		return &reference.Journal{Base: model.Base{ID: 1}, OrganizationID: 10, DefaultAccountID: ptrUint64(300)}, nil
	}
	return payments, invoices, accounts, journals
}

func TestPaymentHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Payment], error) {
			return &query.Page[accounting.Payment]{Items: []*accounting.Payment{samplePayment()}, Count: 1}, nil
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Payment], error) {
			return &query.Page[accounting.Payment]{Items: []*accounting.Payment{samplePayment()}, Count: 1}, nil
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[accounting.Payment], error) {
			return nil, errors.New("boom")
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPaymentHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.FindFunc = func(_ context.Context, _ uint64) (*accounting.Payment, error) {
			return samplePayment(), nil
		}
		allocations := accounting.PaymentAllocationDAOMock{}
		allocations.ListByPaymentFunc = func(_ context.Context, _ uint64) ([]*accounting.PaymentAllocation, error) {
			return []*accounting.PaymentAllocation{{Base: model.Base{ID: 1}, PaymentID: 1, InvoiceID: 1, Amount: 100}}, nil
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, allocations)
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.FindFunc = func(_ context.Context, _ uint64) (*accounting.Payment, error) {
			return nil, nil
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.FindFunc = func(_ context.Context, _ uint64) (*accounting.Payment, error) {
			return &accounting.Payment{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(99)}, nil
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.FindFunc = func(_ context.Context, _ uint64) (*accounting.Payment, error) {
			return nil, errors.New("boom")
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("allocations error", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		payments.FindFunc = func(_ context.Context, _ uint64) (*accounting.Payment, error) {
			return samplePayment(), nil
		}
		allocations := accounting.PaymentAllocationDAOMock{}
		allocations.ListByPaymentFunc = func(_ context.Context, _ uint64) ([]*accounting.PaymentAllocation, error) {
			return nil, errors.New("boom")
		}
		svc := newPaymentService(payments, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, allocations)
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodGet, "/payments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func paymentCreateBody() string {
	return `{"contact_id":5,"journal_id":1,"amount":100,"invoice_ids":[1]}`
}

func TestPaymentHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payments, invoices, accounts, journals := paymentServiceMocks()
		svc := newPaymentService(payments, invoices, accounts, journals, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("success with date", func(t *testing.T) {
		payments, invoices, accounts, journals := paymentServiceMocks()
		svc := newPaymentService(payments, invoices, accounts, journals, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		body := `{"contact_id":5,"journal_id":1,"amount":100,"date":"2026-01-15","invoice_ids":[1]}`
		resp, err := doRequest(app, http.MethodPost, "/payments", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", `{"contact_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", `{"contact_id":5,"journal_id":1,"amount":100,"date":"bogus","invoice_ids":[1]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, false, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", `{"contact_id":5,"journal_id":1,"amount":0,"invoice_ids":[1]}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invoice not found", func(t *testing.T) {
		payments, _, accounts, journals := paymentServiceMocks()
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return nil, nil
		}
		svc := newPaymentService(payments, invoices, accounts, journals, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("no bank account", func(t *testing.T) {
		payments, invoices, accounts, _ := paymentServiceMocks()
		journals := dao.CRUDMock[reference.Journal]{}
		journals.FindFunc = func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 1}, OrganizationID: 10}, nil
		}
		svc := newPaymentService(payments, invoices, accounts, journals, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no receivable account", func(t *testing.T) {
		payments, invoices, _, journals := paymentServiceMocks()
		accounts := accounting.AccountLookupMock{}
		svc := newPaymentService(payments, invoices, accounts, journals, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("sequence error", func(t *testing.T) {
		payments, invoices, accounts, journals := paymentServiceMocks()
		sequences := accounting.SequenceDAOMock{}
		sequences.ReserveFunc = func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return nil, errors.New("boom")
		}
		svc := accounting.NewPaymentService(payments, invoices, accounting.PosterMock{}, accounts, journals, sequence.NewSequenceService(sequences), accounting.TransactionerMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("poster error", func(t *testing.T) {
		payments, invoices, accounts, journals := paymentServiceMocks()
		poster := accounting.PosterMock{}
		poster.PostTxFunc = func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return nil, errors.New("boom")
		}
		svc := accounting.NewPaymentService(payments, invoices, poster, accounts, journals, sequence.NewSequenceService(accounting.SequenceDAOMock{}), accounting.TransactionerMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestPaymentHandlerCreateOutbound(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStatePosted, AmountResidual: amount.FromFloat64(100), EntryID: ptrUint64(9)}, nil
		}
		accounts := accounting.AccountLookupMock{}
		accounts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 200}, OrganizationID: 10, Type: "payable", Active: true}}}, nil
		}
		journals := dao.CRUDMock[reference.Journal]{}
		journals.FindFunc = func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 1}, OrganizationID: 10, DefaultAccountID: ptrUint64(300)}, nil
		}
		svc := newPaymentService(payments, invoices, accounts, journals, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments/outbound", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("no payable account", func(t *testing.T) {
		payments := accounting.PaymentDAOMock{}
		invoices := accounting.InvoiceDAOMock{}
		invoices.FindFunc = func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 1}, OrganizationID: ptrUint64(10), Type: accounting.InvoiceTypeSupplierBill, State: accounting.InvoiceStatePosted, AmountResidual: amount.FromFloat64(100), EntryID: ptrUint64(9)}, nil
		}
		accounts := accounting.AccountLookupMock{}
		journals := dao.CRUDMock[reference.Journal]{}
		journals.FindFunc = func(_ context.Context, _ uint64) (*reference.Journal, error) {
			return &reference.Journal{Base: model.Base{ID: 1}, OrganizationID: 10, DefaultAccountID: ptrUint64(300)}, nil
		}
		svc := newPaymentService(payments, invoices, accounts, journals, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments/outbound", paymentCreateBody())
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		svc := newPaymentService(accounting.PaymentDAOMock{}, accounting.InvoiceDAOMock{}, accounting.AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, accounting.PaymentAllocationDAOMock{})
		app := paymentTestApp(t, true, svc)

		resp, err := doRequest(app, http.MethodPost, "/payments/outbound", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestWritePaymentError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "payment not found", err: accounting.ErrPaymentNotFound, status: http.StatusNotFound},
		{name: "invoice not found", err: accounting.ErrInvoiceNotFound, status: http.StatusNotFound},
		{name: "amount", err: accounting.ErrPaymentAmount, status: http.StatusUnprocessableEntity},
		{name: "no invoices", err: accounting.ErrPaymentNoInvoices, status: http.StatusUnprocessableEntity},
		{name: "over allocation", err: accounting.ErrOverAllocation, status: http.StatusUnprocessableEntity},
		{name: "invoice not posted", err: accounting.ErrInvoiceNotPosted, status: http.StatusUnprocessableEntity},
		{name: "no bank", err: accounting.ErrNoBankAccount, status: http.StatusUnprocessableEntity},
		{name: "no receivable", err: accounting.ErrNoReceivableAccount, status: http.StatusUnprocessableEntity},
		{name: "no payable", err: accounting.ErrNoPayableAccount, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writePaymentError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
