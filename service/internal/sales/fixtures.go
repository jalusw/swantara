package sales

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func SaleOrderFixture(opts ...func(*SaleOrder) *SaleOrder) *SaleOrder {
	now := time.Now()
	o := &SaleOrder{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:           helper.Ptr(gofakeit.AppName()),
		ContactID:      uint64(gofakeit.Number(1, 10000)),
		ShipAddressID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		BillAddressID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		PriceBookID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		CurrencyCode:   helper.Ptr("USD"),
		SalespersonID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		SalesGroupID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		ProspectID:     helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		WarehouseID:    helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		State:          OrderStateDraft,
		OrderDate:      helper.Ptr(now),
		ExpectedDate:   helper.Ptr(now.AddDate(0, 0, 14)),
		ValidityDate:   helper.Ptr(now.AddDate(0, 1, 0)),
		PaymentTermID:  helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Incoterm:       helper.Ptr("FOB"),
		CustomerPORef:  helper.Ptr(gofakeit.AppName()),
		AmountUntaxed:  gofakeit.Float64Range(1, 10000),
		AmountTax:      gofakeit.Float64Range(1, 10000),
		AmountTotal:    gofakeit.Float64Range(1, 10000),
		InvoiceStatus:  InvoiceStatusNo,
		DeliveryStatus: DeliveryStatusPending,
		Note:           helper.Ptr(gofakeit.Sentence(3)),
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func SaleOrderLineFixture(opts ...func(*SaleOrderLine) *SaleOrderLine) *SaleOrderLine {
	now := time.Now()
	l := &SaleOrderLine{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrderID:       uint64(gofakeit.Number(1, 10000)),
		Sequence:      10,
		ItemID:        helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Description:   helper.Ptr(gofakeit.Sentence(3)),
		QtyOrdered:    gofakeit.Float64Range(1, 100),
		QtyDelivered:  0,
		QtyInvoiced:   0,
		QtyReturns:    0,
		UnitID:        helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		UnitPrice:     gofakeit.Float64Range(1, 10000),
		DiscountPct:   0,
		TaxIDs:        nil,
		DimensionID:   helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		PriceSubtotal: gofakeit.Float64Range(1, 10000),
		PriceTax:      gofakeit.Float64Range(1, 10000),
		PriceTotal:    gofakeit.Float64Range(1, 10000),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}
