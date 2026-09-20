package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ContactAddressResponse struct {
	ID          uint64    `json:"id"`
	ContactID   uint64    `json:"contact_id"`
	Type        *string   `json:"type"`
	Line1       *string   `json:"line1"`
	Line2       *string   `json:"line2"`
	City        *string   `json:"city"`
	State       *string   `json:"state"`
	PostalCode  *string   `json:"postal_code"`
	CountryCode *string   `json:"country_code"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newContactAddressResponse(address *contacts.ContactAddress) ContactAddressResponse {
	return ContactAddressResponse{
		ID:          address.ID,
		ContactID:   address.ContactID,
		Type:        address.Type,
		Line1:       address.Line1,
		Line2:       address.Line2,
		City:        address.City,
		State:       address.State,
		PostalCode:  address.PostalCode,
		CountryCode: address.CountryCode,
		IsDefault:   address.IsDefault,
		CreatedAt:   address.CreatedAt,
		UpdatedAt:   address.UpdatedAt,
	}
}

type ListContactAddressesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListContactAddressesResponse `json:"data"`
}
type ListContactAddressesResponse struct {
	Addresses []ContactAddressResponse `json:"addresses"`
}

// @Summary List contact addresses
// @Description Lists a contact's addresses after verifying the contact exists and is owned by the caller's organization. Each address includes its type, lines, city, state, postal code, country, and default flag, and results can be exported as CSV.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Success 200 {object} ListContactAddressesResponseEnvelope "Addresses retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/addresses [get]
func (h ContactRelationHandler) ListAddresses(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	addresses, err := h.svc.ListAddressesByContact(c, contactID)
	if err != nil {
		httpx.RequestLog(c).Error("contact addresses list failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve addresses.", err)
	}

	items := make([]ContactAddressResponse, len(addresses))
	for i, address := range addresses {
		items[i] = newContactAddressResponse(address)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "contact-addresses.csv", items)
	}

	return httpx.CreateSuccessResponse(c, "Addresses retrieved successfully.", ListContactAddressesResponse{
		Addresses: items,
	})
}

type CreateContactAddressRequest struct {
	Type        *string `json:"type" validate:"omitempty,oneof=billing shipping other"`
	Line1       *string `json:"line1"`
	Line2       *string `json:"line2"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	PostalCode  *string `json:"postal_code"`
	CountryCode *string `json:"country_code"`
	IsDefault   bool    `json:"is_default"`
}

type CreateContactAddressResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateContactAddressResponse `json:"data"`
}
type CreateContactAddressResponse struct {
	Address ContactAddressResponse `json:"address"`
}

// @Summary Create contact address
// @Description Adds a new address to an existing tenant-owned contact, returning 404 when the contact is missing or not owned by the caller's organization. The address type is validated as billing, shipping, or other, and the new address is returned on success.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param body body CreateContactAddressRequest true "Address details"
// @Success 201 {object} CreateContactAddressResponseEnvelope "Address created successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/addresses [post]
func (h ContactRelationHandler) CreateAddress(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	var request CreateContactAddressRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	created, err := h.svc.CreateAddress(c, &contacts.ContactAddress{
		ContactID:   contactID,
		Type:        request.Type,
		Line1:       request.Line1,
		Line2:       request.Line2,
		City:        request.City,
		State:       request.State,
		PostalCode:  request.PostalCode,
		CountryCode: request.CountryCode,
		IsDefault:   request.IsDefault,
	})
	if err != nil {
		httpx.RequestLog(c).Error("contact address creation failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create address.", err)
	}

	return httpx.CreateCreatedResponse(c, "Address created successfully.", CreateContactAddressResponse{
		Address: newContactAddressResponse(created),
	})
}

