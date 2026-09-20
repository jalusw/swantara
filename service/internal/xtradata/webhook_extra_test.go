package xtradata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestWebhookDeliverer_NewWithNilClient_AppliesDefaultTimeout(t *testing.T) {
	deliverer := NewWebhookDeliverer(WebhookSubscriptionDAOMock{}, WebhookDeliveryDAOMock{}, nil)
	if deliverer.client == nil || deliverer.client.Timeout != 10*time.Second {
		t.Errorf("client = %+v, want default timeout client", deliverer.client)
	}
}

func TestWebhookDeliverer_SetBackoff_OverridesBackoff(t *testing.T) {
	deliverer := NewWebhookDeliverer(WebhookSubscriptionDAOMock{}, WebhookDeliveryDAOMock{}, http.DefaultClient)
	deliverer = deliverer.SetBackoff(func(int) time.Duration { return 5 * time.Minute })
	if deliverer.backoff(1) != 5*time.Minute {
		t.Errorf("backoff(1) = %v, want 5m", deliverer.backoff(1))
	}
}

func TestWebhookDeliverer_UsesDefaultBackoffOnRetry(t *testing.T) {
	ctx := context.Background()
	var calls int
	var mu sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	subscription := &WebhookSubscription{URL: server.URL, Secret: "secret"}
	subscription.ID = 1
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{},
		server.Client(),
	)

	if err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`), Topic: "t"}); err != nil {
		t.Fatalf("deliver failed: %v", err)
	}

	mu.Lock()
	callCount := calls
	mu.Unlock()
	if callCount != 2 {
		t.Errorf("calls = %d, want 2 (one failure then success via default backoff)", callCount)
	}
}

func TestWebhookDeliverer_Deliver_PropagatesSubscriptionError(t *testing.T) {
	ctx := context.Background()
	subErr := errors.New("db down")
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return nil, subErr
		}},
		WebhookDeliveryDAOMock{},
		http.DefaultClient,
	)

	err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`)})
	if !errors.Is(err, subErr) {
		t.Errorf("err = %v, want subErr", err)
	}
}

func TestWebhookDeliverer_Deliver_PropagatesDeliveryLookupError(t *testing.T) {
	ctx := context.Background()
	lookupErr := errors.New("db down")
	subscription := &WebhookSubscription{URL: "https://example.com", Secret: "secret"}
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{DeliveredForEventFn: func(context.Context, uint64) ([]*WebhookDelivery, error) {
			return nil, lookupErr
		}},
		http.DefaultClient,
	)

	err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`)})
	if !errors.Is(err, lookupErr) {
		t.Errorf("err = %v, want lookupErr", err)
	}
}

func TestWebhookDeliverer_Deliver_StopsRetryingOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	subscription := &WebhookSubscription{URL: server.URL, Secret: "secret"}
	subscription.ID = 1
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{},
		server.Client(),
	)
	deliverer = deliverer.SetBackoff(func(int) time.Duration { return time.Hour })
	cancel()

	err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`)})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestWebhookDeliverer_Deliver_PropagatesInvalidURLError(t *testing.T) {
	ctx := context.Background()
	subscription := &WebhookSubscription{URL: "://bad", Secret: "secret"}
	subscription.ID = 1
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{},
		http.DefaultClient,
	).SetBackoff(func(int) time.Duration { return 0 })

	err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("deliver should fail on invalid URL")
	}
}

type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("connection refused")
}

func TestWebhookDeliverer_Deliver_PropagatesClientError(t *testing.T) {
	ctx := context.Background()
	subscription := &WebhookSubscription{URL: "https://example.com/hook", Secret: "secret"}
	subscription.ID = 1
	client := &http.Client{Transport: failingTransport{}}
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{},
		client,
	).SetBackoff(func(int) time.Duration { return 0 })

	err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`)})
	if err == nil {
		t.Fatal("deliver should fail on client error")
	}
}

func TestWebhookDeliverer_Record_LogsWhenCreateFails(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	subscription := &WebhookSubscription{URL: server.URL, Secret: "secret"}
	subscription.ID = 1
	deliverer := NewWebhookDeliverer(
		WebhookSubscriptionDAOMock{ListEnabledForOrgFn: func(context.Context, *uint64) ([]*WebhookSubscription, error) {
			return []*WebhookSubscription{subscription}, nil
		}},
		WebhookDeliveryDAOMock{CreateFunc: func(context.Context, *WebhookDelivery) (*WebhookDelivery, error) {
			return nil, errors.New("db down")
		}},
		server.Client(),
	)

	if err := deliverer.Deliver(ctx, &IntegrationEvent{Payload: []byte(`{}`)}); err != nil {
		t.Fatalf("deliver failed: %v", err)
	}
}
