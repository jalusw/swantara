package commission

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	BasisRevenue   = "revenue"
	BasisMargin    = "margin"
	BasisCollected = "collected"

	EntryStateDraft     = "draft"
	EntryStateConfirmed = "confirmed"
	EntryStatePaid      = "paid"
	EntryStateCancelled = "cancelled"
)

type CommissionPlan struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Basis          string  `gorm:"type:text" json:"basis"`
	Active         bool    `gorm:"default:true" json:"active"`
}

func (CommissionPlan) TableName() string {
	return "commission_plans"
}

type CommissionRule struct {
	model.Base
	PlanID         uint64  `json:"plan_id"`
	ItemCategoryID *uint64 `json:"item_category_id"`
	MinAmount      float64 `gorm:"type:numeric(18,4)" json:"min_amount"`
	MaxAmount      float64 `gorm:"type:numeric(18,4)" json:"max_amount"`
	RatePct        float64 `gorm:"type:numeric(8,4)" json:"rate_pct"`
	FixedAmount    float64 `gorm:"type:numeric(18,4)" json:"fixed_amount"`
}

func (CommissionRule) TableName() string {
	return "commission_rules"
}

type CommissionAssignment struct {
	model.Base
	PlanID        uint64     `json:"plan_id"`
	SalespersonID uint64     `json:"salesperson_id"`
	DateStart     *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd       *time.Time `gorm:"type:date" json:"date_end"`
}

func (CommissionAssignment) TableName() string {
	return "commission_assignments"
}

type CommissionEntry struct {
	model.Base
	SalespersonID    uint64  `json:"salesperson_id"`
	PlanID           uint64  `json:"plan_id"`
	SourceType       string  `json:"source_type"`
	SourceID         uint64  `json:"source_id"`
	BaseAmount       float64 `gorm:"type:numeric(18,4)" json:"base_amount"`
	CommissionAmount float64 `gorm:"type:numeric(18,4)" json:"commission_amount"`
	State            string  `gorm:"type:text" json:"state"`
	PeriodID         *uint64 `json:"period_id"`
	PayslipID        *uint64 `json:"payslip_id"`
}

func (CommissionEntry) TableName() string {
	return "commission_entries"
}
