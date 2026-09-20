package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateOrganizationRequest struct {
	Name              string  `json:"name" validate:"required"`
	LegalName         *string `json:"legal_name"`
	ParentID          *uint64 `json:"parent_id"`
	BaseCurrency      string  `json:"base_currency" validate:"required,len=3"`
	CountryCode       *string `json:"country_code"`
	TaxID             *string `json:"tax_id"`
	Logo              *string `json:"logo"`
	Timezone          string  `json:"timezone"`
	TaxYearStartMonth int16   `json:"tax_year_start_month"`
}

type CreateOrganizationResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateOrganizationResponse `json:"data"`
}
type CreateOrganizationResponse struct {
	Organization OrganizationResponse `json:"organization"`
}

// @Summary Create organization
// @Description Creates an organization with a name, three-letter base currency, and optional parent organization, timezone, and tax year start. The base currency must be a known currency and the parent, when given, must exist; the creator is recorded as the organization owner.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param body body CreateOrganizationRequest true "Organization details"
// @Success 201 {object} CreateOrganizationResponseEnvelope "Organization created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown base currency/parent"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations [post]
func (h OrganizationHandler) Create(c fiber.Ctx) error {
	callerID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	var request CreateOrganizationRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	org, err := h.svc.Create(c, &reference.Organization{
		Name:              request.Name,
		LegalName:         request.LegalName,
		ParentID:          request.ParentID,
		BaseCurrency:      request.BaseCurrency,
		CountryCode:       request.CountryCode,
		TaxID:             request.TaxID,
		Logo:              request.Logo,
		Timezone:          request.Timezone,
		TaxYearStartMonth: request.TaxYearStartMonth,
	}, callerID)
	if err != nil {
		return writeOrganizationError(c, err)
	}

	return httpx.CreateCreatedResponseWithLinks(c, "Organization created successfully.", CreateOrganizationResponse{
		Organization: newOrganizationResponse(org),
	}, buildOrganizationLinks(c, org.ID))
}

type QuickCreateOrganizationRequest struct {
	Name        string `json:"name" validate:"required"`
	CountryCode string `json:"country_code" validate:"required,len=2"`
}

type QuickCreateOrganizationResponseEnvelope struct {
	httpx.EnvelopeBase
	Data QuickCreateOrganizationResponse `json:"data"`
}
type QuickCreateOrganizationResponse struct {
	Organization OrganizationResponse `json:"organization"`
}

// @Summary Quick create organization
// @Description Creates an organization with just a name and country code. Currency, timezone, and tax year are auto-provisioned based on the country. Base data is provisioned asynchronously in the background. The creator is assigned as the organization owner.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param body body QuickCreateOrganizationRequest true "Name and country code"
// @Success 201 {object} QuickCreateOrganizationResponseEnvelope "Organization created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unsupported country"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/quick [post]
func (h OrganizationHandler) QuickCreate(c fiber.Ctx) error {
	callerID, ok := httpx.CallerID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", httpx.ErrMissingUserInContext)
	}

	var request QuickCreateOrganizationRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	org, err := h.svc.QuickCreate(c, request.Name, request.CountryCode, callerID)
	if err != nil {
		return writeOrganizationError(c, err)
	}

	return httpx.CreateCreatedResponseWithLinks(c, "Organization created successfully.", QuickCreateOrganizationResponse{
		Organization: newOrganizationResponse(org),
	}, buildOrganizationLinks(c, org.ID))
}
