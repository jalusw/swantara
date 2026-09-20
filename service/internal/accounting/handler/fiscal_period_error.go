package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func writeTaxPeriodError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrPeriodNotFound):
		return httpx.CreateNotFoundResponse(c, "Tax period not found.")
	case errors.Is(err, accounting.ErrInvalidPeriod):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax period date range is invalid.", nil)
	case errors.Is(err, accounting.ErrInvalidPeriodState):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax period state is not valid.", nil)
	case errors.Is(err, accounting.ErrPeriodNotClosed):
		return httpx.CreateConflictResponse(c, "Tax period must be closed before it can be locked.", err)
	case errors.Is(err, accounting.ErrPeriodLocked), errors.Is(err, accounting.ErrPeriodNotOpen), errors.Is(err, accounting.ErrPeriodAlreadyClosed):
		return httpx.CreateConflictResponse(c, "Tax period is closed or locked.", err)
	case errors.Is(err, accounting.ErrCloseRequiresEntries):
		return httpx.CreateConflictResponse(c, "Period close requires closing entries.", err)
	case errors.Is(err, accounting.ErrTaxYearNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Tax year does not exist for the organization.", nil)
	default:
		httpx.RequestLog(c).Error("tax period write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save tax period.", err)
	}
}
