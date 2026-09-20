package payroll

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	SequencePayrollRunCode = "payroll_run"
	OriginPayrollRun       = "payroll_run"

	LeaveStateDraft     = "draft"
	LeaveStateSubmitted = "submitted"
	LeaveStateApproved  = "approved"
	LeaveStateRefused   = "refused"

	ContractStateActive = "active"
	ContractStateClosed = "closed"

	RunStateDraft     = "draft"
	RunStateConfirmed = "confirmed"
	RunStatePaid      = "paid"
	RunStateClosed    = "closed"

	PayslipStateDraft  = "draft"
	PayslipStatePosted = "posted"

	EmploymentTypeFullTime = "full_time"
	EmploymentTypePartTime = "part_time"
	EmploymentTypeContract = "contract"

	WageTypeMonthly = "monthly"
	WageTypeHourly  = "hourly"

	RuleCategoryEarning   = "earning"
	RuleCategoryDeduction = "deduction"

	ComputeTypeFixed   = "fixed"
	ComputeTypePercent = "percent"
	ComputeTypeFormula = "formula"

	AttendanceStatusDraft           = "draft"
	AttendanceStatusConfirmed       = "confirmed"
	AttendanceStatusCancelled       = "cancelled"
	AttendanceStatusPendingApproval = "pending_approval"
	AttendanceStatusAbsent          = "absent"

	AttendanceTypeNormal    = "normal"
	AttendanceTypeOvertime  = "overtime"
	AttendanceTypeWFH       = "wfh"
	AttendanceTypeHalfDay   = "half_day"
	AttendanceTypeSickLeave = "sick_leave"

	AttendanceSourceManual = "manual"
	AttendanceSourceSelf   = "self"
	AttendanceSourceDevice = "device"
	AttendanceSourceAPI    = "api"

	DefaultWorkHoursPerDay = 8.0
	OvertimeThresholdHours = 8.0
)

type Employee struct {
	model.Base
	OrganizationID   *uint64    `json:"organization_id"`
	ContactID        uint64     `json:"contact_id"`
	UserID           *uint64    `json:"user_id"`
	EmployeeNumber   string     `json:"employee_number"`
	DepartmentID     *uint64    `json:"department_id"`
	JobPositionID    *uint64    `json:"job_position_id"`
	ManagerID        *uint64    `json:"manager_id"`
	HireDate         *time.Time `gorm:"type:date" json:"hire_date"`
	TerminationDate  *time.Time `gorm:"type:date" json:"termination_date"`
	EmploymentType   string     `json:"employment_type"`
	WorkLocation     *string    `json:"work_location"`
	Active           bool       `gorm:"default:true" json:"active"`
	RequiresApproval bool       `gorm:"default:false" json:"requires_approval"`
}

type EmploymentContract struct {
	model.Base
	EmployeeID   uint64     `json:"employee_id"`
	DateStart    time.Time  `gorm:"type:date" json:"date_start"`
	DateEnd      *time.Time `gorm:"type:date" json:"date_end"`
	Wage         float64    `gorm:"type:numeric(18,4)" json:"wage"`
	WageType     string     `json:"wage_type"`
	CurrencyCode string     `gorm:"type:char(3)" json:"currency_code"`
	State        string     `json:"state"`
}

type LeaveRequest struct {
	model.Base
	EmployeeID  uint64    `json:"employee_id"`
	LeaveTypeID uint64    `json:"leave_type_id"`
	DateFrom    time.Time `gorm:"type:date" json:"date_from"`
	DateTo      time.Time `gorm:"type:date" json:"date_to"`
	Days        float64   `gorm:"type:numeric(6,2)" json:"days"`
	State       string    `json:"state"`
}

type Attendance struct {
	model.Base
	OrganizationID        uint64     `gorm:"not null" json:"organization_id"`
	EmployeeID            uint64     `gorm:"not null" json:"employee_id"`
	ShiftID               *uint64    `json:"shift_id"`
	CheckIn               *time.Time `json:"check_in"`
	CheckOut              *time.Time `json:"check_out"`
	WorkedHours           float64    `gorm:"type:numeric(8,2)" json:"worked_hours"`
	Status                string     `gorm:"size:20;not null;default:draft" json:"status"`
	AttendanceType        string     `gorm:"size:20;not null;default:normal" json:"attendance_type"`
	Source                string     `gorm:"size:30;not null;default:manual" json:"source"`
	BreakMinutes          int        `gorm:"not null;default:0" json:"break_minutes"`
	OvertimeHours         float64    `gorm:"type:numeric(8,2);not null;default:0" json:"overtime_hours"`
	LateMinutes           int        `gorm:"not null;default:0" json:"late_minutes"`
	EarlyDepartureMinutes int        `gorm:"not null;default:0" json:"early_departure_minutes"`
	Notes                 *string    `json:"notes"`
	ConfirmedAt           *time.Time `json:"confirmed_at"`
	Latitude              *float64   `gorm:"type:numeric(10,7)" json:"latitude"`
	Longitude             *float64   `gorm:"type:numeric(10,7)" json:"longitude"`
	DeviceID              *string    `gorm:"size:255" json:"device_id"`
	IPAddress             *string    `gorm:"size:45" json:"ip_address"`
	UserAgent             *string    `json:"user_agent"`
}

type Timesheet struct {
	model.Base
	EmployeeID  uint64    `json:"employee_id"`
	Date        time.Time `gorm:"type:date" json:"date"`
	ProjectID   *uint64   `json:"project_id"`
	TaskID      *uint64   `json:"task_id"`
	DimensionID *uint64   `json:"dimension_id"`
	Hours       float64   `gorm:"type:numeric(8,2)" json:"hours"`
	Description *string   `json:"description"`
}

type PayrollRun struct {
	model.Base
	OrganizationID uint64    `json:"organization_id"`
	Name           *string   `json:"name"`
	PeriodStart    time.Time `gorm:"type:date" json:"period_start"`
	PeriodEnd      time.Time `gorm:"type:date" json:"period_end"`
	State          string    `json:"state"`
}

type Payslip struct {
	model.Base
	RunID      uint64  `json:"run_id"`
	EmployeeID uint64  `json:"employee_id"`
	ContractID uint64  `json:"contract_id"`
	Gross      float64 `gorm:"type:numeric(18,4)" json:"gross"`
	Net        float64 `gorm:"type:numeric(18,4)" json:"net"`
	EntryID    *uint64 `json:"entry_id"`
	State      string  `json:"state"`
}

type PayslipLine struct {
	model.Base
	PayslipID uint64  `json:"payslip_id"`
	RuleID    uint64  `json:"rule_id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Amount    float64 `gorm:"type:numeric(18,4)" json:"amount"`
}

type Shift struct {
	model.Base
	OrganizationID uint64 `gorm:"not null" json:"organization_id"`
	Name           string `gorm:"size:100;not null" json:"name"`
	StartTime      string `gorm:"size:5;not null" json:"start_time"`
	EndTime        string `gorm:"size:5;not null" json:"end_time"`
}

type ShiftAssignment struct {
	model.Base
	EmployeeID uint64    `gorm:"not null" json:"employee_id"`
	ShiftID    uint64    `gorm:"not null" json:"shift_id"`
	Date       time.Time `gorm:"type:date;not null" json:"date"`
}

type AttendanceSummary struct {
	TotalWorkedHours   float64
	TotalOvertimeHours float64
	TotalLateMinutes   int
	TotalEarlyMinutes  int
	TotalDays          int
}
