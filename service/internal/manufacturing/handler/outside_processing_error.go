package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

func writeSubcontractError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, manufacturing.ErrOutsideProcessingNotFound):
		return httpx.CreateNotFoundResponse(c, "Outside processing order not found.")
	case errors.Is(err, manufacturing.ErrOutsideProcessingDuplicate):
		return httpx.CreateConflictResponse(c, "Outside processing order already exists for this production order.", err)
	case errors.Is(err, manufacturing.ErrOutsideProcessingState), errors.Is(err, manufacturing.ErrOutsideProcessingOrderState):
		return httpx.CreateConflictResponse(c, "State does not allow this operation.", err)
	case errors.Is(err, manufacturing.ErrProductionOrderNotFound):
		return httpx.CreateNotFoundResponse(c, "Production order not found.")
	case errors.Is(err, manufacturing.ErrProductionOrderRecipe):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order recipe is missing.", nil)
	case errors.Is(err, manufacturing.ErrOutsideProcessingNotSubcontracted):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order recipe is not a subcontract recipe.", nil)
	case errors.Is(err, manufacturing.ErrOutsideProcessingComponents):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Outside processing order has no components to send.", nil)
	case errors.Is(err, manufacturing.ErrOutsideProcessingSupplierLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Supplier location does not exist.", nil)
	case errors.Is(err, manufacturing.ErrOutsideProcessingOperation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subcontract operation amount could not be determined.", nil)
	case errors.Is(err, manufacturing.ErrOutsideProcessingPurchaseOrder):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Subcontract purchase order is missing.", nil)
	case errors.Is(err, manufacturing.ErrOutsideProcessingSupplier):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Outside processing order requires an active supplier supplier.", nil)
	case errors.Is(err, manufacturing.ErrWIPAccount):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Work in progress account is required.", nil)
	default:
		httpx.RequestLog(c).Error("outside processing order write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process outside processing order.", err)
	}
}
