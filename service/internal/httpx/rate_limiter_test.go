package httpx

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

func TestRateLimiterRejectsRequestsBeyondLimit(t *testing.T) {
	app := fiber.New()
	app.Use(limiter.New(limiter.Config{
		Max:          2,
		Expiration:   time.Minute,
		LimitReached: RateLimitReached,
	}))
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})

	statuses := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		request := httptest.NewRequest(fiber.MethodGet, "/ping", nil)
		response, err := app.Test(request)
		if err != nil {
			t.Fatalf("request %d failed: %v", i, err)
		}
		statuses = append(statuses, response.StatusCode)
		_ = response.Body.Close()
	}

	if statuses[0] != fiber.StatusOK || statuses[1] != fiber.StatusOK {
		t.Fatalf("expected first two requests to pass, got %v", statuses)
	}
	if statuses[2] != fiber.StatusTooManyRequests {
		t.Fatalf("expected third request to be rate limited, got %v", statuses)
	}
}
