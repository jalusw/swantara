package project

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProjectService_CreateProject(t *testing.T) {
	tests := []struct {
		name       string
		input      *Project
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, p *Project)
	}{
		{
			name: "validates and creates draft",
			input: &Project{
				OrganizationID: 1,
				Name:           "Website Revamp",
				ContactID:      7,
				BillingType:    BillingTypeTimeMaterial,
				BillableRate:   250000,
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.CreateFunc = func(_ context.Context, p *Project) (*Project, error) {
					return p, nil
				}
			},
			check: func(t *testing.T, p *Project) {
				if p.State != ProjectStateDraft {
					t.Fatalf("state = %s, want %s", p.State, ProjectStateDraft)
				}
			},
		},
		{
			name: "rejects unknown billing type",
			input: &Project{
				OrganizationID: 1,
				Name:           "Website Revamp",
				ContactID:      7,
				BillingType:    "nonsense",
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidBillingType,
		},
		{
			name: "rejects missing contact",
			input: &Project{
				OrganizationID: 1,
				Name:           "Website Revamp",
				ContactID:      7,
				BillingType:    BillingTypeFixed,
			},
			setup: func(ctx *projectTestContext) {
				ctx.contacts.FindFunc = func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectContactNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			created, err := ctx.svc.CreateProject(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, created)
			}
		})
	}
}

func TestProjectService_SetProjectState(t *testing.T) {
	tests := []struct {
		name       string
		projectID  uint64
		target     string
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, p *Project)
	}{
		{
			name:      "transitions from draft to open",
			projectID: 1,
			target:    ProjectStateOpen,
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateDraft}, nil
				}
				ctx.projects.UpdateFunc = func(_ context.Context, p *Project) (*Project, error) {
					return p, nil
				}
			},
			check: func(t *testing.T, p *Project) {
				if p.State != ProjectStateOpen {
					t.Fatalf("state = %s, want %s", p.State, ProjectStateOpen)
				}
			},
		},
		{
			name:      "rejects transition from closed to open",
			projectID: 1,
			target:    ProjectStateOpen,
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateClosed}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			updated, err := ctx.svc.SetProjectState(context.Background(), tt.projectID, tt.target)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, updated)
			}
		})
	}
}

func TestProjectService_CreateTask(t *testing.T) {
	tests := []struct {
		name       string
		input      *ProjectTask
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
	}{
		{
			name:  "requires open project",
			input: &ProjectTask{ProjectID: 1, Name: "Design", PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateDraft}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidState,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := ctx.svc.CreateTask(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestProjectService_Summary(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, s ProjectSummary)
	}{
		{
			name: "shows unbilled margin and utilization",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{
						Base:           model.Base{ID: 1},
						OrganizationID: 1,
						BillingType:    BillingTypeTimeMaterial,
						BillableRate:   250000,
					}, nil
				}
				ctx.tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectTask, error) {
					return []*ProjectTask{
						{Base: model.Base{ID: 1}, ProjectID: 1, PlannedHours: 40},
						{Base: model.Base{ID: 2}, ProjectID: 1, PlannedHours: 20},
					}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{
						{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20},
						{Base: model.Base{ID: 2}, EmployeeID: 8, Hours: 10},
					}, nil
				}
				ctx.contracts.FindActiveByEmployeeFunc = func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
					return &payroll.EmploymentContract{Wage: 3460000}, nil
				}
			},
			check: func(t *testing.T, s ProjectSummary) {
				if s.PlannedHours != 60 {
					t.Fatalf("planned = %v, want 60", s.PlannedHours)
				}
				if s.EffectiveHours != 30 {
					t.Fatalf("effective = %v, want 30", s.EffectiveHours)
				}
				if s.Utilization != 0.5 {
					t.Fatalf("utilization = %v, want 0.5", s.Utilization)
				}
				if s.UnbilledAmount != 7500000 {
					t.Fatalf("unbilled = %v, want 7500000", s.UnbilledAmount)
				}
				cost := 30 * (3460000 / StandardMonthlyHours)
				if s.CostAmount != cost {
					t.Fatalf("cost = %v, want %v", s.CostAmount, cost)
				}
				if s.MarginAmount != s.BillableAmount-cost {
					t.Fatalf("margin = %v, want %v", s.MarginAmount, s.BillableAmount-cost)
				}
			},
		},
		{
			name: "excludes billed timesheets",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{
						Base:           model.Base{ID: 1},
						OrganizationID: 1,
						BillingType:    BillingTypeTimeMaterial,
						BillableRate:   250000,
					}, nil
				}
				ctx.tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectTask, error) {
					return []*ProjectTask{{Base: model.Base{ID: 1}, ProjectID: 1, PlannedHours: 40}}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
				}
				ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
					return []*ProjectInvoiceLine{
						{Base: model.Base{ID: 1}, ProjectID: 1, TimesheetID: 1, Qty: 20, Amount: 5000000},
					}, nil
				}
				ctx.contracts.FindActiveByEmployeeFunc = func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
					return &payroll.EmploymentContract{Wage: 3460000}, nil
				}
			},
			check: func(t *testing.T, s ProjectSummary) {
				if s.BilledAmount != 5000000 {
					t.Fatalf("billed = %v, want 5000000", s.BilledAmount)
				}
				if s.UnbilledAmount != 0 {
					t.Fatalf("unbilled = %v, want 0", s.UnbilledAmount)
				}
				if s.BillableAmount != 5000000 {
					t.Fatalf("billable = %v, want 5000000", s.BillableAmount)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			summary, err := ctx.svc.Summary(context.Background(), 1)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, summary)
			}
		})
	}
}

