package xtradata

import (
	"encoding/json"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	IntegrationEventStatusPending = "pending"
	IntegrationEventStatusSent    = "sent"
	IntegrationEventStatusFailed  = "failed"
)

const (
	WebhookDeliveryStatusDelivered = "delivered"
	WebhookDeliveryStatusFailed    = "failed"
)

type IntegrationEvent struct {
	model.Base
	OrganizationID *uint64         `json:"organization_id"`
	Topic          string          `json:"topic"`
	Payload        json.RawMessage `gorm:"type:jsonb" json:"payload" swaggertype:"object"`
	Status         string          `json:"status"`
	Retries        int             `json:"retries"`
}

func (IntegrationEvent) TableName() string {
	return "integration_events"
}

type WebhookSubscription struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	URL            string  `json:"url"`
	Secret         string  `json:"-" audit:"redact"`
	Enabled        bool    `json:"enabled"`
}

func (WebhookSubscription) TableName() string {
	return "webhook_subscriptions"
}

type WebhookDelivery struct {
	model.Base
	EventID        uint64     `json:"event_id"`
	SubscriptionID *uint64    `json:"subscription_id"`
	OrganizationID *uint64    `json:"organization_id"`
	Status         string     `json:"status"`
	Attempt        int        `json:"attempt"`
	ResponseCode   *int       `json:"response_code"`
	Error          *string    `json:"error"`
	DeliveredAt    *time.Time `json:"delivered_at"`
}

func (WebhookDelivery) TableName() string {
	return "webhook_deliveries"
}

type IdempotencyKey struct {
	model.Base
	OrganizationID *uint64         `json:"organization_id"`
	Key            string          `json:"key"`
	Resource       string          `json:"resource"`
	StatusCode     int             `json:"status_code"`
	Response       json.RawMessage `gorm:"type:jsonb" json:"response" swaggertype:"object"`
}

func (IdempotencyKey) TableName() string {
	return "idempotency_keys"
}
