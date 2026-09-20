package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

// @Summary Create demand forecast
// @Description Creates a tenant-scoped demand forecast for a item and warehouse over a date period, which is later consumed by Planning runs as a source of demand. Returns the created forecast.
// @Tags Planning
// @Accept json
// @Produce json
// @Param body body CreateForecastRequest true "Forecast details"
// @Success 200 {object} DemandPlanResponseEnvelope "Demand forecast created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/planning/forecasts [post]
func (h PlanningHandler) CreateForecast(c fiber.Ctx) error {
	var request CreateForecastRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	periodStart, err := helper.ParseDateOrToday(request.PeriodStart)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid period start date provided.", nil)
	}
	periodEnd, err := helper.ParseDateOrToday(request.PeriodEnd)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid period end date provided.", nil)
	}
	forecast, err := h.svc.CreateForecast(c, &manufacturing.DemandPlan{
		OrganizationID: httpx.TenantOrganizationID(c, nil),
		ItemID:         request.ItemID,
		WarehouseID:    request.WarehouseID,
		PeriodStart:    helper.Ptr(periodStart),
		PeriodEnd:      helper.Ptr(periodEnd),
		ForecastQty:    request.ForecastQty,
	})
	if err != nil {
		httpx.RequestLog(c).Error("demand forecast create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create demand forecast.", err)
	}
	return httpx.CreateSuccessResponse(c, "Demand forecast created successfully.", newDemandPlanResponse(forecast))
}

// @Summary List demand forecasts
// @Description Lists tenant-scoped demand forecasts with pagination, optionally filtered by item. The result set is always scoped to the caller's organization even when no organization_id is supplied.
// @Tags Planning
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Success 200 {object} ListDemandPlansResponseEnvelope "Demand forecasts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/planning/forecasts [get]
func (h PlanningHandler) ListForecasts(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, map[string]struct{}{
		"organization_id": {},
		"item_id":         {},
	})
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.ListForecasts(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("forecast list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve demand forecasts.", err)
	}

	items := make([]DemandPlanResponse, len(page.Items))
	for i, forecast := range page.Items {
		items[i] = newDemandPlanResponse(forecast)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Demand forecasts retrieved successfully.", ListDemandPlansResponse{
		Forecasts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
