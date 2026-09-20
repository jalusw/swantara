package model

import "context"

type actorKey struct{}

var ActorKey actorKey

func ContextWithActor(ctx context.Context, actorID uint64) context.Context {
	return context.WithValue(ctx, ActorKey, actorID)
}

func ActorID(ctx context.Context) uint64 {
	id, _ := ctx.Value(ActorKey).(uint64)
	return id
}

type tenantKey struct{}

var TenantKey tenantKey

func ContextWithTenant(ctx context.Context, organizationID uint64) context.Context {
	return context.WithValue(ctx, TenantKey, organizationID)
}

func TenantID(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(TenantKey).(uint64)
	return id, ok
}
