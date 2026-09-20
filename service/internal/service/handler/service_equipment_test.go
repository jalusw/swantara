package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

func equipmentHandlerApp(
	t *testing.T,
	equipments service.EquipmentDAOMock,
) *fiber.App {
	return serviceHandlerTest(t, equipments, service.ServiceContractDAOMock{}, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}, service.MaintenancePlanDAOMock{}, emptyServiceSvc(equipments, service.ServiceContractDAOMock{}, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}), emptyMaintenanceSvc(service.MaintenancePlanDAOMock{}, equipments, service.ServiceOrderDAOMock{}))
}

func equipmentHandlerAppNoTenant(
	t *testing.T,
	equipments service.EquipmentDAOMock,
) *fiber.App {
	return serviceHandlerTestNoTenant(t, equipments, service.ServiceContractDAOMock{}, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}, service.MaintenancePlanDAOMock{}, emptyServiceSvc(equipments, service.ServiceContractDAOMock{}, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}), emptyMaintenanceSvc(service.MaintenancePlanDAOMock{}, equipments, service.ServiceOrderDAOMock{}))
}

func TestServiceHandler_ListEquipments_ReturnsEquipments(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.Equipment], error) {
			return &query.Page[service.Equipment]{Items: []*service.Equipment{sampleEquipment()}, Count: 1}, nil
		},
	}
	app := equipmentHandlerApp(t, equipments)

	resp, err := doRequest(app, http.MethodGet, "/equipments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_ListEquipments_RejectsInvalidQuery(t *testing.T) {
	app := equipmentHandlerApp(t, service.EquipmentDAOMock{})

	resp, err := doRequest(app, http.MethodGet, "/equipments/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ListEquipments_RequiresTenant(t *testing.T) {
	app := equipmentHandlerAppNoTenant(t, service.EquipmentDAOMock{})

	resp, err := doRequest(app, http.MethodGet, "/equipments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestServiceHandler_ListEquipments_ReturnsServerError(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.Equipment], error) {
			return nil, errors.New("db down")
		},
	}
	app := equipmentHandlerApp(t, equipments)

	resp, err := doRequest(app, http.MethodGet, "/equipments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_GetEquipment_ReturnsEquipment(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return sampleEquipment(), nil
		},
	}
	app := equipmentHandlerApp(t, equipments)

	resp, err := doRequest(app, http.MethodGet, "/equipments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_GetEquipment_ReturnsNotFound(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return nil, nil
		},
	}
	app := equipmentHandlerApp(t, equipments)

	resp, err := doRequest(app, http.MethodGet, "/equipments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetEquipment_ReturnsNotFoundForForeignTenant(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return foreignEquipment(), nil
		},
	}
	app := equipmentHandlerApp(t, equipments)

	resp, err := doRequest(app, http.MethodGet, "/equipments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetEquipment_ReturnsServerError(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.Equipment, error) {
			return nil, errors.New("db down")
		},
	}
	app := equipmentHandlerApp(t, equipments)

	resp, err := doRequest(app, http.MethodGet, "/equipments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CreateEquipment_CreatesEquipment(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		CreateFunc: func(_ context.Context, entity *service.Equipment) (*service.Equipment, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	app := equipmentHandlerApp(t, equipments)

	body := `{"name":"Chiller","category":"hvac"}`
	resp, err := doRequest(app, http.MethodPost, "/equipments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestServiceHandler_CreateEquipment_RejectsValidation(t *testing.T) {
	app := equipmentHandlerApp(t, service.EquipmentDAOMock{})

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/equipments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateEquipment_RejectsMissingTenant(t *testing.T) {
	app := equipmentHandlerAppNoTenant(t, service.EquipmentDAOMock{})

	body := `{"name":"Chiller"}`
	resp, err := doRequest(app, http.MethodPost, "/equipments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateEquipment_ReturnsServerError(t *testing.T) {
	equipments := service.EquipmentDAOMock{
		CreateFunc: func(_ context.Context, _ *service.Equipment) (*service.Equipment, error) {
			return nil, errors.New("db down")
		},
	}
	app := equipmentHandlerApp(t, equipments)

	body := `{"name":"Chiller"}`
	resp, err := doRequest(app, http.MethodPost, "/equipments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
