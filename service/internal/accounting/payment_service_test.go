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

func TestPaymentService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("AllocatesAndReconcilesInvoice", func(t *testing.T) {
		date := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

		var capturedAllocations []*PaymentAllocation
		var updatedInvoices []*Invoice
		payments := PaymentDAOMock{
			CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
				payment.ID = 1
				capturedAllocations = allocations
				return payment, nil
			},
		}
		openInvoice := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(220)}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return openInvoice, nil },
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice) (*Invoice, error) {
				updatedInvoices = append(updatedInvoices, invoice)
				return invoice, nil
			},
		}
		var capturedLines []PostingLine
		poster := PosterMock{
			PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
				capturedLines = request.Lines
				return &JournalEntry{Base: model.Base{ID: 60}}, nil
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
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		sequences := sequence.NewSequenceService(SequenceDAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return &sequence.Reservation{Value: 1, Number: "PAY/00001"}, nil
			},
		})
		svc := NewPaymentService(payments, invoices, poster, accounts, journals, sequences, TransactionerMock{})

		payment, err := svc.Create(ctx, CreatePaymentRequest{
			OrganizationID: 10,
			ContactID:      5,
			JournalID:      20,
			Amount:         220,
			Date:           date,
			InvoiceIDs:     []uint64{11},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if payment.Name == nil || *payment.Name != "PAY/00001" {
			t.Errorf("name = %v, want PAY/00001", payment.Name)
		}
		if payment.Type != PaymentTypeInbound || payment.State != PaymentStateReconciled {
			t.Errorf("type/state = %s/%s, want inbound/reconciled", payment.Type, payment.State)
		}
		if len(capturedLines) != 2 {
			t.Fatalf("posting lines = %d, want 2", len(capturedLines))
		}
		if capturedLines[0].Debit.Float64() != 220 || capturedLines[1].Credit.Float64() != 220 {
			t.Errorf("lines = %v/%v, want bank debit 220 / AR credit 220", capturedLines[0].Debit, capturedLines[1].Credit)
		}
		if len(capturedAllocations) != 1 || capturedAllocations[0].InvoiceID != 11 || capturedAllocations[0].Amount != 220 {
			t.Errorf("allocations = %+v, want single 220 on invoice 11", capturedAllocations)
		}
		if len(updatedInvoices) != 1 || updatedInvoices[0].AmountResidual.Float64() != 0 || updatedInvoices[0].PaymentState != PaymentStatePaid {
			t.Errorf("invoice updated = residual %v state %s, want 0/paid", updatedInvoices[0].AmountResidual, updatedInvoices[0].PaymentState)
		}
	})

	t.Run("PartialPaymentKeepsInvoicePartial", func(t *testing.T) {
		var updatedInvoices []*Invoice
		payments := PaymentDAOMock{
			CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
				payment.ID = 1
				return payment, nil
			},
		}
		openInvoice := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(220)}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return openInvoice, nil },
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice) (*Invoice, error) {
				updatedInvoices = append(updatedInvoices, invoice)
				return invoice, nil
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(payments, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		payment, err := svc.Create(ctx, CreatePaymentRequest{
			OrganizationID: 10,
			ContactID:      5,
			JournalID:      20,
			Amount:         100,
			InvoiceIDs:     []uint64{11},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payment.Amount != 100 {
			t.Errorf("amount = %v, want 100", payment.Amount)
		}
		if len(updatedInvoices) != 1 || updatedInvoices[0].AmountResidual.Float64() != 120 || updatedInvoices[0].PaymentState != PaymentStatePartial {
			t.Errorf("invoice updated = residual %v state %s, want 120/partial", updatedInvoices[0].AmountResidual, updatedInvoices[0].PaymentState)
		}
	})

	t.Run("AllocatesAcrossMultipleInvoices", func(t *testing.T) {
		invoiceByID := map[uint64]*Invoice{
			11: {Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(150)},
			12: {Base: model.Base{ID: 12}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(100)},
		}
		var capturedAllocations []*PaymentAllocation
		payments := PaymentDAOMock{
			CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
				payment.ID = 1
				capturedAllocations = allocations
				return payment, nil
			},
		}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, id uint64) (*Invoice, error) { return invoiceByID[id], nil },
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(payments, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		payment, err := svc.Create(ctx, CreatePaymentRequest{
			OrganizationID: 10,
			ContactID:      5,
			JournalID:      20,
			Amount:         200,
			InvoiceIDs:     []uint64{11, 12},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payment.Amount != 200 {
			t.Errorf("amount = %v, want 200", payment.Amount)
		}
		if len(capturedAllocations) != 2 {
			t.Fatalf("allocations = %d, want 2", len(capturedAllocations))
		}
		if capturedAllocations[0].InvoiceID != 11 || capturedAllocations[0].Amount != 150 {
			t.Errorf("allocation 0 = invoice %d amount %v, want 11/150", capturedAllocations[0].InvoiceID, capturedAllocations[0].Amount)
		}
		if capturedAllocations[1].InvoiceID != 12 || capturedAllocations[1].Amount != 50 {
			t.Errorf("allocation 1 = invoice %d amount %v, want 12/50", capturedAllocations[1].InvoiceID, capturedAllocations[1].Amount)
		}
		if invoiceByID[11].AmountResidual.Float64() != 0 || invoiceByID[11].PaymentState != PaymentStatePaid {
			t.Errorf("invoice 11 = residual %v state %s, want 0/paid", invoiceByID[11].AmountResidual, invoiceByID[11].PaymentState)
		}
		if invoiceByID[12].AmountResidual.Float64() != 50 || invoiceByID[12].PaymentState != PaymentStatePartial {
			t.Errorf("invoice 12 = residual %v state %s, want 50/partial", invoiceByID[12].AmountResidual, invoiceByID[12].PaymentState)
		}
	})

	t.Run("RejectsOverAllocation", func(t *testing.T) {
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
					return &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, AmountResidual: amount.FromInt64(220)}, nil
				},
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{
			OrganizationID: 10,
			ContactID:      5,
			JournalID:      20,
			Amount:         300,
			InvoiceIDs:     []uint64{11},
		})
		if helper.AssertError(t, err, true, ErrOverAllocation) {
			return
		}
	})

	t.Run("RejectsNonPositiveAmount", func(t *testing.T) {
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 0, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrPaymentAmount) {
			return
		}
	})

	t.Run("RejectsWithoutInvoices", func(t *testing.T) {
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100})
		if helper.AssertError(t, err, true, ErrPaymentNoInvoices) {
			return
		}
	})

	t.Run("RejectsWhenNoBankAccount", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrNoBankAccount) {
			return
		}
	})

	t.Run("RejectsWhenJournalNotFound", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) { return nil, nil },
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrEntryNotFound) {
			return
		}
	})

	t.Run("PropagatesJournalError", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return nil, errors.New("db down")
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})

	t.Run("RejectsWhenNoReceivableAccount", func(t *testing.T) {
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrNoReceivableAccount) {
			return
		}
	})

	t.Run("PropagatesAccountListError", func(t *testing.T) {
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return nil, errors.New("db down")
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})

	t.Run("RejectsWhenInvoiceNotFound", func(t *testing.T) {
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return nil, nil },
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrInvoiceNotFound) {
			return
		}
	})

	t.Run("RejectsWhenInvoiceNotPosted", func(t *testing.T) {
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
					return &Invoice{Base: model.Base{ID: 11}, State: InvoiceStateDraft, AmountResidual: amount.FromInt64(220)}, nil
				},
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrInvoiceNotPosted) {
			return
		}
	})

	t.Run("PropagatesInvoiceFindError", func(t *testing.T) {
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return nil, errors.New("db down") },
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})

	t.Run("SkipsFullyPaidInvoices", func(t *testing.T) {
		invoiceByID := map[uint64]*Invoice{
			11: {Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStatePaid, AmountResidual: amount.FromInt64(0)},
			12: {Base: model.Base{ID: 12}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(100)},
		}
		var capturedAllocations []*PaymentAllocation
		payments := PaymentDAOMock{
			CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
				payment.ID = 1
				capturedAllocations = allocations
				return payment, nil
			},
		}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, id uint64) (*Invoice, error) { return invoiceByID[id], nil },
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(payments, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		if _, err := svc.Create(ctx, CreatePaymentRequest{
			OrganizationID: 10,
			ContactID:      5,
			JournalID:      20,
			Amount:         100,
			InvoiceIDs:     []uint64{11, 12},
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(capturedAllocations) != 1 || capturedAllocations[0].InvoiceID != 12 || capturedAllocations[0].Amount != 100 {
			t.Errorf("allocations = %+v, want only invoice 12 for 100", capturedAllocations)
		}
	})

	t.Run("PropagatesPostError", func(t *testing.T) {
		openInvoice := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(220)}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return openInvoice, nil },
			},
		}
		poster := PosterMock{
			PostTxFunc: func(_ context.Context, _ *gorm.DB, _ PostRequest) (*JournalEntry, error) {
				return nil, errors.New("post failed")
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, invoices, poster, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, JournalID: 20, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})

	t.Run("PropagatesPaymentCreateError", func(t *testing.T) {
		openInvoice := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(220)}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return openInvoice, nil },
			},
		}
		payments := PaymentDAOMock{
			CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, _ *Payment, _ []*PaymentAllocation) (*Payment, error) {
				return nil, errors.New("insert failed")
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(payments, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, JournalID: 20, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})

	t.Run("PropagatesInvoiceUpdateError", func(t *testing.T) {
		openInvoice := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(220)}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return openInvoice, nil },
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *Invoice) (*Invoice, error) {
				return nil, errors.New("update failed")
			},
		}
		payments := PaymentDAOMock{
			CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, _ []*PaymentAllocation) (*Payment, error) {
				payment.ID = 1
				return payment, nil
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(payments, invoices, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, JournalID: 20, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})

	t.Run("PropagatesSequenceError", func(t *testing.T) {
		openInvoice := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(220)}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return openInvoice, nil },
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		sequences := sequence.NewSequenceService(SequenceDAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return nil, errors.New("sequence failed")
			},
		})
		svc := NewPaymentService(PaymentDAOMock{}, invoices, PosterMock{}, accounts, journals, sequences, TransactionerMock{})

		_, err := svc.Create(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, JournalID: 20, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})
}

