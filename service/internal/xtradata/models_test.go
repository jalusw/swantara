package xtradata

import "testing"

func TestModels_TableNames(t *testing.T) {
	if (IntegrationEvent{}).TableName() != "integration_events" {
		t.Errorf("event table = %s, want integration_events", (IntegrationEvent{}).TableName())
	}
	if (WebhookSubscription{}).TableName() != "webhook_subscriptions" {
		t.Errorf("subscription table = %s, want webhook_subscriptions", (WebhookSubscription{}).TableName())
	}
	if (WebhookDelivery{}).TableName() != "webhook_deliveries" {
		t.Errorf("delivery table = %s, want webhook_deliveries", (WebhookDelivery{}).TableName())
	}
	if (IdempotencyKey{}).TableName() != "idempotency_keys" {
		t.Errorf("key table = %s, want idempotency_keys", (IdempotencyKey{}).TableName())
	}
}
