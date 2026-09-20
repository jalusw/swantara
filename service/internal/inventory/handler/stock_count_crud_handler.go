package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type ListStockCountsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListStockCountsResponse `json:"data"`
}
type ListStockCountsResponse struct {
	Counts []StockCountResponse `json:"counts"`
}

// @Summary List inventory counts
// @Description Lists the inventory counts belonging to the caller's organization, honoring pagination, sorting, and filtering. Only allowlisted query fields are accepted and the tenant filter is always enforced, so callers can never see another organization's counts. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Inventory Counts
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListStockCountsResponseEnvelope "Inventory counts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/inventory-counts [get]
func (h StockCountHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, inventoryCountQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("inventory count list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve inventory counts.", err)
	}

	items := make([]StockCountResponse, len(page.Items))
	for i, count := range page.Items {
		items[i] = newStockCountResponse(count)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "inventory-counts.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Inventory counts retrieved successfully.", ListStockCountsResponse{
		Counts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetStockCountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetStockCountResponse `json:"data"`
}
type GetStockCountResponse struct {
	Count StockCountResponse `json:"count"`
}

// @Summary Get inventory count
// @Description Gets a single inventory count by its id, including its state, location, and count date. The count is looked up and returned only when it belongs to the caller's organization; otherwise the request is answered with 404 Not Found to avoid leaking the record's existence.
// @Tags Inventory Counts
// @Accept json
// @Produce json
// @Param id path integer true "Inventory count ID"
// @Success 200 {object} GetStockCountResponseEnvelope "Inventory count retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Inventory count not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/inventory-counts/{id} [get]
func (h StockCountHandler) Get(c fiber.Ctx) error {
	stockCountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inventory count id provided.", nil)
	}

	count, err := h.svc.Find(c, stockCountID)
	if err != nil {
		httpx.RequestLog(c).Error("inventory count lookup failed", "stock_count_id", stockCountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get inventory count.", err)
	}
	if count == nil || !httpx.OwnsTenant(c, count.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Inventory count not found.")
	}

	return httpx.CreateSuccessResponse(c, "Inventory count retrieved successfully.", GetStockCountResponse{
		Count: newStockCountResponse(count),
	})
}

type StockCountLineRequest struct {
	ItemID     uint64  `json:"item_id" validate:"required,gt=0"`
	BatchID    *uint64 `json:"batch_id"`
	CountedQty float64 `json:"counted_qty"`
}

type CreateStockCountRequest struct {
	OrganizationID *uint64                 `json:"organization_id"`
	Name           *string                 `json:"name"`
	LocationID     *uint64                 `json:"location_id" validate:"required"`
	CountDate      *string                 `json:"count_date"`
	Lines          []StockCountLineRequest `json:"lines"`
}

type CreateStockCountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateStockCountResponse `json:"data"`
}
type CreateStockCountResponse struct {
	Count StockCountResponse `json:"count"`
}

// @Summary Create inventory count
// @Description Creates a draft inventory count for a stock location, which is required. Each counted line stores the item's current on-hand quantity as the theoretical quantity alongside the counted quantity and the difference between them. Returns the created count with a 201 status.
// @Tags Inventory Counts
// @Accept json
// @Produce json
// @Param body body CreateStockCountRequest true "Inventory count details"
// @Success 201 {object} CreateStockCountResponseEnvelope "Inventory count created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/inventory-counts [post]
func (h StockCountHandler) Create(c fiber.Ctx) error {
	var request CreateStockCountRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	lines := make([]inventory.StockCountLineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = inventory.StockCountLineRequest{
			ItemID:     line.ItemID,
			BatchID:    line.BatchID,
			CountedQty: line.CountedQty,
		}
	}

	countDate, err := helper.ParseDate(request.CountDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid count date provided.", nil)
	}

	count, err := h.svc.Create(c, &inventory.StockCount{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		Name:           request.Name,
		LocationID:     request.LocationID,
		CountDate:      countDate,
	}, lines)
	if err != nil {
		return writeCountError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Inventory count created successfully.", CreateStockCountResponse{
		Count: newStockCountResponse(count),
	})
}

// @Summary Delete inventory count
// @Description Deletes an inventory count by its id. The count is removed only when it exists and belongs to the caller's organization; otherwise a 404 Not Found is returned so the record's existence is not revealed. Responds with 204 No Content on success.
// @Tags Inventory Counts
// @Accept json
// @Produce json
// @Param id path integer true "Inventory count ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Inventory count not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/inventory-counts/{id} [delete]
func (h StockCountHandler) Delete(c fiber.Ctx) error {
	stockCountID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inventory count id provided.", nil)
	}

	count, err := h.svc.Find(c, stockCountID)
	if err != nil {
		httpx.RequestLog(c).Error("inventory count lookup failed", "stock_count_id", stockCountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete inventory count.", err)
	}
	if count == nil || !httpx.OwnsTenant(c, count.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Inventory count not found.")
	}

	if err := h.svc.Delete(c, stockCountID); err != nil {
		httpx.RequestLog(c).Error("inventory count deletion failed", "stock_count_id", stockCountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete inventory count.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
