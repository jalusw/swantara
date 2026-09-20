package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListCarriersResponse struct {
	Carriers []CarrierResponse `json:"carriers"`
}

type ListCarriersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCarriersResponse `json:"data"`
}

// @Summary List carriers
// @Description Lists carriers with pagination, sorting, and filtering. Supports CSV export.
// @Tags Carriers
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param size query int false "Page size"
// @Param sort query string false "Sort fields (e.g. name:asc)"
// @Success 200 {object} ListCarriersResponseEnvelope "Carriers retrieved successfully."
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /carriers [get]
func (h CarrierHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, carrierQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("carrier list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve carriers.", err)
	}

	items := make([]CarrierResponse, len(page.Items))
	for i, carrier := range page.Items {
		items[i] = newCarrierResponse(carrier)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "carriers.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Carriers retrieved successfully.", ListCarriersResponse{
		Carriers: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
