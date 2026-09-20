package expense

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ExpenseReportDAO interface {
	dao.CRUD[ExpenseReport]
	CreateTx(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error)
}

type expenseReportDAO struct {
	dao.Base[ExpenseReport]
	db *gorm.DB
}

func NewExpenseReportDAO(db *gorm.DB) ExpenseReportDAO {
	return expenseReportDAO{Base: dao.NewBase[ExpenseReport](db), db: db}
}

func (d expenseReportDAO) CreateTx(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
	if err := tx.WithContext(ctx).Create(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

func (d expenseReportDAO) UpdateTx(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
	if err := tx.WithContext(ctx).Save(report).Error; err != nil {
		return nil, err
	}
	return report, nil
}

type ExpenseLineDAO interface {
	dao.CRUD[ExpenseLine]
	ListByReport(ctx context.Context, reportID uint64) ([]*ExpenseLine, error)
	CreateTx(ctx context.Context, tx *gorm.DB, line *ExpenseLine) (*ExpenseLine, error)
}

type expenseLineDAO struct {
	dao.Base[ExpenseLine]
	db *gorm.DB
}

func NewExpenseLineDAO(db *gorm.DB) ExpenseLineDAO {
	return expenseLineDAO{Base: dao.NewBase[ExpenseLine](db), db: db}
}

func (d expenseLineDAO) ListByReport(ctx context.Context, reportID uint64) ([]*ExpenseLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "report_id", Operator: query.Equal, Value: reportID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d expenseLineDAO) CreateTx(ctx context.Context, tx *gorm.DB, line *ExpenseLine) (*ExpenseLine, error) {
	if err := tx.WithContext(ctx).Create(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}
