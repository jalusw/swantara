package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

func writeProductionOrderError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, manufacturing.ErrProductionOrderNotFound):
		return httpx.CreateNotFoundResponse(c, "Production order not found.")
	case errors.Is(err, manufacturing.ErrProductionOrderState):
		return httpx.CreateConflictResponse(c, "Production order state does not allow this operation.", err)
	case errors.Is(err, manufacturing.ErrProductionOrderItem):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order item variant does not exist.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderRecipe):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order recipe is missing or does not match the item.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order quantity must be greater than zero.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order requires an organization.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order location does not exist.", nil)
	case errors.Is(err, manufacturing.ErrConsumedMaterial):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order has no components.", nil)
	case errors.Is(err, sequence.ErrSequenceNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order document sequence is not configured.", nil)
	case errors.Is(err, manufacturing.ErrHoldOverflow):
		return httpx.CreateConflictResponse(c, "Insufficient stock to reserve production order components.", err)
	case errors.Is(err, manufacturing.ErrBalanceNotFound):
		return httpx.CreateConflictResponse(c, "No stock quant found to reserve production order components.", err)
	default:
		httpx.RequestLog(c).Error("production order write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process production order.", err)
	}
}
