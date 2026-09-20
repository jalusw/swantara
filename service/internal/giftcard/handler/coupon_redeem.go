package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

// @Summary Redeem coupon
// @Description Redeems a coupon code and returns the calculated discount amount.
// @Tags Coupons
// @Accept json
// @Produce json
// @Param organization_id path int true "Organization ID"
// @Param body body RedeemCouponRequest true "Redeem payload"
// @Success 200 {object} RedeemCouponResponseEnvelope "Coupon redeemed successfully."
// @Failure 404 {object} httpx.ErrorResponse "Coupon not found"
// @Failure 422 {object} httpx.ErrorResponse "Coupon expired or usage limit exceeded"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/coupons/redeem [post]
func (h CouponHandler) Redeem(c fiber.Ctx) error {
	var request RedeemCouponRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	date := time.Now().UTC()
	if request.Date != nil {
		parsed, err := helper.ParseDate(request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
		}
		date = *parsed
	}
	result, err := h.svc.Redeem(c, giftcard.RedeemCouponRequest{
		OrganizationID: *organizationID,
		Code:           request.Code,
		Subtotal:       amount.FromFloat64(request.Subtotal),
		Date:           date,
	})
	if err != nil {
		return writeCouponError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Coupon redeemed successfully.", RedeemCouponResponse{
		Coupon:   newCouponResponse(result.Coupon),
		Discount: result.Discount.Float64(),
	})
}
