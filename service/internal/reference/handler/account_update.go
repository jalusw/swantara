package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateAccountRequest struct {
	Code         string  `json:"code" validate:"required"`
	Name         string  `json:"name" validate:"required"`
	Type         string  `json:"type" validate:"required"`
	Reconcilable bool    `json:"reconcilable"`
	CurrencyCode *string `json:"currency_code"`
	ParentID     *uint64 `json:"parent_id"`
	Active       *bool   `json:"active"`
}

type UpdateAccountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateAccountResponse `json:"data"`
}
type UpdateAccountResponse struct {
	Account AccountResponse `json:"account"`
}

// @Summary Update account
// @Description Updates a chart of accounts entry's code, name, type, currency, parent, and flags. The code must remain unique within the organization, the type must be valid, and any new parent must belong to the same organization; a 404 is returned if the entry does not belong to the caller.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path integer true "Account ID"
// @Param body body UpdateAccountRequest true "Account details"
// @Success 200 {object} UpdateAccountResponseEnvelope "Account updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Account not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or duplicate code"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accounts/{id} [put]
func (h AccountHandler) Update(c fiber.Ctx) error {
	accountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid account id provided.", nil)
	}

	var request UpdateAccountRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	account, err := h.svc.Find(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update account.", err)
	}
	if account == nil || !httpx.OwnsTenant(c, &account.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Account not found.")
	}

	account.Code = request.Code
	account.Name = request.Name
	account.Type = request.Type
	account.Reconcilable = request.Reconcilable
	account.CurrencyCode = request.CurrencyCode
	account.ParentID = request.ParentID
	account.Active = boolOrTrue(request.Active)

	updated, err := h.svc.Update(c, account)
	if err != nil {
		return writeAccountError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Account updated successfully.", UpdateAccountResponse{
		Account: newAccountResponse(updated),
	})
}
