package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type MarkActivityDoneResponseEnvelope struct {
	httpx.EnvelopeBase
	Data MarkActivityDoneResponse `json:"data"`
}

type MarkActivityDoneResponse struct {
	Activity ProspectActivityResponse `json:"activity"`
}

// @Summary Mark activity done
// @Description Marks a pending CRM activity as done, setting its done flag and recording the completion timestamp. Returns a conflict when the activity is already done.
// @Tags CRM Activities
// @Accept json
// @Produce json
// @Param id path integer true "CRM activity ID"
// @Success 200 {object} MarkActivityDoneResponseEnvelope "Activity marked done successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Activity not found"
// @Failure 409 {object} httpx.ErrorResponse "Activity already done"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /crm/activities/{id}/done [post]
func (h ProspectActivityHandler) MarkDone(c fiber.Ctx) error {
	activityID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid activity id provided.", nil)
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	activity, err := h.svc.FindInOrg(c, activityID, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("crm activity lookup failed", "activity_id", activityID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to mark activity done.", err)
	}
	if activity == nil {
		return httpx.CreateNotFoundResponse(c, "Activity not found.")
	}

	done, err := h.svc.MarkDone(c, activityID)
	if err != nil {
		return writeProspectActivityError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Activity marked done successfully.", MarkActivityDoneResponse{
		Activity: newProspectActivityResponse(done),
	})
}
