package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListBatchsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListBatchsResponse `json:"data"`
}
type ListBatchsResponse struct {
	Batches []BatchResponse `json:"batches"`
}

// @Summary List stock lots
// @Description Lists stock lots with pagination, sorting, and filtering over the allowlisted query fields, such as the item and batch name. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Stock Lots
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListBatchsResponseEnvelope "Stock lots retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-lots [get]
func (h BatchHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, stockLotQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	page, err := h.svc.ListInOrg(c, parsedQuery, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("stock batch list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve stock lots.", err)
	}

	items := make([]BatchResponse, len(page.Items))
	for i, batch := range page.Items {
		items[i] = newBatchResponse(batch)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "stock-lots.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Stock lots retrieved successfully.", ListBatchsResponse{
		Batches: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
