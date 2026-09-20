package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
)

func TestPrettyJSONHandlerFormatsIndentedJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(newPrettyJSONHandler(&buf, slog.LevelInfo, false))
	logger.Info("user created", "user_id", 42, "email", "a@b.com")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("expected valid JSON output, got %q: %v", buf.String(), err)
	}

	if record["msg"] != "user created" {
		t.Errorf("expected msg user created, got %v", record["msg"])
	}
	if record["user_id"] != float64(42) {
		t.Errorf("expected user_id 42, got %v", record["user_id"])
	}
	if record["email"] != "a@b.com" {
		t.Errorf("expected email a@b.com, got %v", record["email"])
	}
	if !strings.Contains(buf.String(), "\n  ") {
		t.Error("expected indented JSON output, got single line")
	}
}

func TestPrettyJSONHandlerPreservesGroupAndAttrs(t *testing.T) {
	var buf bytes.Buffer
	handler := newPrettyJSONHandler(&buf, slog.LevelInfo, false).WithGroup("request").WithAttrs([]slog.Attr{slog.String("id", "req-1")})
	logger := slog.New(handler)
	logger.Info("handled")

	var record map[string]any
	if err := json.Unmarshal(buf.Bytes(), &record); err != nil {
		t.Fatalf("expected valid JSON output, got %q: %v", buf.String(), err)
	}

	group, ok := record["request"].(map[string]any)
	if !ok {
		t.Fatalf("expected request group, got %v", record["request"])
	}
	if group["id"] != "req-1" {
		t.Errorf("expected request.id req-1, got %v", group["id"])
	}
}

func TestPrettyJSONHandlerLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(newPrettyJSONHandler(&buf, slog.LevelInfo, false))
	logger.Debug("skipped")
	logger.Info("kept")

	if strings.Contains(buf.String(), "skipped") {
		t.Error("expected debug record to be filtered out")
	}
	if !strings.Contains(buf.String(), "kept") {
		t.Error("expected info record to be logged")
	}
}

func TestPrettyJSONHandlerColorsOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(newPrettyJSONHandler(&buf, slog.LevelInfo, true))
	logger.Info("colored", "count", 1, "ok", true)

	if !bytes.Contains(buf.Bytes(), []byte("\x1b[")) {
		t.Fatal("expected ANSI color codes in output, got none")
	}

	ansiColorCode := regexp.MustCompile("\x1b\\[[0-9;]*m")
	stripped := ansiColorCode.ReplaceAllString(buf.String(), "")
	var record map[string]any
	if err := json.Unmarshal([]byte(stripped), &record); err != nil {
		t.Fatalf("expected valid JSON after stripping colors, got %q: %v", stripped, err)
	}
	if record["count"] != float64(1) || record["ok"] != true {
		t.Errorf("expected colored output to keep values, got %v", record)
	}
}

func TestColorizeJSONDisabled(t *testing.T) {
	in := []byte("{\n  \"level\": \"WARN\",\n  \"ok\": true,\n  \"n\": 1,\n  \"none\": null\n}\n")
	if got := colorizeJSON(in, false); !bytes.Equal(got, in) {
		t.Error("expected unchanged output when color is disabled")
	}
}

func TestColorizeJSONEnabled(t *testing.T) {
	in := []byte("{\"level\":\"INFO\",\"msg\":\"hello\",\"n\":42,\"ok\":true,\"none\":null}")
	got := colorizeJSON(in, true)
	if !bytes.Contains(got, []byte("\x1b[")) {
		t.Fatal("expected ANSI color codes in colored output")
	}
	if !bytes.Contains(got, []byte("\x1b[34;1m\"level\"\x1b[0m")) {
		t.Error("expected keys to be colored")
	}
}

func TestSetup_SetsDefaultLogger(t *testing.T) {
	for _, tc := range []struct {
		format string
		debug  bool
	}{
		{format: "json"},
		{format: "text"},
		{format: "pretty"},
		{format: "pretty-json"},
		{format: ""},
		{format: "json", debug: true},
	} {
		t.Run(tc.format, func(t *testing.T) {
			Setup(tc.debug, "debug", Options{Format: tc.format})
			if slog.Default() == nil {
				t.Fatal("Setup() did not configure a default logger")
			}
			slog.Info("setup test", "format", tc.format)
		})
	}
}

func TestSetup_FileWriter(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	Setup(true, "warn", Options{File: path, MaxSize: 5, MaxAge: 30, MaxBackups: 1, Compress: false})
	slog.Error("file log test")

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected log file to be created: %v", err)
	}
}

func TestTerminalColor(t *testing.T) {
	if terminalColor(&bytes.Buffer{}) {
		t.Error("terminalColor(buffer) = true, want false for non-terminal writer")
	}

	file, err := os.CreateTemp(t.TempDir(), "term")
	if err != nil {
		t.Fatalf("create temp file error = %v", err)
	}
	defer func() { _ = file.Close() }()
	if terminalColor(file) {
		t.Error("terminalColor(file) = true, want false for regular file")
	}
}

func TestPrettyJSONHandler_HandleError(t *testing.T) {
	var buf bytes.Buffer
	handler := &prettyJSONHandler{writer: &buf, level: slog.LevelInfo}
	err := handler.Handle(context.Background(), slog.Record{})
	if err != nil {
		t.Fatalf("Handle() error = %v, want nil for empty record", err)
	}
	if buf.Len() == 0 {
		t.Error("Handle() wrote no output")
	}
}

func TestPrettyJSONHandler_WithGroupAndAttrs(t *testing.T) {
	var buf bytes.Buffer
	base := &prettyJSONHandler{writer: &buf, level: slog.LevelInfo}
	grouped := base.WithGroup("outer")
	attributed := grouped.WithAttrs([]slog.Attr{slog.String("k", "v")})

	if err := attributed.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !strings.Contains(buf.String(), "outer") || !strings.Contains(buf.String(), "\"k\": \"v\"") {
		t.Errorf("output = %q, want grouped attrs", buf.String())
	}
}

func TestForcedColor(t *testing.T) {
	c := forcedColor(color.FgBlue, color.Bold)
	out := c.Sprint("blue")
	if !strings.Contains(out, "\x1b[") {
		t.Error("forcedColor() did not produce ANSI output")
	}
}
