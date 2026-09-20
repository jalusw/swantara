package httpx

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type bindRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"email"`
}

func TestBindAndValidate(t *testing.T) {
	app := fiber.New()
	app.Post("/bind", func(c fiber.Ctx) error {
		var req bindRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return c.SendStatus(http.StatusOK)
	})

	send := func(body string) int {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(body))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request error = %v", err)
		}
		_ = resp.Body.Close()
		return resp.StatusCode
	}

	if got := send(`{"name":"Ada","email":"ada@example.com"}`); got != http.StatusOK {
		t.Errorf("valid body status = %d, want 200", got)
	}
	if got := send(`{"name":"","email":"not-an-email"}`); got != http.StatusUnprocessableEntity {
		t.Errorf("invalid body status = %d, want 422", got)
	}
	if got := send(`not-json`); got != http.StatusBadRequest {
		t.Errorf("malformed body status = %d, want 400", got)
	}
}

func TestRequestHelpers(t *testing.T) {
	app := fiber.New()
	app.Get("/req", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"ip":      ClientIP(c),
			"ua":      UserAgent(c),
			"fwd_ua":  UserAgent(c),
			"device":  DeviceName(c),
			"browser": Browser(c),
			"os":      OS(c),
		})
	})

	chromeUA := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	req := httptest.NewRequest(http.MethodGet, "/req", nil)
	req.Header.Set("User-Agent", chromeUA)
	req.Header.Set("X-Forwarded-User-Agent", "CustomForwarded/1.0")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["fwd_ua"] != "CustomForwarded/1.0" {
		t.Errorf("forwarded ua = %v, want CustomForwarded/1.0", out["fwd_ua"])
	}

	req = httptest.NewRequest(http.MethodGet, "/req", nil)
	req.Header.Set("User-Agent", chromeUA)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("request error = %v", err)
	}
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["browser"] != "Chrome" {
		t.Errorf("browser = %v, want Chrome", out["browser"])
	}
}

func TestValidateRequestAndBuildFieldError(t *testing.T) {
	req := bindRequest{Name: "", Email: "bad"}
	fieldErrors := ValidateRequest(&req)
	if fieldErrors == nil {
		t.Fatal("ValidateRequest() expected field errors")
	}
	codes := map[string]string{}
	for _, fe := range *fieldErrors {
		codes[fe.Field] = fe.Code
	}
	if codes["name"] != RequiredCode {
		t.Errorf("name code = %q, want REQUIRED", codes["name"])
	}
	if codes["email"] != EmailCode {
		t.Errorf("email code = %q, want EMAIL", codes["email"])
	}

	valid := bindRequest{Name: "Ada", Email: "ada@example.com"}
	if got := ValidateRequest(&valid); got != nil {
		t.Errorf("ValidateRequest() = %+v, want nil", got)
	}
}

func TestValidateRequest_NonStruct(t *testing.T) {
	var n int
	fieldErrors := ValidateRequest(n)
	if fieldErrors == nil {
		t.Fatal("ValidateRequest() expected error for non-struct")
	}
}

func TestBuildFieldErrorMessages(t *testing.T) {
	v := newRequestValidator()
	type sample struct {
		Phone string `json:"phone" validate:"e164"`
		Code  string `json:"code" validate:"min=8,max=12,oneof=abc def"`
		Num   int    `json:"num" validate:"gt=0,gte=1,lt=5,lte=4"`
		Req   string `json:"req" validate:"required_with=Phone"`
	}
	data := sample{}
	if err := v.Struct(data); err != nil {
		errs := ValidateRequest(data)
		for _, fe := range *errs {
			if fe.Message == "" || fe.Code == "" {
				t.Errorf("field %q has empty code/message: %+v", fe.Field, fe)
			}
		}
	}
}

func TestRequestElapsed(t *testing.T) {
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(requestStartKey, time.Now().Add(-time.Millisecond))
		return c.Next()
	})
	app.Get("/elapsed", func(c fiber.Ctx) error {
		return c.SendString(requestElapsed(c))
	})

	resp := doTestRequest(app, http.MethodGet, "/elapsed")
	body := new(strings.Builder)
	_, _ = copyBody(body, resp)
	if body.String() == "" {
		t.Error("requestElapsed() = empty, want duration string")
	}
}

func TestResponseHelpers(t *testing.T) {
	app := fiber.New()
	app.Get("/accepted", func(c fiber.Ctx) error {
		return CreateAcceptedResponse(c, "", fiber.Map{"ok": true})
	})
	app.Get("/created-links", func(c fiber.Ctx) error {
		return CreateCreatedResponseWithLinks(c, "Created.", fiber.Map{"id": 1}, []Link{{Rel: "self", Href: "/x/1"}})
	})
	app.Get("/success-links", func(c fiber.Ctx) error {
		return CreateSuccessResponseWithLinks(c, "", fiber.Map{"id": 2}, []Link{{Rel: "self", Href: "/x/2"}})
	})
	app.Get("/no-content", CreateNoContentResponse)
	app.Get("/bad", func(c fiber.Ctx) error {
		return CreateBadRequestResponse(c, "Bad input.", errors.New("bad detail"))
	})
	app.Get("/conflict", func(c fiber.Ctx) error {
		return CreateConflictResponse(c, "Conflict.", errors.New("dup"))
	})
	app.Get("/forbidden", func(c fiber.Ctx) error {
		return CreateForbiddenErrorResponse(c, "No.", errors.New("denied"))
	})
	app.Get("/unprocessable", func(c fiber.Ctx) error {
		fields := []FieldError{{Field: "name", Code: "REQUIRED", Message: "name is required."}}
		return CreateUnprocessableEntityErrorResponse(c, "Invalid.", &fields)
	})
	app.Get("/service-unavailable", func(c fiber.Ctx) error {
		return CreateServiceUnavailableErrorResponse(c, "Busy.", errors.New("down"))
	})
	app.Get("/unauthorized", func(c fiber.Ctx) error {
		return CreateUnauthorizedErrorResponse(c, "Nope.", errors.New("bad token"))
	})

	checks := []struct {
		path   string
		status int
	}{
		{"/accepted", http.StatusAccepted},
		{"/created-links", http.StatusCreated},
		{"/success-links", http.StatusOK},
		{"/no-content", http.StatusNoContent},
		{"/bad", http.StatusBadRequest},
		{"/conflict", http.StatusConflict},
		{"/forbidden", http.StatusForbidden},
		{"/unprocessable", http.StatusUnprocessableEntity},
		{"/service-unavailable", http.StatusServiceUnavailable},
		{"/unauthorized", http.StatusUnauthorized},
	}
	for _, tc := range checks {
		if resp := doTestRequest(app, http.MethodGet, tc.path); resp.StatusCode != tc.status {
			t.Errorf("%s status = %d, want %d", tc.path, resp.StatusCode, tc.status)
		}
	}
}

func TestCreateErrorResponseAndSetVersion(t *testing.T) {
	SetResponseVersion("1.2.3")
	er := CreateErrorResponse("Oops.", ErrBadRequestCode, errors.New("detail"))
	if er.Success || er.Message != "Oops." || er.ErrorCode != ErrBadRequestCode || er.Detail != "detail" {
		t.Errorf("CreateErrorResponse() = %+v", er)
	}

	er = CreateErrorResponse("", "", nil)
	if er.Message != DefaultErrorMessage {
		t.Errorf("default message = %q, want %q", er.Message, DefaultErrorMessage)
	}
}

func TestLoggingRedaction(t *testing.T) {
	if got := redactErrorDetail(""); got != "" {
		t.Errorf("redactErrorDetail('') = %q, want empty", got)
	}
	got := redactErrorDetail("user alice@example.com tried")
	if strings.Contains(got, "alice@example.com") {
		t.Errorf("redactErrorDetail() = %q, want masked email", got)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	handler := &captureHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(prev)

	app := fiber.New()
	app.Use(RequestLogMiddleware())
	app.Get("/ok", func(c fiber.Ctx) error { return c.SendStatus(http.StatusOK) })
	app.Get("/warn", func(c fiber.Ctx) error {
		setResponseProblem(c, ErrNotFoundCode, "detail", nil)
		return c.SendStatus(http.StatusNotFound)
	})

	doTestRequest(app, http.MethodGet, "/ok")
	doTestRequest(app, http.MethodGet, "/warn")

	if len(handler.records) < 2 {
		t.Fatalf("expected 2 log records, got %d", len(handler.records))
	}
}

type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func TestCallerAttrs(t *testing.T) {
	app := fiber.New()
	app.Get("/attrs", func(c fiber.Ctx) error {
		c.Locals(LocalUserID, uint64(1))
		c.Locals(LocalOrganizationID, uint64(3))
		attrs := callerAttrs(c)
		return c.JSON(fiber.Map{"len": len(attrs)})
	})
	app.Get("/attrs-empty", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"len": len(callerAttrs(c))})
	})

	resp := doTestRequest(app, http.MethodGet, "/attrs")
	var out map[string]any
	if err := jsonDecode(resp, &out); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if out["len"] != float64(4) {
		t.Errorf("callerAttrs len = %v, want 4", out["len"])
	}
}

func TestRequestLogHelper(t *testing.T) {
	app := fiber.New()
	app.Get("/log", func(c fiber.Ctx) error {
		l := RequestLog(c)
		if l == nil {
			t.Fatal("RequestLog() returned nil")
		}
		return c.SendStatus(http.StatusOK)
	})
	if resp := doTestRequest(app, http.MethodGet, "/log"); resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestHealthHandler(t *testing.T) {
	queueHealth := &queueHealthMock{workerRunning: nil}
	db, _ := query.NewMockDB(t)

	cfg := &config.Config{ApplicationName: "swantara"}
	handler := NewHealthHandler(cfg, db, queueHealth)

	app := fiber.New()
	app.Get("/health", handler.Health)

	resp := doTestRequest(app, http.MethodGet, "/health")
	if resp.StatusCode != http.StatusOK {
		t.Errorf("health status = %d, want 200", resp.StatusCode)
	}

	queueHealth.workerRunning = errors.New("no worker")
	app2 := fiber.New()
	app2.Get("/health", NewHealthHandler(cfg, db, queueHealth).Health)
	if resp := doTestRequest(app2, http.MethodGet, "/health"); resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("health with worker down status = %d, want 503", resp.StatusCode)
	}
}

func TestRateLimiterStorageMemory(t *testing.T) {
	cfg := &config.Config{RateLimiterStorage: "memory"}
	storage, err := NewRateLimiterStorage(cfg)
	if err != nil || storage != nil {
		t.Errorf("NewRateLimiterStorage(memory) = (%v, %v), want (nil, nil)", storage, err)
	}
}

func TestRateLimiterStorageRedisFailure(t *testing.T) {
	cfg := &config.Config{
		RateLimiterStorage:   "redis",
		RateLimiterRedisHost: "127.0.0.1",
		RateLimiterRedisPort: 1,
	}
	storage, err := NewRateLimiterStorage(cfg)
	if err == nil {
		if storage != nil {
			_ = storage.Close()
		}
		t.Fatal("NewRateLimiterStorage(redis) expected error for unreachable redis")
	}
}

type queueHealthMock struct {
	workerRunning error
}

func (m *queueHealthMock) WorkerRunning() error {
	return m.workerRunning
}

func copyBody(dst *strings.Builder, resp *http.Response) (int64, error) {
	defer func() { _ = resp.Body.Close() }()
	buf := make([]byte, 32*1024)
	var total int64
	for {
		n, err := resp.Body.Read(buf)
		dst.Write(buf[:n])
		total += int64(n)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return total, nil
			}
			return total, err
		}
	}
}

var _ = context.Background
