package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	PaymentBatchStateDraft     = "draft"
	PaymentBatchStateConfirmed = "confirmed"
	PaymentBatchStateSent      = "sent"
	PaymentBatchStateCancelled = "cancelled"
)

type PaymentBatch struct {
	model.Base
	OrganizationID uint64     `gorm:"not null" json:"organization_id"`
	Name           *string    `json:"name"`
	JournalID      uint64     `gorm:"not null" json:"journal_id"`
	TotalAmount    float64    `gorm:"type:numeric(18,4);default:0" json:"total_amount"`
	PaymentCount   int        `gorm:"default:0" json:"payment_count"`
	State          string     `gorm:"type:text" json:"state"`
	BatchDate      *time.Time `gorm:"type:date" json:"batch_date"`
	GeneratedAt    *time.Time `json:"generated_at"`
}

func (PaymentBatch) TableName() string {
	return "payment_batches"
}

type PaymentBatchLine struct {
	model.Base
	BatchID   uint64 `gorm:"not null" json:"batch_id"`
	PaymentID uint64 `gorm:"not null" json:"payment_id"`
}

func (PaymentBatchLine) TableName() string {
	return "payment_batch_lines"
}
