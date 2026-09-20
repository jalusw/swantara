package subscription

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestNewSubscriptionDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewSubscriptionDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewSubscriptionLineDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewSubscriptionLineDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestSubscriptionDAO_ListDue_ReturnsDueSubscriptions(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscriptions" WHERE state = $1 AND next_invoice_date IS NOT NULL AND next_invoice_date <= $2`)).
		WithArgs(SubscriptionStateActive, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "state", "next_invoice_date"}).
			AddRow(1, 10, SubscriptionStateActive, asOf))

	dao := NewSubscriptionDAO(db)

	items, err := dao.ListDue(ctx, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("items = %+v, want one due subscription", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionDAO_ListDue_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscriptions"`)).
		WillReturnError(errors.New("db down"))

	dao := NewSubscriptionDAO(db)

	_, err := dao.ListDue(ctx, asOf)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionDAO_UpdateTx_SavesSubscription(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "subscriptions" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dao := NewSubscriptionDAO(db)
	subscription := &Subscription{Base: model.Base{ID: 1}, State: SubscriptionStateActive}

	updated, err := dao.UpdateTx(ctx, db, subscription)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != subscription {
		t.Errorf("UpdateTx returned a different subscription")
	}

	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "subscriptions" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	dao := NewSubscriptionDAO(db)

	_, err := dao.UpdateTx(ctx, db, &Subscription{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionLineDAO_ListBySubscription_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "subscription_lines" WHERE subscription_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "subscription_lines" WHERE subscription_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "subscription_id", "item_id", "qty", "unit_price", "discount_pct"}).
			AddRow(1, 7, 100, 1, 120, 0).
			AddRow(2, 7, 101, 2, 60, 10))

	dao := NewSubscriptionLineDAO(db)

	items, err := dao.ListBySubscription(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].ItemID == nil || *items[0].ItemID != 100 {
		t.Errorf("items = %+v, want two lines for subscription 7", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSubscriptionLineDAO_ListBySubscription_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "subscription_lines"`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	dao := NewSubscriptionLineDAO(db)

	_, err := dao.ListBySubscription(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
