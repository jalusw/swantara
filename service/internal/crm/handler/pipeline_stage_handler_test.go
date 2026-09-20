package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func stageApp(stages dao.CRUD[reference.PipelineStage], orgID uint64) *fiber.App {
	return crmTestApp(orgID, func(api fiber.Router, guards httpx.RouteGuards) {
		NewPipelineStageHandler(crm.NewPipelineStageService(stages)).Register(api, guards)
	})
}

func teamApp(teams dao.CRUD[reference.SalesGroup], orgID uint64) *fiber.App {
	return crmTestApp(orgID, func(api fiber.Router, guards httpx.RouteGuards) {
		NewSalesGroupHandler(crm.NewSalesGroupService(teams)).Register(api, guards)
	})
}

func pipelineApp(svc crm.PipelineService, orgID uint64) *fiber.App {
	return crmTestApp(orgID, func(api fiber.Router, guards httpx.RouteGuards) {
		NewPipelineHandler(svc).Register(api, guards)
	})
}

func TestPipelineStageHandler_List_ReturnsStages(t *testing.T) {
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	app := stageApp(stages, 10)

	resp, err := doRequest(app, http.MethodGet, "/crm/stages", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPipelineStageHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	stages := dao.CRUDMock[reference.PipelineStage]{}
	app := stageApp(stages, 0)

	resp, err := doRequest(app, http.MethodGet, "/crm/stages", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestPipelineStageHandler_List_ReturnsServerError(t *testing.T) {
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	app := stageApp(stages, 10)

	resp, err := doRequest(app, http.MethodGet, "/crm/stages", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSalesGroupHandler_List_ReturnsTeams(t *testing.T) {
	teams := dao.CRUDMock[reference.SalesGroup]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalesGroup], error) {
			return &query.Page[reference.SalesGroup]{Items: []*reference.SalesGroup{sampleSalesGroup()}, Count: 1}, nil
		},
	}
	app := teamApp(teams, 10)

	resp, err := doRequest(app, http.MethodGet, "/crm/teams", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSalesGroupHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	teams := dao.CRUDMock[reference.SalesGroup]{}
	app := teamApp(teams, 0)

	resp, err := doRequest(app, http.MethodGet, "/crm/teams", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestSalesGroupHandler_List_ReturnsServerError(t *testing.T) {
	teams := dao.CRUDMock[reference.SalesGroup]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SalesGroup], error) {
			return nil, errors.New("db down")
		},
	}
	app := teamApp(teams, 10)

	resp, err := doRequest(app, http.MethodGet, "/crm/teams", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPipelineHandler_Forecast_ReturnsForecast(t *testing.T) {
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{samplePipelineStage()}, Count: 1}, nil
		},
	}
	leads := crm.ProspectDAOMock{
		ListOpenFunc: func(_ context.Context, _ *uint64) ([]*crm.Prospect, error) {
			prospect := sampleProspect()
			prospect.StageID = helper.Ptr(uint64(1))
			return []*crm.Prospect{prospect}, nil
		},
		CountWonFunc: func(_ context.Context, _ *uint64) (int64, error) {
			return 2, nil
		},
		CountLostFunc: func(_ context.Context, _ *uint64) (int64, error) {
			return 1, nil
		},
	}
	svc := crm.NewPipelineService(leads, stages)
	app := pipelineApp(svc, 10)

	resp, err := doRequest(app, http.MethodGet, "/crm/pipeline", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPipelineHandler_Forecast_ReturnsServerError(t *testing.T) {
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return nil, errors.New("db down")
		},
	}
	svc := crm.NewPipelineService(crm.ProspectDAOMock{}, stages)
	app := pipelineApp(svc, 10)

	resp, err := doRequest(app, http.MethodGet, "/crm/pipeline", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPipelineHandler_Forecast_ReturnsServerErrorOnOpen(t *testing.T) {
	stages := dao.CRUDMock[reference.PipelineStage]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.PipelineStage], error) {
			return &query.Page[reference.PipelineStage]{Items: []*reference.PipelineStage{}, Count: 0}, nil
		},
	}
	leads := crm.ProspectDAOMock{
		ListOpenFunc: func(_ context.Context, _ *uint64) ([]*crm.Prospect, error) {
			return nil, errors.New("db down")
		},
	}
	svc := crm.NewPipelineService(leads, stages)
	app := pipelineApp(svc, 10)

	resp, err := doRequest(app, http.MethodGet, "/crm/pipeline", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
