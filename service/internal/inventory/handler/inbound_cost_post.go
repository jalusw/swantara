package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type PostInboundCostResponseEnvelope struct {
	httpx.EnvelopeBase
	Data InboundCostResponse `json:"data"`
}

// @Summary Post inbound cost
// @Description Posts a draft inbound cost: splits each line across the received movements of its item, increases the corresponding valuation layers, and posts Dr Inventory / Cr Landed Cost Clearing.
// @Tags Landed Costs
// @Accept json
// @Produce json
// @Param id path integer true "Inbound cost ID"
// @Success 200 {object} PostInboundCostResponseEnvelope "Inbound cost posted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Inbound cost not found"
// @Failure 409 {object} httpx.ErrorResponse "Inbound cost state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/landed-costs/{id}/post [post]
func (h InboundCostHandler) Post(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	costID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inbound cost id provided.", nil)
	}

	updated, err := h.svc.Post(c, *organizationID, costID)
	if err != nil {
		return writeInboundCostError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Inbound cost posted successfully.", PostInboundCostResponseEnvelope{
		Data: newInboundCostResponse(updated),
	})
}
