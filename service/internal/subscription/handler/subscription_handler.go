package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

type SubscriptionHandler struct {
	svc subscription.SubscriptionService
}

func NewSubscriptionHandler(svc subscription.SubscriptionService) SubscriptionHandler {
	return SubscriptionHandler{svc: svc}
}

type SubscriptionLineResponse struct {
	ID          uint64  `json:"id"`
	ItemID      *uint64 `json:"item_id"`
	Qty         float64 `json:"qty"`
	UnitPrice   float64 `json:"unit_price"`
	DiscountPct float64 `json:"discount_pct"`
}

type SubscriptionResponse struct {
	ID              uint64                     `json:"id"`
	OrganizationID  *uint64                    `json:"organization_id"`
	Name            string                     `json:"name"`
	ContactID       *uint64                    `json:"contact_id"`
	PlanID          *uint64                    `json:"plan_id"`
	PriceBookID     *uint64                    `json:"price_book_id"`
	CurrencyCode    *string                    `json:"currency_code"`
	DateStart       *time.Time                 `json:"date_start"`
	NextInvoiceDate *time.Time                 `json:"next_invoice_date"`
	DateEnd         *time.Time                 `json:"date_end"`
	State           string                     `json:"state"`
	MRR             float64                    `json:"mrr"`
	Lines           []SubscriptionLineResponse `json:"lines,omitempty"`
}

func newSubscriptionResponse(s *subscription.Subscription, lines []*subscription.SubscriptionLine) SubscriptionResponse {
	response := SubscriptionResponse{
		ID:              s.ID,
		OrganizationID:  s.OrganizationID,
		Name:            s.Name,
		ContactID:       s.ContactID,
		PlanID:          s.PlanID,
		PriceBookID:     s.PriceBookID,
		CurrencyCode:    s.CurrencyCode,
		DateStart:       s.DateStart,
		NextInvoiceDate: s.NextInvoiceDate,
		DateEnd:         s.DateEnd,
		State:           s.State,
		MRR:             s.MRR,
	}
	for _, line := range lines {
		response.Lines = append(response.Lines, SubscriptionLineResponse{
			ID:          line.ID,
			ItemID:      line.ItemID,
			Qty:         line.Qty,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
		})
	}
	return response
}

func writeSubscriptionError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, subscription.ErrSubscriptionNotFound):
		return httpx.CreateNotFoundResponse(c, "Subscription not found.")
	case errors.Is(err, subscription.ErrSubscriptionState):
		return httpx.CreateConflictResponse(c, "Subscription state is invalid for this operation.", err)
	case errors.Is(err, subscription.ErrSubscriptionNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription must have at least one line.", nil)
	case errors.Is(err, subscription.ErrSubscriptionNoPlan):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription plan is required.", nil)
	case errors.Is(err, subscription.ErrSubscriptionPlanNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription plan does not exist.", nil)
	case errors.Is(err, subscription.ErrSubscriptionPlanOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription plan does not belong to this organization.", nil)
	case errors.Is(err, subscription.ErrSubscriptionLineProduct):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription line item is required.", nil)
	case errors.Is(err, subscription.ErrSubscriptionLineProductOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription line item does not belong to this organization.", nil)
	case errors.Is(err, subscription.ErrSubscriptionLineQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription line quantity must be greater than zero.", nil)
	case errors.Is(err, subscription.ErrSubscriptionLinePrice):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription line unit price must not be negative.", nil)
	case errors.Is(err, subscription.ErrSubscriptionLineDiscount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription line discount must be between 0 and 100.", nil)
	case errors.Is(err, subscription.ErrSubscriptionContact):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription contact is required.", nil)
	case errors.Is(err, subscription.ErrSubscriptionContactOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription contact does not belong to this organization.", nil)
	case errors.Is(err, subscription.ErrSubscriptionCurrency):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription currency is required.", nil)
	case errors.Is(err, subscription.ErrSubscriptionPriceBook):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription price_book is required.", nil)
	case errors.Is(err, subscription.ErrSubscriptionPriceBookOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subscription price_book does not belong to this organization.", nil)
	default:
		httpx.RequestLog(c).Error("subscription write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save subscription.", err)
	}
}
