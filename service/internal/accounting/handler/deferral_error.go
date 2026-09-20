package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func writeDeferralError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrScheduleNotFound):
		return httpx.CreateNotFoundResponse(c, "Deferral schedule not found.")
	case errors.Is(err, accounting.ErrScheduleType),
		errors.Is(err, accounting.ErrScheduleMethod),
		errors.Is(err, accounting.ErrScheduleAmount),
		errors.Is(err, accounting.ErrScheduleNoAccount),
		errors.Is(err, accounting.ErrScheduleInvalidDate),
		errors.Is(err, accounting.ErrScheduleNoLines),
		errors.Is(err, accounting.ErrScheduleLineAmount),
		errors.Is(err, accounting.ErrScheduleNoJournal):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid deferral schedule.", nil)
	default:
		httpx.RequestLog(c).Error("deferral write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process deferral.", err)
	}
}
