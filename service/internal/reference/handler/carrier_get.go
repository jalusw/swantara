package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetCarrierResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetCarrierResponse `json:"data"`
}
type GetCarrierResponse struct {
	Carrier CarrierResponse `json:"carrier"`
}

// @Summary Get carrier
// @Description Gets a single carrier by id.
// @Tags Carriers
// @Accept json
// @Produce json
// @Param id path int true "Carrier ID"
// @Success 200 {object} GetCarrierResponseEnvelope "Carrier retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Carrier not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /carriers/{id} [get]
func (h CarrierHandler) Get(c fiber.Ctx) error {
	carrierID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid carrier id provided.", nil)
	}

	carrier, err := h.svc.Find(c, carrierID)
	if err != nil {
		httpx.RequestLog(c).Error("carrier lookup failed", "carrier_id", carrierID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get carrier.", err)
	}
	if carrier == nil {
		return httpx.CreateNotFoundResponse(c, "Carrier not found.")
	}

	return httpx.CreateSuccessResponse(c, "Carrier retrieved successfully.", GetCarrierResponse{
		Carrier: newCarrierResponse(carrier),
	})
}
