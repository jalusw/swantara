package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/service"
)

func contractHandlerApp(
	t *testing.T,
	contracts service.ServiceContractDAOMock,
) *fiber.App {
	return serviceHandlerTest(t, service.EquipmentDAOMock{}, contracts, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}, service.MaintenancePlanDAOMock{}, emptyServiceSvc(service.EquipmentDAOMock{}, contracts, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}), emptyMaintenanceSvc(service.MaintenancePlanDAOMock{}, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))
}

func contractHandlerAppNoTenant(
	t *testing.T,
	contracts service.ServiceContractDAOMock,
) *fiber.App {
	return serviceHandlerTestNoTenant(t, service.EquipmentDAOMock{}, contracts, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}, service.MaintenancePlanDAOMock{}, emptyServiceSvc(service.EquipmentDAOMock{}, contracts, service.ServiceOrderDAOMock{}, service.ServiceOrderLineDAOMock{}), emptyMaintenanceSvc(service.MaintenancePlanDAOMock{}, service.EquipmentDAOMock{}, service.ServiceOrderDAOMock{}))
}

func TestServiceHandler_ListContracts_ReturnsContracts(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.ServiceContract], error) {
			return &query.Page[service.ServiceContract]{Items: []*service.ServiceContract{sampleContract()}, Count: 1}, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_ListContracts_RejectsInvalidQuery(t *testing.T) {
	app := contractHandlerApp(t, service.ServiceContractDAOMock{})

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ListContracts_RequiresTenant(t *testing.T) {
	app := contractHandlerAppNoTenant(t, service.ServiceContractDAOMock{})

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestServiceHandler_ListContracts_ReturnsServerError(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.ServiceContract], error) {
			return nil, errors.New("db down")
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_GetContract_ReturnsContract(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return sampleContract(), nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_GetContract_ReturnsNotFound(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return nil, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetContract_ReturnsNotFoundForForeignTenant(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return foreignContract(), nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetContract_ReturnsServerError(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodGet, "/service-contracts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CreateContract_CreatesContract(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		CreateFunc: func(_ context.Context, entity *service.ServiceContract) (*service.ServiceContract, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	body := `{"name":"Gold SLA"}`
	resp, err := doRequest(app, http.MethodPost, "/service-contracts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestServiceHandler_CreateContract_RejectsValidation(t *testing.T) {
	app := contractHandlerApp(t, service.ServiceContractDAOMock{})

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/service-contracts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateContract_RejectsMissingTenant(t *testing.T) {
	app := contractHandlerAppNoTenant(t, service.ServiceContractDAOMock{})

	body := `{"name":"Gold SLA"}`
	resp, err := doRequest(app, http.MethodPost, "/service-contracts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateContract_ReturnsServerError(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		CreateFunc: func(_ context.Context, _ *service.ServiceContract) (*service.ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	app := contractHandlerApp(t, contracts)

	body := `{"name":"Gold SLA"}`
	resp, err := doRequest(app, http.MethodPost, "/service-contracts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_ActivateContract_ActivatesContract(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return draftContract(), nil
		},
		UpdateFunc: func(_ context.Context, entity *service.ServiceContract) (*service.ServiceContract, error) {
			return entity, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_ActivateContract_ReturnsNotFound(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return nil, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_ActivateContract_ReturnsNotFoundForForeignTenant(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			contract := draftContract()
			contract.OrganizationID = helper.Ptr(uint64(99))
			return contract, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_ActivateContract_ReturnsServerErrorOnLookup(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_ActivateContract_RejectsNonDraftState(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return sampleContract(), nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ActivateContract_RejectsMissingDates(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return draftContractWithoutDates(), nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ActivateContract_ReturnsServerErrorOnUpdate(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return draftContract(), nil
		},
		UpdateFunc: func(_ context.Context, _ *service.ServiceContract) (*service.ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/activate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CancelContract_CancelsContract(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return sampleContract(), nil
		},
		UpdateFunc: func(_ context.Context, entity *service.ServiceContract) (*service.ServiceContract, error) {
			return entity, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_CancelContract_ReturnsNotFound(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return nil, nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CancelContract_ReturnsNotFoundForForeignTenant(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return foreignContract(), nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CancelContract_ReturnsServerErrorOnFind(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CancelContract_RejectsNonActiveState(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return draftContract(), nil
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CancelContract_ReturnsServerErrorOnUpdate(t *testing.T) {
	contracts := service.ServiceContractDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceContract, error) {
			return sampleContract(), nil
		},
		UpdateFunc: func(_ context.Context, _ *service.ServiceContract) (*service.ServiceContract, error) {
			return nil, errors.New("db down")
		},
	}
	app := contractHandlerApp(t, contracts)

	resp, err := doRequest(app, http.MethodPost, "/service-contracts/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
