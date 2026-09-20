package helper

import (
	"context"
	"testing"
)

func TestContextWithRequestIDRoundTrip(t *testing.T) {
	ctx := ContextWithRequestID(context.Background(), "req-123")

	if got := RequestID(ctx); got != "req-123" {
		t.Errorf("expected request id req-123, got %v", got)
	}
}

func TestRequestIDEmptyWhenAbsent(t *testing.T) {
	if got := RequestID(context.Background()); got != "" {
		t.Errorf("expected empty request id, got %v", got)
	}
}
