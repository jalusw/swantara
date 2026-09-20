package expense

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	SequenceExpenseReportCode = "expense_report"

	ExpenseStateDraft      = "draft"
	ExpenseStateSubmitted  = "submitted"
	ExpenseStateApproved   = "approved"
	ExpenseStateRefused    = "refused"
	ExpenseStatePosted     = "posted"
	ExpenseStateReimbursed = "reimbursed"

	ExpensePaymentOwnAccount = "own_account"
	ExpensePaymentOrgAccount = "organization_account"
)

type ExpenseReport struct {
	model.Base
	OrganizationID       *uint64    `json:"organization_id"`
	Name                 string     `json:"name"`
	EmployeeID           uint64     `json:"employee_id"`
	State                string     `json:"state"`
	PaymentMode          string     `json:"payment_mode"`
	TotalAmount          float64    `gorm:"type:numeric(18,4)" json:"total_amount"`
	EntryID              *uint64    `json:"entry_id"`
	ReimbursementEntryID *uint64    `json:"reimbursement_entry_id"`
	SubmittedAt          *time.Time `json:"submitted_at"`
	ApprovedBy           *uint64    `json:"approved_by"`
}

func (ExpenseReport) TableName() string { return "expense_reports" }

type ExpenseLine struct {
	model.Base
	ReportID            uint64            `json:"report_id"`
	EmployeeID          *uint64           `json:"employee_id"`
	CategoryID          *uint64           `json:"category_id"`
	ItemID              *uint64           `json:"item_id"`
	Description         *string           `json:"description"`
	ExpenseDate         *time.Time        `gorm:"type:date" json:"expense_date"`
	Quantity            float64           `gorm:"type:numeric(18,4);default:1" json:"quantity"`
	UnitPrice           float64           `gorm:"type:numeric(18,4)" json:"unit_price"`
	Amount              float64           `gorm:"type:numeric(18,4)" json:"amount"`
	TaxIDs              helper.Int64Array `gorm:"type:bigint[]" json:"tax_ids"`
	CurrencyCode        *string           `gorm:"type:char(3)" json:"currency_code"`
	DimensionID         *uint64           `json:"dimension_id"`
	ProjectID           *uint64           `json:"project_id"`
	Reimbursable        bool              `gorm:"default:true" json:"reimbursable"`
	ReceiptAttachmentID *uint64           `json:"receipt_attachment_id"`
}

func (ExpenseLine) TableName() string { return "expense_lines" }
