package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/expense"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ExpenseCategoryHandler struct {
	categories expense.ExpenseCategoryService
}

func NewExpenseCategoryHandler(categories expense.ExpenseCategoryService) ExpenseCategoryHandler {
	return ExpenseCategoryHandler{categories: categories}
}

type ExpenseCategoryResponse struct {
	ID               uint64            `json:"id"`
	OrganizationID   *uint64           `json:"organization_id"`
	Name             string            `json:"name"`
	ExpenseAccountID *uint64           `json:"expense_account_id"`
	DefaultTaxIDs    helper.Int64Array `json:"default_tax_ids"`
}

func newExpenseCategoryResponse(category *reference.ExpenseCategory) ExpenseCategoryResponse {
	return ExpenseCategoryResponse{
		ID:               category.ID,
		OrganizationID:   category.OrganizationID,
		Name:             category.Name,
		ExpenseAccountID: category.ExpenseAccountID,
		DefaultTaxIDs:    category.DefaultTaxIDs,
	}
}

var expenseCategoryQueryAllowlist = map[string]struct{}{
	"name":            {},
	"organization_id": {},
	"created_at":      {},
	"updated_at":      {},
}

type ListExpenseCategoriesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListExpenseCategoriesResponse `json:"data"`
}
type ListExpenseCategoriesResponse struct {
	Categories []ExpenseCategoryResponse `json:"categories"`
}

type GetExpenseCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetExpenseCategoryResponse `json:"data"`
}
type GetExpenseCategoryResponse struct {
	Category ExpenseCategoryResponse `json:"category"`
}

type CreateExpenseCategoryRequest struct {
	OrganizationID   *uint64           `json:"organization_id"`
	Name             string            `json:"name" validate:"required"`
	ExpenseAccountID *uint64           `json:"expense_account_id" validate:"required"`
	DefaultTaxIDs    helper.Int64Array `json:"default_tax_ids"`
}

type CreateExpenseCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetExpenseCategoryResponse `json:"data"`
}

type UpdateExpenseCategoryRequest struct {
	Name             string            `json:"name" validate:"required"`
	ExpenseAccountID *uint64           `json:"expense_account_id" validate:"required"`
	DefaultTaxIDs    helper.Int64Array `json:"default_tax_ids"`
}

type UpdateExpenseCategoryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetExpenseCategoryResponse `json:"data"`
}
