package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestProspectHandler_List_ReturnsLeads(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return &query.Page[crm.Prospect]{Items: []*crm.Prospect{sampleProspect()}, Count: 1}, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProspectHandler_List_ExportsCSV(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return &query.Page[crm.Prospect]{Items: []*crm.Prospect{sampleProspect()}, Count: 1}, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProspectHandler_List_RejectsInvalidQuery(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(0, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestProspectHandler_List_ReturnsServerErrorOnLeads(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_List_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return &query.Page[crm.Prospect]{Items: []*crm.Prospect{sampleProspect()}, Count: 1}, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Get_ReturnsLead(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProspectHandler_Get_RejectsInvalidID(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Get_ReturnsServerError(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Get_ReturnsNotFound(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProspectHandler_Get_ReturnsNotFoundForOtherOrg(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return leadInOtherOrg(), nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProspectHandler_Get_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Create_CreatesLead(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			CreateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				prospect.ID = 1
				return prospect, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp","probability":50}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProspectHandler_Create_RejectsValidation(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":"","probability":50}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Create_RejectsBadBody(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestProspectHandler_Create_RejectsInvalidExpectedClose(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp","expected_close":"tomorrow"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Create_MapsServiceErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"prospect not found", crm.ErrLeadNotFound, 404},
		{"stage not found", crm.ErrStageNotFound, 422},
		{"stage organization", crm.ErrStageOrganization, 422},
		{"stage required", crm.ErrStageRequired, 422},
		{"contact not found", crm.ErrContactNotFound, 422},
		{"salesperson not found", crm.ErrSalespersonNotFound, 422},
		{"sales team not found", crm.ErrSalesGroupNotFound, 422},
		{"invalid prospect type", crm.ErrInvalidLeadType, 422},
		{"invalid probability", crm.ErrInvalidProbability, 422},
		{"invalid revenue", crm.ErrInvalidRevenue, 422},
		{"invalid priority", crm.ErrInvalidPriority, 422},
		{"name required", crm.ErrLeadNameRequired, 422},
		{"not prospect", crm.ErrNotLead, 409},
		{"not opportunity", crm.ErrNotOpportunity, 409},
		{"prospect closed", crm.ErrLeadClosed, 409},
		{"no won stage", crm.ErrNoWonStage, 409},
		{"lost reason required", crm.ErrLostReasonRequired, 422},
		{"internal error", errors.New("db down"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			leads := crm.ProspectDAOMock{
				CRUDMock: dao.CRUDMock[crm.Prospect]{
					CreateFunc: func(_ context.Context, _ *crm.Prospect) (*crm.Prospect, error) {
						return nil, tt.err
					},
				},
			}
			svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
			app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
				NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
			})

			body := `{"name":"Acme Corp","probability":50}`
			resp, err := doRequest(app, http.MethodPost, "/crm/leads/", body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestProspectHandler_Create_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			CreateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				prospect.ID = 1
				return prospect, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp","probability":50}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Update_UpdatesLead(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
			UpdateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				return prospect, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp","probability":60}`
	resp, err := doRequest(app, http.MethodPut, "/crm/leads/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProspectHandler_Update_RejectsInvalidID(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/leads/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Update_RejectsValidation(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPut, "/crm/leads/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Update_ReturnsNotFound(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/leads/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProspectHandler_Update_ReturnsNotFoundForOtherOrg(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return leadInOtherOrg(), nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/leads/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProspectHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/leads/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Update_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
			UpdateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				return prospect, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/crm/leads/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Delete_DeletesLead(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodDelete, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestProspectHandler_Delete_RejectsInvalidID(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodDelete, "/crm/leads/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Delete_ReturnsNotFound(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodDelete, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProspectHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodDelete, "/crm/leads/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_PromotesLead(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
			UpdateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				return prospect, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return samplePipelineStage(), nil
		},
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"stage_id":1,"expected_revenue":2000,"probability":70}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_RejectsInvalidID(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/abc/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_RejectsValidation(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_ReturnsNotFound(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_RejectsOpportunityLead(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_ReturnsNotFoundForOtherOrg(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return leadInOtherOrg(), nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_ReturnsServerErrorOnFind(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_RejectsInvalidExpectedClose(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	svc := crmLeadTestSvc(leads, dao.CRUDMock[reference.PipelineStage]{}, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(dao.CRUDMock[reference.PipelineStage]{}), svc).Register(api, guards)
	})

	body := `{"stage_id":1,"expected_close":"soon"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_ReturnsUnprocessableOnStageNotFound(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return nil, nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"stage_id":99}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_ReturnsUnprocessableOnInvalidRevenue(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return samplePipelineStage(), nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"stage_id":1,"expected_revenue":-1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_ReturnsUnprocessableOnInvalidProbability(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return samplePipelineStage(), nil
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"stage_id":1,"probability":150}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProspectHandler_Promote_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
			UpdateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				return prospect, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return samplePipelineStage(), nil
		},
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewProspectHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/leads/1/promote", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
