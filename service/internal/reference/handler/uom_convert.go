package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

// @Summary Convert between UoMs
// @Description Converts a decimal quantity between two units of measure using their conversion factors. Both UoMs must exist and belong to the same category, and the quantity must be a valid decimal; the result is rounded according to the target UoM's rounding.
// @Tags UoM
// @Accept json
// @Produce json
// @Param body body ConvertUnitRequest true "Conversion details"
// @Success 200 {object} ConvertUnitResponseEnvelope "Quantity converted successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, unknown UoM, or category mismatch"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /units/convert [post]
func (h UnitHandler) Convert(c fiber.Ctx) error {
	var request ConvertUnitRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	qty, err := amount.FromString(request.Qty)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Quantity must be a valid decimal.", nil)
	}

	converted, err := h.svc.Convert(c, qty, request.FromID, request.ToID)
	if err != nil {
		return writeUnitError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Quantity converted successfully.", ConvertUnitResponse{
		Value: converted,
	})
}
