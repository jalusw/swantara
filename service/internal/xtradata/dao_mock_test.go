package xtradata

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestDAOMockDefaults(t *testing.T) {
	ctx := context.Background()

	var systemConfigs SystemConfigDAOMock
	page, err := systemConfigs.List(ctx, nil)
	if err != nil || page != nil {
		t.Errorf("SystemConfig List() = (%v, %v), want (nil, nil)", page, err)
	}
	if got, err := systemConfigs.Find(ctx, 1); err != nil || got != nil {
		t.Errorf("SystemConfig Find() = (%v, %v), want (nil, nil)", got, err)
	}
	config := &reference.SystemConfig{}
	if got, err := systemConfigs.Create(ctx, config); err != nil || got != config {
		t.Errorf("SystemConfig Create() = (%v, %v), want original", got, err)
	}
	if got, err := systemConfigs.Update(ctx, config); err != nil || got != config {
		t.Errorf("SystemConfig Update() = (%v, %v), want original", got, err)
	}
	if err := systemConfigs.Delete(ctx, 1); err != nil {
		t.Errorf("SystemConfig Delete() error = %v, want nil", err)
	}
	if err := systemConfigs.HardDelete(ctx, 1); err != nil {
		t.Errorf("SystemConfig HardDelete() error = %v, want nil", err)
	}
	if got, err := systemConfigs.Search(ctx, "key", "x"); err != nil || got != nil {
		t.Errorf("SystemConfig Search() = (%v, %v), want (nil, nil)", got, err)
	}

	var events IntegrationEventDAOMock
	eventPage, err := events.List(ctx, nil)
	if err != nil || eventPage != nil {
		t.Errorf("IntegrationEvent List() = (%v, %v), want (nil, nil)", eventPage, err)
	}
	if got, err := events.Find(ctx, 1); err != nil || got != nil {
		t.Errorf("IntegrationEvent Find() = (%v, %v), want (nil, nil)", got, err)
	}
	event := &IntegrationEvent{}
	if got, err := events.Create(ctx, event); err != nil || got != event {
		t.Errorf("IntegrationEvent Create() = (%v, %v), want original", got, err)
	}
	if got, err := events.Update(ctx, event); err != nil || got != event {
		t.Errorf("IntegrationEvent Update() = (%v, %v), want original", got, err)
	}
	if err := events.Delete(ctx, 1); err != nil {
		t.Errorf("IntegrationEvent Delete() error = %v, want nil", err)
	}
	if err := events.HardDelete(ctx, 1); err != nil {
		t.Errorf("IntegrationEvent HardDelete() error = %v, want nil", err)
	}
	if got, err := events.Search(ctx, "key", "x"); err != nil || got != nil {
		t.Errorf("IntegrationEvent Search() = (%v, %v), want (nil, nil)", got, err)
	}

	var keys IdempotencyKeyDAOMock
	keyPage, err := keys.List(ctx, nil)
	if err != nil || keyPage != nil {
		t.Errorf("IdempotencyKey List() = (%v, %v), want (nil, nil)", keyPage, err)
	}
	if got, err := keys.Find(ctx, 1); err != nil || got != nil {
		t.Errorf("IdempotencyKey Find() = (%v, %v), want (nil, nil)", got, err)
	}
	key := &IdempotencyKey{}
	if got, err := keys.Create(ctx, key); err != nil || got != key {
		t.Errorf("IdempotencyKey Create() = (%v, %v), want original", got, err)
	}
	if got, err := keys.Update(ctx, key); err != nil || got != key {
		t.Errorf("IdempotencyKey Update() = (%v, %v), want original", got, err)
	}
	if err := keys.Delete(ctx, 1); err != nil {
		t.Errorf("IdempotencyKey Delete() error = %v, want nil", err)
	}
	if err := keys.HardDelete(ctx, 1); err != nil {
		t.Errorf("IdempotencyKey HardDelete() error = %v, want nil", err)
	}
	if got, err := keys.Search(ctx, "key", "x"); err != nil || got != nil {
		t.Errorf("IdempotencyKey Search() = (%v, %v), want (nil, nil)", got, err)
	}
	if claimed, got, err := keys.Claim(ctx, nil, "x", "resource"); err != nil || !claimed || got != nil {
		t.Errorf("IdempotencyKey Claim() = (%v, %v, %v), want (true, nil, nil)", claimed, got, err)
	}
	if err := keys.Complete(ctx, nil, "x", 200, nil); err != nil {
		t.Errorf("IdempotencyKey Complete() error = %v, want nil", err)
	}
	if err := keys.Release(ctx, nil, "x"); err != nil {
		t.Errorf("IdempotencyKey Release() error = %v, want nil", err)
	}

	var subscriptions WebhookSubscriptionDAOMock
	subPage, err := subscriptions.List(ctx, nil)
	if err != nil || subPage != nil {
		t.Errorf("WebhookSubscription List() = (%v, %v), want (nil, nil)", subPage, err)
	}
	if got, err := subscriptions.Find(ctx, 1); err != nil || got != nil {
		t.Errorf("WebhookSubscription Find() = (%v, %v), want (nil, nil)", got, err)
	}
	subscription := &WebhookSubscription{}
	if got, err := subscriptions.Create(ctx, subscription); err != nil || got != subscription {
		t.Errorf("WebhookSubscription Create() = (%v, %v), want original", got, err)
	}
	if got, err := subscriptions.Update(ctx, subscription); err != nil || got != subscription {
		t.Errorf("WebhookSubscription Update() = (%v, %v), want original", got, err)
	}
	if err := subscriptions.Delete(ctx, 1); err != nil {
		t.Errorf("WebhookSubscription Delete() error = %v, want nil", err)
	}
	if err := subscriptions.HardDelete(ctx, 1); err != nil {
		t.Errorf("WebhookSubscription HardDelete() error = %v, want nil", err)
	}
	if got, err := subscriptions.Search(ctx, "key", "x"); err != nil || got != nil {
		t.Errorf("WebhookSubscription Search() = (%v, %v), want (nil, nil)", got, err)
	}
	if got, err := subscriptions.ListEnabledForOrg(ctx, nil); err != nil || got != nil {
		t.Errorf("WebhookSubscription ListEnabledForOrg() = (%v, %v), want (nil, nil)", got, err)
	}

	var deliveries WebhookDeliveryDAOMock
	deliveryPage, err := deliveries.List(ctx, nil)
	if err != nil || deliveryPage != nil {
		t.Errorf("WebhookDelivery List() = (%v, %v), want (nil, nil)", deliveryPage, err)
	}
	if got, err := deliveries.Find(ctx, 1); err != nil || got != nil {
		t.Errorf("WebhookDelivery Find() = (%v, %v), want (nil, nil)", got, err)
	}
	delivery := &WebhookDelivery{}
	if got, err := deliveries.Create(ctx, delivery); err != nil || got != delivery {
		t.Errorf("WebhookDelivery Create() = (%v, %v), want original", got, err)
	}
	if got, err := deliveries.Update(ctx, delivery); err != nil || got != delivery {
		t.Errorf("WebhookDelivery Update() = (%v, %v), want original", got, err)
	}
	if err := deliveries.Delete(ctx, 1); err != nil {
		t.Errorf("WebhookDelivery Delete() error = %v, want nil", err)
	}
	if err := deliveries.HardDelete(ctx, 1); err != nil {
		t.Errorf("WebhookDelivery HardDelete() error = %v, want nil", err)
	}
	if got, err := deliveries.Search(ctx, "key", "x"); err != nil || got != nil {
		t.Errorf("WebhookDelivery Search() = (%v, %v), want (nil, nil)", got, err)
	}
	if got, err := deliveries.DeliveredForEvent(ctx, 1); err != nil || got != nil {
		t.Errorf("WebhookDelivery DeliveredForEvent() = (%v, %v), want (nil, nil)", got, err)
	}
}
