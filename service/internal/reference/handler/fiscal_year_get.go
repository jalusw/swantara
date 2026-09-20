package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Get tax year
// @Description Returns a single tax year by id with its date range and state. The year must belong to the caller's organization, otherwise a 404 is returned.
// @Tags Fiscal Years
// @Accept json
// @Produce json
// @Param id path integer true "Tax year ID"
// @Success 200 {object} GetTaxYearResponseEnvelope "Tax year retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax year not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-years/{id} [get]
func (h TaxYearHandler) Get(c fiber.Ctx) error {
	yearID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax year id provided.", nil)
	}

	year, err := h.svc.Find(c, yearID)
	if err != nil {
		httpx.RequestLog(c).Error("tax year lookup failed", "tax_year_id", yearID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get tax year.", err)
	}
	if year == nil || !httpx.OwnsTenant(c, year.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Tax year not found.")
	}

	return httpx.CreateSuccessResponse(c, "Tax year retrieved successfully.", GetTaxYearResponse{
		TaxYear: newTaxYearResponse(year),
	})
}
