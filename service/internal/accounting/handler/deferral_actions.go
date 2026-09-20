package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type RecognizeDeferralsRequest struct {
	AsOf string `json:"as_of" validate:"required"`
}

type RecognizeDeferralsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data RecognizeDeferralsResponse `json:"data"`
}
type RecognizeDeferralsResponse struct {
	Posted int `json:"posted"`
}

// @Summary Recognize due deferrals
// @Description Posts recognition entries for all deferral lines due as of the given date belonging to the caller's organization, returning how many were recognized.
// @Tags Deferrals
// @Accept json
// @Produce json
// @Param body body RecognizeDeferralsRequest true "As-of date"
// @Success 200 {object} RecognizeDeferralsResponseEnvelope "Deferrals recognized successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/deferrals/recognize [post]
func (h DeferralHandler) Recognize(c fiber.Ctx) error {
	var request RecognizeDeferralsRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	asOf, err := time.Parse("2006-01-02", request.AsOf)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid as-of date provided.", nil)
	}

	posted, err := h.svc.RecognizeDue(c, organizationID, asOf)
	if err != nil {
		httpx.RequestLog(c).Error("deferral recognition failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to recognize due deferrals.", err)
	}

	return httpx.CreateSuccessResponse(c, "Deferrals recognized successfully.", RecognizeDeferralsResponseEnvelope{
		Data: RecognizeDeferralsResponse{Posted: posted},
	})
}
