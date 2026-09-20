package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/audit"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func auditTestApp(t *testing.T, logs dao.CRUDMock[audit.Log]) *fiber.App {
	t.Helper()
	app := fiber.New()
	NewAuditLogHandler(logs).Register(app, httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error { return c.Next() },
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error { return c.Next() }
		},
	})
	return app
}

func doAuditRequest(app *fiber.App, method, path string) *http.Response {
	req := httptest.NewRequest(method, path, strings.NewReader(""))
	resp, err := app.Test(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func sampleAuditLog() *audit.Log {
	return &audit.Log{
		Base:        model.Base{ID: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		EntityTable: "orders",
		RecordID:    7,
		Action:      "update",
		ChangedBy:   3,
		ChangedAt:   time.Now(),
	}
}

func TestAuditLogHandler_List_Get(t *testing.T) {
	t.Run("lists logs", func(t *testing.T) {
		mock := dao.CRUDMock[audit.Log]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[audit.Log], error) {
				return &query.Page[audit.Log]{Items: []*audit.Log{sampleAuditLog()}, Count: 1}, nil
			},
		}
		resp := doAuditRequest(auditTestApp(t, mock), http.MethodGet, "/audit-logs/")
		defer func() { _ = resp.Body.Close() }()
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("rejects invalid query and reports failure", func(t *testing.T) {
		resp := doAuditRequest(auditTestApp(t, dao.CRUDMock[audit.Log]{}), http.MethodGet, "/audit-logs/?size=abc")
		defer func() { _ = resp.Body.Close() }()
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		mock := dao.CRUDMock[audit.Log]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[audit.Log], error) {
				return nil, errors.New("db down")
			},
		}
		resp = doAuditRequest(auditTestApp(t, mock), http.MethodGet, "/audit-logs/")
		defer func() { _ = resp.Body.Close() }()
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("rejects invalid id", func(t *testing.T) {
		resp := doAuditRequest(auditTestApp(t, dao.CRUDMock[audit.Log]{}), http.MethodGet, "/audit-logs/abc")
		defer func() { _ = resp.Body.Close() }()
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("returns log", func(t *testing.T) {
		mock := dao.CRUDMock[audit.Log]{
			FindFunc: func(_ context.Context, _ uint64) (*audit.Log, error) { return sampleAuditLog(), nil },
		}
		resp := doAuditRequest(auditTestApp(t, mock), http.MethodGet, "/audit-logs/1")
		defer func() { _ = resp.Body.Close() }()
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("returns not found and reports failure", func(t *testing.T) {
		resp := doAuditRequest(auditTestApp(t, dao.CRUDMock[audit.Log]{}), http.MethodGet, "/audit-logs/1")
		defer func() { _ = resp.Body.Close() }()
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		mock := dao.CRUDMock[audit.Log]{
			FindFunc: func(_ context.Context, _ uint64) (*audit.Log, error) { return nil, errors.New("db down") },
		}
		resp = doAuditRequest(auditTestApp(t, mock), http.MethodGet, "/audit-logs/1")
		defer func() { _ = resp.Body.Close() }()
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}
