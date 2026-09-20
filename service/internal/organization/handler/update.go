package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type UpdateOrganizationRequest struct {
	Name              string  `json:"name" validate:"required"`
	LegalName         *string `json:"legal_name"`
	ParentID          *uint64 `json:"parent_id"`
	BaseCurrency      string  `json:"base_currency"`
	CountryCode       *string `json:"country_code"`
	TaxID             *string `json:"tax_id"`
	Logo              *string `json:"logo"`
	Timezone          string  `json:"timezone"`
	TaxYearStartMonth int16   `json:"tax_year_start_month"`
}

type UpdateOrganizationResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateOrganizationResponse `json:"data"`
}
type UpdateOrganizationResponse struct {
	Organization OrganizationResponse `json:"organization"`
}

// @Summary Update organization
// @Description Updates an organization's details, including its legal identity, base currency, tax id, timezone, and tax year settings. The base currency must be a known currency, and assigning a parent that would create a cycle is rejected.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param id path integer true "Organization ID"
// @Param body body UpdateOrganizationRequest true "Organization details"
// @Success 200 {object} UpdateOrganizationResponseEnvelope "Organization updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "Organization not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or parent cycle"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id} [put]
func (h OrganizationHandler) Update(c fiber.Ctx) error {
	orgID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid organization id provided.", nil)
	}

	var request UpdateOrganizationRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	org, err := h.svc.Update(c, orgID, &reference.Organization{
		Name:              request.Name,
		LegalName:         request.LegalName,
		ParentID:          request.ParentID,
		BaseCurrency:      request.BaseCurrency,
		CountryCode:       request.CountryCode,
		TaxID:             request.TaxID,
		Logo:              request.Logo,
		Timezone:          request.Timezone,
		TaxYearStartMonth: request.TaxYearStartMonth,
	})
	if err != nil {
		return writeOrganizationError(c, err)
	}

	return httpx.CreateSuccessResponseWithLinks(c, "Organization updated successfully.", UpdateOrganizationResponse{
		Organization: newOrganizationResponse(org),
	}, buildOrganizationLinks(c, org.ID))
}
