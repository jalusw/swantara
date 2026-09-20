package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete contact
// @Description Soft-deletes a contact after enforcing tenant ownership, returning 404 when the contact does not exist or belongs to another organization. Deletion fails when other records still reference the contact, and the endpoint responds with 204 No Content on success.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id} [delete]
func (h ContactHandler) Delete(c fiber.Ctx) error {
	contactID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid contact id provided.", nil)
	}

	existing, err := h.svc.Find(c, contactID)
	if err != nil {
		httpx.RequestLog(c).Error("contact lookup failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete contact.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Contact not found.")
	}

	if err := h.svc.Delete(c, contactID); err != nil {
		httpx.RequestLog(c).Error("contact deletion failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete contact.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
