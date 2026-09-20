package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/service"
	"gorm.io/gorm"
)

func orderHandlerApp(
	t *testing.T,
	orders service.ServiceOrderDAOMock,
	lines service.ServiceOrderLineDAOMock,
	svc service.ServiceService,
) *fiber.App {
	return serviceHandlerTest(t, service.EquipmentDAOMock{}, service.ServiceContractDAOMock{}, orders, lines, service.MaintenancePlanDAOMock{}, svc, emptyMaintenanceSvc(service.MaintenancePlanDAOMock{}, service.EquipmentDAOMock{}, orders))
}

func orderHandlerAppNoTenant(
	t *testing.T,
	orders service.ServiceOrderDAOMock,
	lines service.ServiceOrderLineDAOMock,
	svc service.ServiceService,
) *fiber.App {
	return serviceHandlerTestNoTenant(t, service.EquipmentDAOMock{}, service.ServiceContractDAOMock{}, orders, lines, service.MaintenancePlanDAOMock{}, svc, emptyMaintenanceSvc(service.MaintenancePlanDAOMock{}, service.EquipmentDAOMock{}, orders))
}

func orderSvc(
	orders service.ServiceOrderDAOMock,
	lines service.ServiceOrderLineDAOMock,
	poster accounting.Poster,
	invoices service.InvoiceBuilder,
) service.ServiceService {
	return service.NewTestServiceService(
		service.EquipmentDAOMock{},
		service.ServiceContractDAOMock{},
		orders,
		lines,
		poster,
		invoices,
		service.ServiceTransactionerMock{},
	)
}

