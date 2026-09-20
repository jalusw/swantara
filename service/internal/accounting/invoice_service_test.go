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

func TestInvoiceService_Create(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "posts invoice and computes amounts",
			run: func(t *testing.T) {
				var createdLines []*InvoiceLine
				var createdTaxes []*InvoiceTax
				var capturedLines []PostingLine
				invoices := InvoiceDAOMock{
					CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
						invoice.ID = 1
						createdLines = lines
						createdTaxes = taxes
						return invoice, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
						capturedLines = request.Lines
						return &JournalEntry{Base: model.Base{ID: 50}}, nil
					},
				}
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
						if len(q.Filters) > 0 && q.Filters[0].Field == "type" && q.Filters[0].Value == "receivable" {
							return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
						}
						return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
					},
				}
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(2100))}, nil
					},
				}
				sequences := sequence.NewSequenceService(SequenceDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
						return &sequence.Reservation{Value: 1, Number: "INV/00001"}, nil
					},
				})
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, taxes, sequences, TransactionerMock{})

				invoice, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					ContactID:      5,
					Date:           date,
					Lines: []InvoiceLineRequest{
						{ItemID: helper.Ptr(uint64(100)), Description: "Widget", Qty: 2, UnitPrice: 100, TaxIDs: helper.Int64Array{9}, AccountID: 4100},
					},
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if invoice.ID != 1 {
					t.Errorf("invoice id = %d, want 1", invoice.ID)
				}
				if invoice.Name == nil || *invoice.Name != "INV/00001" {
					t.Errorf("name = %v, want INV/00001", invoice.Name)
				}
				if invoice.Type != InvoiceTypeCustomerInvoice {
					t.Errorf("type = %s, want customer_invoice", invoice.Type)
				}
				if invoice.State != InvoiceStatePosted || invoice.PaymentState != PaymentStateNotPaid {
					t.Errorf("state/payment_state = %s/%s, want posted/not_paid", invoice.State, invoice.PaymentState)
				}
				if invoice.AmountUntaxed.Float64() != 200 || invoice.AmountTax.Float64() != 20 || invoice.AmountTotal.Float64() != 220 || invoice.AmountResidual.Float64() != 220 {
					t.Errorf("amounts = %v/%v/%v/%v, want 200/20/220/220", invoice.AmountUntaxed, invoice.AmountTax, invoice.AmountTotal, invoice.AmountResidual)
				}
				if len(capturedLines) != 3 {
					t.Fatalf("posting lines = %d, want 3", len(capturedLines))
				}
				if !capturedLines[0].Debit.Equal(amount.FromFloat64(220)) {
					t.Errorf("AR debit = %v, want 220", capturedLines[0].Debit)
				}
				if !capturedLines[1].Credit.Equal(amount.FromFloat64(200)) {
					t.Errorf("revenue credit = %v, want 200", capturedLines[1].Credit)
				}
				if !capturedLines[2].Credit.Equal(amount.FromFloat64(20)) {
					t.Errorf("tax credit = %v, want 20", capturedLines[2].Credit)
				}
				if len(createdLines) != 1 {
					t.Fatalf("invoice lines = %d, want 1", len(createdLines))
				}
				if createdLines[0].PriceSubtotal.Float64() != 200 {
					t.Errorf("line subtotal = %v, want 200", createdLines[0].PriceSubtotal)
				}
				if len(createdTaxes) != 1 {
					t.Fatalf("invoice taxes = %d, want 1", len(createdTaxes))
				}
				if createdTaxes[0].BaseAmount.Float64() != 200 || createdTaxes[0].Amount.Float64() != 20 {
					t.Errorf("tax = base %v amount %v, want 200/20", createdTaxes[0].BaseAmount, createdTaxes[0].Amount)
				}
			},
		},
		{
			name: "rejects without lines",
			run: func(t *testing.T) {
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{OrganizationID: 10, JournalID: 20})
				helper.AssertError(t, err, true, ErrInvoiceNoLines)
			},
		},
		{
			name: "rejects when no receivable account",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, ErrNoReceivableAccount)
			},
		},
		{
			name: "rejects invalid tax",
			run: func(t *testing.T) {
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopePurchase, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(2100))}, nil
					},
				}
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, accounts, taxes, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, TaxIDs: helper.Int64Array{9}, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, ErrInvoiceTaxInvalid)
			},
		},
		{
			name: "rounds tax base from line amount",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
					},
				}
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(11.0), TaxAccountID: helper.Ptr(uint64(2100))}, nil
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, accounts, taxes, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				invoice, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines: []InvoiceLineRequest{
						{Qty: 3, UnitPrice: 100.10, TaxIDs: helper.Int64Array{9}, AccountID: 4100},
					},
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if invoice.AmountUntaxed.Float64() != 300.3 || invoice.AmountTax.Float64() != 33.033 {
					t.Errorf("amounts = %v/%v, want 300.3/33.033", invoice.AmountUntaxed, invoice.AmountTax)
				}
			},
		},
		{
			name: "rejects missing tax",
			run: func(t *testing.T) {
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return nil, nil
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, taxes, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, TaxIDs: helper.Int64Array{9}, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, ErrInvoiceTaxInvalid)
			},
		},
		{
			name: "rejects tax without account",
			run: func(t *testing.T) {
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0)}, nil
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, taxes, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, TaxIDs: helper.Int64Array{9}, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, ErrInvoiceTaxInvalid)
			},
		},
		{
			name: "propagates post error",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ PostRequest) (*JournalEntry, error) {
						return nil, errors.New("poster failed")
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates tax find error",
			run: func(t *testing.T) {
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, taxes, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, TaxIDs: helper.Int64Array{9}, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates receivable list error",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates create error",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
					},
				}
				invoices := InvoiceDAOMock{
					CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, _ *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
						return nil, errors.New("insert failed")
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.Create(ctx, CreateInvoiceRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, AccountID: 4100}},
				})
				helper.AssertError(t, err, true, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}