func TestProjectService_BillTimeMaterial(t *testing.T) {
	var createdLines []*ProjectInvoiceLine

	tests := []struct {
		name       string
		input      BillTimeMaterialRequest
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, inv *accounting.Invoice)
	}{
		{
			name: "generates invoice from unbilled timesheets",
			input: BillTimeMaterialRequest{
				OrganizationID: 1,
				ProjectID:      1,
				JournalID:      1,
				Date:           time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{
						Base:           model.Base{ID: 1},
						OrganizationID: 1,
						ContactID:      7,
						BillingType:    BillingTypeTimeMaterial,
						BillableRate:   250000,
						DimensionID:    helper.Ptr(uint64(5)),
						State:          ProjectStateOpen,
					}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
				}
				ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
					return []*ProjectInvoiceLine{}, nil
				}
				ctx.invoiceLookup.ListByInvoiceFunc = func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
					return []*accounting.InvoiceLine{{Base: model.Base{ID: 9}, InvoiceID: 1}}, nil
				}
				createdLines = nil
				ctx.invoiceLines.CreateFunc = func(_ context.Context, line *ProjectInvoiceLine) (*ProjectInvoiceLine, error) {
					createdLines = append(createdLines, line)
					return line, nil
				}
				ctx.invoices.CreateFunc = func(_ context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
					return &accounting.Invoice{
						Base:        model.Base{ID: 1},
						ContactID:   request.ContactID,
						AmountTotal: amount.FromFloat64(5000000),
						Name:        helper.Ptr("INV/00001"),
					}, nil
				}
			},
			check: func(t *testing.T, inv *accounting.Invoice) {
				if inv.ID != 1 {
					t.Fatalf("invoice id = %d, want 1", inv.ID)
				}
				if len(createdLines) != 1 {
					t.Fatalf("join lines = %d, want 1", len(createdLines))
				}
				if createdLines[0].TimesheetID != 1 {
					t.Fatalf("join timesheet id = %d, want 1", createdLines[0].TimesheetID)
				}
				if createdLines[0].InvoiceLineID == nil || *createdLines[0].InvoiceLineID != 9 {
					t.Fatalf("join invoice line id = %v, want 9", createdLines[0].InvoiceLineID)
				}
			},
		},
		{
			name: "rejects when nothing to bill",
			input: BillTimeMaterialRequest{
				OrganizationID: 1,
				ProjectID:      1,
				JournalID:      1,
				Date:           time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{
						Base:           model.Base{ID: 1},
						OrganizationID: 1,
						BillingType:    BillingTypeTimeMaterial,
						BillableRate:   250000,
						State:          ProjectStateOpen,
					}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{}, nil
				}
				ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
					return []*ProjectInvoiceLine{}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectNothingToBill,
		},
		{
			name: "requires open project",
			input: BillTimeMaterialRequest{
				OrganizationID: 1,
				ProjectID:      1,
				JournalID:      1,
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateDraft, BillingType: BillingTypeTimeMaterial}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidState,
		},
		{
			name: "rejects fixed billing type",
			input: BillTimeMaterialRequest{
				OrganizationID: 1,
				ProjectID:      1,
				JournalID:      1,
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateOpen, BillingType: BillingTypeFixed}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectNotTimeMaterial,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			inv, err := ctx.svc.BillTimeMaterial(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, inv)
			}
		})
	}
}

type projectTestContext struct {
	projects      ProjectDAOMock
	tasks         ProjectTaskDAOMock
	milestones    ProjectMilestoneDAOMock
	invoiceLines  ProjectInvoiceLineDAOMock
	timesheets    payroll.TimesheetDAOMock
	contracts     payroll.EmploymentContractDAOMock
	contacts      contacts.ContactDAOMock
	accounts      projectAccountLookupMock
	dimensions    dao.CRUDMock[reference.Dimension]
	invoices      invoiceEngineMock
	invoiceLookup invoiceLineLookupMock
	svc           ProjectService
}

type projectAccountLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

func (m projectAccountLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
}

type invoiceEngineMock struct {
	CreateFunc func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m invoiceEngineMock) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 1}, ContactID: request.ContactID}, nil
}

type invoiceLineLookupMock struct {
	ListByInvoiceFunc func(ctx context.Context, invoiceID uint64) ([]*accounting.InvoiceLine, error)
}

func (m invoiceLineLookupMock) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*accounting.InvoiceLine, error) {
	if m.ListByInvoiceFunc != nil {
		return m.ListByInvoiceFunc(ctx, invoiceID)
	}
	return []*accounting.InvoiceLine{}, nil
}

func newProjectTestContext() *projectTestContext {
	ctx := &projectTestContext{
		accounts: projectAccountLookupMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 4200}, OrganizationID: 1, Active: true}}}, nil
			},
		},
		dimensions: dao.CRUDMock[reference.Dimension]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.Dimension, error) {
				return &reference.Dimension{Base: model.Base{ID: 5}}, nil
			},
		},
		contacts: contacts.ContactDAOMock{
			CRUDMock: dao.CRUDMock[contacts.Contact]{
				FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return &contacts.Contact{Base: model.Base{ID: 7}}, nil
				},
			},
		},
		contracts: payroll.EmploymentContractDAOMock{},
		invoices:  invoiceEngineMock{},
	}
	ctx.svc = NewProjectService(
		&ctx.projects,
		&ctx.tasks,
		&ctx.milestones,
		&ctx.invoiceLines,
		&ctx.timesheets,
		&ctx.contracts,
		&ctx.contacts,
		&ctx.accounts,
		&ctx.dimensions,
		&ctx.invoices,
		&ctx.invoiceLookup,
		inventory.TransactionerMock{},
	)
	ctx.svc.now = func() time.Time { return time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC) }
	return ctx
}
