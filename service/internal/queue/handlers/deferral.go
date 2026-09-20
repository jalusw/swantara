package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/jobs"
	"github.com/jalusw/swantara/apps/service/internal/queue/tasks"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

type DeferralRecognitionHandler struct {
	deferrals accounting.DeferralService
	jobRuns   jobs.Service
	events    xtradata.EventService
}

func NewDeferralRecognitionHandler(deferrals accounting.DeferralService, jobRuns jobs.Service, events xtradata.EventService) DeferralRecognitionHandler {
	return DeferralRecognitionHandler{deferrals: deferrals, jobRuns: jobRuns, events: events}
}

func (h DeferralRecognitionHandler) Handle(ctx context.Context, t *asynq.Task) (err error) {
	var payload tasks.DeferralRecognitionPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	ctx = helper.ContextWithRequestID(ctx, payload.RequestID)

	defer func() {
		if err != nil {
			slog.Error("task failed", "task", t.Type(), "request_id", payload.RequestID, "error", err)
		}
	}()

	slog.Info("task started", "task", t.Type(), "request_id", payload.RequestID)

	run, err := h.jobRuns.Start(ctx, jobs.TypeDeferralRecognition)
	if err != nil {
		return fmt.Errorf("failed to start job run checkpoint: %w", err)
	}
	defer func() {
		if err != nil {
			_ = h.jobRuns.Fail(ctx, run, err.Error())
		}
	}()

	recognized, err := h.deferrals.RecognizeDue(ctx, nil, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to recognize due deferrals: %w", err)
	}

	if err := h.jobRuns.Complete(ctx, run, recognized); err != nil {
		return fmt.Errorf("failed to complete job run checkpoint: %w", err)
	}

	publishRunEvent(ctx, h.events, "deferral.recognition.completed", map[string]any{
		"recognized": recognized,
		"as_of":      time.Now().UTC(),
	})

	slog.Info("task completed", "task", t.Type(), "request_id", payload.RequestID, "recognized", recognized)

	return nil
}
