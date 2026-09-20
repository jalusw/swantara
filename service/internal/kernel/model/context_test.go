package model

import (
	"context"
	"testing"
)

func TestContextWithActorAndActorID(t *testing.T) {
	ctx := context.Background()
	if got := ActorID(ctx); got != 0 {
		t.Errorf("ActorID(empty ctx) = %d, want 0", got)
	}

	withActor := ContextWithActor(ctx, 42)
	if got := ActorID(withActor); got != 42 {
		t.Errorf("ActorID() = %d, want 42", got)
	}
}

func TestContextWithTenantAndTenantID(t *testing.T) {
	ctx := context.Background()
	if _, ok := TenantID(ctx); ok {
		t.Errorf("TenantID(empty ctx) = present, want absent")
	}

	withTenant := ContextWithTenant(ctx, 7)
	id, ok := TenantID(withTenant)
	if !ok || id != 7 {
		t.Errorf("TenantID() = %d, %v; want 7, true", id, ok)
	}
}

func TestBase_GetID(t *testing.T) {
	base := Base{ID: 7}
	if got := base.GetID(); got != 7 {
		t.Errorf("GetID() = %d, want 7", got)
	}
}
