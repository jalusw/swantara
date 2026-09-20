package handlers

import (
	"context"
	"log/slog"

	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

func publishRunEvent(ctx context.Context, events xtradata.EventService, topic string, payload any) {
	if _, err := events.Enqueue(ctx, nil, topic, payload); err != nil {
		slog.Warn("failed to enqueue integration event", "topic", topic, "error", err)
	}
}
