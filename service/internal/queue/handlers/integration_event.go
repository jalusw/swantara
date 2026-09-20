package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

type IntegrationEventDispatchHandler struct {
	events xtradata.EventService
}

func NewIntegrationEventDispatchHandler(events xtradata.EventService) IntegrationEventDispatchHandler {
	return IntegrationEventDispatchHandler{events: events}
}

func (h IntegrationEventDispatchHandler) Handle(ctx context.Context, t *asynq.Task) (err error) {
	var payload tasks.IntegrationEventDispatchPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	ctx = helper.ContextWithRequestID(ctx, payload.RequestID)

	defer func() {
		if err != nil {
			slog.Error("task failed", "task", t.Type(), "event_id", payload.EventID, "request_id", payload.RequestID, "error", err)
		}
	}()

	slog.Info("task started", "task", t.Type(), "event_id", payload.EventID, "request_id", payload.RequestID)

	if err := h.events.Dispatch(ctx, payload.EventID); err != nil {
		return err
	}

	slog.Info("task completed", "task", t.Type(), "event_id", payload.EventID, "request_id", payload.RequestID)
	return nil
}