func TestServiceHandler_ListOrders_ReturnsOrders(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.ServiceOrder], error) {
			return &query.Page[service.ServiceOrder]{Items: []*service.ServiceOrder{sampleOrder()}, Count: 1}, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrders_RejectsInvalidQuery(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrders_RequiresTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerAppNoTenant(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrders_ReturnsServerError(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[service.ServiceOrder], error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_GetOrder_ReturnsOrder(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_GetOrder_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetOrder_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_GetOrder_ReturnsServerError(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CreateOrder_CreatesOrder(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		CreateFunc: func(_ context.Context, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, entity *service.ServiceOrderLine) (*service.ServiceOrderLine, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"name":"Fix chiller","type":"repair","lines":[{"type":"part","qty":1,"unit_price":100}]}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestServiceHandler_CreateOrder_RejectsValidation(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"name":"","type":"repair","lines":[{"type":"part","qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateOrder_RejectsMissingTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerAppNoTenant(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"name":"Fix chiller","type":"repair","lines":[{"type":"part","qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateOrder_RejectsInvalidLinePrice(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"name":"Fix chiller","type":"repair","lines":[{"type":"part","qty":1,"unit_price":-5}]}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CreateOrder_ReturnsServerError(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		CreateFunc: func(_ context.Context, _ *service.ServiceOrder) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"name":"Fix chiller","type":"repair","lines":[{"type":"part","qty":1}]}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrderLines_ReturnsLines(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*service.ServiceOrderLine, error) {
			return []*service.ServiceOrderLine{sampleLine()}, nil
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrderLines_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrderLines_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrderLines_ReturnsServerErrorOnOrderLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_ListOrderLines_ReturnsServerErrorOnLinesLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*service.ServiceOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodGet, "/service-orders/1/lines", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_AddOrderLine_AddsLine(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, entity *service.ServiceOrderLine) (*service.ServiceOrderLine, error) {
			entity.ID = 1
			return entity, nil
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"type":"part","qty":1,"unit_price":100}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/lines", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestServiceHandler_AddOrderLine_RejectsValidation(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"type":"invalid","qty":1}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/lines", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_AddOrderLine_ReturnsServerErrorOnOrderLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"type":"part","qty":1}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/lines", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_AddOrderLine_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"type":"part","qty":1}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/lines", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_AddOrderLine_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"type":"part","qty":1}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/lines", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_AddOrderLine_RejectsInvalidState(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return doneOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"type":"part","qty":1}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/lines", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_AddOrderLine_ReturnsServerError(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *service.ServiceOrderLine) (*service.ServiceOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"type":"part","qty":1}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/lines", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_SchedulesOrder(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			return entity, nil
		},
		UpdateFunc: func(_ context.Context, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			return entity, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"scheduled_date":"2026-08-01T10:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_RejectsValidation(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_RejectsInvalidDate(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"scheduled_date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_ReturnsServerErrorOnOrderLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"scheduled_date":"2026-08-01T10:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"scheduled_date":"2026-08-01T10:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"scheduled_date":"2026-08-01T10:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_RejectsInvalidState(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return doneOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"scheduled_date":"2026-08-01T10:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_ScheduleOrder_ReturnsServerErrorOnUpdate(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *service.ServiceOrder) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"scheduled_date":"2026-08-01T10:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/schedule", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_StartOrder_StartsOrder(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return scheduledOrder(), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			return entity, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_StartOrder_ReturnsServerErrorOnOrderLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_StartOrder_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_StartOrder_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_StartOrder_RejectsInvalidState(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_StartOrder_ReturnsServerErrorOnUpdate(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return scheduledOrder(), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *service.ServiceOrder) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_CompletesOrder(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return inProgressOrder(), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			return entity, nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*service.ServiceOrderLine, error) {
			return []*service.ServiceOrderLine{sampleLine()}, nil
		},
	}
	poster := service.ServicePosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, poster, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"2026-02-01","cogs_account_id":200,"stock_cost_account_id":300}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_RejectsValidation(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_RejectsInvalidDate(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"bogus","cogs_account_id":200,"stock_cost_account_id":300}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_ReturnsServerErrorOnOrderLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"2026-02-01","cogs_account_id":200,"stock_cost_account_id":300}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"2026-02-01","cogs_account_id":200,"stock_cost_account_id":300}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"2026-02-01","cogs_account_id":200,"stock_cost_account_id":300}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_RejectsInvalidState(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return scheduledOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"2026-02-01","cogs_account_id":200,"stock_cost_account_id":300}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CompleteOrder_ReturnsServerErrorOnLinesLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return inProgressOrder(), nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*service.ServiceOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"2026-02-01","cogs_account_id":200,"stock_cost_account_id":300}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/complete", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_BillsOrder(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return doneOrder(), nil
		},
		UpdateFunc: func(_ context.Context, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			return entity, nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*service.ServiceOrderLine, error) {
			return []*service.ServiceOrderLine{sampleLine()}, nil
		},
	}
	invoices := service.ServiceInvoiceBuilderMock{
		CreateFunc: func(_ context.Context, _ accounting.CreateInvoiceRequest) (*accounting.Invoice, error) {
			return &accounting.Invoice{Base: model.Base{ID: 500}}, nil
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, invoices))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_RejectsValidation(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_RejectsInvalidDate(t *testing.T) {
	orders := service.ServiceOrderDAOMock{}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"date":"bogus","revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_ReturnsServerErrorOnOrderLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_RejectsOrderNotDone(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_RejectsMissingContact(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return doneOrderWithoutContact(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_RejectsNoBillableLines(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return doneOrder(), nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*service.ServiceOrderLine, error) {
			return []*service.ServiceOrderLine{{Type: service.LineTypePart, Qty: 1, UnitPrice: 0, Billable: false}}, nil
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_BillOrder_ReturnsServerErrorOnLinesLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return doneOrder(), nil
		},
	}
	lines := service.ServiceOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*service.ServiceOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, lines, orderSvc(orders, lines, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	body := `{"journal_id":3,"revenue_account_id":400}`
	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/bill", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CancelOrder_CancelsOrder(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, entity *service.ServiceOrder) (*service.ServiceOrder, error) {
			return entity, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestServiceHandler_CancelOrder_ReturnsServerErrorOnOrderLookup(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestServiceHandler_CancelOrder_ReturnsNotFound(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return nil, nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CancelOrder_ReturnsNotFoundForForeignTenant(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return foreignOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestServiceHandler_CancelOrder_RejectsInvalidState(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return doneOrder(), nil
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestServiceHandler_CancelOrder_ReturnsServerErrorOnUpdate(t *testing.T) {
	orders := service.ServiceOrderDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*service.ServiceOrder, error) {
			return sampleOrder(), nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *service.ServiceOrder) (*service.ServiceOrder, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderHandlerApp(t, orders, service.ServiceOrderLineDAOMock{}, orderSvc(orders, service.ServiceOrderLineDAOMock{}, service.ServicePosterMock{}, service.ServiceInvoiceBuilderMock{}))

	resp, err := doRequest(app, http.MethodPost, "/service-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
