package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete tax
// @Description Deletes a tax definition by id. The tax must belong to the caller's organization; a 404 is returned otherwise.
// @Tags Taxes
// @Accept json
// @Produce json
// @Param id path integer true "Tax ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/taxes/{id} [delete]
func (h TaxHandler) Delete(c fiber.Ctx) error {
	taxID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax id provided.", nil)
	}

	tax, err := h.svc.Find(c, taxID)
	if err != nil {
		httpx.RequestLog(c).Error("tax lookup failed", "tax_id", taxID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete tax.", err)
	}
	if tax == nil || !httpx.OwnsTenant(c, tax.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Tax not found.")
	}

	if err := h.svc.Delete(c, taxID); err != nil {
		httpx.RequestLog(c).Error("tax deletion failed", "tax_id", taxID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete tax.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
