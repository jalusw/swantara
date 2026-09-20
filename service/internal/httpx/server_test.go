package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"

	"github.com/jalusw/swantara/apps/service/internal/config"
)

func testAppConfig(debug bool) *config.Config {
	return &config.Config{
		ApplicationName:      "swantara-test",
		ApplicationHost:      "localhost",
		ApplicationPort:      8080,
		ApplicationVersion:   "1.0.0",
		ApplicationDebug:     debug,
		StorageMaxUploadSize: 10 * 1024 * 1024,
		RateLimiterEnabled:   true,
		RateLimiterMax:       5,
		RateLimiterDuration:  time.Minute,
		TrustedProxies:       "10.0.0.1, 10.0.0.2",
	}
}

func TestNewApp(t *testing.T) {
	app := NewApp(testAppConfig(false), nil)
	app.Get("/ok", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	if resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/ok", nil)); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Errorf("ok status = %d, want 200", resp.StatusCode)
	}
}

func TestNewApp_Debug(t *testing.T) {
	app := NewApp(testAppConfig(true), nil)
	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/cmdline", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("pprof status = %d, want 200", resp.StatusCode)
	}

	if resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/metrics", nil)); err != nil {
		t.Fatal(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Errorf("metrics status = %d, want 200", resp.StatusCode)
	}
}

func TestNewFiberConfig(t *testing.T) {
	cfg := testAppConfig(false)
	fiberConfig := newFiberConfig(cfg)
	if fiberConfig.AppName != cfg.ApplicationName || !fiberConfig.CaseSensitive {
		t.Errorf("fiberConfig = %+v", fiberConfig)
	}
	if !fiberConfig.TrustProxy || len(fiberConfig.TrustProxyConfig.Proxies) == 0 {
		t.Errorf("expected trust proxy enabled, got %+v", fiberConfig)
	}

	plain := newFiberConfig(&config.Config{ApplicationName: "x", StorageMaxUploadSize: 1})
	if plain.TrustProxy {
		t.Error("expected trust proxy disabled without trusted proxies")
	}
}

func TestRegisterDocs(t *testing.T) {
	app := fiber.New()
	RegisterDocs(app, testAppConfig(false))
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/docs/", nil))
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
}

func TestCreateCreatedResponse(t *testing.T) {
	app := fiber.New()
	app.Post("/", func(c fiber.Ctx) error {
		return CreateCreatedResponse(c, "Created.", fiber.Map{"id": 1})
	})
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

func TestRedisRateLimiterStorageOperations(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	storage := &redisRateLimiterStorage{client: client}

	if _, err := storage.Get("k"); err == nil {
		t.Error("Get() expected error against unreachable redis")
	}
	if _, err := storage.GetWithContext(context.Background(), "k"); err == nil {
		t.Error("GetWithContext() expected error against unreachable redis")
	}
	if err := storage.Set("k", []byte("v"), time.Minute); err == nil {
		t.Error("Set() expected error against unreachable redis")
	}
	if err := storage.SetWithContext(context.Background(), "k", []byte("v"), time.Minute); err == nil {
		t.Error("SetWithContext() expected error against unreachable redis")
	}
	if err := storage.Delete("k"); err == nil {
		t.Error("Delete() expected error against unreachable redis")
	}
	if err := storage.DeleteWithContext(context.Background(), "k"); err == nil {
		t.Error("DeleteWithContext() expected error against unreachable redis")
	}
	if err := storage.Reset(); err == nil {
		t.Error("Reset() expected error against unreachable redis")
	}
	if err := storage.ResetWithContext(context.Background()); err == nil {
		t.Error("ResetWithContext() expected error against unreachable redis")
	}
	if err := storage.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
