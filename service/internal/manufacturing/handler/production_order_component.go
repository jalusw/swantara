package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List production order components
// @Description Lists the planned and consumed component lines of a tenant-scoped production order, including planned and consumed quantities and the associated stock movement for each component. Results can be exported in CSV format when requested.
// @Tags Manufacturing Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListConsumedMaterialsResponseEnvelope "Components retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/components [get]
func (h ProductionOrderHandler) ListComponents(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	productionOrder, err := h.svc.Find(c, productionOrderID)
	if err != nil {
		httpx.RequestLog(c).Error("production order lookup failed", "production_order_id", productionOrderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve components.", err)
	}
	if productionOrder == nil || !httpx.OwnsTenant(c, productionOrder.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Production order not found.")
	}

	components, err := h.svc.ListComponents(c, productionOrderID)
	if err != nil {
		httpx.RequestLog(c).Error("production order components list failed", "production_order_id", productionOrderID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve components.", err)
	}

	items := make([]ConsumedMaterialResponse, len(components))
	for i, component := range components {
		items[i] = newConsumedMaterialResponse(component)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "manufacturing-order-components.csv", items)
	}

	return httpx.CreateSuccessResponse(c, "Components retrieved successfully.", ListConsumedMaterialsResponse{
		Components: items,
	})
}
