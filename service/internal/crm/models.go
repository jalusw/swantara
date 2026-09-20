package crm

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	ProspectKindLead        = "prospect"
	ProspectKindOpportunity = "opportunity"
)

const (
	ProspectActivityTypeCall    = "call"
	ProspectActivityTypeMeeting = "meeting"
	ProspectActivityTypeEmail   = "email"
	ProspectActivityTypeNote    = "note"
	ProspectActivityTypeTask    = "task"
)

type Prospect struct {
	model.Base
	OrganizationID  *uint64    `json:"organization_id"`
	Name            string     `gorm:"not null" json:"name"`
	Type            string     `gorm:"type:text" json:"type"`
	ContactID       *uint64    `json:"contact_id"`
	ContactName     *string    `json:"contact_name"`
	Email           *string    `json:"email" audit:"redact"`
	Phone           *string    `json:"phone" audit:"redact"`
	JobPosition     *string    `json:"job_position"`
	StageID         *uint64    `json:"stage_id"`
	ExpectedRevenue float64    `gorm:"type:numeric(18,4)" json:"expected_revenue"`
	Probability     float64    `gorm:"type:numeric(5,2)" json:"probability"`
	Priority        int16      `gorm:"default:0" json:"priority"`
	SalespersonID   *uint64    `json:"salesperson_id"`
	SalesGroupID    *uint64    `json:"sales_group_id"`
	Source          *string    `json:"source"`
	Medium          *string    `json:"medium"`
	Campaign        *string    `json:"campaign"`
	LostReason      *string    `json:"lost_reason"`
	ExpectedClose   *time.Time `gorm:"type:date" json:"expected_close"`
	ClosedAt        *time.Time `json:"closed_at"`
}

func (Prospect) TableName() string {
	return "prospects"
}

type ProspectActivity struct {
	model.Base
	ProspectID *uint64    `json:"lead_id"`
	ContactID  *uint64    `json:"contact_id"`
	Type       string     `gorm:"type:text" json:"type"`
	Summary    string     `gorm:"type:text" json:"summary"`
	Note       *string    `json:"note"`
	DueDate    *time.Time `json:"due_date"`
	Done       bool       `gorm:"default:false" json:"done"`
	DoneAt     *time.Time `json:"done_at"`
	UserID     *uint64    `json:"user_id"`
}

func (ProspectActivity) TableName() string {
	return "prospect_activities"
}
