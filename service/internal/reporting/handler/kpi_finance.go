package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Finance KPI
// @Description Returns revenue, expenses, gross/net margin, EBITDA, and current ratio from the general ledger for a date range.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param start query string false "Start date (YYYY-MM-DD)"
// @Param end query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} reporting.FinanceKPI "Finance KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/finance [get]
func (h KpiHandler) Finance(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	start, end, ok := parseDateRange(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Dates must be in YYYY-MM-DD format.", nil)
	}
	kpi, err := h.svc.FinanceKPI(c, *organizationID, start, end)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Finance KPI retrieved successfully.", kpi)
}

// @Summary AR/AP KPI
// @Description Returns DSO, DPO, and overdue percentages from open invoices as of a date.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param as_of query string false "As-of date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} reporting.ArApKPI "AR/AP KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/ar-ap [get]
func (h KpiHandler) ArAp(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	asOf := time.Now().UTC()
	if raw := c.Query("as_of"); raw != "" {
		parsed, err := helper.ParseDate(&raw)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "as_of must be in YYYY-MM-DD format.", nil)
		}
		asOf = *parsed
	}
	kpi, err := h.svc.ArApKPI(c, *organizationID, asOf)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "AR/AP KPI retrieved successfully.", kpi)
}

// @Summary Cash KPI
// @Description Returns cash position, burn, and forecast from the GL bank accounts, payments, and open AR/AP.
// @Tags KPIs
// @Accept json
// @Produce json
// @Param as_of query string false "As-of date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} reporting.CashKPI "Cash KPI retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/kpis/cash [get]
func (h KpiHandler) Cash(c fiber.Ctx) error {
	organizationID, err := organizationOf(c)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	asOf := time.Now().UTC()
	if raw := c.Query("as_of"); raw != "" {
		parsed, err := helper.ParseDate(&raw)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "as_of must be in YYYY-MM-DD format.", nil)
		}
		asOf = *parsed
	}
	kpi, err := h.svc.CashKPI(c, *organizationID, asOf)
	if err != nil {
		return writeKpiError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Cash KPI retrieved successfully.", kpi)
}
