package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type UpdateItemCategoryRequest struct {
	Name                    string  `json:"name" validate:"required"`
	ParentID                *uint64 `json:"parent_id"`
	IncomeAccountID         *uint64 `json:"income_account_id"`
	ExpenseAccountID        *uint64 `json:"expense_account_id"`
	StockValuationAccountID *uint64 `json:"stock_cost_account_id"`
	StockInputAccountID     *uint64 `json:"stock_input_account_id"`
	StockOutputAccountID    *uint64 `json:"stock_output_account_id"`
	CogsAccountID           *uint64 `json:"cogs_account_id"`
	CostMethod              *string `json:"cost_method"`
	Valuation               *string `json:"valuation"`
}

type UpdateItemCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateItemCategoryResponse `json:"data"`
}
type UpdateItemCategoryResponse struct {
	Category ItemCategoryResponse `json:"category"`
}

// @Summary Update item category
// @Description Updates a item category's name, parent, GL accounts, cost method, and valuation. The name remains required and a 404 is returned if the category does not exist or belongs to another organization.
// @Tags Item Categories
// @Accept json
// @Produce json
// @Param id path integer true "Category ID"
// @Param body body UpdateItemCategoryRequest true "Category details"
// @Success 200 {object} UpdateItemCategoryResponseEnvelope "Category updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Category not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/item-categories/{id} [put]
func (h ItemCategoryHandler) Update(c fiber.Ctx) error {
	categoryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid item category id provided.", nil)
	}

	category, err := h.categories.Find(c, categoryID)
	if err != nil {
		httpx.RequestLog(c).Error("item category lookup failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update item category.", err)
	}
	if category == nil || !httpx.OwnsTenant(c, category.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Item category not found.")
	}

	var request UpdateItemCategoryRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	category.Name = request.Name
	category.ParentID = request.ParentID
	category.IncomeAccountID = request.IncomeAccountID
	category.ExpenseAccountID = request.ExpenseAccountID
	category.StockValuationAccountID = request.StockValuationAccountID
	category.StockInputAccountID = request.StockInputAccountID
	category.StockOutputAccountID = request.StockOutputAccountID
	category.CogsAccountID = request.CogsAccountID
	category.CostMethod = request.CostMethod
	category.Valuation = request.Valuation

	updated, err := h.categories.Update(c, category)
	if err != nil {
		httpx.RequestLog(c).Error("item category update failed", "category_id", categoryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update item category.", err)
	}

	return httpx.CreateSuccessResponse(c, "Item category updated successfully.", UpdateItemCategoryResponse{
		Category: newItemCategoryResponse(updated),
	})
}