func TestInvoiceService_CreateCreditNote(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "mirrors original and reverses posting",
			run: func(t *testing.T) {
				var capturedReverse ReverseRequest
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, id uint64) (*Invoice, error) {
							return &Invoice{
								Base:           model.Base{ID: 7},
								OrganizationID: helper.Ptr(uint64(10)),
								EntryID:        helper.Ptr(uint64(50)),
								Type:           InvoiceTypeCustomerInvoice,
								ContactID:      5,
								State:          InvoiceStatePosted,
								PaymentState:   PaymentStateNotPaid,
								AmountUntaxed:  amount.FromInt64(200),
								AmountTax:      amount.FromInt64(20),
								AmountTotal:    amount.FromInt64(220),
								AmountResidual: amount.FromInt64(220),
							}, nil
						},
					},
					CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
						invoice.ID = 8
						return invoice, nil
					},
				}
				lines := InvoiceLineDAOMock{
					ListByInvoiceFunc: func(_ context.Context, invoiceID uint64) ([]*InvoiceLine, error) {
						return []*InvoiceLine{
							{Sequence: 10, ItemID: helper.Ptr(uint64(100)), Description: helper.Ptr("Widget"), Qty: 2, UnitPrice: amount.FromInt64(100), AccountID: helper.Ptr(uint64(4100)), PriceSubtotal: amount.FromInt64(200), SaleLineID: helper.Ptr(uint64(30))},
						}, nil
					},
				}
				taxes := InvoiceTaxDAOMock{
					ListByInvoiceFunc: func(_ context.Context, invoiceID uint64) ([]*InvoiceTax, error) {
						return []*InvoiceTax{
							{TaxID: helper.Ptr(uint64(9)), BaseAmount: amount.FromInt64(200), Amount: amount.FromInt64(20), AccountID: helper.Ptr(uint64(2100))},
						}, nil
					},
				}
				poster := PosterMock{
					ReverseFunc: func(_ context.Context, request ReverseRequest) (*JournalEntry, error) {
						capturedReverse = request
						return &JournalEntry{Base: model.Base{ID: 60}}, nil
					},
				}
				sequences := sequence.NewSequenceService(SequenceDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
						return &sequence.Reservation{Value: 2, Number: "INV/00002"}, nil
					},
				})
				svc := NewInvoiceService(invoices, lines, taxes, poster, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequences, TransactionerMock{})

				creditNote, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{
					OrganizationID:    10,
					OriginalInvoiceID: 7,
					JournalID:         20,
					Date:              date,
					Reference:         "CN-1",
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if creditNote.ID != 8 {
					t.Errorf("credit note id = %d, want 8", creditNote.ID)
				}
				if creditNote.Type != InvoiceTypeCustomerCredit {
					t.Errorf("type = %s, want customer_credit_note", creditNote.Type)
				}
				if creditNote.EntryID == nil || *creditNote.EntryID != 60 {
					t.Errorf("entry_id = %v, want 60", creditNote.EntryID)
				}
				if creditNote.OriginInvoiceID == nil || *creditNote.OriginInvoiceID != 7 {
					t.Errorf("origin_invoice_id = %v, want 7", creditNote.OriginInvoiceID)
				}
				if creditNote.AmountTotal.Float64() != 220 || creditNote.AmountResidual.Float64() != 220 {
					t.Errorf("amounts = %v/%v, want 220/220", creditNote.AmountTotal, creditNote.AmountResidual)
				}
				if capturedReverse.EntryID != 50 {
					t.Errorf("reversed entry id = %d, want 50", capturedReverse.EntryID)
				}
				if capturedReverse.OrganizationID != 10 || capturedReverse.JournalID != 20 {
					t.Errorf("reverse request org/journal = %d/%d, want 10/20", capturedReverse.OrganizationID, capturedReverse.JournalID)
				}
			},
		},
		{
			name: "rejects not posted original",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, id uint64) (*Invoice, error) {
							return &Invoice{
								Base:           model.Base{ID: 7},
								OrganizationID: helper.Ptr(uint64(10)),
								Type:           InvoiceTypeCustomerInvoice,
								State:          InvoiceStateDraft,
							}, nil
						},
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, ErrInvoiceNotPosted)
			},
		},
		{
			name: "rejects duplicate credit note",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, id uint64) (*Invoice, error) {
							return &Invoice{
								Base:           model.Base{ID: 7},
								OrganizationID: helper.Ptr(uint64(10)),
								EntryID:        helper.Ptr(uint64(50)),
								Type:           InvoiceTypeCustomerInvoice,
								State:          InvoiceStatePosted,
								AmountTotal:    amount.FromInt64(220),
							}, nil
						},
					},
					FindByOriginFunc: func(_ context.Context, originID uint64) (*Invoice, error) {
						return &Invoice{Base: model.Base{ID: 8}, Type: InvoiceTypeCustomerCredit}, nil
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, ErrInvoiceReversed)
			},
		},
		{
			name: "rejects wrong organization",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, id uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99)), State: InvoiceStatePosted}, nil
						},
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, ErrInvoiceNotFound)
			},
		},
		{
			name: "rejects original without lines",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), EntryID: helper.Ptr(uint64(50)), Type: InvoiceTypeCustomerInvoice, State: InvoiceStatePosted}, nil
						},
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, ErrInvoiceNoLines)
			},
		},
		{
			name: "propagates find error",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return nil, errors.New("db down")
						},
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates find by origin error",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), EntryID: helper.Ptr(uint64(50)), Type: InvoiceTypeCustomerInvoice, State: InvoiceStatePosted}, nil
						},
					},
					FindByOriginFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates line list error",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), EntryID: helper.Ptr(uint64(50)), Type: InvoiceTypeCustomerInvoice, State: InvoiceStatePosted}, nil
						},
					},
				}
				lines := InvoiceLineDAOMock{
					ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceLine, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewInvoiceService(invoices, lines, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates tax list error",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), EntryID: helper.Ptr(uint64(50)), Type: InvoiceTypeCustomerInvoice, State: InvoiceStatePosted}, nil
						},
					},
				}
				lines := InvoiceLineDAOMock{
					ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceLine, error) {
						return []*InvoiceLine{{Sequence: 10, PriceSubtotal: amount.FromInt64(200)}}, nil
					},
				}
				taxes := InvoiceTaxDAOMock{
					ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceTax, error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewInvoiceService(invoices, lines, taxes, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates sequence error",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), EntryID: helper.Ptr(uint64(50)), Type: InvoiceTypeCustomerInvoice, State: InvoiceStatePosted}, nil
						},
					},
				}
				lines := InvoiceLineDAOMock{
					ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceLine, error) {
						return []*InvoiceLine{{Sequence: 10, PriceSubtotal: amount.FromInt64(200)}}, nil
					},
				}
				sequences := sequence.NewSequenceService(SequenceDAOMock{
					ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
						return nil, errors.New("sequence exhausted")
					},
				})
				svc := NewInvoiceService(invoices, lines, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequences, TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates reverse error",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), EntryID: helper.Ptr(uint64(50)), Type: InvoiceTypeCustomerInvoice, State: InvoiceStatePosted}, nil
						},
					},
				}
				lines := InvoiceLineDAOMock{
					ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceLine, error) {
						return []*InvoiceLine{{Sequence: 10, PriceSubtotal: amount.FromInt64(200)}}, nil
					},
				}
				poster := PosterMock{
					ReverseFunc: func(_ context.Context, _ ReverseRequest) (*JournalEntry, error) {
						return nil, errors.New("poster failed")
					},
				}
				svc := NewInvoiceService(invoices, lines, InvoiceTaxDAOMock{}, poster, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates create error",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
							return &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), EntryID: helper.Ptr(uint64(50)), Type: InvoiceTypeCustomerInvoice, State: InvoiceStatePosted}, nil
						},
					},
					CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, _ *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
						return nil, errors.New("insert failed")
					},
				}
				lines := InvoiceLineDAOMock{
					ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceLine, error) {
						return []*InvoiceLine{{Sequence: 10, PriceSubtotal: amount.FromInt64(200)}}, nil
					},
				}
				svc := NewInvoiceService(invoices, lines, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				helper.AssertError(t, err, true, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}

func TestInvoiceService_CreateSupplierBill(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "posts supplier bill and computes amounts",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
						invoice.ID = 2
						return invoice, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
						return &JournalEntry{Base: model.Base{ID: 51}}, nil
					},
				}
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
						if len(q.Filters) > 0 && q.Filters[0].Field == "type" && q.Filters[0].Value == "payable" {
							return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 2100}, OrganizationID: 10, Active: true}}}, nil
						}
						return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
					},
				}
				taxes := dao.CRUDMock[reference.Tax]{
					FindFunc: func(_ context.Context, _ uint64) (*reference.Tax, error) {
						return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopePurchase, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(2110))}, nil
					},
				}
				svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, taxes, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

				invoice, err := svc.CreateSupplierBill(ctx, CreateSupplierBillRequest{
					OrganizationID: 10,
					JournalID:      20,
					ContactID:      5,
					Date:           date,
					Lines: []InvoiceLineRequest{
						{ItemID: helper.Ptr(uint64(100)), Description: "Raw material", Qty: 2, UnitPrice: 100, TaxIDs: helper.Int64Array{9}, AccountID: 4200},
					},
				})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if invoice.Type != InvoiceTypeSupplierBill {
					t.Errorf("type = %s, want supplier_bill", invoice.Type)
				}
				if invoice.AmountUntaxed.Float64() != 200 || invoice.AmountTax.Float64() != 20 || invoice.AmountTotal.Float64() != 220 {
					t.Errorf("amounts = %v/%v/%v, want 200/20/220", invoice.AmountUntaxed, invoice.AmountTax, invoice.AmountTotal)
				}
			},
		},
		{
			name: "rejects when no payable account",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateSupplierBill(ctx, CreateSupplierBillRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines: []InvoiceLineRequest{
						{Qty: 1, UnitPrice: 100, AccountID: 4200},
					},
				})
				helper.AssertError(t, err, true, ErrNoPayableAccount)
			},
		},
		{
			name: "propagates payable list error",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return nil, errors.New("db down")
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateSupplierBill(ctx, CreateSupplierBillRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, AccountID: 4200}},
				})
				helper.AssertError(t, err, true, nil)
			},
		},
		{
			name: "propagates post error",
			run: func(t *testing.T) {
				accounts := AccountLookupMock{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 2100}, OrganizationID: 10, Active: true}}}, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ PostRequest) (*JournalEntry, error) {
						return nil, errors.New("poster failed")
					},
				}
				svc := NewInvoiceService(InvoiceDAOMock{}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})
				_, err := svc.CreateSupplierBill(ctx, CreateSupplierBillRequest{
					OrganizationID: 10,
					JournalID:      20,
					Lines:          []InvoiceLineRequest{{Qty: 1, UnitPrice: 100, AccountID: 4200}},
				})
				helper.AssertError(t, err, true, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}

