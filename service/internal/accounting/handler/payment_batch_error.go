package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func writePaymentBatchError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrBatchNotFound):
		return httpx.CreateNotFoundResponse(c, "Payment batch not found.")
	case errors.Is(err, accounting.ErrPaymentNotFound):
		return httpx.CreateNotFoundResponse(c, "Payment not found.")
	case errors.Is(err, accounting.ErrBatchNoPayments), errors.Is(err, accounting.ErrPaymentNoInvoices):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment batch must have at least one payment.", nil)
	case errors.Is(err, accounting.ErrBatchNotDraft), errors.Is(err, accounting.ErrEntryNotPosted), errors.Is(err, accounting.ErrInvalidPeriodState):
		return httpx.CreateConflictResponse(c, "Invalid payment batch state.", err)
	default:
		httpx.RequestLog(c).Error("payment batch write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save payment batch.", err)
	}
}
