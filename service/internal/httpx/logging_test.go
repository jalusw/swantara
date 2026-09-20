package httpx

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

func TestRequestLogMiddlewareCorrelatesRequestID(t *testing.T) {
	app := fiber.New()
	app.Use(requestid.New())
	app.Use(RequestLogMiddleware())
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString(helper.RequestID(c))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := resp.Header.Get("X-Request-ID")
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if string(body) != want {
		t.Errorf("expected request id %q to be readable from context, got %q", want, string(body))
	}
}
