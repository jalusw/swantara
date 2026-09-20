package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete supplier item
// @Description Deletes a supplier catalog entry by id; returns 404 if the entry does not exist or belongs to another organization and 204 No Content on success.
// @Tags Supplier Products
// @Accept json
// @Produce json
// @Param id path integer true "Supplier item ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Supplier item not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/supplier-products/{id} [delete]
func (h SupplierProductHandler) Delete(c fiber.Ctx) error {
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
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete supplier item.", err)
	}
	if offer == nil {
		return httpx.CreateNotFoundResponse(c, "Supplier item not found.")
	}

	if err := h.svc.Delete(c, offerID); err != nil {
		httpx.RequestLog(c).Error("supplier item deletion failed", "supplier_item_id", offerID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete supplier item.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
