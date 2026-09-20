package project

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	BillingTypeFixed        = "fixed"
	BillingTypeTimeMaterial = "time_material"
	BillingTypeMilestone    = "milestone"

	ProjectStateDraft     = "draft"
	ProjectStateOpen      = "open"
	ProjectStateClosed    = "closed"
	ProjectStateCancelled = "cancelled"

	TaskStageBacklog    = "backlog"
	TaskStageTodo       = "todo"
	TaskStageInProgress = "in_progress"
	TaskStageDone       = "done"

	StandardMonthlyHours = 173.0
)

type Project struct {
	model.Base
	OrganizationID uint64     `json:"organization_id"`
	Name           string     `json:"name"`
	ContactID      uint64     `json:"contact_id"`
	ManagerID      *uint64    `json:"manager_id"`
	DimensionID    *uint64    `json:"dimension_id"`
	SaleOrderID    *uint64    `json:"sale_order_id"`
	BillingType    string     `json:"billing_type"`
	BillableRate   float64    `gorm:"type:numeric(18,4)" json:"billable_rate"`
	DateStart      *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd        *time.Time `gorm:"type:date" json:"date_end"`
	State          string     `json:"state"`
}

type ProjectTask struct {
	model.Base
	ProjectID      uint64     `json:"project_id"`
	Name           string     `json:"name"`
	AssigneeID     *uint64    `json:"assignee_id"`
	Stage          string     `json:"stage"`
	PlannedHours   float64    `gorm:"type:numeric(8,2)" json:"planned_hours"`
	EffectiveHours float64    `gorm:"type:numeric(8,2)" json:"effective_hours"`
	ParentTaskID   *uint64    `json:"parent_task_id"`
	Deadline       *time.Time `gorm:"type:date" json:"deadline"`
	Priority       *int       `json:"priority"`
}

type ProjectMilestone struct {
	model.Base
	ProjectID  uint64     `json:"project_id"`
	Name       string     `json:"name"`
	Deadline   *time.Time `gorm:"type:date" json:"deadline"`
	Reached    bool       `json:"reached"`
	SaleLineID *uint64    `json:"sale_line_id"`
}

type ProjectInvoiceLine struct {
	model.Base
	ProjectID     uint64  `json:"project_id"`
	InvoiceID     uint64  `json:"invoice_id"`
	InvoiceLineID *uint64 `json:"invoice_line_id"`
	TimesheetID   uint64  `json:"timesheet_id"`
	Qty           float64 `gorm:"type:numeric(18,4)" json:"qty"`
	UnitPrice     float64 `gorm:"type:numeric(18,4)" json:"unit_price"`
	Amount        float64 `gorm:"type:numeric(18,4)" json:"amount"`
}

func (Project) TableName() string            { return "projects" }
func (ProjectTask) TableName() string        { return "project_tasks" }
func (ProjectMilestone) TableName() string   { return "project_milestones" }
func (ProjectInvoiceLine) TableName() string { return "project_invoice_lines" }
