package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/giftcard"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type IssueGiftCardRequest struct {
	OrganizationID     *uint64 `json:"organization_id"`
	Code               string  `json:"code"`
	ContactID          *uint64 `json:"contact_id"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	CurrencyCode       string  `json:"currency_code" validate:"required,len=3"`
	ExpiryDate         *string `json:"expiry_date"`
	IssuedFromOrderID  *uint64 `json:"issued_from_order_id"`
	JournalID          uint64  `json:"journal_id" validate:"required,gt=0"`
	CashAccountID      uint64  `json:"cash_account_id" validate:"required,gt=0"`
	LiabilityAccountID uint64  `json:"liability_account_id" validate:"required,gt=0"`
	Date               string  `json:"date"`
}

// @Summary Issue gift card
// @Description Issues a gift card with a unique code, posting Dr Cash / Cr Gift Card Liability. When no code is provided one is generated.
// @Tags Gift Cards
// @Accept json
// @Produce json
// @Param body body IssueGiftCardRequest true "Gift card issue details"
// @Success 201 {object} CreateGiftCardResponseEnvelope "Gift card issued successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/gift-cards [post]
func (h GiftCardHandler) Issue(c fiber.Ctx) error {
	var request IssueGiftCardRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	var expiry *time.Time
	if request.ExpiryDate != nil {
		parsed, err := helper.ParseDateStr(*request.ExpiryDate)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expiry_date.", nil)
		}
		expiry = &parsed
	}
	date, err := parseOptionalDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
	}
	created, err := h.svc.Issue(c, giftcard.IssueRequest{
		OrganizationID:     *organizationID,
		Code:               request.Code,
		ContactID:          request.ContactID,
		Amount:             request.Amount,
		CurrencyCode:       request.CurrencyCode,
		ExpiryDate:         expiry,
		IssuedFromOrderID:  request.IssuedFromOrderID,
		JournalID:          request.JournalID,
		CashAccountID:      request.CashAccountID,
		LiabilityAccountID: request.LiabilityAccountID,
		Date:               date,
	})
	if err != nil {
		return writeGiftCardError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Gift card issued successfully.", CreateGiftCardResponse{
		GiftCard: newGiftCardResponse(created),
	})
}
