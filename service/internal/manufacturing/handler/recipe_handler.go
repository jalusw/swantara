package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type RecipeHandler struct {
	svc manufacturing.RecipeService
}

func NewRecipeHandler(svc manufacturing.RecipeService) RecipeHandler {
	return RecipeHandler{svc: svc}
}

type RecipeResponse struct {
	ID             uint64    `json:"id"`
	OrganizationID *uint64   `json:"organization_id"`
	ItemID         uint64    `json:"item_id"`
	Code           *string   `json:"code"`
	Qty            float64   `json:"qty"`
	UnitID         *uint64   `json:"unit_id"`
	Type           string    `json:"type"`
	Version        int       `json:"version"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func newRecipeResponse(recipe *manufacturing.Recipe) RecipeResponse {
	return RecipeResponse{
		ID:             recipe.ID,
		OrganizationID: recipe.OrganizationID,
		ItemID:         recipe.ItemID,
		Code:           recipe.Code,
		Qty:            recipe.Qty,
		UnitID:         recipe.UnitID,
		Type:           recipe.Type,
		Version:        recipe.Version,
		Active:         recipe.Active,
		CreatedAt:      recipe.CreatedAt,
		UpdatedAt:      recipe.UpdatedAt,
	}
}

type RecipeLineResponse struct {
	ID               uint64    `json:"id"`
	RecipeID         uint64    `json:"recipe_id"`
	ComponentID      uint64    `json:"component_id"`
	Qty              float64   `json:"qty"`
	UnitID           *uint64   `json:"unit_id"`
	ScrapPct         float64   `json:"scrap_pct"`
	ProductionStepID *uint64   `json:"production_step_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func newRecipeLineResponse(line *manufacturing.RecipeLine) RecipeLineResponse {
	return RecipeLineResponse{
		ID:               line.ID,
		RecipeID:         line.RecipeID,
		ComponentID:      line.ComponentID,
		Qty:              line.Qty,
		UnitID:           line.UnitID,
		ScrapPct:         line.ScrapPct,
		ProductionStepID: line.ProductionStepID,
		CreatedAt:        line.CreatedAt,
		UpdatedAt:        line.UpdatedAt,
	}
}

var bomQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"item_id":         {},
	"code":            {},
	"type":            {},
	"version":         {},
	"active":          {},
	"created_at":      {},
	"updated_at":      {},
}

func writeRecipeError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, manufacturing.ErrRecipeNotFound):
		return httpx.CreateNotFoundResponse(c, "recipe not found.")
	case errors.Is(err, manufacturing.ErrRecipeItem):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "recipe item variant does not exist.", nil)
	case errors.Is(err, manufacturing.ErrInvalidRecipeType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid recipe type.", nil)
	case errors.Is(err, manufacturing.ErrRecipeComponent):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "recipe line component variant does not exist.", nil)
	case errors.Is(err, manufacturing.ErrRecipeLineQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "recipe line quantity must be greater than zero.", nil)
	case errors.Is(err, manufacturing.ErrRecipeLineScrap):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "recipe line scrap percentage cannot be negative.", nil)
	case errors.Is(err, manufacturing.ErrRecipeQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "recipe quantity cannot be negative.", nil)
	case errors.Is(err, manufacturing.ErrRecipeOutputQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "recipe output quantity must be greater than zero.", nil)
	case errors.Is(err, manufacturing.ErrRecipeNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "recipe has no lines.", nil)
	default:
		httpx.RequestLog(c).Error("recipe write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to save recipe.", err)
	}
}
