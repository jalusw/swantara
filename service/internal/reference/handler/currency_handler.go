package handler

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CurrencyHandler struct {
	svc reference.FxRateService
}

func NewCurrencyHandler(svc reference.FxRateService) CurrencyHandler {
	return CurrencyHandler{svc: svc}
}

type CurrencyResponse struct {
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Symbol        *string   `json:"symbol"`
	DecimalPlaces int16     `json:"decimal_places"`
	Rounding      float64   `json:"rounding"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func newCurrencyResponse(currency *reference.Currency) CurrencyResponse {
	return CurrencyResponse{
		Code:          currency.Code,
		Name:          currency.Name,
		Symbol:        currency.Symbol,
		DecimalPlaces: currency.DecimalPlaces,
		Rounding:      currency.Rounding,
		CreatedAt:     currency.CreatedAt,
		UpdatedAt:     currency.UpdatedAt,
	}
}

var currencyQueryAllowlist = map[string]struct{}{
	"code":       {},
	"name":       {},
	"created_at": {},
	"updated_at": {},
}

type ListCurrenciesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCurrenciesResponse `json:"data"`
}
type ListCurrenciesResponse struct {
	Currencies []CurrencyResponse `json:"currencies"`
}

// @Summary List currencies
// @Description Lists available currencies with pagination, sorting, and filtering, returning results as JSON, XML, or CSV.
// @Tags Currencies
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. code:asc)"
// @Param filter query string false "Filters (repeatable, e.g. code:eq:USD)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListCurrenciesResponseEnvelope "Currencies retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /currencies [get]
func (h CurrencyHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, currencyQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	page, err := h.svc.ListCurrencies(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("currency list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve currencies.", err)
	}

	items := make([]CurrencyResponse, len(page.Items))
	for i, currency := range page.Items {
		items[i] = newCurrencyResponse(currency)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "currencies.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Currencies retrieved successfully.", ListCurrenciesResponse{
		Currencies: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetCurrencyResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetCurrencyResponse `json:"data"`
}
type GetCurrencyResponse struct {
	Currency CurrencyResponse `json:"currency"`
}

// @Summary Get currency
// @Description Returns a single currency by its ISO 4217 code (case-insensitive); a 404 is returned if the currency does not exist.
// @Tags Currencies
// @Accept json
// @Produce json
// @Param code path string true "ISO 4217 currency code"
// @Success 200 {object} GetCurrencyResponseEnvelope "Currency retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Currency not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /currencies/{code} [get]
func (h CurrencyHandler) Get(c fiber.Ctx) error {
	currency, err := h.svc.SearchCurrency(c, "code", strings.ToUpper(c.Params("code")))
	if err != nil {
		httpx.RequestLog(c).Error("currency lookup failed", "code", c.Params("code"), "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get currency.", err)
	}
	if currency == nil {
		return httpx.CreateNotFoundResponse(c, "Currency not found.")
	}

	return httpx.CreateSuccessResponse(c, "Currency retrieved successfully.", GetCurrencyResponse{
		Currency: newCurrencyResponse(currency),
	})
}
