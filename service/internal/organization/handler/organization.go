package handler

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/organization"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type OrganizationHandler struct {
	svc organization.Service
}

func NewOrganizationHandler(svc organization.Service) OrganizationHandler {
	return OrganizationHandler{svc: svc}
}

type OrganizationResponse struct {
	ID                uint64    `json:"id"`
	Name              string    `json:"name"`
	LegalName         *string   `json:"legal_name"`
	ParentID          *uint64   `json:"parent_id"`
	BaseCurrency      string    `json:"base_currency"`
	CountryCode       *string   `json:"country_code"`
	TaxID             *string   `json:"tax_id"`
	Logo              *string   `json:"logo"`
	Timezone          string    `json:"timezone"`
	TaxYearStartMonth int16     `json:"tax_year_start_month"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func newOrganizationResponse(org *reference.Organization) OrganizationResponse {
	return OrganizationResponse{
		ID:                org.ID,
		Name:              org.Name,
		LegalName:         org.LegalName,
		ParentID:          org.ParentID,
		BaseCurrency:      org.BaseCurrency,
		CountryCode:       org.CountryCode,
		TaxID:             org.TaxID,
		Logo:              org.Logo,
		Timezone:          org.Timezone,
		TaxYearStartMonth: org.TaxYearStartMonth,
		CreatedAt:         org.CreatedAt,
		UpdatedAt:         org.UpdatedAt,
	}
}

var organizationQueryAllowlist = map[string]struct{}{
	"name":          {},
	"legal_name":    {},
	"base_currency": {},
	"country_code":  {},
	"logo":          {},
	"timezone":      {},
	"created_at":    {},
	"updated_at":    {},
}

func writeOrganizationError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, organization.ErrOrganizationNotFound):
		return httpx.CreateNotFoundResponse(c, "Organization not found.")
	case errors.Is(err, organization.ErrNameRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization name is required.", nil)
	case errors.Is(err, organization.ErrBaseCurrencyNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Base currency does not exist.", nil)
	case errors.Is(err, organization.ErrParentNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Parent organization does not exist.", nil)
	case errors.Is(err, organization.ErrParentCycle):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Parent organization would create a cycle.", nil)
	case errors.Is(err, organization.ErrCountryCodeInvalid):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unsupported country code.", nil)
	default:
		httpx.RequestLog(c).Error("organization write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save organization.", err)
	}
}

func (h OrganizationHandler) findOrganization(c fiber.Ctx) (*reference.Organization, bool) {
	orgID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		_ = httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid organization id provided.", nil)
		return nil, false
	}

	org, err := h.svc.Find(c, orgID)
	if err != nil {
		httpx.RequestLog(c).Error("organization lookup failed", "organization_id", orgID, "error", err)
		_ = httpx.CreateInternalServerErrorResponse(c, "Failed to get organization.", err)
		return nil, false
	}
	if org == nil {
		_ = httpx.CreateNotFoundResponse(c, "Organization not found.")
		return nil, false
	}
	return org, true
}

func buildOrganizationLinks(c fiber.Ctx, orgID uint64) []httpx.Link {
	base := fmt.Sprintf("%s/organizations/%d", httpx.LinkBaseFor(c), orgID)
	return []httpx.Link{
		{Rel: "self", Method: fiber.MethodGet, Href: base},
		{Rel: "update", Method: fiber.MethodPut, Href: base},
		{Rel: "delete", Method: fiber.MethodDelete, Href: base},
	}
}
