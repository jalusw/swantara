package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type BankStatementHandler struct {
	svc accounting.BankStatementService
}

func NewBankStatementHandler(svc accounting.BankStatementService) BankStatementHandler {
	return BankStatementHandler{svc: svc}
}

type BankStatementLineResponse struct {
	ID              uint64    `json:"id"`
	StatementID     uint64    `json:"statement_id"`
	Date            *string   `json:"date"`
	Amount          float64   `json:"amount"`
	CurrencyCode    *string   `json:"currency_code"`
	FeeAmount       float64   `json:"fee_amount"`
	InterestAmount  float64   `json:"interest_amount"`
	ContactID       *uint64   `json:"contact_id"`
	Ref             *string   `json:"ref"`
	Narration       *string   `json:"narration"`
	Reconciled      bool      `json:"reconciled"`
	PaymentID       *uint64   `json:"payment_id"`
	JournalLineID   *uint64   `json:"journal_line_id"`
	FeeEntryID      *uint64   `json:"fee_entry_id"`
	InterestEntryID *uint64   `json:"interest_entry_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func newBankStatementLineResponse(line *accounting.BankStatementLine) BankStatementLineResponse {
	return BankStatementLineResponse{
		ID:              line.ID,
		StatementID:     line.StatementID,
		Date:            helper.FormatDatePtr(line.Date),
		Amount:          line.Amount.Float64(),
		CurrencyCode:    line.CurrencyCode,
		FeeAmount:       line.FeeAmount.Float64(),
		InterestAmount:  line.InterestAmount.Float64(),
		ContactID:       line.ContactID,
		Ref:             line.Ref,
		Narration:       line.Narration,
		Reconciled:      line.Reconciled,
		PaymentID:       line.PaymentID,
		JournalLineID:   line.JournalLineID,
		FeeEntryID:      line.FeeEntryID,
		InterestEntryID: line.InterestEntryID,
		CreatedAt:       line.CreatedAt,
		UpdatedAt:       line.UpdatedAt,
	}
}

type BankStatementResponse struct {
	ID           uint64    `json:"id"`
	JournalID    *uint64   `json:"journal_id"`
	Name         *string   `json:"name"`
	Date         *string   `json:"date"`
	BalanceStart float64   `json:"balance_start"`
	BalanceEnd   float64   `json:"balance_end"`
	CurrencyCode *string   `json:"currency_code"`
	State        string    `json:"state"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func newBankStatementResponse(statement *accounting.BankStatement) BankStatementResponse {
	return BankStatementResponse{
		ID:           statement.ID,
		JournalID:    statement.JournalID,
		Name:         statement.Name,
		Date:         helper.FormatDatePtr(statement.Date),
		BalanceStart: statement.BalanceStart.Float64(),
		BalanceEnd:   statement.BalanceEnd.Float64(),
		CurrencyCode: statement.CurrencyCode,
		State:        statement.State,
		CreatedAt:    statement.CreatedAt,
		UpdatedAt:    statement.UpdatedAt,
	}
}

func writeBankStatementError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrStatementNotFound):
		return httpx.CreateNotFoundResponse(c, "Bank statement not found.")
	case errors.Is(err, accounting.ErrStatementNoLines), errors.Is(err, accounting.ErrStatementCancelled):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Bank statement must have at least one line and must not be cancelled.", nil)
	default:
		httpx.RequestLog(c).Error("bank statement write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save bank statement.", err)
	}
}
