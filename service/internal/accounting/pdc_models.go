package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	PdcStateHeld      = "held"
	PdcStateDeposited = "deposited"
	PdcStateCleared   = "cleared"
	PdcStateBounced   = "bounced"
	PdcStateCancelled = "cancelled"
)

type PdcInstrument struct {
	model.Base
	OrganizationID *uint64       `json:"organization_id"`
	ContactID      uint64        `gorm:"not null" json:"contact_id"`
	Direction      string        `gorm:"type:text" json:"direction"`
	Number         *string       `json:"number"`
	BankName       *string       `json:"bank_name"`
	Amount         amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
	CurrencyCode   *string       `gorm:"type:char(3)" json:"currency_code"`
	DueDate        *time.Time    `gorm:"type:date" json:"due_date"`
	State          string        `gorm:"type:text" json:"state"`
	InvoiceID      *uint64       `json:"invoice_id"`
	PaymentID      *uint64       `json:"payment_id"`
	JournalID      *uint64       `json:"journal_id"`
}

func (PdcInstrument) TableName() string {
	return "pdc_instruments"
}
