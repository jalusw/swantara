package xtradata

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func IntegrationEventFixture(opts ...func(*IntegrationEvent) *IntegrationEvent) *IntegrationEvent {
	now := time.Now()
	evt := &IntegrationEvent{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Topic:          gofakeit.Word(),
		Status:         IntegrationEventStatusPending,
		Retries:        0,
	}
	for _, opt := range opts {
		opt(evt)
	}
	return evt
}

func WebhookSubscriptionFixture(opts ...func(*WebhookSubscription) *WebhookSubscription) *WebhookSubscription {
	now := time.Now()
	sub := &WebhookSubscription{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		URL:            gofakeit.URL(),
		Secret:         gofakeit.Password(true, true, true, true, false, 32),
		Enabled:        gofakeit.Bool(),
	}
	for _, opt := range opts {
		opt(sub)
	}
	return sub
}

func WebhookDeliveryFixture(opts ...func(*WebhookDelivery) *WebhookDelivery) *WebhookDelivery {
	now := time.Now()
	code := gofakeit.Number(200, 500)
	del := &WebhookDelivery{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		EventID:        uint64(gofakeit.Number(1, 10000)),
		SubscriptionID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Status:         WebhookDeliveryStatusDelivered,
		Attempt:        gofakeit.Number(1, 5),
		ResponseCode:   helper.Ptr(code),
		Error:          nil,
		DeliveredAt:    helper.Ptr(time.Now()),
	}
	for _, opt := range opts {
		opt(del)
	}
	return del
}

func IdempotencyKeyFixture(opts ...func(*IdempotencyKey) *IdempotencyKey) *IdempotencyKey {
	now := time.Now()
	key := &IdempotencyKey{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Key:            gofakeit.UUID(),
		Resource:       gofakeit.Word(),
		StatusCode:     gofakeit.Number(200, 500),
	}
	for _, opt := range opts {
		opt(key)
	}
	return key
}
