package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type BankStatementDAOMock struct {
	dao.CRUDMock[BankStatement]
	CreateWithLinesTxFunc func(ctx context.Context, tx *gorm.DB, statement *BankStatement, lines []*BankStatementLine) (*BankStatement, error)
	UpdateTxFunc          func(ctx context.Context, tx *gorm.DB, statement *BankStatement) (*BankStatement, error)
}

func (m BankStatementDAOMock) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, statement *BankStatement, lines []*BankStatementLine) (*BankStatement, error) {
	if m.CreateWithLinesTxFunc != nil {
		return m.CreateWithLinesTxFunc(ctx, tx, statement, lines)
	}
	return statement, nil
}

func (m BankStatementDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, statement *BankStatement) (*BankStatement, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, statement)
	}
	return statement, nil
}

type BankStatementLineDAOMock struct {
	dao.CRUDMock[BankStatementLine]
	ListByStatementFunc             func(ctx context.Context, statementID uint64) ([]*BankStatementLine, error)
	ListUnreconciledByStatementFunc func(ctx context.Context, statementID uint64) ([]*BankStatementLine, error)
	UpdateTxFunc                    func(ctx context.Context, tx *gorm.DB, line *BankStatementLine) (*BankStatementLine, error)
}

func (m BankStatementLineDAOMock) ListByStatement(ctx context.Context, statementID uint64) ([]*BankStatementLine, error) {
	if m.ListByStatementFunc != nil {
		return m.ListByStatementFunc(ctx, statementID)
	}
	return nil, nil
}

func (m BankStatementLineDAOMock) ListUnreconciledByStatement(ctx context.Context, statementID uint64) ([]*BankStatementLine, error) {
	if m.ListUnreconciledByStatementFunc != nil {
		return m.ListUnreconciledByStatementFunc(ctx, statementID)
	}
	return nil, nil
}

func (m BankStatementLineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, line *BankStatementLine) (*BankStatementLine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, line)
	}
	return line, nil
}

type ReminderLevelDAOMock struct {
	dao.CRUDMock[reference.ReminderLevel]
	ListSortedFunc func(ctx context.Context) ([]*reference.ReminderLevel, error)
}

func (m ReminderLevelDAOMock) ListSorted(ctx context.Context) ([]*reference.ReminderLevel, error) {
	if m.ListSortedFunc != nil {
		return m.ListSortedFunc(ctx)
	}
	return nil, nil
}
