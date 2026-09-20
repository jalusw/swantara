package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type ReorderHandler struct {
	svc inventory.ReorderService
}

func NewReorderHandler(svc inventory.ReorderService) ReorderHandler {
	return ReorderHandler{svc: svc}
}

func writeReorderError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrInvalidRuleQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Minimum, maximum and multiple quantities are invalid.", nil)
	case errors.Is(err, inventory.ErrLocationRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A stock location is required.", nil)
	default:
		httpx.RequestLog(c).Error("reorder rule write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save reorder rule.", err)
	}
}
