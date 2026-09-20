package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func closedOpportunity() *crm.Prospect {
	prospect := sampleCRMOpportunity()
	now := time.Now()
	prospect.ClosedAt = &now
	return prospect
}

func wonStage() *reference.PipelineStage {
	stage := samplePipelineStage()
	stage.IsWon = true
	return stage
}

func opportunityApp(leads crm.ProspectDAOMock, stages dao.CRUDMock[reference.PipelineStage]) *fiber.App {
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	return crmTestApp(10, func(api fiber.Router, guards httpx.RouteGuards) {
		NewOpportunityHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})
}

func TestOpportunityHandler_List_ReturnsOpportunities(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return &query.Page[crm.Prospect]{Items: []*crm.Prospect{sampleCRMOpportunity()}, Count: 1}, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOpportunityHandler_List_ExportsCSV(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return &query.Page[crm.Prospect]{Items: []*crm.Prospect{sampleCRMOpportunity()}, Count: 1}, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOpportunityHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	stages := dao.CRUDMock[reference.PipelineStage]{}
	svc := crmLeadTestSvc(leads, stages, dao.CRUDMock[reference.SalesGroup]{}, contacts.ContactDAOMock{})
	app := crmTestApp(0, func(api fiber.Router, guards httpx.RouteGuards) {
		NewOpportunityHandler(newTestPipelineStageService(stages), svc).Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestOpportunityHandler_List_ReturnsServerErrorOnLeads(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_List_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[crm.Prospect], error) {
				return &query.Page[crm.Prospect]{Items: []*crm.Prospect{sampleCRMOpportunity()}, Count: 1}, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Get_ReturnsOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOpportunityHandler_Get_RejectsInvalidID(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Get_ReturnsNotFound(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Get_ReturnsNotFoundForLeadType(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Get_ReturnsNotFoundForOtherOrg(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				prospect := sampleCRMOpportunity()
				prospect.OrganizationID = helper.Ptr(uint64(99))
				return prospect, nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Get_ReturnsServerErrorOnFind(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Get_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodGet, "/crm/opportunities/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_CreatesOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			CreateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				prospect.ID = 2
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
	app := opportunityApp(leads, stages)

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_RejectsValidation(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"","stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_RejectsBadBody(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", `{"name":`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_RejectsInvalidExpectedClose(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"Acme Deal","stage_id":1,"expected_close":"soon"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_RejectsMissingStage(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"Acme Deal"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_RejectsUnknownStage(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return nil, nil
		},
	}
	app := opportunityApp(leads, stages)

	body := `{"name":"Acme Deal","stage_id":99}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_RejectsStageInOtherOrg(t *testing.T) {
	leads := crm.ProspectDAOMock{}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			stage := samplePipelineStage()
			stage.OrganizationID = helper.Ptr(uint64(99))
			return stage, nil
		},
	}
	app := opportunityApp(leads, stages)

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Create_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			CreateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				prospect.ID = 2
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
	app := opportunityApp(leads, stages)

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_UpdatesOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
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
	app := opportunityApp(leads, stages)

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/2", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_RejectsInvalidID(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_RejectsValidation(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/2", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_ReturnsNotFound(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/2", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_ReturnsNotFoundForLeadType(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_ReturnsNotFoundForOtherOrg(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				prospect := sampleCRMOpportunity()
				prospect.OrganizationID = helper.Ptr(uint64(99))
				return prospect, nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/2", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/2", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Update_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
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
	app := opportunityApp(leads, stages)

	body := `{"name":"Acme Deal","stage_id":1}`
	resp, err := doRequest(app, http.MethodPut, "/crm/opportunities/2", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Delete_DeletesOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodDelete, "/crm/opportunities/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestOpportunityHandler_Delete_RejectsInvalidID(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodDelete, "/crm/opportunities/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Delete_ReturnsNotFoundForLeadType(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodDelete, "/crm/opportunities/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodDelete, "/crm/opportunities/2", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_AdvanceStage_AdvancesOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
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
	app := opportunityApp(leads, stages)

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/advance-stage", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOpportunityHandler_AdvanceStage_RejectsInvalidID(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/abc/advance-stage", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_AdvanceStage_RejectsValidation(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/advance-stage", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_AdvanceStage_ReturnsConflictForClosedOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return closedOpportunity(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/advance-stage", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOpportunityHandler_AdvanceStage_ReturnsNotFoundForLeadType(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/1/advance-stage", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_AdvanceStage_ReturnsUnprocessableOnUnknownStage(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.PipelineStage, error) {
			return nil, nil
		},
	}
	app := opportunityApp(leads, stages)

	body := `{"stage_id":99}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/advance-stage", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_AdvanceStage_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
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
	app := opportunityApp(leads, stages)

	body := `{"stage_id":1}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/advance-stage", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Win_WinsOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
			UpdateFunc: func(_ context.Context, prospect *crm.Prospect) (*crm.Prospect, error) {
				return prospect, nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{wonStage()}, Count: 1}, nil
		},
	}
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/win", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOpportunityHandler_Win_RejectsInvalidID(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/abc/win", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Win_ReturnsConflictForNoWonStage(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
		},
	}
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/win", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOpportunityHandler_Win_ReturnsConflictForClosedOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return closedOpportunity(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/win", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOpportunityHandler_Win_ReturnsNotFoundForLeadType(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/1/win", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Win_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
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
	app := opportunityApp(leads, stages)

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/win", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOpportunityHandler_Lose_LosesOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
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
	app := opportunityApp(leads, stages)

	body := `{"lost_reason":"Budget"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/lose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOpportunityHandler_Lose_RejectsInvalidID(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"lost_reason":"Budget"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/abc/lose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Lose_RejectsMissingReason(t *testing.T) {
	app := opportunityApp(crm.ProspectDAOMock{}, dao.CRUDMock[reference.PipelineStage]{})

	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/lose", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Lose_RejectsBlankReason(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"lost_reason":"   "}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/lose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOpportunityHandler_Lose_ReturnsConflictForClosedOpportunity(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return closedOpportunity(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"lost_reason":"Budget"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/lose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOpportunityHandler_Lose_ReturnsNotFoundForLeadType(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleProspect(), nil
			},
		},
	}
	app := opportunityApp(leads, dao.CRUDMock[reference.PipelineStage]{})

	body := `{"lost_reason":"Budget"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/1/lose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOpportunityHandler_Lose_ReturnsServerErrorOnStages(t *testing.T) {
	leads := crm.ProspectDAOMock{
		CRUDMock: dao.CRUDMock[crm.Prospect]{
			FindFunc: func(_ context.Context, _ uint64) (*crm.Prospect, error) {
				return sampleCRMOpportunity(), nil
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
	app := opportunityApp(leads, stages)

	body := `{"lost_reason":"Budget"}`
	resp, err := doRequest(app, http.MethodPost, "/crm/opportunities/2/lose", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
