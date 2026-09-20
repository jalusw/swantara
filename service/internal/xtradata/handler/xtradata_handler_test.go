package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

func modelBase() model.Base {
	return model.Base{ID: 1}
}

func systemConfigHandlerTest(
	t *testing.T,
	svc xtradata.ConfigService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewSystemConfigHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func integrationEventHandlerTest(
	t *testing.T,
	events xtradata.IntegrationEventDAOMock,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	h := NewIntegrationEventHandler(xtradata.NewEventService(events, nil))
	h.Register(app, passthroughGuards())
	return app
}

func sampleSystemConfig() *reference.SystemConfig {
	return &reference.SystemConfig{
		Base:           modelBase(),
		OrganizationID: nil,
		Key:            "currency",
		Value:          json.RawMessage(`{"rate":1}`),
	}
}

func sampleIntegrationEvent() *xtradata.IntegrationEvent {
	return &xtradata.IntegrationEvent{
		Base:    modelBase(),
		Topic:   "contact.created",
		Payload: json.RawMessage(`{"id":1}`),
		Status:  "pending",
		Retries: 0,
	}
}

func TestSystemConfigHandler_List_ReturnsConfigs(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{sampleSystemConfig()}, Count: 1}, nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSystemConfigHandler_List_ExportsCSV(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return &query.Page[reference.SystemConfig]{Items: []*reference.SystemConfig{sampleSystemConfig()}, Count: 1}, nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSystemConfigHandler_List_RejectsInvalidQuery(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSystemConfigHandler_List_ReturnsServerError(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
			return nil, errors.New("db down")
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Get_ReturnsConfig(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return sampleSystemConfig(), nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Get_ReturnsNotFound(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return nil, nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Get_RejectsInvalidID(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Get_ReturnsServerError(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return nil, errors.New("db down")
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/system-configs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Create_CreatesConfig(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*reference.SystemConfig, error) {
			return nil, nil
		},
		CreateFunc: func(_ context.Context, config *reference.SystemConfig) (*reference.SystemConfig, error) {
			config.ID = 1
			return config, nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{"key":"currency","value":{"rate":1}}`
	resp, err := doRequest(app, http.MethodPost, "/system-configs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Create_ReturnsConflictOnDuplicate(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*reference.SystemConfig, error) {
			return sampleSystemConfig(), nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{"key":"currency","value":{"rate":1}}`
	resp, err := doRequest(app, http.MethodPost, "/system-configs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Create_RejectsValidation(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/system-configs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Create_ReturnsServerError(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*reference.SystemConfig, error) {
			return nil, errors.New("db down")
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{"key":"currency","value":{"rate":1}}`
	resp, err := doRequest(app, http.MethodPost, "/system-configs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Update_UpdatesConfig(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return sampleSystemConfig(), nil
		},
		UpdateFunc: func(_ context.Context, config *reference.SystemConfig) (*reference.SystemConfig, error) {
			return config, nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{"value":{"rate":2}}`
	resp, err := doRequest(app, http.MethodPut, "/system-configs/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Update_ReturnsNotFound(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return nil, nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{"value":{"rate":2}}`
	resp, err := doRequest(app, http.MethodPut, "/system-configs/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Update_RejectsInvalidID(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{"value":{"rate":2}}`
	resp, err := doRequest(app, http.MethodPut, "/system-configs/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Update_RejectsValidation(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/system-configs/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Update_ReturnsServerError(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return nil, errors.New("db down")
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	body := `{"value":{"rate":2}}`
	resp, err := doRequest(app, http.MethodPut, "/system-configs/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Delete_DeletesConfig(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return sampleSystemConfig(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodDelete, "/system-configs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Delete_ReturnsNotFound(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return nil, nil
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodDelete, "/system-configs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestSystemConfigHandler_Delete_ReturnsServerError(t *testing.T) {
	configs := xtradata.SystemConfigDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
			return sampleSystemConfig(), nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	}
	svc := xtradata.NewConfigService(configs)
	app := systemConfigHandlerTest(t, svc)

	resp, err := doRequest(app, http.MethodDelete, "/system-configs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_List_ReturnsEvents(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[xtradata.IntegrationEvent], error) {
			return &query.Page[xtradata.IntegrationEvent]{Items: []*xtradata.IntegrationEvent{sampleIntegrationEvent()}, Count: 1}, nil
		},
	}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_List_ExportsCSV(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[xtradata.IntegrationEvent], error) {
			return &query.Page[xtradata.IntegrationEvent]{Items: []*xtradata.IntegrationEvent{sampleIntegrationEvent()}, Count: 1}, nil
		},
	}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_List_RejectsInvalidQuery(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_List_ReturnsServerError(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[xtradata.IntegrationEvent], error) {
			return nil, errors.New("db down")
		},
	}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_Get_ReturnsEvent(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
			return sampleIntegrationEvent(), nil
		},
	}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_Get_ReturnsNotFound(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
			return nil, nil
		},
	}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_Get_RejectsInvalidID(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestIntegrationEventHandler_Get_ReturnsServerError(t *testing.T) {
	events := xtradata.IntegrationEventDAOMock{
		FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
			return nil, errors.New("db down")
		},
	}
	app := integrationEventHandlerTest(t, events)

	resp, err := doRequest(app, http.MethodGet, "/integration-events/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
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
