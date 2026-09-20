package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

func serviceHandlerTest(
	t *testing.T,
	equipments service.EquipmentDAOMock,
	contracts service.ServiceContractDAOMock,
	orders service.ServiceOrderDAOMock,
	lines service.ServiceOrderLineDAOMock,
	plans service.MaintenancePlanDAOMock,
	svc service.ServiceService,
	maintenance service.MaintenanceService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	h := NewServiceHandler(svc, maintenance)
	h.Register(app, passthroughGuards())
	return app
}

func serviceHandlerTestNoTenant(
	t *testing.T,
	equipments service.EquipmentDAOMock,
	contracts service.ServiceContractDAOMock,
	orders service.ServiceOrderDAOMock,
	lines service.ServiceOrderLineDAOMock,
	plans service.MaintenancePlanDAOMock,
	svc service.ServiceService,
	maintenance service.MaintenanceService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewServiceHandler(svc, maintenance)
	h.Register(app, passthroughGuards())
	return app
}

func emptyServiceSvc(
	equipments service.EquipmentDAOMock,
	contracts service.ServiceContractDAOMock,
	orders service.ServiceOrderDAOMock,
	lines service.ServiceOrderLineDAOMock,
) service.ServiceService {
	return service.NewTestServiceService(
		equipments,
		contracts,
		orders,
		lines,
		service.ServicePosterMock{},
		service.ServiceInvoiceBuilderMock{},
		service.ServiceTransactionerMock{},
	)
}

func emptyMaintenanceSvc(
	plans service.MaintenancePlanDAOMock,
	equipments service.EquipmentDAOMock,
	orders service.ServiceOrderDAOMock,
) service.MaintenanceService {
	return service.NewTestMaintenanceService(plans, equipments, orders, service.ServiceTransactionerMock{})
}

func sampleEquipment() *service.Equipment {
	return &service.Equipment{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Chiller",
		Category:       "hvac",
	}
}

func foreignEquipment() *service.Equipment {
	equipment := sampleEquipment()
	equipment.OrganizationID = helper.Ptr(uint64(99))
	return equipment
}

func sampleContract() *service.ServiceContract {
	return &service.ServiceContract{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Gold SLA",
		State:          service.ContractStateActive,
	}
}

func draftContract() *service.ServiceContract {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	return &service.ServiceContract{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Gold SLA",
		State:          service.ContractStateDraft,
		DateStart:      &start,
		DateEnd:        &end,
	}
}

func draftContractWithoutDates() *service.ServiceContract {
	contract := draftContract()
	contract.DateStart = nil
	contract.DateEnd = nil
	return contract
}

func foreignContract() *service.ServiceContract {
	contract := sampleContract()
	contract.OrganizationID = helper.Ptr(uint64(99))
	return contract
}

func sampleOrder() *service.ServiceOrder {
	return &service.ServiceOrder{
		Base:           model.Base{ID: 1},
		OrganizationID: helper.Ptr(uint64(10)),
		Name:           "Fix chiller",
		Type:           service.OrderTypeRepair,
		State:          service.OrderStateNew,
	}
}

func foreignOrder() *service.ServiceOrder {
	order := sampleOrder()
	order.OrganizationID = helper.Ptr(uint64(99))
	return order
}

func scheduledOrder() *service.ServiceOrder {
	order := sampleOrder()
	order.State = service.OrderStateScheduled
	return order
}

func inProgressOrder() *service.ServiceOrder {
	order := sampleOrder()
	order.State = service.OrderStateInProgress
	return order
}

func doneOrder() *service.ServiceOrder {
	order := sampleOrder()
	order.State = service.OrderStateDone
	order.ContactID = helper.Ptr(uint64(7))
	return order
}

func doneOrderWithoutContact() *service.ServiceOrder {
	order := doneOrder()
	order.ContactID = nil
	return order
}

func sampleLine() *service.ServiceOrderLine {
	return &service.ServiceOrderLine{
		Base:           model.Base{ID: 1},
		ServiceOrderID: 1,
		Type:           service.LineTypePart,
		Qty:            1,
		UnitCost:       50,
		UnitPrice:      100,
		Billable:       true,
	}
}

func samplePlan() *service.MaintenancePlan {
	next := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	return &service.MaintenancePlan{
		Base:         model.Base{ID: 1},
		EquipmentID:  helper.Ptr(uint64(5)),
		Name:         "Quarterly service",
		IntervalDays: 30,
		NextDue:      &next,
		Active:       true,
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
