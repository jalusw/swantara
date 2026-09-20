package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type CreateItemCategoryRequest struct {
	OrganizationID          *uint64 `json:"organization_id"`
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

type CreateItemCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateItemCategoryResponse `json:"data"`
}
type CreateItemCategoryResponse struct {
	Category ItemCategoryResponse `json:"category"`
}

// @Summary Create item category
// @Description Creates a item category with a required name and optional parent, default GL accounts, cost method, and valuation. The name must be provided or validation fails with a 422.
// @Tags Item Categories
// @Accept json
// @Produce json
// @Param body body CreateItemCategoryRequest true "Category details"
// @Success 201 {object} CreateItemCategoryResponseEnvelope "Category created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/item-categories [post]
func (h ItemCategoryHandler) Create(c fiber.Ctx) error {
	var request CreateItemCategoryRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	category, err := h.categories.Create(c, &reference.ItemCategory{
		OrganizationID:          organizationID,
		Name:                    request.Name,
		ParentID:                request.ParentID,
		IncomeAccountID:         request.IncomeAccountID,
		ExpenseAccountID:        request.ExpenseAccountID,
		StockValuationAccountID: request.StockValuationAccountID,
		StockInputAccountID:     request.StockInputAccountID,
		StockOutputAccountID:    request.StockOutputAccountID,
		CogsAccountID:           request.CogsAccountID,
		CostMethod:              request.CostMethod,
		Valuation:               request.Valuation,
	})
	if err != nil {
		httpx.RequestLog(c).Error("item category create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create item category.", err)
	}

	return httpx.CreateCreatedResponse(c, "Item category created successfully.", CreateItemCategoryResponse{
		Category: newItemCategoryResponse(category),
	})
}
