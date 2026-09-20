package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/expense"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Submit expense report
// @Description Transitions a draft expense report into the submitted state for approval.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense report ID"
// @Success 200 {object} TransitionExpenseReportResponseEnvelope "Expense report submitted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense report not found"
// @Failure 409 {object} httpx.ErrorResponse "Expense report state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports/{id}/submit [post]
func (h ExpenseReportHandler) Submit(c fiber.Ctx) error {
	return h.transition(c, func(c fiber.Ctx, organizationID, reportID uint64) (*expense.ExpenseReport, error) {
		return h.svc.Submit(c, organizationID, reportID)
	})
}

// @Summary Approve expense report
// @Description Approves a submitted expense report, allowing it to be posted.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense report ID"
// @Success 200 {object} TransitionExpenseReportResponseEnvelope "Expense report approved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense report not found"
// @Failure 409 {object} httpx.ErrorResponse "Expense report state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports/{id}/approve [post]
func (h ExpenseReportHandler) Approve(c fiber.Ctx) error {
	return h.transition(c, func(c fiber.Ctx, organizationID, reportID uint64) (*expense.ExpenseReport, error) {
		approverID, _ := httpx.CallerID(c)
		return h.svc.Approve(c, organizationID, reportID, approverID)
	})
}

// @Summary Refuse expense report
// @Description Refuses a submitted expense report, preventing posting.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense report ID"
// @Success 200 {object} TransitionExpenseReportResponseEnvelope "Expense report refused successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense report not found"
// @Failure 409 {object} httpx.ErrorResponse "Expense report state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports/{id}/refuse [post]
func (h ExpenseReportHandler) Refuse(c fiber.Ctx) error {
	return h.transition(c, func(c fiber.Ctx, organizationID, reportID uint64) (*expense.ExpenseReport, error) {
		return h.svc.Refuse(c, organizationID, reportID)
	})
}

// @Summary Post expense report
// @Description Posts an approved expense report to the general ledger, recording expense, input tax, and payable or card clearing accounts.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense report ID"
// @Success 200 {object} TransitionExpenseReportResponseEnvelope "Expense report posted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense report not found"
// @Failure 409 {object} httpx.ErrorResponse "Expense report state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports/{id}/post [post]
func (h ExpenseReportHandler) Post(c fiber.Ctx) error {
	return h.transition(c, func(c fiber.Ctx, organizationID, reportID uint64) (*expense.ExpenseReport, error) {
		return h.svc.Post(c, organizationID, reportID)
	})
}

// @Summary Reimburse expense report
// @Description Reimburses a posted own-account expense report, clearing the employee payable against the bank.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense report ID"
// @Success 200 {object} TransitionExpenseReportResponseEnvelope "Expense report reimbursed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense report not found"
// @Failure 409 {object} httpx.ErrorResponse "Expense report state is invalid for this operation"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports/{id}/reimburse [post]
func (h ExpenseReportHandler) Reimburse(c fiber.Ctx) error {
	return h.transition(c, func(c fiber.Ctx, organizationID, reportID uint64) (*expense.ExpenseReport, error) {
		return h.svc.Reimburse(c, organizationID, reportID)
	})
}

// @Summary Bill expense report to customer invoice
// @Description Creates a customer invoice from the billable lines of a posted expense report.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense report ID"
// @Success 201 {object} BillExpenseReportResponseEnvelope "Expense report billed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense report not found"
// @Failure 409 {object} httpx.ErrorResponse "Expense report state is invalid for this operation"
// @Failure 422 {object} httpx.ErrorResponse "Expense report has no billable lines"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/expense-reports/{id}/bill [post]
func (h ExpenseReportHandler) Bill(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	reportID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expense report id provided.", nil)
	}

	invoice, err := h.svc.BillToInvoice(c, *organizationID, reportID)
	if err != nil {
		return writeExpenseError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Expense report billed successfully.", BillExpenseReportResponseEnvelope{
		Data: BillExpenseReportResponse{InvoiceID: invoice.ID},
	})
}

func (h ExpenseReportHandler) transition(c fiber.Ctx, transition func(c fiber.Ctx, organizationID, reportID uint64) (*expense.ExpenseReport, error)) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	reportID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expense report id provided.", nil)
	}

	updated, err := transition(c, *organizationID, reportID)
	if err != nil {
		return writeExpenseError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Expense report updated successfully.", TransitionExpenseReportResponseEnvelope{
		Data: newExpenseReportResponse(updated),
	})
}

func writeExpenseError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, expense.ErrExpenseReportNotFound):
		return httpx.CreateNotFoundResponse(c, "Expense report not found.")
	case errors.Is(err, expense.ErrExpenseReportState):
		return httpx.CreateConflictResponse(c, "Expense report state is invalid for this operation.", err)
	case errors.Is(err, expense.ErrExpenseNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Expense report must have at least one line.", nil)
	case errors.Is(err, expense.ErrExpenseNoCategory):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Expense line category does not exist.", nil)
	case errors.Is(err, expense.ErrExpenseNoAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Expense category has no expense account configured.", nil)
	case errors.Is(err, expense.ErrExpenseConfig):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Expense posting is not configured for this organization.", nil)
	case errors.Is(err, expense.ErrExpenseNoBillable):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Expense report has no billable lines.", nil)
	case errors.Is(err, expense.ErrExpenseBillableProject):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Billable expense lines must reference a single project.", nil)
	case errors.Is(err, expense.ErrExpenseProjectNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Billable expense project does not exist.", nil)
	case errors.Is(err, expense.ErrExpenseNoIncomeAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Billable expense lines require a item with an income account.", nil)
	case errors.Is(err, expense.ErrExpenseInvalidLine):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Expense report has invalid data.", nil)
	default:
		httpx.RequestLog(c).Error("expense report write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save expense report.", err)
	}
}
