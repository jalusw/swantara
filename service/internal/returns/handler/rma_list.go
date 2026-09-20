package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListRMAsResponse struct {
	RMAs []RMAResponse `json:"rmas"`
}

type ListRMAsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListRMAsResponse `json:"data"`
}

// @Summary List RMAs
// @Description Lists return merchandise authorizations with pagination, sorting, and filtering. Results are scoped to the caller's organization and can be narrowed by type, contact, originating order, and state so the returns queue can be triaged at a glance.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListRMAsResponseEnvelope "RMAs retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas [get]
func (h RMAHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, rmaQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("rma list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve RMAs.", err)
	}

	items := make([]RMAResponse, len(page.Items))
	for i, rma := range page.Items {
		items[i] = newRMAResponse(rma)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "RMAs retrieved successfully.", ListRMAsResponse{
		RMAs: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
