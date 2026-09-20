package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ReleaseByMovementRequest struct {
	MovementID uint64 `json:"movement_id" validate:"required,gt=0"`
}

// @Summary Release reservations by movement
// @Description Releases every stock reservation linked to a given stock movement, freeing that stock back to the available pool. Each affected quant has its reserved quantity decremented by the released amounts. Returns an empty envelope on success.
// @Tags Stock Reservations
// @Accept json
// @Produce json
// @Param body body ReleaseByMovementRequest true "Movement details"
// @Success 200 {object} httpx.EmptyEnvelope "Stock released successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-reservations/release-by-movement [post]
func (h HoldHandler) ReleaseByMovement(c fiber.Ctx) error {
	var request ReleaseByMovementRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	if err := h.svc.ReleaseByMovementInOrg(c, *organizationID, request.MovementID); err != nil {
		return writeReservationError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Stock released successfully.", struct{}{})
}

// @Summary Release stock reservation
// @Description Releases a single stock reservation by its id, freeing the reserved quantity back and decrementing it on the reservation's quant. A missing reservation is answered with 404 Not Found. Responds with 204 No Content on success.
// @Tags Stock Reservations
// @Accept json
// @Produce json
// @Param id path integer true "Reservation ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Reservation not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /stock-reservations/{id} [delete]
func (h HoldHandler) Release(c fiber.Ctx) error {
	reservationID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid reservation id provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	reservation, err := h.svc.FindInOrg(c, reservationID, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("reservation lookup failed", "reservation_id", reservationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to release reservation.", err)
	}
	if reservation == nil {
		return httpx.CreateNotFoundResponse(c, "Reservation not found.")
	}

	if err := h.svc.Release(c, reservationID); err != nil {
		return writeReservationError(c, err)
	}

	return httpx.CreateNoContentResponse(c)
}
