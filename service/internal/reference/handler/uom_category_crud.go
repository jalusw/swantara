package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

// @Summary List UoM categories
// @Description Lists unit of measure categories with pagination, sorting, and filtering, returning results as JSON, XML, or CSV.
// @Tags UoM
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. name:like:Weight)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListUnitCategoriesResponseEnvelope "UoM categories retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /unit-categories [get]
func (h UnitGroupHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, unitGroupQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}

	page, err := h.svc.ListCategories(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("unit category list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve UoM categories.", err)
	}

	items := make([]UnitGroupResponse, len(page.Items))
	for i, category := range page.Items {
		items[i] = newUnitGroupResponse(category)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "unit-categories.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "UoM categories retrieved successfully.", ListUnitCategoriesResponse{
		Categories: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get UoM category
// @Description Returns a single unit of measure category by id, including HATEOAS-style links for retrieving, updating, and deleting the category; a 404 is returned if the category does not exist.
// @Tags UoM
// @Accept json
// @Produce json
// @Param id path integer true "UoM category ID"
// @Success 200 {object} GetUnitGroupResponseEnvelope "UoM category retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "UoM category not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /unit-categories/{id} [get]
func (h UnitGroupHandler) Get(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid unit category id provided.", nil)
	}

	category, err := h.svc.FindCategory(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("unit category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get UoM category.", err)
	}
	if category == nil {
		return httpx.CreateNotFoundResponse(c, "UoM category not found.")
	}

	return httpx.CreateSuccessResponseWithLinks(c, "UoM category retrieved successfully.", GetUnitGroupResponse{
		Category: newUnitGroupResponse(category),
	}, buildUnitGroupLinks(c, category.ID))
}

// @Summary Create UoM category
// @Description Creates a unit of measure category with a required name. The name must be provided or validation fails with a 422; on success the new category is returned.
// @Tags UoM
// @Accept json
// @Produce json
// @Param body body CreateUnitGroupRequest true "UoM category details"
// @Success 201 {object} CreateUnitGroupResponseEnvelope "UoM category created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /unit-categories [post]
func (h UnitGroupHandler) Create(c fiber.Ctx) error {
	var request CreateUnitGroupRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	category, err := h.svc.CreateCategory(c, &reference.UnitGroup{Name: request.Name})
	if err != nil {
		httpx.RequestLog(c).Error("unit category creation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create UoM category.", err)
	}

	return httpx.CreateCreatedResponse(c, "UoM category created successfully.", CreateUnitGroupResponse{
		Category: newUnitGroupResponse(category),
	})
}

// @Summary Update UoM category
// @Description Updates a unit of measure category's name, which remains required. Returns 404 if the category does not exist and 422 when the name is missing; the updated category is returned with its links.
// @Tags UoM
// @Accept json
// @Produce json
// @Param id path integer true "UoM category ID"
// @Param body body UpdateUnitGroupRequest true "UoM category details"
// @Success 200 {object} UpdateUnitGroupResponseEnvelope "UoM category updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "UoM category not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /unit-categories/{id} [put]
func (h UnitGroupHandler) Update(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid unit category id provided.", nil)
	}

	var request UpdateUnitGroupRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	category, err := h.svc.FindCategory(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("unit category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update UoM category.", err)
	}
	if category == nil {
		return httpx.CreateNotFoundResponse(c, "UoM category not found.")
	}

	category.Name = request.Name

	updated, err := h.svc.UpdateCategory(c, category)
	if err != nil {
		httpx.RequestLog(c).Error("unit category update failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update UoM category.", err)
	}

	return httpx.CreateSuccessResponseWithLinks(c, "UoM category updated successfully.", UpdateUnitGroupResponse{
		Category: newUnitGroupResponse(updated),
	}, buildUnitGroupLinks(c, updated.ID))
}

// @Summary Delete UoM category
// @Description Deletes a unit of measure category by id; returns 404 if the category does not exist and 204 No Content on success.
// @Tags UoM
// @Accept json
// @Produce json
// @Param id path integer true "UoM category ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "UoM category not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /unit-categories/{id} [delete]
func (h UnitGroupHandler) Delete(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid unit category id provided.", nil)
	}

	category, err := h.svc.FindCategory(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("unit category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete UoM category.", err)
	}
	if category == nil {
		return httpx.CreateNotFoundResponse(c, "UoM category not found.")
	}

	if err := h.svc.DeleteCategory(c, categoryID); err != nil {
		httpx.RequestLog(c).Error("unit category deletion failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete UoM category.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
