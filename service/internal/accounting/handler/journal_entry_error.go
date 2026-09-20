package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func writeJournalEntryError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrEntryNotFound):
		return httpx.CreateNotFoundResponse(c, "Account entry not found.")
	case errors.Is(err, accounting.ErrEntryNotDraft):
		return httpx.CreateConflictResponse(c, "Only draft journal entries can be confirmed.", err)
	case errors.Is(err, accounting.ErrEntryNotPosted):
		return httpx.CreateConflictResponse(c, "Only posted account movements can be reversed.", err)
	case errors.Is(err, accounting.ErrEntryReversed), errors.Is(err, accounting.ErrEntryImmutable):
		return httpx.CreateConflictResponse(c, "Account entry is already reversed.", err)
	case errors.Is(err, accounting.ErrPeriodLocked):
		return httpx.CreateConflictResponse(c, "The tax period for this date is closed or locked.", err)
	case errors.Is(err, accounting.ErrOriginIncomplete):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Origin entry is invalid: origin_id requires an origin_type.", nil)
	case errors.Is(err, accounting.ErrReverserMissing):
		httpx.RequestLog(c).Error("account entry reversal engine missing", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Reversal engine is not configured.", err)
	case errors.Is(err, accounting.ErrNoLines), errors.Is(err, accounting.ErrInvalidLine), errors.Is(err, accounting.ErrUnbalanced):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Journal entry is invalid: it must be balanced with valid debit and credit lines.", nil)
	default:
		httpx.RequestLog(c).Error("account entry write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to post journal entry.", err)
	}
}
