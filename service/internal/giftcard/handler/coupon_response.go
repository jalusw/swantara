package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CouponResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	Code           string     `json:"code"`
	DiscountType   string     `json:"discount_type"`
	DiscountValue  float64    `json:"discount_value"`
	PriceRuleID    *uint64    `json:"price_rule_id"`
	UsageLimit     *int       `json:"usage_limit"`
	UsedCount      int        `json:"used_count"`
	ExpiryDate     *time.Time `json:"expiry_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newCouponResponse(coupon *giftcard.Coupon) CouponResponse {
	return CouponResponse{
		ID:             coupon.ID,
		OrganizationID: coupon.OrganizationID,
		Code:           coupon.Code,
		DiscountType:   coupon.DiscountType,
		DiscountValue:  coupon.DiscountValue,
		PriceRuleID:    coupon.PriceRuleID,
		UsageLimit:     coupon.UsageLimit,
		UsedCount:      coupon.UsedCount,
		ExpiryDate:     coupon.ExpiryDate,
		CreatedAt:      coupon.CreatedAt,
		UpdatedAt:      coupon.UpdatedAt,
	}
}

var couponQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"code":            {},
	"discount_type":   {},
	"price_rule_id":   {},
	"usage_limit":     {},
	"used_count":      {},
	"expiry_date":     {},
	"created_at":      {},
	"updated_at":      {},
}

type ListCouponsResponse struct {
	Coupons []CouponResponse `json:"coupons"`
}

type ListCouponsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCouponsResponse `json:"data"`
}

type GetCouponResponse struct {
	Coupon CouponResponse `json:"coupon"`
}

type GetCouponResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetCouponResponse `json:"data"`
}

type CreateCouponRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Code           string  `json:"code" validate:"required"`
	DiscountType   string  `json:"discount_type" validate:"required"`
	DiscountValue  float64 `json:"discount_value"`
	PriceRuleID    *uint64 `json:"price_rule_id"`
	UsageLimit     *int    `json:"usage_limit"`
	ExpiryDate     *string `json:"expiry_date"`
}

type CreateCouponResponse struct {
	Coupon CouponResponse `json:"coupon"`
}

type CreateCouponResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCouponResponse `json:"data"`
}

type UpdateCouponRequest struct {
	Code          string  `json:"code" validate:"required"`
	DiscountType  string  `json:"discount_type" validate:"required"`
	DiscountValue float64 `json:"discount_value"`
	PriceRuleID   *uint64 `json:"price_rule_id"`
	UsageLimit    *int    `json:"usage_limit"`
	ExpiryDate    *string `json:"expiry_date"`
}

type UpdateCouponResponse struct {
	Coupon CouponResponse `json:"coupon"`
}

type UpdateCouponResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateCouponResponse `json:"data"`
}

type RedeemCouponRequest struct {
	Code     string  `json:"code" validate:"required"`
	Subtotal float64 `json:"subtotal"`
	Date     *string `json:"date"`
}

type RedeemCouponResponse struct {
	Coupon   CouponResponse `json:"coupon"`
	Discount float64        `json:"discount"`
}

type RedeemCouponResponseEnvelope struct {
	httpx.EnvelopeBase
	Data RedeemCouponResponse `json:"data"`
}
