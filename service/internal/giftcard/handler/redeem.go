package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type RedeemGiftCardRequest struct {
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	OrderType          string  `json:"order_type"`
	OrderID            uint64  `json:"order_id"`
	JournalID          uint64  `json:"journal_id" validate:"required,gt=0"`
	RevenueAccountID   uint64  `json:"revenue_account_id" validate:"required,gt=0"`
	LiabilityAccountID uint64  `json:"liability_account_id" validate:"required,gt=0"`
	Date               string  `json:"date"`
}

// @Summary Redeem gift card
// @Description Redeems a gift card balance against a sale, posting Dr Gift Card Liability / Cr Revenue and reducing the balance.
// @Tags Gift Cards
// @Accept json
// @Produce json
// @Param id path integer true "Gift card ID"
// @Param body body RedeemGiftCardRequest true "Redemption details"
// @Success 200 {object} GetGiftCardResponseEnvelope "Gift card redeemed successfully."
// @Failure 422 {object} httpx.ErrorResponse "Insufficient balance or validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/gift-cards/{id}/redeem [post]
func (h GiftCardHandler) Redeem(c fiber.Ctx) error {
	var request RedeemGiftCardRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := parseOptionalDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	updated, err := h.svc.Redeem(c, giftcard.RedeemRequest{
		GiftCardID:         id,
		Amount:             request.Amount,
		OrderType:          request.OrderType,
		OrderID:            request.OrderID,
		JournalID:          request.JournalID,
		RevenueAccountID:   request.RevenueAccountID,
		LiabilityAccountID: request.LiabilityAccountID,
		Date:               date,
	})
	if err != nil {
		return writeGiftCardError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Gift card redeemed successfully.", GetGiftCardResponse{
		GiftCard: newGiftCardResponse(updated),
	})
}
