package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List fiscal positions
// @Description Lists fiscal positions for the caller's tenant, applying pagination, sorting, and filtering over fields such as country code, auto-apply, and active state. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Fiscal Positions
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. active:eq:true)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListTaxRulesResponseEnvelope "Fiscal positions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-rules [get]
func (h TaxRuleHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, taxRuleQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.resolver.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("fiscal position list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve fiscal positions.", err)
	}

	items := make([]TaxRuleResponse, len(page.Items))
	for i, position := range page.Items {
		items[i] = newTaxRuleResponse(position)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "tax-rules.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Fiscal positions retrieved successfully.", ListTaxRulesResponse{
		TaxRules: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
