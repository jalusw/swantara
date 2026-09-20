package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List gift cards
// @Description Lists gift cards with pagination, sorting, and filtering. Results are scoped to the caller's organization.
// @Tags Gift Cards
// @Accept json
// @Produce json
// @Success 200 {object} ListGiftCardsResponseEnvelope "Gift cards retrieved successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/gift-cards [get]
func (h GiftCardHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, giftCardQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("gift card list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve gift cards.", err)
	}
	items := make([]GiftCardResponse, len(page.Items))
	for i, card := range page.Items {
		items[i] = newGiftCardResponse(card)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Gift cards retrieved successfully.", ListGiftCardsResponse{
		GiftCards: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
