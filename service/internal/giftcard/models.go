package giftcard

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	TransactionIssue   = "issue"
	TransactionRedeem  = "redeem"
	TransactionRefund  = "refund"
	TransactionAdjust  = "adjust"
	TransactionForfeit = "forfeit"

	GiftCardStateActive    = "active"
	GiftCardStateUsed      = "used"
	GiftCardStateExpired   = "expired"
	GiftCardStateCancelled = "cancelled"
)

type GiftCard struct {
	model.Base
	OrganizationID    *uint64    `json:"organization_id"`
	Code              string     `json:"code" audit:"redact"`
	ContactID         *uint64    `json:"contact_id"`
	InitialAmount     float64    `gorm:"type:numeric(18,4)" json:"initial_amount"`
	Balance           float64    `gorm:"type:numeric(18,4)" json:"balance"`
	CurrencyCode      string     `gorm:"type:char(3)" json:"currency_code"`
	ExpiryDate        *time.Time `gorm:"type:date" json:"expiry_date"`
	State             string     `gorm:"type:text" json:"state"`
	IssuedFromOrderID *uint64    `json:"issued_from_order_id"`
}

func (GiftCard) TableName() string {
	return "gift_cards"
}

type GiftCardTransaction struct {
	model.Base
	GiftCardID uint64  `json:"gift_card_id"`
	Type       string  `gorm:"type:text" json:"type"`
	Amount     float64 `gorm:"type:numeric(18,4)" json:"amount"`
	OrderType  string  `json:"order_type"`
	OrderID    uint64  `json:"order_id"`
	EntryID    *uint64 `json:"entry_id"`
}

func (GiftCardTransaction) TableName() string {
	return "gift_card_transactions"
}

const (
	CouponDiscountPercent = "percent"
	CouponDiscountFixed   = "fixed"
)

var validCouponDiscountType = map[string]struct{}{
	CouponDiscountPercent: {},
	CouponDiscountFixed:   {},
}

type Coupon struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Code           string     `json:"code" audit:"redact"`
	DiscountType   string     `gorm:"type:text" json:"discount_type"`
	DiscountValue  float64    `gorm:"type:numeric(18,4)" json:"discount_value"`
	PriceRuleID    *uint64    `json:"price_rule_id"`
	UsageLimit     *int       `json:"usage_limit"`
	UsedCount      int        `gorm:"default:0" json:"used_count"`
	ExpiryDate     *time.Time `gorm:"type:date" json:"expiry_date"`
}

func (Coupon) TableName() string {
	return "coupons"
}