func TestInvoiceService_CreateSupplierBill_StockableRoutesToInventory(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	var captured PostRequest
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			captured = request
			return &JournalEntry{Base: model.Base{ID: 51}}, nil
		},
	}
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 2100}, OrganizationID: 10, Active: true}}}, nil
		},
	}
	products := itemResolverMock{
		FindTemplateFunc: func(_ context.Context, itemID uint64) (*struct {
			Type       string
			CategoryID *uint64
		}, error) {
			return &struct {
				Type       string
				CategoryID *uint64
			}{Type: "stockable", CategoryID: helper.Ptr(uint64(99))}, nil
		},
	}
	categories := dao.CRUDMock[reference.ItemCategory]{
		FindFunc: func(_ context.Context, id uint64) (*reference.ItemCategory, error) {
			return &reference.ItemCategory{Base: model.Base{ID: 99}, StockValuationAccountID: helper.Ptr(uint64(5000))}, nil
		},
	}
	svc := NewInvoiceService(InvoiceDAOMock{CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
		invoice.ID = 2
		return invoice, nil
	}}, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, accounts, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{}).WithProducts(products, categories)
	_, err := svc.CreateSupplierBill(ctx, CreateSupplierBillRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		Lines:          []InvoiceLineRequest{{ItemID: helper.Ptr(uint64(100)), Qty: 2, UnitPrice: 100, AccountID: 4200}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	foundInventory := false
	foundExpense := false
	for _, line := range captured.Lines {
		if line.AccountID == 5000 && line.Debit.Equal(amount.FromFloat64(200)) {
			foundInventory = true
		}
		if line.AccountID == 4200 {
			foundExpense = true
		}
	}
	if !foundInventory {
		t.Errorf("posting lines = %+v, want inventory 5000 debit 200", captured.Lines)
	}
	if foundExpense {
		t.Errorf("posting lines should not contain expense 4200 when stockable routed, got %+v", captured.Lines)
	}
}

type itemResolverMock struct {
	FindTemplateFunc func(ctx context.Context, itemID uint64) (*struct {
		Type       string
		CategoryID *uint64
	}, error)
}

func (m itemResolverMock) FindTemplate(ctx context.Context, itemID uint64) (*struct {
	Type       string
	CategoryID *uint64
}, error) {
	return m.FindTemplateFunc(ctx, itemID)
}

func TestInvoiceService_CreateVendorCreditNote(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "uses supplier type",
			run: func(t *testing.T) {
				invoices := InvoiceDAOMock{
					CRUDMock: dao.CRUDMock[Invoice]{
						FindFunc: func(_ context.Context, id uint64) (*Invoice, error) {
							return &Invoice{
								Base:           model.Base{ID: 7},
								OrganizationID: helper.Ptr(uint64(10)),
								EntryID:        helper.Ptr(uint64(50)),
								Type:           InvoiceTypeSupplierBill,
								State:          InvoiceStatePosted,
								AmountUntaxed:  amount.FromInt64(150),
								AmountTax:      amount.FromInt64(15),
								AmountTotal:    amount.FromInt64(165),
							}, nil
						},
					},
					CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
						invoice.ID = 9
						return invoice, nil
					},
				}
				lines := InvoiceLineDAOMock{
					ListByInvoiceFunc: func(_ context.Context, invoiceID uint64) ([]*InvoiceLine, error) {
						return []*InvoiceLine{{Sequence: 10, Qty: 1, UnitPrice: amount.FromInt64(150), PriceSubtotal: amount.FromInt64(150)}}, nil
					},
				}
				poster := PosterMock{
					ReverseFunc: func(_ context.Context, request ReverseRequest) (*JournalEntry, error) {
						return &JournalEntry{Base: model.Base{ID: 61}}, nil
					},
				}
				svc := NewInvoiceService(invoices, lines, InvoiceTaxDAOMock{}, poster, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

				creditNote, err := svc.CreateVendorCreditNote(ctx, CreateCreditNoteRequest{OrganizationID: 10, OriginalInvoiceID: 7, JournalID: 20})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if creditNote.Type != InvoiceTypeSupplierCredit {
					t.Errorf("type = %s, want supplier_credit_note", creditNote.Type)
				}
				if creditNote.AmountTotal.Float64() != 165 {
					t.Errorf("amount_total = %v, want 165", creditNote.AmountTotal)
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, tc.run)
	}
}
