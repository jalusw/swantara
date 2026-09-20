package helper

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"
)

func TestRecordOpLogsOperationWithRequestID(t *testing.T) {
	sink := captureOperations(t)

	var err error
	RecordOp(ContextWithRequestID(context.Background(), "req-op-1"), "auth.login", time.Now().Add(-5*time.Millisecond), &err)

	record := sink.find("operation")
	if record == nil {
		t.Fatal("expected an operation record, got none")
	}

	if got := attr(record, "request_id"); got != "req-op-1" {
		t.Errorf("expected request_id req-op-1, got %v", got)
	}

	if got := attr(record, "operation"); got != "auth.login" {
		t.Errorf("expected operation auth.login, got %v", got)
	}
}

func TestRecordOpLogsOperationFailure(t *testing.T) {
	sink := captureOperations(t)

	wantErr := errors.New("boom")
	err := wantErr
	RecordOp(context.Background(), "auth.login", time.Now(), &err)

	record := sink.find("operation_failed")
	if record == nil {
		t.Fatal("expected an operation_failed record, got none")
	}

	if gotErr, ok := attr(record, "error").(error); !ok || !errors.Is(gotErr, wantErr) {
		t.Errorf("expected error boom, got %v", attr(record, "error"))
	}
}

func captureOperations(t *testing.T) *operationSink {
	t.Helper()
	sink := &operationSink{}
	prevDefault := slog.Default()
	slog.SetDefault(slog.New(sink))
	t.Cleanup(func() { slog.SetDefault(prevDefault) })
	return sink
}

type operationSink struct {
	records []slog.Record
}

func (s *operationSink) Enabled(context.Context, slog.Level) bool { return true }

func (s *operationSink) Handle(_ context.Context, r slog.Record) error {
	s.records = append(s.records, r)
	return nil
}

func (s *operationSink) WithAttrs([]slog.Attr) slog.Handler { return s }

func (s *operationSink) WithGroup(string) slog.Handler { return s }

func (s *operationSink) find(message string) *slog.Record {
	for i := range s.records {
		if s.records[i].Message == message {
			return &s.records[i]
		}
	}
	return nil
}

func attr(record *slog.Record, key string) any {
	var value any
	record.Attrs(func(a slog.Attr) bool {
		if a.Key == key {
			value = a.Value.Any()
		}
		return true
	})
	return value
}
