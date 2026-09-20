package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Inventory KPI
// @Description Returns on-hand inventory value, quantity, and item count.
// @Tags KPIs
// @Accept json
// @Produce json
// @Success 200 {object} reporting.InventoryKPI "Inventory KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/inventory [get]
func (h KpiHandler) Inventory(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	kpi, err := h.svc.InventoryKPI(c, *organizationID)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Inventory KPI retrieved successfully.", kpi)
}

// @Summary Procurement KPI
// @Description Returns PO cycle time, on-time delivery, and price variance for a date range.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} reporting.ProcurementKPI "Procurement KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/procurement [get]
func (h KpiHandler) Procurement(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	start, end, ok := parseDateRange(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}
	kpi, err := h.svc.ProcurementKPI(c, *organizationID, start, end)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Procurement KPI retrieved successfully.", kpi)
}

// @Summary Manufacturing KPI
// @Description Returns OEE, yield, scrap, and cost variance from shop tasks and production orders for a date range.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} reporting.ManufacturingKPI "Manufacturing KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/manufacturing [get]
func (h KpiHandler) Manufacturing(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	start, end, ok := parseDateRange(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}
	kpi, err := h.svc.ManufacturingKPI(c, *organizationID, start, end)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Manufacturing KPI retrieved successfully.", kpi)
}

// @Summary Inventory ratio KPI
// @Description Returns inventory turnover, days on hand, and stockout count for a date range.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} reporting.InventoryRatioKPI "Inventory ratio KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/inventory-ratio [get]
func (h KpiHandler) InventoryRatio(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	start, end, ok := parseDateRange(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}
	kpi, err := h.svc.InventoryRatioKPI(c, *organizationID, start, end)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Inventory ratio KPI retrieved successfully.", kpi)
}
