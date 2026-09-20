package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetPOSOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPOSOrderResponse `json:"data"`
}
type GetPOSOrderResponse struct {
	Order POSOrderResponse `json:"order"`
}

// @Summary Get POS order
// @Description Returns a single POS order with its lines and payments, forming the receipt data. The order must belong to the caller's organization, otherwise a 404 is returned.
// @Tags POS Orders
// @Accept json
// @Produce json
// @Param id path integer true "POS order ID"
// @Success 200 {object} GetPOSOrderResponseEnvelope "Order retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "POS order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/orders/{id} [get]
func (h POSOrderHandler) Get(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid POS order id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	order, err := h.svc.GetOrder(c, organizationID, orderID)
	if err != nil {
		return writePOSError(c, err)
	}
	lines, err := h.svc.ListOrderLines(c, order.ID)
	if err != nil {
		httpx.RequestLog(c).Error("pos order lines failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve POS order lines.", err)
	}
	payments, err := h.svc.ListOrderPayments(c, order.ID)
	if err != nil {
		httpx.RequestLog(c).Error("pos order payments failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve POS order payments.", err)
	}
	response := newPOSOrderResponse(order)
	response.Lines = make([]POSOrderLineResponse, len(lines))
	for i, line := range lines {
		response.Lines[i] = newPOSOrderLineResponse(line)
	}
	response.Payments = make([]POSPaymentResponse, len(payments))
	for i, payment := range payments {
		response.Payments[i] = newPOSPaymentResponse(payment)
	}
	return httpx.CreateSuccessResponse(c, "POS order retrieved successfully.", GetPOSOrderResponse{
		Order: response,
	})
}
