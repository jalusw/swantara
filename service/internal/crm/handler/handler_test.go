package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func crmTestApp(orgID uint64, register func(fiber.Router, httpx.RouteGuards)) *fiber.App {
	app := fiber.New()
	if orgID != 0 {
		app.Use(func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, orgID)
			c.Locals(model.ActorKey, uint64(5))
			return c.Next()
		})
	}
	register(app, passthroughGuards())
	return app
}

func crmLeadTestSvc(
	leads crm.ProspectDAOMock,
	stages dao.CRUDMock[reference.PipelineStage],
	teams dao.CRUDMock[reference.SalesGroup],
	contacts contacts.ContactDAOMock,
) crm.ProspectService {
	return crm.NewProspectService(leads, stages, teams, contacts)
}

func newTestPipelineStageService(stages dao.CRUDMock[reference.PipelineStage]) crm.PipelineStageService {
	return crm.NewPipelineStageService(stages)
}

func crmActivityTestSvc(
	activities crm.ProspectActivityDAOMock,
	leads crm.ProspectDAOMock,
	contacts contacts.ContactDAOMock,
) crm.ProspectActivityService {
	return crm.NewProspectActivityService(activities, leads, contacts)
}

func sampleProspect() *crm.Prospect {
	return &crm.Prospect{
		Base:            model.Base{ID: 1},
		OrganizationID:  helper.Ptr(uint64(10)),
		Name:            "Acme Corp",
		Type:            crm.ProspectKindLead,
		ExpectedRevenue: 1000,
		Probability:     50,
		Priority:        1,
	}
}

func sampleCRMOpportunity() *crm.Prospect {
	prospect := sampleProspect()
	prospect.ID = 2
	prospect.Type = crm.ProspectKindOpportunity
	prospect.StageID = helper.Ptr(uint64(1))
	return prospect
}

func leadInOtherOrg() *crm.Prospect {
	prospect := sampleProspect()
	prospect.OrganizationID = helper.Ptr(uint64(99))
	return prospect
}

func samplePipelineStage() *reference.PipelineStage {
	return &reference.PipelineStage{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Prospecting",
		Sequence:       1,
		Probability:    50,
	}
}

func sampleSalesGroup() *reference.SalesGroup {
	return &reference.SalesGroup{
		Base:           model.Base{ID: 1},
		Name:           "Enterprise",
		OrganizationID: helper.Ptr(uint64(10)),
	}
}

func sampleProspectActivity() *crm.ProspectActivity {
	return &crm.ProspectActivity{
		Base:    model.Base{ID: 1},
		Type:    crm.ProspectActivityTypeCall,
		Summary: "Intro call",
	}
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
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
