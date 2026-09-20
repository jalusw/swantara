package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateJournalRequest struct {
	OrganizationID   *uint64 `json:"organization_id"`
	Code             *string `json:"code"`
	Name             string  `json:"name" validate:"required"`
	Type             string  `json:"type" validate:"required"`
	DefaultAccountID *uint64 `json:"default_account_id"`
	BankAccountID    *uint64 `json:"bank_account_id"`
}

type CreateJournalResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateJournalResponse `json:"data"`
}
type CreateJournalResponse struct {
	Journal JournalResponse `json:"journal"`
}

// @Summary Create journal
// @Description Creates an accounting journal with a required name and type, plus an optional code, default account, and bank account. The type must be valid and the default account, if given, must belong to the same organization; the journal is assigned to the caller's organization.
// @Tags Journals
// @Accept json
// @Produce json
// @Param body body CreateJournalRequest true "Journal details"
// @Success 201 {object} CreateJournalResponseEnvelope "Journal created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/journals [post]
func (h JournalHandler) Create(c fiber.Ctx) error {
	var request CreateJournalRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	journal, err := h.svc.Create(c, &reference.Journal{
		OrganizationID:   *organizationID,
		Code:             request.Code,
		Name:             request.Name,
		Type:             request.Type,
		DefaultAccountID: request.DefaultAccountID,
		BankAccountID:    request.BankAccountID,
	})
	if err != nil {
		return writeJournalError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Journal created successfully.", CreateJournalResponse{
		Journal: newJournalResponse(journal),
	})
}
