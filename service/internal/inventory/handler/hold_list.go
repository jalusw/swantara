package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListStockHoldsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListStockHoldsResponse `json:"data"`
}
type ListStockHoldsResponse struct {
	Reservations []StockHoldResponse `json:"reservations"`
}

// @Summary List stock reservations
// @Description Lists stock reservations with pagination, sorting, and filtering over the allowlisted query fields, such as the movement and quant they reference. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Stock Reservations
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListStockHoldsResponseEnvelope "Reservations retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-reservations [get]
func (h HoldHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, stockReservationQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	page, err := h.svc.ListInOrg(c, parsedQuery, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("reservation list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve reservations.", err)
	}

	items := make([]StockHoldResponse, len(page.Items))
	for i, reservation := range page.Items {
		items[i] = newStockHoldResponse(reservation)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "stock-reservations.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Reservations retrieved successfully.", ListStockHoldsResponse{
		Reservations: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
