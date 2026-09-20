package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CashFlowHandler struct {
	svc accounting.CashFlowService
}

func NewCashFlowHandler(svc accounting.CashFlowService) CashFlowHandler {
	return CashFlowHandler{svc: svc}
}

type CashFlowLineResponse struct {
	AccountID       uint64  `json:"account_id"`
	AccountCode     string  `json:"account_code"`
	Name            string  `json:"name"`
	Amount          float64 `json:"amount"`
	CashFlowSection string  `json:"cash_flow_section"`
	OriginType      string  `json:"origin_type"`
}

type CashFlowResponse struct {
	Operating   []CashFlowLineResponse `json:"operating"`
	Investing   []CashFlowLineResponse `json:"investing"`
	Financing   []CashFlowLineResponse `json:"financing"`
	NetChange   float64                `json:"net_change"`
	OpeningCash float64                `json:"opening_cash"`
	ClosingCash float64                `json:"closing_cash"`
	Balanced    bool                   `json:"balanced"`
}

type CashFlowResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CashFlowResponse `json:"data"`
}

// @Summary Generate cash flow statement
// @Description Generates the cash flow statement grouped by operating, investing, and financing sections for a date range
// @Tags Accounting Reports
// @Produce json
// @Param date_start query string true "Start date (YYYY-MM-DD)"
// @Param date_end query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} CashFlowResponseEnvelope "Cash flow statement generated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Missing date range"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid date format"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /reports/cash-flow [get]
func (h CashFlowHandler) Generate(c fiber.Ctx) error {
	dateStartStr := c.Query("date_start")
	dateEndStr := c.Query("date_end")
	if dateStartStr == "" || dateEndStr == "" {
		return httpx.CreateBadRequestResponse(c, "date_start and date_end are required", nil)
	}

	dateStart, err := time.Parse("2006-01-02", dateStartStr)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date_start format. Use YYYY-MM-DD.", nil)
	}

	dateEnd, err := time.Parse("2006-01-02", dateEndStr)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date_end format. Use YYYY-MM-DD.", nil)
	}

	orgID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	cf, err := h.svc.Generate(c, orgID, dateStart, dateEnd)
	if err != nil {
		return writeCashFlowError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Cash flow statement generated successfully.", CashFlowResponse{
		Operating:   toCashFlowLines(cf.Operating),
		Investing:   toCashFlowLines(cf.Investing),
		Financing:   toCashFlowLines(cf.Financing),
		NetChange:   cf.NetChange,
		OpeningCash: cf.OpeningCash,
		ClosingCash: cf.ClosingCash,
		Balanced:    cf.Balanced,
	})
}

func toCashFlowLines(lines []accounting.CashFlowLine) []CashFlowLineResponse {
	result := make([]CashFlowLineResponse, len(lines))
	for i, l := range lines {
		result[i] = CashFlowLineResponse{
			AccountID:       l.AccountID,
			AccountCode:     l.AccountCode,
			Name:            l.Name,
			Amount:          l.Amount,
			CashFlowSection: l.CashFlowSection,
			OriginType:      l.OriginType,
		}
	}
	return result
}

func writeCashFlowError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax period not found.")
	case errors.Is(err, accounting.ErrInvalidPeriod):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax period date range is invalid.", nil)
	default:
		httpx.RequestLog(c).Error("cash flow generation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to generate cash flow statement.", err)
	}
}
