package sales

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestSaleOrderDAOMock_Defaults(t *testing.T) {
	ctx := context.Background()
	order := &SaleOrder{Base: model.Base{ID: 1}}
	orders := SaleOrderDAOMock{}

	if got, err := orders.CreateWithLines(ctx, order, nil); err != nil || got != order {
		t.Errorf("CreateWithLines = (%v, %v), want original order", got, err)
	}
	if got, err := orders.UpdateTx(ctx, &gorm.DB{}, order); err != nil || got != order {
		t.Errorf("UpdateTx = (%v, %v), want original order", got, err)
	}

	invoked := false
	orders.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, o *SaleOrder) (*SaleOrder, error) {
		invoked = true
		return o, nil
	}
	if got, err := orders.UpdateTx(ctx, &gorm.DB{}, order); err != nil || got != order || !invoked {
		t.Errorf("UpdateTx = (%v, %v, invoked %v), want func invoked", got, err, invoked)
	}
}

func TestSaleOrderLineDAOMock_Defaults(t *testing.T) {
	ctx := context.Background()
	line := &SaleOrderLine{Base: model.Base{ID: 1}}
	lines := SaleOrderLineDAOMock{}

	if got, err := lines.UpdateTx(ctx, &gorm.DB{}, line); err != nil || got != line {
		t.Errorf("UpdateTx = (%v, %v), want original line", got, err)
	}
	if err := lines.ReplaceLines(ctx, 1, nil); err != nil {
		t.Errorf("ReplaceLines error = %v, want nil", err)
	}

	invoked := false
	lines.UpdateTxFunc = func(_ context.Context, _ *gorm.DB, l *SaleOrderLine) (*SaleOrderLine, error) {
		invoked = true
		return l, nil
	}
	if got, err := lines.UpdateTx(ctx, &gorm.DB{}, line); err != nil || got != line || !invoked {
		t.Errorf("UpdateTx = (%v, %v, invoked %v), want func invoked", got, err, invoked)
	}
}

func TestSequenceDAOMock_Defaults(t *testing.T) {
	ctx := context.Background()
	seq := SequenceDAOMock{}
	now := time.Now()

	res, err := seq.Reserve(ctx, 1, "sale_order", now)
	if err != nil || res.Number != "SO/00001" {
		t.Errorf("Reserve = (%v, %v), want reservation SO/00001", res, err)
	}
	res, err = seq.ReserveTx(ctx, &gorm.DB{}, 1, "sale_order", now)
	if err != nil || res.Number != "SO/00001" {
		t.Errorf("ReserveTx = (%v, %v), want reservation SO/00001", res, err)
	}
}

func TestShipEngineMock_Default(t *testing.T) {
	ctx := context.Background()

	layer, err := ShipEngineMock{}.Ship(ctx, 1, 2, time.Now())
	if err != nil || layer == nil {
		t.Errorf("Ship = (%v, %v), want a valuation layer", layer, err)
	}
}

func TestInvoiceEngineMock_Default(t *testing.T) {
	ctx := context.Background()

	invoice, err := InvoiceEngineMock{}.Create(ctx, accounting.CreateInvoiceRequest{ContactID: 5})
	if err != nil || invoice.ContactID != 5 {
		t.Errorf("Create = (%v, %v), want invoice for contact 5", invoice, err)
	}
}

func TestInvoiceLookupMock_Default(t *testing.T) {
	ctx := context.Background()

	invoices, err := InvoiceLookupMock{}.ListOpenByContact(ctx, 5)
	if err != nil || len(invoices) != 1 {
		t.Errorf("ListOpenByContact = (%v, %v), want one open invoice", invoices, err)
	}
}

func TestPaymentEngineMock_Default(t *testing.T) {
	ctx := context.Background()

	payment, err := PaymentEngineMock{}.Create(ctx, accounting.CreatePaymentRequest{ContactID: 5, Amount: 100})
	if err != nil || payment.ContactID != 5 || payment.Amount != 100 {
		t.Errorf("Create = (%v, %v), want payment for contact 5", payment, err)
	}
}
