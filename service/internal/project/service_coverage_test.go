package project

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProjectService_CreateProject_Coverage(t *testing.T) {
	tests := []struct {
		name       string
		input      *Project
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, p *Project)
	}{
		{
			name: "requires name",
			input: &Project{
				OrganizationID: 1,
				ContactID:      7,
				BillingType:    BillingTypeFixed,
			},
			wantErr:    true,
			wantErrVal: ErrProjectNameRequired,
		},
		{
			name: "rejects invalid dates",
			input: &Project{
				OrganizationID: 1,
				Name:           "Website Revamp",
				ContactID:      7,
				BillingType:    BillingTypeFixed,
				DateStart:      helper.Ptr(time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)),
				DateEnd:        helper.Ptr(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)),
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidDates,
		},
		{
			name: "propagates contact error",
			input: &Project{
				OrganizationID: 1,
				Name:           "Website Revamp",
				ContactID:      7,
				BillingType:    BillingTypeFixed,
			},
			setup: func(ctx *projectTestContext) {
				ctx.contacts.FindFunc = func(_ context.Context, _ uint64) (*contacts.Contact, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "propagates dimension error",
			input: &Project{
				OrganizationID: 1,
				Name:           "Website Revamp",
				ContactID:      7,
				BillingType:    BillingTypeFixed,
				DimensionID:    helper.Ptr(uint64(5)),
			},
			setup: func(ctx *projectTestContext) {
				ctx.dimensions.FindFunc = func(_ context.Context, _ uint64) (*reference.Dimension, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "rejects missing dimension",
			input: &Project{
				OrganizationID: 1,
				Name:           "Website Revamp",
				ContactID:      7,
				BillingType:    BillingTypeFixed,
				DimensionID:    helper.Ptr(uint64(5)),
			},
			setup: func(ctx *projectTestContext) {
				ctx.dimensions.FindFunc = func(_ context.Context, _ uint64) (*reference.Dimension, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectDimensionNotFound,
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

func TestProjectService_UpdateProject_Coverage(t *testing.T) {
	tests := []struct {
		name       string
		input      *Project
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, p *Project)
	}{
		{
			name: "updates and preserves state",
			input: &Project{
				Base:        model.Base{ID: 1},
				Name:        "Website Revamp",
				ContactID:   7,
				BillingType: BillingTypeFixed,
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateOpen}, nil
				}
				ctx.projects.UpdateFunc = func(_ context.Context, p *Project) (*Project, error) {
					return p, nil
				}
			},
			check: func(t *testing.T, p *Project) {
				if p == nil {
					t.Fatal("expected updated project, got nil")
				}
				if p.State != ProjectStateOpen {
					t.Fatalf("state = %s, want %s preserved", p.State, ProjectStateOpen)
				}
			},
		},
		{
			name: "returns not found",
			input: &Project{
				Base:        model.Base{ID: 1},
				Name:        "Website Revamp",
				ContactID:   7,
				BillingType: BillingTypeFixed,
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectNotFound,
		},
		{
			name: "propagates find error",
			input: &Project{
				Base:        model.Base{ID: 1},
				Name:        "Website Revamp",
				ContactID:   7,
				BillingType: BillingTypeFixed,
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "rejects invalid billing type",
			input: &Project{
				Base:        model.Base{ID: 1},
				Name:        "Website Revamp",
				ContactID:   7,
				BillingType: "nonsense",
			},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateOpen}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidBillingType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			updated, err := ctx.svc.UpdateProject(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, updated)
			}
		})
	}
}

func TestProjectService_SetProjectState_Coverage(t *testing.T) {
	tests := []struct {
		name       string
		projectID  uint64
		target     string
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
	}{
		{
			name:      "returns not found",
			projectID: 1,
			target:    ProjectStateOpen,
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectNotFound,
		},
		{
			name:      "propagates find error",
			projectID: 1,
			target:    ProjectStateOpen,
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := ctx.svc.SetProjectState(context.Background(), tt.projectID, tt.target)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestProjectService_CreateTask_Coverage(t *testing.T) {
	tests := []struct {
		name       string
		input      *ProjectTask
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, p *ProjectTask)
	}{
		{
			name:  "creates with default stage",
			input: &ProjectTask{ProjectID: 1, Name: "Design", PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateOpen}, nil
				}
				ctx.tasks.CreateFunc = func(_ context.Context, task *ProjectTask) (*ProjectTask, error) {
					return task, nil
				}
			},
			check: func(t *testing.T, p *ProjectTask) {
				if p == nil {
					t.Fatal("expected created task, got nil")
				}
				if p.Stage != TaskStageBacklog {
					t.Fatalf("stage = %s, want %s", p.Stage, TaskStageBacklog)
				}
			},
		},
		{
			name:  "returns not found",
			input: &ProjectTask{ProjectID: 1, Name: "Design", PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectNotFound,
		},
		{
			name:  "propagates find error",
			input: &ProjectTask{ProjectID: 1, Name: "Design", PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name:  "requires name",
			input: &ProjectTask{ProjectID: 1, PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateOpen}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectTaskNameRequired,
		},
		{
			name:  "rejects negative hours",
			input: &ProjectTask{ProjectID: 1, Name: "Design", PlannedHours: -1},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, State: ProjectStateOpen}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidHours,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			created, err := ctx.svc.CreateTask(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, created)
			}
		})
	}
}

func TestProjectService_UpdateTask_Coverage(t *testing.T) {
	tests := []struct {
		name       string
		input      *ProjectTask
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, p *ProjectTask)
	}{
		{
			name:  "updates and defaults stage",
			input: &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Design", PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.tasks.FindFunc = func(_ context.Context, _ uint64) (*ProjectTask, error) {
					return &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Design", Stage: TaskStageInProgress}, nil
				}
				ctx.tasks.UpdateFunc = func(_ context.Context, task *ProjectTask) (*ProjectTask, error) {
					return task, nil
				}
			},
			check: func(t *testing.T, p *ProjectTask) {
				if p == nil {
					t.Fatal("expected updated task, got nil")
				}
				if p.Stage != TaskStageBacklog {
					t.Fatalf("stage = %s, want %s", p.Stage, TaskStageBacklog)
				}
			},
		},
		{
			name:  "returns not found",
			input: &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Design", PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.tasks.FindFunc = func(_ context.Context, _ uint64) (*ProjectTask, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectTaskNotFound,
		},
		{
			name:  "propagates find error",
			input: &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Design", PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.tasks.FindFunc = func(_ context.Context, _ uint64) (*ProjectTask, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name:  "requires name",
			input: &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1, PlannedHours: 8},
			setup: func(ctx *projectTestContext) {
				ctx.tasks.FindFunc = func(_ context.Context, _ uint64) (*ProjectTask, error) {
					return &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectTaskNameRequired,
		},
		{
			name:  "rejects negative hours",
			input: &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Design", PlannedHours: -1},
			setup: func(ctx *projectTestContext) {
				ctx.tasks.FindFunc = func(_ context.Context, _ uint64) (*ProjectTask, error) {
					return &ProjectTask{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Design"}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectInvalidHours,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			updated, err := ctx.svc.UpdateTask(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, updated)
			}
		})
	}
}

func TestProjectService_CreateMilestone_Coverage(t *testing.T) {
	tests := []struct {
		name       string
		input      *ProjectMilestone
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
		check      func(t *testing.T, p *ProjectMilestone)
	}{
		{
			name:  "creates",
			input: &ProjectMilestone{ProjectID: 1, Name: "Launch"},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}}, nil
				}
				ctx.milestones.CreateFunc = func(_ context.Context, milestone *ProjectMilestone) (*ProjectMilestone, error) {
					return milestone, nil
				}
			},
			check: func(t *testing.T, p *ProjectMilestone) {
				if p == nil {
					t.Fatal("expected created milestone, got nil")
				}
				if p.ProjectID != 1 {
					t.Fatalf("project id = %d, want 1", p.ProjectID)
				}
			},
		},
		{
			name:  "returns not found",
			input: &ProjectMilestone{ProjectID: 1, Name: "Launch"},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectNotFound,
		},
		{
			name:  "propagates find error",
			input: &ProjectMilestone{ProjectID: 1, Name: "Launch"},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name:  "requires name",
			input: &ProjectMilestone{ProjectID: 1},
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}}, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectMilestoneNameRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			created, err := ctx.svc.CreateMilestone(context.Background(), tt.input)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, created)
			}
		})
	}
}

