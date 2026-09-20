package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Subscription KPI
// @Description Returns MRR, ARR, churn, and LTV for active subscriptions.
// @Tags KPIs
// @Accept json
// @Produce json
// @Success 200 {object} reporting.SubscriptionKPI "Subscription KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/subscription [get]
func (h KpiHandler) Subscription(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	kpi, err := h.svc.SubscriptionKPI(c, *organizationID)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Subscription KPI retrieved successfully.", kpi)
}

// @Summary Projects KPI
// @Description Returns aggregated project margin, cost, billed amounts, and utilization.
// @Tags KPIs
// @Accept json
// @Produce json
// @Success 200 {object} reporting.ProjectKPI "Projects KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/projects [get]
func (h KpiHandler) Projects(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	kpi, err := h.svc.ProjectKPI(c, *organizationID)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Projects KPI retrieved successfully.", kpi)
}

// @Summary Payroll KPI
// @Description Returns gross and net payroll cost for a date range from paid payroll runs.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} reporting.PayrollKPI "Payroll KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/payroll [get]
func (h KpiHandler) Payroll(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	start, end, ok := parseDateRange(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}
	kpi, err := h.svc.PayrollKPI(c, *organizationID, start, end)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Payroll KPI retrieved successfully.", kpi)
}
