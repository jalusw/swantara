package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type BatchHandler struct {
	svc inventory.BatchService
}

func NewBatchHandler(svc inventory.BatchService) BatchHandler {
	return BatchHandler{svc: svc}
}

type BatchResponse struct {
	ID             uint64     `json:"id"`
	ItemID         uint64     `json:"item_id"`
	Name           string     `json:"name"`
	Ref            *string    `json:"ref"`
	ExpiryDate     *time.Time `json:"expiry_date"`
	BestBeforeDate *time.Time `json:"best_before_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newBatchResponse(batch *inventory.Batch) BatchResponse {
	return BatchResponse{
		ID:             batch.ID,
		ItemID:         batch.ItemID,
		Name:           batch.Name,
		Ref:            batch.Ref,
		ExpiryDate:     batch.ExpiryDate,
		BestBeforeDate: batch.BestBeforeDate,
		CreatedAt:      batch.CreatedAt,
		UpdatedAt:      batch.UpdatedAt,
	}
}

var stockLotQueryAllowlist = map[string]struct{}{
	"item_id":    {},
	"name":       {},
	"created_at": {},
	"updated_at": {},
}

func writeLotError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrVariantNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item variant does not exist.", nil)
	case errors.Is(err, inventory.ErrItemNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item template does not exist.", nil)
	case errors.Is(err, inventory.ErrBatchTrackingDisabled):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item does not use batch/serial tracking.", nil)
	case errors.Is(err, inventory.ErrBatchNameTaken):
		return httpx.CreateConflictResponse(c, "Lot name already exists for this item.", err)
	default:
		httpx.RequestLog(c).Error("stock batch write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save stock batch.", err)
	}
}
