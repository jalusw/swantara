package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type SupplierQuoteLineRequest struct {
	QuoteRequestLineID uint64  `json:"quoteRequest_line_id" validate:"required,gt=0"`
	ItemID             *uint64 `json:"item_id"`
	Description        *string `json:"description"`
	Qty                float64 `json:"qty" validate:"required,gt=0"`
	UnitPrice          float64 `json:"unit_price" validate:"required"`
	DiscountPct        float64 `json:"discount_pct"`
}

type SubmitSupplierQuoteRequest struct {
	SupplierID   uint64                     `json:"supplier_id" validate:"required,gt=0"`
	CurrencyCode *string                    `json:"currency_code"`
	ValidUntil   *time.Time                 `json:"valid_until"`
	Notes        *string                    `json:"notes"`
	Lines        []SupplierQuoteLineRequest `json:"lines" validate:"required,min=1"`
}

type SupplierQuoteResponse struct {
	ID             uint64     `json:"id"`
	QuoteRequestID uint64     `json:"quoteRequest_id"`
	SupplierID     uint64     `json:"supplier_id"`
	CurrencyCode   *string    `json:"currency_code"`
	State          string     `json:"state"`
	QuoteDate      *time.Time `json:"quote_date"`
	ValidUntil     *time.Time `json:"valid_until"`
	Notes          *string    `json:"notes"`
	AmountUntaxed  float64    `json:"amount_untaxed"`
	AmountTax      float64    `json:"amount_tax"`
	AmountTotal    float64    `json:"amount_total"`
}

type ListSupplierQuotesResponse struct {
	Quotes []SupplierQuoteResponse `json:"quotes"`
}

type ListSupplierQuotesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSupplierQuotesResponse `json:"data"`
}

type SubmitSupplierQuoteResponse struct {
	Quote SupplierQuoteResponse `json:"quote"`
}

type SubmitSupplierQuoteResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SubmitSupplierQuoteResponse `json:"data"`
}

func newSupplierQuoteResponse(quote *procurement.SupplierQuote) SupplierQuoteResponse {
	return SupplierQuoteResponse{
		ID:             quote.ID,
		QuoteRequestID: quote.QuoteRequestID,
		SupplierID:     quote.SupplierID,
		CurrencyCode:   quote.CurrencyCode,
		State:          quote.State,
		QuoteDate:      quote.QuoteDate,
		ValidUntil:     quote.ValidUntil,
		Notes:          quote.Notes,
		AmountUntaxed:  quote.AmountUntaxed,
		AmountTax:      quote.AmountTax,
		AmountTotal:    quote.AmountTotal,
	}
}

// @Summary Submit purchase quotation
// @Description Submits a priced supplier quotation for a sent purchase QuoteRequest, creating it in the submitted state. Each line must reference a known QuoteRequest line with a positive quantity and a non-negative unit price, and the untaxed and total amounts are computed from the quoted lines. Only sent QuoteRequests can receive quotations, and the submitting supplier must exist.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Param body body SubmitSupplierQuoteRequest true "Quotation details"
// @Success 201 {object} SubmitSupplierQuoteResponseEnvelope "Purchase quotation submitted successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase QuoteRequest not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id}/quotes [post]
func (h SupplierQuoteRequestHandler) SubmitQuote(c fiber.Ctx) error {
	rfqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase QuoteRequest id provided.", nil)
	}
	var request SubmitSupplierQuoteRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	lines := make([]*procurement.SupplierQuoteLine, 0, len(request.Lines))
	for _, line := range request.Lines {
		lines = append(lines, &procurement.SupplierQuoteLine{
			QuoteRequestLineID: line.QuoteRequestLineID,
			ItemID:             line.ItemID,
			Description:        line.Description,
			Qty:                line.Qty,
			UnitPrice:          line.UnitPrice,
			DiscountPct:        line.DiscountPct,
		})
	}
	quote, err := h.svc.SubmitQuote(c, rfqID, &procurement.SupplierQuote{
		SupplierID:   request.SupplierID,
		CurrencyCode: request.CurrencyCode,
		ValidUntil:   request.ValidUntil,
		Notes:        request.Notes,
	}, lines)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase quotation submitted successfully.", SubmitSupplierQuoteResponse{
		Quote: newSupplierQuoteResponse(quote),
	})
}

// @Summary List purchase quotations
// @Description Lists the supplier quotations submitted for a purchase QuoteRequest, including supplier, pricing, validity period, and computed totals. The quotations are returned for the QuoteRequest identified by the path id.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Success 200 {object} ListSupplierQuotesResponseEnvelope "Purchase quotations retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase QuoteRequest not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id}/quotes [get]
func (h SupplierQuoteRequestHandler) ListQuotes(c fiber.Ctx) error {
	rfqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase QuoteRequest id provided.", nil)
	}
	quotes, err := h.svc.ListQuotes(c, rfqID)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	items := make([]SupplierQuoteResponse, len(quotes))
	for i, quote := range quotes {
		items[i] = newSupplierQuoteResponse(quote)
	}
	return httpx.CreateSuccessResponse(c, "Purchase quotations retrieved successfully.", ListSupplierQuotesResponse{
		Quotes: items,
	})
}

// @Summary Accept purchase quotation
// @Description Accepts a submitted quotation for a sent purchase QuoteRequest, marking it as the accepted offer while rejecting every other submitted quotation. The QuoteRequest is then moved to the done state, making the accepted quotation the winning offer for conversion into a purchase order.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Param supplier_quote_id path integer true "Purchase quotation ID"
// @Success 200 {object} GetSupplierQuoteRequestResponseEnvelope "Purchase quotation accepted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase quotation not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id}/quotes/{supplier_quote_id}/accept [post]
func (h SupplierQuoteRequestHandler) AcceptQuote(c fiber.Ctx) error {
	quoteID, err := strconv.ParseUint(c.Params("supplier_quote_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase quotation id provided.", nil)
	}
	quoteRequest, err := h.svc.AcceptQuote(c, quoteID)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase quotation accepted successfully.", GetSupplierQuoteRequestResponse{
		QuoteRequest: newSupplierQuoteRequestResponse(quoteRequest),
	})
}
