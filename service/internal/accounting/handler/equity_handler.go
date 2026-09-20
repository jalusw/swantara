package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type EquityHandler struct {
	svc accounting.EquityService
}

func NewEquityHandler(svc accounting.EquityService) EquityHandler {
	return EquityHandler{svc: svc}
}

type EquityMovementResponse struct {
	AccountID    uint64  `json:"account_id"`
	AccountCode  string  `json:"account_code"`
	AccountName  string  `json:"account_name"`
	AccountType  string  `json:"account_type"`
	MovementType string  `json:"movement_type"`
	Amount       float64 `json:"amount"`
}

type EquityRollforwardResponse struct {
	OpeningBalance       []EquityMovementResponse `json:"opening_balance"`
	CapitalContributions []EquityMovementResponse `json:"capital_contributions"`
	NetIncome            float64                  `json:"net_income"`
	Dividends            []EquityMovementResponse `json:"dividends"`
	OtherChanges         []EquityMovementResponse `json:"other_changes"`
	ClosingBalance       []EquityMovementResponse `json:"closing_balance"`
	PeriodID             uint64                   `json:"period_id"`
	Balanced             bool                     `json:"balanced"`
}

type EquityResponseEnvelope struct {
	httpx.EnvelopeBase
	Data EquityRollforwardResponse `json:"data"`
}

// @Summary Generate equity rollforward
// @Description Generates the equity rollforward for a tax period including contributions, net income, and dividends
// @Tags Accounting Reports
// @Produce json
// @Param id path integer true "Tax period ID"
// @Success 200 {object} EquityResponseEnvelope "Equity rollforward generated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Tax period not found"
// @Failure 422 {object} httpx.ErrorResponse "Invalid tax period"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /tax-periods/{id}/equity [get]
func (h EquityHandler) Generate(c fiber.Ctx) error {
	periodID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid tax period id provided.", nil)
	}

	orgID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	eq, err := h.svc.Generate(c, orgID, periodID)
	if err != nil {
		return writeEquityError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Equity rollforward generated successfully.", EquityRollforwardResponse{
		OpeningBalance:       toEquityMovements(eq.OpeningBalance),
		CapitalContributions: toEquityMovements(eq.CapitalContributions),
		NetIncome:            eq.NetIncome,
		Dividends:            toEquityMovements(eq.Dividends),
		OtherChanges:         toEquityMovements(eq.OtherChanges),
		ClosingBalance:       toEquityMovements(eq.ClosingBalance),
		PeriodID:             eq.PeriodID,
		Balanced:             eq.Balanced,
	})
}

func toEquityMovements(movements []accounting.EquityMovement) []EquityMovementResponse {
	result := make([]EquityMovementResponse, len(movements))
	for i, m := range movements {
		result[i] = EquityMovementResponse{
			AccountID:    m.AccountID,
			AccountCode:  m.AccountCode,
			AccountName:  m.AccountName,
			AccountType:  m.AccountType,
			MovementType: m.MovementType,
			Amount:       m.Amount,
		}
	}
	return result
}

func writeEquityError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax period not found.")
	case errors.Is(err, accounting.ErrInvalidPeriod):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax period date range is invalid.", nil)
	default:
		httpx.RequestLog(c).Error("equity rollforward generation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to generate equity rollforward.", err)
	}
}
