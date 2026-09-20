package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

type SubscriptionPlanHandler struct {
	plans subscription.SubscriptionPlanService
}

func NewSubscriptionPlanHandler(plans subscription.SubscriptionPlanService) SubscriptionPlanHandler {
	return SubscriptionPlanHandler{plans: plans}
}

type SubscriptionPlanResponse struct {
	ID                uint64  `json:"id"`
	OrganizationID    *uint64 `json:"organization_id"`
	Name              string  `json:"name"`
	RecurringInterval string  `json:"recurring_interval"`
	RecurringCount    int     `json:"recurring_count"`
}

func newSubscriptionPlanResponse(plan *reference.SubscriptionPlan) SubscriptionPlanResponse {
	return SubscriptionPlanResponse{
		ID:                plan.ID,
		OrganizationID:    plan.OrganizationID,
		Name:              plan.Name,
		RecurringInterval: plan.RecurringInterval,
		RecurringCount:    plan.RecurringCount,
	}
}

type ListSubscriptionPlansResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSubscriptionPlansResponse `json:"data"`
}
type ListSubscriptionPlansResponse struct {
	Plans []SubscriptionPlanResponse `json:"plans"`
}

type GetSubscriptionPlanResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSubscriptionPlanResponse `json:"data"`
}
type GetSubscriptionPlanResponse struct {
	Plan SubscriptionPlanResponse `json:"plan"`
}

type CreateSubscriptionPlanRequest struct {
	OrganizationID    *uint64 `json:"organization_id"`
	Name              string  `json:"name" validate:"required"`
	RecurringInterval string  `json:"recurring_interval" validate:"required"`
	RecurringCount    int     `json:"recurring_count"`
}

type CreateSubscriptionPlanResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetSubscriptionPlanResponse `json:"data"`
}

type UpdateSubscriptionPlanRequest struct {
	Name              string `json:"name" validate:"required"`
	RecurringInterval string `json:"recurring_interval" validate:"required"`
	RecurringCount    int    `json:"recurring_count"`
}
