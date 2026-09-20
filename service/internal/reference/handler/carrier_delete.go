package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete carrier
// @Description Deletes a carrier.
// @Tags Carriers
// @Accept json
// @Produce json
// @Param id path int true "Carrier ID"
// @Success 204 "No content"
// @Failure 404 {object} httpx.ErrorResponse "Carrier not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /carriers/{id} [delete]
func (h CarrierHandler) Delete(c fiber.Ctx) error {
	carrierID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid carrier id provided.", nil)
	}

	carrier, err := h.svc.Find(c, carrierID)
	if err != nil {
		httpx.RequestLog(c).Error("carrier lookup failed", "carrier_id", carrierID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete carrier.", err)
	}
	if carrier == nil {
		return httpx.CreateNotFoundResponse(c, "Carrier not found.")
	}

	if err := h.svc.Delete(c, carrierID); err != nil {
		httpx.RequestLog(c).Error("carrier deletion failed", "carrier_id", carrierID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete carrier.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
