package reporting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	FxRevaluationStatePosted   = "posted"
	FxRevaluationStateReversed = "reversed"

	AccrualStatePosted   = "posted"
	AccrualStateReversed = "reversed"
)

type FxRevaluation struct {
	model.Base
	OrganizationID uint64    `json:"organization_id"`
	PeriodID       uint64    `json:"period_id"`
	Name           *string   `json:"name"`
	Date           time.Time `gorm:"type:date" json:"date"`
	State          string    `gorm:"type:text" json:"state"`
	TotalGainLoss  float64   `gorm:"type:numeric(18,4);default:0" json:"total_gain_loss"`
}

func (FxRevaluation) TableName() string {
	return "fx_revaluations"
}

type FxRevaluationLine struct {
	model.Base
	RevaluationID  uint64  `json:"revaluation_id"`
	AccountID      uint64  `json:"account_id"`
	CurrencyCode   string  `gorm:"type:char(3)" json:"currency_code"`
	ForeignBalance float64 `gorm:"type:numeric(18,4);default:0" json:"foreign_balance"`
	BaseBalance    float64 `gorm:"type:numeric(18,4);default:0" json:"base_balance"`
	ClosingRate    float64 `gorm:"type:numeric(18,8);default:0" json:"closing_rate"`
	GainLoss       float64 `gorm:"type:numeric(18,4);default:0" json:"gain_loss"`
	EntryID        *uint64 `json:"entry_id"`
	Reversed       bool    `gorm:"default:false" json:"reversed"`
}

func (FxRevaluationLine) TableName() string {
	return "fx_revaluation_lines"
}

type Accrual struct {
	model.Base
	OrganizationID  uint64     `json:"organization_id"`
	PeriodID        uint64     `json:"period_id"`
	Name            *string    `json:"name"`
	Description     *string    `json:"description"`
	ReversalDate    *time.Time `gorm:"type:date" json:"reversal_date"`
	State           string     `gorm:"type:text" json:"state"`
	EntryID         *uint64    `json:"entry_id"`
	ReversalEntryID *uint64    `json:"reversal_entry_id"`
}

func (Accrual) TableName() string {
	return "accruals"
}

type AccrualLine struct {
	model.Base
	AccrualID uint64  `json:"accrual_id"`
	AccountID uint64  `json:"account_id"`
	Name      *string `json:"name"`
	Debit     float64 `gorm:"type:numeric(18,4);default:0" json:"debit"`
	Credit    float64 `gorm:"type:numeric(18,4);default:0" json:"credit"`
}

func (AccrualLine) TableName() string {
	return "accrual_lines"
}
