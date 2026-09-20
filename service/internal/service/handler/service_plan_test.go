package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/service"
	"gorm.io/gorm"
)

func planHandlerApp(
	t *testing.T,
	plans service.MaintenancePlanDAOMock,
	equipments service.EquipmentDAOMock,
	orders service.ServiceOrderDAOMock,
	maintenance service.MaintenanceService,
) *fiber.App {
	return serviceHandlerTest(t, equipments, service.ServiceContractDAOMock{}, orders, service.ServiceOrderLineDAOMock{}, plans, emptyServiceSvc(equipments, service.ServiceContractDAOMock{}, orders, service.ServiceOrderLineDAOMock{}), maintenance)
}

func planHandlerAppNoTenant(
	t *testing.T,
	plans service.MaintenancePlanDAOMock,
	equipments service.EquipmentDAOMock,
	orders service.ServiceOrderDAOMock,
	maintenance service.MaintenanceService,
) *fiber.App {
	return serviceHandlerTestNoTenant(t, equipments, service.ServiceContractDAOMock{}, orders, service.ServiceOrderLineDAOMock{}, plans, emptyServiceSvc(equipments, service.ServiceContractDAOMock{}, orders, service.ServiceOrderLineDAOMock{}), maintenance)
}

func TestServiceHandler_ListMaintenancePlans_ReturnsPlans(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.MaintenancePlan], error) {
			return &query.Page[service.MaintenancePlan]{Items: []*service.MaintenancePlan{samplePlan()}, Count: 1}, nil
		},
	}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_ListMaintenancePlans_RejectsInvalidQuery(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ListMaintenancePlans_ReturnsServerError(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.MaintenancePlan], error) {
			return nil, errors.New("db down")
		},
	}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_GetMaintenancePlan_ReturnsPlan(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.MaintenancePlan, error) {
			return samplePlan(), nil
		},
	}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return sampleEquipment(), nil
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_GetMaintenancePlan_ReturnsServerError(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.MaintenancePlan, error) {
			return nil, errors.New("db down")
		},
	}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_GetMaintenancePlan_ReturnsNotFound(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.MaintenancePlan, error) {
			return nil, nil
		},
	}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetMaintenancePlan_ReturnsNotFoundForForeignEquipment(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.MaintenancePlan, error) {
			return samplePlan(), nil
		},
	}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return foreignEquipment(), nil
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetMaintenancePlan_ReturnsNotFoundOnEquipmentError(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.MaintenancePlan, error) {
			return samplePlan(), nil
		},
	}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return nil, errors.New("db down")
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetMaintenancePlan_ReturnsNotFoundForPlanWithoutEquipment(t *testing.T) {
	plan := samplePlan()
	plan.EquipmentID = nil
	plans := service.MaintenancePlanDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.MaintenancePlan, error) {
			return plan, nil
		},
	}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/maintenance-plans/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CreateMaintenancePlan_CreatesPlan(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		CreateFunc: func(_ context.Context, entity *service.MaintenancePlan) (*service.MaintenancePlan, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return sampleEquipment(), nil
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	body := `{"equipment_id":1,"name":"Quarterly service","interval_days":30,"next_due":"2026-08-01"}`
	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestServiceHandler_CreateMaintenancePlan_RejectsValidation(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateMaintenancePlan_RejectsInvalidNextDue(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	body := `{"equipment_id":1,"name":"Quarterly service","interval_days":30,"next_due":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateMaintenancePlan_ReturnsServerErrorOnEquipmentLookup(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return nil, errors.New("db down")
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	body := `{"equipment_id":1,"name":"Quarterly service","interval_days":30,"next_due":"2026-08-01"}`
	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CreateMaintenancePlan_ReturnsNotFound(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return nil, nil
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	body := `{"equipment_id":1,"name":"Quarterly service","interval_days":30,"next_due":"2026-08-01"}`
	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CreateMaintenancePlan_ReturnsNotFoundForForeignEquipment(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return foreignEquipment(), nil
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	body := `{"equipment_id":1,"name":"Quarterly service","interval_days":30,"next_due":"2026-08-01"}`
	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CreateMaintenancePlan_ReturnsServerErrorOnPlanCreate(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		CreateFunc: func(_ context.Context, _ *service.MaintenancePlan) (*service.MaintenancePlan, error) {
			return nil, errors.New("db down")
		},
	}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return sampleEquipment(), nil
		},
	}
	app := planHandlerApp(t, plans, equipments, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, equipments, service.ServiceOrderDAOMock{}))

	body := `{"equipment_id":1,"name":"Quarterly service","interval_days":30,"next_due":"2026-08-01"}`
	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_GenerateMaintenanceOrders_GeneratesOrders(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*service.MaintenancePlan, error) {
			return []*service.MaintenancePlan{samplePlan()}, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *service.MaintenancePlan) (*service.MaintenancePlan, error) {
			return entity, nil
		},
	}
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return sampleEquipment(), nil
		},
	}
	orders := service.ServiceOrderDAOMock{
		CreateFunc: func(_ context.Context, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	app := planHandlerApp(t, plans, equipments, orders, emptyMaintenanceSvc(plans, equipments, orders))

	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/generate-orders", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_GenerateMaintenanceOrders_RejectsMissingTenant(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{}
	app := planHandlerAppNoTenant(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/generate-orders", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_GenerateMaintenanceOrders_ReturnsServerError(t *testing.T) {
	plans := service.MaintenancePlanDAOMock{
		ListDueFunc: func(_ context.Context, _ uint64, _ time.Time) ([]*service.MaintenancePlan, error) {
			return nil, errors.New("db down")
		},
	}
	app := planHandlerApp(t, plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}, emptyMaintenanceSvc(plans, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))

	resp, err := doRequest(app, http.MethodPost, "/maintenance-plans/generate-orders", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_WriteServiceError_MapsErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "equipment not found", err: service.ErrEquipmentNotFound, want: http.StatusNotFound},
		{name: "contract not found", err: service.ErrContractNotFound, want: http.StatusNotFound},
		{name: "order not found", err: service.ErrOrderNotFound, want: http.StatusNotFound},
		{name: "plan not found", err: service.ErrPlanNotFound, want: http.StatusNotFound},
		{name: "equipment name required", err: service.ErrEquipmentNameRequired, want: http.StatusUnprocessableEntity},
		{name: "contract name required", err: service.ErrContractNameRequired, want: http.StatusUnprocessableEntity},
		{name: "contract invalid state", err: service.ErrContractInvalidState, want: http.StatusUnprocessableEntity},
		{name: "order name required", err: service.ErrOrderNameRequired, want: http.StatusUnprocessableEntity},
		{name: "order invalid type", err: service.ErrOrderInvalidType, want: http.StatusUnprocessableEntity},
		{name: "order invalid state", err: service.ErrOrderInvalidState, want: http.StatusUnprocessableEntity},
		{name: "line invalid type", err: service.ErrOrderLineInvalidType, want: http.StatusUnprocessableEntity},
		{name: "line invalid qty", err: service.ErrOrderLineInvalidQty, want: http.StatusUnprocessableEntity},
		{name: "line invalid price", err: service.ErrOrderLineInvalidPrice, want: http.StatusUnprocessableEntity},
		{name: "order no lines", err: service.ErrOrderNoLines, want: http.StatusUnprocessableEntity},
		{name: "order not done", err: service.ErrOrderNotDone, want: http.StatusUnprocessableEntity},
		{name: "order not completed", err: service.ErrOrderNotCompleted, want: http.StatusUnprocessableEntity},
		{name: "order requires journal", err: service.ErrOrderRequiresJournal, want: http.StatusUnprocessableEntity},
		{name: "order requires accounts", err: service.ErrOrderRequiresAccounts, want: http.StatusUnprocessableEntity},
		{name: "order requires contact", err: service.ErrOrderRequiresContact, want: http.StatusUnprocessableEntity},
		{name: "order empty lines", err: service.ErrOrderEmptyLines, want: http.StatusUnprocessableEntity},
		{name: "order contact for cogs", err: service.ErrOrderContactForCOGS, want: http.StatusUnprocessableEntity},
		{name: "plan name required", err: service.ErrPlanNameRequired, want: http.StatusUnprocessableEntity},
		{name: "plan interval invalid", err: service.ErrPlanIntervalInvalid, want: http.StatusUnprocessableEntity},
		{name: "plan next due required", err: service.ErrPlanNextDueRequired, want: http.StatusUnprocessableEntity},
		{name: "unexpected error", err: errors.New("boom"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/err", func(c fiber.Ctx) error {
				return writeServiceError(c, tt.err)
			})

			resp, err := doRequest(app, http.MethodPost, "/err", "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.want {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}
