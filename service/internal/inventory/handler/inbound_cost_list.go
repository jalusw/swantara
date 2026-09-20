package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListInboundCostsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListInboundCostsResponse `json:"data"`
}

type ListInboundCostsResponse struct {
	Costs []InboundCostResponse `json:"costs"`
}

// @Summary List inbound costs
// @Description Lists the inbound costs belonging to the caller's organization.
// @Tags Landed Costs
// @Accept json
// @Produce json
// @Success 200 {object} ListInboundCostsResponseEnvelope "Inbound costs retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Organization is required"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/landed-costs [get]
func (h InboundCostHandler) List(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	costs, err := h.svc.List(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("inbound cost list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve inbound costs.", err)
	}

	items := make([]InboundCostResponse, len(costs))
	for i, cost := range costs {
		items[i] = newInboundCostResponse(cost)
	}

	return httpx.CreateSuccessResponse(c, "Inbound costs retrieved successfully.", ListInboundCostsResponse{
		Costs: items,
	})
}
