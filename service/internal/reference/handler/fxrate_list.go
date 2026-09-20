package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListFxRatesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListFxRatesResponse `json:"data"`
}
type ListFxRatesResponse struct {
	FxRates []FxRateResponse `json:"fx_rates"`
}

// @Summary List FX rates
// @Description Lists FX rates with pagination, sorting, and filtering, automatically scoping results to the caller's organization; the response can be exported as JSON, XML, or CSV.
// @Tags FX Rates
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. valid_from:desc)"
// @Param filter query string false "Filters (repeatable, e.g. currency_code:eq:USD)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListFxRatesResponseEnvelope "FX rates retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/fx-rates [get]
func (h FxRateHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, fxRateQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.fxRateSvc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("fx rate list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve FX rates.", err)
	}

	items := make([]FxRateResponse, len(page.Items))
	for i, rate := range page.Items {
		items[i] = newFxRateResponse(rate)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "fx-rates.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "FX rates retrieved successfully.", ListFxRatesResponse{
		FxRates: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
