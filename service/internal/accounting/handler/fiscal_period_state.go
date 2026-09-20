package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateTaxPeriodResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateTaxPeriodResponse `json:"data"`
}
type UpdateTaxPeriodResponse struct {
	TaxPeriod TaxPeriodResponse `json:"tax_period"`
}

// @Summary Close tax period
// @Description Closes an open tax period, rejecting the transition when the period is locked. Once closed, the period rejects new journal entries until it is reopened or locked.
// @Tags Fiscal Periods
// @Accept json
// @Produce json
// @Param id path integer true "Tax period ID"
// @Success 200 {object} UpdateTaxPeriodResponseEnvelope "Tax period closed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 422 {object} httpx.ErrorResponse "Period is locked"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-periods/{id}/close [post]
func (h TaxPeriodHandler) Close(c fiber.Ctx) error {
	periodID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax period id provided.", nil)
	}

	period, err := h.svc.Close(c, periodID)
	if err != nil {
		return writeTaxPeriodError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Tax period closed successfully.", UpdateTaxPeriodResponse{
		TaxPeriod: newTaxPeriodResponse(period),
	})
}

// @Summary Lock tax period
// @Description Locks a tax period to make it final, requiring the period to be closed first. Locked periods reject new journal entries and any further state changes, acting as a permanent close.
// @Tags Fiscal Periods
// @Accept json
// @Produce json
// @Param id path integer true "Tax period ID"
// @Success 200 {object} UpdateTaxPeriodResponseEnvelope "Tax period locked successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 422 {object} httpx.ErrorResponse "Period must be closed first"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-periods/{id}/lock [post]
func (h TaxPeriodHandler) Lock(c fiber.Ctx) error {
	periodID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax period id provided.", nil)
	}

	period, err := h.svc.Lock(c, periodID)
	if err != nil {
		return writeTaxPeriodError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Tax period locked successfully.", UpdateTaxPeriodResponse{
		TaxPeriod: newTaxPeriodResponse(period),
	})
}

// @Summary Reopen tax period
// @Description Reopens a closed tax period back to open so that journal entries can be posted again. The transition is rejected when the period is locked.
// @Tags Fiscal Periods
// @Accept json
// @Produce json
// @Param id path integer true "Tax period ID"
// @Success 200 {object} UpdateTaxPeriodResponseEnvelope "Tax period reopened successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 422 {object} httpx.ErrorResponse "Period is locked"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-periods/{id}/open [post]
func (h TaxPeriodHandler) Open(c fiber.Ctx) error {
	periodID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax period id provided.", nil)
	}

	period, err := h.svc.Open(c, periodID)
	if err != nil {
		return writeTaxPeriodError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Tax period reopened successfully.", UpdateTaxPeriodResponse{
		TaxPeriod: newTaxPeriodResponse(period),
	})
}
