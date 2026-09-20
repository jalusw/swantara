package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

func writePaymentBatchError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, procurement.ErrPaymentBatchNotFound):
		return httpx.CreateNotFoundResponse(c, "Payment batch not found.")
	case errors.Is(err, procurement.ErrPaymentBatchNoOrders):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment batch must have at least one order.", nil)
	case errors.Is(err, procurement.ErrPaymentBatchMixedVendors):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "All orders must belong to the same supplier.", nil)
	case errors.Is(err, procurement.ErrPaymentBatchInvalidOrderState):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "One or more orders are not in a valid state.", nil)
	case errors.Is(err, procurement.ErrPaymentBatchInvalidAmount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Amount must be positive.", nil)
	case errors.Is(err, procurement.ErrPaymentBatchInvalidState):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Payment batch is not in a valid state.", nil)
	default:
		httpx.RequestLog(c).Error("payment batch write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process payment batch.", err)
	}
}
