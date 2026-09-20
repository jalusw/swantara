package pos

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"gorm.io/gorm"
)

func TestPOSFixtures_WithOpts(t *testing.T) {
	if POSSessionFixture(func(s *POSSession) *POSSession { return s }) == nil {
		t.Error("session = nil")
	}
	if POSOrderFixture(func(o *POSOrder) *POSOrder { return o }) == nil {
		t.Error("order = nil")
	}
	if POSOrderLineFixture(func(l *POSOrderLine) *POSOrderLine { return l }) == nil {
		t.Error("line = nil")
	}
	if POSPaymentFixture(func(p *POSPayment) *POSPayment { return p }) == nil {
		t.Error("payment = nil")
	}
}

func TestPOSMock_WithFunc(t *testing.T) {
	ctx := context.Background()

	t.Run("transactioner", func(t *testing.T) {
		called := false
		mock := TransactionerMock{
			RunFunc: func(_ context.Context, fn func(tx *gorm.DB) error) error {
				called = true
				return fn(nil)
			},
		}
		if err := mock.Run(ctx, func(_ *gorm.DB) error { return nil }); err != nil {
			t.Errorf("Run = %v", err)
		}
		if !called {
			t.Error("RunFunc not called")
		}
	})

	t.Run("ship engine", func(t *testing.T) {
		mock := ShipEngineMock{
			ShipTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64, _ uint64, _ time.Time) (*inventory.CostLayer, error) {
				return &inventory.CostLayer{}, nil
			},
		}
		if _, err := mock.ShipTx(ctx, nil, 1, 2, time.Now()); err != nil {
			t.Errorf("ShipTx = %v", err)
		}
	})

	t.Run("credit note", func(t *testing.T) {
		mock := InvoiceEngineMock{
			CreateCreditNoteFunc: func(_ context.Context, _ accounting.CreateCreditNoteRequest) (*accounting.Invoice, error) {
				return &accounting.Invoice{}, nil
			},
		}
		if _, err := mock.CreateCreditNote(ctx, accounting.CreateCreditNoteRequest{}); err != nil {
			t.Errorf("CreateCreditNote = %v", err)
		}
	})

	t.Run("order tx", func(t *testing.T) {
		mock := POSOrderDAOMock{
			CreateWithLinesAndPaymentsTxFunc: func(_ context.Context, _ *gorm.DB, o *POSOrder, _ []*POSOrderLine, _ []*POSPayment) (*POSOrder, error) {
				return o, nil
			},
		}
		if _, err := mock.CreateWithLinesAndPaymentsTx(ctx, nil, &POSOrder{}, nil, nil); err != nil {
			t.Errorf("CreateWithLinesAndPaymentsTx = %v", err)
		}
	})
}
