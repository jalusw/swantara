package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type StockHandler struct {
	stock     inventory.StockService
	ledger    inventory.LedgerService
	valuation inventory.ValuationService
}

func NewStockHandler(
	stock inventory.StockService,
	ledger inventory.LedgerService,
	valuation inventory.ValuationService,
) StockHandler {
	return StockHandler{stock: stock, ledger: ledger, valuation: valuation}
}

func writeStockError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrMovementNotFound):
		return httpx.CreateNotFoundResponse(c, "Stock movement not found.")
	case errors.Is(err, inventory.ErrMovementItem):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock movement item variant is required.", nil)
	case errors.Is(err, inventory.ErrMovementQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock movement quantity must be greater than zero.", nil)
	case errors.Is(err, inventory.ErrMovementState):
		return httpx.CreateConflictResponse(c, "Stock movement cannot be processed in its current state.", err)
	case errors.Is(err, inventory.ErrLocationNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock location does not exist.", nil)
	case errors.Is(err, inventory.ErrInsufficientStock):
		return httpx.CreateConflictResponse(c, "Insufficient stock for this movement.", err)
	case errors.Is(err, inventory.ErrNotReceipt):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Movement source is not a supplier location.", nil)
	case errors.Is(err, inventory.ErrNotShipment):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Movement destination is not a customer location.", nil)
	case errors.Is(err, inventory.ErrBatchRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A batch is required for this item.", nil)
	case errors.Is(err, inventory.ErrNegativeCost):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unit cost cannot be negative.", nil)
	case errors.Is(err, inventory.ErrOrganizationMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required for posting.", nil)
	case errors.Is(err, inventory.ErrValuationAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no stock valuation account.", nil)
	case errors.Is(err, inventory.ErrInputAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no stock input account.", nil)
	case errors.Is(err, inventory.ErrCogsAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no COGS account.", nil)
	default:
		httpx.RequestLog(c).Error("stock write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process stock operation.", err)
	}
}
