package procurement

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func PurchaseOrderFixture(opts ...func(*PurchaseOrder) *PurchaseOrder) *PurchaseOrder {
	now := time.Now()
	po := &PurchaseOrder{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: nil,
		Name:           nil,
		SupplierID:     uint64(gofakeit.Number(1, 10000)),
		SupplierRef:    nil,
		CurrencyCode:   nil,
		WarehouseID:    nil,
		DestLocationID: nil,
		State:          "draft",
		OrderDate:      helper.Ptr(now),
		ExpectedDate:   nil,
		PaymentTermID:  nil,
		Incoterm:       nil,
		AmountUntaxed:  gofakeit.Float64Range(1, 10000),
		AmountTax:      gofakeit.Float64Range(1, 10000),
		AmountTotal:    gofakeit.Float64Range(1, 10000),
		InvoiceStatus:  "no",
		ReceiptStatus:  "no",
	}
	for _, opt := range opts {
		opt(po)
	}
	return po
}

func PurchaseOrderLineFixture(opts ...func(*PurchaseOrderLine) *PurchaseOrderLine) *PurchaseOrderLine {
	now := time.Now()
	l := &PurchaseOrderLine{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrderID:       uint64(gofakeit.Number(1, 10000)),
		Sequence:      gofakeit.Number(1, 100),
		ItemID:        nil,
		Description:   nil,
		QtyOrdered:    gofakeit.Float64Range(1, 1000),
		QtyReceived:   0,
		QtyBilled:     0,
		QtyReturns:    0,
		UnitID:        nil,
		UnitPrice:     gofakeit.Float64Range(1, 10000),
		DiscountPct:   0,
		DimensionID:   nil,
		CostCenterID:  nil,
		PriceSubtotal: gofakeit.Float64Range(1, 10000),
		TaxIDs:        nil,
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

func PurchaseRequestFixture(opts ...func(*PurchaseRequest) *PurchaseRequest) *PurchaseRequest {
	now := time.Now()
	pr := &PurchaseRequest{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: nil,
		Name:           nil,
		RequesterID:    uint64(gofakeit.Number(1, 10000)),
		DepartmentID:   nil,
		State:          "draft",
		NeededBy:       nil,
	}
	for _, opt := range opts {
		opt(pr)
	}
	return pr
}

func SupplierQuoteRequestFixture(opts ...func(*SupplierQuoteRequest) *SupplierQuoteRequest) *SupplierQuoteRequest {
	now := time.Now()
	quoteRequest := &SupplierQuoteRequest{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: nil,
		Name:           nil,
		RequesterID:    uint64(gofakeit.Number(1, 10000)),
		SupplierID:     nil,
		CurrencyCode:   nil,
		State:          "draft",
		OrderDate:      helper.Ptr(now),
		QuoteDeadline:  nil,
		Notes:          nil,
	}
	for _, opt := range opts {
		opt(quoteRequest)
	}
	return quoteRequest
}
