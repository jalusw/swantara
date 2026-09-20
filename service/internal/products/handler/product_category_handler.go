package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ItemCategoryHandler struct {
	categories products.ItemCategoryService
}

func NewItemCategoryHandler(categories products.ItemCategoryService) ItemCategoryHandler {
	return ItemCategoryHandler{categories: categories}
}

type ItemCategoryResponse struct {
	ID                      uint64    `json:"id"`
	OrganizationID          *uint64   `json:"organization_id"`
	Name                    string    `json:"name"`
	ParentID                *uint64   `json:"parent_id"`
	IncomeAccountID         *uint64   `json:"income_account_id"`
	ExpenseAccountID        *uint64   `json:"expense_account_id"`
	StockValuationAccountID *uint64   `json:"stock_cost_account_id"`
	StockInputAccountID     *uint64   `json:"stock_input_account_id"`
	StockOutputAccountID    *uint64   `json:"stock_output_account_id"`
	CogsAccountID           *uint64   `json:"cogs_account_id"`
	CostMethod              *string   `json:"cost_method"`
	Valuation               *string   `json:"valuation"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

func newItemCategoryResponse(category *reference.ItemCategory) ItemCategoryResponse {
	return ItemCategoryResponse{
		ID:                      category.ID,
		OrganizationID:          category.OrganizationID,
		Name:                    category.Name,
		ParentID:                category.ParentID,
		IncomeAccountID:         category.IncomeAccountID,
		ExpenseAccountID:        category.ExpenseAccountID,
		StockValuationAccountID: category.StockValuationAccountID,
		StockInputAccountID:     category.StockInputAccountID,
		StockOutputAccountID:    category.StockOutputAccountID,
		CogsAccountID:           category.CogsAccountID,
		CostMethod:              category.CostMethod,
		Valuation:               category.Valuation,
		CreatedAt:               category.CreatedAt,
		UpdatedAt:               category.UpdatedAt,
	}
}

var itemCategoryQueryAllowlist = map[string]struct{}{
	"name":            {},
	"parent_id":       {},
	"cost_method":     {},
	"organization_id": {},
	"created_at":      {},
	"updated_at":      {},
}
