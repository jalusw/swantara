package xtradata

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestConfigService_Read(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	svc := NewConfigService(SystemConfigDAOMock{})
	if _, err := svc.List(ctx, q); err != nil {
		t.Errorf("List = %v", err)
	}
	if _, err := svc.Find(ctx, 1); err != nil {
		t.Errorf("Find = %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Errorf("Delete = %v", err)
	}
}

func TestIdempotencyService_Delegate(t *testing.T) {
	ctx := context.Background()

	svc := NewIdempotencyService(IdempotencyKeyDAOMock{})
	if err := svc.Complete(ctx, nil, "k", 200, []byte(`{}`)); err != nil {
		t.Errorf("Complete = %v", err)
	}
	if err := svc.Release(ctx, nil, "k"); err != nil {
		t.Errorf("Release = %v", err)
	}
}

func TestXtradataMock_All(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("system config dao", func(t *testing.T) {
		bare := SystemConfigDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := bare.Update(ctx, &reference.SystemConfig{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		wired := SystemConfigDAOMock{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
				return &query.Page[reference.SystemConfig]{}, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*reference.SystemConfig, error) {
				return &reference.SystemConfig{}, nil
			},
			UpdateFunc: func(_ context.Context, e *reference.SystemConfig) (*reference.SystemConfig, error) {
				return e, nil
			},
		}
		if _, err := wired.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := wired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := wired.Update(ctx, &reference.SystemConfig{}); err != nil {
			t.Errorf("Update = %v", err)
		}
	})

	t.Run("event dao", func(t *testing.T) {
		bare := IntegrationEventDAOMock{}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := bare.Create(ctx, &IntegrationEvent{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &IntegrationEvent{}); err != nil {
			t.Errorf("Update = %v", err)
		}
	})

	t.Run("idempotency dao", func(t *testing.T) {
		bare := IdempotencyKeyDAOMock{}
		if err := bare.Complete(ctx, nil, "k", 200, nil); err != nil {
			t.Errorf("Complete = %v", err)
		}
		if err := bare.Release(ctx, nil, "k"); err != nil {
			t.Errorf("Release = %v", err)
		}
		wired := IdempotencyKeyDAOMock{
			CompleteFn: func(_ context.Context, _ *uint64, _ string, _ int, _ []byte) error { return nil },
			ReleaseFn:  func(_ context.Context, _ *uint64, _ string) error { return nil },
		}
		if err := wired.Complete(ctx, nil, "k", 200, nil); err != nil {
			t.Errorf("Complete = %v", err)
		}
		if err := wired.Release(ctx, nil, "k"); err != nil {
			t.Errorf("Release = %v", err)
		}
	})
}

func TestXtradataFixtures(t *testing.T) {
	if IntegrationEventFixture() == nil {
		t.Error("event = nil")
	}
	if WebhookSubscriptionFixture() == nil {
		t.Error("subscription = nil")
	}
	if WebhookDeliveryFixture() == nil {
		t.Error("delivery = nil")
	}
	if IdempotencyKeyFixture() == nil {
		t.Error("key = nil")
	}
}
