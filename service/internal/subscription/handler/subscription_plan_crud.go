package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

// @Summary List subscription plans
// @Description Lists subscription plans with pagination, sorting, and filtering over fields such as name, recurring interval, and recurring count. The tenant filter is always enforced so callers only see their own organization's plans.
// @Tags Subscription Plans
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. recurring_interval:eq:monthly)"
// @Success 200 {object} ListSubscriptionPlansResponseEnvelope "Subscription plans retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscription-plans [get]
func (h SubscriptionPlanHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, map[string]struct{}{
		"name":               {},
		"recurring_interval": {},
		"recurring_count":    {},
		"organization_id":    {},
	})
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.plans.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("subscription plan list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve subscription plans.", err)
	}

	items := make([]SubscriptionPlanResponse, len(page.Items))
	for i, plan := range page.Items {
		items[i] = newSubscriptionPlanResponse(plan)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Subscription plans retrieved successfully.", ListSubscriptionPlansResponse{
		Plans: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get subscription plan
// @Description Gets a single subscription plan by id with its recurring interval and count. A 404 is returned when the plan does not exist or belongs to another organization.
// @Tags Subscription Plans
// @Accept json
// @Produce json
// @Param id path integer true "Subscription plan ID"
// @Success 200 {object} GetSubscriptionPlanResponseEnvelope "Subscription plan retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription plan not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscription-plans/{id} [get]
func (h SubscriptionPlanHandler) Get(c fiber.Ctx) error {
	planID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid subscription plan id provided.", nil)
	}

	plan, err := h.plans.Find(c, planID)
	if err != nil {
		httpx.RequestLog(c).Error("subscription plan lookup failed", "plan_id", planID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get subscription plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Subscription plan not found.")
	}

	return httpx.CreateSuccessResponse(c, "Subscription plan retrieved successfully.", GetSubscriptionPlanResponse{
		Plan: newSubscriptionPlanResponse(plan),
	})
}

// @Summary Create subscription plan
// @Description Creates a subscription plan with a name, recurring interval, and recurring count. The recurring count defaults to 1 when not provided.
// @Tags Subscription Plans
// @Accept json
// @Produce json
// @Param body body CreateSubscriptionPlanRequest true "Subscription plan details"
// @Success 201 {object} CreateSubscriptionPlanResponseEnvelope "Subscription plan created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscription-plans [post]
func (h SubscriptionPlanHandler) Create(c fiber.Ctx) error {
	var request CreateSubscriptionPlanRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	if request.RecurringCount == 0 {
		request.RecurringCount = 1
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	created, err := h.plans.Create(c, &reference.SubscriptionPlan{
		OrganizationID:    organizationID,
		Name:              request.Name,
		RecurringInterval: request.RecurringInterval,
		RecurringCount:    request.RecurringCount,
	})
	if err != nil {
		httpx.RequestLog(c).Error("subscription plan creation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create subscription plan.", err)
	}

	return httpx.CreateCreatedResponse(c, "Subscription plan created successfully.", CreateSubscriptionPlanResponseEnvelope{
		Data: GetSubscriptionPlanResponse{Plan: newSubscriptionPlanResponse(created)},
	})
}

// @Summary Update subscription plan
// @Description Updates an existing subscription plan's name, recurring interval, and recurring count. A 404 is returned when the plan does not exist or belongs to another organization.
// @Tags Subscription Plans
// @Accept json
// @Produce json
// @Param id path integer true "Subscription plan ID"
// @Param body body UpdateSubscriptionPlanRequest true "Subscription plan details"
// @Success 200 {object} GetSubscriptionPlanResponseEnvelope "Subscription plan updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription plan not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscription-plans/{id} [put]
func (h SubscriptionPlanHandler) Update(c fiber.Ctx) error {
	planID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid subscription plan id provided.", nil)
	}

	plan, err := h.plans.Find(c, planID)
	if err != nil {
		httpx.RequestLog(c).Error("subscription plan lookup failed", "plan_id", planID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update subscription plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Subscription plan not found.")
	}

	var request UpdateSubscriptionPlanRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	if request.RecurringCount == 0 {
		request.RecurringCount = 1
	}
	plan.Name = request.Name
	plan.RecurringInterval = request.RecurringInterval
	plan.RecurringCount = request.RecurringCount

	updated, err := h.plans.Update(c, plan)
	if err != nil {
		httpx.RequestLog(c).Error("subscription plan update failed", "plan_id", planID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update subscription plan.", err)
	}

	return httpx.CreateSuccessResponse(c, "Subscription plan updated successfully.", GetSubscriptionPlanResponse{
		Plan: newSubscriptionPlanResponse(updated),
	})
}

// @Summary Delete subscription plan
// @Description Permanently deletes a subscription plan by id. A 404 is returned when the plan does not exist or belongs to another organization.
// @Tags Subscription Plans
// @Accept json
// @Produce json
// @Param id path integer true "Subscription plan ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Subscription plan not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscription-plans/{id} [delete]
func (h SubscriptionPlanHandler) Delete(c fiber.Ctx) error {
	planID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid subscription plan id provided.", nil)
	}

	plan, err := h.plans.Find(c, planID)
	if err != nil {
		httpx.RequestLog(c).Error("subscription plan lookup failed", "plan_id", planID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete subscription plan.", err)
	}
	if plan == nil || !httpx.OwnsTenant(c, plan.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Subscription plan not found.")
	}

	if err := h.plans.Delete(c, planID); err != nil {
		httpx.RequestLog(c).Error("subscription plan deletion failed", "plan_id", planID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete subscription plan.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
