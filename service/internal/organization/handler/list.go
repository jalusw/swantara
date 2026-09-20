package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListOrganizationsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListOrganizationsResponse `json:"data"`
}
type ListOrganizationsResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
}

// @Summary List organizations
// @Description Lists organizations with pagination, sorting, and filtering. Results can be narrowed by name, legal name, base currency, country, and timezone, and a CSV export is supported.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. name:eq:Acme)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListOrganizationsResponseEnvelope "Organizations retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations [get]
func (h OrganizationHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, organizationQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("organization list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve organizations.", err)
	}

	items := make([]OrganizationResponse, len(page.Items))
	for i, org := range page.Items {
		items[i] = newOrganizationResponse(org)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "organizations.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Organizations retrieved successfully.", ListOrganizationsResponse{
		Organizations: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
