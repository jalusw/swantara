package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

type ListSubscriptionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSubscriptionsResponse `json:"data"`
}
type ListSubscriptionsResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions"`
}

// @Summary List subscriptions
// @Description Lists subscriptions for the caller's organization.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} ListSubscriptionsResponseEnvelope "Subscriptions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Organization is required"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions [get]
func (h SubscriptionHandler) List(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	subscriptions, err := h.svc.List(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("subscription list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve subscriptions.", err)
	}

	items := make([]SubscriptionResponse, len(subscriptions))
	for i, subscription := range subscriptions {
		items[i] = newSubscriptionResponse(subscription, nil)
	}

	return httpx.CreateSuccessResponse(c, "Subscriptions retrieved successfully.", ListSubscriptionsResponse{
		Subscriptions: items,
	})
}

type GetSubscriptionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SubscriptionResponse `json:"data"`
}

// @Summary Get subscription
// @Description Gets a single subscription by id for the caller's organization. A 404 is returned when the subscription does not exist or belongs to another organization.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path integer true "Subscription ID"
// @Success 200 {object} GetSubscriptionResponseEnvelope "Subscription retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions/{id} [get]
func (h SubscriptionHandler) Get(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	subscriptionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid subscription id provided.", nil)
	}

	subscription, err := h.svc.Get(c, *organizationID, subscriptionID)
	if err != nil {
		return writeSubscriptionError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Subscription retrieved successfully.", GetSubscriptionResponseEnvelope{
		Data: newSubscriptionResponse(subscription, nil),
	})
}

type CreateSubscriptionRequest struct {
	Name         string                          `json:"name" validate:"required"`
	ContactID    uint64                          `json:"contact_id" validate:"required"`
	PlanID       uint64                          `json:"plan_id" validate:"required"`
	PriceBookID  uint64                          `json:"price_book_id" validate:"required"`
	CurrencyCode string                          `json:"currency_code" validate:"required"`
	Lines        []CreateSubscriptionLineRequest `json:"lines" validate:"required"`
}

type CreateSubscriptionLineRequest struct {
	ItemID      uint64  `json:"item_id" validate:"required"`
	Qty         float64 `json:"qty" validate:"required"`
	UnitPrice   float64 `json:"unit_price"`
	DiscountPct float64 `json:"discount_pct"`
}

type CreateSubscriptionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SubscriptionResponse `json:"data"`
}

// @Summary Create subscription
// @Description Creates a subscription for the caller's organization with a name, contact, plan, price_book, currency, and at least one line.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param body body CreateSubscriptionRequest true "Subscription details"
// @Success 201 {object} CreateSubscriptionResponseEnvelope "Subscription created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown references"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions [post]
func (h SubscriptionHandler) Create(c fiber.Ctx) error {
	var request CreateSubscriptionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	lines := make([]subscription.CreateSubscriptionLineRequest, 0, len(request.Lines))
	for _, line := range request.Lines {
		lines = append(lines, subscription.CreateSubscriptionLineRequest{
			ItemID:      line.ItemID,
			Qty:         line.Qty,
			UnitPrice:   line.UnitPrice,
			DiscountPct: line.DiscountPct,
		})
	}

	created, err := h.svc.Create(c, subscription.CreateSubscriptionRequest{
		OrganizationID: *organizationID,
		Name:           request.Name,
		ContactID:      request.ContactID,
		PlanID:         request.PlanID,
		PriceBookID:    request.PriceBookID,
		CurrencyCode:   request.CurrencyCode,
		Lines:          lines,
	})
	if err != nil {
		return writeSubscriptionError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Subscription created successfully.", CreateSubscriptionResponseEnvelope{
		Data: newSubscriptionResponse(created, nil),
	})
}
