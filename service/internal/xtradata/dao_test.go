package xtradata

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestNewIntegrationEventDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewIntegrationEventDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewSystemConfigDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewSystemConfigDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewWebhookSubscriptionDAO_ListEnabledForOrg_ForOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	orgID := helper.Ptr(uint64(10))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "webhook_subscriptions" WHERE enabled = $1 AND (organization_id = $2 OR organization_id IS NULL)`)).
		WithArgs(true, uint64(10)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "url", "secret", "enabled"}).
			AddRow(1, 10, "https://example.com/hook", "sec", true))

	dao := NewWebhookSubscriptionDAO(db)

	items, err := dao.ListEnabledForOrg(ctx, orgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].URL != "https://example.com/hook" || !items[0].Enabled {
		t.Errorf("items = %+v, want one enabled subscription", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewWebhookSubscriptionDAO_ListEnabledForOrg_ForGlobal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "webhook_subscriptions" WHERE enabled = $1 AND organization_id IS NULL`)).
		WithArgs(true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "url", "secret", "enabled"}))

	dao := NewWebhookSubscriptionDAO(db)

	items, err := dao.ListEnabledForOrg(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("items = %+v, want none", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewWebhookSubscriptionDAO_ListEnabledForOrg_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "webhook_subscriptions"`)).
		WillReturnError(errors.New("db down"))

	dao := NewWebhookSubscriptionDAO(db)

	_, err := dao.ListEnabledForOrg(ctx, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewWebhookDeliveryDAO_DeliveredForEvent_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	deliveredAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "webhook_deliveries" WHERE event_id = $1 AND status = $2`)).
		WithArgs(uint64(7), WebhookDeliveryStatusDelivered).
		WillReturnRows(sqlmock.NewRows([]string{"id", "event_id", "subscription_id", "organization_id", "status", "attempt", "response_code", "error", "delivered_at"}).
			AddRow(1, 7, 2, 10, WebhookDeliveryStatusDelivered, 1, 200, nil, deliveredAt))

	dao := NewWebhookDeliveryDAO(db)

	items, err := dao.DeliveredForEvent(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].SubscriptionID == nil || *items[0].SubscriptionID != 2 {
		t.Errorf("items = %+v, want one delivery for subscription 2", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewWebhookDeliveryDAO_DeliveredForEvent_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "webhook_deliveries"`)).
		WillReturnError(errors.New("db down"))

	dao := NewWebhookDeliveryDAO(db)

	_, err := dao.DeliveredForEvent(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewIdempotencyKeyDAO_Claim_InsertsWinner(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	organizationID := uint64(10)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_keys"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := NewIdempotencyKeyDAO(db)

	claimed, existing, err := dao.Claim(ctx, &organizationID, "abc", "sale-order-confirm")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !claimed || existing != nil {
		t.Errorf("claimed = %v existing = %+v, want winner claim", claimed, existing)
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewIdempotencyKeyDAO_Claim_ConflictReturnsExisting(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	organizationID := uint64(10)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "idempotency_keys"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "idempotency_keys"`)).
		WithArgs("abc", 10, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "key", "resource", "status_code", "response"}).
			AddRow(7, "abc", "invoice-create", 200, []byte(`{"ok":true}`)))
	mock.ExpectCommit()

	dao := NewIdempotencyKeyDAO(db)

	claimed, existing, err := dao.Claim(ctx, &organizationID, "abc", "payment-create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claimed || existing == nil || existing.Resource != "invoice-create" || existing.StatusCode != 200 {
		t.Errorf("claimed = %v existing = %+v, want stored conflict row", claimed, existing)
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewIdempotencyKeyDAO_Complete_FillsResponse(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	organizationID := uint64(10)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "idempotency_keys" SET "response"=$1,"status_code"=$2,"updated_at"=$3 WHERE (key = $4 AND status_code IS NULL AND response IS NULL) AND organization_id = $5`)).
		WithArgs([]byte(`{"ok":true}`), 200, sqlmock.AnyArg(), "abc", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	dao := NewIdempotencyKeyDAO(db)

	if err := dao.Complete(ctx, &organizationID, "abc", 200, []byte(`{"ok":true}`)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewIdempotencyKeyDAO_Release_HardDeletesClaim(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	organizationID := uint64(10)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "idempotency_keys" WHERE key = $1 AND organization_id = $2`)).
		WithArgs("abc", 10).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	dao := NewIdempotencyKeyDAO(db)

	if err := dao.Release(ctx, &organizationID, "abc"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSystemConfigDAO_Create(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "system_configs"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	dao := NewSystemConfigDAO(db)

	config := &reference.SystemConfig{Key: "currency", Value: []byte(`"IDR"`)}
	created, err := dao.Create(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created != config {
		t.Errorf("created = %+v, want original config", created)
	}

	query.AssertDBMockDone(t, mock)
}
