package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

type MetricsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data subscription.Metrics `json:"data"`
}

// @Summary Get subscription metrics
// @Description Computes MRR and related metrics for the caller's organization.
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} MetricsResponseEnvelope "Subscription metrics retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Organization is required"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subscriptions/metrics [get]
func (h SubscriptionHandler) Metrics(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	metrics, err := h.svc.Metrics(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("subscription metrics failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to compute subscription metrics.", err)
	}

	return httpx.CreateSuccessResponse(c, "Subscription metrics retrieved successfully.", MetricsResponseEnvelope{
		Data: metrics,
	})
}
