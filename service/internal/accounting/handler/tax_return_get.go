package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetTaxReturnResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetTaxReturnResponse `json:"data"`
}
type GetTaxReturnResponse struct {
	TaxReturn TaxReturnResponse `json:"tax_return"`
}

// @Summary Get tax return
// @Description Gets a single tax return by its id, returning 404 when the return does not exist or belongs to another tenant. The response includes the computed output tax, input tax, net payable, and current state.
// @Tags Tax Returns
// @Accept json
// @Produce json
// @Param id path integer true "Tax return ID"
// @Success 200 {object} GetTaxReturnResponseEnvelope "Tax return retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax return not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-returns/{id} [get]
func (h TaxReturnHandler) Get(c fiber.Ctx) error {
	taxReturnID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax return id provided.", nil)
	}

	taxReturn, err := h.svc.Find(c, taxReturnID)
	if err != nil {
		httpx.RequestLog(c).Error("tax return lookup failed", "tax_return_id", taxReturnID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get tax return.", err)
	}
	if taxReturn == nil || !httpx.OwnsTenant(c, taxReturn.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Tax return not found.")
	}

	return httpx.CreateSuccessResponse(c, "Tax return retrieved successfully.", GetTaxReturnResponse{
		TaxReturn: newTaxReturnResponse(taxReturn),
	})
}
