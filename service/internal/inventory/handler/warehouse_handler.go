package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type WarehouseHandler struct {
	svc inventory.WarehouseService
}

func NewWarehouseHandler(
	svc inventory.WarehouseService,
) WarehouseHandler {
	return WarehouseHandler{svc: svc}
}

func writeWarehouseError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrWarehouseNameRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Warehouse name is required.", nil)
	case errors.Is(err, inventory.ErrWarehouseCodeTaken):
		return httpx.CreateConflictResponse(c, "Warehouse code already exists.", err)
	case errors.Is(err, inventory.ErrWarehouseNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Warehouse does not exist.", nil)
	case errors.Is(err, inventory.ErrLocationNameRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock location name is required.", nil)
	case errors.Is(err, inventory.ErrInvalidLocationType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock location usage is invalid.", nil)
	case errors.Is(err, inventory.ErrLocationParent):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock location parent does not exist.", nil)
	case errors.Is(err, inventory.ErrWarehouseMismatch):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Stock location organization does not match its warehouse.", nil)
	default:
		httpx.RequestLog(c).Error("warehouse write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save warehouse.", err)
	}
}
