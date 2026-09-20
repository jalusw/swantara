package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Complete outside processing order
// @Description Completes a received outside processing order, moving it from the received to the done state.
// @Tags Subcontract Orders
// @Accept json
// @Produce json
// @Param id path integer true "Outside processing order ID"
// @Success 200 {object} OutsideProcessingOrderEnvelope "Outside processing order completed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Outside processing order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subcontract-orders/{id}/done [post]
func (h OutsideProcessingHandler) Done(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid outside processing order id provided.", nil)
	}

	order, err := h.svc.Done(c, orderID)
	if err != nil {
		return writeSubcontractError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Outside processing order completed successfully.", newOutsideProcessingOrderResponse(order))
}

// @Summary Cancel outside processing order
// @Description Cancels a outside processing order that has not been completed, moving it to the cancelled state. Done or already cancelled orders cannot be cancelled.
// @Tags Subcontract Orders
// @Accept json
// @Produce json
// @Param id path integer true "Outside processing order ID"
// @Success 200 {object} OutsideProcessingOrderEnvelope "Outside processing order cancelled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Outside processing order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subcontract-orders/{id}/cancel [post]
func (h OutsideProcessingHandler) Cancel(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid outside processing order id provided.", nil)
	}

	order, err := h.svc.Cancel(c, orderID)
	if err != nil {
		return writeSubcontractError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Outside processing order cancelled successfully.", newOutsideProcessingOrderResponse(order))
}
