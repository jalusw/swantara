package helper

import (
	"context"
	"log/slog"
	"time"
)

func RecordOp(ctx context.Context, name string, start time.Time, err *error) {
	attrs := []any{"operation", name, "duration", time.Since(start)}
	if requestID := RequestID(ctx); requestID != "" {
		attrs = append(attrs, "request_id", requestID)
	}

	if *err != nil {
		slog.Error("operation_failed", append(attrs, "error", *err)...)
	} else {
		slog.Debug("operation", attrs...)
	}
}
