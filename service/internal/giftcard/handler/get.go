package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Get gift card
// @Description Gets a single gift card by id with its current balance. The card must belong to the caller's organization; a request for a card owned by another tenant returns 404.
// @Tags Gift Cards
// @Accept json
// @Produce json
// @Param id path integer true "Gift card ID"
// @Success 200 {object} GetGiftCardResponseEnvelope "Gift card retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Gift card not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/gift-cards/{id} [get]
func (h GiftCardHandler) Get(c fiber.Ctx) error {
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
	return httpx.CreateSuccessResponse(c, "Gift card retrieved successfully.", GetGiftCardResponse{
		GiftCard: newGiftCardResponse(card),
	})
}
