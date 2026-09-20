package accounting

import (
	"context"
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

func lifecycleReceivableMock() AccountLookupMock {
	return AccountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			if len(q.Filters) > 0 && q.Filters[0].Field == "type" && q.Filters[0].Value == "receivable" {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
			}
			return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
		},
	}
}

func lifecycleSequenceSvc() sequence.Service {
	return sequence.NewSequenceService(SequenceDAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "INV/00001"}, nil
		},
	})
}

func TestInvoiceService_CreateDraft(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	posterCalled := false
	invoices := InvoiceDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 1
			return invoice, nil
		},
	}
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ PostRequest) (*JournalEntry, error) {
			posterCalled = true
			return &JournalEntry{Base: model.Base{ID: 50}}, nil
		},
	}
	svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, lifecycleReceivableMock(), dao.CRUDMock[reference.Tax]{}, lifecycleSequenceSvc(), TransactionerMock{})

	invoice, err := svc.Create(ctx, CreateInvoiceRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		Draft:          true,
		Lines:          []InvoiceLineRequest{{Description: "Widget", Qty: 2, UnitPrice: 100, AccountID: 4100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invoice.State != InvoiceStateDraft {
		t.Errorf("state = %s, want draft", invoice.State)
	}
	if invoice.EntryID != nil {
		t.Errorf("entry_id = %v, want nil", invoice.EntryID)
	}
	if posterCalled {
		t.Error("poster must not be called for draft invoices")
	}
	if invoice.AmountTotal.Float64() != 200 || invoice.AmountResidual.Float64() != 200 {
		t.Errorf("total/residual = %v/%v, want 200/200", invoice.AmountTotal, invoice.AmountResidual)
	}
}

func TestInvoiceService_PostDraft(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	draft := &Invoice{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(10)), Type: InvoiceTypeCustomerInvoice, ContactID: 5, Name: helper.Ptr("INV/00001"), InvoiceDate: &date, JournalID: helper.Ptr(uint64(20)), State: InvoiceStateDraft, PaymentState: PaymentStateNotPaid, AmountUntaxed: amount.FromInt64(200), AmountTotal: amount.FromInt64(200), AmountResidual: amount.FromInt64(200)}
	var capturedLines []PostingLine
	invoices := InvoiceDAOMock{
		CRUDMock: dao.CRUDMock[Invoice]{
			FindFunc: func(_ context.Context, id uint64) (*Invoice, error) {
				if id == 7 {
					return draft, nil
				}
				return nil, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice) (*Invoice, error) {
			return invoice, nil
		},
	}
	lines := InvoiceLineDAOMock{
		ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceLine, error) {
			return []*InvoiceLine{{Base: model.Base{ID: 8}, Description: helper.Ptr("Widget"), Qty: 2, UnitPrice: amount.FromInt64(100), AccountID: helper.Ptr(uint64(4100)), PriceSubtotal: amount.FromInt64(200)}}, nil
		},
	}
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			capturedLines = request.Lines
			return &JournalEntry{Base: model.Base{ID: 50}}, nil
		},
	}
	svc := NewInvoiceService(invoices, lines, InvoiceTaxDAOMock{}, poster, lifecycleReceivableMock(), dao.CRUDMock[reference.Tax]{}, lifecycleSequenceSvc(), TransactionerMock{})

	posted, err := svc.Post(ctx, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posted.State != InvoiceStatePosted {
		t.Errorf("state = %s, want posted", posted.State)
	}
	if posted.EntryID == nil || *posted.EntryID != 50 {
		t.Errorf("entry_id = %v, want 50", posted.EntryID)
	}
	if len(capturedLines) != 2 {
		t.Fatalf("posting lines = %d, want 2", len(capturedLines))
	}
	if !capturedLines[0].Debit.Equal(amount.FromInt64(200)) || !capturedLines[1].Credit.Equal(amount.FromInt64(200)) {
		t.Errorf("lines = %+v, want AR debit 200 and revenue credit 200", capturedLines)
	}

	if _, err := svc.Post(ctx, 7, 10); err != ErrInvoiceNotDraft {
		t.Errorf("second post err = %v, want ErrInvoiceNotDraft", err)
	}
	if _, err := svc.Post(ctx, 999, 10); err != ErrInvoiceNotFound {
		t.Errorf("missing post err = %v, want ErrInvoiceNotFound", err)
	}
}

