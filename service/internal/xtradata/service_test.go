package xtradata

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestConfigService_Create_RejectsDuplicateKey(t *testing.T) {
	ctx := context.Background()
	svc := NewConfigService(SystemConfigDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*reference.SystemConfig, error) {
			return &reference.SystemConfig{Key: "currency"}, nil
		},
	})

	_, err := svc.Create(ctx, &reference.SystemConfig{Key: "currency", Value: []byte(`"IDR"`)})
	if !errors.Is(err, ErrConfigKeyExists) {
		t.Fatalf("err = %v, want ErrConfigKeyExists", err)
	}
}

func TestEventService_Enqueue_PersistsPendingAndEnqueues(t *testing.T) {
	ctx := context.Background()
	created := &IntegrationEvent{Topic: "order.created", Status: IntegrationEventStatusPending}
	var enqueued *asynq.Task
	svc := NewEventService(
		IntegrationEventDAOMock{CreateFunc: func(_ context.Context, e *IntegrationEvent) (*IntegrationEvent, error) {
			created = e
			created.ID = 7
			return created, nil
		}},
		queue.TaskEnqueuerMock{EnqueueFunc: func(task *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
			enqueued = task
			return nil, nil
		}},
	)

	event, err := svc.Enqueue(ctx, nil, "order.created", map[string]any{"id": 1})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if event.ID != 7 || event.Status != IntegrationEventStatusPending {
		t.Errorf("event = %+v, want pending with id 7", event)
	}
	if enqueued == nil {
		t.Errorf("expected a dispatch task to be enqueued")
	}
}

func TestEventService_Dispatch_MarksSentOnSuccess(t *testing.T) {
	ctx := context.Background()
	svc := NewEventService(
		IntegrationEventDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*IntegrationEvent, error) {
				return &IntegrationEvent{Status: IntegrationEventStatusPending}, nil
			},
			UpdateFunc: func(_ context.Context, e *IntegrationEvent) (*IntegrationEvent, error) { return e, nil },
		},
		queue.TaskEnqueuerMock{},
	)

	if err := svc.Dispatch(ctx, 1); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}
}

func TestEventService_Dispatch_RetriesMarkFailedWithoutDuplicates(t *testing.T) {
	ctx := context.Background()
	deliverErr := errors.New("delivery down")
	seenStatuses := []string{}
	svc := NewEventService(
		IntegrationEventDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*IntegrationEvent, error) {
				return &IntegrationEvent{Status: IntegrationEventStatusFailed}, nil
			},
			UpdateFunc: func(_ context.Context, e *IntegrationEvent) (*IntegrationEvent, error) {
				seenStatuses = append(seenStatuses, e.Status)
				return e, nil
			},
		},
		queue.TaskEnqueuerMock{},
	).WithDeliver(func(context.Context, *IntegrationEvent) error { return deliverErr })

	firstErr := svc.Dispatch(ctx, 1)
	secondErr := svc.Dispatch(ctx, 1)

	if !errors.Is(firstErr, deliverErr) || !errors.Is(secondErr, deliverErr) {
		t.Errorf("firstErr = %v, secondErr = %v, want deliverErr on both retries", firstErr, secondErr)
	}
	if len(seenStatuses) != 2 || seenStatuses[0] != IntegrationEventStatusFailed || seenStatuses[1] != IntegrationEventStatusFailed {
		t.Errorf("statuses after retries = %v, want failed,failed", seenStatuses)
	}
}

func TestEventService_Dispatch_ReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewEventService(
		IntegrationEventDAOMock{FindFunc: func(_ context.Context, _ uint64) (*IntegrationEvent, error) {
			return nil, nil
		}},
		queue.TaskEnqueuerMock{},
	)

	if err := svc.Dispatch(ctx, 99); !errors.Is(err, ErrEventNotFound) {
		t.Errorf("err = %v, want ErrEventNotFound", err)
	}
}

func TestIdempotencyService_ClaimNewKey_ReturnsNilEntry(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(10)
	svc := NewIdempotencyService(IdempotencyKeyDAOMock{})

	entry, err := svc.Claim(ctx, &organizationID, "abc", "sale-order-confirm")
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	if entry != nil {
		t.Errorf("entry = %+v, want nil for a fresh claim", entry)
	}
}

func TestIdempotencyService_ClaimCompletedKey_ReplaysEntry(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(10)
	svc := NewIdempotencyService(IdempotencyKeyDAOMock{
		ClaimFn: func(_ context.Context, _ *uint64, _, _ string) (bool, *IdempotencyKey, error) {
			return false, &IdempotencyKey{Resource: "sale-order-confirm", StatusCode: 200, Response: []byte(`{"ok":true}`)}, nil
		},
	})

	entry, err := svc.Claim(ctx, &organizationID, "abc", "sale-order-confirm")
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	if entry == nil || entry.StatusCode != 200 || string(entry.Response) != `{"ok":true}` {
		t.Errorf("entry = %+v, want stored response", entry)
	}
}

func TestIdempotencyService_ClaimInFlightKey_ReturnsInFlight(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(10)
	svc := NewIdempotencyService(IdempotencyKeyDAOMock{
		ClaimFn: func(_ context.Context, _ *uint64, _, _ string) (bool, *IdempotencyKey, error) {
			return false, &IdempotencyKey{Resource: "sale-order-confirm"}, nil
		},
	})

	if _, err := svc.Claim(ctx, &organizationID, "abc", "sale-order-confirm"); !errors.Is(err, httpx.ErrIdempotencyKeyInFlight) {
		t.Errorf("err = %v, want ErrIdempotencyKeyInFlight", err)
	}
}

func TestIdempotencyService_ClaimReusedAcrossResources_ReturnsReused(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(10)
	svc := NewIdempotencyService(IdempotencyKeyDAOMock{
		ClaimFn: func(_ context.Context, _ *uint64, _, _ string) (bool, *IdempotencyKey, error) {
			return false, &IdempotencyKey{Resource: "invoice-create", StatusCode: 200, Response: []byte(`{"id":1}`)}, nil
		},
	})

	if _, err := svc.Claim(ctx, &organizationID, "abc", "payment-create"); !errors.Is(err, httpx.ErrIdempotencyKeyReused) {
		t.Errorf("err = %v, want ErrIdempotencyKeyReused", err)
	}
}
