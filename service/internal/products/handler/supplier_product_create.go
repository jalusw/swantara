package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

type CreateSupplierProductRequest struct {
	ItemID              uint64     `json:"item_id" validate:"required,gt=0"`
	SupplierID          uint64     `json:"supplier_id" validate:"required,gt=0"`
	SupplierSku         *string    `json:"supplier_sku"`
	SupplierProductName *string    `json:"supplier_product_name"`
	MinQty              float64    `json:"min_qty"`
	Price               *float64   `json:"price"`
	CurrencyCode        *string    `json:"currency_code"`
	LeadTimeDays        *int       `json:"lead_time_days"`
	Priority            int        `json:"priority"`
	ValidFrom           *time.Time `json:"valid_from"`
	ValidTo             *time.Time `json:"valid_to"`
}

type CreateSupplierProductResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateSupplierProductResponse `json:"data"`
}
type CreateSupplierProductResponse struct {
	SupplierProduct SupplierProductResponse `json:"supplier_product"`
}

// @Summary Create supplier item
// @Description Adds a supplier catalog entry for a item variant within the caller's organization, validating that the variant exists and that the supplier is an active supplier. The validity window must be valid and the minimum quantity cannot be negative.
// @Tags Supplier Products
// @Accept json
// @Produce json
// @Param body body CreateSupplierProductRequest true "Supplier item details"
// @Success 201 {object} CreateSupplierProductResponseEnvelope "Supplier item created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Variant or supplier not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, unknown variant, or supplier not a supplier"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-products [post]
func (h SupplierProductHandler) Create(c fiber.Ctx) error {
	var request CreateSupplierProductRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	offer, err := h.svc.Create(c, organizationID, &products.SupplierProduct{
		ItemID:              request.ItemID,
		SupplierID:          request.SupplierID,
		SupplierSku:         request.SupplierSku,
		SupplierProductName: request.SupplierProductName,
		MinQty:              request.MinQty,
		Price:               request.Price,
		CurrencyCode:        request.CurrencyCode,
		LeadTimeDays:        request.LeadTimeDays,
		Priority:            request.Priority,
		ValidFrom:           request.ValidFrom,
		ValidTo:             request.ValidTo,
	})
	if err != nil {
		return writeSupplierProductError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Supplier item created successfully.", CreateSupplierProductResponse{
		SupplierProduct: newSupplierProductResponse(offer),
	})
}
