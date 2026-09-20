package db

import (
	"context"
	"errors"

	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGormLoggerLogMode(t *testing.T) {
	base := newGormSlogLogger(time.Millisecond)
	mode := base.LogMode(logger.Error)
	if mode == base {
		t.Error("LogMode() returned the same instance")
	}
	if got := mode.(*gormSlogLogger).config.LogLevel; got != logger.Error {
		t.Errorf("LogMode() level = %v, want Error", got)
	}
}

func TestGormLoggerInfoWarnErrorLevels(t *testing.T) {
	tests := []struct {
		name   string
		mode   logger.LogLevel
		log    func(l *gormSlogLogger)
		expect string
	}{
		{name: "info at info level", mode: logger.Info, log: func(l *gormSlogLogger) { l.Info(context.Background(), "info-msg") }, expect: "info-msg"},
		{name: "info suppressed at warn level", mode: logger.Warn, log: func(l *gormSlogLogger) { l.Info(context.Background(), "info-msg") }, expect: ""},
		{name: "warn at warn level", mode: logger.Warn, log: func(l *gormSlogLogger) { l.Warn(context.Background(), "warn-msg") }, expect: "warn-msg"},
		{name: "warn suppressed at error level", mode: logger.Error, log: func(l *gormSlogLogger) { l.Warn(context.Background(), "warn-msg") }, expect: ""},
		{name: "error at error level", mode: logger.Error, log: func(l *gormSlogLogger) { l.Error(context.Background(), "error-msg") }, expect: "error-msg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := captureLogs(t)
			tt.log(newGormSlogLogger(time.Millisecond).LogMode(tt.mode).(*gormSlogLogger))
			if tt.expect == "" && len(handler.records) != 0 {
				t.Errorf("expected no records, got %d", len(handler.records))
			}
			if tt.expect != "" && findRecord(handler.records, tt.expect) == nil {
				t.Errorf("expected record %q, got %+v", tt.expect, handler.records)
			}
		})
	}
}

func TestGormLoggerTraceSlowQuery(t *testing.T) {
	handler := captureLogs(t)
	times := func() (string, int64) { return "select * from users", 5 }
	l := newGormSlogLogger(time.Nanosecond).LogMode(logger.Warn)

	l.Trace(context.Background(), time.Now().Add(-time.Millisecond), times, nil)

	record := findRecord(handler.records, "slow_query")
	if record == nil {
		t.Fatalf("expected slow_query record, got %+v", handler.records)
	}
}

func TestGormLoggerTraceWithActor(t *testing.T) {
	handler := captureLogs(t)
	times := func() (string, int64) { return "select * from orders", 1 }
	l := newGormSlogLogger(time.Millisecond).LogMode(logger.Info)
	ctx := model.ContextWithActor(context.Background(), 7)

	l.Trace(ctx, time.Now(), times, nil)

	record := findRecord(handler.records, "query")
	if record == nil {
		t.Fatalf("expected query record, got none")
	}
	if got := attrValue(record, "user_id"); got != uint64(7) {
		t.Errorf("user_id = %v, want 7", got)
	}
}

func TestGormLoggerTraceIgnoresRecordNotFound(t *testing.T) {
	handler := captureLogs(t)
	times := func() (string, int64) { return "select * from users", 0 }
	l := newGormSlogLogger(time.Millisecond).LogMode(logger.Error)

	l.Trace(context.Background(), time.Now(), times, logger.ErrRecordNotFound)

	if findRecord(handler.records, "query_error") != nil {
		t.Error("expected no query_error record for record-not-found")
	}
}

func TestTrimSQL(t *testing.T) {
	if got := trimSQL("  select\n  *\n  from users "); got != "select * from users" {
		t.Errorf("trimSQL() = %q, want normalized sql", got)
	}

	long := "select " + string(make([]byte, maxSQLLogLength)) + " from users"
	trimmed := trimSQL(long)
	if len(trimmed) != maxSQLLogLength+3 {
		t.Errorf("trimSQL() length = %d, want %d", len(trimmed), maxSQLLogLength+3)
	}
	if trimmed[len(trimmed)-3:] != "..." {
		t.Errorf("trimSQL() suffix = %q, want ...", trimmed[len(trimmed)-3:])
	}
}

func TestExtractTable(t *testing.T) {
	tests := []struct {
		sql  string
		want string
	}{
		{sql: "SELECT * FROM users", want: "users"},
		{sql: "UPDATE orders SET x=1", want: "orders"},
		{sql: "INSERT INTO invoices (a) VALUES (1)", want: "invoices"},
		{sql: "SELECT * FROM public.accounts", want: "accounts"},
		{sql: `SELECT * FROM "inventory_items"`, want: "inventory_items"},
		{sql: "SELECT 1", want: ""},
	}
	for _, tt := range tests {
		if got := extractTable(tt.sql); got != tt.want {
			t.Errorf("extractTable(%q) = %q, want %q", tt.sql, got, tt.want)
		}
	}
}

func TestDBTransactioner_Run(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectCommit()

	tx := NewDBTransactioner(db)
	if err := tx.Run(context.Background(), func(tx *gorm.DB) error {
		return nil
	}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDBTransactioner_RunRollbackOnError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectRollback()

	tx := NewDBTransactioner(db)
	err := tx.Run(context.Background(), func(tx *gorm.DB) error {
		return errors.New("boom")
	})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("Run() error = %v, want boom", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewGormSlogLogger(t *testing.T) {
	l := newGormSlogLogger(2 * time.Second)
	if l.config.SlowThreshold != 2*time.Second {
		t.Errorf("SlowThreshold = %v, want 2s", l.config.SlowThreshold)
	}
	if !l.config.IgnoreRecordNotFoundError {
		t.Error("expected IgnoreRecordNotFoundError to be true")
	}
}