func TestPaymentService_CreateOutbound(t *testing.T) {
	ctx := context.Background()

	t.Run("PaysVendorInvoices", func(t *testing.T) {
		date := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)

		var capturedLines []PostingLine
		poster := PosterMock{
			PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
				capturedLines = request.Lines
				return &JournalEntry{Base: model.Base{ID: 60}}, nil
			},
		}
		payments := PaymentDAOMock{
			CreateWithAllocationsTxFunc: func(_ context.Context, _ *gorm.DB, payment *Payment, allocations []*PaymentAllocation) (*Payment, error) {
				payment.ID = 1
				return payment, nil
			},
		}
		openInvoice := &Invoice{Base: model.Base{ID: 11}, State: InvoiceStatePosted, PaymentState: PaymentStateNotPaid, AmountResidual: amount.FromInt64(220)}
		invoices := InvoiceDAOMock{
			CRUDMock: dao.CRUDMock[Invoice]{
				FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) { return openInvoice, nil },
			},
		}
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
				if len(q.Filters) > 0 && q.Filters[0].Value == "payable" {
					return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 2100}, OrganizationID: 10, Active: true}}}, nil
				}
				return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(payments, invoices, poster, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		payment, err := svc.CreateOutbound(ctx, CreatePaymentRequest{
			OrganizationID: 10,
			ContactID:      5,
			JournalID:      20,
			Amount:         220,
			Date:           date,
			InvoiceIDs:     []uint64{11},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if payment.Type != PaymentTypeOutbound || payment.State != PaymentStateReconciled {
			t.Errorf("type/state = %s/%s, want outbound/reconciled", payment.Type, payment.State)
		}
		if len(capturedLines) != 2 {
			t.Fatalf("posting lines = %d, want 2", len(capturedLines))
		}
		if capturedLines[0].AccountID != 2100 || capturedLines[1].AccountID != 1100 {
			t.Errorf("lines = %d/%d, want payable debit 2100 / bank credit 1100", capturedLines[0].AccountID, capturedLines[1].AccountID)
		}
		if capturedLines[0].Debit.Float64() != 220 || capturedLines[1].Credit.Float64() != 220 {
			t.Errorf("lines = %v/%v, want 220 both sides", capturedLines[0].Debit, capturedLines[1].Credit)
		}
	})

	t.Run("RejectsWhenNoPayableAccount", func(t *testing.T) {
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.CreateOutbound(ctx, CreatePaymentRequest{
			OrganizationID: 10,
			ContactID:      5,
			JournalID:      20,
			Amount:         100,
			InvoiceIDs:     []uint64{11},
		})
		if helper.AssertError(t, err, true, ErrNoPayableAccount) {
			return
		}
	})

	t.Run("RejectsNonPositiveAmount", func(t *testing.T) {
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.CreateOutbound(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 0, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrPaymentAmount) {
			return
		}
	})

	t.Run("RejectsWithoutInvoices", func(t *testing.T) {
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Journal]{}, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.CreateOutbound(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100})
		if helper.AssertError(t, err, true, ErrPaymentNoInvoices) {
			return
		}
	})

	t.Run("RejectsWhenNoBankAccount", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.CreateOutbound(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrNoBankAccount) {
			return
		}
	})

	t.Run("RejectsWhenJournalNotFound", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) { return nil, nil },
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.CreateOutbound(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, ErrEntryNotFound) {
			return
		}
	})

	t.Run("PropagatesJournalError", func(t *testing.T) {
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return nil, errors.New("db down")
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, AccountLookupMock{}, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.CreateOutbound(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})

	t.Run("PropagatesAccountListError", func(t *testing.T) {
		accounts := AccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return nil, errors.New("db down")
			},
		}
		journals := dao.CRUDMock[reference.Journal]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Journal, error) {
				return &reference.Journal{Base: model.Base{ID: 20}, DefaultAccountID: helper.Ptr(uint64(1100))}, nil
			},
		}
		svc := NewPaymentService(PaymentDAOMock{}, InvoiceDAOMock{}, PosterMock{}, accounts, journals, sequence.NewSequenceService(SequenceDAOMock{}), TransactionerMock{})

		_, err := svc.CreateOutbound(ctx, CreatePaymentRequest{OrganizationID: 10, ContactID: 5, Amount: 100, InvoiceIDs: []uint64{11}})
		if helper.AssertError(t, err, true, nil) {
			return
		}
	})
}
