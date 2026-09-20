package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProjectHandler_List(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[project.Project], error) {
					return &query.Page[project.Project]{
						Items: []*project.Project{{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Website Revamp"}},
						Count: 1,
					}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/?page=1&size=10",
			body:   "",
			status: http.StatusOK,
		},
		{
			name: "invalid query",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/?filter=bogus:eq:1",
			body:   "",
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "unauthorized",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), false)
			},
			method: http.MethodGet,
			path:   "/projects/",
			body:   "",
			status: http.StatusUnauthorized,
		},
		{
			name: "dao error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[project.Project], error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/",
			body:   "",
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_Get_ErrorPaths(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "tenant mismatch",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 99}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/1",
			body:   "",
			status: http.StatusNotFound,
		},
		{
			name: "dao error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/1",
			body:   "",
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_Create_ErrorPaths(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "organization required",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), false)
			},
			method: http.MethodPost,
			path:   "/projects",
			body:   `{"name":"Website Revamp","contact_id":7,"billing_type":"fixed"}`,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "contact not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects",
			body:   `{"name":"Website Revamp","contact_id":7,"billing_type":"fixed"}`,
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.CreateFunc = func(_ context.Context, p *project.Project) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				svc := buildProjectSvc(
					projects,
					project.ProjectTaskDAOMock{},
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{CRUDMock: dao.CRUDMock[contacts.Contact]{FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
						return &contacts.Contact{Base: model.Base{ID: 7}}, nil
					}}},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPost,
			path:   "/projects",
			body:   `{"name":"Website Revamp","contact_id":7,"billing_type":"fixed"}`,
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_Update(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, State: project.ProjectStateOpen}, nil
				}
				projects.UpdateFunc = func(_ context.Context, p *project.Project) (*project.Project, error) {
					return p, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, &contacts.Contact{Base: model.Base{ID: 7}}), true)
			},
			method: http.MethodPut,
			path:   "/projects/1",
			body:   `{"name":"Website Revamp","contact_id":7,"billing_type":"fixed"}`,
			status: http.StatusOK,
		},
		{
			name: "validation error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1",
			body:   `{"name":"Website Revamp","contact_id":7,"billing_type":"nonsense"}`,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1",
			body:   `{"name":"Website Revamp","contact_id":7,"billing_type":"fixed"}`,
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1",
			body:   `{"name":"Website Revamp","contact_id":7,"billing_type":"fixed"}`,
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_SetState(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, State: project.ProjectStateDraft}, nil
				}
				projects.UpdateFunc = func(_ context.Context, p *project.Project) (*project.Project, error) {
					return p, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1/state",
			body:   `{"state":"open"}`,
			status: http.StatusOK,
		},
		{
			name: "validation error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1/state",
			body:   `{"state":"nonsense"}`,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "invalid transition",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, State: project.ProjectStateClosed}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1/state",
			body:   `{"state":"open"}`,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1/state",
			body:   `{"state":"open"}`,
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1/state",
			body:   `{"state":"open"}`,
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_ListTasks(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				tasks := project.ProjectTaskDAOMock{}
				tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*project.ProjectTask, error) {
					return []*project.ProjectTask{
						{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Design", Stage: project.TaskStageBacklog},
					}, nil
				}
				return projectHandlerApp(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/1/tasks",
			body:   "",
			status: http.StatusOK,
		},
		{
			name: "dao error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				tasks := project.ProjectTaskDAOMock{}
				tasks.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*project.ProjectTask, error) {
					return nil, errors.New("boom")
				}
				svc := buildProjectSvc(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, payroll.TimesheetDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, handlerAccountLookup{}, dao.CRUDMock[reference.Dimension]{}, handlerInvoiceEngine{}, handlerInvoiceLineLookup{})
				return projectHandlerApp(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodGet,
			path:   "/projects/1/tasks",
			body:   "",
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_CreateTask(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, State: project.ProjectStateOpen}, nil
				}
				tasks := project.ProjectTaskDAOMock{}
				tasks.CreateFunc = func(_ context.Context, task *project.ProjectTask) (*project.ProjectTask, error) {
					return task, nil
				}
				svc := buildProjectSvc(
					projects,
					tasks,
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPost,
			path:   "/projects/1/tasks",
			body:   `{"name":"Design","planned_hours":8}`,
			status: http.StatusCreated,
		},
		{
			name: "validation error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, State: project.ProjectStateOpen}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/tasks",
			body:   `{"planned_hours":8}`,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/tasks",
			body:   `{"name":"Design","planned_hours":8}`,
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/tasks",
			body:   `{"name":"Design","planned_hours":8}`,
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_UpdateTask(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				tasks := project.ProjectTaskDAOMock{}
				tasks.FindFunc = func(_ context.Context, _ uint64) (*project.ProjectTask, error) {
					return &project.ProjectTask{Base: model.Base{ID: 2}, ProjectID: 1, Name: "Design"}, nil
				}
				tasks.UpdateFunc = func(_ context.Context, task *project.ProjectTask) (*project.ProjectTask, error) {
					return task, nil
				}
				svc := buildProjectSvc(
					projects,
					tasks,
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPut,
			path:   "/projects/1/tasks/2",
			body:   `{"name":"Design","planned_hours":8}`,
			status: http.StatusOK,
		},
		{
			name: "validation error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				tasks := project.ProjectTaskDAOMock{}
				tasks.FindFunc = func(_ context.Context, _ uint64) (*project.ProjectTask, error) {
					return &project.ProjectTask{Base: model.Base{ID: 2}, ProjectID: 1}, nil
				}
				svc := buildProjectSvc(
					projects,
					tasks,
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPut,
			path:   "/projects/1/tasks/2",
			body:   `{"planned_hours":8}`,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				tasks := project.ProjectTaskDAOMock{}
				svc := buildProjectSvc(
					projects,
					tasks,
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPut,
			path:   "/projects/1/tasks/2",
			body:   `{"name":"Design","planned_hours":8}`,
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				tasks := project.ProjectTaskDAOMock{}
				tasks.FindFunc = func(_ context.Context, _ uint64) (*project.ProjectTask, error) {
					return nil, errors.New("boom")
				}
				svc := buildProjectSvc(
					projects,
					tasks,
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, tasks, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPut,
			path:   "/projects/1/tasks/2",
			body:   `{"name":"Design","planned_hours":8}`,
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_ListMilestones(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				milestones := project.ProjectMilestoneDAOMock{}
				milestones.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*project.ProjectMilestone, error) {
					return []*project.ProjectMilestone{{Base: model.Base{ID: 1}, ProjectID: 1, Name: "Launch"}}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, milestones, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/1/milestones",
			body:   "",
			status: http.StatusOK,
		},
		{
			name: "dao error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				milestones := project.ProjectMilestoneDAOMock{}
				milestones.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*project.ProjectMilestone, error) {
					return nil, errors.New("boom")
				}
				svc := buildProjectSvc(projects, project.ProjectTaskDAOMock{}, milestones, project.ProjectInvoiceLineDAOMock{}, payroll.TimesheetDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, handlerAccountLookup{}, dao.CRUDMock[reference.Dimension]{}, handlerInvoiceEngine{}, handlerInvoiceLineLookup{})
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, milestones, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodGet,
			path:   "/projects/1/milestones",
			body:   "",
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_CreateMilestone(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1}, nil
				}
				milestones := project.ProjectMilestoneDAOMock{}
				milestones.CreateFunc = func(_ context.Context, milestone *project.ProjectMilestone) (*project.ProjectMilestone, error) {
					return milestone, nil
				}
				svc := buildProjectSvc(
					projects,
					project.ProjectTaskDAOMock{},
					milestones,
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, milestones, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPost,
			path:   "/projects/1/milestones",
			body:   `{"name":"Launch"}`,
			status: http.StatusCreated,
		},
		{
			name: "validation error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/milestones",
			body:   `{}`,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/milestones",
			body:   `{"name":"Launch"}`,
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/milestones",
			body:   `{"name":"Launch"}`,
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_SetMilestoneReached(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "success",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				milestones := project.ProjectMilestoneDAOMock{}
				milestones.FindFunc = func(_ context.Context, _ uint64) (*project.ProjectMilestone, error) {
					return &project.ProjectMilestone{Base: model.Base{ID: 2}, ProjectID: 1, Name: "Launch"}, nil
				}
				milestones.UpdateFunc = func(_ context.Context, milestone *project.ProjectMilestone) (*project.ProjectMilestone, error) {
					return milestone, nil
				}
				svc := buildProjectSvc(
					projects,
					project.ProjectTaskDAOMock{},
					milestones,
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, milestones, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPut,
			path:   "/projects/1/milestones/2/reached",
			body:   `{"reached":true}`,
			status: http.StatusOK,
		},
		{
			name: "bad request",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPut,
			path:   "/projects/1/milestones/2/reached",
			body:   `{"reached":`,
			status: http.StatusBadRequest,
		},
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				milestones := project.ProjectMilestoneDAOMock{}
				svc := buildProjectSvc(
					projects,
					project.ProjectTaskDAOMock{},
					milestones,
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, milestones, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPut,
			path:   "/projects/1/milestones/2/reached",
			body:   `{"reached":true}`,
			status: http.StatusNotFound,
		},
		{
			name: "internal error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				milestones := project.ProjectMilestoneDAOMock{}
				milestones.FindFunc = func(_ context.Context, _ uint64) (*project.ProjectMilestone, error) {
					return nil, errors.New("boom")
				}
				svc := buildProjectSvc(
					projects,
					project.ProjectTaskDAOMock{},
					milestones,
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, milestones, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodPut,
			path:   "/projects/1/milestones/2/reached",
			body:   `{"reached":true}`,
			status: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_BillTimeMaterial_ErrorPaths(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		do     func(*fiber.App, string, string, string) (*http.Response, error)
		status int
	}{
		{
			name: "invalid date",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/bill",
			body:   `{"journal_id":1,"date":"bogus"}`,
			do:     doRequestWithOrg,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "organization required",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/abc/bill",
			body:   `{"journal_id":1,"date":"2026-08-01"}`,
			do:     doRequest,
			status: http.StatusUnprocessableEntity,
		},
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/bill",
			body:   `{"journal_id":1,"date":"2026-08-01"}`,
			do:     doRequestWithOrg,
			status: http.StatusNotFound,
		},
		{
			name: "tenant mismatch",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 99}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/bill",
			body:   `{"journal_id":1,"date":"2026-08-01"}`,
			do:     doRequestWithOrg,
			status: http.StatusNotFound,
		},
		{
			name: "dao error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodPost,
			path:   "/projects/1/bill",
			body:   `{"journal_id":1,"date":"2026-08-01"}`,
			do:     doRequestWithOrg,
			status: http.StatusInternalServerError,
		},
		{
			name: "nothing to bill",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, BillingType: project.BillingTypeTimeMaterial, BillableRate: 250000, State: project.ProjectStateOpen}, nil
				}
				timesheets := payroll.TimesheetDAOMock{}
				timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{}, nil
				}
				invoiceLines := project.ProjectInvoiceLineDAOMock{}
				invoiceLines.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*project.ProjectInvoiceLine, error) {
					return []*project.ProjectInvoiceLine{}, nil
				}
				svc := buildProjectSvc(
					projects,
					project.ProjectTaskDAOMock{},
					project.ProjectMilestoneDAOMock{},
					invoiceLines,
					timesheets,
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, invoiceLines, svc, true)
			},
			method: http.MethodPost,
			path:   "/projects/1/bill",
			body:   `{"journal_id":1,"date":"2026-08-01"}`,
			do:     doRequestWithOrg,
			status: http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := tt.do(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestProjectHandler_Summary_ErrorPaths(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() *fiber.App
		method string
		path   string
		body   string
		status int
	}{
		{
			name: "not found",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/1/summary",
			body:   "",
			status: http.StatusNotFound,
		},
		{
			name: "tenant mismatch",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 99}, nil
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/1/summary",
			body:   "",
			status: http.StatusNotFound,
		},
		{
			name: "dao error",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return nil, errors.New("boom")
				}
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, projectSvcWithContact(projects, nil), true)
			},
			method: http.MethodGet,
			path:   "/projects/1/summary",
			body:   "",
			status: http.StatusInternalServerError,
		},
		{
			name: "cost not attributable",
			setup: func() *fiber.App {
				projects := project.ProjectDAOMock{}
				projects.FindFunc = func(_ context.Context, _ uint64) (*project.Project, error) {
					return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, BillingType: project.BillingTypeTimeMaterial, BillableRate: 250000}, nil
				}
				timesheets := payroll.TimesheetDAOMock{}
				timesheets.ListByProjectFunc = func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
					return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
				}
				contracts := payroll.EmploymentContractDAOMock{}
				svc := buildProjectSvc(
					projects,
					project.ProjectTaskDAOMock{},
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					timesheets,
					contracts,
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
				)
				return projectHandlerApp(projects, project.ProjectTaskDAOMock{}, project.ProjectMilestoneDAOMock{}, project.ProjectInvoiceLineDAOMock{}, svc, true)
			},
			method: http.MethodGet,
			path:   "/projects/1/summary",
			body:   "",
			status: http.StatusUnprocessableEntity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := tt.setup()
			resp, err := doRequest(app, tt.method, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.status {
				t.Fatalf("expected %d, got %d", tt.status, resp.StatusCode)
			}
		})
	}
}

