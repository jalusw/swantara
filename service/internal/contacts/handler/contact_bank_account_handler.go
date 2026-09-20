package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ContactBankAccountResponse struct {
	ID            uint64    `json:"id"`
	ContactID     uint64    `json:"contact_id"`
	AccountHolder *string   `json:"account_holder"`
	BankName      *string   `json:"bank_name"`
	IBAN          *string   `json:"iban"`
	SwiftBIC      *string   `json:"swift_bic"`
	AccountNumber *string   `json:"account_number"`
	RoutingNumber *string   `json:"routing_number"`
	CurrencyCode  *string   `json:"currency_code"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func newContactBankAccountResponse(account *contacts.ContactBankAccount) ContactBankAccountResponse {
	return ContactBankAccountResponse{
		ID:            account.ID,
		ContactID:     account.ContactID,
		AccountHolder: account.AccountHolder,
		BankName:      account.BankName,
		IBAN:          account.IBAN,
		SwiftBIC:      account.SwiftBIC,
		AccountNumber: account.AccountNumber,
		RoutingNumber: account.RoutingNumber,
		CurrencyCode:  account.CurrencyCode,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}
}

type ListContactBankAccountsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListContactBankAccountsResponse `json:"data"`
}
type ListContactBankAccountsResponse struct {
	BankAccounts []ContactBankAccountResponse `json:"bank_accounts"`
}

// @Summary List contact bank accounts
// @Description Lists a contact's bank accounts after verifying the contact exists and is owned by the caller's organization. Each account includes the account holder, bank name, IBAN, SWIFT/BIC, account and routing numbers, and currency.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Success 200 {object} ListContactBankAccountsResponseEnvelope "Bank accounts retrieved successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/bank-accounts [get]
func (h ContactRelationHandler) ListBankAccounts(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	accounts, err := h.svc.ListBankAccountsByContact(c, contactID)
	if err != nil {
		httpx.RequestLog(c).Error("contact bank accounts list failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve bank accounts.", err)
	}

	items := make([]ContactBankAccountResponse, len(accounts))
	for i, account := range accounts {
		items[i] = newContactBankAccountResponse(account)
	}

	return httpx.CreateSuccessResponse(c, "Bank accounts retrieved successfully.", ListContactBankAccountsResponse{
		BankAccounts: items,
	})
}

type CreateContactBankAccountRequest struct {
	AccountHolder *string `json:"account_holder"`
	BankName      *string `json:"bank_name"`
	IBAN          *string `json:"iban"`
	SwiftBIC      *string `json:"swift_bic"`
	AccountNumber *string `json:"account_number"`
	RoutingNumber *string `json:"routing_number"`
	CurrencyCode  *string `json:"currency_code"`
}

type CreateContactBankAccountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateContactBankAccountResponse `json:"data"`
}
type CreateContactBankAccountResponse struct {
	BankAccount ContactBankAccountResponse `json:"bank_account"`
}

// @Summary Create contact bank account
// @Description Adds a new bank account to an existing tenant-owned contact, returning 404 when the contact is missing or not owned by the caller's organization. The account holder, bank name, IBAN, SWIFT/BIC, account and routing numbers, and currency are recorded, and the new account is returned on success.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param body body CreateContactBankAccountRequest true "Bank account details"
// @Success 201 {object} CreateContactBankAccountResponseEnvelope "Bank account created successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/bank-accounts [post]
func (h ContactRelationHandler) CreateBankAccount(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}

	var request CreateContactBankAccountRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	created, err := h.svc.CreateBankAccount(c, &contacts.ContactBankAccount{
		ContactID:     contactID,
		AccountHolder: request.AccountHolder,
		BankName:      request.BankName,
		IBAN:          request.IBAN,
		SwiftBIC:      request.SwiftBIC,
		AccountNumber: request.AccountNumber,
		RoutingNumber: request.RoutingNumber,
		CurrencyCode:  request.CurrencyCode,
	})
	if err != nil {
		httpx.RequestLog(c).Error("contact bank account creation failed", "contact_id", contactID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create bank account.", err)
	}

	return httpx.CreateCreatedResponse(c, "Bank account created successfully.", CreateContactBankAccountResponse{
		BankAccount: newContactBankAccountResponse(created),
	})
}

type UpdateContactBankAccountRequest struct {
	AccountHolder *string `json:"account_holder"`
	BankName      *string `json:"bank_name"`
	IBAN          *string `json:"iban"`
	SwiftBIC      *string `json:"swift_bic"`
	AccountNumber *string `json:"account_number"`
	RoutingNumber *string `json:"routing_number"`
	CurrencyCode  *string `json:"currency_code"`
}

type UpdateContactBankAccountResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateContactBankAccountResponse `json:"data"`
}
type UpdateContactBankAccountResponse struct {
	BankAccount ContactBankAccountResponse `json:"bank_account"`
}

// @Summary Update contact bank account
// @Description Updates an existing bank account of a tenant-owned contact, returning 404 when the contact or account does not exist or the account belongs to a different contact. All bank account fields are replaced with the submitted values.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param account_id path integer true "Bank account ID"
// @Param body body UpdateContactBankAccountRequest true "Bank account details"
// @Success 200 {object} UpdateContactBankAccountResponseEnvelope "Bank account updated successfully."
// @Failure 404 {object} httpx.ErrorResponse "Contact or bank account not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/bank-accounts/{account_id} [put]
func (h ContactRelationHandler) UpdateBankAccount(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}
	accountID, err := strconv.ParseUint(c.Params("account_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid bank account id provided.", nil)
	}

	var request UpdateContactBankAccountRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	existing, err := h.svc.FindBankAccount(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("contact bank account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update bank account.", err)
	}
	if existing == nil || existing.ContactID != contactID {
		return httpx.CreateNotFoundResponse(c, "Bank account not found.")
	}

	existing.AccountHolder = request.AccountHolder
	existing.BankName = request.BankName
	existing.IBAN = request.IBAN
	existing.SwiftBIC = request.SwiftBIC
	existing.AccountNumber = request.AccountNumber
	existing.RoutingNumber = request.RoutingNumber
	existing.CurrencyCode = request.CurrencyCode

	updated, err := h.svc.UpdateBankAccount(c, existing)
	if err != nil {
		httpx.RequestLog(c).Error("contact bank account update failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update bank account.", err)
	}

	return httpx.CreateSuccessResponse(c, "Bank account updated successfully.", UpdateContactBankAccountResponse{
		BankAccount: newContactBankAccountResponse(updated),
	})
}

// @Summary Delete contact bank account
// @Description Deletes a bank account after verifying it belongs to the specified tenant-owned contact, returning 404 when the account is missing or associated with a different contact. The endpoint responds with 204 No Content on success.
// @Tags Contacts
// @Accept json
// @Produce json
// @Param id path integer true "Contact ID"
// @Param account_id path integer true "Bank account ID"
// @Success 204 "No Content"
// @Failure 404 {object} httpx.ErrorResponse "Contact or bank account not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/contacts/{id}/bank-accounts/{account_id} [delete]
func (h ContactRelationHandler) DeleteBankAccount(c fiber.Ctx) error {
	contactID, ok := h.contactID(c)
	if !ok {
		return nil
	}
	if !h.contactExists(c, contactID) {
		return nil
	}
	accountID, err := strconv.ParseUint(c.Params("account_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid bank account id provided.", nil)
	}

	existing, err := h.svc.FindBankAccount(c, accountID)
	if err != nil {
		httpx.RequestLog(c).Error("contact bank account lookup failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete bank account.", err)
	}
	if existing == nil || existing.ContactID != contactID {
		return httpx.CreateNotFoundResponse(c, "Bank account not found.")
	}

	if err := h.svc.DeleteBankAccount(c, accountID); err != nil {
		httpx.RequestLog(c).Error("contact bank account deletion failed", "account_id", accountID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete bank account.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
