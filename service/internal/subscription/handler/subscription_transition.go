package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

type TransitionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data SubscriptionResponse `json:"data"`
}

// @Summary Activate subscription
// @Description Transitions a subscription into the active state.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path integer true "Subscription ID"
// @Success 200 {object} TransitionResponseEnvelope "Subscription activated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription not found"
// @Failure 409 {object} httpx.ErrorResponse "Subscription state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions/{id}/activate [post]
func (h SubscriptionHandler) Activate(c fiber.Ctx) error {
	return h.transition(c, func(ctx fiber.Ctx, organizationID, subscriptionID uint64) (*subscription.Subscription, error) {
		return h.svc.Activate(ctx, organizationID, subscriptionID)
	})
}

// @Summary Pause subscription
// @Description Transitions a subscription into the paused state.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path integer true "Subscription ID"
// @Success 200 {object} TransitionResponseEnvelope "Subscription paused successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription not found"
// @Failure 409 {object} httpx.ErrorResponse "Subscription state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions/{id}/pause [post]
func (h SubscriptionHandler) Pause(c fiber.Ctx) error {
	return h.transition(c, func(ctx fiber.Ctx, organizationID, subscriptionID uint64) (*subscription.Subscription, error) {
		return h.svc.Pause(ctx, organizationID, subscriptionID)
	})
}

// @Summary Resume subscription
// @Description Transitions a subscription out of the paused state and back into active billing.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path integer true "Subscription ID"
// @Success 200 {object} TransitionResponseEnvelope "Subscription resumed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription not found"
// @Failure 409 {object} httpx.ErrorResponse "Subscription state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions/{id}/resume [post]
func (h SubscriptionHandler) Resume(c fiber.Ctx) error {
	return h.transition(c, func(ctx fiber.Ctx, organizationID, subscriptionID uint64) (*subscription.Subscription, error) {
		return h.svc.Resume(ctx, organizationID, subscriptionID)
	})
}

// @Summary Churn subscription
// @Description Marks a subscription as churned, ending recurring billing for it.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path integer true "Subscription ID"
// @Success 200 {object} TransitionResponseEnvelope "Subscription churned successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription not found"
// @Failure 409 {object} httpx.ErrorResponse "Subscription state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions/{id}/churn [post]
func (h SubscriptionHandler) Churn(c fiber.Ctx) error {
	return h.transition(c, func(ctx fiber.Ctx, organizationID, subscriptionID uint64) (*subscription.Subscription, error) {
		return h.svc.Churn(ctx, organizationID, subscriptionID)
	})
}

// @Summary Close subscription
// @Description Closes an ended subscription, finalizing its state and preventing further transitions.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Param id path integer true "Subscription ID"
// @Success 200 {object} TransitionResponseEnvelope "Subscription closed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription not found"
// @Failure 409 {object} httpx.ErrorResponse "Subscription state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions/{id}/close [post]
func (h SubscriptionHandler) Close(c fiber.Ctx) error {
	return h.transition(c, func(ctx fiber.Ctx, organizationID, subscriptionID uint64) (*subscription.Subscription, error) {
		return h.svc.Close(ctx, organizationID, subscriptionID)
	})
}

func (h SubscriptionHandler) transition(c fiber.Ctx, transition func(c fiber.Ctx, organizationID, subscriptionID uint64) (*subscription.Subscription, error)) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	subscriptionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid subscription id provided.", nil)
	}

	updated, err := transition(c, *organizationID, subscriptionID)
	if err != nil {
		return writeSubscriptionError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Subscription updated successfully.", TransitionResponseEnvelope{
		Data: newSubscriptionResponse(updated, nil),
	})
}
