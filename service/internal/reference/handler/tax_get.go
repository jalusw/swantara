package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Get tax
// @Description Returns a single tax definition by id with its rate, type, scope, and account mappings. The tax must belong to the caller's organization, otherwise a 404 is returned.
// @Tags Taxes
// @Accept json
// @Produce json
// @Param id path integer true "Tax ID"
// @Success 200 {object} GetTaxResponseEnvelope "Tax retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/taxes/{id} [get]
func (h TaxHandler) Get(c fiber.Ctx) error {
	taxID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax id provided.", nil)
	}

	tax, err := h.svc.Find(c, taxID)
	if err != nil {
		httpx.RequestLog(c).Error("tax lookup failed", "tax_id", taxID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get tax.", err)
	}
	if tax == nil || !httpx.OwnsTenant(c, tax.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Tax not found.")
	}

	return httpx.CreateSuccessResponse(c, "Tax retrieved successfully.", GetTaxResponse{
		Tax: newTaxResponse(tax),
	})
}