type UpdateContactAddressRequest struct {
	Type        *string `json:"type" validate:"omitempty,oneof=billing shipping other"`
	Line1       *string `json:"line1"`
	Line2       *string `json:"line2"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	PostalCode  *string `json:"postal_code"`
	CountryCode *string `json:"country_code"`
	IsDefault   bool    `json:"is_default"`
}

type UpdateContactAddressResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateContactAddressResponse `json:"data"`
}
type UpdateContactAddressResponse struct {
	Address ContactAddressResponse `json:"address"`
}

// @Summary Update contact address
// @Description Updates an existing address of a tenant-owned contact, returning 404 when the contact or address does not exist or the address belongs to a different contact. The address type is validated as billing, shipping, or other, and all address fields are replaced with the submitted values.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param address_id path integer true "Address ID"
// @Param body body UpdateContactAddressRequest true "Address details"
// @Success 200 {object} UpdateContactAddressResponseEnvelope "Address updated successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact or address not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/addresses/{address_id} [put]
func (h ContactRelationHandler) UpdateAddress(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}
	addressID, err := strconv.ParseUint(c.Params("address_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid address id provided.", nil)
	}

	var request UpdateContactAddressRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	existing, err := h.svc.FindAddress(c, addressID)
	if err != nil {
		httpx.RequestLog(c).Error("contact address lookup failed", "address_id", addressID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update address.", err)
	}
	if existing == nil || existing.ContactID != contactID {
		return httpx.CreateNotFoundResponse(c, "Address not found.")
	}

	existing.Type = request.Type
	existing.Line1 = request.Line1
	existing.Line2 = request.Line2
	existing.City = request.City
	existing.State = request.State
	existing.PostalCode = request.PostalCode
	existing.CountryCode = request.CountryCode
	existing.IsDefault = request.IsDefault

	updated, err := h.svc.UpdateAddress(c, existing)
	if err != nil {
		httpx.RequestLog(c).Error("contact address update failed", "address_id", addressID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update address.", err)
	}

	return httpx.CreateSuccessResponse(c, "Address updated successfully.", UpdateContactAddressResponse{
		Address: newContactAddressResponse(updated),
	})
}

// @Summary Delete contact address
// @Description Deletes an address after verifying it belongs to the specified tenant-owned contact, returning 404 when the address is missing or associated with a different contact. The endpoint responds with 204 No Content on success.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param address_id path integer true "Address ID"
// @Success 204 "No Content"
// @Failure 404 {object} httpx.ErrorResponse "Contact or address not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/addresses/{address_id} [delete]
func (h ContactRelationHandler) DeleteAddress(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}
	addressID, err := strconv.ParseUint(c.Params("address_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid address id provided.", nil)
	}

	existing, err := h.svc.FindAddress(c, addressID)
	if err != nil {
		httpx.RequestLog(c).Error("contact address lookup failed", "address_id", addressID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete address.", err)
	}
	if existing == nil || existing.ContactID != contactID {
		return httpx.CreateNotFoundResponse(c, "Address not found.")
	}

	if err := h.svc.DeleteAddress(c, addressID); err != nil {
		httpx.RequestLog(c).Error("contact address deletion failed", "address_id", addressID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete address.", err)
	}

	return httpx.CreateNoContentResponse(c)
}

type ContactAddressDefaultResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ContactAddressDefaultResponse `json:"data"`
}
type ContactAddressDefaultResponse struct {
	Address ContactAddressResponse `json:"address"`
}

// @Summary Set default contact address
// @Description Marks an address as the default for a tenant-owned contact, clearing the default flag on the contact's other addresses in a single transaction. Returns 404 when the contact or address does not exist or the address belongs to a different contact.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param address_id path integer true "Address ID"
// @Success 200 {object} ContactAddressDefaultResponseEnvelope "Default address updated successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact or address not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/addresses/{address_id}/default [post]
func (h ContactRelationHandler) SetDefaultAddress(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}
	addressID, err := strconv.ParseUint(c.Params("address_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid address id provided.", nil)
	}

	if err := h.svc.SetDefaultAddress(c, contactID, addressID); err != nil {
		switch {
		case errors.Is(err, contacts.ErrAddressNotFound), errors.Is(err, contacts.ErrAddressNotForContact):
			return httpx.CreateNotFoundResponse(c, "Address not found.")
		default:
			httpx.RequestLog(c).Error("set default address failed", "address_id", addressID, "error", err)
			return httpx.CreateInternalServerErrorResponse(c, "Failed to set default address.", err)
		}
	}

	address, err := h.svc.FindAddress(c, addressID)
	if err != nil {
		httpx.RequestLog(c).Error("contact address lookup failed", "address_id", addressID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get address.", err)
	}

	return httpx.CreateSuccessResponse(c, "Default address updated successfully.", ContactAddressDefaultResponse{
		Address: newContactAddressResponse(address),
	})
}
