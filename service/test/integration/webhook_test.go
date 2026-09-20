//go:build integration

package integration

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/queue"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/xtradata"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type webhookCapture struct {
	mu     sync.Mutex
	reqs   []*http.Request
	bodies []string
}

func (c *webhookCapture) handler(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.reqs = append(c.reqs, r.Clone(r.Context()))
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		c.bodies = append(c.bodies, string(buf))
		w.WriteHeader(status)
	}
}

func (c *webhookCapture) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.reqs)
}

func (c *webhookCapture) first() (*http.Request, string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.reqs) == 0 {
		return nil, ""
	}
	return c.reqs[0], c.bodies[0]
}

func TestWebhookDeliverySignedAndAudited(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	capture := &webhookCapture{}
	server := httptest.NewServer(capture.handler(http.StatusOK))
	defer server.Close()

	const secret = "webhook-test-secret"
	subscriptionDAO := xtradata.NewWebhookSubscriptionDAO(testDB)
	subscription, err := subscriptionDAO.Create(ctx, &xtradata.WebhookSubscription{
		OrganizationID: helper.Ptr(org.ID),
		URL:            server.URL,
		Secret:         secret,
		Enabled:        true,
	})
	if err != nil {
		t.Fatalf("create webhook subscription failed: %v", err)
	}

	eventDAO := xtradata.NewIntegrationEventDAO(testDB)
	payload := json.RawMessage(`{"topic":"scheduled.job.completed","count":3}`)
	event, err := eventDAO.Create(ctx, &xtradata.IntegrationEvent{
		OrganizationID: helper.Ptr(org.ID),
		Topic:          "scheduled.job.completed",
		Payload:        payload,
		Status:         xtradata.IntegrationEventStatusPending,
	})
	if err != nil {
		t.Fatalf("create integration event failed: %v", err)
	}

	deliveryDAO := xtradata.NewWebhookDeliveryDAO(testDB)
	svc := xtradata.NewEventService(eventDAO, queue.TaskEnqueuerMock{}).
		WithDeliver(xtradata.NewWebhookDeliverer(subscriptionDAO, deliveryDAO, server.Client()).Deliver)

	if err := svc.Dispatch(ctx, event.ID); err != nil {
		t.Fatalf("dispatch failed: %v", err)
	}

	got, err := eventDAO.Find(ctx, event.ID)
	if err != nil {
		t.Fatalf("find event failed: %v", err)
	}
	if got.Status != xtradata.IntegrationEventStatusSent {
		t.Errorf("event status = %s, want sent", got.Status)
	}

	deliveriesPage, err := deliveryDAO.List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "event_id", Operator: query.Equal, Value: event.ID}},
	})
	if err != nil {
		t.Fatalf("list deliveries failed: %v", err)
	}
	if len(deliveriesPage.Items) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(deliveriesPage.Items))
	}
	delivery := deliveriesPage.Items[0]
	if delivery.Status != xtradata.WebhookDeliveryStatusDelivered || delivery.Attempt != 1 {
		t.Errorf("delivery = status:%s attempt:%d, want delivered attempt 1", delivery.Status, delivery.Attempt)
	}
	if delivery.SubscriptionID == nil || *delivery.SubscriptionID != subscription.ID {
		t.Errorf("delivery subscription_id = %v, want %d", delivery.SubscriptionID, subscription.ID)
	}
	if delivery.DeliveredAt == nil {
		t.Errorf("delivery delivered_at is nil, want set")
	}

	req, body := capture.first()
	if req == nil {
		t.Fatalf("no webhook request received")
	}
	var sentPayload map[string]any
	if err := json.Unmarshal([]byte(body), &sentPayload); err != nil {
		t.Fatalf("received body is not json: %v", err)
	}
	var wantPayload map[string]any
	_ = json.Unmarshal(payload, &wantPayload)
	if !reflect.DeepEqual(sentPayload, wantPayload) {
		t.Errorf("body = %s, want payload %s", body, payload)
	}
	wantSig := "sha256=" + webhookSignature(secret, []byte(body))
	if req.Header.Get("X-Webhook-Signature") != wantSig {
		t.Errorf("X-Webhook-Signature = %s, want %s", req.Header.Get("X-Webhook-Signature"), wantSig)
	}
	if req.Header.Get("X-Webhook-Id") != strconv.FormatUint(event.ID, 10) {
		t.Errorf("X-Webhook-Id = %s, want %d", req.Header.Get("X-Webhook-Id"), event.ID)
	}
	if req.Header.Get("X-Webhook-Timestamp") == "" {
		t.Errorf("X-Webhook-Timestamp is empty")
	}
}

