package accounting

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	DeferredTypeDeferredRevenue = "deferred_revenue"
	DeferredTypeDeferredExpense = "deferred_expense"
	DeferredTypePrepaid         = "prepaid"

	DeferredMethodLinear    = "linear"
	DeferredMethodManual    = "manual"
	DeferredMethodMilestone = "milestone"

	DeferredStateDraft     = "draft"
	DeferredStateRunning   = "running"
	DeferredStateDone      = "done"
	DeferredStateCancelled = "cancelled"
)

type DeferredSchedule struct {
	model.Base
	OrganizationID        *uint64    `json:"organization_id"`
	Type                  string     `gorm:"type:text" json:"type"`
	SourceType            string     `json:"source_type"`
	SourceID              uint64     `json:"source_id"`
	ContactID             *uint64    `json:"contact_id"`
	ItemID                *uint64    `json:"item_id"`
	TotalAmount           float64    `gorm:"type:numeric(18,4)" json:"total_amount"`
	RecognizedAmount      float64    `gorm:"type:numeric(18,4)" json:"recognized_amount"`
	BalanceSheetAccountID *uint64    `json:"balance_sheet_account_id"`
	PLAccountID           *uint64    `json:"pl_account_id"`
	Method                string     `gorm:"type:text" json:"method"`
	DateStart             *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd               *time.Time `gorm:"type:date" json:"date_end"`
	Periods               int        `json:"periods"`
	State                 string     `gorm:"type:text" json:"state"`
	DimensionID           *uint64    `json:"dimension_id"`
}

func (DeferredSchedule) TableName() string {
	return "deferred_schedules"
}

type DeferredScheduleLine struct {
	model.Base
	ScheduleID      uint64     `json:"schedule_id"`
	Sequence        int        `json:"sequence"`
	RecognitionDate *time.Time `gorm:"type:date" json:"recognition_date"`
	Amount          float64    `gorm:"type:numeric(18,4)" json:"amount"`
	EntryID         *uint64    `json:"entry_id"`
	Posted          bool       `json:"posted"`
}

func (DeferredScheduleLine) TableName() string {
	return "deferred_schedule_lines"
}
