package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

var taxReturnQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"period_id":       {},
	"type":            {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}

type ListTaxReturnsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListTaxReturnsResponse `json:"data"`
}
type ListTaxReturnsResponse struct {
	TaxReturns []TaxReturnResponse `json:"tax_returns"`
}

// @Summary List tax returns
// @Description Lists tax returns for the caller's tenant, applying pagination, sorting, and filtering over fields such as tax period, type, and state. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Tax Returns
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. created_at:desc)"
// @Param filter query string false "Filters (repeatable, e.g. state:eq:draft)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListTaxReturnsResponseEnvelope "Tax returns retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-returns [get]
func (h TaxReturnHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, taxReturnQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("tax return list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve tax returns.", err)
	}

	items := make([]TaxReturnResponse, len(page.Items))
	for i, taxReturn := range page.Items {
		items[i] = newTaxReturnResponse(taxReturn)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "tax-returns.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Tax returns retrieved successfully.", ListTaxReturnsResponse{
		TaxReturns: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
