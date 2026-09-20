package commission

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestCommissionMock_All(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("plan dao", func(t *testing.T) {
		bare := CommissionPlanDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := bare.Create(ctx, &CommissionPlan{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &CommissionPlan{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}

		wired := CommissionPlanDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionPlan], error) {
				return &query.Page[CommissionPlan]{}, nil
			},
			FindFunc:       func(_ context.Context, _ uint64) (*CommissionPlan, error) { return &CommissionPlan{}, nil },
			CreateFunc:     func(_ context.Context, e *CommissionPlan) (*CommissionPlan, error) { return e, nil },
			UpdateFunc:     func(_ context.Context, e *CommissionPlan) (*CommissionPlan, error) { return e, nil },
			DeleteFunc:     func(_ context.Context, _ uint64) error { return nil },
			HardDeleteFunc: func(_ context.Context, _ uint64) error { return nil },
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := wired.Create(ctx, &CommissionPlan{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Update(ctx, &CommissionPlan{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := wired.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := wired.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
	})

	t.Run("rule dao", func(t *testing.T) {
		bare := CommissionRuleDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := bare.Create(ctx, &CommissionRule{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &CommissionRule{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.CreateTx(ctx, nil, &CommissionRule{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := bare.ListByPlan(ctx, 1); err != nil {
			t.Errorf("ListByPlan = %v", err)
		}

		wired := CommissionRuleDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionRule], error) {
				return &query.Page[CommissionRule]{}, nil
			},
			FindFunc:       func(_ context.Context, _ uint64) (*CommissionRule, error) { return &CommissionRule{}, nil },
			CreateFunc:     func(_ context.Context, e *CommissionRule) (*CommissionRule, error) { return e, nil },
			UpdateFunc:     func(_ context.Context, e *CommissionRule) (*CommissionRule, error) { return e, nil },
			DeleteFunc:     func(_ context.Context, _ uint64) error { return nil },
			HardDeleteFunc: func(_ context.Context, _ uint64) error { return nil },
			CreateTxFunc:   func(_ context.Context, _ *gorm.DB, e *CommissionRule) (*CommissionRule, error) { return e, nil },
			ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) { return []*CommissionRule{{}}, nil },
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := wired.Create(ctx, &CommissionRule{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Update(ctx, &CommissionRule{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := wired.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := wired.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := wired.CreateTx(ctx, nil, &CommissionRule{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := wired.ListByPlan(ctx, 1); err != nil {
			t.Errorf("ListByPlan = %v", err)
		}
	})

	t.Run("assignment dao", func(t *testing.T) {
		bare := CommissionAssignmentDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := bare.Create(ctx, &CommissionAssignment{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &CommissionAssignment{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.CreateTx(ctx, nil, &CommissionAssignment{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}

		wired := CommissionAssignmentDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				return &query.Page[CommissionAssignment]{}, nil
			},
			FindFunc:       func(_ context.Context, _ uint64) (*CommissionAssignment, error) { return &CommissionAssignment{}, nil },
			CreateFunc:     func(_ context.Context, e *CommissionAssignment) (*CommissionAssignment, error) { return e, nil },
			UpdateFunc:     func(_ context.Context, e *CommissionAssignment) (*CommissionAssignment, error) { return e, nil },
			DeleteFunc:     func(_ context.Context, _ uint64) error { return nil },
			HardDeleteFunc: func(_ context.Context, _ uint64) error { return nil },
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, e *CommissionAssignment) (*CommissionAssignment, error) {
				return e, nil
			},
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := wired.Create(ctx, &CommissionAssignment{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Update(ctx, &CommissionAssignment{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := wired.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := wired.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := wired.CreateTx(ctx, nil, &CommissionAssignment{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
	})

	t.Run("entry dao", func(t *testing.T) {
		bare := CommissionEntryDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := bare.Create(ctx, &CommissionEntry{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &CommissionEntry{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.CreateTx(ctx, nil, &CommissionEntry{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := bare.UpdateTx(ctx, nil, &CommissionEntry{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}

		wired := CommissionEntryDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionEntry], error) {
				return &query.Page[CommissionEntry]{}, nil
			},
			FindFunc:       func(_ context.Context, _ uint64) (*CommissionEntry, error) { return &CommissionEntry{}, nil },
			CreateFunc:     func(_ context.Context, e *CommissionEntry) (*CommissionEntry, error) { return e, nil },
			UpdateFunc:     func(_ context.Context, e *CommissionEntry) (*CommissionEntry, error) { return e, nil },
			DeleteFunc:     func(_ context.Context, _ uint64) error { return nil },
			HardDeleteFunc: func(_ context.Context, _ uint64) error { return nil },
			CreateTxFunc:   func(_ context.Context, _ *gorm.DB, e *CommissionEntry) (*CommissionEntry, error) { return e, nil },
			UpdateTxFunc:   func(_ context.Context, _ *gorm.DB, e *CommissionEntry) (*CommissionEntry, error) { return e, nil },
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := wired.Create(ctx, &CommissionEntry{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Update(ctx, &CommissionEntry{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := wired.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := wired.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := wired.CreateTx(ctx, nil, &CommissionEntry{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := wired.UpdateTx(ctx, nil, &CommissionEntry{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
	})
}

func TestCommissionDAO_Tx(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("rule create tx", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_rules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		tx := db.Begin()
		got, err := NewCommissionRuleDAO(db).CreateTx(ctx, tx, &CommissionRule{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("rule create tx error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_rules"`)).WillReturnError(dbErr)

		tx := db.Begin()
		_, err := NewCommissionRuleDAO(db).CreateTx(ctx, tx, &CommissionRule{})
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})

	t.Run("assignment create tx", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_assignments"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		tx := db.Begin()
		got, err := NewCommissionAssignmentDAO(db).CreateTx(ctx, tx, &CommissionAssignment{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("entry tx", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_entries"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		tx := db.Begin()
		got, err := NewCommissionEntryDAO(db).CreateTx(ctx, tx, &CommissionEntry{})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 1 {
			t.Errorf("id = %d", got.ID)
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)

		db, mock = query.NewMockDB(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "commission_entries"`)).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		tx = db.Begin()
		updated, err := NewCommissionEntryDAO(db).UpdateTx(ctx, tx, &CommissionEntry{Base: model.Base{ID: 7}})
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if updated == nil {
			t.Error("entry = nil")
		}
		if err := tx.Commit().Error; err != nil {
			t.Fatalf("commit = %v", err)
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("rule lists by plan", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "commission_rules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "commission_rules"`)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		items, err := NewCommissionRuleDAO(db).ListByPlan(ctx, 3)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(items) != 1 {
			t.Errorf("len = %d", len(items))
		}
		query.AssertDBMockDone(t, mock)
	})

	t.Run("rule list error", func(t *testing.T) {
		db, mock := query.NewMockDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "commission_rules"`)).WillReturnError(dbErr)

		_, err := NewCommissionRuleDAO(db).ListByPlan(ctx, 3)
		helper.AssertError(t, err, true, dbErr)
		query.AssertDBMockDone(t, mock)
	})
}

func TestCommissionFixtures(t *testing.T) {
	if CommissionPlanFixture() == nil {
		t.Error("plan = nil")
	}
	if CommissionRuleFixture() == nil {
		t.Error("rule = nil")
	}
	if CommissionAssignmentFixture() == nil {
		t.Error("assignment = nil")
	}
	if CommissionEntryFixture() == nil {
		t.Error("entry = nil")
	}
}
