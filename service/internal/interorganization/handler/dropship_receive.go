package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

type ReceiveDropshipOrderRequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date"`
}

type GetDropshipOrderResponse struct {
	PurchaseOrder DropshipOrderResponse `json:"purchase_order"`
}

type GetDropshipOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetDropshipOrderResponse `json:"data"`
}

// @Summary Receive drop-ship order
// @Description Records the supplier-to-customer receipt: movements are applied without entering on-hand stock, COGS and stock-input are posted per line, and the linked sale lines are marked delivered.
// @Tags Drop-shipping
// @Accept json
// @Produce json
// @Param id path integer true "Drop-ship purchase order ID"
// @Param body body ReceiveDropshipOrderRequest true "Receipt details"
// @Success 200 {object} GetDropshipOrderResponseEnvelope "Drop-ship order received successfully."
// @Failure 404 {object} httpx.ErrorResponse "Drop-ship order not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/dropship-orders/{id}/receive [post]
func (h DropShipHandler) Receive(c fiber.Ctx) error {
	var request ReceiveDropshipOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	date := time.Now()
	if request.Date != "" {
		parsed, err := helper.ParseDateStr(request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
		}
		date = parsed
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	order, err := h.svc.Receive(c, interorganization.ReceiveDropshipRequest{
		OrganizationID:  *organizationID,
		PurchaseOrderID: id,
		JournalID:       request.JournalID,
		Date:            date,
	})
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Drop-ship order received successfully.", GetDropshipOrderResponse{
		PurchaseOrder: newDropshipOrderResponse(order),
	})
}
