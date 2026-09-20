package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type AccountFullReconcile struct {
	model.Base
	Name *string `json:"name"`
}

func (AccountFullReconcile) TableName() string {
	return "account_full_reconciles"
}

type AccountPartialReconcile struct {
	model.Base
	DebitLineID     uint64  `json:"debit_line_id"`
	CreditLineID    uint64  `json:"credit_line_id"`
	Amount          float64 `gorm:"type:numeric(18,4)" json:"amount"`
	FullReconcileID *uint64 `json:"full_reconcile_id"`
}

func (AccountPartialReconcile) TableName() string {
	return "account_partial_reconciles"
}

type ReminderAction struct {
	model.Base
	ContactID uint64     `json:"contact_id"`
	InvoiceID uint64     `json:"invoice_id"`
	LevelID   uint64     `json:"level_id"`
	SentAt    *time.Time `json:"sent_at"`
	Channel   *string    `json:"channel"`
}

func (ReminderAction) TableName() string {
	return "reminder_actions"
}
