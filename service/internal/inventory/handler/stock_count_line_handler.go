package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListStockCountLinesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListStockCountLinesResponse `json:"data"`
}
type ListStockCountLinesResponse struct {
	Lines []StockCountLineResponse `json:"lines"`
}

// @Summary List inventory count lines
// @Description Lists the lines of an inventory count, including the theoretical, counted, and difference quantities for each item, optionally exported as CSV. The count must belong to the caller's organization, otherwise a 404 Not Found is returned.
// @Tags Inventory Counts
// @Accept json
// @Produce json
// @Param id path integer true "Inventory count ID"
// @Success 200 {object} ListStockCountLinesResponseEnvelope "Inventory count lines retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Inventory count not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/inventory-counts/{id}/lines [get]
func (h StockCountHandler) ListLines(c fiber.Ctx) error {
	stockCountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inventory count id provided.", nil)
	}

	count, err := h.svc.Find(c, stockCountID)
	if err != nil {
		httpx.RequestLog(c).Error("inventory count lookup failed", "stock_count_id", stockCountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve inventory count lines.", err)
	}
	if count == nil || !httpx.OwnsTenant(c, count.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Inventory count not found.")
	}

	lines, err := h.svc.ListLines(c, stockCountID)
	if err != nil {
		httpx.RequestLog(c).Error("inventory count line list failed", "stock_count_id", stockCountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve inventory count lines.", err)
	}

	items := make([]StockCountLineResponse, len(lines))
	for i, line := range lines {
		items[i] = newStockCountLineResponse(line)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "inventory-count-lines.csv", items)
	}

	return httpx.CreateSuccessResponse(c, "Inventory count lines retrieved successfully.", ListStockCountLinesResponse{
		Lines: items,
	})
}
