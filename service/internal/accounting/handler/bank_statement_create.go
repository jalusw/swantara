package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type CreateBankStatementLineRequest struct {
	Date           *string `json:"date"`
	Amount         float64 `json:"amount" validate:"required"`
	CurrencyCode   *string `json:"currency_code" validate:"omitempty,len=3"`
	FeeAmount      float64 `json:"fee_amount" validate:"gte=0"`
	InterestAmount float64 `json:"interest_amount" validate:"gte=0"`
	ContactID      *uint64 `json:"contact_id"`
	Ref            string  `json:"ref"`
	Narration      string  `json:"narration"`
}

type CreateBankStatementRequest struct {
	OrganizationID *uint64                          `json:"organization_id"`
	JournalID      uint64                           `json:"journal_id" validate:"required,gt=0"`
	Date           *string                          `json:"date"`
	BalanceStart   float64                          `json:"balance_start"`
	BalanceEnd     float64                          `json:"balance_end"`
	CurrencyCode   *string                          `json:"currency_code" validate:"omitempty,len=3"`
	Lines          []CreateBankStatementLineRequest `json:"lines" validate:"required,min=1,dive"`
}

type CreateBankStatementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateBankStatementResponse `json:"data"`
}
type CreateBankStatementResponse struct {
	BankStatement BankStatementResponse `json:"bank_statement"`
}

// @Summary Create bank statement
// @Description Creates an open bank statement against a journal with its statement lines, requiring at least one line and valid statement and line dates. Opening and closing balances are recorded alongside the statement header, which starts in the open state.
// @Tags Bank Statements
// @Accept json
// @Produce json
// @Param body body CreateBankStatementRequest true "Bank statement details"
// @Success 201 {object} CreateBankStatementResponseEnvelope "Bank statement created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/bank-statements [post]
func (h BankStatementHandler) Create(c fiber.Ctx) error {
	var request CreateBankStatementRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	statementDate, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	date := time.Time{}
	if statementDate != nil {
		date = *statementDate
	}

	lines := make([]accounting.BankStatementLineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lineDate, err := helper.ParseDate(line.Date)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Line date must be in YYYY-MM-DD format.", nil)
		}
		lines[i] = accounting.BankStatementLineRequest{
			Date:           lineDate,
			Amount:         line.Amount,
			CurrencyCode:   line.CurrencyCode,
			FeeAmount:      line.FeeAmount,
			InterestAmount: line.InterestAmount,
			ContactID:      line.ContactID,
			Ref:            line.Ref,
			Narration:      line.Narration,
		}
	}

	statement, err := h.svc.Create(c, accounting.CreateBankStatementRequest{
		OrganizationID: *organizationID,
		JournalID:      request.JournalID,
		Date:           date,
		BalanceStart:   request.BalanceStart,
		BalanceEnd:     request.BalanceEnd,
		CurrencyCode:   request.CurrencyCode,
		Lines:          lines,
	})
	if err != nil {
		return writeBankStatementError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Bank statement created successfully.", CreateBankStatementResponse{
		BankStatement: newBankStatementResponse(statement),
	})
}
