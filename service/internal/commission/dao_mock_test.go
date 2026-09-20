package commission

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestCommissionPlanDAOMock_CRUD(t *testing.T) {
	ctx := context.Background()
	found := &CommissionPlan{Base: baseID(1)}
	mock := CommissionPlanDAOMock{
		ListFunc: func(context.Context, *query.Query) (*query.Page[CommissionPlan], error) {
			return &query.Page[CommissionPlan]{Items: []*CommissionPlan{found}}, nil
		},
		FindFunc:       func(context.Context, uint64) (*CommissionPlan, error) { return found, nil },
		CreateFunc:     func(context.Context, *CommissionPlan) (*CommissionPlan, error) { return found, nil },
		UpdateFunc:     func(context.Context, *CommissionPlan) (*CommissionPlan, error) { return found, nil },
		DeleteFunc:     func(context.Context, uint64) error { return nil },
		HardDeleteFunc: func(context.Context, uint64) error { return nil },
	}
	if _, err := mock.List(ctx, &query.Query{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Find(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Create(ctx, found); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Update(ctx, found); err != nil {
		t.Fatal(err)
	}
	if err := mock.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := mock.HardDelete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Search(ctx, "name", "x"); err != nil {
		t.Fatal(err)
	}
}

func TestCommissionRuleDAOMock_CRUD(t *testing.T) {
	ctx := context.Background()
	found := &CommissionRule{Base: baseID(1)}
	mock := CommissionRuleDAOMock{
		ListFunc: func(context.Context, *query.Query) (*query.Page[CommissionRule], error) {
			return &query.Page[CommissionRule]{Items: []*CommissionRule{found}}, nil
		},
		FindFunc:       func(context.Context, uint64) (*CommissionRule, error) { return found, nil },
		CreateFunc:     func(context.Context, *CommissionRule) (*CommissionRule, error) { return found, nil },
		UpdateFunc:     func(context.Context, *CommissionRule) (*CommissionRule, error) { return found, nil },
		DeleteFunc:     func(context.Context, uint64) error { return nil },
		HardDeleteFunc: func(context.Context, uint64) error { return nil },
		CreateTxFunc:   func(context.Context, *gorm.DB, *CommissionRule) (*CommissionRule, error) { return found, nil },
	}
	if _, err := mock.List(ctx, &query.Query{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Find(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Create(ctx, found); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Update(ctx, found); err != nil {
		t.Fatal(err)
	}
	if err := mock.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := mock.HardDelete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Search(ctx, "plan_id", uint64(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.CreateTx(ctx, nil, found); err != nil {
		t.Fatal(err)
	}
}

func TestCommissionAssignmentDAOMock_CRUD(t *testing.T) {
	ctx := context.Background()
	found := &CommissionAssignment{Base: baseID(1)}
	mock := CommissionAssignmentDAOMock{
		ListFunc: func(context.Context, *query.Query) (*query.Page[CommissionAssignment], error) {
			return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{found}}, nil
		},
		FindFunc:       func(context.Context, uint64) (*CommissionAssignment, error) { return found, nil },
		CreateFunc:     func(context.Context, *CommissionAssignment) (*CommissionAssignment, error) { return found, nil },
		UpdateFunc:     func(context.Context, *CommissionAssignment) (*CommissionAssignment, error) { return found, nil },
		DeleteFunc:     func(context.Context, uint64) error { return nil },
		HardDeleteFunc: func(context.Context, uint64) error { return nil },
		CreateTxFunc: func(context.Context, *gorm.DB, *CommissionAssignment) (*CommissionAssignment, error) {
			return found, nil
		},
	}
	if _, err := mock.List(ctx, &query.Query{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Find(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Create(ctx, found); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Update(ctx, found); err != nil {
		t.Fatal(err)
	}
	if err := mock.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := mock.HardDelete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Search(ctx, "salesperson_id", uint64(5)); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.CreateTx(ctx, nil, found); err != nil {
		t.Fatal(err)
	}
}

func TestCommissionEntryDAOMock_CRUD(t *testing.T) {
	ctx := context.Background()
	found := &CommissionEntry{Base: model.Base{ID: 1}}
	mock := CommissionEntryDAOMock{
		ListFunc: func(context.Context, *query.Query) (*query.Page[CommissionEntry], error) {
			return &query.Page[CommissionEntry]{Items: []*CommissionEntry{found}}, nil
		},
		FindFunc:       func(context.Context, uint64) (*CommissionEntry, error) { return found, nil },
		CreateFunc:     func(context.Context, *CommissionEntry) (*CommissionEntry, error) { return found, nil },
		UpdateFunc:     func(context.Context, *CommissionEntry) (*CommissionEntry, error) { return found, nil },
		DeleteFunc:     func(context.Context, uint64) error { return nil },
		HardDeleteFunc: func(context.Context, uint64) error { return nil },
		CreateTxFunc:   func(context.Context, *gorm.DB, *CommissionEntry) (*CommissionEntry, error) { return found, nil },
		UpdateTxFunc:   func(context.Context, *gorm.DB, *CommissionEntry) (*CommissionEntry, error) { return found, nil },
	}
	if _, err := mock.List(ctx, &query.Query{}); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Find(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Create(ctx, found); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Update(ctx, found); err != nil {
		t.Fatal(err)
	}
	if err := mock.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := mock.HardDelete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.Search(ctx, "source_id", uint64(7)); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.CreateTx(ctx, nil, found); err != nil {
		t.Fatal(err)
	}
	if _, err := mock.UpdateTx(ctx, nil, found); err != nil {
		t.Fatal(err)
	}
}
