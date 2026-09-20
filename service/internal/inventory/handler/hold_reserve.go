package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type ReserveStockRequest struct {
	ItemID     uint64  `json:"item_id" validate:"required,gt=0"`
	LocationID uint64  `json:"location_id" validate:"required,gt=0"`
	BatchID    *uint64 `json:"batch_id"`
	Qty        string  `json:"qty" validate:"required"`
	MovementID *uint64 `json:"movement_id"`
}

type ReserveStockResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ReserveStockResponse `json:"data"`
}
type ReserveStockResponse struct {
	Reservation StockHoldResponse `json:"reservation"`
}

// @Summary Reserve stock
// @Description Reserves a positive decimal quantity of a item at a stock location, optionally linked to a stock movement. The stock quant must exist and the quantity must not exceed the stock still available after subtracting already-reserved quantities, otherwise a 409 Conflict is returned. The reservation is recorded and the quant's reserved quantity is incremented accordingly.
// @Tags Stock Reservations
// @Accept json
// @Produce json
// @Param body body ReserveStockRequest true "Reservation details"
// @Success 201 {object} ReserveStockResponseEnvelope "Stock reserved successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-reservations [post]
func (h HoldHandler) Reserve(c fiber.Ctx) error {
	var request ReserveStockRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	qty, err := amount.FromString(request.Qty)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Quantity must be a valid decimal.", nil)
	}

	reservation, err := h.svc.Reserve(c, *organizationID, request.ItemID, request.LocationID, request.BatchID, qty, request.MovementID)
	if err != nil {
		return writeReservationError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Stock reserved successfully.", ReserveStockResponse{
		Reservation: newStockHoldResponse(reservation),
	})
}
