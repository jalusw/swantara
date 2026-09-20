package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetSupplierProductResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSupplierProductResponse `json:"data"`
}
type GetSupplierProductResponse struct {
	SupplierProduct SupplierProductResponse `json:"supplier_product"`
}

// @Summary Get supplier item
// @Description Returns a single supplier catalog entry by id with its supplier pricing, minimum quantity, lead time, and validity window; a 404 is returned if the entry does not exist or belongs to another organization.
// @Tags Supplier Products
// @Accept json
// @Produce json
// @Param id path integer true "Supplier item ID"
// @Success 200 {object} GetSupplierProductResponseEnvelope "Supplier item retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Supplier item not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-products/{id} [get]
func (h SupplierProductHandler) Get(c fiber.Ctx) error {
	offerID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid supplier item id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	offer, err := h.svc.FindInOrg(c, offerID, organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("supplier item lookup failed", "supplier_item_id", offerID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get supplier item.", err)
	}
	if offer == nil {
		return httpx.CreateNotFoundResponse(c, "Supplier item not found.")
	}

	return httpx.CreateSuccessResponse(c, "Supplier item retrieved successfully.", GetSupplierProductResponse{
		SupplierProduct: newSupplierProductResponse(offer),
	})
}
