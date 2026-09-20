package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CarrierHandler struct {
	svc reference.CarrierService
}

func NewCarrierHandler(svc reference.CarrierService) CarrierHandler {
	return CarrierHandler{svc: svc}
}

type CarrierResponse struct {
	ID             uint64    `json:"id"`
	Name           string    `json:"name"`
	TrackingURLTpl *string   `json:"tracking_url_tpl"`
	DeliveryItemID *uint64   `json:"delivery_item_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newCarrierResponse(carrier *reference.Carrier) CarrierResponse {
	return CarrierResponse{
		ID:             carrier.ID,
		Name:           carrier.Name,
		TrackingURLTpl: carrier.TrackingURLTpl,
		DeliveryItemID: carrier.DeliveryItemID,
		CreatedAt:      carrier.CreatedAt,
		UpdatedAt:      carrier.UpdatedAt,
	}
}

var carrierQueryAllowlist = map[string]struct{}{
	"name":             {},
	"tracking_url_tpl": {},
	"delivery_item_id": {},
	"created_at":       {},
	"updated_at":       {},
}

func writeCarrierError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reference.ErrCarrierName):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Carrier name is required.", nil)
	case errors.Is(err, reference.ErrCarrierDeliveryProduct):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Delivery item does not exist.", nil)
	default:
		httpx.RequestLog(c).Error("carrier write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save carrier.", err)
	}
}
