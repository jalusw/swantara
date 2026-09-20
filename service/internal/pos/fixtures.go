package pos

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func POSSessionFixture(opts ...func(*POSSession) *POSSession) *POSSession {
	now := time.Now()
	s := &POSSession{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		ConfigID:       uint64(gofakeit.Number(1, 10000)),
		CashierID:      uint64(gofakeit.Number(1, 10000)),
		OpenedAt:       helper.Ptr(now),
		ClosedAt:       nil,
		OpeningBalance: gofakeit.Float64Range(1, 10000),
		ClosingBalance: nil,
		State:          "opened",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func POSOrderFixture(opts ...func(*POSOrder) *POSOrder) *POSOrder {
	now := time.Now()
	o := &POSOrder{
		Base:        model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		SessionID:   uint64(gofakeit.Number(1, 10000)),
		ContactID:   nil,
		Name:        helper.Ptr(gofakeit.AppName()),
		State:       "draft",
		InvoiceID:   nil,
		OrderTime:   helper.Ptr(now),
		AmountTotal: amount.FromFloat64(gofakeit.Float64Range(1, 10000)),
		AmountTax:   amount.FromFloat64(gofakeit.Float64Range(1, 10000)),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func POSOrderLineFixture(opts ...func(*POSOrderLine) *POSOrderLine) *POSOrderLine {
	now := time.Now()
	l := &POSOrderLine{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrderID:       uint64(gofakeit.Number(1, 10000)),
		ItemID:        nil,
		Qty:           gofakeit.Float64Range(1, 100),
		DiscountPct:   0,
		UnitPrice:     amount.FromFloat64(gofakeit.Float64Range(1, 10000)),
		PriceSubtotal: amount.FromFloat64(gofakeit.Float64Range(1, 10000)),
		PriceTax:      amount.FromFloat64(gofakeit.Float64Range(1, 10000)),
		PriceTotal:    amount.FromFloat64(gofakeit.Float64Range(1, 10000)),
		TaxIDs:        nil,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func POSPaymentFixture(opts ...func(*POSPayment) *POSPayment) *POSPayment {
	now := time.Now()
	p := &POSPayment{
		Base:    model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrderID: uint64(gofakeit.Number(1, 10000)),
		Method:  "cash",
		Amount:  gofakeit.Float64Range(1, 10000),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}
