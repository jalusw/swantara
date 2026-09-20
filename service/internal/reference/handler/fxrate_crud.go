package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type GetFxRateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetFxRateResponse `json:"data"`
}
type GetFxRateResponse struct {
	FxRate FxRateResponse `json:"fx_rate"`
}

// @Summary Get FX rate
// @Description Returns a single FX rate by id. The rate must belong to the caller's organization, otherwise a 404 is returned; a non-numeric id yields a 422.
// @Tags FX Rates
// @Accept json
// @Produce json
// @Param id path integer true "FX rate ID"
// @Success 200 {object} GetFxRateResponseEnvelope "FX rate retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "FX rate not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/fx-rates/{id} [get]
func (h FxRateHandler) Get(c fiber.Ctx) error {
	rateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid fx rate id provided.", nil)
	}

	rate, err := h.fxRateSvc.Find(c, rateID)
	if err != nil {
		httpx.RequestLog(c).Error("fx rate lookup failed", "rate_id", rateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get FX rate.", err)
	}
	if rate == nil || !httpx.OwnsTenant(c, rate.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "FX rate not found.")
	}

	return httpx.CreateSuccessResponse(c, "FX rate retrieved successfully.", GetFxRateResponse{
		FxRate: newFxRateResponse(rate),
	})
}

type CreateFxRateRequest struct {
	CurrencyCode   string  `json:"currency_code" validate:"required,len=3"`
	OrganizationID *uint64 `json:"organization_id"`
	Rate           float64 `json:"rate" validate:"required,gt=0"`
	RateType       string  `json:"rate_type"`
	ValidFrom      string  `json:"valid_from" validate:"required"`
}

type CreateFxRateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateFxRateResponse `json:"data"`
}
type CreateFxRateResponse struct {
	FxRate FxRateResponse `json:"fx_rate"`
}

// @Summary Create FX rate
// @Description Creates an FX rate for a currency with a positive rate, a rate type of spot, average, or closing, and a valid-from date in YYYY-MM-DD format. The currency must exist, the rate must be greater than zero, and the rate type must be valid; the created rate is associated with the caller's organization.
// @Tags FX Rates
// @Accept json
// @Produce json
// @Param body body CreateFxRateRequest true "FX rate details"
// @Success 201 {object} CreateFxRateResponseEnvelope "FX rate created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, unknown currency, or invalid rate"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/fx-rates [post]
func (h FxRateHandler) Create(c fiber.Ctx) error {
	var request CreateFxRateRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	validFrom, err := helper.ParseDate(&request.ValidFrom)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Valid from must be a date in YYYY-MM-DD format.", nil)
	}

	rate, err := h.fxRateSvc.Create(c, &reference.FxRate{
		CurrencyCode:   request.CurrencyCode,
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		Rate:           request.Rate,
		RateType:       request.RateType,
		ValidFrom:      *validFrom,
	})
	if err != nil {
		return writeFxRateError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "FX rate created successfully.", CreateFxRateResponse{
		FxRate: newFxRateResponse(rate),
	})
}

type UpdateFxRateRequest struct {
	Rate      float64 `json:"rate" validate:"required,gt=0"`
	RateType  string  `json:"rate_type"`
	ValidFrom string  `json:"valid_from" validate:"required"`
}

type UpdateFxRateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateFxRateResponse `json:"data"`
}
type UpdateFxRateResponse struct {
	FxRate FxRateResponse `json:"fx_rate"`
}

// @Summary Update FX rate
// @Description Updates an FX rate's rate, rate type, and valid-from date, requiring a positive rate, a valid rate type (spot, average, or closing), and a date in YYYY-MM-DD format. The rate must belong to the caller's organization and the currency must exist, otherwise an error is returned.
// @Tags FX Rates
// @Accept json
// @Produce json
// @Param id path integer true "FX rate ID"
// @Param body body UpdateFxRateRequest true "FX rate details"
// @Success 200 {object} UpdateFxRateResponseEnvelope "FX rate updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "FX rate not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid rate"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/fx-rates/{id} [put]
func (h FxRateHandler) Update(c fiber.Ctx) error {
	rateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid fx rate id provided.", nil)
	}

	var request UpdateFxRateRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	validFrom, err := helper.ParseDate(&request.ValidFrom)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Valid from must be a date in YYYY-MM-DD format.", nil)
	}

	existing, err := h.fxRateSvc.Find(c, rateID)
	if err != nil {
		httpx.RequestLog(c).Error("fx rate lookup failed", "rate_id", rateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update FX rate.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "FX rate not found.")
	}

	existing.Rate = request.Rate
	existing.RateType = request.RateType
	existing.ValidFrom = *validFrom
	existing.OrganizationID = httpx.TenantOrganizationID(c, existing.OrganizationID)

	updated, err := h.fxRateSvc.Update(c, existing)
	if err != nil {
		return writeFxRateError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "FX rate updated successfully.", UpdateFxRateResponse{
		FxRate: newFxRateResponse(updated),
	})
}

// @Summary Delete FX rate
// @Description Deletes an FX rate by id. The rate must belong to the caller's organization; a 404 is returned otherwise.
// @Tags FX Rates
// @Accept json
// @Produce json
// @Param id path integer true "FX rate ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "FX rate not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/fx-rates/{id} [delete]
func (h FxRateHandler) Delete(c fiber.Ctx) error {
	rateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid fx rate id provided.", nil)
	}

	existing, err := h.fxRateSvc.Find(c, rateID)
	if err != nil {
		httpx.RequestLog(c).Error("fx rate lookup failed", "rate_id", rateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete FX rate.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "FX rate not found.")
	}

	if err := h.fxRateSvc.Delete(c, rateID); err != nil {
		httpx.RequestLog(c).Error("fx rate deletion failed", "rate_id", rateID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete FX rate.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
