package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type ExplodeRecipeRequest struct {
	Qty string `json:"qty" validate:"required"`
}

type ExplodeRecipeResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ExplodeRecipeResponse `json:"data"`
}
type ExplodeRecipeResponse struct {
	Components []manufacturing.RecipeComponentRequirement `json:"components"`
}

// @Summary Explode recipe
// @Description Explodes a tenant-scoped bill of materials for a given output quantity and returns the required component quantities, scaling each line's quantity by the output ratio and accounting for its scrap percentage. The output quantity must be a positive decimal value.
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path integer true "recipe ID"
// @Param body body ExplodeRecipeRequest true "Output quantity"
// @Success 200 {object} ExplodeRecipeResponseEnvelope "recipe exploded successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "recipe not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/recipes/{id}/explode [post]
func (h RecipeHandler) Explode(c fiber.Ctx) error {
	recipeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid recipe id provided.", nil)
	}

	var request ExplodeRecipeRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	recipe, err := h.svc.Find(c, recipeID)
	if err != nil {
		httpx.RequestLog(c).Error("recipe lookup failed", "recipe_id", recipeID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to explode recipe.", err)
	}
	if recipe == nil || !httpx.OwnsTenant(c, recipe.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "recipe not found.")
	}

	qty, err := amount.FromString(request.Qty)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Quantity must be a valid decimal.", nil)
	}

	requirements, err := h.svc.Explode(c, recipeID, qty)
	if err != nil {
		return writeRecipeError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "recipe exploded successfully.", ExplodeRecipeResponse{
		Components: requirements,
	})
}
