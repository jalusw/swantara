package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

type StatementHandler struct {
	svc reporting.StatementService
}

func NewStatementHandler(svc reporting.StatementService) StatementHandler {
	return StatementHandler{svc: svc}
}

// @Summary Profit and loss
// @Description Returns the P&L statement for a tax period: revenue, COGS, expenses, and net income per income-statement account, reconciling to the trial balance.
// @Tags Reporting
// @Accept json
// @Produce json
// @Param period_id query integer true "Tax period ID"
// @Success 200 {object} reporting.ProfitAndLoss "Profit and loss retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reports/profit-and-loss [get]
func (h StatementHandler) ProfitAndLoss(c fiber.Ctx) error {
	periodID, err := strconv.ParseUint(c.Query("period_id"), 10, 64)
	if err != nil || periodID == 0 {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid period_id provided.", nil)
	}
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	statement, err := h.svc.ProfitAndLoss(c, *organizationID, periodID)
	if err != nil {
		return writeStatementError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Profit and loss retrieved successfully.", statement)
}

// @Summary Balance sheet
// @Description Returns the balance sheet as of a date: asset, liability, and equity balances including current earnings, reconciling to the trial balance.
// @Tags Reporting
// @Accept json
// @Produce json
// @Param as_of query string false "As-of date (YYYY-MM-DD), defaults to today"
// @Success 200 {object} reporting.BalanceSheet "Balance sheet retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reports/balance-sheet [get]
func (h StatementHandler) BalanceSheet(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	asOf := time.Now().UTC()
	if raw := c.Query("as_of"); raw != "" {
		parsed, err := helper.ParseDate(&raw)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "as_of must be in YYYY-MM-DD format.", nil)
		}
		asOf = *parsed
	}

	sheet, err := h.svc.BalanceSheet(c, *organizationID, asOf)
	if err != nil {
		return writeStatementError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Balance sheet retrieved successfully.", sheet)
}

// @Summary Cash flow statement
// @Description Returns the cash flow statement for a date range: operating, investing, and financing cash flows from the general ledger cash and bank accounts.
// @Tags Reporting
// @Accept json
// @Produce json
// @Param start query string true "Start date (YYYY-MM-DD)"
// @Param end query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} reporting.CashFlow "Cash flow statement retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/reports/cash-flow [get]
func (h StatementHandler) CashFlow(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	startRaw := c.Query("start")
	endRaw := c.Query("end")
	if startRaw == "" || endRaw == "" {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "start and end must be in YYYY-MM-DD format.", nil)
	}
	start, err := helper.ParseDate(&startRaw)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "start must be in YYYY-MM-DD format.", nil)
	}
	end, err := helper.ParseDate(&endRaw)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "end must be in YYYY-MM-DD format.", nil)
	}
	if end.Before(*start) {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "end must be on or after start.", nil)
	}

	flow, err := h.svc.CashFlow(c, *organizationID, *start, *end)
	if err != nil {
		return writeStatementError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Cash flow statement retrieved successfully.", flow)
}

type YearEndRollRequest struct {
	OrganizationID            *uint64 `json:"organization_id"`
	TaxYearID                 uint64  `json:"tax_year_id" validate:"required,gt=0"`
	JournalID                 uint64  `json:"journal_id" validate:"required,gt=0"`
	RetainedEarningsAccountID uint64  `json:"retained_earnings_account_id" validate:"required,gt=0"`
}

// @Summary Year-end retained earnings roll
// @Description Posts the year-end closing entry that rolls net income into retained earnings, zeroing the income statement accounts. Guards against double runs.
// @Tags Period Close
// @Accept json
// @Produce json
// @Param body body YearEndRollRequest true "Year-end roll details"
// @Success 200 {object} reporting.YearEndRollResult "Year-end roll posted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax year or period not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/period-close/year-end-roll [post]
func (h StatementHandler) YearEndRoll(c fiber.Ctx) error {
	var request YearEndRollRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	result, err := h.svc.YearEndRoll(c, reporting.YearEndRollRequest{
		OrganizationID:            *organizationID,
		TaxYearID:                 request.TaxYearID,
		JournalID:                 request.JournalID,
		RetainedEarningsAccountID: request.RetainedEarningsAccountID,
	})
	if err != nil {
		return writeStatementError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Year-end retained earnings roll posted successfully.", result)
}

func writeStatementError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reporting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax period or year not found.")
	case errors.Is(err, reporting.ErrYearEndAlreadyClosed):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Year-end retained earnings roll has already been posted for this period.", nil)
	case errors.Is(err, reporting.ErrPeriodLocked):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax period is locked.", nil)
	default:
		httpx.RequestLog(c).Error("statement query failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to generate the statement.", err)
	}
}
