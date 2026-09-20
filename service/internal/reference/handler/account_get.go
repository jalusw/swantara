package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetAccountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetAccountResponse `json:"data"`
}
type GetAccountResponse struct {
	Account AccountResponse `json:"account"`
}

// @Summary Get account
// @Description Returns a single chart of accounts entry by id. The entry must belong to the caller's organization, otherwise a 404 is returned.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path integer true "Account ID"
// @Success 200 {object} GetAccountResponseEnvelope "Account retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Account not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accounts/{id} [get]
func (h AccountHandler) Get(c fiber.Ctx) error {
	accountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid account id provided.", nil)
	}

	account, err := h.svc.Find(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get account.", err)
	}
	if account == nil || !httpx.OwnsTenant(c, &account.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Account not found.")
	}

	return httpx.CreateSuccessResponse(c, "Account retrieved successfully.", GetAccountResponse{
		Account: newAccountResponse(account),
	})
}
