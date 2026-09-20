package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetProspectActivityResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetProspectActivityResponse `json:"data"`
}

type GetProspectActivityResponse struct {
	Activity ProspectActivityResponse `json:"activity"`
}

// @Summary Get activity
// @Description Gets a single CRM activity by id, returning 404 when no activity with that id exists. The response includes the activity's type, summary, note, due date, completion state, and linked prospect or contact references.
// @Tags CRM Activities
// @Accept json
// @Produce json
// @Param id path integer true "CRM activity ID"
// @Success 200 {object} GetProspectActivityResponseEnvelope "Activity retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Activity not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /crm/activities/{id} [get]
func (h ProspectActivityHandler) Get(c fiber.Ctx) error {
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
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get activity.", err)
	}
	if activity == nil {
		return httpx.CreateNotFoundResponse(c, "Activity not found.")
	}

	return httpx.CreateSuccessResponse(c, "Activity retrieved successfully.", GetProspectActivityResponse{
		Activity: newProspectActivityResponse(activity),
	})
}
