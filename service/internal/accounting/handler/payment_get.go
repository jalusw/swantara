package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Get payment
// @Description Gets a single payment by its id together with its invoice allocations, returning 404 when the payment does not exist or belongs to another tenant. The allocations show how the payment amount was distributed across invoices.
// @Tags Payments
// @Accept json
// @Produce json
// @Param id path integer true "Payment ID"
// @Success 200 {object} GetPaymentResponseEnvelope "Payment retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Payment not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payments/{id} [get]
func (h PaymentHandler) Get(c fiber.Ctx) error {
	paymentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid payment id provided.", nil)
	}

	payment, err := h.svc.Find(c, paymentID)
	if err != nil {
		httpx.RequestLog(c).Error("payment lookup failed", "payment_id", paymentID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payment.", err)
	}
	if payment == nil || !httpx.OwnsTenant(c, payment.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Payment not found.")
	}

	allocations, err := h.svc.ListAllocations(c, payment.ID)
	if err != nil {
		httpx.RequestLog(c).Error("payment allocations lookup failed", "payment_id", payment.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get payment allocations.", err)
	}
	allocationItems := make([]PaymentAllocationResponse, len(allocations))
	for i, allocation := range allocations {
		allocationItems[i] = newPaymentAllocationResponse(allocation)
	}

	return httpx.CreateSuccessResponse(c, "Payment retrieved successfully.", GetPaymentResponse{
		Payment:     newPaymentResponse(payment),
		Allocations: allocationItems,
	})
}
