package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type PaymentHandler struct {
	svc accounting.PaymentService
}

func NewPaymentHandler(svc accounting.PaymentService) PaymentHandler {
	return PaymentHandler{svc: svc}
}

func writePaymentError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, accounting.ErrPaymentNotFound):
		return httpx.CreateNotFoundResponse(c, "Payment not found.")
	case errors.Is(err, accounting.ErrInvoiceNotFound):
		return httpx.CreateNotFoundResponse(c, "Invoice not found.")
	case errors.Is(err, accounting.ErrPaymentAmount), errors.Is(err, accounting.ErrPaymentNoInvoices), errors.Is(err, accounting.ErrOverAllocation), errors.Is(err, accounting.ErrInvoiceNotPosted):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment is invalid: it must have a positive amount, at least one open posted invoice, and not exceed the open balance.", nil)
	case errors.Is(err, accounting.ErrNoBankAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No bank or cash account configured for the journal.", nil)
	case errors.Is(err, accounting.ErrNoReceivableAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No receivable account configured for the organization.", nil)
	case errors.Is(err, accounting.ErrNoPayableAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No payable account configured for the organization.", nil)
	default:
		httpx.RequestLog(c).Error("payment write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save payment.", err)
	}
}
