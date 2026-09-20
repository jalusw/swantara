package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type ListRecipesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListRecipesResponse `json:"data"`
}
type ListRecipesResponse struct {
	Recipes []RecipeResponse `json:"recipes"`
}

// @Summary List recipes
// @Description Lists tenant-scoped bills of materials with pagination, sorting, and filtering on attributes such as item, code, type, and version. Results can be exported in CSV format when requested. Returns the matching recipes together with pagination metadata.
// @Tags recipes
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. code:asc)"
// @Param filter query string false "Filters (repeatable, e.g. type:eq:manufacture)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListRecipesResponseEnvelope "recipes retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/recipes [get]
func (h RecipeHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, bomQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("recipe list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve recipes.", err)
	}

	items := make([]RecipeResponse, len(page.Items))
	for i, recipe := range page.Items {
		items[i] = newRecipeResponse(recipe)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "recipes.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "recipes retrieved successfully.", ListRecipesResponse{
		Recipes: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetRecipeResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetRecipeResponse `json:"data"`
}
type GetRecipeResponse struct {
	Recipe RecipeResponse `json:"recipe"`
}

// @Summary Get recipe
// @Description Gets a single tenant-scoped bill of materials by id. A 404 is returned when the recipe does not exist or belongs to another organization.
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path integer true "recipe ID"
// @Success 200 {object} GetRecipeResponseEnvelope "recipe retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "recipe not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/recipes/{id} [get]
func (h RecipeHandler) Get(c fiber.Ctx) error {
	recipeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid recipe id provided.", nil)
	}

	recipe, err := h.svc.Find(c, recipeID)
	if err != nil {
		httpx.RequestLog(c).Error("recipe lookup failed", "recipe_id", recipeID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get recipe.", err)
	}
	if recipe == nil || !httpx.OwnsTenant(c, recipe.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "recipe not found.")
	}

	return httpx.CreateSuccessResponse(c, "recipe retrieved successfully.", GetRecipeResponse{
		Recipe: newRecipeResponse(recipe),
	})
}

type CreateRecipeLineRequest struct {
	ComponentID      uint64  `json:"component_id" validate:"required,gt=0"`
	Qty              float64 `json:"qty" validate:"required,gt=0"`
	UnitID           *uint64 `json:"unit_id"`
	ScrapPct         float64 `json:"scrap_pct"`
	ProductionStepID *uint64 `json:"production_step_id"`
}

type CreateRecipeRequest struct {
	OrganizationID *uint64                   `json:"organization_id"`
	ItemID         uint64                    `json:"item_id" validate:"required,gt=0"`
	Code           *string                   `json:"code"`
	Qty            float64                   `json:"qty"`
	UnitID         *uint64                   `json:"unit_id"`
	Type           string                    `json:"type" validate:"required"`
	Version        int                       `json:"version"`
	Lines          []CreateRecipeLineRequest `json:"lines"`
}

type CreateRecipeResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateRecipeResponse `json:"data"`
}
type CreateRecipeResponse struct {
	Recipe RecipeResponse       `json:"recipe"`
	Lines  []RecipeLineResponse `json:"lines"`
}

