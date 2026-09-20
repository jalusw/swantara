package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type StockCountHandler struct {
	svc inventory.StockCountService
}

func NewStockCountHandler(svc inventory.StockCountService) StockCountHandler {
	return StockCountHandler{svc: svc}
}

type StockCountResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	LocationID     *uint64    `json:"location_id"`
	State          string     `json:"state"`
	CountDate      *time.Time `json:"count_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newStockCountResponse(count *inventory.StockCount) StockCountResponse {
	return StockCountResponse{
		ID:             count.ID,
		OrganizationID: count.OrganizationID,
		Name:           count.Name,
		LocationID:     count.LocationID,
		State:          count.State,
		CountDate:      count.CountDate,
		CreatedAt:      count.CreatedAt,
		UpdatedAt:      count.UpdatedAt,
	}
}

type StockCountLineResponse struct {
	ID             uint64  `json:"id"`
	StockCountID   uint64  `json:"stock_count_id"`
	ItemID         uint64  `json:"item_id"`
	BatchID        *uint64 `json:"batch_id"`
	TheoreticalQty float64 `json:"theoretical_qty"`
	CountedQty     float64 `json:"counted_qty"`
	DiffQty        float64 `json:"diff_qty"`
}

func newStockCountLineResponse(line *inventory.StockCountLine) StockCountLineResponse {
	return StockCountLineResponse{
		ID:             line.ID,
		StockCountID:   line.StockCountID,
		ItemID:         line.ItemID,
		BatchID:        line.BatchID,
		TheoreticalQty: line.TheoreticalQty,
		CountedQty:     line.CountedQty,
		DiffQty:        line.DiffQty,
	}
}

var inventoryCountQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"location_id":     {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}

func writeCountError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrStockCountNotFound):
		return httpx.CreateNotFoundResponse(c, "Inventory count not found.")
	case errors.Is(err, inventory.ErrStockCountState):
		return httpx.CreateConflictResponse(c, "Inventory count cannot be posted in its current state.", err)
	case errors.Is(err, inventory.ErrStockCountNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Inventory count has no lines.", nil)
	case errors.Is(err, inventory.ErrLocationRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A stock location is required.", nil)
	case errors.Is(err, inventory.ErrGainLossAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "An inventory gain/loss account is required.", nil)
	case errors.Is(err, inventory.ErrOrganizationMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required for posting.", nil)
	case errors.Is(err, inventory.ErrValuationAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no stock valuation account.", nil)
	default:
		httpx.RequestLog(c).Error("inventory count write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save inventory count.", err)
	}
}