func TestWriteProjectError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "project not found", err: project.ErrProjectNotFound, status: http.StatusNotFound},
		{name: "contact not found", err: project.ErrProjectContactNotFound, status: http.StatusNotFound},
		{name: "dimension not found", err: project.ErrProjectDimensionNotFound, status: http.StatusNotFound},
		{name: "task not found", err: project.ErrProjectTaskNotFound, status: http.StatusNotFound},
		{name: "milestone not found", err: project.ErrProjectMilestoneNotFound, status: http.StatusNotFound},
		{name: "invalid state", err: project.ErrProjectInvalidState, status: http.StatusUnprocessableEntity},
		{name: "not time material", err: project.ErrProjectNotTimeMaterial, status: http.StatusUnprocessableEntity},
		{name: "no billable rate", err: project.ErrProjectNoBillableRate, status: http.StatusUnprocessableEntity},
		{name: "nothing to bill", err: project.ErrProjectNothingToBill, status: http.StatusUnprocessableEntity},
		{name: "no revenue account", err: project.ErrProjectNoRevenueAccount, status: http.StatusUnprocessableEntity},
		{name: "cost not attributable", err: project.ErrProjectCostNotAttributable, status: http.StatusUnprocessableEntity},
		{name: "name required", err: project.ErrProjectNameRequired, status: http.StatusUnprocessableEntity},
		{name: "invalid billing type", err: project.ErrProjectInvalidBillingType, status: http.StatusUnprocessableEntity},
		{name: "invalid dates", err: project.ErrProjectInvalidDates, status: http.StatusUnprocessableEntity},
		{name: "task name required", err: project.ErrProjectTaskNameRequired, status: http.StatusUnprocessableEntity},
		{name: "invalid hours", err: project.ErrProjectInvalidHours, status: http.StatusUnprocessableEntity},
		{name: "milestone name required", err: project.ErrProjectMilestoneNameRequired, status: http.StatusUnprocessableEntity},
		{name: "internal error", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeProjectError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}

