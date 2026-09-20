package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List taxes
// @Description Lists tax definitions with pagination, sorting, and filtering, automatically scoping results to the caller's organization; the response can be exported as JSON, XML, or CSV.
// @Tags Taxes
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. scope:eq:sale)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListTaxesResponseEnvelope "Taxes retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/taxes [get]
func (h TaxHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, taxQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("tax list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve taxes.", err)
	}

	items := make([]TaxResponse, len(page.Items))
	for i, tax := range page.Items {
		items[i] = newTaxResponse(tax)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "taxes.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Taxes retrieved successfully.", ListTaxesResponse{
		Taxes: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
