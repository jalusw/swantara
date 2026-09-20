package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListSystemConfigsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSystemConfigsResponse `json:"data"`
}

type ListSystemConfigsResponse struct {
	SystemConfigs []SystemConfigResponse `json:"system_configs"`
}

// @Summary List system configs
// @Description Lists system configs with pagination, sorting, and filtering, scoping results to the caller's organization when present; the response can be exported as JSON, XML, or CSV.
// @Tags System Configs
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. key:desc)"
// @Param filter query string false "Filters (repeatable, e.g. key:eq:currency)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListSystemConfigsResponseEnvelope "System configs retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/system-configs [get]
func (h SystemConfigHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, systemConfigQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.configSvc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("system config list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve system configs.", err)
	}

	items := make([]SystemConfigResponse, len(page.Items))
	for i, config := range page.Items {
		items[i] = newSystemConfigResponse(config)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "system-configs.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "System configs retrieved successfully.", ListSystemConfigsResponse{
		SystemConfigs: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
