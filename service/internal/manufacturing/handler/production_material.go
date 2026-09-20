package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ConsumeMaterialRequest struct {
	ComponentID  uint64  `json:"component_id" validate:"required,gt=0"`
	Qty          float64 `json:"qty" validate:"required,gt=0"`
	JournalID    uint64  `json:"journal_id" validate:"required,gt=0"`
	WIPAccountID uint64  `json:"wip_account_id" validate:"required,gt=0"`
	Date         string  `json:"date"`
}

type GetConsumedMaterialResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetConsumedMaterialResponse `json:"data"`
}
type GetConsumedMaterialResponse struct {
	Component ConsumedMaterialResponse `json:"component"`
}

// @Summary Consume raw materials
// @Description Consumes raw materials against a production order component for an in-progress order, creating a stock movement from the order's source location to the production location and posting a Dr WIP / Cr Inventory journal entry. The quantity must not exceed the remaining planned quantity of the component, which must belong to the order. The component's consumed quantity and stock movement are updated and returned.
// @Tags Work Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Param body body ConsumeMaterialRequest true "Consumption details"
// @Success 200 {object} GetConsumedMaterialResponseEnvelope "Raw materials consumed successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed or insufficient stock"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown references"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/consume [post]
func (h ProductionHandler) Consume(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	var request ConsumeMaterialRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	component, err := h.svc.Consume(c, productionOrderID, request.ComponentID, request.Qty, request.JournalID, request.WIPAccountID, date)
	if err != nil {
		return writeProductionError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Raw materials consumed successfully.", GetConsumedMaterialResponse{
		Component: newConsumedMaterialResponse(component),
	})
}

type ProduceGoodsRequest struct {
	Qty          float64 `json:"qty" validate:"required,gt=0"`
	JournalID    uint64  `json:"journal_id" validate:"required,gt=0"`
	WIPAccountID uint64  `json:"wip_account_id" validate:"required,gt=0"`
	Date         string  `json:"date"`
}

// @Summary Produce finished goods
// @Description Produces finished goods for an in-progress production order, moving the output quantity from the production location into the destination stock location and posting a Dr Inventory / Cr WIP journal entry at the item's standard cost. The quantity must not exceed the remaining quantity to produce. The order's produced quantity is updated and the order is returned.
// @Tags Work Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Param body body ProduceGoodsRequest true "Production details"
// @Success 200 {object} GetProductionOrderResponseEnvelope "Finished goods produced successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed or insufficient WIP"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown references"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/produce [post]
func (h ProductionHandler) Produce(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	var request ProduceGoodsRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	productionOrder, err := h.svc.Produce(c, productionOrderID, request.Qty, request.JournalID, request.WIPAccountID, date)
	if err != nil {
		return writeProductionError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Finished goods produced successfully.", GetProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
	})
}
