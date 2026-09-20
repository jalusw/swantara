package expense

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type ExpenseCategoryService struct {
	categories dao.CRUD[reference.ExpenseCategory]
}

func NewExpenseCategoryService(categories dao.CRUD[reference.ExpenseCategory]) ExpenseCategoryService {
	return ExpenseCategoryService{categories: categories}
}

func (s ExpenseCategoryService) List(ctx context.Context, q *query.Query) (*query.Page[reference.ExpenseCategory], error) {
	return s.categories.List(ctx, q)
}

func (s ExpenseCategoryService) Find(ctx context.Context, id uint64) (*reference.ExpenseCategory, error) {
	return s.categories.Find(ctx, id)
}

func (s ExpenseCategoryService) Create(ctx context.Context, category *reference.ExpenseCategory) (*reference.ExpenseCategory, error) {
	return s.categories.Create(ctx, category)
}

func (s ExpenseCategoryService) Update(ctx context.Context, category *reference.ExpenseCategory) (*reference.ExpenseCategory, error) {
	return s.categories.Update(ctx, category)
}

func (s ExpenseCategoryService) Delete(ctx context.Context, id uint64) error {
	return s.categories.Delete(ctx, id)
}
