package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lmittmann/tint"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Options struct {
	File       string
	Format     string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

func Setup(debug bool, level string, opts Options) {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	var writer io.Writer
	if opts.File != "" {
		writer = &lumberjack.Logger{
			Filename:   opts.File,
			MaxSize:    opts.MaxSize,
			MaxAge:     opts.MaxAge,
			MaxBackups: opts.MaxBackups,
			Compress:   opts.Compress,
		}
	} else {
		writer = os.Stdout
	}

	format := opts.Format
	if format == "" {
		if debug {
			format = "text"
		} else {
			format = "json"
		}
	}

	var handler slog.Handler
	switch strings.ToLower(format) {
	case "pretty":
		handler = tint.NewTextHandler(writer, &tint.Options{
			Level:      l,
			NoColor:    !terminalColor(writer),
			TimeFormat: "15:04:05.000",
		})
	case "pretty-json":
		handler = newPrettyJSONHandler(writer, l, terminalColor(writer) && os.Getenv("NO_COLOR") == "")
	case "text":
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{Level: l})
	default:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: l})
	}

	slog.SetDefault(slog.New(handler))
}

func Fatal(msg string, err error) {
	slog.Error(msg, "error", err)
	os.Exit(1)
}

func Fatalf(format string, args ...any) {
	slog.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func terminalColor(out io.Writer) bool {
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type prettyJSONHandler struct {
	writer io.Writer
	level  slog.Level
	color  bool
	attrs  []slog.Attr
	groups []string
}

func newPrettyJSONHandler(writer io.Writer, level slog.Level, colorEnabled bool) *prettyJSONHandler {
	return &prettyJSONHandler{writer: writer, level: level, color: colorEnabled}
}

func (h *prettyJSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *prettyJSONHandler) Handle(ctx context.Context, r slog.Record) error {
	var buf bytes.Buffer
	var inner slog.Handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: h.level})
	for _, group := range h.groups {
		inner = inner.WithGroup(group)
	}
	inner = inner.WithAttrs(h.attrs)
	if err := inner.Handle(ctx, r); err != nil {
		return err
	}

	var out bytes.Buffer
	if err := json.Indent(&out, buf.Bytes(), "", "  "); err != nil {
		_, err := io.WriteString(h.writer, buf.String())
		return err
	}
	_, err := h.writer.Write(colorizeJSON(out.Bytes(), h.color))
	return err
}

func (h *prettyJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	merged := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	merged = append(merged, h.attrs...)
	merged = append(merged, attrs...)
	return &prettyJSONHandler{writer: h.writer, level: h.level, color: h.color, attrs: merged, groups: h.groups}
}

func (h *prettyJSONHandler) WithGroup(name string) slog.Handler {
	groups := make([]string, 0, len(h.groups)+1)
	groups = append(groups, h.groups...)
	groups = append(groups, name)
	return &prettyJSONHandler{writer: h.writer, level: h.level, color: h.color, attrs: h.attrs, groups: groups}
}

var (
	prettyJSONKeyColor    = forcedColor(color.FgBlue, color.Bold)
	prettyJSONStringColor = forcedColor(color.FgGreen, color.Bold)
	prettyJSONNumberColor = forcedColor(color.FgCyan, color.Bold)
	prettyJSONBoolColor   = forcedColor(color.FgYellow, color.Bold)
	prettyJSONNullColor   = forcedColor(color.FgBlack, color.Bold)
)

func forcedColor(attrs ...color.Attribute) *color.Color {
	c := color.New(attrs...)
	c.EnableColor()
	return c
}

func colorizeJSON(src []byte, enabled bool) []byte {
	if !enabled {
		return src
	}
	var out bytes.Buffer
	for i := 0; i < len(src); {
		switch {
		case src[i] == '"':
			start := i
			i++
			for i < len(src) && src[i] != '"' {
				if src[i] == '\\' {
					i += 2
				} else {
					i++
				}
			}
			i++
			isKey := false
			for j := i; j < len(src); j++ {
				if src[j] == ' ' || src[j] == '\t' {
					continue
				}
				isKey = src[j] == ':'
				break
			}
			if isKey {
				out.WriteString(prettyJSONKeyColor.Sprint(string(src[start:i])))
			} else {
				out.WriteString(prettyJSONStringColor.Sprint(string(src[start:i])))
			}
		case src[i] == '-' || src[i] >= '0' && src[i] <= '9':
			start := i
			for i < len(src) && (src[i] == '-' || src[i] == '+' || src[i] == '.' || src[i] == 'e' || src[i] == 'E' || src[i] >= '0' && src[i] <= '9') {
				i++
			}
			out.WriteString(prettyJSONNumberColor.Sprint(string(src[start:i])))
		case src[i] == 't' && bytes.HasPrefix(src[i:], []byte("true")):
			out.WriteString(prettyJSONBoolColor.Sprint("true"))
			i += 4
		case src[i] == 'f' && bytes.HasPrefix(src[i:], []byte("false")):
			out.WriteString(prettyJSONBoolColor.Sprint("false"))
			i += 5
		case src[i] == 'n' && bytes.HasPrefix(src[i:], []byte("null")):
			out.WriteString(prettyJSONNullColor.Sprint("null"))
			i += 4
		default:
			out.WriteByte(src[i])
			i++
		}
	}
	return out.Bytes()
}
