package commission

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type CommissionPlanDAO interface {
	dao.CRUD[CommissionPlan]
}

type commissionPlanDAO struct {
	dao.Base[CommissionPlan]
}

func NewCommissionPlanDAO(db *gorm.DB) CommissionPlanDAO {
	return commissionPlanDAO{Base: dao.NewBase[CommissionPlan](db)}
}

type CommissionRuleDAO interface {
	dao.CRUD[CommissionRule]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionRule) (*CommissionRule, error)
	ListByPlan(ctx context.Context, planID uint64) ([]*CommissionRule, error)
}

type commissionRuleDAO struct {
	dao.Base[CommissionRule]
	db *gorm.DB
}

func NewCommissionRuleDAO(db *gorm.DB) CommissionRuleDAO {
	return commissionRuleDAO{Base: dao.NewBase[CommissionRule](db), db: db}
}

func (d commissionRuleDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionRule) (*CommissionRule, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d commissionRuleDAO) ListByPlan(ctx context.Context, planID uint64) ([]*CommissionRule, error) {
	page, err := d.List(ctx, &query.Query{Filters: []query.Filter{{Field: "plan_id", Operator: query.Equal, Value: planID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

type CommissionAssignmentDAO interface {
	dao.CRUD[CommissionAssignment]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionAssignment) (*CommissionAssignment, error)
}

type commissionAssignmentDAO struct {
	dao.Base[CommissionAssignment]
	db *gorm.DB
}

func NewCommissionAssignmentDAO(db *gorm.DB) CommissionAssignmentDAO {
	return commissionAssignmentDAO{Base: dao.NewBase[CommissionAssignment](db), db: db}
}

func (d commissionAssignmentDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionAssignment) (*CommissionAssignment, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

type CommissionEntryDAO interface {
	dao.CRUD[CommissionEntry]
	CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error)
	UpdateTx(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error)
}

type commissionEntryDAO struct {
	dao.Base[CommissionEntry]
	db *gorm.DB
}

func NewCommissionEntryDAO(db *gorm.DB) CommissionEntryDAO {
	return commissionEntryDAO{Base: dao.NewBase[CommissionEntry](db), db: db}
}

func (d commissionEntryDAO) CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error) {
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}

func (d commissionEntryDAO) UpdateTx(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error) {
	if err := tx.WithContext(ctx).Save(entity).Error; err != nil {
		return nil, err
	}
	return entity, nil
}
