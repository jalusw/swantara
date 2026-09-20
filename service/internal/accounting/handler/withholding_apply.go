package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type WithholdRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	JournalID      uint64  `json:"journal_id" validate:"required,gt=0"`
	ContactID      uint64  `json:"contact_id" validate:"required,gt=0"`
	Amount         float64 `json:"amount" validate:"required,gt=0"`
	Date           *string `json:"date"`
	Ref            string  `json:"ref"`
	Scope          string  `json:"scope" validate:"required,oneof=sale purchase"`
	WHTID          uint64  `json:"withholding_tax_id" validate:"required,gt=0"`
	BankAccountID  uint64  `json:"bank_account_id" validate:"required,gt=0"`
	PayableID      uint64  `json:"payable_account_id" validate:"required,gt=0"`
	Description    string  `json:"description"`
}

type WithholdResponseEnvelope struct {
	httpx.EnvelopeBase
	Data WithholdResponse `json:"data"`
}
type WithholdResponse struct {
	EntryID uint64 `json:"entry_id"`
}

// @Summary Apply withholding tax
// @Description Posts a withholding payment journal entry that computes the withheld amount from the withholding tax rate and derives the net amount settled with the contact, requiring a non-empty date. The entry debits payables for the gross amount and credits the withholding tax and bank accounts, and the withholding tax must exist with an account and a scope matching the request.
// @Tags Withholding Taxes
// @Accept json
// @Produce json
// @Param body body WithholdRequest true "Withholding details"
// @Success 201 {object} WithholdResponseEnvelope "Withholding applied successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/withholding-taxes/apply [post]
func (h WithholdingTaxHandler) Apply(c fiber.Ctx) error {
	var request WithholdRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	if date == nil || date.IsZero() {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date is required.", nil)
	}

	entry, err := h.svc.Withhold(c, accounting.WithholdRequest{
		OrganizationID: *organizationID,
		JournalID:      request.JournalID,
		ContactID:      request.ContactID,
		Amount:         request.Amount,
		Date:           *date,
		Ref:            request.Ref,
		Scope:          request.Scope,
		WHTID:          request.WHTID,
		BankAccountID:  request.BankAccountID,
		PayableID:      request.PayableID,
		Description:    request.Description,
	})
	if err != nil {
		return writeWithholdingTaxError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Withholding applied successfully.", WithholdResponse{
		EntryID: entry.ID,
	})
}
