package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
)

type CreateDropshipOrderRequest struct {
	SaleOrderID    uint64  `json:"sale_order_id" validate:"required,gt=0"`
	SupplierID     uint64  `json:"supplier_id" validate:"required,gt=0"`
	DestLocationID *uint64 `json:"dest_location_id"`
	Date           string  `json:"date"`
}

type CreateDropshipOrderResponse struct {
	PurchaseOrder DropshipOrderResponse  `json:"purchase_order"`
	Links         []DropshipLinkResponse `json:"links"`
}

type CreateDropshipOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateDropshipOrderResponse `json:"data"`
}

// @Summary Create drop-ship order
// @Description Creates a purchase order from a sale order that ships directly from the supplier to the customer without entering stock. Each mirrored sale line is linked through dropship_links.
// @Tags Drop-shipping
// @Accept json
// @Produce json
// @Param body body CreateDropshipOrderRequest true "Drop-ship details"
// @Success 201 {object} CreateDropshipOrderResponseEnvelope "Drop-ship order created successfully."
// @Failure 404 {object} httpx.ErrorResponse "Source order not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/dropship-orders [post]
func (h DropShipHandler) Create(c fiber.Ctx) error {
	var request CreateDropshipOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	date := time.Now()
	if request.Date != "" {
		parsed, err := helper.ParseDateStr(request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date.", nil)
		}
		date = parsed
	}
	order, links, err := h.svc.Create(c, interorganization.CreateDropshipOrderRequest{
		OrganizationID: *organizationID,
		SaleOrderID:    request.SaleOrderID,
		SupplierID:     request.SupplierID,
		DestLocationID: request.DestLocationID,
		Date:           date,
	})
	if err != nil {
		return writeInterorganizationError(c, err)
	}
	items := make([]DropshipLinkResponse, len(links))
	for i, link := range links {
		items[i] = newDropshipLinkResponse(link)
	}
	return httpx.CreateCreatedResponse(c, "Drop-ship order created successfully.", CreateDropshipOrderResponse{
		PurchaseOrder: newDropshipOrderResponse(order),
		Links:         items,
	})
}
