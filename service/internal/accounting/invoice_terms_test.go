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

func termTestService(invoices InvoiceDAO, terms PaymentTermSplitter, positions TaxRuleMapper, installments InvoiceInstallmentDAO) InvoiceService {
	accounts := AccountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 1200}, OrganizationID: 10, Active: true}}}, nil
		},
	}
	taxes := dao.CRUDMock[reference.Tax]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Tax, error) {
			if id == 7 {
				return &reference.Tax{Base: model.Base{ID: 7}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(2100))}, nil
			}
			return &reference.Tax{Base: model.Base{ID: 9}, Type: reference.TaxTypePercent, Scope: reference.TaxScopeSale, Amount: helper.Ptr(10.0), TaxAccountID: helper.Ptr(uint64(2100))}, nil
		},
	}
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
	if terms != nil {
		svc = svc.WithPaymentTerms(terms)
	}
	if positions != nil {
		svc = svc.WithTaxRules(positions)
	}
	if installments != nil {
		svc = svc.WithInstallments(installments)
	}
	return svc
}

func TestInvoiceService_PaymentTermSetsDueDate(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	wantDue := date.AddDate(0, 0, 30)
	var captured *Invoice
	var capturedInstallments []*InvoiceInstallment
	invoices := InvoiceDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 1
			captured = invoice
			return invoice, nil
		},
	}
	terms := PaymentTermSplitterMock{
		SplitsFunc: func(_ context.Context, _ uint64, _ amount.Amount, _ time.Time) ([]PaymentTermSplit, error) {
			return []PaymentTermSplit{{Amount: amount.FromFloat64(220), DueDate: wantDue, Sequence: 10}}, nil
		},
	}
	installments := InvoiceInstallmentDAOMock{
		CreateManyTxFunc: func(_ context.Context, _ *gorm.DB, records []*InvoiceInstallment) error {
			capturedInstallments = append(capturedInstallments, records...)
			return nil
		},
	}
	svc := termTestService(invoices, terms, nil, installments)
	invoice, err := svc.Create(ctx, CreateInvoiceRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		PaymentTermID:  helper.Ptr(uint64(3)),
		Lines:          []InvoiceLineRequest{{Description: "Widget", Qty: 2, UnitPrice: 100, AccountID: 4100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.DueDate == nil || !captured.DueDate.Equal(wantDue) {
		t.Errorf("due date = %v, want %v", captured.DueDate, wantDue)
	}
	if captured.PaymentTermID == nil || *captured.PaymentTermID != 3 {
		t.Errorf("payment term = %v, want 3", captured.PaymentTermID)
	}
	if invoice.AmountTotal.Float64() != 200 {
		t.Errorf("total = %v, want 200", invoice.AmountTotal)
	}
	if len(capturedInstallments) != 1 {
		t.Fatalf("installments = %d, want 1", len(capturedInstallments))
	}
	if capturedInstallments[0].InvoiceID != 1 || capturedInstallments[0].State != InstallmentStatePending {
		t.Errorf("installment = %+v, want invoice 1 pending", capturedInstallments[0])
	}
	if capturedInstallments[0].DueDate == nil || !capturedInstallments[0].DueDate.Equal(wantDue) {
		t.Errorf("installment due = %v, want %v", capturedInstallments[0].DueDate, wantDue)
	}
	if capturedInstallments[0].Amount.Float64() != 220 {
		t.Errorf("installment amount = %v, want 220", capturedInstallments[0].Amount)
	}
}

func TestInvoiceService_PaymentTermMultiSplitSchedule(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	first := date.AddDate(0, 0, 15)
	second := date.AddDate(0, 0, 30)
	var captured *Invoice
	var capturedInstallments []*InvoiceInstallment
	invoices := InvoiceDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 2
			captured = invoice
			return invoice, nil
		},
	}
	terms := PaymentTermSplitterMock{
		SplitsFunc: func(_ context.Context, _ uint64, _ amount.Amount, _ time.Time) ([]PaymentTermSplit, error) {
			return []PaymentTermSplit{
				{Amount: amount.FromFloat64(100), DueDate: first, Sequence: 10},
				{Amount: amount.FromFloat64(100), DueDate: second, Sequence: 20},
			}, nil
		},
	}
	installments := InvoiceInstallmentDAOMock{
		CreateManyTxFunc: func(_ context.Context, _ *gorm.DB, records []*InvoiceInstallment) error {
			capturedInstallments = append(capturedInstallments, records...)
			return nil
		},
	}
	svc := termTestService(invoices, terms, nil, installments)
	_, err := svc.Create(ctx, CreateInvoiceRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		PaymentTermID:  helper.Ptr(uint64(3)),
		Lines:          []InvoiceLineRequest{{Description: "Widget", Qty: 2, UnitPrice: 100, AccountID: 4100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.DueDate == nil || !captured.DueDate.Equal(second) {
		t.Errorf("due date = %v, want %v", captured.DueDate, second)
	}
	if len(capturedInstallments) != 2 {
		t.Fatalf("installments = %d, want 2", len(capturedInstallments))
	}
	if !capturedInstallments[0].DueDate.Equal(first) || !capturedInstallments[1].DueDate.Equal(second) {
		t.Errorf("installment dues = %v, %v, want %v, %v", capturedInstallments[0].DueDate, capturedInstallments[1].DueDate, first, second)
	}
}

func TestInvoiceService_ExplicitDueDateWins(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	explicit := date.AddDate(0, 0, 7)
	called := false
	invoices := InvoiceDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, _ []*InvoiceLine, _ []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 1
			return invoice, nil
		},
	}
	terms := PaymentTermSplitterMock{
		SplitsFunc: func(_ context.Context, _ uint64, _ amount.Amount, _ time.Time) ([]PaymentTermSplit, error) {
			called = true
			return []PaymentTermSplit{{DueDate: date.AddDate(0, 0, 30)}}, nil
		},
	}
	svc := termTestService(invoices, terms, nil, nil)
	invoice, err := svc.Create(ctx, CreateInvoiceRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		DueDate:        &explicit,
		PaymentTermID:  helper.Ptr(uint64(3)),
		Lines:          []InvoiceLineRequest{{Description: "Widget", Qty: 1, UnitPrice: 50, AccountID: 4100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Errorf("splitter should be called to build the installment schedule")
	}
	if invoice.DueDate == nil || !invoice.DueDate.Equal(explicit) {
		t.Errorf("due date = %v, want %v", invoice.DueDate, explicit)
	}
}

func TestInvoiceService_TaxRuleRemaps(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	var capturedLines []*InvoiceLine
	var capturedTaxes []*InvoiceTax
	var captured *Invoice
	invoices := InvoiceDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, _ *gorm.DB, invoice *Invoice, lines []*InvoiceLine, taxes []*InvoiceTax) (*Invoice, error) {
			invoice.ID = 1
			captured = invoice
			capturedLines = lines
			capturedTaxes = taxes
			return invoice, nil
		},
	}
	positions := TaxRuleMapperMock{
		FindFunc: func(_ context.Context, _ uint64) (*TaxRule, error) {
			return &TaxRule{Base: model.Base{ID: 4}, OrganizationID: helper.Ptr(uint64(10)), Name: helper.Ptr("Export"), Active: true}, nil
		},
		ResolveFunc: func(_ context.Context, _ uint64, srcTaxID *uint64, srcAccountID uint64) (*ResolveResult, error) {
			if srcTaxID != nil {
				return &ResolveResult{Account: helper.Ptr(srcAccountID)}, nil
			}
			return &ResolveResult{Account: helper.Ptr(uint64(4200))}, nil
		},
	}
	svc := termTestService(invoices, nil, positions, nil)
	_, err := svc.Create(ctx, CreateInvoiceRequest{
		OrganizationID: 10,
		JournalID:      20,
		ContactID:      5,
		Date:           date,
		TaxRuleID:      helper.Ptr(uint64(4)),
		Lines:          []InvoiceLineRequest{{Description: "Export", Qty: 1, UnitPrice: 100, TaxIDs: helper.Int64Array{9}, AccountID: 4100}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedLines) != 1 {
		t.Fatalf("lines = %d, want 1", len(capturedLines))
	}
	if capturedLines[0].AccountID == nil || *capturedLines[0].AccountID != 4200 {
		t.Errorf("account = %v, want 4200", capturedLines[0].AccountID)
	}
	if len(capturedTaxes) != 0 {
		t.Errorf("taxes = %d, want 0 after exemption", len(capturedTaxes))
	}
	if captured.TaxRuleID == nil || *captured.TaxRuleID != 4 {
		t.Errorf("fiscal position = %v, want 4", captured.TaxRuleID)
	}
}
