package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateTaxYearRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	DateStart      string  `json:"date_start" validate:"required"`
	DateEnd        string  `json:"date_end" validate:"required"`
}

// @Summary Create tax year
// @Description Creates a tax year with a required name and a start and end date in YYYY-MM-DD format. The date range must be valid with date_end not preceding date_start, and the year is assigned to the caller's organization.
// @Tags Fiscal Years
// @Accept json
// @Produce json
// @Param body body CreateTaxYearRequest true "Tax year details"
// @Success 201 {object} CreateTaxYearResponseEnvelope "Tax year created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid date range"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-years [post]
func (h TaxYearHandler) Create(c fiber.Ctx) error {
	var request CreateTaxYearRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	dateStart, err := helper.ParseDate(&request.DateStart)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	dateEnd, err := helper.ParseDate(&request.DateEnd)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	year, err := h.svc.Create(c, &reference.TaxYear{
		OrganizationID: organizationID,
		Name:           request.Name,
		DateStart:      dateStart,
		DateEnd:        dateEnd,
	})
	if err != nil {
		return writeTaxYearError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Tax year created successfully.", CreateTaxYearResponse{
		TaxYear: newTaxYearResponse(year),
	})
}
