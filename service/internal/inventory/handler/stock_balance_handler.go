package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type StockBalanceResponse struct {
	ID          uint64    `json:"id"`
	ItemID      uint64    `json:"item_id"`
	LocationID  uint64    `json:"location_id"`
	BatchID     *uint64   `json:"batch_id"`
	Quantity    float64   `json:"quantity"`
	ReservedQty float64   `json:"reserved_qty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newStockBalanceResponse(quant *inventory.StockBalance) StockBalanceResponse {
	return StockBalanceResponse{
		ID:          quant.ID,
		ItemID:      quant.ItemID,
		LocationID:  quant.LocationID,
		BatchID:     quant.BatchID,
		Quantity:    quant.Quantity,
		ReservedQty: quant.ReservedQty,
		CreatedAt:   quant.CreatedAt,
		UpdatedAt:   quant.UpdatedAt,
	}
}

var stockQuantQueryAllowlist = map[string]struct{}{
	"item_id":     {},
	"location_id": {},
	"batch_id":    {},
	"created_at":  {},
	"updated_at":  {},
}

type ListStockBalancesResponse struct {
	Balances []StockBalanceResponse `json:"balances"`
}

type ListStockBalancesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListStockBalancesResponse `json:"data"`
}

// @Summary List stock quants
// @Description Lists the stock quants belonging to the caller's organization, honoring pagination, sorting, and filtering. Only allowlisted query fields are accepted and the tenant filter is always enforced, so callers can never see another organization's quants. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Stock
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListStockBalancesResponseEnvelope "Stock quants retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock/quants [get]
func (h StockHandler) ListBalances(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, stockQuantQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.stock.ListBalances(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("stock quant list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve stock quants.", err)
	}

	items := make([]StockBalanceResponse, len(page.Items))
	for i, quant := range page.Items {
		items[i] = newStockBalanceResponse(quant)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "stock-quants.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Stock quants retrieved successfully.", ListStockBalancesResponse{
		Balances: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type OnHandResponse struct {
	ItemID     uint64  `json:"item_id"`
	LocationID uint64  `json:"location_id"`
	OnHand     float64 `json:"on_hand"`
}

type OnHandResponseEnvelope struct {
	httpx.EnvelopeBase
	Data OnHandResponse `json:"data"`
}

// @Summary Get on hand quantity
// @Description Gets the current on-hand quantity of a item at a given stock location for the caller's organization. Both item_id and location_id query parameters are required and must be positive integers. If the location does not exist or does not belong to the caller's organization, a 404 Not Found is returned; otherwise the total quantity held in the quant is reported.
// @Tags Stock
// @Accept json
// @Produce json
// @Param item_id query integer true "Item ID"
// @Param location_id query integer true "Stock location ID"
// @Success 200 {object} OnHandResponseEnvelope "On hand quantity retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock location not found"
// @Failure 422 {object} httpx.ErrorResponse "Missing or invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock/on-hand [get]
func (h StockHandler) OnHand(c fiber.Ctx) error {
	itemID, err := strconv.ParseUint(c.Query("item_id"), 10, 64)
	if err != nil || itemID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "item_id is required.", nil)
	}
	locationID, err := strconv.ParseUint(c.Query("location_id"), 10, 64)
	if err != nil || locationID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "location_id is required.", nil)
	}

	var organizationID *uint64
	if id, ok := httpx.CallerOrganizationID(c); ok {
		organizationID = helper.Ptr(id)
	}

	onHand, err := h.ledger.OnHand(c, organizationID, itemID, locationID)
	if err != nil {
		if errors.Is(err, inventory.ErrLocationNotFound) {
			return httpx.CreateNotFoundResponse(c, "Stock location not found.")
		}
		httpx.RequestLog(c).Error("on hand lookup failed", "item_id", itemID, "location_id", locationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get on hand quantity.", err)
	}

	return httpx.CreateSuccessResponse(c, "On hand quantity retrieved successfully.", OnHandResponse{
		ItemID:     itemID,
		LocationID: locationID,
		OnHand:     onHand,
	})
}

type AvailableToPromiseResponse struct {
	ItemID     uint64  `json:"item_id"`
	LocationID uint64  `json:"location_id"`
	Available  float64 `json:"available"`
}

type AvailableToPromiseResponseEnvelope struct {
	httpx.EnvelopeBase
	Data AvailableToPromiseResponse `json:"data"`
}

// @Summary Get available to promise
// @Description Gets the available-to-promise quantity of a item at a given stock location for the caller's organization, computed as the on-hand quantity minus the quantity already reserved. Both item_id and location_id query parameters are required and must be positive integers. If the location does not exist or does not belong to the caller's organization, a 404 Not Found is returned.
// @Tags Stock
// @Accept json
// @Produce json
// @Param item_id query integer true "Item ID"
// @Param location_id query integer true "Stock location ID"
// @Success 200 {object} AvailableToPromiseResponseEnvelope "Available to promise retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock location not found"
// @Failure 422 {object} httpx.ErrorResponse "Missing or invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock/available-to-promise [get]
func (h StockHandler) AvailableToPromise(c fiber.Ctx) error {
	itemID, err := strconv.ParseUint(c.Query("item_id"), 10, 64)
	if err != nil || itemID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "item_id is required.", nil)
	}
	locationID, err := strconv.ParseUint(c.Query("location_id"), 10, 64)
	if err != nil || locationID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "location_id is required.", nil)
	}

	var organizationID *uint64
	if id, ok := httpx.CallerOrganizationID(c); ok {
		organizationID = helper.Ptr(id)
	}

	available, err := h.ledger.AvailableToPromise(c, organizationID, itemID, locationID)
	if err != nil {
		if errors.Is(err, inventory.ErrLocationNotFound) {
			return httpx.CreateNotFoundResponse(c, "Stock location not found.")
		}
		httpx.RequestLog(c).Error("atp lookup failed", "item_id", itemID, "location_id", locationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get available to promise.", err)
	}

	return httpx.CreateSuccessResponse(c, "Available to promise retrieved successfully.", AvailableToPromiseResponse{
		ItemID:     itemID,
		LocationID: locationID,
		Available:  available,
	})
}

// @Summary Rebuild stock quants
// @Description Rebuilds the stock quants of the caller's organization from the stock movement ledger. Each item/location/batch quant is re-derived from the aggregate ledger totals, updating existing quants, creating missing ones, and zeroing any quant that no longer has a corresponding ledger balance. This is useful for reconciling quants after manual or external data changes.
// @Tags Stock
// @Accept json
// @Produce json
// @Success 200 {object} httpx.EmptyEnvelope "Stock quants rebuilt successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock/rebuild [post]
func (h StockHandler) RebuildBalances(c fiber.Ctx) error {
	var organizationID *uint64
	if id, ok := httpx.CallerOrganizationID(c); ok {
		organizationID = helper.Ptr(id)
	}

	if err := h.ledger.RebuildBalances(c, organizationID); err != nil {
		httpx.RequestLog(c).Error("stock quant rebuild failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to rebuild stock quants.", err)
	}

	return httpx.CreateSuccessResponse(c, "Stock quants rebuilt successfully.", struct{}{})
}
