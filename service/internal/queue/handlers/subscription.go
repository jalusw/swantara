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
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
)

type SubscriptionBillingHandler struct {
	subscriptions subscription.SubscriptionService
	deferrals     accounting.DeferralService
	jobRuns       jobs.Service
	events        xtradata.EventService
}

func NewSubscriptionBillingHandler(subscriptions subscription.SubscriptionService, deferrals accounting.DeferralService, jobRuns jobs.Service, events xtradata.EventService) SubscriptionBillingHandler {
	return SubscriptionBillingHandler{subscriptions: subscriptions, deferrals: deferrals, jobRuns: jobRuns, events: events}
}

func (h SubscriptionBillingHandler) Handle(ctx context.Context, t *asynq.Task) (err error) {
	var payload tasks.SubscriptionBillingPayload

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

	run, err := h.jobRuns.Start(ctx, jobs.TypeSubscriptionBilling)
	if err != nil {
		return fmt.Errorf("failed to start job run checkpoint: %w", err)
	}
	defer func() {
		if err != nil {
			_ = h.jobRuns.Fail(ctx, run, err.Error())
		}
	}()

	asOf := time.Now().UTC()
	billed, err := h.subscriptions.BillDue(ctx, asOf)
	if err != nil {
		return fmt.Errorf("failed to bill due subscriptions: %w", err)
	}
	recognized, err := h.deferrals.RecognizeDue(ctx, nil, asOf)
	if err != nil {
		return fmt.Errorf("failed to recognize revenue: %w", err)
	}

	if err := h.jobRuns.Complete(ctx, run, billed+recognized); err != nil {
		return fmt.Errorf("failed to complete job run checkpoint: %w", err)
	}

	publishRunEvent(ctx, h.events, "subscription.billing.completed", map[string]any{
		"billed":     billed,
		"recognized": recognized,
		"as_of":      asOf,
	})

	slog.Info("task completed", "task", t.Type(), "request_id", payload.RequestID, "billed", billed, "recognized", recognized)

	return nil
}
