package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	BankStatementStateDraft      = "draft"
	BankStatementStateOpen       = "open"
	BankStatementStateReconciled = "reconciled"
	BankStatementStateCancelled  = "cancelled"
)

type BankStatement struct {
	model.Base
	JournalID    *uint64       `json:"journal_id"`
	Name         *string       `json:"name"`
	Date         *time.Time    `gorm:"type:date" json:"date"`
	BalanceStart amount.Amount `gorm:"type:numeric(18,4);default:0" json:"balance_start"`
	BalanceEnd   amount.Amount `gorm:"type:numeric(18,4);default:0" json:"balance_end"`
	CurrencyCode *string       `gorm:"type:char(3)" json:"currency_code"`
	State        string        `gorm:"type:text" json:"state"`
}

func (BankStatement) TableName() string {
	return "bank_statements"
}

type BankStatementLine struct {
	model.Base
	StatementID     uint64        `gorm:"not null" json:"statement_id"`
	Date            *time.Time    `gorm:"type:date" json:"date"`
	Amount          amount.Amount `gorm:"type:numeric(18,4)" json:"amount"`
	CurrencyCode    *string       `gorm:"type:char(3)" json:"currency_code"`
	FeeAmount       amount.Amount `gorm:"type:numeric(18,4);default:0" json:"fee_amount"`
	InterestAmount  amount.Amount `gorm:"type:numeric(18,4);default:0" json:"interest_amount"`
	ContactID       *uint64       `json:"contact_id"`
	Ref             *string       `json:"ref"`
	Narration       *string       `json:"narration"`
	Reconciled      bool          `gorm:"default:false" json:"reconciled"`
	PaymentID       *uint64       `json:"payment_id"`
	JournalLineID   *uint64       `json:"journal_line_id"`
	FeeEntryID      *uint64       `json:"fee_entry_id"`
	InterestEntryID *uint64       `json:"interest_entry_id"`
}

func StatementGross(line *BankStatementLine) amount.Amount {
	net := line.Amount.Abs()
	gross := net.Add(line.FeeAmount.Abs()).Sub(line.InterestAmount.Abs())
	if gross.IsNegative() {
		gross = amount.Zero()
	}
	if line.Amount.IsNegative() {
		return gross.Neg()
	}
	return gross
}

func (BankStatementLine) TableName() string {
	return "bank_statement_lines"
}
