package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type InvoicePOSOrderRequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date" validate:"required"`
}

type InvoicePOSOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data InvoicePOSOrderResponse `json:"data"`
}
type InvoicePOSOrderResponse struct {
	InvoiceID uint64 `json:"invoice_id"`
}

// @Summary Generate invoice for POS order
// @Description Generates a formal customer invoice for a POS order that has a contact, and reconciles the revenue already recognized at sale.
// @Tags POS Orders
// @Accept json
// @Produce json
// @Param id path integer true "POS order ID"
// @Param body body InvoicePOSOrderRequest true "Invoice details"
// @Success 201 {object} InvoicePOSOrderResponseEnvelope "Invoice generated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "POS order not found"
// @Failure 409 {object} httpx.ErrorResponse "Order already invoiced"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/orders/{id}/invoice [post]
func (h POSOrderHandler) Invoice(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid POS order id provided.", nil)
	}
	var request InvoicePOSOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := parsePOSDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	invoice, err := h.svc.CreateInvoice(c, orderID, request.JournalID, date)
	if err != nil {
		return writePOSError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Invoice generated successfully.", InvoicePOSOrderResponse{
		InvoiceID: invoice.ID,
	})
}

func parsePOSDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}