// @Summary Create recipe
// @Description Creates a bill of materials with its component lines, validating the recipe type, the output item variant, and each line's component variant, quantity, and scrap percentage. The base quantity and version default to 1 and the recipe is marked active when not specified. The recipe and its component lines are saved together and returned.
// @Tags recipes
// @Accept json
// @Produce json
// @Param body body CreateRecipeRequest true "recipe details"
// @Success 201 {object} CreateRecipeResponseEnvelope "recipe created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error, unknown variant, or invalid recipe"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/recipes [post]
func (h RecipeHandler) Create(c fiber.Ctx) error {
	var request CreateRecipeRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	lines := make([]*manufacturing.RecipeLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = &manufacturing.RecipeLine{
			ComponentID:      line.ComponentID,
			Qty:              line.Qty,
			UnitID:           line.UnitID,
			ScrapPct:         line.ScrapPct,
			ProductionStepID: line.ProductionStepID,
		}
	}

	recipe, err := h.svc.CreateRecipe(c, &manufacturing.Recipe{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		ItemID:         request.ItemID,
		Code:           request.Code,
		Qty:            request.Qty,
		UnitID:         request.UnitID,
		Type:           request.Type,
		Version:        request.Version,
	}, lines)
	if err != nil {
		return writeRecipeError(c, err)
	}

	lineResponses := make([]RecipeLineResponse, len(lines))
	for i, line := range lines {
		lineResponses[i] = newRecipeLineResponse(line)
	}

	return httpx.CreateCreatedResponse(c, "recipe created successfully.", CreateRecipeResponse{
		Recipe: newRecipeResponse(recipe),
		Lines:  lineResponses,
	})
}

type UpdateRecipeRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	ItemID         uint64  `json:"item_id" validate:"required,gt=0"`
	Code           *string `json:"code"`
	Qty            float64 `json:"qty"`
	UnitID         *uint64 `json:"unit_id"`
	Type           string  `json:"type" validate:"required"`
	Version        int     `json:"version"`
	Active         *bool   `json:"active"`
}

type UpdateRecipeResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateRecipeResponse `json:"data"`
}
type UpdateRecipeResponse struct {
	Recipe RecipeResponse `json:"recipe"`
}

// @Summary Update recipe
// @Description Updates a tenant-scoped bill of materials header, including its version and active flag. The recipe type, output quantity, and item variant are validated before the changes are persisted.
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path integer true "recipe ID"
// @Param body body UpdateRecipeRequest true "recipe details"
// @Success 200 {object} UpdateRecipeResponseEnvelope "recipe updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "recipe not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/recipes/{id} [put]
func (h RecipeHandler) Update(c fiber.Ctx) error {
	recipeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid recipe id provided.", nil)
	}

	var request UpdateRecipeRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	recipe, err := h.svc.Find(c, recipeID)
	if err != nil {
		httpx.RequestLog(c).Error("recipe lookup failed", "recipe_id", recipeID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update recipe.", err)
	}
	if recipe == nil || !httpx.OwnsTenant(c, recipe.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "recipe not found.")
	}

	recipe.OrganizationID = httpx.TenantOrganizationID(c, recipe.OrganizationID)
	recipe.ItemID = request.ItemID
	recipe.Code = request.Code
	recipe.Qty = request.Qty
	recipe.UnitID = request.UnitID
	recipe.Type = request.Type
	recipe.Version = request.Version
	if request.Active != nil {
		recipe.Active = *request.Active
	}

	updated, err := h.svc.UpdateRecipe(c, recipe)
	if err != nil {
		return writeRecipeError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "recipe updated successfully.", UpdateRecipeResponse{
		Recipe: newRecipeResponse(updated),
	})
}

// @Summary Delete recipe
// @Description Deletes a tenant-scoped bill of materials by id. A 404 is returned when the recipe does not exist or belongs to another organization, and the operation removes the recipe record entirely.
// @Tags recipes
// @Accept json
// @Produce json
// @Param id path integer true "recipe ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "recipe not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/recipes/{id} [delete]
func (h RecipeHandler) Delete(c fiber.Ctx) error {
	recipeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid recipe id provided.", nil)
	}

	recipe, err := h.svc.Find(c, recipeID)
	if err != nil {
		httpx.RequestLog(c).Error("recipe lookup failed", "recipe_id", recipeID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete recipe.", err)
	}
	if recipe == nil || !httpx.OwnsTenant(c, recipe.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "recipe not found.")
	}

	if err := h.svc.Delete(c, recipeID); err != nil {
		httpx.RequestLog(c).Error("recipe deletion failed", "recipe_id", recipeID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete recipe.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
