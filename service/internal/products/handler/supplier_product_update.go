package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateSupplierProductRequest struct {
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

type UpdateSupplierProductResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateSupplierProductResponse `json:"data"`
}
type UpdateSupplierProductResponse struct {
	SupplierProduct SupplierProductResponse `json:"supplier_product"`
}

// @Summary Update supplier item
// @Description Updates a supplier catalog entry's supplier, pricing, minimum quantity, lead time, and validity window. A 404 is returned if the entry does not exist or belongs to another organization, and the validity window and minimum quantity are validated.
// @Tags Supplier Products
// @Accept json
// @Produce json
// @Param id path integer true "Supplier item ID"
// @Param body body UpdateSupplierProductRequest true "Supplier item details"
// @Success 200 {object} UpdateSupplierProductResponseEnvelope "Supplier item updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Supplier item not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-products/{id} [put]
func (h SupplierProductHandler) Update(c fiber.Ctx) error {
	offerID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid supplier item id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	var request UpdateSupplierProductRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	existing, err := h.svc.FindInOrg(c, offerID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("supplier item lookup failed", "supplier_item_id", offerID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update supplier item.", err)
	}
	if existing == nil {
		return httpx.CreateNotFoundResponse(c, "Supplier item not found.")
	}

	existing.SupplierID = request.SupplierID
	existing.SupplierSku = request.SupplierSku
	existing.SupplierProductName = request.SupplierProductName
	existing.MinQty = request.MinQty
	existing.Price = request.Price
	existing.CurrencyCode = request.CurrencyCode
	existing.LeadTimeDays = request.LeadTimeDays
	existing.Priority = request.Priority
	existing.ValidFrom = request.ValidFrom
	existing.ValidTo = request.ValidTo

	updated, err := h.svc.Update(c, organizationID, existing)
	if err != nil {
		return writeSupplierProductError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Supplier item updated successfully.", UpdateSupplierProductResponse{
		SupplierProduct: newSupplierProductResponse(updated),
	})
}
