package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateCarrierRequest struct {
	Name           string  `json:"name" validate:"required"`
	TrackingURLTpl *string `json:"tracking_url_tpl"`
	DeliveryItemID *uint64 `json:"delivery_item_id"`
}

type UpdateCarrierResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateCarrierResponse `json:"data"`
}
type UpdateCarrierResponse struct {
	Carrier CarrierResponse `json:"carrier"`
}

// @Summary Update carrier
// @Description Updates an existing carrier.
// @Tags Carriers
// @Accept json
// @Produce json
// @Param id path int true "Carrier ID"
// @Param body body UpdateCarrierRequest true "Carrier payload"
// @Success 200 {object} UpdateCarrierResponseEnvelope "Carrier updated successfully."
// @Failure 404 {object} httpx.ErrorResponse "Carrier not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /carriers/{id} [put]
func (h CarrierHandler) Update(c fiber.Ctx) error {
	carrierID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid carrier id provided.", nil)
	}

	var request UpdateCarrierRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	carrier, err := h.svc.Find(c, carrierID)
	if err != nil {
		httpx.RequestLog(c).Error("carrier lookup failed", "carrier_id", carrierID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update carrier.", err)
	}
	if carrier == nil {
		return httpx.CreateNotFoundResponse(c, "Carrier not found.")
	}

	carrier.Name = request.Name
	carrier.TrackingURLTpl = request.TrackingURLTpl
	carrier.DeliveryItemID = request.DeliveryItemID

	updated, err := h.svc.Update(c, carrier)
	if err != nil {
		return writeCarrierError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Carrier updated successfully.", UpdateCarrierResponse{
		Carrier: newCarrierResponse(updated),
	})
}
