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
)

func TestCommissionPlanDAO_ConstructorAndList(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "commission_plans"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "commission_plans"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Sales Plan"))

	plans := NewCommissionPlanDAO(db)
	page, err := plans.List(ctx, &query.Query{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 {
		t.Errorf("page = %+v, want one plan", page)
	}

	query.AssertDBMockDone(t, mock)
}

func TestCommissionRuleDAO_CreateTx_CreatesRule(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_rules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	rules := NewCommissionRuleDAO(db)
	tx := db.Begin()

	rule := &CommissionRule{PlanID: 1, RatePct: 5}
	created, err := rules.CreateTx(ctx, tx, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("created id = %d, want 1", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestCommissionRuleDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_rules"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	rules := NewCommissionRuleDAO(db)
	tx := db.Begin()

	_, err := rules.CreateTx(ctx, tx, &CommissionRule{})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCommissionRuleDAO_ListByPlan_ReturnsRules(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "commission_rules" WHERE plan_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "commission_rules" WHERE plan_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "plan_id", "rate_pct"}).
			AddRow(1, 1, 5).
			AddRow(2, 1, 10))

	rules := NewCommissionRuleDAO(db)

	items, err := rules.ListByPlan(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].RatePct != 5 || items[1].RatePct != 10 {
		t.Errorf("items = %+v, want two rules", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestCommissionAssignmentDAO_CreateTx_CreatesAssignment(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_assignments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	assigns := NewCommissionAssignmentDAO(db)
	tx := db.Begin()

	assignment := &CommissionAssignment{PlanID: 1, SalespersonID: 5}
	created, err := assigns.CreateTx(ctx, tx, assignment)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("created id = %d, want 1", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestCommissionAssignmentDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_assignments"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	assigns := NewCommissionAssignmentDAO(db)
	tx := db.Begin()

	_, err := assigns.CreateTx(ctx, tx, &CommissionAssignment{})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestCommissionEntryDAO_CreateTx_CreatesEntry(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "commission_entries"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	entries := NewCommissionEntryDAO(db)
	tx := db.Begin()

	entry := &CommissionEntry{PlanID: 1, SalespersonID: 5, State: EntryStateConfirmed}
	created, err := entries.CreateTx(ctx, tx, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("created id = %d, want 1", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestCommissionEntryDAO_UpdateTx_SavesEntry(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "commission_entries" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	entries := NewCommissionEntryDAO(db)
	tx := db.Begin()

	entry := &CommissionEntry{Base: model.Base{ID: 1}, State: EntryStatePaid}
	updated, err := entries.UpdateTx(ctx, tx, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != entry {
		t.Errorf("UpdateTx returned a different entry")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}
