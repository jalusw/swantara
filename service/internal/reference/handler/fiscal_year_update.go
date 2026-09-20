package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateTaxYearRequest struct {
	Name      string  `json:"name" validate:"required"`
	DateStart string  `json:"date_start" validate:"required"`
	DateEnd   string  `json:"date_end" validate:"required"`
	State     *string `json:"state"`
}

// @Summary Update tax year
// @Description Updates a tax year's name, date range, and state. Dates must be in YYYY-MM-DD format and date_end must not precede date_start; a 404 is returned if the year does not belong to the caller.
// @Tags Fiscal Years
// @Accept json
// @Produce json
// @Param id path integer true "Tax year ID"
// @Param body body UpdateTaxYearRequest true "Tax year details"
// @Success 200 {object} UpdateTaxYearResponseEnvelope "Tax year updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax year not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid date range"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-years/{id} [put]
func (h TaxYearHandler) Update(c fiber.Ctx) error {
	yearID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax year id provided.", nil)
	}

	var request UpdateTaxYearRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	year, err := h.svc.Find(c, yearID)
	if err != nil {
		httpx.RequestLog(c).Error("tax year lookup failed", "tax_year_id", yearID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update tax year.", err)
	}
	if year == nil || !httpx.OwnsTenant(c, year.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Tax year not found.")
	}

	dateStart, err := helper.ParseDate(&request.DateStart)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	dateEnd, err := helper.ParseDate(&request.DateEnd)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	year.Name = request.Name
	year.DateStart = dateStart
	year.DateEnd = dateEnd
	year.State = request.State

	updated, err := h.svc.Update(c, year)
	if err != nil {
		return writeTaxYearError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Tax year updated successfully.", UpdateTaxYearResponse{
		TaxYear: newTaxYearResponse(updated),
	})
}
