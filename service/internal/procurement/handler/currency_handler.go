package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type CurrencyRateHandler struct {
	svc procurement.CurrencyRateService
}

func NewCurrencyRateHandler(svc procurement.CurrencyRateService) CurrencyRateHandler {
	return CurrencyRateHandler{svc: svc}
}

type CurrencyRateResponse struct {
	ID           uint64     `json:"id"`
	FromCurrency string     `json:"from_currency"`
	ToCurrency   string     `json:"to_currency"`
	Rate         float64    `json:"rate"`
	RateDate     *time.Time `json:"rate_date"`
}

func newCurrencyRateResponse(rate *procurement.CurrencyRate) CurrencyRateResponse {
	return CurrencyRateResponse{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		RateDate:     rate.RateDate,
	}
}

var currencyRateQueryAllowlist = map[string]struct{}{
	"from_currency": {},
	"to_currency":   {},
	"rate_date":     {},
	"created_at":    {},
	"updated_at":    {},
}

type ListCurrencyRatesResponse struct {
	Rates []CurrencyRateResponse `json:"rates"`
}

type ListCurrencyRatesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCurrencyRatesResponse `json:"data"`
}

// @Summary List currency rates
// @Description Lists currency exchange rates across the caller's organization with pagination, sorting, and filtering on attributes such as currency pair and rate date.
// @Tags Currency Rates
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListCurrencyRatesResponseEnvelope "Currency rates retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/currency-rates [get]
func (h CurrencyRateHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, currencyRateQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("currency rate list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve currency rates.", err)
	}
	items := make([]CurrencyRateResponse, len(page.Items))
	for i, rate := range page.Items {
		items[i] = newCurrencyRateResponse(rate)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Currency rates retrieved successfully.", ListCurrencyRatesResponse{
		Rates: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetCurrencyRateResponse struct {
	Rate CurrencyRateResponse `json:"rate"`
}

type GetCurrencyRateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetCurrencyRateResponse `json:"data"`
}

// @Summary Get currency rate
// @Description Gets a single currency exchange rate by id, scoped to the caller's organization.
// @Tags Currency Rates
// @Accept json
// @Produce json
// @Param id path integer true "Currency rate ID"
// @Success 200 {object} GetCurrencyRateResponseEnvelope "Currency rate retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Currency rate not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/currency-rates/{id} [get]
func (h CurrencyRateHandler) Get(c fiber.Ctx) error {
	rateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid currency rate id provided.", nil)
	}
	rate, err := h.svc.Find(c, rateID)
	if err != nil {
		httpx.RequestLog(c).Error("currency rate lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get currency rate.", err)
	}
	if rate == nil || !httpx.OwnsTenant(c, rate.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Currency rate not found.")
	}
	return httpx.CreateSuccessResponse(c, "Currency rate retrieved successfully.", GetCurrencyRateResponse{
		Rate: newCurrencyRateResponse(rate),
	})
}

type CreateCurrencyRateRequest struct {
	FromCurrency string  `json:"from_currency" validate:"required,len=3"`
	ToCurrency   string  `json:"to_currency" validate:"required,len=3"`
	Rate         float64 `json:"rate" validate:"required,gt=0"`
	RateDate     *string `json:"rate_date"`
}

type CreateCurrencyRateResponse struct {
	Rate CurrencyRateResponse `json:"rate"`
}

type CreateCurrencyRateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCurrencyRateResponse `json:"data"`
}

// @Summary Create currency rate
// @Description Creates a new currency exchange rate for a currency pair on a given date. The rate must be positive and the currency codes must be 3-letter ISO codes.
// @Tags Currency Rates
// @Accept json
// @Produce json
// @Param body body CreateCurrencyRateRequest true "Currency rate details"
// @Success 201 {object} CreateCurrencyRateResponseEnvelope "Currency rate created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/currency-rates [post]
func (h CurrencyRateHandler) Create(c fiber.Ctx) error {
	var request CreateCurrencyRateRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	rateDate, err := helper.ParseDate(request.RateDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rate date provided.", nil)
	}
	rate := &procurement.CurrencyRate{
		OrganizationID: httpx.TenantOrganizationID(c, nil),
		FromCurrency:   request.FromCurrency,
		ToCurrency:     request.ToCurrency,
		Rate:           request.Rate,
		RateDate:       rateDate,
	}
	created, err := h.svc.Create(c, rate)
	if err != nil {
		return writeCurrencyRateError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Currency rate created successfully.", CreateCurrencyRateResponse{
		Rate: newCurrencyRateResponse(created),
	})
}

type UpdateCurrencyRateRequest struct {
	Rate float64 `json:"rate" validate:"required,gt=0"`
}

// @Summary Update currency rate
// @Description Updates the rate of an existing currency exchange rate. The rate must be positive.
// @Tags Currency Rates
// @Accept json
// @Produce json
// @Param id path integer true "Currency rate ID"
// @Param body body UpdateCurrencyRateRequest true "Rate update"
// @Success 200 {object} GetCurrencyRateResponseEnvelope "Currency rate updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Currency rate not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/currency-rates/{id} [put]
func (h CurrencyRateHandler) Update(c fiber.Ctx) error {
	rateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid currency rate id provided.", nil)
	}
	existing, err := h.svc.Find(c, rateID)
	if err != nil {
		httpx.RequestLog(c).Error("currency rate lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get currency rate.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Currency rate not found.")
	}
	var request UpdateCurrencyRateRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	existing.Rate = request.Rate
	updated, err := h.svc.Update(c, existing)
	if err != nil {
		return writeCurrencyRateError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Currency rate updated successfully.", GetCurrencyRateResponse{
		Rate: newCurrencyRateResponse(updated),
	})
}

type DeleteCurrencyRateResponse struct {
	Message string `json:"message"`
}

type DeleteCurrencyRateResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DeleteCurrencyRateResponse `json:"data"`
}

// @Summary Delete currency rate
// @Description Deletes a currency exchange rate by id. The rate must belong to the caller's organization.
// @Tags Currency Rates
// @Accept json
// @Produce json
// @Param id path integer true "Currency rate ID"
// @Success 200 {object} DeleteCurrencyRateResponseEnvelope "Currency rate deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Currency rate not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/currency-rates/{id} [delete]
func (h CurrencyRateHandler) Delete(c fiber.Ctx) error {
	rateID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid currency rate id provided.", nil)
	}
	existing, err := h.svc.Find(c, rateID)
	if err != nil {
		httpx.RequestLog(c).Error("currency rate lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get currency rate.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Currency rate not found.")
	}
	if err := h.svc.Delete(c, rateID); err != nil {
		httpx.RequestLog(c).Error("currency rate delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete currency rate.", err)
	}
	return httpx.CreateSuccessResponse(c, "Currency rate deleted successfully.", DeleteCurrencyRateResponse{
		Message: "Currency rate deleted successfully.",
	})
}

func writeCurrencyRateError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, procurement.ErrCurrencyRateExists):
		return httpx.CreateConflictResponse(c, "Currency rate already exists for this pair and date.", err)
	default:
		httpx.RequestLog(c).Error("currency rate write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process currency rate.", err)
	}
}
