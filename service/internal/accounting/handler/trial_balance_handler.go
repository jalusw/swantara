package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type TrialBalanceHandler struct {
	svc accounting.TrialBalanceService
}

func NewTrialBalanceHandler(svc accounting.TrialBalanceService) TrialBalanceHandler {
	return TrialBalanceHandler{svc: svc}
}

type TrialBalanceLineResponse struct {
	AccountID   uint64        `json:"account_id"`
	AccountCode string        `json:"account_code"`
	AccountName string        `json:"account_name"`
	AccountType string        `json:"account_type"`
	Debit       amount.Amount `json:"debit"`
	Credit      amount.Amount `json:"credit"`
	Balance     amount.Amount `json:"balance"`
}

type TrialBalanceResponse struct {
	Lines       []TrialBalanceLineResponse `json:"lines"`
	TotalDebit  amount.Amount              `json:"total_debit"`
	TotalCredit amount.Amount              `json:"total_credit"`
	PeriodID    uint64                     `json:"period_id"`
	Balanced    bool                       `json:"balanced"`
	Difference  amount.Amount              `json:"difference"`
}

type TrialBalanceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data TrialBalanceResponse `json:"data"`
}

// @Summary Generate trial balance
// @Description Generates a trial balance for the specified tax period, listing all accounts with their debit, credit, and net balance totals for posted journal entries within the period.
// @Tags Trial Balance
// @Accept json
// @Produce json
// @Param id path integer true "Fiscal Period ID"
// @Success 200 {object} TrialBalanceResponseEnvelope "Trial balance generated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/tax-periods/{id}/trial-balance [get]
func (h TrialBalanceHandler) Generate(c fiber.Ctx) error {
	orgID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}
	asOf := c.Query("as_of")
	includeZero := c.Query("include_zero") == "true"
	rollup := c.Query("rollup") == "true"
	if asOf != "" {
		tb, err := h.svc.GenerateAsOf(c, orgID, asOf, includeZero)
		if err != nil {
			return writeTrialBalanceError(c, err)
		}
		tbLines := tb.Lines
		if rollup {
			tbLines, err = h.svc.Rollup(c, orgID, tbLines)
			if err != nil {
				return writeTrialBalanceError(c, err)
			}
		}
		lines := make([]TrialBalanceLineResponse, len(tbLines))
		for i, line := range tbLines {
			lines[i] = TrialBalanceLineResponse{
				AccountID:   line.AccountID,
				AccountCode: line.AccountCode,
				AccountName: line.AccountName,
				AccountType: line.AccountType,
				Debit:       line.Debit,
				Credit:      line.Credit,
				Balance:     line.Balance,
			}
		}
		return httpx.CreateSuccessResponse(c, "Trial balance generated successfully.", TrialBalanceResponse{
			Lines:       lines,
			TotalDebit:  tb.TotalDebit,
			TotalCredit: tb.TotalCredit,
			Balanced:    tb.Balanced,
			Difference:  tb.Difference,
		})
	}
	if periodIDStr := c.Query("period_id"); periodIDStr != "" {
		periodID, err := strconv.ParseUint(periodIDStr, 10, 64)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid period_id provided.", nil)
		}
		tb, err := h.svc.Generate(c, orgID, periodID)
		if err != nil {
			return writeTrialBalanceError(c, err)
		}
		tbLines := tb.Lines
		if rollup {
			tbLines, err = h.svc.Rollup(c, orgID, tbLines)
			if err != nil {
				return writeTrialBalanceError(c, err)
			}
		}
		lines := make([]TrialBalanceLineResponse, len(tbLines))
		for i, line := range tbLines {
			lines[i] = TrialBalanceLineResponse{
				AccountID:   line.AccountID,
				AccountCode: line.AccountCode,
				AccountName: line.AccountName,
				AccountType: line.AccountType,
				Debit:       line.Debit,
				Credit:      line.Credit,
				Balance:     line.Balance,
			}
		}
		return httpx.CreateSuccessResponse(c, "Trial balance generated successfully.", TrialBalanceResponse{
			Lines:       lines,
			TotalDebit:  tb.TotalDebit,
			TotalCredit: tb.TotalCredit,
			PeriodID:    tb.PeriodID,
			Balanced:    tb.Balanced,
			Difference:  tb.Difference,
		})
	}
	periodID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax period id provided.", nil)
	}

	tb, err := h.svc.Generate(c, orgID, periodID)
	if err != nil {
		return writeTrialBalanceError(c, err)
	}
	tbLines := tb.Lines
	if rollup {
		tbLines, err = h.svc.Rollup(c, orgID, tbLines)
		if err != nil {
			return writeTrialBalanceError(c, err)
		}
	}

	lines := make([]TrialBalanceLineResponse, len(tbLines))
	for i, line := range tbLines {
		lines[i] = TrialBalanceLineResponse{
			AccountID:   line.AccountID,
			AccountCode: line.AccountCode,
			AccountName: line.AccountName,
			AccountType: line.AccountType,
			Debit:       line.Debit,
			Credit:      line.Credit,
			Balance:     line.Balance,
		}
	}

	return httpx.CreateSuccessResponse(c, "Trial balance generated successfully.", TrialBalanceResponse{
		Lines:       lines,
		TotalDebit:  tb.TotalDebit,
		TotalCredit: tb.TotalCredit,
		PeriodID:    tb.PeriodID,
		Balanced:    tb.Balanced,
		Difference:  tb.Difference,
	})
}

func writeTrialBalanceError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax period not found.")
	case errors.Is(err, accounting.ErrInvalidPeriod):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax period date range is invalid.", nil)
	default:
		httpx.RequestLog(c).Error("trial balance generation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to generate trial balance.", err)
	}
}
