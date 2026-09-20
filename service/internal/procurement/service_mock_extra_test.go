package procurement

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/crosscutting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"gorm.io/gorm"
)

func TestProcurementMock_WithFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("update tx delegates", func(t *testing.T) {
		if _, err := (PurchaseRequestDAOMock{
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, r *PurchaseRequest) (*PurchaseRequest, error) {
				return r, nil
			},
		}).UpdateTx(ctx, nil, &PurchaseRequest{}); err != nil {
			t.Errorf("requisition UpdateTx = %v", err)
		}
		if _, err := (PurchaseOrderDAOMock{
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, o *PurchaseOrder) (*PurchaseOrder, error) { return o, nil },
		}).UpdateTx(ctx, nil, &PurchaseOrder{}); err != nil {
			t.Errorf("order UpdateTx = %v", err)
		}
		if _, err := (PurchaseOrderLineDAOMock{
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, l *PurchaseOrderLine) (*PurchaseOrderLine, error) { return l, nil },
		}).UpdateTx(ctx, nil, &PurchaseOrderLine{}); err != nil {
			t.Errorf("order line UpdateTx = %v", err)
		}
		if _, err := (SupplierQuoteRequestDAOMock{
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, r *SupplierQuoteRequest) (*SupplierQuoteRequest, error) {
				return r, nil
			},
		}).UpdateTx(ctx, nil, &SupplierQuoteRequest{}); err != nil {
			t.Errorf("quoteRequest UpdateTx = %v", err)
		}
		if _, err := (SupplierQuoteDAOMock{
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, q *SupplierQuote) (*SupplierQuote, error) { return q, nil },
		}).UpdateTx(ctx, nil, &SupplierQuote{}); err != nil {
			t.Errorf("quote UpdateTx = %v", err)
		}
	})

	t.Run("replace lines delegates", func(t *testing.T) {
		if err := (PurchaseRequestLineDAOMock{
			ReplaceLinesFunc: func(_ context.Context, _ uint64, _ []*PurchaseRequestLine) error { return nil },
		}).ReplaceLines(ctx, 1, nil); err != nil {
			t.Errorf("requisition ReplaceLines = %v", err)
		}
		if err := (PurchaseOrderLineDAOMock{
			ReplaceLinesFunc: func(_ context.Context, _ uint64, _ []*PurchaseOrderLine) error { return nil },
		}).ReplaceLines(ctx, 1, nil); err != nil {
			t.Errorf("order ReplaceLines = %v", err)
		}
	})

	t.Run("create with lines delegates", func(t *testing.T) {
		if _, err := (SupplyAgreementDAOMock{
			CreateWithLinesFunc: func(_ context.Context, a *SupplyAgreement, _ []*SupplyAgreementLine) (*SupplyAgreement, error) {
				return a, nil
			},
		}).CreateWithLines(ctx, &SupplyAgreement{}, nil); err != nil {
			t.Errorf("agreement CreateWithLines = %v", err)
		}
		if _, err := (PaymentBatchDAOMock{
			CreateWithLinesFunc: func(_ context.Context, b *PaymentBatch, _ []*PaymentBatchLine) (*PaymentBatch, error) { return b, nil },
		}).CreateWithLines(ctx, &PaymentBatch{}, nil); err != nil {
			t.Errorf("batch CreateWithLines = %v", err)
		}
	})

	t.Run("order creator and converter delegates", func(t *testing.T) {
		if _, err := (OrderCreatorMock{
			CreateFunc: func(_ context.Context, o *PurchaseOrder, _ []*PurchaseOrderLine) (*PurchaseOrder, error) {
				return o, nil
			},
		}).Create(ctx, &PurchaseOrder{}, nil); err != nil {
			t.Errorf("OrderCreator = %v", err)
		}
		if _, err := (CurrencyConverterMock{
			ConvertFunc: func(_ context.Context, a amount.Amount, _, _ string, _ *uint64, _ time.Time) (amount.Amount, error) {
				return a, nil
			},
		}).Convert(ctx, amount.FromFloat64(1), "USD", "IDR", nil, time.Now()); err != nil {
			t.Errorf("Convert = %v", err)
		}
		if _, err := (CurrencyConverterMock{
			HasRateFunc: func(_ context.Context, _, _ string, _ *uint64) (bool, error) { return true, nil },
		}).HasRate(ctx, "USD", "IDR", nil); err != nil {
			t.Errorf("HasRate = %v", err)
		}
	})

	t.Run("approval state delegates", func(t *testing.T) {
		if _, _, err := (ApprovalEngineMock{
			StateFunc: func(_ context.Context, _ string, _ uint64) (*crosscutting.ApprovalRequest, []*crosscutting.ApprovalStep, error) {
				return nil, nil, nil
			},
		}).State(ctx, "po", 1); err != nil {
			t.Errorf("State = %v", err)
		}
	})
}

func TestProcurementFixtures_WithOpts(t *testing.T) {
	order := PurchaseOrderFixture(func(o *PurchaseOrder) *PurchaseOrder {
		o.State = PurchaseOrderStateConfirmed
		return o
	})
	if order.State != PurchaseOrderStateConfirmed {
		t.Errorf("order state = %q", order.State)
	}
	line := PurchaseOrderLineFixture(func(l *PurchaseOrderLine) *PurchaseOrderLine {
		l.QtyOrdered = 9
		return l
	})
	if line.QtyOrdered != 9 {
		t.Errorf("qty = %v", line.QtyOrdered)
	}
	req := PurchaseRequestFixture(func(r *PurchaseRequest) *PurchaseRequest {
		r.State = RequestStateConfirmed
		return r
	})
	if req.State != RequestStateConfirmed {
		t.Errorf("requisition state = %q", req.State)
	}
	quoteRequest := SupplierQuoteRequestFixture(func(r *SupplierQuoteRequest) *SupplierQuoteRequest {
		r.State = QuoteRequestStateSent
		return r
	})
	if quoteRequest.State != QuoteRequestStateSent {
		t.Errorf("quoteRequest state = %q", quoteRequest.State)
	}
}
