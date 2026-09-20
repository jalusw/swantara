package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type JournalHandler struct {
	svc reference.JournalService
}

func NewJournalHandler(svc reference.JournalService) JournalHandler {
	return JournalHandler{svc: svc}
}

type JournalResponse struct {
	ID               uint64    `json:"id"`
	OrganizationID   uint64    `json:"organization_id"`
	Code             *string   `json:"code"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	DefaultAccountID *uint64   `json:"default_account_id"`
	BankAccountID    *uint64   `json:"bank_account_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func newJournalResponse(journal *reference.Journal) JournalResponse {
	return JournalResponse{
		ID:               journal.ID,
		OrganizationID:   journal.OrganizationID,
		Code:             journal.Code,
		Name:             journal.Name,
		Type:             journal.Type,
		DefaultAccountID: journal.DefaultAccountID,
		BankAccountID:    journal.BankAccountID,
		CreatedAt:        journal.CreatedAt,
		UpdatedAt:        journal.UpdatedAt,
	}
}

var journalQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"code":            {},
	"name":            {},
	"type":            {},
	"created_at":      {},
	"updated_at":      {},
}

func writeJournalError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reference.ErrInvalidJournalType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Journal type is not valid.", nil)
	case errors.Is(err, reference.ErrJournalDefaultAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Default account does not exist for the organization.", nil)
	default:
		httpx.RequestLog(c).Error("journal write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save journal.", err)
	}
}
