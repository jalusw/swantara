package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CustomerProfileResponse struct {
	ContactID             uint64   `json:"contact_id"`
	CustomerPaymentTermID *uint64  `json:"customer_payment_term_id"`
	CreditLimit           *float64 `json:"credit_limit"`
	ReceivableAccountID   *uint64  `json:"receivable_account_id"`
	Active                bool     `json:"active"`
}

func newCustomerProfileResponse(record *contacts.CustomerProfile) CustomerProfileResponse {
	return CustomerProfileResponse{
		ContactID:             record.ContactID,
		CustomerPaymentTermID: record.CustomerPaymentTermID,
		CreditLimit:           record.CreditLimit,
		ReceivableAccountID:   record.ReceivableAccountID,
		Active:                record.Active,
	}
}

type EnableCustomerRequest struct {
	CustomerPaymentTermID *uint64  `json:"customer_payment_term_id"`
	CreditLimit           *float64 `json:"credit_limit"`
	ReceivableAccountID   *uint64  `json:"receivable_account_id"`
	Active                bool     `json:"active"`
}

type EnableCustomerResponseEnvelope struct {
	httpx.EnvelopeBase
	Data EnableCustomerResponse `json:"data"`
}
type EnableCustomerResponse struct {
	Customer CustomerProfileResponse `json:"customer"`
}

// @Summary Enable or update customer role
// @Description Enables the customer role for a tenant-owned contact, creating the customer record when it does not exist or updating it when it does, with payment term, credit limit, and receivable account. The role is always activated regardless of the submitted active flag, and the resulting customer record is returned.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param body body EnableCustomerRequest true "Customer details"
// @Success 200 {object} EnableCustomerResponseEnvelope "Customer role enabled successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/customer [put]
func (h ContactRelationHandler) EnableCustomer(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	var request EnableCustomerRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	record, err := h.svc.EnableCustomer(c, contactID, &contacts.CustomerProfile{
		CustomerPaymentTermID: request.CustomerPaymentTermID,
		CreditLimit:           request.CreditLimit,
		ReceivableAccountID:   request.ReceivableAccountID,
		Active:                request.Active,
	})
	if err != nil {
		if errors.Is(err, contacts.ErrContactNotFound) {
			return httpx.CreateNotFoundResponse(c, "Contact not found.")
		}
		httpx.RequestLog(c).Error("enable customer role failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to enable customer role.", err)
	}

	return httpx.CreateSuccessResponse(c, "Customer role enabled successfully.", EnableCustomerResponse{
		Customer: newCustomerProfileResponse(record),
	})
}

// @Summary Disable customer role
// @Description Deactivates the customer role for a tenant-owned contact by setting its active flag to false. Does nothing when no customer record exists, and responds with a success message.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Success 200 "Customer role disabled successfully."
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/customer [delete]
func (h ContactRelationHandler) DisableCustomer(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	if err := h.svc.DisableCustomer(c, contactID); err != nil {
		httpx.RequestLog(c).Error("disable customer role failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to disable customer role.", err)
	}

	return httpx.CreateSuccessResponse(c, "Customer role disabled successfully.", struct{}{})
}