func TestInvoiceService_WriteOffBadDebt(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	open := &Invoice{Base: model.Base{ID: 9}, OrganizationID: helper.Ptr(uint64(10)), Type: InvoiceTypeCustomerInvoice, ContactID: 5, Name: helper.Ptr("INV/00002"), InvoiceDate: &date, JournalID: helper.Ptr(uint64(20)), State: InvoiceStatePosted, PaymentState: PaymentStatePartial, AmountUntaxed: amount.FromInt64(200), AmountTotal: amount.FromInt64(200), AmountResidual: amount.FromInt64(80)}
	var capturedLines []PostingLine
	invoices := InvoiceDAOMock{
		CRUDMock: dao.CRUDMock[Invoice]{
			FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
				return open, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice) (*Invoice, error) {
			return invoice, nil
		},
	}
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			capturedLines = request.Lines
			return &JournalEntry{Base: model.Base{ID: 60}}, nil
		},
	}
	svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, poster, lifecycleReceivableMock(), dao.CRUDMock[reference.Tax]{}, lifecycleSequenceSvc(), TransactionerMock{})

	writtenOff, err := svc.WriteOffBadDebt(ctx, WriteOffBadDebtRequest{InvoiceID: 9, OrganizationID: 10, ExpenseAccountID: 6000, Date: date})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !writtenOff.AmountResidual.IsZero() {
		t.Errorf("residual = %v, want zero", writtenOff.AmountResidual)
	}
	if writtenOff.PaymentState != PaymentStateBadDebt {
		t.Errorf("payment_state = %s, want bad_debt", writtenOff.PaymentState)
	}
	if len(capturedLines) != 2 {
		t.Fatalf("posting lines = %d, want 2", len(capturedLines))
	}
	if capturedLines[0].AccountID != 6000 || !capturedLines[0].Debit.Equal(amount.FromInt64(80)) {
		t.Errorf("expense line = %+v, want debit 80 on 6000", capturedLines[0])
	}
	if capturedLines[1].AccountID != 1200 || !capturedLines[1].Credit.Equal(amount.FromInt64(80)) {
		t.Errorf("receivable line = %+v, want credit 80 on 1200", capturedLines[1])
	}

	open.State = InvoiceStateDraft
	if _, err := svc.WriteOffBadDebt(ctx, WriteOffBadDebtRequest{InvoiceID: 9, OrganizationID: 10, ExpenseAccountID: 6000}); err != ErrBadDebtNotAllowed {
		t.Errorf("draft write-off err = %v, want ErrBadDebtNotAllowed", err)
	}
}

