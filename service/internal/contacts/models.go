package contacts

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type Contact struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	DisplayName    *string `json:"display_name"`
	IsOrganization bool    `json:"is_organization"`
	ParentID       *uint64 `json:"parent_id"`
	Email          *string `json:"email" audit:"redact"`
	Phone          *string `json:"phone" audit:"redact"`
	Mobile         *string `json:"mobile" audit:"redact"`
	Website        *string `json:"website"`
	TaxID          *string `json:"tax_id" audit:"redact"`
	Industry       *string `json:"industry"`
	CurrencyCode   *string `json:"currency_code"`
	Lang           string  `json:"lang"`
	Active         bool    `json:"active"`
}

type ContactAddress struct {
	model.Base
	ContactID   uint64  `json:"contact_id"`
	Type        *string `json:"type"`
	Line1       *string `json:"line1"`
	Line2       *string `json:"line2"`
	City        *string `json:"city"`
	State       *string `json:"state"`
	PostalCode  *string `json:"postal_code"`
	CountryCode *string `json:"country_code"`
	IsDefault   bool    `json:"is_default"`
}

type ContactBankAccount struct {
	model.Base
	ContactID     uint64  `json:"contact_id"`
	AccountHolder *string `json:"account_holder"`
	BankName      *string `json:"bank_name"`
	IBAN          *string `json:"iban" audit:"redact"`
	SwiftBIC      *string `json:"swift_bic"`
	AccountNumber *string `json:"account_number" audit:"redact"`
	RoutingNumber *string `json:"routing_number" audit:"redact"`
	CurrencyCode  *string `json:"currency_code"`
}

type CustomerProfile struct {
	model.Base
	ContactID             uint64   `json:"contact_id"`
	CustomerPaymentTermID *uint64  `json:"customer_payment_term_id"`
	CreditLimit           *float64 `json:"credit_limit" gorm:"type:numeric(18,4)"`
	ReceivableAccountID   *uint64  `json:"receivable_account_id"`
	Active                bool     `json:"active"`
}

func (CustomerProfile) TableName() string { return "contact_customers" }

type SupplierProfile struct {
	model.Base
	ContactID             uint64  `json:"contact_id"`
	SupplierPaymentTermID *uint64 `json:"supplier_payment_term_id"`
	PayableAccountID      *uint64 `json:"payable_account_id"`
	Active                bool    `json:"active"`
}

func (SupplierProfile) TableName() string { return "contact_suppliers" }
