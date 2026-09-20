package expense

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type ExpenseReportDAOMock struct {
	dao.CRUDMock[ExpenseReport]
	SearchFunc   func(ctx context.Context, field string, value any) (*ExpenseReport, error)
	CreateTxFunc func(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error)
	UpdateTxFunc func(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error)
}

func (m ExpenseReportDAOMock) Search(ctx context.Context, field string, value any) (*ExpenseReport, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, field, value)
	}
	return nil, nil
}

func (m ExpenseReportDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, report)
	}
	return report, nil
}

func (m ExpenseReportDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, report *ExpenseReport) (*ExpenseReport, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, report)
	}
	return report, nil
}

type ExpenseLineDAOMock struct {
	dao.CRUDMock[ExpenseLine]
	ListByReportFunc func(ctx context.Context, reportID uint64) ([]*ExpenseLine, error)
	CreateTxFunc     func(ctx context.Context, tx *gorm.DB, line *ExpenseLine) (*ExpenseLine, error)
}

func (m ExpenseLineDAOMock) ListByReport(ctx context.Context, reportID uint64) ([]*ExpenseLine, error) {
	if m.ListByReportFunc != nil {
		return m.ListByReportFunc(ctx, reportID)
	}
	return []*ExpenseLine{}, nil
}

func (m ExpenseLineDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, line *ExpenseLine) (*ExpenseLine, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, line)
	}
	return line, nil
}
