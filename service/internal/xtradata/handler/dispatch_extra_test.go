package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

func TestSystemConfigHandler_Delete_Branches(t *testing.T) {
	t.Run("rejects invalid id", func(t *testing.T) {
		app := systemConfigHandlerTest(t, xtradata.NewConfigService(xtradata.SystemConfigDAOMock{}))
		resp, err := doRequest(app, http.MethodDelete, "/system-configs/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("reports lookup error and missing", func(t *testing.T) {
		failing := xtradata.NewConfigService(xtradata.SystemConfigDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
				return nil, errors.New("db down")
			},
		})
		app := systemConfigHandlerTest(t, failing)
		resp, err := doRequest(app, http.MethodDelete, "/system-configs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		app = systemConfigHandlerTest(t, xtradata.NewConfigService(xtradata.SystemConfigDAOMock{}))
		resp, err = doRequest(app, http.MethodDelete, "/system-configs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("reports delete error", func(t *testing.T) {
		svc := xtradata.NewConfigService(xtradata.SystemConfigDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
				return sampleSystemConfig(), nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error { return errors.New("db down") },
		})
		app := systemConfigHandlerTest(t, svc)
		resp, err := doRequest(app, http.MethodDelete, "/system-configs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("deletes config", func(t *testing.T) {
		svc := xtradata.NewConfigService(xtradata.SystemConfigDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
				return sampleSystemConfig(), nil
			},
		})
		app := systemConfigHandlerTest(t, svc)
		resp, err := doRequest(app, http.MethodDelete, "/system-configs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNoContent)
	})
}

func TestIntegrationEventHandler_Dispatch(t *testing.T) {
	t.Run("rejects invalid id", func(t *testing.T) {
		app := integrationEventHandlerTest(t, xtradata.IntegrationEventDAOMock{})
		resp, err := doRequest(app, http.MethodPost, "/integration-events/abc/dispatch", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("maps not found", func(t *testing.T) {
		app := integrationEventHandlerTest(t, xtradata.IntegrationEventDAOMock{})
		resp, err := doRequest(app, http.MethodPost, "/integration-events/1/dispatch", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})

	t.Run("maps already sent", func(t *testing.T) {
		events := xtradata.IntegrationEventDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
				event := sampleIntegrationEvent()
				event.Status = xtradata.IntegrationEventStatusSent
				return event, nil
			},
		}
		app := integrationEventHandlerTest(t, events)
		resp, err := doRequest(app, http.MethodPost, "/integration-events/1/dispatch", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusConflict)
	})

	t.Run("reports dispatch error", func(t *testing.T) {
		events := xtradata.IntegrationEventDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
				return nil, errors.New("db down")
			},
		}
		app := integrationEventHandlerTest(t, events)
		resp, err := doRequest(app, http.MethodPost, "/integration-events/1/dispatch", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("dispatches pending event", func(t *testing.T) {
		events := xtradata.IntegrationEventDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*xtradata.IntegrationEvent, error) {
				return sampleIntegrationEvent(), nil
			},
			UpdateFunc: func(_ context.Context, e *xtradata.IntegrationEvent) (*xtradata.IntegrationEvent, error) {
				return e, nil
			},
		}
		app := integrationEventHandlerTest(t, events)
		resp, err := doRequest(app, http.MethodPost, "/integration-events/1/dispatch", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})
}
