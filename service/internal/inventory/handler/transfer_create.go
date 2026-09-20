package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

// @Summary Create warehouse transfer
// @Description Creates a draft warehouse transfer between two distinct warehouses, validating that both warehouses exist and differ and that the organization has a transit location. For each line, outbound stock movements from the source location to the transit location and inbound movements from transit to the destination location are generated in draft state, together with their outbound and inbound shipments. Returns the created warehouse transfer with a 201 status.
// @Tags Transfer Orders
// @Accept json
// @Produce json
// @Param body body CreateWarehouseTransferRequest true "Warehouse transfer details"
// @Success 201 {object} CreateWarehouseTransferResponseEnvelope "Warehouse transfer created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/transfer-orders [post]
func (h TransferHandler) Create(c fiber.Ctx) error {
	var request CreateWarehouseTransferRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	lines := make([]inventory.TransferLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = inventory.TransferLine{
			ItemID:        line.ItemID,
			Qty:           line.Qty,
			BatchID:       line.BatchID,
			SrcLocationID: line.SrcLocationID,
			DstLocationID: line.DstLocationID,
		}
	}

	scheduledDate, err := helper.ParseDate(request.ScheduledDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid scheduled date provided.", nil)
	}

	transfer, err := h.svc.Create(c, &inventory.WarehouseTransfer{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		Name:           request.Name,
		SrcWarehouseID: request.SrcWarehouseID,
		DstWarehouseID: request.DstWarehouseID,
		ScheduledDate:  scheduledDate,
	}, lines)
	if err != nil {
		return writeTransferError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Warehouse transfer created successfully.", CreateWarehouseTransferResponse{
		WarehouseTransfer: newWarehouseTransferResponse(transfer),
	})
}
