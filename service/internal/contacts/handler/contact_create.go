package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ContactAddressRequest struct {
	Type        *string `json:"type" validate:"omitempty,oneof=billing shipping other"`
	Line1       *string `json:"line1"`
	Line2       *string `json:"line2"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	PostalCode  *string `json:"postal_code"`
	CountryCode *string `json:"country_code"`
	IsDefault   bool    `json:"is_default"`
}

type ContactBankAccountRequest struct {
	AccountHolder *string `json:"account_holder"`
	BankName      *string `json:"bank_name"`
	IBAN          *string `json:"iban"`
	SwiftBIC      *string `json:"swift_bic"`
	AccountNumber *string `json:"account_number"`
	RoutingNumber *string `json:"routing_number"`
	CurrencyCode  *string `json:"currency_code"`
}

type CustomerProfileRequest struct {
	CustomerPaymentTermID *uint64  `json:"customer_payment_term_id"`
	CreditLimit           *float64 `json:"credit_limit"`
	ReceivableAccountID   *uint64  `json:"receivable_account_id"`
	Active                bool     `json:"active"`
}

type SupplierProfileRequest struct {
	SupplierPaymentTermID *uint64 `json:"supplier_payment_term_id"`
	PayableAccountID      *uint64 `json:"payable_account_id"`
	Active                bool    `json:"active"`
}

type CreateContactRequest struct {
	OrganizationID *uint64                     `json:"organization_id"`
	Name           string                      `json:"name" validate:"required"`
	DisplayName    *string                     `json:"display_name"`
	IsOrganization bool                        `json:"is_organization"`
	ParentID       *uint64                     `json:"parent_id"`
	Email          *string                     `json:"email" validate:"omitempty,email"`
	Phone          *string                     `json:"phone"`
	Mobile         *string                     `json:"mobile"`
	Website        *string                     `json:"website"`
	TaxID          *string                     `json:"tax_id"`
	Industry       *string                     `json:"industry"`
	CurrencyCode   *string                     `json:"currency_code"`
	Lang           string                      `json:"lang"`
	Active         bool                        `json:"active"`
	Addresses      []ContactAddressRequest     `json:"addresses"`
	BankAccounts   []ContactBankAccountRequest `json:"bank_accounts"`
	Customer       *CustomerProfileRequest     `json:"customer"`
	Supplier       *SupplierProfileRequest     `json:"supplier"`
}

type CreateContactResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateContactResponse `json:"data"`
}
type CreateContactResponse struct {
	Contact ContactResponse `json:"contact"`
}

// @Summary Create contact
// @Description Creates a contact with optional addresses, bank accounts, and customer or supplier roles, all persisted in a single transaction. The contact name is required, and addresses are validated for a valid type and a single default per type. The resulting contact is returned with HATEOAS links.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param body body CreateContactRequest true "Contact details"
// @Success 201 {object} CreateContactResponseEnvelope "Contact created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 403 {object} httpx.ErrorResponse "Forbidden"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts [post]
func (h ContactHandler) Create(c fiber.Ctx) error {
	var request CreateContactRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	addresses := make([]*contacts.ContactAddress, len(request.Addresses))
	for i, address := range request.Addresses {
		addresses[i] = &contacts.ContactAddress{
			Type:        address.Type,
			Line1:       address.Line1,
			Line2:       address.Line2,
			City:        address.City,
			State:       address.State,
			PostalCode:  address.PostalCode,
			CountryCode: address.CountryCode,
			IsDefault:   address.IsDefault,
		}
	}

	banks := make([]*contacts.ContactBankAccount, len(request.BankAccounts))
	for i, bank := range request.BankAccounts {
		banks[i] = &contacts.ContactBankAccount{
			AccountHolder: bank.AccountHolder,
			BankName:      bank.BankName,
			IBAN:          bank.IBAN,
			SwiftBIC:      bank.SwiftBIC,
			AccountNumber: bank.AccountNumber,
			RoutingNumber: bank.RoutingNumber,
			CurrencyCode:  bank.CurrencyCode,
		}
	}

	var customer *contacts.CustomerProfile
	if request.Customer != nil {
		customer = &contacts.CustomerProfile{
			CustomerPaymentTermID: request.Customer.CustomerPaymentTermID,
			CreditLimit:           request.Customer.CreditLimit,
			ReceivableAccountID:   request.Customer.ReceivableAccountID,
			Active:                request.Customer.Active,
		}
	}

	var supplier *contacts.SupplierProfile
	if request.Supplier != nil {
		supplier = &contacts.SupplierProfile{
			SupplierPaymentTermID: request.Supplier.SupplierPaymentTermID,
			PayableAccountID:      request.Supplier.PayableAccountID,
			Active:                request.Supplier.Active,
		}
	}

	contact, err := h.svc.Create(c, &contacts.Contact{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
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
	}, addresses, banks, customer, supplier)
	if err != nil {
		return writeContactError(c, err)
	}

	return httpx.CreateCreatedResponseWithLinks(c, "Contact created successfully.", CreateContactResponse{
		Contact: newContactResponse(contact),
	}, buildContactLinks(c, contact.ID))
}
