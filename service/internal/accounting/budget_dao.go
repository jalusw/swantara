package accounting

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type BudgetDAO interface {
	dao.CRUD[Budget]
	CreateWithLinesTx(ctx context.Context, tx *gorm.DB, budget *Budget, lines []*BudgetLine) (*Budget, error)
}

type budgetDAO struct {
	dao.Base[Budget]
	db *gorm.DB
}

func NewBudgetDAO(db *gorm.DB) BudgetDAO {
	return budgetDAO{Base: dao.NewBase[Budget](db), db: db}
}

func (d budgetDAO) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, budget *Budget, lines []*BudgetLine) (*Budget, error) {
	if err := tx.WithContext(ctx).Create(budget).Error; err != nil {
		return nil, err
	}
	for _, line := range lines {
		line.BudgetID = budget.ID
		if err := tx.WithContext(ctx).Create(line).Error; err != nil {
			return nil, err
		}
	}
	return budget, nil
}

type BudgetLineDAO interface {
	dao.CRUD[BudgetLine]
	ListByBudget(ctx context.Context, budgetID uint64) ([]*BudgetLine, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, line *BudgetLine) (*BudgetLine, error)
	SumPracticalByBudget(ctx context.Context, budgetID uint64) (float64, error)
}

type budgetLineDAO struct {
	dao.Base[BudgetLine]
	db *gorm.DB
}

func NewBudgetLineDAO(db *gorm.DB) BudgetLineDAO {
	return budgetLineDAO{Base: dao.NewBase[BudgetLine](db), db: db}
}

func (d budgetLineDAO) ListByBudget(ctx context.Context, budgetID uint64) ([]*BudgetLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "budget_id", Operator: query.Equal, Value: budgetID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d budgetLineDAO) UpdateTx(ctx context.Context, tx *gorm.DB, line *BudgetLine) (*BudgetLine, error) {
	if err := tx.WithContext(ctx).Save(line).Error; err != nil {
		return nil, err
	}
	return line, nil
}

