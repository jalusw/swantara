package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func couponExpiry(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	parsed, err := helper.ParseDateStr(*raw)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func writeCouponError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, giftcard.ErrCouponNotFound):
		return httpx.CreateNotFoundResponse(c, "Coupon not found.")
	case errors.Is(err, giftcard.ErrCouponCodeRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Coupon code is required.", nil)
	case errors.Is(err, giftcard.ErrCouponOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Coupon requires an organization.", nil)
	case errors.Is(err, giftcard.ErrCouponDiscountInvalid):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Coupon discount is invalid.", nil)
	case errors.Is(err, giftcard.ErrCouponAlreadyExists):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Coupon code already exists for the organization.", nil)
	case errors.Is(err, giftcard.ErrCouponExpired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Coupon has expired.", nil)
	case errors.Is(err, giftcard.ErrCouponOverLimit):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Coupon usage limit exceeded.", nil)
	default:
		httpx.RequestLog(c).Error("coupon write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save coupon.", err)
	}
}
