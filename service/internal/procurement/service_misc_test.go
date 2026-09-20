package procurement

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
)

func TestProcurementFixtures_Construct(t *testing.T) {
	if PurchaseOrderFixture() == nil {
		t.Error("order fixture = nil")
	}
	if PurchaseOrderLineFixture() == nil {
		t.Error("order line fixture = nil")
	}
	if PurchaseRequestFixture() == nil {
		t.Error("requisition fixture = nil")
	}
	if SupplierQuoteRequestFixture() == nil {
		t.Error("quoteRequest fixture = nil")
	}
}

func TestProcurementModels_TableNames(t *testing.T) {
	tables := map[string]string{
		(CostCenter{}).TableName():          "cost_centers",
		(PurchaseCreditMemo{}).TableName():  "purchase_credit_memos",
		(PurchaseDebitMemo{}).TableName():   "purchase_debit_memos",
		(PaymentBatch{}).TableName():        "payment_batches",
		(PaymentBatchLine{}).TableName():    "payment_batch_lines",
		(CurrencyRate{}).TableName():        "currency_rates",
		(SupplyAgreement{}).TableName():     "supply_agreements",
		(SupplyAgreementLine{}).TableName(): "supply_agreement_lines",
		(SupplierScorecard{}).TableName():   "supplier_scorecards",
	}
	for got, want := range tables {
		if got != want {
			t.Errorf("table = %q, want %q", got, want)
		}
	}
}

func TestProcurementMock_Fallbacks(t *testing.T) {
	ctx := context.Background()

	t.Run("dao mock fallbacks", func(t *testing.T) {
		if _, err := (PurchaseOrderDAOMock{}).UpdateTx(ctx, nil, &PurchaseOrder{}); err != nil {
			t.Errorf("order UpdateTx = %v", err)
		}
		if _, err := (PurchaseOrderLineDAOMock{}).UpdateTx(ctx, nil, &PurchaseOrderLine{}); err != nil {
			t.Errorf("order line UpdateTx = %v", err)
		}
		if _, err := (PurchaseRequestDAOMock{}).UpdateTx(ctx, nil, &PurchaseRequest{}); err != nil {
			t.Errorf("requisition UpdateTx = %v", err)
		}
		if err := (PurchaseRequestLineDAOMock{}).ReplaceLines(ctx, 1, nil); err != nil {
			t.Errorf("requisition ReplaceLines = %v", err)
		}
		if _, err := (SupplierQuoteRequestDAOMock{}).UpdateTx(ctx, nil, &SupplierQuoteRequest{}); err != nil {
			t.Errorf("quoteRequest UpdateTx = %v", err)
		}
		if _, err := (SupplierQuoteDAOMock{}).UpdateTx(ctx, nil, &SupplierQuote{}); err != nil {
			t.Errorf("quote UpdateTx = %v", err)
		}
		if _, err := (CurrencyRateDAOMock{}).FindByPair(ctx, "USD", "IDR", nil); err != nil {
			t.Errorf("FindByPair = %v", err)
		}
		if _, err := (CurrencyRateDAOMock{}).FindByPairDate(ctx, "USD", "IDR", nil, time.Now()); err != nil {
			t.Errorf("FindByPairDate = %v", err)
		}
		agreement := &SupplyAgreement{}
		if _, err := (SupplyAgreementDAOMock{}).CreateWithLines(ctx, agreement, nil); err != nil {
			t.Errorf("agreement CreateWithLines = %v", err)
		}
		if agreement.ID == 0 {
			t.Error("expected fallback to assign id")
		}
		if _, err := (SupplyAgreementLineDAOMock{}).ListByAgreement(ctx, 1); err != nil {
			t.Errorf("ListByAgreement = %v", err)
		}
		if _, err := (SupplierScorecardDAOMock{}).FindByVendorAndPeriod(ctx, 1, nil, nil); err != nil {
			t.Errorf("FindByVendorAndPeriod = %v", err)
		}
		batch := &PaymentBatch{}
		if _, err := (PaymentBatchDAOMock{}).CreateWithLines(ctx, batch, nil); err != nil {
			t.Errorf("batch CreateWithLines = %v", err)
		}
		if batch.ID == 0 {
			t.Error("expected fallback to assign id")
		}
		if _, err := (PaymentBatchLineDAOMock{}).ListByBatch(ctx, 1); err != nil {
			t.Errorf("ListByBatch = %v", err)
		}
	})

	t.Run("service mock fallbacks", func(t *testing.T) {
		if _, err := (OrderCreatorMock{}).Create(ctx, &PurchaseOrder{}, nil); err != nil {
			t.Errorf("OrderCreator = %v", err)
		}
		if _, err := (CurrencyConverterMock{}).Convert(ctx, amount.FromFloat64(1), "USD", "IDR", nil, time.Now()); err != nil {
			t.Errorf("Convert = %v", err)
		}
		if _, err := (CurrencyConverterMock{}).HasRate(ctx, "USD", "IDR", nil); err != nil {
			t.Errorf("HasRate = %v", err)
		}
	})

	t.Run("engine mock fallbacks", func(t *testing.T) {
		if _, err := (BillEngineMock{}).CreateCreditNote(ctx, accounting.CreateCreditNoteRequest{}); err != nil {
			t.Errorf("CreateCreditNote = %v", err)
		}
		if _, err := (BillEngineMock{}).CreateSupplierBill(ctx, accounting.CreateSupplierBillRequest{}); err != nil {
			t.Errorf("CreateSupplierBill = %v", err)
		}
		if _, _, err := (ApprovalEngineMock{}).State(ctx, "po", 1); err != nil {
			t.Errorf("Approval State = %v", err)
		}
	})

	t.Run("test constructors", func(t *testing.T) {
		_ = NewTestPurchaseOrderService(PurchaseOrderServiceTestDeps{})
		_ = NewTestPurchaseRequestService(PurchaseRequestServiceTestDeps{})
		_ = NewTestSupplierQuoteRequestService(SupplierQuoteRequestServiceTestDeps{})
		_ = NewTestApprovalConfig(1, 1000, []uint64{5})
	})

	t.Run("test constructors with presets", func(t *testing.T) {
		seqPreset := sequence.DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return &sequence.Reservation{Value: 2, Number: "X/2"}, nil
			},
		}
		contactPreset := contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: 9}}, nil
				},
			},
		}
		_ = NewTestPurchaseOrderService(PurchaseOrderServiceTestDeps{
			Sequences:  seqPreset,
			Contacts:   contactPreset,
			Warehouses: inventory.WarehouseDAOMock{},
			Expense: ExpenseAccountEngineMock{
				ResolveExpenseAccountFunc: func(_ context.Context, _ uint64) (uint64, error) { return 1, nil },
			},
		})
		_ = NewTestPurchaseRequestService(PurchaseRequestServiceTestDeps{
			Sequences: seqPreset,
			Contacts:  contactPreset,
		})
		_ = NewTestSupplierQuoteRequestService(SupplierQuoteRequestServiceTestDeps{
			Sequences: seqPreset,
			Contacts:  contactPreset,
			Orders:    OrderCreatorMock{},
		})
		_ = NewTestApprovalConfig(1, 0, nil)
	})
}
