package project

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestProjectDAO_Create_InsertsProject(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "projects"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	projects := NewProjectDAO(db)
	created, err := projects.Create(ctx, &Project{OrganizationID: 1, Name: "Website Revamp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectDAO_Create_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "projects"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	projects := NewProjectDAO(db)
	_, err := projects.Create(ctx, &Project{Name: "Website Revamp"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectDAO_Find_ReturnsProject(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "projects" WHERE id = $1`)).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "name", "billing_type", "state"}).
			AddRow(7, 1, "Website Revamp", BillingTypeFixed, ProjectStateDraft))

	projects := NewProjectDAO(db)
	found, err := projects.Find(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.Name != "Website Revamp" {
		t.Fatalf("found = %+v, want Website Revamp", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectDAO_Update_SavesProject(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "projects" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	projects := NewProjectDAO(db)
	updated, err := projects.Update(ctx, &Project{Base: model.Base{ID: 7}, Name: "Website Revamp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated == nil || updated.ID != 7 {
		t.Fatalf("updated = %+v, want id 7", updated)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectTaskDAO_ListByProject_ReturnsTasks(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_tasks" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "project_tasks" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name"}).
			AddRow(1, 7, "Design").
			AddRow(2, 7, "Build"))

	tasks := NewProjectTaskDAO(db)
	items, err := tasks.ListByProject(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Name != "Design" {
		t.Fatalf("items = %+v, want two tasks", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectTaskDAO_ListByProject_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_tasks" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	tasks := NewProjectTaskDAO(db)
	_, err := tasks.ListByProject(ctx, 7)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectMilestoneDAO_ListByProject_ReturnsMilestones(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_milestones" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "project_milestones" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "reached"}).
			AddRow(1, 7, "Launch", true))

	milestones := NewProjectMilestoneDAO(db)
	items, err := milestones.ListByProject(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || !items[0].Reached {
		t.Fatalf("items = %+v, want one reached milestone", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectMilestoneDAO_ListByProject_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_milestones" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	milestones := NewProjectMilestoneDAO(db)
	_, err := milestones.ListByProject(ctx, 7)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectInvoiceLineDAO_ListByProject_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_invoice_lines" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "project_invoice_lines" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "invoice_id", "timesheet_id"}).
			AddRow(1, 7, 9, 3))

	lines := NewProjectInvoiceLineDAO(db)
	items, err := lines.ListByProject(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].TimesheetID != 3 {
		t.Fatalf("items = %+v, want one line for timesheet 3", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectInvoiceLineDAO_ListByProject_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_invoice_lines" WHERE project_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	lines := NewProjectInvoiceLineDAO(db)
	_, err := lines.ListByProject(ctx, 7)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectInvoiceLineDAO_ListByTimesheet_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_invoice_lines" WHERE timesheet_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "project_invoice_lines" WHERE timesheet_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "invoice_id", "timesheet_id"}).
			AddRow(1, 7, 9, 3))

	lines := NewProjectInvoiceLineDAO(db)
	items, err := lines.ListByTimesheet(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].TimesheetID != 3 {
		t.Fatalf("items = %+v, want one line for timesheet 3", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestProjectInvoiceLineDAO_ListByTimesheet_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "project_invoice_lines" WHERE timesheet_id = $1`)).
		WithArgs(3).
		WillReturnError(errors.New("db down"))

	lines := NewProjectInvoiceLineDAO(db)
	_, err := lines.ListByTimesheet(ctx, 3)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	query.AssertDBMockDone(t, mock)
}
