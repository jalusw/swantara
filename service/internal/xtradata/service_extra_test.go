package xtradata

import (
	"context"
	"errors"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestConfigService_Create_PropagatesSearchError(t *testing.T) {
	ctx := context.Background()
	searchErr := errors.New("db down")
	svc := NewConfigService(SystemConfigDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*reference.SystemConfig, error) {
			return nil, searchErr
		},
	})

	_, err := svc.Create(ctx, &reference.SystemConfig{Key: "currency"})
	if helper.AssertError(t, err, true, searchErr) {
		return
	}
}

func TestConfigService_Create_PersistsNewConfig(t *testing.T) {
	ctx := context.Background()
	created := &reference.SystemConfig{Key: "currency"}
	svc := NewConfigService(SystemConfigDAOMock{
		SearchFunc: func(_ context.Context, _ string, _ any) (*reference.SystemConfig, error) {
			return nil, nil
		},
		CreateFunc: func(_ context.Context, config *reference.SystemConfig) (*reference.SystemConfig, error) {
			created = config
			return config, nil
		},
	})

	config := &reference.SystemConfig{Key: "currency", Value: []byte(`"IDR"`)}
	saved, err := svc.Create(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved != config || created != config {
		t.Errorf("saved = %+v, want original config", saved)
	}
}

func TestConfigService_Update_PersistsConfig(t *testing.T) {
	ctx := context.Background()
	config := &reference.SystemConfig{Key: "currency", Value: []byte(`"USD"`)}
	svc := NewConfigService(SystemConfigDAOMock{})

	saved, err := svc.Update(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if saved != config {
		t.Errorf("saved = %+v, want original config", saved)
	}
}

func TestEventService_Enqueue_PropagatesMarshalError(t *testing.T) {
	svc := NewEventService(IntegrationEventDAOMock{}, queue.TaskEnqueuerMock{})

	_, err := svc.Enqueue(context.Background(), nil, "topic", make(chan int))
	if helper.AssertError(t, err, true, nil) {
		return
	}
}

func TestEventService_Enqueue_PropagatesCreateError(t *testing.T) {
	ctx := context.Background()
	createErr := errors.New("db down")
	svc := NewEventService(
		IntegrationEventDAOMock{CreateFunc: func(_ context.Context, _ *IntegrationEvent) (*IntegrationEvent, error) {
			return nil, createErr
		}},
		queue.TaskEnqueuerMock{},
	)

	_, err := svc.Enqueue(ctx, nil, "topic", map[string]any{"id": 1})
	if helper.AssertError(t, err, true, createErr) {
		return
	}
}

func TestEventService_Enqueue_PropagatesEnqueueError(t *testing.T) {
	ctx := context.Background()
	enqueueErr := errors.New("queue down")
	svc := NewEventService(
		IntegrationEventDAOMock{CreateFunc: func(_ context.Context, e *IntegrationEvent) (*IntegrationEvent, error) {
			e.ID = 7
			return e, nil
		}},
		queue.TaskEnqueuerMock{EnqueueFunc: func(_ *asynq.Task, _ ...asynq.Option) (*asynq.TaskInfo, error) {
			return nil, enqueueErr
		}},
	)

	event, err := svc.Enqueue(ctx, nil, "topic", map[string]any{"id": 1})
	if !errors.Is(err, enqueueErr) {
		t.Fatalf("err = %v, want enqueueErr", err)
	}
	if event == nil || event.ID != 7 {
		t.Errorf("event = %+v, want persisted event", event)
	}
}

func TestEventService_Dispatch_PropagatesFindError(t *testing.T) {
	ctx := context.Background()
	findErr := errors.New("db down")
	svc := NewEventService(
		IntegrationEventDAOMock{FindFunc: func(_ context.Context, _ uint64) (*IntegrationEvent, error) {
			return nil, findErr
		}},
		queue.TaskEnqueuerMock{},
	)

	err := svc.Dispatch(ctx, 1)
	if helper.AssertError(t, err, true, findErr) {
		return
	}
}

func TestEventService_Dispatch_RejectsAlreadySent(t *testing.T) {
	ctx := context.Background()
	svc := NewEventService(
		IntegrationEventDAOMock{FindFunc: func(_ context.Context, _ uint64) (*IntegrationEvent, error) {
			return &IntegrationEvent{Status: IntegrationEventStatusSent}, nil
		}},
		queue.TaskEnqueuerMock{},
	)

	err := svc.Dispatch(ctx, 1)
	if helper.AssertError(t, err, true, ErrEventAlreadySent) {
		return
	}
}

func TestEventService_Dispatch_PropagatesUpdateErrorOnFailure(t *testing.T) {
	ctx := context.Background()
	updateErr := errors.New("db down")
	svc := NewEventService(
		IntegrationEventDAOMock{
			FindFunc: func(_ context.Context, _ uint64) (*IntegrationEvent, error) {
				return &IntegrationEvent{Status: IntegrationEventStatusPending}, nil
			},
			UpdateFunc: func(_ context.Context, _ *IntegrationEvent) (*IntegrationEvent, error) {
				return nil, updateErr
			},
		},
		queue.TaskEnqueuerMock{},
	).WithDeliver(func(context.Context, *IntegrationEvent) error { return errors.New("delivery down") })

	err := svc.Dispatch(ctx, 1)
	if helper.AssertError(t, err, true, updateErr) {
		return
	}
}

func TestIdempotencyService_Claim_PropagatesError(t *testing.T) {
	ctx := context.Background()
	claimErr := errors.New("db down")
	svc := NewIdempotencyService(IdempotencyKeyDAOMock{
		ClaimFn: func(_ context.Context, _ *uint64, _, _ string) (bool, *IdempotencyKey, error) {
			return false, nil, claimErr
		},
	})

	_, err := svc.Claim(ctx, nil, "abc", "sale-order-confirm")
	if helper.AssertError(t, err, true, claimErr) {
		return
	}
}
