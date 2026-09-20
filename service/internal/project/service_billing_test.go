package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func openTimeMaterialProject() *Project {
	return &Project{
		Base:           model.Base{ID: 1},
		OrganizationID: 1,
		ContactID:      7,
		BillingType:    BillingTypeTimeMaterial,
		BillableRate:   250000,
		DimensionID:    helper.Ptr(uint64(5)),
		State:          ProjectStateOpen,
	}
}

func billTimeMaterialRequest() BillTimeMaterialRequest {
	return BillTimeMaterialRequest{
		OrganizationID: 1,
		ProjectID:      1,
		JournalID:      1,
		Date:           time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestProjectService_BillTimeMaterial_PropagatesFindError(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return nil, errors.New("boom")
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectService_BillTimeMaterial_ReturnsNotFound(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return nil, nil
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err != ErrProjectNotFound {
		t.Fatalf("error = %v, want %v", err, ErrProjectNotFound)
	}
}

func TestProjectService_BillTimeMaterial_RejectsNoBillableRate(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		project := openTimeMaterialProject()
		project.BillableRate = 0
		return project, nil
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err != ErrProjectNoBillableRate {
		t.Fatalf("error = %v, want %v", err, ErrProjectNoBillableRate)
	}
}

func TestProjectService_BillTimeMaterial_PropagatesTimesheetError(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return nil, errors.New("boom")
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectService_BillTimeMaterial_PropagatesInvoiceLineError(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return nil, errors.New("boom")
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectService_BillTimeMaterial_PropagatesRevenueAccountError(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return []*ProjectInvoiceLine{}, nil
	}
	ctx.accounts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
		return nil, errors.New("boom")
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectService_BillTimeMaterial_RejectsNoRevenueAccount(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return []*ProjectInvoiceLine{}, nil
	}
	ctx.accounts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
		return &query.Page[reference.Account]{Items: []*reference.Account{
			{Base: model.Base{ID: 4200}, OrganizationID: 99, Active: true},
		}}, nil
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err != ErrProjectNoRevenueAccount {
		t.Fatalf("error = %v, want %v", err, ErrProjectNoRevenueAccount)
	}
}

func TestProjectService_BillTimeMaterial_PropagatesInvoiceCreateError(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return []*ProjectInvoiceLine{}, nil
	}
	ctx.invoices.CreateFunc = func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
		return nil, errors.New("boom")
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectService_BillTimeMaterial_PropagatesInvoiceLineLookupError(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return []*ProjectInvoiceLine{}, nil
	}
	ctx.invoiceLookup.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
		return nil, errors.New("boom")
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectService_BillTimeMaterial_PropagatesJoinLineCreateError(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return []*ProjectInvoiceLine{}, nil
	}
	ctx.invoiceLines.CreateFunc = func(_ context.Context, _ *ProjectInvoiceLine) (*ProjectInvoiceLine, error) {
		return nil, errors.New("boom")
	}

	_, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProjectService_BillTimeMaterial_SkipsAlreadyBilledTimesheets(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{
			{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20},
			{Base: model.Base{ID: 2}, EmployeeID: 8, Hours: 10},
		}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return []*ProjectInvoiceLine{{Base: model.Base{ID: 1}, ProjectID: 1, TimesheetID: 1, Amount: 5000000}}, nil
	}
	var createdLines []*ProjectInvoiceLine
	ctx.invoiceLines.CreateFunc = func(_ context.Context, line *ProjectInvoiceLine) (*ProjectInvoiceLine, error) {
		createdLines = append(createdLines, line)
		return line, nil
	}

	invoice, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invoice.ID != 1 {
		t.Fatalf("invoice id = %d, want 1", invoice.ID)
	}
	if len(createdLines) != 1 {
		t.Fatalf("join lines = %d, want 1", len(createdLines))
	}
	if createdLines[0].TimesheetID != 2 {
		t.Fatalf("join timesheet id = %d, want 2", createdLines[0].TimesheetID)
	}
}

func TestProjectService_BillTimeMaterial_LeavesInvoiceLineIDNilWhenShort(t *testing.T) {
	ctx := newProjectTestContext()
	ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
		return openTimeMaterialProject(), nil
	}
	ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
		return []*payroll.Timesheet{
			{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20},
			{Base: model.Base{ID: 2}, EmployeeID: 8, Hours: 10},
		}, nil
	}
	ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
		return []*ProjectInvoiceLine{}, nil
	}
	ctx.invoiceLookup.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
		return []*accounting.InvoiceLine{{Base: model.Base{ID: 9}, InvoiceID: 1}}, nil
	}
	var createdLines []*ProjectInvoiceLine
	ctx.invoiceLines.CreateFunc = func(_ context.Context, line *ProjectInvoiceLine) (*ProjectInvoiceLine, error) {
		createdLines = append(createdLines, line)
		return line, nil
	}

	invoice, err := ctx.svc.BillTimeMaterial(context.Background(), billTimeMaterialRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if invoice.ID != 1 {
		t.Fatalf("invoice id = %d, want 1", invoice.ID)
	}
	if len(createdLines) != 2 {
		t.Fatalf("join lines = %d, want 2", len(createdLines))
	}
	if createdLines[0].InvoiceLineID == nil || *createdLines[0].InvoiceLineID != 9 {
		t.Fatalf("first join invoice line id = %v, want 9", createdLines[0].InvoiceLineID)
	}
	if createdLines[1].InvoiceLineID != nil {
		t.Fatalf("second join invoice line id = %v, want nil", createdLines[1].InvoiceLineID)
	}
}

func TestProjectDAOMockDefaults_ReturnEmptyLists(t *testing.T) {
	tasks := ProjectTaskDAOMock{}
	taskItems, err := tasks.ListByProject(context.Background(), 1)
	if err != nil || len(taskItems) != 0 {
		t.Fatalf("tasks = %v, err = %v, want empty", taskItems, err)
	}

	milestones := ProjectMilestoneDAOMock{}
	milestoneItems, err := milestones.ListByProject(context.Background(), 1)
	if err != nil || len(milestoneItems) != 0 {
		t.Fatalf("milestones = %v, err = %v, want empty", milestoneItems, err)
	}

	invoiceLines := ProjectInvoiceLineDAOMock{}
	lineItems, err := invoiceLines.ListByTimesheet(context.Background(), 1)
	if err != nil || len(lineItems) != 0 {
		t.Fatalf("invoice lines = %v, err = %v, want empty", lineItems, err)
	}
}

func TestProjectDAOMock_DelegatesListByProjectAndTimesheet(t *testing.T) {
	milestones := ProjectMilestoneDAOMock{
		ListByProjectFunc: func(_ context.Context, _ uint64) ([]*ProjectMilestone, error) {
			return []*ProjectMilestone{{Base: model.Base{ID: 1}, Name: "Launch"}}, nil
		},
	}
	milestoneItems, err := milestones.ListByProject(context.Background(), 1)
	if err != nil || len(milestoneItems) != 1 {
		t.Fatalf("milestones = %v, err = %v, want one milestone", milestoneItems, err)
	}

	invoiceLines := ProjectInvoiceLineDAOMock{
		ListByTimesheetFunc: func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
			return []*ProjectInvoiceLine{{Base: model.Base{ID: 1}, TimesheetID: 1}}, nil
		},
	}
	lineItems, err := invoiceLines.ListByTimesheet(context.Background(), 1)
	if err != nil || len(lineItems) != 1 {
		t.Fatalf("invoice lines = %v, err = %v, want one line", lineItems, err)
	}
}
