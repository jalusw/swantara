package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/pos"
)

func writePOSError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, pos.ErrConfigNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS config does not exist.", nil)
	case errors.Is(err, pos.ErrSessionNotFound):
		return httpx.CreateNotFoundResponse(c, "POS session not found.")
	case errors.Is(err, pos.ErrSessionState):
		return httpx.CreateConflictResponse(c, "POS session cannot be processed in its current state.", err)
	case errors.Is(err, pos.ErrSessionOpen):
		return httpx.CreateConflictResponse(c, "A POS session for this config is already open.", err)
	case errors.Is(err, pos.ErrSessionReconciliation):
		return httpx.CreateConflictResponse(c, "POS session closing balance does not reconcile with recorded payments.", err)
	case errors.Is(err, pos.ErrOrderNotFound):
		return httpx.CreateNotFoundResponse(c, "POS order not found.")
	case errors.Is(err, pos.ErrOrderNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order must have at least one line.", nil)
	case errors.Is(err, pos.ErrOrderProduct):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order line item is required.", nil)
	case errors.Is(err, pos.ErrOrderQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order line quantity must be positive.", nil)
	case errors.Is(err, pos.ErrOrderDiscount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order line discount must be between 0 and 100.", nil)
	case errors.Is(err, pos.ErrOrderTax):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order line tax is invalid.", nil)
	case errors.Is(err, pos.ErrOrderStockUnavailable):
		return httpx.CreateConflictResponse(c, "POS order quantity is not available in stock.", err)
	case errors.Is(err, pos.ErrOrderNoStock):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order cannot create stock movements.", nil)
	case errors.Is(err, pos.ErrPaymentMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order requires at least one payment.", nil)
	case errors.Is(err, pos.ErrPaymentTotal):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS payments total must match the order total.", nil)
	case errors.Is(err, pos.ErrOrderInvoiced):
		return httpx.CreateConflictResponse(c, "POS order is already invoiced.", err)
	case errors.Is(err, pos.ErrConfigWarehouse):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS config must define a warehouse.", nil)
	case errors.Is(err, pos.ErrConfigJournal):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS config must define a journal with a default account.", nil)
	case errors.Is(err, pos.ErrConfigPriceBook):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS config must define a price_book.", nil)
	case errors.Is(err, pos.ErrNoCashier):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "The cashier is not a member of the POS config organization.", nil)
	case errors.Is(err, pos.ErrOrderState):
		return httpx.CreateConflictResponse(c, "POS order cannot be processed in its current state.", err)
	case errors.Is(err, pos.ErrOrderCost):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "POS order original stock cost cannot be resolved.", nil)
	default:
		httpx.RequestLog(c).Error("pos write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process POS request.", err)
	}
}
