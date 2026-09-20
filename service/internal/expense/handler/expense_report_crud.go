package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/expense"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary List expense reports
// @Description Lists expense reports for the caller's organization.
// @Tags Expense
// @Accept json
// @Produce json
// @Success 200 {object} ListExpenseReportsResponseEnvelope "Expense reports retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Organization is required"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports [get]
func (h ExpenseReportHandler) List(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	reports, err := h.svc.List(c, *organizationID)
	if err != nil {
		httpx.RequestLog(c).Error("expense report list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve expense reports.", err)
	}

	items := make([]ExpenseReportResponse, len(reports))
	for i, report := range reports {
		items[i] = newExpenseReportResponse(report)
	}

	return httpx.CreateSuccessResponse(c, "Expense reports retrieved successfully.", ListExpenseReportsResponse{
		Reports: items,
	})
}

// @Summary Get expense report
// @Description Gets a single expense report with its lines for the caller's organization. A 404 is returned when the report does not exist or belongs to another organization.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense report ID"
// @Success 200 {object} GetExpenseReportResponseEnvelope "Expense report retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense report not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports/{id} [get]
func (h ExpenseReportHandler) Get(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	reportID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expense report id provided.", nil)
	}

	report, err := h.svc.Get(c, *organizationID, reportID)
	if err != nil {
		return writeExpenseError(c, err)
	}

	lines, err := h.svc.ListLines(c, *organizationID, reportID)
	if err != nil {
		return writeExpenseError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Expense report retrieved successfully.", GetExpenseReportResponseEnvelope{
		Data: withExpenseLines(newExpenseReportResponse(report), lines),
	})
}

// @Summary Create expense report
// @Description Creates a draft expense report with its lines for the caller's organization.
// @Tags Expense
// @Accept json
// @Produce json
// @Param body body CreateExpenseReportRequest true "Expense report details"
// @Success 201 {object} CreateExpenseReportResponseEnvelope "Expense report created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports [post]
func (h ExpenseReportHandler) Create(c fiber.Ctx) error {
	var request CreateExpenseReportRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	lines := make([]expense.CreateExpenseLineRequest, 0, len(request.Lines))
	for _, line := range request.Lines {
		expenseDate, err := time.Parse("2006-01-02", line.ExpenseDate)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expense date provided.", nil)
		}
		reimbursable := true
		if line.Reimbursable != nil {
			reimbursable = *line.Reimbursable
		}
		lines = append(lines, expense.CreateExpenseLineRequest{
			CategoryID:          line.CategoryID,
			ItemID:              line.ItemID,
			Description:         line.Description,
			ExpenseDate:         expenseDate,
			Quantity:            line.Quantity,
			UnitPrice:           line.UnitPrice,
			TaxIDs:              line.TaxIDs,
			CurrencyCode:        line.CurrencyCode,
			DimensionID:         line.DimensionID,
			ProjectID:           line.ProjectID,
			Reimbursable:        reimbursable,
			ReceiptAttachmentID: line.ReceiptAttachmentID,
		})
	}

	created, err := h.svc.Create(c, expense.CreateExpenseReportRequest{
		OrganizationID: *organizationID,
		Name:           request.Name,
		EmployeeID:     request.EmployeeID,
		PaymentMode:    request.PaymentMode,
		Lines:          lines,
	})
	if err != nil {
		return writeExpenseError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Expense report created successfully.", CreateExpenseReportResponseEnvelope{
		Data: newExpenseReportResponse(created),
	})
}
