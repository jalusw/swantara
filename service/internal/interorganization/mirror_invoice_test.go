package interorganization_test

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

type vendorBillCreatorMock struct {
	CreateSupplierBillFunc func(ctx context.Context, request accounting.CreateSupplierBillRequest) (*accounting.Invoice, error)
}

func (m vendorBillCreatorMock) CreateSupplierBill(ctx context.Context, request accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
	if m.CreateSupplierBillFunc != nil {
		return m.CreateSupplierBillFunc(ctx, request)
	}
	return &accounting.Invoice{}, nil
}

func mirrorInvoiceServiceForTest(
	invoices accounting.InvoiceDAO,
	lines accounting.InvoiceLineDAO,
	bills vendorBillCreatorMock,
	rule *interorganization.InterorganizationRule,
) interorganization.InterorganizationService {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				items := []*interorganization.InterorganizationRule{}
				if rule != nil {
					items = append(items, rule)
				}
				return &query.Page[interorganization.InterorganizationRule]{Items: items}, nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			CreateFunc: func(_ context.Context, transaction *interorganization.InterorganizationTransaction) (*interorganization.InterorganizationTransaction, error) {
				transaction.ID = 500
				return transaction, nil
			},
		},
	}
	svc := newInterorganizationServiceForTest(rules, trans, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	return svc.WithInvoiceMirror(invoices, lines, bills)
}

func postedCustomerInvoice() *accounting.Invoice {
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	return &accounting.Invoice{
		Base:           model.Base{ID: 11},
		OrganizationID: helper.Ptr(uint64(10)),
		Type:           accounting.InvoiceTypeCustomerInvoice,
		ContactID:      5,
		Name:           helper.Ptr("INV/00001"),
		InvoiceDate:    &date,
		State:          accounting.InvoiceStatePosted,
		PaymentState:   accounting.PaymentStateNotPaid,
		AmountTotal:    amount.FromFloat64(220),
		AmountResidual: amount.FromFloat64(220),
	}
}

func mirrorRule() *interorganization.InterorganizationRule {
	return &interorganization.InterorganizationRule{
		Base:               baseID(7),
		FromOrganizationID: helper.Ptr(uint64(10)),
		ToOrganizationID:   helper.Ptr(uint64(20)),
		AutoMirror:         true,
		SupplierContactID:  helper.Ptr(uint64(77)),
		CustomerContactID:  helper.Ptr(uint64(88)),
	}
}

func TestInterorganizationService_MirrorInvoice_CreatesSupplierBill(t *testing.T) {
	invoices := accounting.InvoiceDAOMock{
		CRUDMock: dao.CRUDMock[accounting.Invoice]{
			FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
				return postedCustomerInvoice(), nil
			},
		},
	}
	lines := accounting.InvoiceLineDAOMock{
		ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
			return []*accounting.InvoiceLine{{
				Base:        model.Base{ID: 1},
				Description: helper.Ptr("Widget"),
				Qty:         2,
				UnitPrice:   amount.FromFloat64(100),
				AccountID:   helper.Ptr(uint64(4100)),
			}}, nil
		},
	}
	var billed accounting.CreateSupplierBillRequest
	bills := vendorBillCreatorMock{
		CreateSupplierBillFunc: func(_ context.Context, request accounting.CreateSupplierBillRequest) (*accounting.Invoice, error) {
			billed = request
			return &accounting.Invoice{Base: model.Base{ID: 33}, AmountTotal: amount.FromFloat64(200), AmountResidual: amount.FromFloat64(200)}, nil
		},
	}
	svc := mirrorInvoiceServiceForTest(invoices, lines, bills, mirrorRule())

	bill, transaction, err := svc.MirrorInvoice(context.Background(), interorganization.MirrorInvoiceRequest{
		InvoiceID:        11,
		ToOrganizationID: 20,
		JournalID:        9,
		ExpenseAccountID: 5100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bill.ID != 33 {
		t.Errorf("bill = %d, want 33", bill.ID)
	}
	if billed.OrganizationID != 20 || billed.ContactID != 77 || billed.JournalID != 9 {
		t.Errorf("bill request = %+v, want org 20 contact 77 journal 9", billed)
	}
	if len(billed.Lines) != 1 || billed.Lines[0].AccountID != 5100 || billed.Lines[0].Qty != 2 {
		t.Errorf("bill lines = %+v, want 1 line on expense 5100 qty 2", billed.Lines)
	}
	if transaction.MirrorType != "invoice" || transaction.SourceType != "customer_invoice" {
		t.Errorf("transaction = %s/%s, want customer_invoice/invoice", transaction.SourceType, transaction.MirrorType)
	}
	if transaction.MirrorID == nil || *transaction.MirrorID != 33 {
		t.Errorf("mirror id = %v, want 33", transaction.MirrorID)
	}
}

func TestInterorganizationService_MirrorInvoice_Rejects(t *testing.T) {
	draft := postedCustomerInvoice()
	draft.State = accounting.InvoiceStateDraft
	vendorBill := postedCustomerInvoice()
	vendorBill.Type = accounting.InvoiceTypeSupplierBill

	tests := []struct {
		name    string
		invoice *accounting.Invoice
		rule    *interorganization.InterorganizationRule
		wantErr error
	}{
		{name: "draft source", invoice: draft, rule: mirrorRule(), wantErr: interorganization.ErrMirrorSourceNotPosted},
		{name: "non invoice source", invoice: vendorBill, rule: mirrorRule(), wantErr: interorganization.ErrMirrorSourceNotPosted},
		{name: "missing rule", invoice: postedCustomerInvoice(), rule: nil, wantErr: interorganization.ErrMirrorRuleMissing},
		{name: "disabled rule", invoice: postedCustomerInvoice(), rule: func() *interorganization.InterorganizationRule { r := mirrorRule(); r.AutoMirror = false; return r }(), wantErr: interorganization.ErrMirrorRuleDisabled},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			invoices := accounting.InvoiceDAOMock{
				CRUDMock: dao.CRUDMock[accounting.Invoice]{
					FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
						return tt.invoice, nil
					},
				},
			}
			svc := mirrorInvoiceServiceForTest(invoices, accounting.InvoiceLineDAOMock{}, vendorBillCreatorMock{}, tt.rule)
			_, _, err := svc.MirrorInvoice(context.Background(), interorganization.MirrorInvoiceRequest{
				InvoiceID:        11,
				ToOrganizationID: 20,
				JournalID:        9,
				ExpenseAccountID: 5100,
			})
			if err != tt.wantErr {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
