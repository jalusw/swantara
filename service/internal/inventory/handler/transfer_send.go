package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Send warehouse transfer
// @Description Sends a draft warehouse transfer, applying its outbound stock movements from source to transit and transitioning the order to in-transit. Each outbound movement is checked to ensure sufficient on-hand stock exists at its source location before it is applied, otherwise a 409 Conflict is returned. Goods-in-transit journal entries are posted at the item's current unit cost, debiting the transit account and crediting Inventory.
// @Tags Transfer Orders
// @Accept json
// @Produce json
// @Param id path integer true "Warehouse transfer ID"
// @Param body body TransferActionRequest true "Send details"
// @Success 200 {object} httpx.EmptyEnvelope "Warehouse transfer sent successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Warehouse transfer not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/transfer-orders/{id}/send [post]
func (h TransferHandler) Send(c fiber.Ctx) error {
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
		return httpx.CreateInternalServerErrorResponse(c, "Failed to send warehouse transfer.", err)
	}
	if transfer == nil || !httpx.OwnsTenant(c, transfer.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Warehouse transfer not found.")
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	if err := h.svc.Send(c, transferID, request.JournalID, request.TransitAccountID, date); err != nil {
		return writeTransferError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Warehouse transfer sent successfully.", struct{}{})
}
