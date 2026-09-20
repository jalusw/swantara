package xtradata

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
)

const (
	webhookMaxAttempts = 5
	webhookBackoffBase = 200 * time.Millisecond
)

type WebhookDeliverer struct {
	subscriptions WebhookSubscriptionDAO
	deliveries    WebhookDeliveryDAO
	client        *http.Client
	now           func() time.Time
	backoff       func(attempt int) time.Duration
}

func NewWebhookDeliverer(subscriptions WebhookSubscriptionDAO, deliveries WebhookDeliveryDAO, client *http.Client) WebhookDeliverer {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return WebhookDeliverer{
		subscriptions: subscriptions,
		deliveries:    deliveries,
		client:        client,
		now:           time.Now,
		backoff:       func(attempt int) time.Duration { return webhookBackoffBase << (attempt - 1) },
	}
}

func (d WebhookDeliverer) SetBackoff(backoff func(attempt int) time.Duration) WebhookDeliverer {
	d.backoff = backoff
	return d
}

func (d WebhookDeliverer) Deliver(ctx context.Context, event *IntegrationEvent) error {
	subscriptions, err := d.subscriptions.ListEnabledForOrg(ctx, event.OrganizationID)
	if err != nil {
		return err
	}
	if len(subscriptions) == 0 {
		return nil
	}

	delivered, err := d.deliveries.DeliveredForEvent(ctx, event.ID)
	if err != nil {
		return err
	}
	alreadyDelivered := make(map[uint64]struct{}, len(delivered))
	for _, row := range delivered {
		if row.SubscriptionID != nil {
			alreadyDelivered[*row.SubscriptionID] = struct{}{}
		}
	}

	var firstErr error
	for _, subscription := range subscriptions {
		if _, ok := alreadyDelivered[subscription.ID]; ok {
			continue
		}
		if err := d.deliverTo(ctx, event, subscription); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (d WebhookDeliverer) deliverTo(ctx context.Context, event *IntegrationEvent, subscription *WebhookSubscription) error {
	payload := event.Payload
	timestamp := d.now().Unix()
	var lastErr error
	var lastCode *int
	for attempt := 1; attempt <= webhookMaxAttempts; attempt++ {
		if attempt > 1 {
			if err := sleep(ctx, d.backoff(attempt)); err != nil {
				lastErr = err
				break
			}
		}
		statusCode, err := d.post(ctx, subscription, payload, event.ID, timestamp)
		if err == nil {
			d.record(ctx, event, subscription, &statusCode, nil, attempt)
			return nil
		}
		lastErr = err
		lastCode = &statusCode
	}
	d.record(ctx, event, subscription, lastCode, lastErr, webhookMaxAttempts)
	return lastErr
}

func (d WebhookDeliverer) post(ctx context.Context, subscription *WebhookSubscription, payload []byte, eventID uint64, timestamp int64) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, subscription.URL, bytes.NewReader(payload))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Id", strconv.FormatUint(eventID, 10))
	req.Header.Set("X-Webhook-Timestamp", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-Webhook-Signature", "sha256="+signWebhook(subscription.Secret, payload))

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.StatusCode, fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return resp.StatusCode, nil
}

func (d WebhookDeliverer) record(ctx context.Context, event *IntegrationEvent, subscription *WebhookSubscription, responseCode *int, deliveryErr error, attempt int) {
	delivery := &WebhookDelivery{
		EventID:        event.ID,
		SubscriptionID: helper.Ptr(subscription.ID),
		OrganizationID: event.OrganizationID,
		Status:         WebhookDeliveryStatusDelivered,
		Attempt:        attempt,
		ResponseCode:   responseCode,
	}
	if deliveryErr != nil {
		delivery.Status = WebhookDeliveryStatusFailed
		msg := deliveryErr.Error()
		delivery.Error = &msg
	} else {
		now := d.now()
		delivery.DeliveredAt = &now
	}
	if _, err := d.deliveries.Create(ctx, delivery); err != nil {
		slog.Error("failed to record webhook delivery", "event_id", event.ID, "subscription_id", subscription.ID, "error", err)
	}
}

func signWebhook(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write(payload); err != nil {
		slog.Error("failed to sign webhook payload", "error", err)
		return ""
	}
	return hex.EncodeToString(mac.Sum(nil))
}

func sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
