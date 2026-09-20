package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Receive warehouse transfer
// @Description Completes an in-transit warehouse transfer by applying its inbound stock movements from transit to their destination locations and transitioning the order to the received state. Journal entries are posted at the item's current unit cost, debiting Inventory and crediting the transit account, which closes the goods-in-transit balance opened when the order was sent.
// @Tags Transfer Orders
// @Accept json
// @Produce json
// @Param id path integer true "Warehouse transfer ID"
// @Param body body TransferActionRequest true "Receive details"
// @Success 200 {object} httpx.EmptyEnvelope "Warehouse transfer received successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Warehouse transfer not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/transfer-orders/{id}/receive [post]
func (h TransferHandler) Receive(c fiber.Ctx) error {
	transferID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid warehouse transfer id provided.", nil)
	}

	var request TransferActionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	transfer, err := h.svc.Find(c, transferID)
	if err != nil {
		httpx.RequestLog(c).Error("warehouse transfer lookup failed", "transfer_id", transferID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to receive warehouse transfer.", err)
	}
	if transfer == nil || !httpx.OwnsTenant(c, transfer.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Warehouse transfer not found.")
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	if err := h.svc.Receive(c, transferID, request.JournalID, request.TransitAccountID, date); err != nil {
		return writeTransferError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Warehouse transfer received successfully.", struct{}{})
}