func (d budgetLineDAO) SumPracticalByBudget(ctx context.Context, budgetID uint64) (float64, error) {
	var total float64
	if err := d.db.WithContext(ctx).
		Model(&BudgetLine{}).
		Where("budget_id = ?", budgetID).
		Select("COALESCE(SUM(practical_amount), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

type TaxRuleDAO interface {
	dao.CRUD[TaxRule]
	CreateWithMapsTx(ctx context.Context, tx *gorm.DB, position *TaxRule, taxMaps []*TaxRuleTaxMap, accountMaps []*TaxRuleAccountMap) (*TaxRule, error)
}

type taxRuleDAO struct {
	dao.Base[TaxRule]
	db *gorm.DB
}

func NewTaxRuleDAO(db *gorm.DB) TaxRuleDAO {
	return taxRuleDAO{Base: dao.NewBase[TaxRule](db), db: db}
}

func (d taxRuleDAO) CreateWithMapsTx(ctx context.Context, tx *gorm.DB, position *TaxRule, taxMaps []*TaxRuleTaxMap, accountMaps []*TaxRuleAccountMap) (*TaxRule, error) {
	if err := tx.WithContext(ctx).Create(position).Error; err != nil {
		return nil, err
	}
	for _, taxMap := range taxMaps {
		taxMap.TaxRuleID = position.ID
		if err := tx.WithContext(ctx).Create(taxMap).Error; err != nil {
			return nil, err
		}
	}
	for _, accountMap := range accountMaps {
		accountMap.TaxRuleID = position.ID
		if err := tx.WithContext(ctx).Create(accountMap).Error; err != nil {
			return nil, err
		}
	}
	return position, nil
}

type TaxRuleTaxMapDAO interface {
	dao.CRUD[TaxRuleTaxMap]
	ListByPosition(ctx context.Context, positionID uint64) ([]*TaxRuleTaxMap, error)
}

type taxRuleTaxMapDAO struct {
	dao.Base[TaxRuleTaxMap]
	db *gorm.DB
}

func NewTaxRuleTaxMapDAO(db *gorm.DB) TaxRuleTaxMapDAO {
	return taxRuleTaxMapDAO{Base: dao.NewBase[TaxRuleTaxMap](db), db: db}
}

func (d taxRuleTaxMapDAO) ListByPosition(ctx context.Context, positionID uint64) ([]*TaxRuleTaxMap, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "tax_rule_id", Operator: query.Equal, Value: positionID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type TaxRuleAccountMapDAO interface {
	dao.CRUD[TaxRuleAccountMap]
	ListByPosition(ctx context.Context, positionID uint64) ([]*TaxRuleAccountMap, error)
}

type taxRuleAccountMapDAO struct {
	dao.Base[TaxRuleAccountMap]
	db *gorm.DB
}

func NewTaxRuleAccountMapDAO(db *gorm.DB) TaxRuleAccountMapDAO {
	return taxRuleAccountMapDAO{Base: dao.NewBase[TaxRuleAccountMap](db), db: db}
}

func (d taxRuleAccountMapDAO) ListByPosition(ctx context.Context, positionID uint64) ([]*TaxRuleAccountMap, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "tax_rule_id", Operator: query.Equal, Value: positionID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type WithholdingTaxDAO interface {
	dao.CRUD[WithholdingTax]
	ListActiveByScope(ctx context.Context, organizationID uint64, scope string) ([]*WithholdingTax, error)
}

type withholdingTaxDAO struct {
	dao.Base[WithholdingTax]
	db *gorm.DB
}

func NewWithholdingTaxDAO(db *gorm.DB) WithholdingTaxDAO {
	return withholdingTaxDAO{Base: dao.NewBase[WithholdingTax](db), db: db}
}

func (d withholdingTaxDAO) ListActiveByScope(ctx context.Context, organizationID uint64, scope string) ([]*WithholdingTax, error) {
	var entities []WithholdingTax
	if err := d.db.WithContext(ctx).
		Where("organization_id = ? AND scope = ? AND active = ?", organizationID, scope, true).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*WithholdingTax, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

type TaxReturnDAO interface {
	dao.CRUD[TaxReturn]
	FindByPeriod(ctx context.Context, periodID uint64) (*TaxReturn, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, taxReturn *TaxReturn) (*TaxReturn, error)
}

type taxReturnDAO struct {
	dao.Base[TaxReturn]
	db *gorm.DB
}

func NewTaxReturnDAO(db *gorm.DB) TaxReturnDAO {
	return taxReturnDAO{Base: dao.NewBase[TaxReturn](db), db: db}
}

func (d taxReturnDAO) FindByPeriod(ctx context.Context, periodID uint64) (*TaxReturn, error) {
	var entity TaxReturn
	if err := d.db.WithContext(ctx).Where("period_id = ?", periodID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

func (d taxReturnDAO) UpdateTx(ctx context.Context, tx *gorm.DB, taxReturn *TaxReturn) (*TaxReturn, error) {
	if err := tx.WithContext(ctx).Save(taxReturn).Error; err != nil {
		return nil, err
	}
	return taxReturn, nil
}

type PracticalAmounts struct {
	ByAccount map[uint64]float64
	Total     float64
}

type BudgetQueryDAO interface {
	SumPracticalByPeriod(ctx context.Context, organizationID uint64, accountID uint64, start, end time.Time) (float64, error)
}

type budgetQueryDAO struct {
	db *gorm.DB
}

func NewBudgetQueryDAO(db *gorm.DB) BudgetQueryDAO {
	return budgetQueryDAO{db: db}
}

func (d budgetQueryDAO) SumPracticalByPeriod(ctx context.Context, organizationID uint64, accountID uint64, start, end time.Time) (float64, error) {
	var total float64
	if err := d.db.WithContext(ctx).
		Model(&JournalLine{}).
		Joins("JOIN journal_entrys ON journal_entrys.id = journal_lines.entry_id").
		Where("journal_entrys.organization_id = ? AND journal_entrys.state = ?", organizationID, EntryStatePosted).
		Where("journal_lines.account_id = ?", accountID).
		Where("journal_entrys.date >= ? AND journal_entrys.date <= ?", start, end).
		Select("COALESCE(SUM(journal_lines.debit - journal_lines.credit), 0)").
		Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
