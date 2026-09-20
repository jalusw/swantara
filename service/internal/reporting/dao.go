package reporting

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type FxRevaluationDAO interface {
	dao.CRUD[FxRevaluation]
	ListByOrganization(ctx context.Context, organizationID uint64) ([]*FxRevaluation, error)
}

type FxRevaluationLineDAO interface {
	dao.CRUD[FxRevaluationLine]
	ListByRevaluation(ctx context.Context, revaluationID uint64) ([]*FxRevaluationLine, error)
	ListOpenByOrg(ctx context.Context, organizationID uint64) ([]*FxRevaluationLine, error)
}

type AccrualDAO interface {
	dao.CRUD[Accrual]
	ListByOrganization(ctx context.Context, organizationID uint64) ([]*Accrual, error)
}

type AccrualLineDAO interface {
	dao.CRUD[AccrualLine]
	ListByAccrual(ctx context.Context, accrualID uint64) ([]*AccrualLine, error)
}

type fxRevaluationDAO struct {
	dao.Base[FxRevaluation]
}

func NewFxRevaluationDAO(db *gorm.DB) FxRevaluationDAO {
	return fxRevaluationDAO{Base: dao.NewBase[FxRevaluation](db)}
}

func (d fxRevaluationDAO) ListByOrganization(ctx context.Context, organizationID uint64) ([]*FxRevaluation, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type fxRevaluationLineDAO struct {
	dao.Base[FxRevaluationLine]
	db *gorm.DB
}

func NewFxRevaluationLineDAO(db *gorm.DB) FxRevaluationLineDAO {
	return fxRevaluationLineDAO{Base: dao.NewBase[FxRevaluationLine](db), db: db}
}

func (d fxRevaluationLineDAO) ListByRevaluation(ctx context.Context, revaluationID uint64) ([]*FxRevaluationLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "revaluation_id", Operator: query.Equal, Value: revaluationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (d fxRevaluationLineDAO) ListOpenByOrg(ctx context.Context, organizationID uint64) ([]*FxRevaluationLine, error) {
	var entities []FxRevaluationLine
	if err := d.db.WithContext(ctx).
		Joins("JOIN fx_revaluations ON fx_revaluations.id = fx_revaluation_lines.revaluation_id").
		Where("fx_revaluations.organization_id = ? AND fx_revaluation_lines.reversed = ? AND fx_revaluations.deleted_at IS NULL", organizationID, false).
		Find(&entities).Error; err != nil {
		return nil, err
	}
	items := make([]*FxRevaluationLine, len(entities))
	for i := range entities {
		items[i] = &entities[i]
	}
	return items, nil
}

type accrualDAO struct {
	dao.Base[Accrual]
}

func NewAccrualDAO(db *gorm.DB) AccrualDAO {
	return accrualDAO{Base: dao.NewBase[Accrual](db)}
}

func (d accrualDAO) ListByOrganization(ctx context.Context, organizationID uint64) ([]*Accrual, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type accrualLineDAO struct {
	dao.Base[AccrualLine]
}

func NewAccrualLineDAO(db *gorm.DB) AccrualLineDAO {
	return accrualLineDAO{Base: dao.NewBase[AccrualLine](db)}
}

func (d accrualLineDAO) ListByAccrual(ctx context.Context, accrualID uint64) ([]*AccrualLine, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "accrual_id", Operator: query.Equal, Value: accrualID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}
