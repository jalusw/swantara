package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateCarrierRequest struct {
	Name           string  `json:"name" validate:"required"`
	TrackingURLTpl *string `json:"tracking_url_tpl"`
	DeliveryItemID *uint64 `json:"delivery_item_id"`
}

type CreateCarrierResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCarrierResponse `json:"data"`
}
type CreateCarrierResponse struct {
	Carrier CarrierResponse `json:"carrier"`
}

// @Summary Create carrier
// @Description Creates a new carrier.
// @Tags Carriers
// @Accept json
// @Produce json
// @Param body body CreateCarrierRequest true "Carrier payload"
// @Success 201 {object} CreateCarrierResponseEnvelope "Carrier created successfully."
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /carriers [post]
func (h CarrierHandler) Create(c fiber.Ctx) error {
	var request CreateCarrierRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	carrier, err := h.svc.Create(c, &reference.Carrier{
		Name:           request.Name,
		TrackingURLTpl: request.TrackingURLTpl,
		DeliveryItemID: request.DeliveryItemID,
	})
	if err != nil {
		return writeCarrierError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Carrier created successfully.", CreateCarrierResponse{
		Carrier: newCarrierResponse(carrier),
	})
}
