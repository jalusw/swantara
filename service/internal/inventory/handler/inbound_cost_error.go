package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

func writeInboundCostError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, inventory.ErrInboundCostNotFound):
		return httpx.CreateNotFoundResponse(c, "Inbound cost not found.")
	case errors.Is(err, inventory.ErrInboundCostState):
		return httpx.CreateConflictResponse(c, "Inbound cost state is invalid for this operation.", err)
	case errors.Is(err, inventory.ErrInboundCostNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Inbound cost has no lines.", nil)
	case errors.Is(err, inventory.ErrInboundCostNoShipments):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Inbound cost has no target shipments.", nil)
	case errors.Is(err, inventory.ErrInboundCostNoMovements):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A inbound cost line has no matching receipt movements.", nil)
	case errors.Is(err, inventory.ErrInboundCostNoAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A inbound cost line has no clearing account.", nil)
	case errors.Is(err, inventory.ErrInboundCostNoJournal):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Inbound cost journal is not configured for this organization.", nil)
	case errors.Is(err, inventory.ErrInboundCostSplit):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inbound cost split method.", nil)
	case errors.Is(err, inventory.ErrInboundCostValue):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Inbound cost value split requires received movements with a valuation layer.", nil)
	case errors.Is(err, inventory.ErrValuationAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Item category has no stock valuation account.", nil)
	case errors.Is(err, inventory.ErrOrganizationMissing):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required for posting.", nil)
	default:
		httpx.RequestLog(c).Error("inbound cost write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save inbound cost.", err)
	}
}
