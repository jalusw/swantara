package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type SupplierProfileResponse struct {
	ContactID             uint64  `json:"contact_id"`
	SupplierPaymentTermID *uint64 `json:"supplier_payment_term_id"`
	PayableAccountID      *uint64 `json:"payable_account_id"`
	Active                bool    `json:"active"`
}

func newSupplierProfileResponse(record *contacts.SupplierProfile) SupplierProfileResponse {
	return SupplierProfileResponse{
		ContactID:             record.ContactID,
		SupplierPaymentTermID: record.SupplierPaymentTermID,
		PayableAccountID:      record.PayableAccountID,
		Active:                record.Active,
	}
}

type EnableSupplierRequest struct {
	SupplierPaymentTermID *uint64 `json:"supplier_payment_term_id"`
	PayableAccountID      *uint64 `json:"payable_account_id"`
	Active                bool    `json:"active"`
}

type EnableSupplierResponseEnvelope struct {
	httpx.EnvelopeBase
	Data EnableSupplierResponse `json:"data"`
}
type EnableSupplierResponse struct {
	Supplier SupplierProfileResponse `json:"supplier"`
}

// @Summary Enable or update supplier role
// @Description Enables the supplier role for a tenant-owned contact, creating the supplier record when it does not exist or updating it when it does, with supplier payment term and payable account. The role is always activated regardless of the submitted active flag, and the resulting supplier record is returned.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param body body EnableSupplierRequest true "Supplier details"
// @Success 200 {object} EnableSupplierResponseEnvelope "Supplier role enabled successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/supplier [put]
func (h ContactRelationHandler) EnableSupplier(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	var request EnableSupplierRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	record, err := h.svc.EnableSupplier(c, contactID, &contacts.SupplierProfile{
		SupplierPaymentTermID: request.SupplierPaymentTermID,
		PayableAccountID:      request.PayableAccountID,
		Active:                request.Active,
	})
	if err != nil {
		if errors.Is(err, contacts.ErrContactNotFound) {
			return httpx.CreateNotFoundResponse(c, "Contact not found.")
		}
		httpx.RequestLog(c).Error("enable supplier role failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to enable supplier role.", err)
	}

	return httpx.CreateSuccessResponse(c, "Supplier role enabled successfully.", EnableSupplierResponse{
		Supplier: newSupplierProfileResponse(record),
	})
}

// @Summary Disable supplier role
// @Description Deactivates the supplier role for a tenant-owned contact by setting its active flag to false. Does nothing when no supplier record exists, and responds with a success message.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Success 200 "Supplier role disabled successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/supplier [delete]
func (h ContactRelationHandler) DisableSupplier(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	if err := h.svc.DisableSupplier(c, contactID); err != nil {
		httpx.RequestLog(c).Error("disable supplier role failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to disable supplier role.", err)
	}

	return httpx.CreateSuccessResponse(c, "Supplier role disabled successfully.", struct{}{})
}
