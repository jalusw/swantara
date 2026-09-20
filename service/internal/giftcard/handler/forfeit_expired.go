package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ForfeitExpiredGiftCardsRequest struct {
	JournalID          uint64 `json:"journal_id" validate:"required,gt=0"`
	LiabilityAccountID uint64 `json:"liability_account_id" validate:"required,gt=0"`
	IncomeAccountID    uint64 `json:"income_account_id" validate:"required,gt=0"`
	AsOf               string `json:"as_of"`
	Date               string `json:"date"`
}

// @Summary Forfeit expired gift cards
// @Description Forfeits the residual balance of gift cards past their expiry date, posting Dr Gift Card Liability / Cr Other Income. The run is scoped to the caller's organization and is idempotent.
// @Tags Gift Cards
// @Accept json
// @Produce json
// @Param body body ForfeitExpiredGiftCardsRequest true "Forfeit run details"
// @Success 200 {object} ForfeitExpiredGiftCardsResponseEnvelope "Expired gift cards forfeited successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/gift-cards/forfeit-expired [post]
func (h GiftCardHandler) ForfeitExpired(c fiber.Ctx) error {
	var request ForfeitExpiredGiftCardsRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	asOf, err := parseOptionalDate(request.AsOf)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid as_of.", nil)
	}
	date, err := parseOptionalDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	forfeited, err := h.svc.ForfeitExpired(c, giftcard.ForfeitExpiredRequest{
		OrganizationID:     *organizationID,
		AsOf:               asOf,
		Date:               date,
		JournalID:          request.JournalID,
		LiabilityAccountID: request.LiabilityAccountID,
		IncomeAccountID:    request.IncomeAccountID,
	})
	if err != nil {
		return writeGiftCardError(c, err)
	}
	items := make([]GiftCardResponse, len(forfeited))
	for i, card := range forfeited {
		items[i] = newGiftCardResponse(card)
	}
	return httpx.CreateSuccessResponse(c, "Expired gift cards forfeited successfully.", ForfeitExpiredGiftCardsResponse{
		GiftCards: items,
	})
}
