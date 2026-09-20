package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/pos"
)

type POSOrderLineRequest struct {
	ItemID      uint64            `json:"item_id" validate:"required,gt=0"`
	Qty         float64           `json:"qty" validate:"required,gt=0"`
	DiscountPct float64           `json:"discount_pct" validate:"gte=0,lte=100"`
	TaxIDs      helper.Int64Array `json:"tax_ids"`
}

type POSPaymentRequest struct {
	Method string  `json:"method" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gt=0"`
}

type SellPOSOrderRequest struct {
	SessionID uint64                `json:"session_id" validate:"required,gt=0"`
	ContactID *uint64               `json:"contact_id"`
	Lines     []POSOrderLineRequest `json:"lines" validate:"required,min=1,dive"`
	Payments  []POSPaymentRequest   `json:"payments" validate:"required,min=1,dive"`
}

type SellPOSOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SellPOSOrderResponse `json:"data"`
}
type SellPOSOrderResponse struct {
	Order POSOrderResponse `json:"order"`
}

// @Summary Sell at the register
// @Description Creates a POS order with lines and payments, posts the outgoing stock movement with COGS, and recognizes revenue immediately. For orders with a contact, a formal invoice is generated and the receivable is reconciled.
// @Tags POS Orders
// @Accept json
// @Produce json
// @Param body body SellPOSOrderRequest true "POS order details"
// @Success 201 {object} SellPOSOrderResponseEnvelope "POS order created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Session not found"
// @Failure 409 {object} httpx.ErrorResponse "Stock unavailable or session not open"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown reference"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/orders [post]
func (h POSOrderHandler) Sell(c fiber.Ctx) error {
	var request SellPOSOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	lines := make([]pos.SellLineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = pos.SellLineRequest{
			ItemID:      line.ItemID,
			Qty:         line.Qty,
			DiscountPct: line.DiscountPct,
			TaxIDs:      line.TaxIDs,
		}
	}
	payments := make([]pos.SellPaymentRequest, len(request.Payments))
	for i, payment := range request.Payments {
		payments[i] = pos.SellPaymentRequest{
			Method: payment.Method,
			Amount: payment.Amount,
		}
	}

	order, err := h.svc.Sell(c, pos.SellRequest{
		SessionID: request.SessionID,
		ContactID: request.ContactID,
		Lines:     lines,
		Payments:  payments,
	})
	if err != nil {
		return writePOSError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "POS order created successfully.", SellPOSOrderResponse{
		Order: newPOSOrderResponse(order),
	})
}
