package pos

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	SessionStateOpened  = "opened"
	SessionStateClosing = "closing"
	SessionStateClosed  = "closed"

	OrderStateDone     = "done"
	OrderStateRefunded = "refunded"

	SequencePOSOrderCode = "pos_order"
)

type POSSession struct {
	model.Base
	ConfigID       uint64     `gorm:"not null" json:"config_id"`
	CashierID      uint64     `gorm:"not null" json:"cashier_id"`
	OpenedAt       *time.Time `gorm:"type:timestamptz" json:"opened_at"`
	ClosedAt       *time.Time `gorm:"type:timestamptz" json:"closed_at"`
	OpeningBalance float64    `gorm:"type:numeric(18,4);default:0" json:"opening_balance"`
	ClosingBalance *float64   `gorm:"type:numeric(18,4)" json:"closing_balance"`
	State          string     `gorm:"type:text" json:"state"`
}

func (POSSession) TableName() string {
	return "pos_sessions"
}

type POSOrder struct {
	model.Base
	SessionID   uint64        `gorm:"not null" json:"session_id"`
	ContactID   *uint64       `json:"contact_id"`
	Name        *string       `json:"name"`
	AmountTotal amount.Amount `gorm:"type:numeric(18,4);default:0" json:"amount_total"`
	AmountTax   amount.Amount `gorm:"type:numeric(18,4);default:0" json:"amount_tax"`
	State       string        `gorm:"type:text" json:"state"`
	InvoiceID   *uint64       `json:"invoice_id"`
	OrderTime   *time.Time    `gorm:"type:timestamptz" json:"order_time"`
}

func (POSOrder) TableName() string {
	return "pos_orders"
}

type POSOrderLine struct {
	model.Base
	OrderID       uint64            `gorm:"not null" json:"order_id"`
	ItemID        *uint64           `json:"item_id"`
	Qty           float64           `gorm:"type:numeric(18,4);not null" json:"qty"`
	UnitPrice     amount.Amount     `gorm:"type:numeric(18,4)" json:"unit_price"`
	DiscountPct   float64           `gorm:"type:numeric(8,4);default:0" json:"discount_pct"`
	TaxIDs        helper.Int64Array `gorm:"type:bigint[]" json:"tax_ids"`
	PriceSubtotal amount.Amount     `gorm:"type:numeric(18,4)" json:"price_subtotal"`
	PriceTax      amount.Amount     `gorm:"type:numeric(18,4);default:0" json:"price_tax"`
	PriceTotal    amount.Amount     `gorm:"type:numeric(18,4);default:0" json:"price_total"`
}

func (POSOrderLine) TableName() string {
	return "pos_order_lines"
}

type POSPayment struct {
	model.Base
	OrderID uint64  `gorm:"not null" json:"order_id"`
	Method  string  `gorm:"type:text;not null" json:"method"`
	Amount  float64 `gorm:"type:numeric(18,4);not null" json:"amount"`
}

func (POSPayment) TableName() string {
	return "pos_payments"
}
