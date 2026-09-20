package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetContactResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetContactResponse `json:"data"`
}
type GetContactResponse struct {
	Contact ContactResponse `json:"contact"`
}

// @Summary Get contact
// @Description Gets a single contact by id after enforcing tenant ownership, returning 404 when the contact does not exist or belongs to another organization. The response includes HATEOAS links for viewing, updating, and deleting the contact.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Success 200 {object} GetContactResponseEnvelope "Contact retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id} [get]
func (h ContactHandler) Get(c fiber.Ctx) error {
	contact, ok := h.findContact(c)
	if !ok {
		return nil
	}

	return httpx.CreateSuccessResponseWithLinks(c, "Contact retrieved successfully.", GetContactResponse{
		Contact: newContactResponse(contact),
	}, buildContactLinks(c, contact.ID))
}
