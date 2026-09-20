package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete item category
// @Description Deletes a item category by id; returns 404 if the category does not exist or belongs to another organization and 204 No Content on success.
// @Tags Item Categories
// @Accept json
// @Produce json
// @Param id path integer true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Category not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/item-categories/{id} [delete]
func (h ItemCategoryHandler) Delete(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item category id provided.", nil)
	}

	category, err := h.categories.Find(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("item category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete item category.", err)
	}
	if category == nil || !httpx.OwnsTenant(c, category.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item category not found.")
	}

	if err := h.categories.Delete(c, categoryID); err != nil {
		httpx.RequestLog(c).Error("item category deletion failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete item category.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