func projectGuards(withTenant bool) httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			if withTenant {
				c.Locals(httpx.LocalOrganizationID, uint64(1))
				c.Locals(model.ActorKey, uint64(5))
			}
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func projectHandlerApp(projects project.ProjectDAOMock, tasks project.ProjectTaskDAOMock, milestones project.ProjectMilestoneDAOMock, invoiceLines project.ProjectInvoiceLineDAOMock, svc project.ProjectService, withTenant bool) *fiber.App {
	handler := NewProjectHandler(svc)
	app := fiber.New()
	handler.Register(app, projectGuards(withTenant))
	return app
}

func buildProjectSvc(
	projects project.ProjectDAOMock,
	tasks project.ProjectTaskDAOMock,
	milestones project.ProjectMilestoneDAOMock,
	invoiceLines project.ProjectInvoiceLineDAOMock,
	timesheets payroll.TimesheetDAOMock,
	contracts payroll.EmploymentContractDAOMock,
	contacts contacts.ContactDAOMock,
	accounts handlerAccountLookup,
	dimensions dao.CRUDMock[reference.Dimension],
	invoices handlerInvoiceEngine,
	invoiceLookup handlerInvoiceLineLookup,
) project.ProjectService {
	return project.NewProjectService(
		projects, tasks, milestones, invoiceLines,
		timesheets, contracts, contacts, accounts, dimensions,
		invoices, invoiceLookup, inventory.TransactionerMock{},
	)
}

func projectSvcWithContact(projects project.ProjectDAOMock, contact *contacts.Contact) project.ProjectService {
	return buildProjectSvc(
		projects,
		project.ProjectTaskDAOMock{},
		project.ProjectMilestoneDAOMock{},
		project.ProjectInvoiceLineDAOMock{},
		payroll.TimesheetDAOMock{},
		payroll.EmploymentContractDAOMock{},
		contacts.ContactDAOMock{CRUDMock: dao.CRUDMock[contacts.Contact]{FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
			return contact, nil
		}}},
		handlerAccountLookup{},
		dao.CRUDMock[reference.Dimension]{},
		handlerInvoiceEngine{},
		handlerInvoiceLineLookup{},
	)
}

func writeErrorStatus(path string, fn func(c fiber.Ctx) error) int {
	app := fiber.New()
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}
