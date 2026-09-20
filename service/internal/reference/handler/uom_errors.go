package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func writeUnitError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, reference.ErrUnitGroupNotFound):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "UoM category does not exist.", nil)
	case errors.Is(err, reference.ErrUnitNameTaken):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "A UoM with this name already exists in the category.", nil)
	case errors.Is(err, reference.ErrInvalidFactor):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "UoM factor must be greater than zero.", nil)
	case errors.Is(err, reference.ErrUnitNotFound):
		return httpx.CreateNotFoundResponse(c, "UoM not found.")
	case errors.Is(err, reference.ErrUnitGroupMismatch):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "UoMs belong to different categories.", nil)
	default:
		httpx.RequestLog(c).Error("unit write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save UoM.", err)
	}
}