func TestProjectService_SetMilestoneReached_Coverage(t *testing.T) {
	tests := []struct {
		name        string
		milestoneID uint64
		reached     bool
		setup       func(ctx *projectTestContext)
		wantErr     bool
		wantErrVal  error
		check       func(t *testing.T, p *ProjectMilestone)
	}{
		{
			name:        "updates flag",
			milestoneID: 1,
			reached:     true,
			setup: func(ctx *projectTestContext) {
				ctx.milestones.FindFunc = func(_ context.Context, _ uint64) (*ProjectMilestone, error) {
					return &ProjectMilestone{Base: model.Base{ID: 1}, Name: "Launch", Reached: false}, nil
				}
				ctx.milestones.UpdateFunc = func(_ context.Context, milestone *ProjectMilestone) (*ProjectMilestone, error) {
					return milestone, nil
				}
			},
			check: func(t *testing.T, p *ProjectMilestone) {
				if p == nil {
					t.Fatal("expected updated milestone, got nil")
				}
				if !p.Reached {
					t.Fatal("expected reached milestone, got not reached")
				}
			},
		},
		{
			name:        "returns not found",
			milestoneID: 1,
			reached:     true,
			setup: func(ctx *projectTestContext) {
				ctx.milestones.FindFunc = func(_ context.Context, _ uint64) (*ProjectMilestone, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectMilestoneNotFound,
		},
		{
			name:        "propagates find error",
			milestoneID: 1,
			reached:     true,
			setup: func(ctx *projectTestContext) {
				ctx.milestones.FindFunc = func(_ context.Context, _ uint64) (*ProjectMilestone, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			updated, err := ctx.svc.SetMilestoneReached(context.Background(), tt.milestoneID, tt.reached)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.check != nil && !tt.wantErr {
				tt.check(t, updated)
			}
		})
	}
}

func TestProjectService_Summary_Coverage(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(ctx *projectTestContext)
		wantErr    bool
		wantErrVal error
	}{
		{
			name: "returns not found",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectNotFound,
		},
		{
			name: "propagates find error",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "propagates task error",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, BillableRate: 250000}, nil
				}
				ctx.tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectTask, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "propagates timesheet error",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, BillableRate: 250000}, nil
				}
				ctx.tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectTask, error) {
					return []*ProjectTask{}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "propagates invoice line error",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, BillableRate: 250000}, nil
				}
				ctx.tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectTask, error) {
					return []*ProjectTask{}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{}, nil
				}
				ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "propagates cost error",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, BillableRate: 250000}, nil
				}
				ctx.tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectTask, error) {
					return []*ProjectTask{}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
				}
				ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
					return []*ProjectInvoiceLine{}, nil
				}
				ctx.contracts.FindActiveByEmployeeFunc = func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
					return nil, errors.New("boom")
				}
			},
			wantErr: true,
		},
		{
			name: "rejects unattributable cost",
			setup: func(ctx *projectTestContext) {
				ctx.projects.FindFunc = func(_ context.Context, _ uint64) (*Project, error) {
					return &Project{Base: model.Base{ID: 1}, BillableRate: 250000}, nil
				}
				ctx.tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectTask, error) {
					return []*ProjectTask{}, nil
				}
				ctx.timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
				}
				ctx.invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*ProjectInvoiceLine, error) {
					return []*ProjectInvoiceLine{}, nil
				}
				ctx.contracts.FindActiveByEmployeeFunc = func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
					return nil, nil
				}
			},
			wantErr:    true,
			wantErrVal: ErrProjectCostNotAttributable,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newProjectTestContext()
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := ctx.svc.Summary(context.Background(), 1)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}
