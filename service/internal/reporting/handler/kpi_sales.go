package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Sales KPI
// @Description Returns bookings, revenue, COGS, gross margin, and win rate for a date range.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} reporting.SalesKPI "Sales KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/sales [get]
func (h KpiHandler) Sales(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	start, end, ok := parseDateRange(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}
	kpi, err := h.svc.SalesKPI(c, *organizationID, start, end)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Sales KPI retrieved successfully.", kpi)
}

// @Summary Pipeline KPI
// @Description Returns weighted pipeline, expected revenue, and win rate from open opportunities.
// @Tags KPIs
// @Accept json
// @Produce json
// @Success 200 {object} reporting.PipelineKPI "Pipeline KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/pipeline [get]
func (h KpiHandler) Pipeline(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	kpi, err := h.svc.PipelineKPI(c, *organizationID)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Pipeline KPI retrieved successfully.", kpi)
}
