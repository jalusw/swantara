package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type RefundPOSOrderRequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date" validate:"required"`
}

type RefundPOSOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data RefundPOSOrderResponse `json:"data"`
}
type RefundPOSOrderResponse struct {
	Order POSOrderResponse `json:"order"`
}

// @Summary Refund POS order
// @Description Fully refunds a done POS order: restores stock at the original cost, reverses the revenue journal movement, and generates a credit note when the order was invoiced.
// @Tags POS Orders
// @Accept json
// @Produce json
// @Param id path integer true "POS order ID"
// @Param body body RefundPOSOrderRequest true "Refund details"
// @Success 200 {object} RefundPOSOrderResponseEnvelope "Order refunded successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "POS order not found"
// @Failure 409 {object} httpx.ErrorResponse "Order cannot be refunded in its current state"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/orders/{id}/refund [post]
func (h POSOrderHandler) Refund(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid POS order id provided.", nil)
	}
	var request RefundPOSOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := parsePOSDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	order, err := h.svc.Refund(c, orderID, request.JournalID, date)
	if err != nil {
		return writePOSError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "POS order refunded successfully.", RefundPOSOrderResponse{
		Order: newPOSOrderResponse(order),
	})
}
