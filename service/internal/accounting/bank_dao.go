package accounting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type BankStatementDAO interface {
	dao.CRUD[BankStatement]
	CreateWithLinesTx(ctx context.Context, tx *gorm.DB, statement *BankStatement, lines []*BankStatementLine) (*BankStatement, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, statement *BankStatement) (*BankStatement, error)
}

type bankStatementDAO struct {
	dao.Base[BankStatement]
	db *gorm.DB
}

func NewBankStatementDAO(db *gorm.DB) BankStatementDAO {
	return bankStatementDAO{Base: dao.NewBase[BankStatement](db), db: db}
}

func (d bankStatementDAO) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, statement *BankStatement, lines []*BankStatementLine) (*BankStatement, error) {
	if err := tx.WithContext(ctx).Create(statement).Error; err != nil {
		return nil, err
	}
	for _, line := range lines {
		line.StatementID = statement.ID
		if err := tx.WithContext(ctx).Create(line).Error; err != nil {
			return nil, err
		}
	}
	return statement, nil
}

func (d bankStatementDAO) UpdateTx(ctx context.Context, tx *gorm.DB, statement *BankStatement) (*BankStatement, error) {
	if err := tx.WithContext(ctx).Save(statement).Error; err != nil {
		return nil, err
	}
	return statement, nil
}

type BankStatementLineDAO interface {
	dao.CRUD[BankStatementLine]
	ListByStatement(ctx context.Context, statementID uint64) ([]*BankStatementLine, error)
	ListUnreconciledByStatement(ctx context.Context, statementID uint64) ([]*BankStatementLine, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, line *BankStatementLine) (*BankStatementLine, error)
}

type bankStatementLineDAO struct {
	dao.Base[BankStatementLine]
	db *gorm.DB
}

func NewBankStatementLineDAO(db *gorm.DB) BankStatementLineDAO {
	return bankStatementLineDAO{Base: dao.NewBase[BankStatementLine](db), db: db}
}

func (d bankStatementLineDAO) ListByStatement(ctx context.Context, statementID uint64) ([]*BankStatementLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "statement_id", Operator: query.Equal, Value: statementID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d bankStatementLineDAO) ListUnreconciledByStatement(ctx context.Context, statementID uint64) ([]*BankStatementLine, error) {
	var entities []BankStatementLine
	if err := d.db.WithContext(ctx).Where("statement_id = ? AND reconciled = ?", statementID, false).Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*BankStatementLine, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

func (d bankStatementLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, line *BankStatementLine) (*BankStatementLine, error) {
	if err := tx.WithContext(ctx).Save(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}