func TestInvoiceService_AgingReport(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	dueSoon := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	overdue10 := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	overdue45 := time.Date(2026, 7, 17, 0, 0, 0, 0, time.UTC)
	overdue100 := time.Date(2026, 5, 23, 0, 0, 0, 0, time.UTC)
	invoices := InvoiceDAOMock{
		ListOpenByOrganizationFunc: func(_ context.Context, _ uint64) ([]*Invoice, error) {
			return []*Invoice{
				{Base: model.Base{ID: 1}, ContactID: 5, DueDate: &dueSoon, AmountResidual: amount.FromInt64(100)},
				{Base: model.Base{ID: 2}, ContactID: 5, DueDate: &overdue10, AmountResidual: amount.FromInt64(50)},
				{Base: model.Base{ID: 3}, ContactID: 6, DueDate: &overdue45, AmountResidual: amount.FromInt64(70)},
				{Base: model.Base{ID: 4}, ContactID: 6, DueDate: &overdue100, AmountResidual: amount.FromInt64(30)},
				{Base: model.Base{ID: 5}, ContactID: 6, DueDate: nil, AmountResidual: amount.FromInt64(20)},
			}, nil
		},
	}
	svc := NewInvoiceService(invoices, InvoiceLineDAOMock{}, InvoiceTaxDAOMock{}, PosterMock{}, AccountLookupMock{}, dao.CRUDMock[reference.Tax]{}, lifecycleSequenceSvc(), TransactionerMock{})

	rows, err := svc.AgingReport(ctx, 10, &asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	if rows[0].ContactID != 5 || rows[0].Current != 100 || rows[0].Days1_30 != 50 || rows[0].Total != 150 {
		t.Errorf("contact 5 row = %+v, want current 100 / 1-30 50 / total 150", rows[0])
	}
	if rows[1].ContactID != 6 || rows[1].Days31_60 != 70 || rows[1].Over90 != 30 || rows[1].Current != 20 || rows[1].Total != 120 {
		t.Errorf("contact 6 row = %+v, want 31-60 70 / over90 30 / current 20 / total 120", rows[1])
	}
}

func TestInvoiceService_VendorDraftPost(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			if len(q.Filters) > 0 && q.Filters[0].Field == "type" && q.Filters[0].Value == "payable" {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 2100}, OrganizationID: 10, Active: true}}}, nil
			}
			return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
		},
	}
	draft := &Invoice{Base: model.Base{ID: 12}, OrganizationID: helper.Ptr(uint64(10)), Type: InvoiceTypeSupplierBill, ContactID: 7, Name: helper.Ptr("INV/00003"), InvoiceDate: &date, JournalID: helper.Ptr(uint64(20)), State: InvoiceStateDraft, PaymentState: PaymentStateNotPaid, AmountUntaxed: amount.FromInt64(300), AmountTotal: amount.FromInt64(300), AmountResidual: amount.FromInt64(300)}
	invoices := InvoiceDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 12
			return invoice, nil
		},
		CRUDMock: dao.CRUDMock[Invoice]{
			FindFunc: func(_ context.Context, _ uint64) (*Invoice, error) {
				return draft, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice) (*Invoice, error) {
			return invoice, nil
		},
	}
	lines := InvoiceLineDAOMock{
		ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*InvoiceLine, error) {
			return []*InvoiceLine{{Base: model.Base{ID: 13}, Description: helper.Ptr("Supplies"), Qty: 3, UnitPrice: amount.FromInt64(100), AccountID: helper.Ptr(uint64(6100)), PriceSubtotal: amount.FromInt64(300)}}, nil
		},
	}
	var capturedLines []PostingLine
	poster := PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request PostRequest) (*JournalEntry, error) {
			capturedLines = request.Lines
			return &JournalEntry{Base: model.Base{ID: 55}}, nil
		},
	}
	svc := NewInvoiceService(invoices, lines, InvoiceTaxDAOMock{}, poster, accounts, dao.CRUDMock[reference.Tax]{}, lifecycleSequenceSvc(), TransactionerMock{})

	created, err := svc.CreateSupplierBill(ctx, CreateSupplierBillRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      7,
		Date:           date,
		Draft:          true,
		Lines:          []InvoiceLineRequest{{Description: "Supplies", Qty: 3, UnitPrice: 100, AccountID: 6100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.State != InvoiceStateDraft || created.EntryID != nil {
		t.Errorf("draft = %s/%v, want draft/nil", created.State, created.EntryID)
	}

	posted, err := svc.Post(ctx, 12, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posted.State != InvoiceStatePosted || posted.EntryID == nil || *posted.EntryID != 55 {
		t.Errorf("posted = %s/%v, want posted/55", posted.State, posted.EntryID)
	}
	if len(capturedLines) != 2 {
		t.Fatalf("posting lines = %d, want 2", len(capturedLines))
	}
	if !capturedLines[0].Credit.Equal(amount.FromInt64(300)) || !capturedLines[1].Debit.Equal(amount.FromInt64(300)) {
		t.Errorf("lines = %+v, want payable credit 300 and expense debit 300", capturedLines)
	}
}

type ruleDAOFake struct {
	dao.CRUDMock[ReconcileRule]
	items []*ReconcileRule
}

func (f ruleDAOFake) ListByOrganization(_ context.Context, _ uint64) ([]*ReconcileRule, error) {
	return f.items, nil
}

func TestReconcileRuleEngine_SuggestMatches(t *testing.T) {
	ctx := context.Background()
	debit := &JournalLine{Base: model.Base{ID: 11}, AccountID: 1100, ContactID: helper.Ptr(uint64(5)), Debit: amount.FromInt64(100)}
	credit := &JournalLine{Base: model.Base{ID: 22}, AccountID: 1100, ContactID: helper.Ptr(uint64(5)), Credit: amount.FromInt64(100)}
	lines := JournalLineDAOMock{
		ListUnreconciledByAccountFunc: func(_ context.Context, _ uint64) ([]*JournalLine, error) {
			return []*JournalLine{debit, credit}, nil
		},
		CRUDMock: dao.CRUDMock[JournalLine]{
			FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
				if id == 11 {
					return debit, nil
				}
				return credit, nil
			},
		},
	}
	rules := ruleDAOFake{items: []*ReconcileRule{{Base: model.Base{ID: 3}, Name: helper.Ptr("Contact+Amount"), AccountID: helper.Ptr(uint64(1100)), MatchContact: true, MatchAmount: true, AmountTolerance: 0.01}}}
	engine := NewReconcileRuleEngine(rules, dao.CRUDMock[ReconcileRuleMatch]{}, lines, AccountPartialReconcileDAOMock{}, nil)

	proposals, err := engine.SuggestMatches(ctx, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(proposals) != 1 {
		t.Fatalf("proposals = %d, want 1", len(proposals))
	}
	if proposals[0].Score != 70 {
		t.Errorf("score = %d, want 70", proposals[0].Score)
	}
	if !proposals[0].Amount.Equal(amount.FromInt64(100)) {
		t.Errorf("amount = %v, want 100", proposals[0].Amount)
	}
	if proposals[0].RuleName != "Contact+Amount" || proposals[0].DebitLineID != 11 || proposals[0].CreditLineID != 22 {
		t.Errorf("proposal = %+v, want rule Contact+Amount debit 11 credit 22", proposals[0])
	}

	filtered, err := engine.SuggestMatches(ctx, 10, 71)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 0 {
		t.Errorf("filtered = %d, want 0", len(filtered))
	}
}
