package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetInboundCostResponseEnvelope struct {
	httpx.EnvelopeBase
	Data InboundCostResponse `json:"data"`
}

// @Summary Get inbound cost
// @Description Gets a single inbound cost by id for the caller's organization. A 404 is returned when it does not exist or belongs to another organization.
// @Tags Landed Costs
// @Accept json
// @Produce json
// @Param id path integer true "Inbound cost ID"
// @Success 200 {object} GetInboundCostResponseEnvelope "Inbound cost retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Inbound cost not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/landed-costs/{id} [get]
func (h InboundCostHandler) Get(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	costID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inbound cost id provided.", nil)
	}

	cost, err := h.svc.Get(c, *organizationID, costID)
	if err != nil {
		return writeInboundCostError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Inbound cost retrieved successfully.", GetInboundCostResponseEnvelope{
		Data: newInboundCostResponse(cost),
	})
}
