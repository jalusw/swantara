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

func settlementTestService(invoices InvoiceDAO, poster Poster) InvoiceService {
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
		},
	}
	taxes := dao.CRUDMock[reference.Tax]{}
	sequences := sequence.NewSequenceService(SequenceDAOMock{})
	svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, taxes, sequences, TransactionerMock{})
	return svc.WithSettlements(InvoiceCreditApplicationDAOMock{}, InvoiceContraSettlementDAOMock{}).WithDownPayments(DownPaymentLinkDAOMock{})
}

func postedInvoice(id uint64, invoiceType string, contact uint64, residual float64) *Invoice {
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	return &Invoice{
		Base:           model.Base{ID: id},
		OrganizationID: helper.Ptr(uint64(10)),
		Type:           invoiceType,
		ContactID:      contact,
		Name:           helper.Ptr("INV"),
		InvoiceDate:    &date,
		State:          InvoiceStatePosted,
		PaymentState:   PaymentStateNotPaid,
		AmountTotal:    amount.FromFloat64(residual),
		AmountResidual: amount.FromFloat64(residual),
	}
}

func TestInvoiceService_ApplyCredit(t *testing.T) {
	tests := []struct {
		name           string
		invoice        *Invoice
		credit         *Invoice
		amount         float64
		wantApplied    float64
		wantInvState   string
		wantInvRemain  float64
		wantCredState  string
		wantCredRemain float64
		wantErr        error
	}{
		{
			name:           "full settle defaults to invoice residual",
			invoice:        postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 220),
			credit:         postedInvoice(12, InvoiceTypeCustomerCredit, 5, 220),
			amount:         0,
			wantApplied:    220,
			wantInvState:   PaymentStatePaid,
			wantInvRemain:  0,
			wantCredState:  PaymentStatePaid,
			wantCredRemain: 0,
		},
		{
			name:           "partial leaves both open",
			invoice:        postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 220),
			credit:         postedInvoice(12, InvoiceTypeCustomerCredit, 5, 220),
			amount:         120,
			wantApplied:    120,
			wantInvState:   PaymentStatePartial,
			wantInvRemain:  100,
			wantCredState:  PaymentStatePartial,
			wantCredRemain: 100,
		},
		{
			name:           "auto caps at smaller credit residual",
			invoice:        postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 220),
			credit:         postedInvoice(12, InvoiceTypeCustomerCredit, 5, 80),
			amount:         0,
			wantApplied:    80,
			wantInvState:   PaymentStatePartial,
			wantInvRemain:  140,
			wantCredState:  PaymentStatePaid,
			wantCredRemain: 0,
		},
		{
			name:           "supplier pair settles",
			invoice:        postedInvoice(11, InvoiceTypeSupplierBill, 7, 150),
			credit:         postedInvoice(12, InvoiceTypeSupplierCredit, 7, 150),
			amount:         0,
			wantApplied:    150,
			wantInvState:   PaymentStatePaid,
			wantInvRemain:  0,
			wantCredState:  PaymentStatePaid,
			wantCredRemain: 0,
		},
		{
			name:    "over apply fails",
			invoice: postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 100),
			credit:  postedInvoice(12, InvoiceTypeCustomerCredit, 5, 220),
			amount:  150,
			wantErr: ErrOverAllocation,
		},
		{
			name:    "mismatched types fail",
			invoice: postedInvoice(11, InvoiceTypeSupplierBill, 5, 100),
			credit:  postedInvoice(12, InvoiceTypeCustomerCredit, 5, 100),
			amount:  50,
			wantErr: ErrCreditMismatch,
		},
		{
			name:    "different contact fails",
			invoice: postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 100),
			credit:  postedInvoice(12, InvoiceTypeCustomerCredit, 6, 100),
			amount:  50,
			wantErr: ErrCreditMismatch,
		},
		{
			name: "draft invoice fails",
			invoice: func() *Invoice {
				inv := postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 100)
				inv.State = InvoiceStateDraft
				return inv
			}(),
			credit:  postedInvoice(12, InvoiceTypeCustomerCredit, 5, 100),
			amount:  50,
			wantErr: ErrInvoiceNotPosted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byID := map[uint64]*Invoice{11: tt.invoice, 12: tt.credit}
			var updated []*Invoice
			invoices := InvoiceDAOMock{
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice) (*Invoice, error) {
					updated = append(updated, invoice)
					return invoice, nil
				},
			}
			invoices.FindFunc = func(_ context.Context, id uint64) (*Invoice, error) {
				return byID[id], nil
			}
			svc := settlementTestService(invoices, PosterMock{})
			applied, err := svc.ApplyCredit(context.Background(), ApplyCreditRequest{
				OrganizationID: 10,
				InvoiceID:      11,
				CreditNoteID:   12,
				Amount:         tt.amount,
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
			if applied.Amount.Float64() != tt.wantApplied {
				t.Errorf("applied = %v, want %v", applied.Amount, tt.wantApplied)
			}
			if len(updated) != 2 {
				t.Fatalf("updated = %d, want 2", len(updated))
			}
			if updated[0].PaymentState != tt.wantInvState || updated[0].AmountResidual.Float64() != tt.wantInvRemain {
				t.Errorf("invoice = %s/%v, want %s/%v", updated[0].PaymentState, updated[0].AmountResidual, tt.wantInvState, tt.wantInvRemain)
			}
			if updated[1].PaymentState != tt.wantCredState || updated[1].AmountResidual.Float64() != tt.wantCredRemain {
				t.Errorf("credit = %s/%v, want %s/%v", updated[1].PaymentState, updated[1].AmountResidual, tt.wantCredState, tt.wantCredRemain)
			}
		})
	}
}

