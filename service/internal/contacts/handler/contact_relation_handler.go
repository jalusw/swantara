package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ContactRelationHandler struct {
	svc contacts.ContactService
}

func NewContactRelationHandler(
	svc contacts.ContactService,
) ContactRelationHandler {
	return ContactRelationHandler{
		svc: svc,
	}
}

func (h ContactRelationHandler) contactID(c fiber.Ctx) (uint64, bool) {
	contactID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		_ = httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid contact id provided.", nil)
		return 0, false
	}
	return contactID, true
}

func (h ContactRelationHandler) contactExists(c fiber.Ctx, contactID uint64) bool {
	contact, err := h.svc.Find(c, contactID)
	if err != nil {
		httpx.RequestLog(c).Error("contact lookup failed", "contact_id", contactID, "error", err)
		_ = httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve contact.", err)
		return false
	}
	if contact == nil || !httpx.OwnsTenant(c, contact.OrganizationID) {
		_ = httpx.CreateNotFoundResponse(c, "Contact not found.")
		return false
	}
	return true
}
