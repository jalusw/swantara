package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete organization
// @Description Permanently deletes an organization. Deletion fails with a conflict when other records still reference the organization.
// @Tags Organizations
// @Accept json
// @Produce json
// @Param id path integer true "Organization ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "Organization not found"
// @Failure 409 {object} httpx.ErrorResponse "Organization is referenced by other records"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{id} [delete]
func (h OrganizationHandler) Delete(c fiber.Ctx) error {
	orgID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid organization id provided.", nil)
	}

	existing, err := h.svc.Find(c, orgID)
	if err != nil {
		httpx.RequestLog(c).Error("organization lookup failed", "organization_id", orgID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete organization.", err)
	}
	if existing == nil {
		return httpx.CreateNotFoundResponse(c, "Organization not found.")
	}

	if err := h.svc.Delete(c, orgID); err != nil {
		httpx.RequestLog(c).Error("organization deletion failed", "organization_id", orgID, "error", err)
		return httpx.CreateConflictResponse(c, "Organization cannot be deleted because other records reference it.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
