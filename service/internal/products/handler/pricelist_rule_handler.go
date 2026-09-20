package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/products"
)

type PriceRuleResponse struct {
	ID          uint64     `json:"id"`
	PriceBookID uint64     `json:"price_book_id"`
	AppliesTo   string     `json:"applies_to"`
	ItemID      *uint64    `json:"item_id"`
	CategoryID  *uint64    `json:"category_id"`
	MinQty      float64    `json:"min_qty"`
	ComputeType string     `json:"compute_type"`
	FixedPrice  *float64   `json:"fixed_price"`
	DiscountPct *float64   `json:"discount_pct"`
	DateStart   *time.Time `json:"date_start"`
	DateEnd     *time.Time `json:"date_end"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func newPriceRuleResponse(rule *products.PriceRule) PriceRuleResponse {
	return PriceRuleResponse{
		ID:          rule.ID,
		PriceBookID: rule.PriceBookID,
		AppliesTo:   rule.AppliesTo,
		ItemID:      rule.ItemID,
		CategoryID:  rule.CategoryID,
		MinQty:      rule.MinQty,
		ComputeType: rule.ComputeType,
		FixedPrice:  rule.FixedPrice,
		DiscountPct: rule.DiscountPct,
		DateStart:   rule.DateStart,
		DateEnd:     rule.DateEnd,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}
}

type ListPriceRulesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPriceRulesResponse `json:"data"`
}
type ListPriceRulesResponse struct {
	Rules []PriceRuleResponse `json:"rules"`
}

// @Summary List price_book rules
// @Description Lists the pricing rules for a price list, each with its applies-to scope, compute type, minimum quantity, and validity window. A 404 is returned if the price list does not belong to the caller's organization.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param id path integer true "PriceBook ID"
// @Success 200 {object} ListPriceRulesResponseEnvelope "Rules retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "PriceBook not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books/{id}/rules [get]
func (h PriceBookHandler) ListRules(c fiber.Ctx) error {
	price_bookID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid price_book id provided.", nil)
	}

	price_book, err := h.svc.FindPriceBook(c, price_bookID)
	if err != nil {
		httpx.RequestLog(c).Error("price_book lookup failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve rules.", err)
	}
	if price_book == nil || !httpx.OwnsTenant(c, price_book.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "PriceBook not found.")
	}

	rules, err := h.svc.ListRulesByPriceBook(c, price_bookID)
	if err != nil {
		httpx.RequestLog(c).Error("price_book rule list failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve rules.", err)
	}

	items := make([]PriceRuleResponse, len(rules))
	for i, rule := range rules {
		items[i] = newPriceRuleResponse(rule)
	}

	return httpx.CreateSuccessResponse(c, "Rules retrieved successfully.", ListPriceRulesResponse{
		Rules: items,
	})
}

type CreatePriceRuleRequest struct {
	AppliesTo   string     `json:"applies_to" validate:"required"`
	ItemID      *uint64    `json:"item_id"`
	CategoryID  *uint64    `json:"category_id"`
	MinQty      float64    `json:"min_qty"`
	ComputeType string     `json:"compute_type" validate:"required"`
	FixedPrice  *float64   `json:"fixed_price"`
	DiscountPct *float64   `json:"discount_pct"`
	DateStart   *time.Time `json:"date_start"`
	DateEnd     *time.Time `json:"date_end"`
}

type CreatePriceRuleResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePriceRuleResponse `json:"data"`
}
type CreatePriceRuleResponse struct {
	Rule PriceRuleResponse `json:"rule"`
}

// @Summary Add price_book rule
// @Description Adds a pricing rule to a price list, defining a item or category scope, a fixed price or discount percentage, a minimum quantity, and an optional validity window. The applies-to and compute types must be valid and the scope must reference a item or category; a 404 is returned if the price list does not belong to the caller.
// @Tags PriceBooks
// @Accept json
// @Produce json
// @Param id path integer true "PriceBook ID"
// @Param body body CreatePriceRuleRequest true "Rule details"
// @Success 201 {object} CreatePriceRuleResponseEnvelope "Rule created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "PriceBook not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or invalid rule"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/price_books/{id}/rules [post]
func (h PriceBookHandler) CreateRule(c fiber.Ctx) error {
	price_bookID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid price_book id provided.", nil)
	}

	var request CreatePriceRuleRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	price_book, err := h.svc.FindPriceBook(c, price_bookID)
	if err != nil {
		httpx.RequestLog(c).Error("price_book lookup failed", "price_book_id", price_bookID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create rule.", err)
	}
	if price_book == nil || !httpx.OwnsTenant(c, price_book.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "PriceBook not found.")
	}

	rule, err := h.svc.CreateRule(c, price_bookID, &products.PriceRule{
		AppliesTo:   request.AppliesTo,
		ItemID:      request.ItemID,
		CategoryID:  request.CategoryID,
		MinQty:      request.MinQty,
		ComputeType: request.ComputeType,
		FixedPrice:  request.FixedPrice,
		DiscountPct: request.DiscountPct,
		DateStart:   request.DateStart,
		DateEnd:     request.DateEnd,
	})
	if err != nil {
		return writePriceBookError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Rule created successfully.", CreatePriceRuleResponse{
		Rule: newPriceRuleResponse(rule),
	})
}
