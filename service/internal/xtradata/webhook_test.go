package xtradata

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestSignWebhook(t *testing.T) {
	got := signWebhook("secret", []byte(`{"count":1}`))
	mac := hmac.New(sha256.New, []byte("secret"))
	_, _ = mac.Write([]byte(`{"count":1}`))
	want := hex.EncodeToString(mac.Sum(nil))
	if got != want {
		t.Errorf("signature = %s, want %s", got, want)
	}
}

func TestWebhookDeliverer_RetriesThenRecordsDelivered(t *testing.T) {
	ctx := context.Background()
	var calls int
	var mu sync.Mutex
	var receivedHeaders http.Header
	var receivedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		receivedHeaders = r.Header.Clone()
		buf := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(buf)
		receivedBody = buf
		if calls < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var recorded *WebhookDelivery
	subscription := &WebhookSubscription{URL: server.URL, Secret: "secret"}
	subscription.ID = 1
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{
			CreateFunc: func(_ context.Context, d *WebhookDelivery) (*WebhookDelivery, error) {
				recorded = d
				return d, nil
			},
		},
		server.Client(),
	)
	deliverer.backoff = func(int) time.Duration { return 0 }

	payload := []byte(`{"count":1}`)
	event := &IntegrationEvent{Payload: payload, Topic: "t"}

	if err := deliverer.Deliver(ctx, event); err != nil {
		t.Fatalf("deliver failed: %v", err)
	}

	mu.Lock()
	callCount := calls
	hdr := receivedHeaders.Clone()
	body := append([]byte(nil), receivedBody...)
	mu.Unlock()

	if callCount != 3 {
		t.Errorf("calls = %d, want 3 (two failures then success)", callCount)
	}
	if recorded == nil {
		t.Fatalf("no delivery row recorded")
	}
	if recorded.Status != WebhookDeliveryStatusDelivered || recorded.Attempt != 3 {
		t.Errorf("delivery = status:%s attempt:%d, want delivered attempt 3", recorded.Status, recorded.Attempt)
	}
	if string(body) != string(payload) {
		t.Errorf("received body = %s, want %s", body, payload)
	}
	wantSig := "sha256=" + signWebhook("secret", payload)
	if hdr.Get("X-Webhook-Signature") != wantSig {
		t.Errorf("X-Webhook-Signature = %s, want %s", hdr.Get("X-Webhook-Signature"), wantSig)
	}
	if hdr.Get("X-Webhook-Id") == "" || hdr.Get("X-Webhook-Timestamp") == "" {
		t.Errorf("missing webhook id/timestamp headers: %v", hdr)
	}
}

func TestWebhookDeliverer_ExhaustsRetriesAndRecordsFailed(t *testing.T) {
	ctx := context.Background()
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	var recorded *WebhookDelivery
	subscription := &WebhookSubscription{URL: server.URL, Secret: "secret"}
	subscription.ID = 1
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{
			CreateFunc: func(_ context.Context, d *WebhookDelivery) (*WebhookDelivery, error) {
				recorded = d
				return d, nil
			},
		},
		&http.Client{Timeout: time.Second},
	)
	deliverer.backoff = func(int) time.Duration { return 0 }

	err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`), Topic: "t"})

	if err == nil {
		t.Fatal("deliver should fail after exhausting retries")
	}
	if calls != webhookMaxAttempts {
		t.Errorf("calls = %d, want %d", calls, webhookMaxAttempts)
	}
	if recorded == nil || recorded.Status != WebhookDeliveryStatusFailed || recorded.Attempt != webhookMaxAttempts {
		t.Errorf("delivery = %+v, want failed with attempt %d", recorded, webhookMaxAttempts)
	}
	if recorded.ResponseCode == nil || *recorded.ResponseCode != http.StatusInternalServerError {
		t.Errorf("response code = %v, want 500", recorded.ResponseCode)
	}
}

func TestWebhookDeliverer_SkipsAlreadyDeliveredSubscription(t *testing.T) {
	ctx := context.Background()
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	subscription := &WebhookSubscription{URL: server.URL, Secret: "secret"}
	subscription.ID = 1
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{
			DeliveredForEventFn: func(_ context.Context, _ uint64) ([]*WebhookDelivery, error) {
				return []*WebhookDelivery{{SubscriptionID: ptrUint64(1), Status: WebhookDeliveryStatusDelivered}}, nil
			},
		},
		server.Client(),
	)

	if err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`), Topic: "t"}); err != nil {
		t.Fatalf("deliver failed: %v", err)
	}
	if calls != 0 {
		t.Errorf("calls = %d, want 0 (already delivered subscription skipped)", calls)
	}
}

func TestWebhookDeliverer_NoSubscriptionsIsSuccess(t *testing.T) {
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return nil, nil
		}},
		WebhookDeliveryDAOMock{},
		http.DefaultClient,
	)
	if err := deliverer.Deliver(context.Background(), &IntegrationEvent{Payload: []byte(`{}`)}); err != nil {
		t.Fatalf("deliver with no subscriptions = %v, want nil", err)
	}
}

func ptrUint64(v uint64) *uint64 { return &v }
