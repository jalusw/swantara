package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ResolveFxRateRequest struct {
	CurrencyCode   string `json:"currency_code" validate:"required,len=3"`
	OrganizationID uint64 `json:"organization_id"`
	RateType       string `json:"rate_type"`
	Date           string `json:"date" validate:"required"`
}

type ResolveFxRateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ResolveFxRateResponse `json:"data"`
}
type ResolveFxRateResponse struct {
	Rate amount.Amount `json:"rate"`
}

// @Summary Resolve FX rate
// @Description Resolves the applicable FX rate for a currency, rate type, and date in YYYY-MM-DD format. The caller's organization is used when present, and a 404 is returned when no rate applies for the criteria; an invalid rate type yields a 422.
// @Tags FX Rates
// @Accept json
// @Produce json
// @Param body body ResolveFxRateRequest true "Resolution criteria"
// @Success 200 {object} ResolveFxRateResponseEnvelope "FX rate resolved successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "FX rate not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/fx-rates/resolve [post]
func (h FxRateHandler) Resolve(c fiber.Ctx) error {
	var request ResolveFxRateRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := helper.ParseDate(&request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	organizationID := request.OrganizationID
	if callerOrg, ok := httpx.CallerOrganizationID(c); ok {
		organizationID = callerOrg
	}

	rate, err := h.rateSource.Rate(c, request.CurrencyCode, organizationID, amount.RateType(request.RateType), *date)
	if err != nil {
		if errors.Is(err, reference.ErrRateNotFound) {
			return httpx.CreateNotFoundResponse(c, "FX rate not found.")
		}
		if errors.Is(err, amount.ErrInvalidRate) {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid FX rate type.", nil)
		}
		httpx.RequestLog(c).Error("fx rate resolution failed", "currency_code", request.CurrencyCode, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to resolve FX rate.", err)
	}

	return httpx.CreateSuccessResponse(c, "FX rate resolved successfully.", ResolveFxRateResponse{
		Rate: rate,
	})
}
