package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateContactRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	DisplayName    *string `json:"display_name"`
	IsOrganization bool    `json:"is_organization"`
	ParentID       *uint64 `json:"parent_id"`
	Email          *string `json:"email" validate:"omitempty,email"`
	Phone          *string `json:"phone"`
	Mobile         *string `json:"mobile"`
	Website        *string `json:"website"`
	TaxID          *string `json:"tax_id"`
	Industry       *string `json:"industry"`
	CurrencyCode   *string `json:"currency_code"`
	Lang           string  `json:"lang"`
	Active         bool    `json:"active"`
}

type UpdateContactResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateContactResponse `json:"data"`
}
type UpdateContactResponse struct {
	Contact ContactResponse `json:"contact"`
}

// @Summary Update contact
// @Description Updates a contact's base identity fields after enforcing tenant ownership, returning 404 when the contact is missing or not owned by the caller's organization. A non-empty name is required, and the updated contact is returned with HATEOAS links.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param body body UpdateContactRequest true "Contact details"
// @Success 200 {object} UpdateContactResponseEnvelope "Contact updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id} [put]
func (h ContactHandler) Update(c fiber.Ctx) error {
	contactID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid contact id provided.", nil)
	}

	var request UpdateContactRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	existing, err := h.svc.Find(c, contactID)
	if err != nil {
		httpx.RequestLog(c).Error("contact lookup failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update contact.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Contact not found.")
	}

	contact, err := h.svc.Update(c, contactID, &contacts.Contact{
		OrganizationID: httpx.TenantOrganizationID(c, existing.OrganizationID),
		Name:           request.Name,
		DisplayName:    request.DisplayName,
		IsOrganization: request.IsOrganization,
		ParentID:       request.ParentID,
		Email:          request.Email,
		Phone:          request.Phone,
		Mobile:         request.Mobile,
		Website:        request.Website,
		TaxID:          request.TaxID,
		Industry:       request.Industry,
		CurrencyCode:   request.CurrencyCode,
		Lang:           request.Lang,
		Active:         request.Active,
	})
	if err != nil {
		return writeContactError(c, err)
	}

	return httpx.CreateSuccessResponseWithLinks(c, "Contact updated successfully.", UpdateContactResponse{
		Contact: newContactResponse(contact),
	}, buildContactLinks(c, contact.ID))
}
