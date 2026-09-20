package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete account
// @Description Deletes a chart of accounts entry by id. The entry must belong to the caller's organization; a 404 is returned otherwise.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param id path integer true "Account ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Account not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accounts/{id} [delete]
func (h AccountHandler) Delete(c fiber.Ctx) error {
	accountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid account id provided.", nil)
	}

	account, err := h.svc.Find(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete account.", err)
	}
	if account == nil || !httpx.OwnsTenant(c, &account.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Account not found.")
	}

	if err := h.svc.Delete(c, accountID); err != nil {
		httpx.RequestLog(c).Error("account deletion failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete account.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
