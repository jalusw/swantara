package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

func writeProductionError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, manufacturing.ErrProductionOrderNotFound):
		return httpx.CreateNotFoundResponse(c, "Production order not found.")
	case errors.Is(err, manufacturing.ErrProductionOrderState):
		return httpx.CreateConflictResponse(c, "Production order state does not allow this operation.", err)
	case errors.Is(err, manufacturing.ErrProductionOrderRecipe):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order recipe is missing.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order requires an organization.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order location does not exist.", nil)
	case errors.Is(err, manufacturing.ErrShopTaskNotFound):
		return httpx.CreateNotFoundResponse(c, "Shop task not found.")
	case errors.Is(err, manufacturing.ErrShopTaskWorkCenter):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Shop task work center does not exist.", nil)
	case errors.Is(err, manufacturing.ErrShopTaskMismatch):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Shop task does not belong to the production order.", nil)
	case errors.Is(err, manufacturing.ErrShopTaskLabor):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Shop task labor hours must be greater than zero.", nil)
	case errors.Is(err, manufacturing.ErrConsumeComponent):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Component does not belong to the production order.", nil)
	case errors.Is(err, manufacturing.ErrConsumeQuantity):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Consume quantity exceeds the remaining component quantity.", nil)
	case errors.Is(err, manufacturing.ErrProduceQuantity):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Produce quantity exceeds the remaining quantity to produce.", nil)
	case errors.Is(err, manufacturing.ErrProductionLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production location does not exist.", nil)
	case errors.Is(err, manufacturing.ErrWIPAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Work in progress account is required.", nil)
	case errors.Is(err, manufacturing.ErrLaborAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Applied labor account is required.", nil)
	case errors.Is(err, manufacturing.ErrVarianceAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Manufacturing variance account is required.", nil)
	case errors.Is(err, sequence.ErrSequenceNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Shop task document sequence is not configured.", nil)
	case errors.Is(err, manufacturing.ErrMovementNotFound):
		return httpx.CreateNotFoundResponse(c, "Stock movement not found.")
	case errors.Is(err, manufacturing.ErrMovementState):
		return httpx.CreateConflictResponse(c, "Stock movement state does not allow this operation.", err)
	case errors.Is(err, manufacturing.ErrInsufficientStock):
		return httpx.CreateConflictResponse(c, "Insufficient stock to consume raw materials.", err)
	case errors.Is(err, manufacturing.ErrNotConsumption):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Movement is not routed to a production location.", nil)
	case errors.Is(err, manufacturing.ErrNotProduction):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Movement does not originate from a production location.", nil)
	case errors.Is(err, manufacturing.ErrValuationAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no stock valuation account.", nil)
	case errors.Is(err, manufacturing.ErrNegativeCost):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unit cost cannot be negative.", nil)
	case errors.Is(err, manufacturing.ErrBatchRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item requires a lot or serial number.", nil)
	default:
		httpx.RequestLog(c).Error("production write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process production operation.", err)
	}
}
