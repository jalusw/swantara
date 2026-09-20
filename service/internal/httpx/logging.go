package httpx

import (
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

const (
	requestStartKey    = "request_start_time"
	responseProblemKey = "response_problem"
)

type responseProblem struct {
	ErrorCode   string
	ErrorDetail string
	FieldErrors *[]FieldError
}

func setResponseProblem(c fiber.Ctx, errorCode, errorDetail string, fieldErrors *[]FieldError) {
	c.Locals(responseProblemKey, &responseProblem{
		ErrorCode:   errorCode,
		ErrorDetail: errorDetail,
		FieldErrors: fieldErrors,
	})
}

func RequestLogMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		c.Locals(requestStartKey, start)
		c.Locals(helper.RequestIDKey, requestid.FromContext(c))
		err := c.Next()
		logRequest(c, start, err)
		return err
	}
}

func logRequest(c fiber.Ctx, start time.Time, err error) {
	attrs := []any{
		"request_id", requestid.FromContext(c),
		"method", c.Method(),
		"path", c.Path(),
		"status", c.Response().StatusCode(),
		"duration", time.Since(start),
		"bytes", c.Response().Header.ContentLength(),
	}
	attrs = append(attrs, callerAttrs(c)...)

	if p, ok := c.Locals(responseProblemKey).(*responseProblem); ok {
		attrs = append(attrs,
			"error_code", p.ErrorCode,
			"error_detail", redactErrorDetail(p.ErrorDetail),
		)
		if p.FieldErrors != nil {
			attrs = append(attrs, "field_errors", *p.FieldErrors)
		}
	}

	status := c.Response().StatusCode()
	switch {
	case err != nil || status >= 500:
		slog.Error("http_request", append(attrs, "error", err)...)
	case status >= 400:
		slog.Warn("http_request", attrs...)
	default:
		slog.Info("http_request", attrs...)
	}
}

func RequestLog(c fiber.Ctx) *slog.Logger {
	return slog.With("request_id", requestid.FromContext(c)).With(callerAttrs(c)...)
}

func callerAttrs(c fiber.Ctx) []any {
	attrs := make([]any, 0, 6)
	if id, ok := CallerID(c); ok {
		attrs = append(attrs, "user_id", id)
	}

	if id, ok := CallerOrganizationID(c); ok {
		attrs = append(attrs, "organization_id", id)
	}
	return attrs
}

func redactErrorDetail(detail string) string {
	if detail == "" {
		return detail
	}

	fields := strings.Fields(detail)
	for i, field := range fields {
		if strings.Contains(field, "@") {
			fields[i] = helper.MaskEmail(field)
		}
	}

	return strings.Join(fields, " ")
}

func RateLimitReached(c fiber.Ctx) error {
	slog.Warn("rate_limited", "request_id", requestid.FromContext(c), "ip", c.IP())
	return c.SendStatus(fiber.StatusTooManyRequests)
}
