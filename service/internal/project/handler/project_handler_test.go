package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProjectHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func() (project.ProjectDAOMock, project.ProjectService)
		body       string
		wantStatus int
	}{
		{
			name: "creates project with valid input",
			setupMocks: func() (project.ProjectDAOMock, project.ProjectService) {
				projects := project.ProjectDAOMock{
					CRUDMock: dao.CRUDMock[project.Project]{
						CreateFunc: func(_ context.Context, p *project.Project) (*project.Project, error) {
							return p, nil
						},
					},
				}
				svc := project.NewProjectService(
					projects,
					project.ProjectTaskDAOMock{},
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{CRUDMock: dao.CRUDMock[contacts.Contact]{
						FindFunc: func(_ context.Context, _ uint64) (*contacts.Contact, error) {
							return &contacts.Contact{Base: model.Base{ID: 7}}, nil
						},
					}},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
					inventory.TransactionerMock{},
				)
				return projects, svc
			},
			body:       `{"name":"Website Revamp","contact_id":7,"billing_type":"fixed","organization_id":1}`,
			wantStatus: http.StatusCreated,
		},
		{
			name: "rejects invalid billing type",
			setupMocks: func() (project.ProjectDAOMock, project.ProjectService) {
				projects := project.ProjectDAOMock{}
				return projects, projectSvc(projects)
			},
			body:       `{"name":"Website Revamp","contact_id":7,"billing_type":"nonsense"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, svc := tt.setupMocks()
			app := newProjectHandlerApp(projects, svc)

			resp, err := doRequest(app, http.MethodPost, "/projects", tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestProjectHandler_Get(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func() (project.ProjectDAOMock, project.ProjectService)
		path       string
		wantStatus int
	}{
		{
			name: "returns project",
			setupMocks: func() (project.ProjectDAOMock, project.ProjectService) {
				projects := project.ProjectDAOMock{
					CRUDMock: dao.CRUDMock[project.Project]{
						FindFunc: func(_ context.Context, _ uint64) (*project.Project, error) {
							return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, Name: "Website Revamp", State: project.ProjectStateOpen}, nil
						},
					},
				}
				return projects, projectSvc(projects)
			},
			path:       "/projects/1",
			wantStatus: http.StatusOK,
		},
		{
			name: "returns not found",
			setupMocks: func() (project.ProjectDAOMock, project.ProjectService) {
				projects := project.ProjectDAOMock{}
				return projects, projectSvc(projects)
			},
			path:       "/projects/1",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, svc := tt.setupMocks()
			app := newProjectHandlerApp(projects, svc)

			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestProjectHandler_BillTimeMaterial(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func() (project.ProjectDAOMock, project.ProjectService)
		body       string
		path       string
		wantStatus int
	}{
		{
			name: "generates invoice",
			setupMocks: func() (project.ProjectDAOMock, project.ProjectService) {
				projects := project.ProjectDAOMock{
					CRUDMock: dao.CRUDMock[project.Project]{
						FindFunc: func(_ context.Context, _ uint64) (*project.Project, error) {
							return &project.Project{
								Base:           model.Base{ID: 1},
								OrganizationID: 1,
								ContactID:      7,
								BillingType:    project.BillingTypeTimeMaterial,
								BillableRate:   250000,
								DimensionID:    helper.Ptr(uint64(5)),
								State:          project.ProjectStateOpen,
							}, nil
						},
					},
				}
				svc := project.NewProjectService(
					projects,
					project.ProjectTaskDAOMock{},
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{
						CRUDMock: dao.CRUDMock[project.ProjectInvoiceLine]{
							CreateFunc: func(_ context.Context, l *project.ProjectInvoiceLine) (*project.ProjectInvoiceLine, error) {
								return l, nil
							},
						},
					},
					payroll.TimesheetDAOMock{
						ListByProjectFunc: func(_ context.Context, _ uint64) ([]*payroll.Timesheet, error) {
							return []*payroll.Timesheet{{Base: model.Base{ID: 1}, EmployeeID: 7, Hours: 20}}, nil
						},
					},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{list: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: model.Base{ID: 4200}, OrganizationID: 1, Active: true}}}, nil
					}},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
					inventory.TransactionerMock{},
				)
				return projects, svc
			},
			body:       `{"journal_id":1,"date":"2026-08-01"}`,
			path:       "/projects/1/bill",
			wantStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, svc := tt.setupMocks()
			app := newProjectHandlerApp(projects, svc)

			resp, err := doRequestWithOrg(app, http.MethodPost, tt.path, tt.body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestProjectHandler_Summary(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func() (project.ProjectDAOMock, project.ProjectService)
		wantStatus int
	}{
		{
			name: "returns margin",
			setupMocks: func() (project.ProjectDAOMock, project.ProjectService) {
				projects := project.ProjectDAOMock{
					CRUDMock: dao.CRUDMock[project.Project]{
						FindFunc: func(_ context.Context, _ uint64) (*project.Project, error) {
							return &project.Project{Base: model.Base{ID: 1}, OrganizationID: 1, BillingType: project.BillingTypeTimeMaterial, BillableRate: 250000}, nil
						},
					},
				}
				svc := project.NewProjectService(
					projects,
					project.ProjectTaskDAOMock{},
					project.ProjectMilestoneDAOMock{},
					project.ProjectInvoiceLineDAOMock{},
					payroll.TimesheetDAOMock{},
					payroll.EmploymentContractDAOMock{},
					contacts.ContactDAOMock{},
					handlerAccountLookup{},
					dao.CRUDMock[reference.Dimension]{},
					handlerInvoiceEngine{},
					handlerInvoiceLineLookup{},
					inventory.TransactionerMock{},
				)
				return projects, svc
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projects, svc := tt.setupMocks()
			app := newProjectHandlerApp(projects, svc)

			resp, err := doRequest(app, http.MethodGet, "/projects/1/summary", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(1))
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}

func doRequestWithOrg(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Organization-ID", "1")
	return app.Test(req)
}

func newProjectHandlerApp(projects project.ProjectDAOMock, svc project.ProjectService) *fiber.App {
	handler := NewProjectHandler(svc)
	app := fiber.New()
	handler.Register(app, passthroughGuards())
	return app
}

func projectSvc(projects project.ProjectDAOMock) project.ProjectService {
	return project.NewProjectService(
		projects,
		project.ProjectTaskDAOMock{},
		project.ProjectMilestoneDAOMock{},
		project.ProjectInvoiceLineDAOMock{},
		payroll.TimesheetDAOMock{},
		payroll.EmploymentContractDAOMock{},
		contacts.ContactDAOMock{CRUDMock: dao.CRUDMock[contacts.Contact]{}},
		handlerAccountLookup{},
		dao.CRUDMock[reference.Dimension]{},
		handlerInvoiceEngine{},
		handlerInvoiceLineLookup{},
		inventory.TransactionerMock{},
	)
}

type handlerAccountLookup struct {
	list func(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

func (m handlerAccountLookup) List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error) {
	if m.list != nil {
		return m.list(ctx, q)
	}
	return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
}

type handlerInvoiceEngine struct {
	create func(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error)
}

func (m handlerInvoiceEngine) Create(ctx context.Context, request accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
	if m.create != nil {
		return m.create(ctx, request)
	}
	return &accounting.Invoice{Base: model.Base{ID: 1}, ContactID: request.ContactID, AmountTotal: amount.FromFloat64(5000000), Name: helper.Ptr("INV/00001")}, nil
}

type handlerInvoiceLineLookup struct {
	listByInvoice func(ctx context.Context, invoiceID uint64) ([]*accounting.InvoiceLine, error)
}

func (m handlerInvoiceLineLookup) ListByInvoice(ctx context.Context, invoiceID uint64) ([]*accounting.InvoiceLine, error) {
	if m.listByInvoice != nil {
		return m.listByInvoice(ctx, invoiceID)
	}
	return []*accounting.InvoiceLine{{Base: model.Base{ID: 9}, InvoiceID: invoiceID}}, nil
}
