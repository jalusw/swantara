package crm

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestNewProspectDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewProspectDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewProspectActivityDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewProspectActivityDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestProspectDAO_ListOpen_ForOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	orgID := helper.Ptr(uint64(10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "prospects" WHERE type = $1 AND closed_at IS NULL AND organization_id = $2`)).
		WithArgs(ProspectKindOpportunity, uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "type"}).
			AddRow(1, "Acme", ProspectKindOpportunity).
			AddRow(2, "Globex", ProspectKindOpportunity))

	dao := NewProspectDAO(db)

	leads, err := dao.ListOpen(ctx, orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leads) != 2 || leads[0].Name != "Acme" || leads[1].Name != "Globex" {
		t.Errorf("leads = %+v, want two open leads", leads)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectDAO_ListOpen_ForAllOrganizations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "prospects" WHERE type = $1 AND closed_at IS NULL`)).
		WithArgs(ProspectKindOpportunity).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	dao := NewProspectDAO(db)

	leads, err := dao.ListOpen(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(leads) != 0 {
		t.Errorf("leads = %+v, want none", leads)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectDAO_ListOpen_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "prospects"`)).
		WillReturnError(errors.New("db down"))

	dao := NewProspectDAO(db)

	_, err := dao.ListOpen(ctx, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectDAO_CountWon_ForOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	orgID := helper.Ptr(uint64(10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "prospects" JOIN pipeline_stages ON pipeline_stages.id = prospects.stage_id WHERE prospects.type = $1 AND pipeline_stages.is_won = true AND prospects.organization_id = $2`)).
		WithArgs(ProspectKindOpportunity, uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	dao := NewProspectDAO(db)

	count, err := dao.CountWon(ctx, orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectDAO_CountWon_ForAllOrganizations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "prospects" JOIN pipeline_stages ON pipeline_stages.id = prospects.stage_id WHERE prospects.type = $1 AND pipeline_stages.is_won = true`)).
		WithArgs(ProspectKindOpportunity).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	dao := NewProspectDAO(db)

	count, err := dao.CountWon(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectDAO_CountLost_ForOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	orgID := helper.Ptr(uint64(10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "prospects" WHERE type = $1 AND lost_reason IS NOT NULL AND organization_id = $2`)).
		WithArgs(ProspectKindOpportunity, uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	dao := NewProspectDAO(db)

	count, err := dao.CountLost(ctx, orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectDAO_CountLost_ForAllOrganizations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "prospects" WHERE type = $1 AND lost_reason IS NOT NULL`)).
		WithArgs(ProspectKindOpportunity).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	dao := NewProspectDAO(db)

	count, err := dao.CountLost(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectActivityDAO_ListInOrg_ReturnsPage(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	orgID := uint64(10)

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities" LEFT JOIN prospects ON prospects.id = prospect_activities.lead_id LEFT JOIN contacts ON contacts.id = prospect_activities.contact_id WHERE (prospect_activities.lead_id IS NOT NULL AND prospects.organization_id = $1) OR (prospect_activities.lead_id IS NULL AND prospect_activities.contact_id IS NOT NULL AND contacts.organization_id = $2)`)).
		WithArgs(orgID, orgID).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities" LEFT JOIN prospects ON prospects.id = prospect_activities.lead_id LEFT JOIN contacts ON contacts.id = prospect_activities.contact_id WHERE (prospect_activities.lead_id IS NOT NULL AND prospects.organization_id = $1) OR (prospect_activities.lead_id IS NULL AND prospect_activities.contact_id IS NOT NULL AND contacts.organization_id = $2)`)).
		WithArgs(orgID, orgID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "summary"}).
			AddRow(1, "Follow up").
			AddRow(2, "Send proposal"))

	dao := NewProspectActivityDAO(db)

	page, err := dao.ListInOrg(ctx, nil, orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 2 || len(page.Items) != 2 || page.Items[0].Summary != "Follow up" {
		t.Errorf("page = %+v, want two activities", page)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectActivityDAO_ListInOrg_AppliesQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	orgID := uint64(10)
	q := &query.Query{Filters: []query.Filter{{Field: "type", Operator: query.Equal, Value: "call"}}}

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities" LEFT JOIN prospects ON prospects.id = prospect_activities.lead_id LEFT JOIN contacts ON contacts.id = prospect_activities.contact_id WHERE (prospect_activities.lead_id IS NOT NULL AND prospects.organization_id = $1) OR (prospect_activities.lead_id IS NULL AND prospect_activities.contact_id IS NOT NULL AND contacts.organization_id = $2)`)).
		WithArgs(orgID, orgID, "call").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities" LEFT JOIN prospects ON prospects.id = prospect_activities.lead_id LEFT JOIN contacts ON contacts.id = prospect_activities.contact_id WHERE (prospect_activities.lead_id IS NOT NULL AND prospects.organization_id = $1) OR (prospect_activities.lead_id IS NULL AND prospect_activities.contact_id IS NOT NULL AND contacts.organization_id = $2)`)).
		WithArgs(orgID, orgID, "call").
		WillReturnRows(sqlmock.NewRows([]string{"id", "type"}).AddRow(1, "call"))

	dao := NewProspectActivityDAO(db)

	page, err := dao.ListInOrg(ctx, q, orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 {
		t.Errorf("page = %+v, want one activity", page)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectActivityDAO_ListInOrg_PropagatesCountError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "prospect_activities"`)).
		WillReturnError(errors.New("db down"))

	dao := NewProspectActivityDAO(db)

	_, err := dao.ListInOrg(ctx, nil, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectActivityDAO_ListInOrg_PropagatesFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "prospect_activities"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities"`)).
		WillReturnError(errors.New("db down"))

	dao := NewProspectActivityDAO(db)

	_, err := dao.ListInOrg(ctx, nil, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectActivityDAO_FindInOrg_FindsActivity(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities" LEFT JOIN prospects ON prospects.id = prospect_activities.lead_id LEFT JOIN contacts ON contacts.id = prospect_activities.contact_id WHERE prospect_activities.id = $1 AND ((prospect_activities.lead_id IS NOT NULL AND prospects.organization_id = $2) OR (prospect_activities.lead_id IS NULL AND prospect_activities.contact_id IS NOT NULL AND contacts.organization_id = $3))`)).
		WithArgs(uint64(7), uint64(10), uint64(10), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "summary"}).AddRow(7, "Follow up"))

	dao := NewProspectActivityDAO(db)

	activity, err := dao.FindInOrg(ctx, 7, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity == nil || activity.ID != 7 {
		t.Errorf("activity = %+v, want activity 7", activity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectActivityDAO_FindInOrg_ReturnsNilWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities"`)).
		WithArgs(uint64(99), uint64(10), uint64(10), 1).
		WillReturnError(gorm.ErrRecordNotFound)

	dao := NewProspectActivityDAO(db)

	activity, err := dao.FindInOrg(ctx, 99, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity != nil {
		t.Errorf("activity = %+v, want nil", activity)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProspectActivityDAO_FindInOrg_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`FROM "prospect_activities"`)).
		WithArgs(uint64(7), uint64(10), uint64(10), 1).
		WillReturnError(errors.New("db down"))

	dao := NewProspectActivityDAO(db)

	_, err := dao.FindInOrg(ctx, 7, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