func TestWebhookRedeliveryResendsOnlyFailedWebhook(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	captureA := &webhookCapture{}
	serverA := httptest.NewServer(captureA.handler(http.StatusOK))
	defer serverA.Close()

	var bCalls int
	var mu sync.Mutex
	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		bCalls++
		mu.Unlock()
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer serverB.Close()

	subscriptionDAO := xtradata.NewWebhookSubscriptionDAO(testDB)
	if _, err := subscriptionDAO.Create(ctx, &xtradata.WebhookSubscription{
		OrganizationID: helper.Ptr(org.ID), URL: serverA.URL, Secret: "secret-a", Enabled: true,
	}); err != nil {
		t.Fatalf("create webhook subscription A failed: %v", err)
	}
	if _, err := subscriptionDAO.Create(ctx, &xtradata.WebhookSubscription{
		OrganizationID: helper.Ptr(org.ID), URL: serverB.URL, Secret: "secret-b", Enabled: true,
	}); err != nil {
		t.Fatalf("create webhook subscription B failed: %v", err)
	}

	eventDAO := xtradata.NewIntegrationEventDAO(testDB)
	event, err := eventDAO.Create(ctx, &xtradata.IntegrationEvent{
		OrganizationID: helper.Ptr(org.ID),
		Topic:          "deferral.recognition.completed",
		Payload:        json.RawMessage(`{"count":1}`),
		Status:         xtradata.IntegrationEventStatusPending,
	})
	if err != nil {
		t.Fatalf("create integration event failed: %v", err)
	}

	deliveryDAO := xtradata.NewWebhookDeliveryDAO(testDB)
	deliverer := xtradata.NewWebhookDeliverer(subscriptionDAO, deliveryDAO, http.DefaultClient).
		SetBackoff(func(int) time.Duration { return 0 })
	svc := xtradata.NewEventService(eventDAO, queue.TaskEnqueuerMock{}).WithDeliver(deliverer.Deliver)

	if firstErr := svc.Dispatch(ctx, event.ID); firstErr == nil {
		t.Fatal("first dispatch should fail (webhook B down)")
	}
	if err := svc.Dispatch(ctx, event.ID); err == nil {
		t.Fatal("second dispatch should still fail (webhook B down)")
	}

	got, err := eventDAO.Find(ctx, event.ID)
	if err != nil {
		t.Fatalf("find event failed: %v", err)
	}
	if got.Status != xtradata.IntegrationEventStatusFailed {
		t.Errorf("event status = %s, want failed", got.Status)
	}
	if captureA.count() != 1 {
		t.Errorf("webhook A received %d requests, want 1 (skipped on redelivery)", captureA.count())
	}
	mu.Lock()
	bCount := bCalls
	mu.Unlock()
	if bCount != 10 {
		t.Errorf("webhook B received %d requests, want 10 (5 attempts x 2 dispatches)", bCount)
	}

	deliveriesPage, err := deliveryDAO.List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "event_id", Operator: query.Equal, Value: event.ID}},
	})
	if err != nil {
		t.Fatalf("list deliveries failed: %v", err)
	}
	delivered := 0
	failed := 0
	for _, d := range deliveriesPage.Items {
		if d.Status == xtradata.WebhookDeliveryStatusDelivered {
			delivered++
		}
		if d.Status == xtradata.WebhookDeliveryStatusFailed {
			failed++
		}
	}
	if delivered != 1 || failed != 2 {
		t.Errorf("deliveries = %d delivered, %d failed, want 1 delivered + 2 failed", delivered, failed)
	}

	req, _ := captureA.first()
	if req != nil {
		if got := req.Header.Get("X-Webhook-Id"); got != strconv.FormatUint(event.ID, 10) {
			t.Errorf("X-Webhook-Id = %s, want %d (dedupe key preserved)", got, event.ID)
		}
	}
}

func webhookSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
