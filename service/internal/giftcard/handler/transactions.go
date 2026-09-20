package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List gift card transactions
// @Description Lists the issue/redeem/refund transactions of a gift card. The card must belong to the caller's organization.
// @Tags Gift Cards
// @Accept json
// @Produce json
// @Param id path integer true "Gift card ID"
// @Success 200 {object} ListGiftCardTransactionsResponseEnvelope "Gift card transactions retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Gift card not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/gift-cards/{id}/transactions [get]
func (h GiftCardHandler) ListTransactions(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	card, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("gift card get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve gift card.", err)
	}
	if card == nil || !httpx.OwnsTenant(c, card.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Gift card not found.")
	}
	transactions, err := h.svc.ListTransactions(c, card.ID)
	if err != nil {
		httpx.RequestLog(c).Error("gift card transactions list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve gift card transactions.", err)
	}
	items := make([]GiftCardTransactionResponse, len(transactions))
	for i, t := range transactions {
		items[i] = newGiftCardTransactionResponse(t)
	}
	return httpx.CreateSuccessResponse(c, "Gift card transactions retrieved successfully.", ListGiftCardTransactionsResponse{
		GiftCardTransactions: items,
	})
}
