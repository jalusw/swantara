package db

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"gorm.io/gorm/logger"
)

func TestGormLoggerTracesQueryWithRequestID(t *testing.T) {
	handler := captureLogs(t)

	times := func() (string, int64) { return "select * from users where id = 1", 1 }
	l := newGormSlogLogger(time.Millisecond).LogMode(logger.Info)
	ctx := helper.ContextWithRequestID(context.Background(), "req-abc")

	l.Trace(ctx, time.Now(), times, nil)

	record := findRecord(handler.records, "query")
	if record == nil {
		t.Fatalf("expected a %q record, got none", "query")
	}

	if got := attrValue(record, "request_id"); got != "req-abc" {
		t.Errorf("expected request_id req-abc, got %v", got)
	}

	if got := attrValue(record, "table"); got != "users" {
		t.Errorf("expected table users, got %v", got)
	}
}

func TestGormLoggerReportsQueryError(t *testing.T) {
	handler := captureLogs(t)

	times := func() (string, int64) { return "select * from users", 0 }
	l := newGormSlogLogger(time.Millisecond).LogMode(logger.Error)

	l.Trace(context.Background(), time.Now(), times, errors.New("boom"))

	record := findRecord(handler.records, "query_error")
	if record == nil {
		t.Fatalf("expected a %q record, got none", "query_error")
	}

	if value, ok := attrValue(record, "error").(error); !ok || value.Error() != "boom" {
		t.Errorf("expected error boom, got %v", attrValue(record, "error"))
	}
}

func TestGormLoggerSilentSuppressesQueries(t *testing.T) {
	handler := captureLogs(t)

	times := func() (string, int64) { return "select * from users", 1 }
	l := newGormSlogLogger(time.Millisecond).LogMode(logger.Silent)

	l.Trace(context.Background(), time.Now(), times, nil)

	if len(handler.records) != 0 {
		t.Errorf("expected no log records, got %d", len(handler.records))
	}
}

func captureLogs(t *testing.T) *captureHandler {
	t.Helper()
	handler := &captureHandler{}
	prevDefault := slog.Default()
	slog.SetDefault(slog.New(handler))
	t.Cleanup(func() { slog.SetDefault(prevDefault) })
	return handler
}

type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *captureHandler) Handle(ctx context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}

func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }

func (h *captureHandler) WithGroup(string) slog.Handler { return h }

func findRecord(records []slog.Record, message string) *slog.Record {
	for i := range records {
		if records[i].Message == message {
			return &records[i]
		}
	}
	return nil
}

func attrValue(record *slog.Record, key string) any {
	var value any
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			value = a.Value.Any()
		}
		return true
	})
	return value
}
