package commission

import (
	"context"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type CommissionPlanDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[CommissionPlan], error)
	FindFunc       func(ctx context.Context, id uint64) (*CommissionPlan, error)
	CreateFunc     func(ctx context.Context, entity *CommissionPlan) (*CommissionPlan, error)
	UpdateFunc     func(ctx context.Context, entity *CommissionPlan) (*CommissionPlan, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
}

func (m CommissionPlanDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[CommissionPlan], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[CommissionPlan]{}, nil
}

func (m CommissionPlanDAOMock) Find(ctx context.Context, id uint64) (*CommissionPlan, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m CommissionPlanDAOMock) Create(ctx context.Context, entity *CommissionPlan) (*CommissionPlan, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionPlanDAOMock) Update(ctx context.Context, entity *CommissionPlan) (*CommissionPlan, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionPlanDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionPlanDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionPlanDAOMock) Search(ctx context.Context, field string, value any) (*CommissionPlan, error) {
	return nil, nil
}

type CommissionRuleDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[CommissionRule], error)
	FindFunc       func(ctx context.Context, id uint64) (*CommissionRule, error)
	CreateFunc     func(ctx context.Context, entity *CommissionRule) (*CommissionRule, error)
	UpdateFunc     func(ctx context.Context, entity *CommissionRule) (*CommissionRule, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	CreateTxFunc   func(ctx context.Context, tx *gorm.DB, entity *CommissionRule) (*CommissionRule, error)
	ListByPlanFunc func(ctx context.Context, planID uint64) ([]*CommissionRule, error)
}

func (m CommissionRuleDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[CommissionRule], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[CommissionRule]{}, nil
}

func (m CommissionRuleDAOMock) Find(ctx context.Context, id uint64) (*CommissionRule, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m CommissionRuleDAOMock) Create(ctx context.Context, entity *CommissionRule) (*CommissionRule, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionRuleDAOMock) Update(ctx context.Context, entity *CommissionRule) (*CommissionRule, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionRuleDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionRuleDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionRuleDAOMock) Search(ctx context.Context, field string, value any) (*CommissionRule, error) {
	return nil, nil
}

func (m CommissionRuleDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionRule) (*CommissionRule, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

func (m CommissionRuleDAOMock) ListByPlan(ctx context.Context, planID uint64) ([]*CommissionRule, error) {
	if m.ListByPlanFunc != nil {
		return m.ListByPlanFunc(ctx, planID)
	}
	return []*CommissionRule{}, nil
}

type CommissionAssignmentDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[CommissionAssignment], error)
	FindFunc       func(ctx context.Context, id uint64) (*CommissionAssignment, error)
	CreateFunc     func(ctx context.Context, entity *CommissionAssignment) (*CommissionAssignment, error)
	UpdateFunc     func(ctx context.Context, entity *CommissionAssignment) (*CommissionAssignment, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	CreateTxFunc   func(ctx context.Context, tx *gorm.DB, entity *CommissionAssignment) (*CommissionAssignment, error)
}

func (m CommissionAssignmentDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[CommissionAssignment], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[CommissionAssignment]{}, nil
}

func (m CommissionAssignmentDAOMock) Find(ctx context.Context, id uint64) (*CommissionAssignment, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m CommissionAssignmentDAOMock) Create(ctx context.Context, entity *CommissionAssignment) (*CommissionAssignment, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionAssignmentDAOMock) Update(ctx context.Context, entity *CommissionAssignment) (*CommissionAssignment, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionAssignmentDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionAssignmentDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionAssignmentDAOMock) Search(ctx context.Context, field string, value any) (*CommissionAssignment, error) {
	return nil, nil
}

func (m CommissionAssignmentDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionAssignment) (*CommissionAssignment, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

type CommissionEntryDAOMock struct {
	ListFunc       func(ctx context.Context, q *query.Query) (*query.Page[CommissionEntry], error)
	FindFunc       func(ctx context.Context, id uint64) (*CommissionEntry, error)
	CreateFunc     func(ctx context.Context, entity *CommissionEntry) (*CommissionEntry, error)
	UpdateFunc     func(ctx context.Context, entity *CommissionEntry) (*CommissionEntry, error)
	DeleteFunc     func(ctx context.Context, id uint64) error
	HardDeleteFunc func(ctx context.Context, id uint64) error
	CreateTxFunc   func(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error)
	UpdateTxFunc   func(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error)
}

func (m CommissionEntryDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[CommissionEntry], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[CommissionEntry]{}, nil
}

func (m CommissionEntryDAOMock) Find(ctx context.Context, id uint64) (*CommissionEntry, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m CommissionEntryDAOMock) Create(ctx context.Context, entity *CommissionEntry) (*CommissionEntry, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionEntryDAOMock) Update(ctx context.Context, entity *CommissionEntry) (*CommissionEntry, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return entity, nil
}

func (m CommissionEntryDAOMock) Delete(ctx context.Context, id uint64) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionEntryDAOMock) HardDelete(ctx context.Context, id uint64) error {
	if m.HardDeleteFunc != nil {
		return m.HardDeleteFunc(ctx, id)
	}
	return nil
}

func (m CommissionEntryDAOMock) Search(ctx context.Context, field string, value any) (*CommissionEntry, error) {
	return nil, nil
}

func (m CommissionEntryDAOMock) CreateTx(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error) {
	if m.CreateTxFunc != nil {
		return m.CreateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}

func (m CommissionEntryDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, entity *CommissionEntry) (*CommissionEntry, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, entity)
	}
	return entity, nil
}
