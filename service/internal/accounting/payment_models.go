package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	PaymentTypeInbound  = "inbound"
	PaymentTypeOutbound = "outbound"

	PaymentStateDraft      = "draft"
	PaymentStatePosted     = "posted"
	PaymentStateReconciled = "reconciled"
	PaymentStateCancelled  = "cancelled"
)

type Payment struct {
	model.Base
	OrganizationID       *uint64   `json:"organization_id"`
	Name                 *string   `json:"name"`
	ContactID            uint64    `gorm:"not null" json:"contact_id"`
	Type                 string    `gorm:"type:text" json:"type"`
	JournalID            *uint64   `json:"journal_id"`
	PaymentMethod        *string   `json:"payment_method"`
	Amount               float64   `gorm:"type:numeric(18,4);not null" json:"amount"`
	CurrencyCode         *string   `gorm:"type:char(3)" json:"currency_code"`
	Date                 time.Time `gorm:"type:date;not null" json:"date"`
	Reference            *string   `json:"reference"`
	EntryID              *uint64   `json:"entry_id"`
	ContactBankAccountID *uint64   `json:"contact_bank_account_id"`
	State                string    `gorm:"type:text" json:"state"`
}

func (Payment) TableName() string {
	return "payments"
}

type PaymentAllocation struct {
	model.Base
	PaymentID uint64  `json:"payment_id"`
	InvoiceID uint64  `gorm:"not null" json:"invoice_id"`
	Amount    float64 `gorm:"type:numeric(18,4)" json:"amount"`
}

func (PaymentAllocation) TableName() string {
	return "payment_allocations"
}
