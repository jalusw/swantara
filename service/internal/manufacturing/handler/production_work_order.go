package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GenerateShopTasksResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GenerateShopTasksResponse `json:"data"`
}
type GenerateShopTasksResponse struct {
	ShopTasks []ShopTaskResponse `json:"shop_tasks"`
}

// @Summary Generate shop tasks
// @Description Generates shop tasks from the production order's routing operations, sequencing them by operation and assigning each one to its configured work center. Each shop task is created in the planned state with its planned duration derived from the operation setup and time minutes, and requires a valid work center. The generated shop tasks are returned in operation order.
// @Tags Work Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Success 200 {object} GenerateShopTasksResponseEnvelope "Shop tasks generated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/work-orders [post]
func (h ProductionHandler) GenerateShopTasks(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	workOrders, err := h.svc.GenerateShopTasks(c, productionOrderID)
	if err != nil {
		return writeProductionOrderError(c, err)
	}

	items := make([]ShopTaskResponse, len(workOrders))
	for i, workOrder := range workOrders {
		items[i] = newShopTaskResponse(workOrder)
	}

	return httpx.CreateSuccessResponse(c, "Shop tasks generated successfully.", GenerateShopTasksResponse{
		ShopTasks: items,
	})
}

type ListShopTasksResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListShopTasksResponse `json:"data"`
}
type ListShopTasksResponse struct {
	ShopTasks []ShopTaskResponse `json:"shop_tasks"`
}

// @Summary List shop tasks
// @Description Lists the shop tasks for a production order, including planned and actual durations, planned and finished dates, and state. Results can be exported in CSV format when requested.
// @Tags Work Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListShopTasksResponseEnvelope "Shop tasks retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/work-orders [get]
func (h ProductionHandler) ListShopTasks(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	workOrders, err := h.svc.ListShopTasksByMO(c, productionOrderID)
	if err != nil {
		httpx.RequestLog(c).Error("shop task list failed", "production_order_id", productionOrderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve shop tasks.", err)
	}

	items := make([]ShopTaskResponse, len(workOrders))
	for i, workOrder := range workOrders {
		items[i] = newShopTaskResponse(workOrder)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "work-orders.csv", items)
	}

	return httpx.CreateSuccessResponse(c, "Shop tasks retrieved successfully.", ListShopTasksResponse{
		ShopTasks: items,
	})
}
