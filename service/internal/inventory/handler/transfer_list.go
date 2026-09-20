package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List warehouse transfers
// @Description Lists the warehouse transfers belonging to the caller's organization, honoring pagination, sorting, and filtering. Only allowlisted query fields are accepted and the tenant filter is always enforced, so callers can never see another organization's warehouse transfers. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Transfer Orders
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListWarehouseTransfersResponseEnvelope "Warehouse transfers retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/transfer-orders [get]
func (h TransferHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, transferOrderQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("warehouse transfer list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve warehouse transfers.", err)
	}

	items := make([]WarehouseTransferResponse, len(page.Items))
	for i, transfer := range page.Items {
		items[i] = newWarehouseTransferResponse(transfer)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "transfer-orders.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Warehouse transfers retrieved successfully.", ListWarehouseTransfersResponse{
		WarehouseTransfers: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
