package handler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ContactHandler struct {
	svc contacts.ContactService
}

func NewContactHandler(svc contacts.ContactService) ContactHandler {
	return ContactHandler{svc: svc}
}

func (h ContactHandler) findContact(c fiber.Ctx) (*contacts.Contact, bool) {
	contactID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		_ = httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid contact id provided.", nil)
		return nil, false
	}

	contact, err := h.svc.Find(c, contactID)
	if err != nil {
		httpx.RequestLog(c).Error("contact lookup failed", "contact_id", contactID, "error", err)
		_ = httpx.CreateInternalServerErrorResponse(c, "Failed to get contact.", err)
		return nil, false
	}
	if contact == nil || !httpx.OwnsTenant(c, contact.OrganizationID) {
		_ = httpx.CreateNotFoundResponse(c, "Contact not found.")
		return nil, false
	}
	return contact, true
}

func writeContactError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, contacts.ErrContactNotFound):
		return httpx.CreateNotFoundResponse(c, "Contact not found.")
	case errors.Is(err, contacts.ErrNameRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Contact name is required.", nil)
	case errors.Is(err, contacts.ErrInvalidAddressType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Address type must be one of billing, shipping, other.", nil)
	case errors.Is(err, contacts.ErrMultipleDefaultAddresses):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Only one default address is allowed per contact.", nil)
	default:
		httpx.RequestLog(c).Error("contact write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save contact.", err)
	}
}

func buildContactLinks(c fiber.Ctx, contactID uint64) []httpx.Link {
	base := fmt.Sprintf("%s/contacts/%d", httpx.LinkBaseFor(c), contactID)
	return []httpx.Link{
		{Rel: "self", Method: fiber.MethodGet, Href: base},
		{Rel: "update", Method: fiber.MethodPut, Href: base},
		{Rel: "delete", Method: fiber.MethodDelete, Href: base},
	}
}
