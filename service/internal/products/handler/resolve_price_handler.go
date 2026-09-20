package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type ResolvePriceRequest struct {
	VariantID uint64 `json:"variant_id" validate:"required,gt=0"`
	Qty       string `json:"qty" validate:"required"`
	Date      string `json:"date"`
}

type ResolvePriceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ResolvePriceResponse `json:"data"`
}
type ResolvePriceResponse struct {
	Price       amount.Amount `json:"price"`
	BasePrice   amount.Amount `json:"base_price"`
	ItemID      uint64        `json:"item_id"`
	VariantID   uint64        `json:"variant_id"`
	RuleID      uint64        `json:"rule_id,omitempty"`
	AppliesTo   string        `json:"applies_to,omitempty"`
	ComputeType string        `json:"compute_type,omitempty"`
}

// @Summary Resolve price for a variant
// @Description Resolves the unit price for a item variant on a price list for a given quantity and date. The quantity must be a valid decimal and the date must be in YYYY-MM-DD format (defaulting to today when omitted); the matching rule's precedence and validity window are honored. A 404 is returned if the price list or variant is not found.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param id path integer true "PriceBook ID"
// @Param body body ResolvePriceRequest true "Resolution details"
// @Success 200 {object} ResolvePriceResponseEnvelope "Price resolved successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "PriceBook or variant not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books/{id}/resolve [post]
func (h PriceBookHandler) ResolvePrice(c fiber.Ctx) error {
	price_bookID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid price_book id provided.", nil)
	}

	var request ResolvePriceRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	price_book, err := h.svc.FindPriceBook(c, price_bookID)
	if err != nil {
		httpx.RequestLog(c).Error("price_book lookup failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to resolve price.", err)
	}
	if price_book == nil || !httpx.OwnsTenant(c, price_book.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "PriceBook not found.")
	}

	qty, err := amount.FromString(request.Qty)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Quantity must be a valid decimal.", nil)
	}

	date := time.Now()
	if request.Date != "" {
		parsed, err := helper.ParseDate(&request.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
		}
		date = *parsed
	}

	resolved, err := h.svc.ResolvePrice(c, price_bookID, request.VariantID, qty, date)
	if err != nil {
		return writePriceBookError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Price resolved successfully.", ResolvePriceResponse{
		Price:       resolved.Price,
		BasePrice:   resolved.BasePrice,
		ItemID:      resolved.ItemID,
		VariantID:   resolved.VariantID,
		RuleID:      resolved.RuleID,
		AppliesTo:   resolved.AppliesTo,
		ComputeType: resolved.ComputeType,
	})
}
