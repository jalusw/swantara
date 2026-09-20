package httpx

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func TestRecoverMiddlewareLogsPanicWithRequestID(t *testing.T) {
	prevDefault := slog.Default()
	sink := &recordSink{}
	slog.SetDefault(slog.New(sink))
	defer slog.SetDefault(prevDefault)

	app := fiber.New()
	app.Use(recover.New(recover.Config{
		EnableStackTrace:  true,
		StackTraceHandler: panicStackTraceHandler,
	}))
	app.Use(requestid.New())
	app.Use(RequestLogMiddleware())
	app.Get("/", func(c fiber.Ctx) error {
		panic("boom")
	})

	_, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	record := sink.find("panic")
	if record == nil {
		t.Fatal("expected a panic log record, got none")
	}

	if id := attrValueString(record, "request_id"); id == "" {
		t.Error("expected panic log to carry request_id")
	}

	if attrValueString(record, "stack") == "" {
		t.Error("expected panic log to carry stack trace")
	}
}

type recordSink struct {
	records []slog.Record
}

func (s *recordSink) Enabled(context.Context, slog.Level) bool { return true }

func (s *recordSink) Handle(_ context.Context, r slog.Record) error {
	s.records = append(s.records, r)
	return nil
}

func (s *recordSink) WithAttrs([]slog.Attr) slog.Handler { return s }

func (s *recordSink) WithGroup(string) slog.Handler { return s }

func (s *recordSink) find(message string) *slog.Record {
	for i := range s.records {
		if s.records[i].Message == message {
			return &s.records[i]
		}
	}
	return nil
}

func attrValueString(record *slog.Record, key string) string {
	var value string
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			value = a.Value.String()
		}
		return true
	})
	return value
}
