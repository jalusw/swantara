package procurement

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

func TestDAOMockDefaults(t *testing.T) {
	ctx := context.Background()

	requisition := &PurchaseRequest{}
	requisitions := PurchaseRequestDAOMock{}
	if got, err := requisitions.CreateWithLines(ctx, requisition, nil); err != nil || got != requisition {
		t.Errorf("CreateWithLines() = (%v, %v), want original requisition", got, err)
	}
	if got, err := requisitions.UpdateTx(ctx, nil, requisition); err != nil || got != requisition {
		t.Errorf("UpdateTx() = (%v, %v), want original requisition", got, err)
	}

	requisitionLines := PurchaseRequestLineDAOMock{}
	if items, err := requisitionLines.ListByRequest(ctx, 1); err != nil || len(items) != 0 {
		t.Errorf("ListByRequest() = (%v, %v), want empty", items, err)
	}
	if err := requisitionLines.ReplaceLines(ctx, 1, nil); err != nil {
		t.Errorf("ReplaceLines() error = %v, want nil", err)
	}

	order := &PurchaseOrder{}
	orders := PurchaseOrderDAOMock{}
	if got, err := orders.CreateWithLines(ctx, order, nil); err != nil || got != order {
		t.Errorf("CreateWithLines() = (%v, %v), want original order", got, err)
	}
	if got, err := orders.UpdateTx(ctx, nil, order); err != nil || got != order {
		t.Errorf("UpdateTx() = (%v, %v), want original order", got, err)
	}

	orderLines := PurchaseOrderLineDAOMock{}
	if items, err := orderLines.ListByOrder(ctx, 1); err != nil || len(items) != 0 {
		t.Errorf("ListByOrder() = (%v, %v), want empty", items, err)
	}
	if err := orderLines.ReplaceLines(ctx, 1, nil); err != nil {
		t.Errorf("ReplaceLines() error = %v, want nil", err)
	}
	line := &PurchaseOrderLine{}
	if got, err := orderLines.UpdateTx(ctx, nil, line); err != nil || got != line {
		t.Errorf("UpdateTx() = (%v, %v), want original line", got, err)
	}

	quoteRequest := &SupplierQuoteRequest{}
	quote_requests := SupplierQuoteRequestDAOMock{}
	if got, err := quote_requests.CreateWithLines(ctx, quoteRequest, nil); err != nil || got != quoteRequest {
		t.Errorf("CreateWithLines() = (%v, %v), want original quoteRequest", got, err)
	}
	if got, err := quote_requests.UpdateTx(ctx, nil, quoteRequest); err != nil || got != quoteRequest {
		t.Errorf("UpdateTx() = (%v, %v), want original quoteRequest", got, err)
	}

	rfqLines := SupplierQuoteRequestLineDAOMock{}
	if items, err := rfqLines.ListByRFQ(ctx, 1); err != nil || len(items) != 0 {
		t.Errorf("ListByRFQ() = (%v, %v), want empty", items, err)
	}

	quote := &SupplierQuote{}
	quotes := SupplierQuoteDAOMock{}
	if got, err := quotes.CreateWithLines(ctx, quote, nil); err != nil || got != quote {
		t.Errorf("CreateWithLines() = (%v, %v), want original quote", got, err)
	}
	if got, err := quotes.UpdateTx(ctx, nil, quote); err != nil || got != quote {
		t.Errorf("UpdateTx() = (%v, %v), want original quote", got, err)
	}

	quoteLines := SupplierQuoteLineDAOMock{}
	if items, err := quoteLines.ListByQuote(ctx, 1); err != nil || len(items) != 0 {
		t.Errorf("ListByQuote() = (%v, %v), want empty", items, err)
	}
}

func TestEngineMockDefaults(t *testing.T) {
	ctx := context.Background()

	offers := OfferEngineMock{}
	if got, err := offers.BestOfferForSupplier(ctx, 1, 2, amount.Zero(), time.Now()); err != nil || got == nil || got.SupplierID != 2 {
		t.Errorf("BestOfferForSupplier() = (%v, %v), want offer for supplier 2", got, err)
	}

	suppliers := SupplierProfileLookupMock{}
	if got, err := suppliers.FindByContact(ctx, 2); err != nil || got == nil || !got.Active {
		t.Errorf("FindByContact() = (%v, %v), want active supplier", got, err)
	}

	receive := ReceiveEngineMock{}
	if got, err := receive.Receive(ctx, 1, amount.Zero(), 1, time.Now()); err != nil || got == nil {
		t.Errorf("Receive() = (%v, %v), want valuation layer", got, err)
	}

	approvals := ApprovalEngineMock{}
	if approved, err := approvals.IsApproved(ctx, "purchase_order", 1); err != nil || approved {
		t.Errorf("IsApproved() = (%v, %v), want not approved", approved, err)
	}
	if created, err := approvals.Create(ctx, 7, "purchase_order", 1, 2, nil); err != nil || created == nil {
		t.Errorf("Create() = (%v, %v), want approval request", created, err)
	}
	if _, steps, err := approvals.State(ctx, "purchase_order", 1); err != nil || len(steps) != 0 {
		t.Errorf("State() = (%v, %v), want empty steps", steps, err)
	}

	bills := BillEngineMock{}
	if created, err := bills.CreateSupplierBill(ctx, accounting.CreateSupplierBillRequest{}); err != nil || created == nil {
		t.Errorf("CreateSupplierBill() = (%v, %v), want invoice", created, err)
	}

	openInvoices := OpenInvoiceLookupMock{}
	if items, err := openInvoices.ListOpenByContact(ctx, 2); err != nil || len(items) != 0 {
		t.Errorf("ListOpenByContact() = (%v, %v), want empty", items, err)
	}

	payments := OutboundPaymentEngineMock{}
	if created, err := payments.CreateOutbound(ctx, accounting.CreatePaymentRequest{}); err != nil || created == nil {
		t.Errorf("CreateOutbound() = (%v, %v), want payment", created, err)
	}

	quality := QualityEngineMock{}
	if triggered, err := quality.TriggerChecks(ctx, 1, 1, nil); err != nil || triggered != 0 {
		t.Errorf("TriggerChecks() = (%v, %v), want 0", triggered, err)
	}
	if failed, err := quality.HasFailedChecks(ctx, 1); err != nil || failed {
		t.Errorf("HasFailedChecks() = (%v, %v), want false", failed, err)
	}

	expense := ExpenseAccountEngineMock{}
	if accountID, err := expense.ResolveExpenseAccount(ctx, 1); err != nil || accountID != 0 {
		t.Errorf("ResolveExpenseAccount() = (%v, %v), want 0", accountID, err)
	}

	resolver := inventory.ItemResolverMock{}
	if _, err := resolver.Resolve(ctx, 1); err != nil {
		t.Errorf("Resolve() error = %v, want nil", err)
	}
}

func TestOrderCreatorMockDefault(t *testing.T) {
	ctx := context.Background()
	order := &PurchaseOrder{}
	creator := OrderCreatorMock{}

	created, err := creator.Create(ctx, order, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created != order {
		t.Errorf("Create() = %v, want original order", created)
	}
}
