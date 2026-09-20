package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/pos"
)

type POSOrderLineResponse struct {
	ID            uint64            `json:"id"`
	OrderID       uint64            `json:"order_id"`
	ItemID        *uint64           `json:"item_id"`
	Qty           float64           `json:"qty"`
	UnitPrice     amount.Amount     `json:"unit_price"`
	DiscountPct   float64           `json:"discount_pct"`
	TaxIDs        helper.Int64Array `json:"tax_ids"`
	PriceSubtotal amount.Amount     `json:"price_subtotal"`
	PriceTax      amount.Amount     `json:"price_tax"`
	PriceTotal    amount.Amount     `json:"price_total"`
}

func newPOSOrderLineResponse(line *pos.POSOrderLine) POSOrderLineResponse {
	return POSOrderLineResponse{
		ID:            line.ID,
		OrderID:       line.OrderID,
		ItemID:        line.ItemID,
		Qty:           line.Qty,
		UnitPrice:     line.UnitPrice,
		DiscountPct:   line.DiscountPct,
		TaxIDs:        line.TaxIDs,
		PriceSubtotal: line.PriceSubtotal,
		PriceTax:      line.PriceTax,
		PriceTotal:    line.PriceTotal,
	}
}

type POSPaymentResponse struct {
	ID      uint64  `json:"id"`
	OrderID uint64  `json:"order_id"`
	Method  string  `json:"method"`
	Amount  float64 `json:"amount"`
}

func newPOSPaymentResponse(payment *pos.POSPayment) POSPaymentResponse {
	return POSPaymentResponse{
		ID:      payment.ID,
		OrderID: payment.OrderID,
		Method:  payment.Method,
		Amount:  payment.Amount,
	}
}

type POSOrderResponse struct {
	ID          uint64                 `json:"id"`
	SessionID   uint64                 `json:"session_id"`
	ContactID   *uint64                `json:"contact_id"`
	Name        *string                `json:"name"`
	AmountTotal amount.Amount          `json:"amount_total"`
	AmountTax   amount.Amount          `json:"amount_tax"`
	State       string                 `json:"state"`
	InvoiceID   *uint64                `json:"invoice_id"`
	OrderTime   *time.Time             `json:"order_time"`
	Lines       []POSOrderLineResponse `json:"lines,omitempty"`
	Payments    []POSPaymentResponse   `json:"payments,omitempty"`
}

func newPOSOrderResponse(order *pos.POSOrder) POSOrderResponse {
	return POSOrderResponse{
		ID:          order.ID,
		SessionID:   order.SessionID,
		ContactID:   order.ContactID,
		Name:        order.Name,
		AmountTotal: order.AmountTotal,
		AmountTax:   order.AmountTax,
		State:       order.State,
		InvoiceID:   order.InvoiceID,
		OrderTime:   order.OrderTime,
	}
}
