package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateAccountRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Code           string  `json:"code" validate:"required"`
	Name           string  `json:"name" validate:"required"`
	Type           string  `json:"type" validate:"required"`
	Reconcilable   bool    `json:"reconcilable"`
	CurrencyCode   *string `json:"currency_code"`
	ParentID       *uint64 `json:"parent_id"`
	Active         *bool   `json:"active"`
}

type CreateAccountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateAccountResponse `json:"data"`
}
type CreateAccountResponse struct {
	Account AccountResponse `json:"account"`
}

// @Summary Create account
// @Description Creates a chart of accounts entry with a required code, name, and account type, plus an optional parent and currency. The account code must be unique within the organization, the type must be valid, and the parent, if given, must belong to the same organization; new accounts default to active.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param body body CreateAccountRequest true "Account details"
// @Success 201 {object} CreateAccountResponseEnvelope "Account created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or duplicate code"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accounts [post]
func (h AccountHandler) Create(c fiber.Ctx) error {
	var request CreateAccountRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	account, err := h.svc.Create(c, &reference.Account{
		OrganizationID: *organizationID,
		Code:           request.Code,
		Name:           request.Name,
		Type:           request.Type,
		Reconcilable:   request.Reconcilable,
		CurrencyCode:   request.CurrencyCode,
		ParentID:       request.ParentID,
		Active:         boolOrTrue(request.Active),
	})
	if err != nil {
		return writeAccountError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Account created successfully.", CreateAccountResponse{
		Account: newAccountResponse(account),
	})
}
