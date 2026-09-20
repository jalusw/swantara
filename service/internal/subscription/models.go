package subscription

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	SubscriptionStateDraft   = "draft"
	SubscriptionStateActive  = "active"
	SubscriptionStatePaused  = "paused"
	SubscriptionStateChurned = "churned"
	SubscriptionStateClosed  = "closed"
)

type Subscription struct {
	model.Base
	OrganizationID  *uint64    `json:"organization_id"`
	Name            string     `json:"name"`
	ContactID       *uint64    `json:"contact_id"`
	PlanID          *uint64    `json:"plan_id"`
	PriceBookID     *uint64    `json:"price_book_id"`
	CurrencyCode    *string    `gorm:"type:char(3)" json:"currency_code"`
	DateStart       *time.Time `gorm:"type:date" json:"date_start"`
	NextInvoiceDate *time.Time `gorm:"type:date" json:"next_invoice_date"`
	DateEnd         *time.Time `gorm:"type:date" json:"date_end"`
	State           string     `gorm:"type:text" json:"state"`
	MRR             float64    `gorm:"type:numeric(18,4)" json:"mrr"`
}

func (Subscription) TableName() string {
	return "subscriptions"
}

type SubscriptionLine struct {
	model.Base
	SubscriptionID uint64  `json:"subscription_id"`
	ItemID         *uint64 `json:"item_id"`
	Qty            float64 `gorm:"type:numeric(18,4)" json:"qty"`
	UnitPrice      float64 `gorm:"type:numeric(18,4)" json:"unit_price"`
	DiscountPct    float64 `gorm:"type:numeric(8,4)" json:"discount_pct"`
}

func (SubscriptionLine) TableName() string {
	return "subscription_lines"
}
