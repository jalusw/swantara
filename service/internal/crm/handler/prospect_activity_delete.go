package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete activity
// @Description Deletes a CRM activity by id after verifying it exists, returning 404 otherwise. The record is removed permanently and the endpoint responds with 204 No Content on success.
// @Tags CRM Activities
// @Accept json
// @Produce json
// @Param id path integer true "CRM activity ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Activity not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /crm/activities/{id} [delete]
func (h ProspectActivityHandler) Delete(c fiber.Ctx) error {
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
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete activity.", err)
	}
	if activity == nil {
		return httpx.CreateNotFoundResponse(c, "Activity not found.")
	}

	if err := h.svc.Delete(c, activityID); err != nil {
		httpx.RequestLog(c).Error("crm activity deletion failed", "activity_id", activityID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete activity.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
