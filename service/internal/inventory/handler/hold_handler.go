package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type HoldHandler struct {
	svc inventory.HoldService
}

func NewHoldHandler(svc inventory.HoldService) HoldHandler {
	return HoldHandler{svc: svc}
}

type StockHoldResponse struct {
	ID         uint64    `json:"id"`
	MovementID *uint64   `json:"movement_id"`
	BalanceID  uint64    `json:"balance_id"`
	Qty        float64   `json:"qty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func newStockHoldResponse(reservation *inventory.StockHold) StockHoldResponse {
	return StockHoldResponse{
		ID:         reservation.ID,
		MovementID: reservation.MovementID,
		BalanceID:  reservation.BalanceID,
		Qty:        reservation.Qty,
		CreatedAt:  reservation.CreatedAt,
		UpdatedAt:  reservation.UpdatedAt,
	}
}

var stockReservationQueryAllowlist = map[string]struct{}{
	"movement_id": {},
	"balance_id":  {},
	"created_at":  {},
	"updated_at":  {},
}

func writeReservationError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrBalanceNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock quant does not exist.", nil)
	case errors.Is(err, inventory.ErrMovementQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Reservation quantity must be greater than zero.", nil)
	case errors.Is(err, inventory.ErrHoldOverflow):
		return httpx.CreateConflictResponse(c, "Reservation exceeds available stock.", err)
	case errors.Is(err, inventory.ErrHoldNotFound):
		return httpx.CreateNotFoundResponse(c, "Reservation not found.")
	default:
		httpx.RequestLog(c).Error("reservation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process reservation.", err)
	}
}
