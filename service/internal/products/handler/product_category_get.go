package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetItemCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetItemCategoryResponse `json:"data"`
}
type GetItemCategoryResponse struct {
	Category ItemCategoryResponse `json:"category"`
}

// @Summary Get item category
// @Description Returns a single item category by id, including its GL account mappings, cost method, and valuation; a 404 is returned if the category does not exist or belongs to another organization.
// @Tags Item Categories
// @Accept json
// @Produce json
// @Param id path integer true "Category ID"
// @Success 200 {object} GetItemCategoryResponseEnvelope "Category retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Category not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/item-categories/{id} [get]
func (h ItemCategoryHandler) Get(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item category id provided.", nil)
	}

	category, err := h.categories.Find(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("item category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get item category.", err)
	}
	if category == nil || !httpx.OwnsTenant(c, category.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item category not found.")
	}

	return httpx.CreateSuccessResponse(c, "Item category retrieved successfully.", GetItemCategoryResponse{
		Category: newItemCategoryResponse(category),
	})
}
