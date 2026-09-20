package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

// @Summary List expense categories
// @Description Lists expense categories with pagination, sorting, and filtering.
// @Tags Expense
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. name:like:Travel)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListExpenseCategoriesResponseEnvelope "Expense categories retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/expense-categories [get]
func (h ExpenseCategoryHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, expenseCategoryQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.categories.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("expense category list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve expense categories.", err)
	}

	items := make([]ExpenseCategoryResponse, len(page.Items))
	for i, category := range page.Items {
		items[i] = newExpenseCategoryResponse(category)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "expense-categories.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Expense categories retrieved successfully.", ListExpenseCategoriesResponse{
		Categories: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get expense category
// @Description Returns a single expense category by id; a 404 is returned if it does not exist or belongs to another organization.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense category ID"
// @Success 200 {object} GetExpenseCategoryResponseEnvelope "Expense category retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense category not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/expense-categories/{id} [get]
func (h ExpenseCategoryHandler) Get(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expense category id provided.", nil)
	}

	category, err := h.categories.Find(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("expense category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get expense category.", err)
	}
	if category == nil || !httpx.OwnsTenant(c, category.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Expense category not found.")
	}

	return httpx.CreateSuccessResponse(c, "Expense category retrieved successfully.", GetExpenseCategoryResponse{
		Category: newExpenseCategoryResponse(category),
	})
}

// @Summary Create expense category
// @Description Creates an expense category with a name and an expense account used for posting.
// @Tags Expense
// @Accept json
// @Produce json
// @Param body body CreateExpenseCategoryRequest true "Expense category details"
// @Success 201 {object} CreateExpenseCategoryResponseEnvelope "Expense category created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/expense-categories [post]
func (h ExpenseCategoryHandler) Create(c fiber.Ctx) error {
	var request CreateExpenseCategoryRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	category, err := h.categories.Create(c, &reference.ExpenseCategory{
		OrganizationID:   organizationID,
		Name:             request.Name,
		ExpenseAccountID: request.ExpenseAccountID,
		DefaultTaxIDs:    request.DefaultTaxIDs,
	})
	if err != nil {
		httpx.RequestLog(c).Error("expense category creation failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create expense category.", err)
	}

	return httpx.CreateCreatedResponse(c, "Expense category created successfully.", CreateExpenseCategoryResponseEnvelope{
		Data: GetExpenseCategoryResponse{Category: newExpenseCategoryResponse(category)},
	})
}

// @Summary Update expense category
// @Description Updates an expense category; returns 404 if the category does not exist or belongs to another organization.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense category ID"
// @Param body body UpdateExpenseCategoryRequest true "Expense category details"
// @Success 200 {object} UpdateExpenseCategoryResponseEnvelope "Expense category updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense category not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/expense-categories/{id} [put]
func (h ExpenseCategoryHandler) Update(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expense category id provided.", nil)
	}

	category, err := h.categories.Find(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("expense category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update expense category.", err)
	}
	if category == nil || !httpx.OwnsTenant(c, category.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Expense category not found.")
	}

	var request UpdateExpenseCategoryRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	category.Name = request.Name
	category.ExpenseAccountID = request.ExpenseAccountID
	category.DefaultTaxIDs = request.DefaultTaxIDs

	updated, err := h.categories.Update(c, category)
	if err != nil {
		httpx.RequestLog(c).Error("expense category update failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update expense category.", err)
	}

	return httpx.CreateSuccessResponse(c, "Expense category updated successfully.", UpdateExpenseCategoryResponseEnvelope{
		Data: GetExpenseCategoryResponse{Category: newExpenseCategoryResponse(updated)},
	})
}

// @Summary Delete expense category
// @Description Deletes an expense category by id; returns 404 if it does not exist and 204 No Content on success.
// @Tags Expense
// @Accept json
// @Produce json
// @Param id path integer true "Expense category ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Expense category not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/expense-categories/{id} [delete]
func (h ExpenseCategoryHandler) Delete(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid expense category id provided.", nil)
	}

	category, err := h.categories.Find(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("expense category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete expense category.", err)
	}
	if category == nil || !httpx.OwnsTenant(c, category.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Expense category not found.")
	}

	if err := h.categories.Delete(c, categoryID); err != nil {
		httpx.RequestLog(c).Error("expense category deletion failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete expense category.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
