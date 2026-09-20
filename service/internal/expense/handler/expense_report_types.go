package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/expense"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ExpenseReportHandler struct {
	svc expense.ExpenseService
}

func NewExpenseReportHandler(svc expense.ExpenseService) ExpenseReportHandler {
	return ExpenseReportHandler{svc: svc}
}

type ExpenseLineResponse struct {
	ID                  uint64            `json:"id"`
	CategoryID          *uint64           `json:"category_id"`
	ItemID              *uint64           `json:"item_id"`
	Description         *string           `json:"description"`
	ExpenseDate         *time.Time        `json:"expense_date"`
	Quantity            float64           `json:"quantity"`
	UnitPrice           float64           `json:"unit_price"`
	Amount              float64           `json:"amount"`
	TaxIDs              helper.Int64Array `json:"tax_ids"`
	DimensionID         *uint64           `json:"dimension_id"`
	ProjectID           *uint64           `json:"project_id"`
	Reimbursable        bool              `json:"reimbursable"`
	ReceiptAttachmentID *uint64           `json:"receipt_attachment_id"`
}

type ExpenseReportResponse struct {
	ID          uint64                `json:"id"`
	Name        string                `json:"name"`
	EmployeeID  uint64                `json:"employee_id"`
	State       string                `json:"state"`
	PaymentMode string                `json:"payment_mode"`
	TotalAmount float64               `json:"total_amount"`
	MovementID  *uint64               `json:"movement_id"`
	SubmittedAt *time.Time            `json:"submitted_at"`
	ApprovedBy  *uint64               `json:"approved_by"`
	Lines       []ExpenseLineResponse `json:"lines,omitempty"`
}

func newExpenseReportResponse(report *expense.ExpenseReport) ExpenseReportResponse {
	return ExpenseReportResponse{
		ID:          report.ID,
		Name:        report.Name,
		EmployeeID:  report.EmployeeID,
		State:       report.State,
		PaymentMode: report.PaymentMode,
		TotalAmount: report.TotalAmount,
		MovementID:  report.EntryID,
		SubmittedAt: report.SubmittedAt,
		ApprovedBy:  report.ApprovedBy,
	}
}

func withExpenseLines(report ExpenseReportResponse, lines []*expense.ExpenseLine) ExpenseReportResponse {
	for _, line := range lines {
		report.Lines = append(report.Lines, ExpenseLineResponse{
			ID:                  line.ID,
			CategoryID:          line.CategoryID,
			ItemID:              line.ItemID,
			Description:         line.Description,
			ExpenseDate:         line.ExpenseDate,
			Quantity:            line.Quantity,
			UnitPrice:           line.UnitPrice,
			Amount:              line.Amount,
			TaxIDs:              line.TaxIDs,
			DimensionID:         line.DimensionID,
			ProjectID:           line.ProjectID,
			Reimbursable:        line.Reimbursable,
			ReceiptAttachmentID: line.ReceiptAttachmentID,
		})
	}
	return report
}

type ListExpenseReportsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListExpenseReportsResponse `json:"data"`
}
type ListExpenseReportsResponse struct {
	Reports []ExpenseReportResponse `json:"reports"`
}

type GetExpenseReportResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ExpenseReportResponse `json:"data"`
}

type CreateExpenseLineRequest struct {
	CategoryID          *uint64           `json:"category_id" validate:"required"`
	ItemID              *uint64           `json:"item_id"`
	Description         string            `json:"description"`
	ExpenseDate         string            `json:"expense_date" validate:"required"`
	Quantity            float64           `json:"quantity"`
	UnitPrice           float64           `json:"unit_price" validate:"required"`
	TaxIDs              helper.Int64Array `json:"tax_ids"`
	CurrencyCode        string            `json:"currency_code"`
	DimensionID         *uint64           `json:"dimension_id"`
	ProjectID           *uint64           `json:"project_id"`
	Reimbursable        *bool             `json:"reimbursable"`
	ReceiptAttachmentID *uint64           `json:"receipt_attachment_id"`
}

type CreateExpenseReportRequest struct {
	Name        string                     `json:"name" validate:"required"`
	EmployeeID  uint64                     `json:"employee_id" validate:"required"`
	PaymentMode string                     `json:"payment_mode" validate:"required"`
	Lines       []CreateExpenseLineRequest `json:"lines" validate:"required"`
}

type CreateExpenseReportResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ExpenseReportResponse `json:"data"`
}

type TransitionExpenseReportResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ExpenseReportResponse `json:"data"`
}

type BillExpenseReportResponseEnvelope struct {
	httpx.EnvelopeBase
	Data BillExpenseReportResponse `json:"data"`
}
type BillExpenseReportResponse struct {
	InvoiceID uint64 `json:"invoice_id"`
}
