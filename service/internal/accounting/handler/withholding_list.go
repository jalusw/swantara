package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListWithholdingTaxesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListWithholdingTaxesResponse `json:"data"`
}
type ListWithholdingTaxesResponse struct {
	WithholdingTaxes []WithholdingTaxResponse `json:"withholding_taxes"`
}

// @Summary List withholding taxes
// @Description Lists withholding tax definitions for the caller's tenant, applying pagination, sorting, and filtering over fields such as name, rate, account, and scope. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Withholding Taxes
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. scope:eq:purchase)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListWithholdingTaxesResponseEnvelope "Withholding taxes retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/withholding-taxes [get]
func (h WithholdingTaxHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, withholdingTaxQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("withholding tax list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve withholding taxes.", err)
	}

	items := make([]WithholdingTaxResponse, len(page.Items))
	for i, tax := range page.Items {
		items[i] = newWithholdingTaxResponse(tax)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "withholding-taxes.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Withholding taxes retrieved successfully.", ListWithholdingTaxesResponse{
		WithholdingTaxes: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
