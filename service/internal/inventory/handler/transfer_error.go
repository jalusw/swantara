package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

func writeTransferError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrWarehouseTransferNotFound):
		return httpx.CreateNotFoundResponse(c, "Warehouse transfer not found.")
	case errors.Is(err, inventory.ErrWarehouseTransferState):
		return httpx.CreateConflictResponse(c, "Warehouse transfer cannot transition in its current state.", err)
	case errors.Is(err, inventory.ErrSameWarehouse):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Source and destination warehouses must differ.", nil)
	case errors.Is(err, inventory.ErrWarehouseNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Warehouse does not exist.", nil)
	case errors.Is(err, inventory.ErrTransitNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "No transit location found for this organization.", nil)
	case errors.Is(err, inventory.ErrTransitAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A goods in transit account is required.", nil)
	case errors.Is(err, inventory.ErrOrganizationMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required for posting.", nil)
	case errors.Is(err, inventory.ErrMovementQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Transfer line quantity must be greater than zero.", nil)
	case errors.Is(err, inventory.ErrInsufficientStock):
		return httpx.CreateConflictResponse(c, "Insufficient stock to send this transfer.", err)
	case errors.Is(err, inventory.ErrLocationRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A stock location is required.", nil)
	case errors.Is(err, inventory.ErrValuationAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no stock valuation account.", nil)
	default:
		httpx.RequestLog(c).Error("warehouse transfer write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process warehouse transfer.", err)
	}
}
