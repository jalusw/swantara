package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type PostStockCountRequest struct {
	JournalID         uint64 `json:"journal_id" validate:"required,gt=0"`
	GainLossAccountID uint64 `json:"gain_loss_account_id" validate:"required,gt=0"`
	Date              string `json:"date"`
}

type PostStockCountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PostStockCountResponse `json:"data"`
}
type PostStockCountResponse struct {
	Count StockCountResponse `json:"count"`
}

// @Summary Post inventory count
// @Description Posts a draft inventory count, applying each line's difference to its stock quant and transitioning the count to the posted state. The count must be in draft and have at least one line, and a journal plus an inventory gain/loss account are required. For every non-zero difference, a journal entry is posted at the item's current unit cost: a gain debits Inventory and credits the gain/loss account, while a loss debits the gain/loss account and credits Inventory.
// @Tags Inventory Counts
// @Accept json
// @Produce json
// @Param id path integer true "Inventory count ID"
// @Param body body PostStockCountRequest true "Posting details"
// @Success 200 {object} PostStockCountResponseEnvelope "Inventory count posted successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Inventory count not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/inventory-counts/{id}/post [post]
func (h StockCountHandler) Post(c fiber.Ctx) error {
	stockCountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inventory count id provided.", nil)
	}

	var request PostStockCountRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	count, err := h.svc.Find(c, stockCountID)
	if err != nil {
		httpx.RequestLog(c).Error("inventory count lookup failed", "stock_count_id", stockCountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to post inventory count.", err)
	}
	if count == nil || !httpx.OwnsTenant(c, count.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Inventory count not found.")
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	updated, err := h.svc.Post(c, stockCountID, request.JournalID, request.GainLossAccountID, date)
	if err != nil {
		return writeCountError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Inventory count posted successfully.", PostStockCountResponse{
		Count: newStockCountResponse(updated),
	})
}
