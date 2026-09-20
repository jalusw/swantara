package accounting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func creditLimitTestService(invoices InvoiceDAO, limits ContactCreditLimiter) InvoiceService {
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
		},
	}
	taxes := dao.CRUDMock[reference.Tax]{}
	sequences := sequence.NewSequenceService(SequenceDAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "INV/00001"}, nil
		},
	})
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ PostRequest) (*JournalEntry, error) {
			return &JournalEntry{Base: model.Base{ID: 50}}, nil
		},
	}
	svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, taxes, sequences, TransactionerMock{})
	return svc.WithCreditLimits(limits)
}

func openBalance(org uint64, residual float64) *Invoice {
	return &Invoice{
		Base:           model.Base{ID: 31},
		OrganizationID: helper.Ptr(org),
		Type:           InvoiceTypeCustomerInvoice,
		ContactID:      5,
		State:          InvoiceStatePosted,
		PaymentState:   PaymentStatePartial,
		AmountTotal:    amount.FromFloat64(residual),
		AmountResidual: amount.FromFloat64(residual),
	}
}

func TestInvoiceService_CreditLimit(t *testing.T) {
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	line := []InvoiceLineRequest{{Description: "Widget", Qty: 2, UnitPrice: 100, AccountID: 4100}}

	tests := []struct {
		name    string
		limit   *float64
		open    []*Invoice
		wantErr error
	}{
		{name: "within limit creates", limit: helper.Ptr(1000.0), open: []*Invoice{openBalance(10, 700)}},
		{name: "over limit fails", limit: helper.Ptr(1000.0), open: []*Invoice{openBalance(10, 900)}, wantErr: ErrCreditLimitExceeded},
		{name: "nil limit skips check", limit: nil, open: []*Invoice{openBalance(10, 9000)}},
		{name: "other organization ignored", limit: helper.Ptr(1000.0), open: []*Invoice{openBalance(99, 9000)}},
		{name: "first invoice over limit fails", limit: helper.Ptr(100.0), open: nil, wantErr: ErrCreditLimitExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoices := InvoiceDAOMock{
				ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*Invoice, error) {
					return tt.open, nil
				},
				CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
					invoice.ID = 1
					return invoice, nil
				},
			}
			limits := ContactCreditLimiterMock{
				CreditLimitFunc: func(_ context.Context, _ uint64) (*float64, error) {
					return tt.limit, nil
				},
			}
			svc := creditLimitTestService(invoices, limits)
			_, err := svc.Create(context.Background(), CreateInvoiceRequest{
				OrganizationID: 10,
				JournalID:      20,
				ContactID:      5,
				Date:           date,
				Lines:          line,
			})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestInvoiceService_SupplierBillSkipsCreditLimit(t *testing.T) {
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	called := false
	invoices := InvoiceDAOMock{
		ListOpenByContactFunc: func(_ context.Context, _ uint64) ([]*Invoice, error) {
			called = true
			return []*Invoice{openBalance(10, 9000)}, nil
		},
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 1
			return invoice, nil
		},
	}
	limits := ContactCreditLimiterMock{
		CreditLimitFunc: func(_ context.Context, _ uint64) (*float64, error) {
			return helper.Ptr(100.0), nil
		},
	}
	svc := creditLimitTestService(invoices, limits)
	_, err := svc.CreateSupplierBill(context.Background(), CreateSupplierBillRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		Lines:          []InvoiceLineRequest{{Description: "Parts", Qty: 2, UnitPrice: 100, AccountID: 5100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Errorf("open-balance lookup should not run for supplier bills")
	}
}
