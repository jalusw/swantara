package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Confirm production order
// @Description Confirms a production order, transitioning it from draft to confirmed. Only draft orders can be confirmed; any other state is rejected with a conflict.
// @Tags Manufacturing Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Success 200 {object} GetProductionOrderResponseEnvelope "Production order confirmed successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/confirm [post]
func (h ProductionOrderHandler) Confirm(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	productionOrder, err := h.svc.Confirm(c, productionOrderID)
	if err != nil {
		return writeProductionOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Production order confirmed successfully.", GetProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
	})
}

// @Summary Plan production order
// @Description Plans a confirmed production order by reserving the planned component quantities from the source location through stock reservations, then transitioning the order to the planned state. The operation fails with a conflict when there is insufficient stock to reserve the components.
// @Tags Manufacturing Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Success 200 {object} GetProductionOrderResponseEnvelope "Production order planned successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed or insufficient stock"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/plan [post]
func (h ProductionOrderHandler) Plan(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	productionOrder, err := h.svc.Plan(c, productionOrderID)
	if err != nil {
		return writeProductionOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Production order planned successfully.", GetProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
	})
}

// @Summary Cancel production order
// @Description Cancels a production order, moving it to the cancelled state. The transition is rejected when the order is already done or already cancelled, so only orders that have not been completed can be cancelled.
// @Tags Manufacturing Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Success 200 {object} GetProductionOrderResponseEnvelope "Production order cancelled successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/cancel [post]
func (h ProductionOrderHandler) Cancel(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	productionOrder, err := h.svc.Cancel(c, productionOrderID)
	if err != nil {
		return writeProductionOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Production order cancelled successfully.", GetProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
	})
}
