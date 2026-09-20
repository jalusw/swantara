package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListTaxPeriodsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListTaxPeriodsResponse `json:"data"`
}
type ListTaxPeriodsResponse struct {
	TaxPeriods []TaxPeriodResponse `json:"tax_periods"`
}

// @Summary List tax periods
// @Description Lists tax periods for the caller's tenant, applying pagination, sorting, and filtering over fields such as tax year, name, and state. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Fiscal Periods
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. state:eq:open)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListTaxPeriodsResponseEnvelope "Tax periods retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-periods [get]
func (h TaxPeriodHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, taxPeriodQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("tax period list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve tax periods.", err)
	}

	items := make([]TaxPeriodResponse, len(page.Items))
	for i, period := range page.Items {
		items[i] = newTaxPeriodResponse(period)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "tax-periods.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Tax periods retrieved successfully.", ListTaxPeriodsResponse{
		TaxPeriods: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetTaxPeriodResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetTaxPeriodResponse `json:"data"`
}
type GetTaxPeriodResponse struct {
	TaxPeriod TaxPeriodResponse `json:"tax_period"`
}

// @Summary Get tax period
// @Description Gets a single tax period by its id, returning 404 when the period does not exist or belongs to another tenant. The response includes the period's start and end dates and its current state.
// @Tags Fiscal Periods
// @Accept json
// @Produce json
// @Param id path integer true "Tax period ID"
// @Success 200 {object} GetTaxPeriodResponseEnvelope "Tax period retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-periods/{id} [get]
func (h TaxPeriodHandler) Get(c fiber.Ctx) error {
	periodID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax period id provided.", nil)
	}

	period, err := h.svc.Find(c, periodID)
	if err != nil {
		httpx.RequestLog(c).Error("tax period lookup failed", "tax_period_id", periodID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get tax period.", err)
	}
	if period == nil || !httpx.OwnsTenant(c, &period.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Tax period not found.")
	}

	return httpx.CreateSuccessResponse(c, "Tax period retrieved successfully.", GetTaxPeriodResponse{
		TaxPeriod: newTaxPeriodResponse(period),
	})
}

type CreateTaxPeriodRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	TaxYearID      uint64  `json:"tax_year_id" validate:"required,gt=0"`
	Name           string  `json:"name" validate:"required"`
	DateStart      string  `json:"date_start" validate:"required"`
	DateEnd        string  `json:"date_end" validate:"required"`
	State          string  `json:"state" validate:"omitempty,oneof=open closed locked"`
	PeriodType     string  `json:"period_type" validate:"omitempty,oneof=standard adjustment"`
}

type CreateTaxPeriodResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateTaxPeriodResponse `json:"data"`
}
type CreateTaxPeriodResponse struct {
	TaxPeriod TaxPeriodResponse `json:"tax_period"`
}

// @Summary Create tax period
// @Description Creates a tax period under a tax year, validating that the tax year belongs to the organization and that the supplied state and start-to-end date range are valid. The new period is returned with its name, dates, and initial state.
// @Tags Fiscal Periods
// @Accept json
// @Produce json
// @Param body body CreateTaxPeriodRequest true "Tax period details"
// @Success 201 {object} CreateTaxPeriodResponseEnvelope "Tax period created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid date range"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-periods [post]
func (h TaxPeriodHandler) Create(c fiber.Ctx) error {
	var request CreateTaxPeriodRequest
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

	period, err := h.svc.Create(c, &accounting.TaxPeriod{
		OrganizationID: *organizationID,
		TaxYearID:      request.TaxYearID,
		Name:           request.Name,
		DateStart:      dateStart,
		DateEnd:        dateEnd,
		State:          request.State,
		PeriodType:     request.PeriodType,
	})
	if err != nil {
		return writeTaxPeriodError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Tax period created successfully.", CreateTaxPeriodResponse{
		TaxPeriod: newTaxPeriodResponse(period),
	})
}
