package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type RefundGiftCardRequest struct {
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	OrderType          string  `json:"order_type"`
	OrderID            uint64  `json:"order_id"`
	JournalID          uint64  `json:"journal_id" validate:"required,gt=0"`
	RefundAccountID    uint64  `json:"refund_account_id" validate:"required,gt=0"`
	LiabilityAccountID uint64  `json:"liability_account_id" validate:"required,gt=0"`
	Date               string  `json:"date"`
}

// @Summary Refund gift card
// @Description Restores a refunded amount to a gift card, posting Dr Sales Refund / Cr Gift Card Liability and increasing the balance.
// @Tags Gift Cards
// @Accept json
// @Produce json
// @Param id path integer true "Gift card ID"
// @Param body body RefundGiftCardRequest true "Refund details"
// @Success 200 {object} GetGiftCardResponseEnvelope "Gift card refunded successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/gift-cards/{id}/refund [post]
func (h GiftCardHandler) Refund(c fiber.Ctx) error {
	var request RefundGiftCardRequest
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
	updated, err := h.svc.Refund(c, giftcard.RefundRequest{
		GiftCardID:         id,
		Amount:             request.Amount,
		OrderType:          request.OrderType,
		OrderID:            request.OrderID,
		JournalID:          request.JournalID,
		RefundAccountID:    request.RefundAccountID,
		LiabilityAccountID: request.LiabilityAccountID,
		Date:               date,
	})
	if err != nil {
		return writeGiftCardError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Gift card refunded successfully.", GetGiftCardResponse{
		GiftCard: newGiftCardResponse(updated),
	})
}
