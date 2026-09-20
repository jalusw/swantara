package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetOrganizationResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetOrganizationResponse `json:"data"`
}
type GetOrganizationResponse struct {
	Organization OrganizationResponse `json:"organization"`
}

// @Summary Get organization
// @Description Gets a single organization by id with its legal identity, base currency, tax id, timezone, and tax year settings, together with HATEOAS links for self, update, and delete.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param id path integer true "Organization ID"
// @Success 200 {object} GetOrganizationResponseEnvelope "Organization retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Organization not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id} [get]
func (h OrganizationHandler) Get(c fiber.Ctx) error {
	org, ok := h.findOrganization(c)
	if !ok {
		return nil
	}

	return httpx.CreateSuccessResponseWithLinks(c, "Organization retrieved successfully.", GetOrganizationResponse{
		Organization: newOrganizationResponse(org),
	}, buildOrganizationLinks(c, org.ID))
}
