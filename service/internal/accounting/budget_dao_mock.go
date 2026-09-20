package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"gorm.io/gorm"
)

type BudgetDAOMock struct {
	dao.CRUDMock[Budget]
	CreateWithLinesTxFunc func(ctx context.Context, tx *gorm.DB, budget *Budget, lines []*BudgetLine) (*Budget, error)
}

func (m BudgetDAOMock) CreateWithLinesTx(ctx context.Context, tx *gorm.DB, budget *Budget, lines []*BudgetLine) (*Budget, error) {
	if m.CreateWithLinesTxFunc != nil {
		return m.CreateWithLinesTxFunc(ctx, tx, budget, lines)
	}
	return budget, nil
}

type BudgetLineDAOMock struct {
	dao.CRUDMock[BudgetLine]
	ListByBudgetFunc         func(ctx context.Context, budgetID uint64) ([]*BudgetLine, error)
	UpdateTxFunc             func(ctx context.Context, tx *gorm.DB, line *BudgetLine) (*BudgetLine, error)
	SumPracticalByBudgetFunc func(ctx context.Context, budgetID uint64) (float64, error)
}

func (m BudgetLineDAOMock) ListByBudget(ctx context.Context, budgetID uint64) ([]*BudgetLine, error) {
	if m.ListByBudgetFunc != nil {
		return m.ListByBudgetFunc(ctx, budgetID)
	}
	return nil, nil
}

func (m BudgetLineDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, line *BudgetLine) (*BudgetLine, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, line)
	}
	return line, nil
}

func (m BudgetLineDAOMock) SumPracticalByBudget(ctx context.Context, budgetID uint64) (float64, error) {
	if m.SumPracticalByBudgetFunc != nil {
		return m.SumPracticalByBudgetFunc(ctx, budgetID)
	}
	return 0, nil
}

type BudgetQueryDAOMock struct {
	SumPracticalByPeriodFunc func(ctx context.Context, organizationID uint64, accountID uint64, start, end time.Time) (float64, error)
}

func (m BudgetQueryDAOMock) SumPracticalByPeriod(ctx context.Context, organizationID uint64, accountID uint64, start, end time.Time) (float64, error) {
	if m.SumPracticalByPeriodFunc != nil {
		return m.SumPracticalByPeriodFunc(ctx, organizationID, accountID, start, end)
	}
	return 0, nil
}

type TaxRuleDAOMock struct {
	dao.CRUDMock[TaxRule]
	CreateWithMapsTxFunc func(ctx context.Context, tx *gorm.DB, position *TaxRule, taxMaps []*TaxRuleTaxMap, accountMaps []*TaxRuleAccountMap) (*TaxRule, error)
}

func (m TaxRuleDAOMock) CreateWithMapsTx(ctx context.Context, tx *gorm.DB, position *TaxRule, taxMaps []*TaxRuleTaxMap, accountMaps []*TaxRuleAccountMap) (*TaxRule, error) {
	if m.CreateWithMapsTxFunc != nil {
		return m.CreateWithMapsTxFunc(ctx, tx, position, taxMaps, accountMaps)
	}
	return position, nil
}

type TaxRuleTaxMapDAOMock struct {
	dao.CRUDMock[TaxRuleTaxMap]
	ListByPositionFunc func(ctx context.Context, positionID uint64) ([]*TaxRuleTaxMap, error)
}

func (m TaxRuleTaxMapDAOMock) ListByPosition(ctx context.Context, positionID uint64) ([]*TaxRuleTaxMap, error) {
	if m.ListByPositionFunc != nil {
		return m.ListByPositionFunc(ctx, positionID)
	}
	return nil, nil
}

type TaxRuleAccountMapDAOMock struct {
	dao.CRUDMock[TaxRuleAccountMap]
	ListByPositionFunc func(ctx context.Context, positionID uint64) ([]*TaxRuleAccountMap, error)
}

func (m TaxRuleAccountMapDAOMock) ListByPosition(ctx context.Context, positionID uint64) ([]*TaxRuleAccountMap, error) {
	if m.ListByPositionFunc != nil {
		return m.ListByPositionFunc(ctx, positionID)
	}
	return nil, nil
}

type WithholdingTaxDAOMock struct {
	dao.CRUDMock[WithholdingTax]
	ListActiveByScopeFunc func(ctx context.Context, organizationID uint64, scope string) ([]*WithholdingTax, error)
}

func (m WithholdingTaxDAOMock) ListActiveByScope(ctx context.Context, organizationID uint64, scope string) ([]*WithholdingTax, error) {
	if m.ListActiveByScopeFunc != nil {
		return m.ListActiveByScopeFunc(ctx, organizationID, scope)
	}
	return nil, nil
}

type TaxReturnDAOMock struct {
	dao.CRUDMock[TaxReturn]
	FindByPeriodFunc func(ctx context.Context, periodID uint64) (*TaxReturn, error)
	UpdateTxFunc     func(ctx context.Context, tx *gorm.DB, taxReturn *TaxReturn) (*TaxReturn, error)
}

func (m TaxReturnDAOMock) FindByPeriod(ctx context.Context, periodID uint64) (*TaxReturn, error) {
	if m.FindByPeriodFunc != nil {
		return m.FindByPeriodFunc(ctx, periodID)
	}
	return nil, nil
}

func (m TaxReturnDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, taxReturn *TaxReturn) (*TaxReturn, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, taxReturn)
	}
	return taxReturn, nil
}