func TestInvoiceService_ContraSettle(t *testing.T) {
	ctx := context.Background()
	invoice := postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 200)
	bill := postedInvoice(12, InvoiceTypeSupplierBill, 5, 150)
	byID := map[uint64]*Invoice{11: invoice, 12: bill}
	var updated []*Invoice
	invoices := InvoiceDAOMock{
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, inv *Invoice) (*Invoice, error) {
			updated = append(updated, inv)
			return inv, nil
		},
	}
	invoices.FindFunc = func(_ context.Context, id uint64) (*Invoice, error) {
		return byID[id], nil
	}
	var posted []PostRequest
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			posted = append(posted, request)
			return &JournalEntry{Base: model.Base{ID: 60}}, nil
		},
	}
	svc := settlementTestService(invoices, poster)
	settlement, err := svc.ContraSettle(ctx, ContraSettleRequest{
		OrganizationID:    10,
		CustomerInvoiceID: 11,
		SupplierBillID:    12,
		JournalID:         20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settlement.Amount.Float64() != 150 {
		t.Errorf("settled = %v, want 150", settlement.Amount)
	}
	if settlement.EntryID == nil || *settlement.EntryID != 60 {
		t.Errorf("entry = %v, want 60", settlement.EntryID)
	}
	if len(updated) != 2 {
		t.Fatalf("updated = %d, want 2", len(updated))
	}
	if updated[0].AmountResidual.Float64() != 50 || updated[0].PaymentState != PaymentStatePartial {
		t.Errorf("invoice = %v/%s, want 50/partial", updated[0].AmountResidual, updated[0].PaymentState)
	}
	if updated[1].AmountResidual.Float64() != 0 || updated[1].PaymentState != PaymentStatePaid {
		t.Errorf("bill = %v/%s, want 0/paid", updated[1].AmountResidual, updated[1].PaymentState)
	}
	if len(posted) != 1 || len(posted[0].Lines) != 2 {
		t.Fatalf("posted movements = %d, want 1 with 2 lines", len(posted))
	}
	if posted[0].Lines[0].Debit.Float64() != 150 || posted[0].Lines[1].Credit.Float64() != 150 {
		t.Errorf("contra lines = %+v, want Dr 150 / Cr 150", posted[0].Lines)
	}
}

func TestInvoiceService_DeductDownPayment(t *testing.T) {
	tests := []struct {
		name          string
		final         *Invoice
		advance       *Invoice
		amount        float64
		wantDeducted  float64
		wantFinState  string
		wantFinRemain float64
		wantAdvState  string
		wantAdvRemain float64
		wantErr       error
		sameID        bool
	}{
		{
			name:          "full deduct defaults to final residual",
			final:         postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			advance:       postedInvoice(22, InvoiceTypeCustomerInvoice, 5, 300),
			amount:        0,
			wantDeducted:  300,
			wantFinState:  PaymentStatePartial,
			wantFinRemain: 700,
			wantAdvState:  PaymentStatePaid,
			wantAdvRemain: 0,
		},
		{
			name:          "partial deduct leaves both open",
			final:         postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			advance:       postedInvoice(22, InvoiceTypeCustomerInvoice, 5, 300),
			amount:        100,
			wantDeducted:  100,
			wantFinState:  PaymentStatePartial,
			wantFinRemain: 900,
			wantAdvState:  PaymentStatePartial,
			wantAdvRemain: 200,
		},
		{
			name:    "over deduct fails",
			final:   postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			advance: postedInvoice(22, InvoiceTypeCustomerInvoice, 5, 300),
			amount:  400,
			wantErr: ErrOverAllocation,
		},
		{
			name:    "same document fails",
			final:   postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			advance: postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			amount:  100,
			wantErr: ErrCreditMismatch,
			sameID:  true,
		},
		{
			name:    "credit note as advance fails",
			final:   postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			advance: postedInvoice(22, InvoiceTypeCustomerCredit, 5, 300),
			amount:  100,
			wantErr: ErrCreditMismatch,
		},
		{
			name:    "different contact fails",
			final:   postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			advance: postedInvoice(22, InvoiceTypeCustomerInvoice, 6, 300),
			amount:  100,
			wantErr: ErrCreditMismatch,
		},
		{
			name:  "draft advance fails",
			final: postedInvoice(21, InvoiceTypeCustomerInvoice, 5, 1000),
			advance: func() *Invoice {
				inv := postedInvoice(22, InvoiceTypeCustomerInvoice, 5, 300)
				inv.State = InvoiceStateDraft
				return inv
			}(),
			amount:  100,
			wantErr: ErrInvoiceNotPosted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byID := map[uint64]*Invoice{21: tt.final, 22: tt.advance}
			finalID, advanceID := uint64(21), uint64(22)
			if tt.sameID {
				byID = map[uint64]*Invoice{21: tt.final}
				advanceID = 21
			}
			var updated []*Invoice
			invoices := InvoiceDAOMock{
				UpdateTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice) (*Invoice, error) {
					updated = append(updated, invoice)
					return invoice, nil
				},
			}
			invoices.FindFunc = func(_ context.Context, id uint64) (*Invoice, error) {
				return byID[id], nil
			}
			svc := settlementTestService(invoices, PosterMock{})
			link, err := svc.DeductDownPayment(context.Background(), DeductDownPaymentRequest{
				OrganizationID:   10,
				FinalInvoiceID:   finalID,
				AdvanceInvoiceID: advanceID,
				Amount:           tt.amount,
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
			if link.Amount.Float64() != tt.wantDeducted {
				t.Errorf("deducted = %v, want %v", link.Amount, tt.wantDeducted)
			}
			if len(updated) != 2 {
				t.Fatalf("updated = %d, want 2", len(updated))
			}
			if updated[0].PaymentState != tt.wantFinState || updated[0].AmountResidual.Float64() != tt.wantFinRemain {
				t.Errorf("final = %s/%v, want %s/%v", updated[0].PaymentState, updated[0].AmountResidual, tt.wantFinState, tt.wantFinRemain)
			}
			if updated[1].PaymentState != tt.wantAdvState || updated[1].AmountResidual.Float64() != tt.wantAdvRemain {
				t.Errorf("advance = %s/%v, want %s/%v", updated[1].PaymentState, updated[1].AmountResidual, tt.wantAdvState, tt.wantAdvRemain)
			}
		})
	}
}

func TestInvoiceService_ContraSettleMismatch(t *testing.T) {
	tests := []struct {
		name    string
		invoice *Invoice
		bill    *Invoice
		wantErr error
	}{
		{
			name:    "different contact",
			invoice: postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 200),
			bill:    postedInvoice(12, InvoiceTypeSupplierBill, 6, 150),
			wantErr: ErrCreditMismatch,
		},
		{
			name:    "wrong types",
			invoice: postedInvoice(11, InvoiceTypeSupplierBill, 5, 200),
			bill:    postedInvoice(12, InvoiceTypeCustomerInvoice, 5, 150),
			wantErr: ErrInvoiceNotPosted,
		},
		{
			name:    "zero residual",
			invoice: postedInvoice(11, InvoiceTypeCustomerInvoice, 5, 0),
			bill:    postedInvoice(12, InvoiceTypeSupplierBill, 5, 150),
			wantErr: ErrOverAllocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			byID := map[uint64]*Invoice{11: tt.invoice, 12: tt.bill}
			invoices := InvoiceDAOMock{}
			invoices.FindFunc = func(_ context.Context, id uint64) (*Invoice, error) {
				return byID[id], nil
			}
			svc := settlementTestService(invoices, PosterMock{})
			_, err := svc.ContraSettle(context.Background(), ContraSettleRequest{
				OrganizationID:    10,
				CustomerInvoiceID: 11,
				SupplierBillID:    12,
				JournalID:         20,
			})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
