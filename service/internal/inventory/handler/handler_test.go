package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

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

func writeErrorStatus(path string, fn func(c fiber.Ctx) error) int {
	app := fiber.New()
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func inventoryTestApp(t *testing.T, withTenant bool, register func(api fiber.Router, guards httpx.RouteGuards)) *fiber.App {
	t.Helper()
	app := fiber.New()
	if withTenant {
		app.Use(func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(10))
			c.Locals(model.ActorKey, uint64(5))
			return c.Next()
		})
	}
	register(app, passthroughGuards())
	return app
}

func inventoryUint64Ptr(value uint64) *uint64 {
	return &value
}

func inventoryStringPtr(value string) *string {
	return &value
}

func inventoryFloat64Ptr(value float64) *float64 {
	return &value
}
