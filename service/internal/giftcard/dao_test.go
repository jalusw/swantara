package giftcard

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

func TestGiftCardDAO_UpdateTx_UpdatesCard(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "gift_cards" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	cards := NewGiftCardDAO(db)
	card := &GiftCard{Base: model.Base{ID: 42}, OrganizationID: helper.Ptr(uint64(10)), Code: "GC-001", Balance: 50}

	updated, err := cards.UpdateTx(ctx, db, card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != card {
		t.Fatalf("expected same card back")
	}

	query.AssertDBMockDone(t, mock)
}

func TestGiftCardDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "gift_cards" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	cards := NewGiftCardDAO(db)

	_, err := cards.UpdateTx(ctx, db, &GiftCard{Base: model.Base{ID: 42}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestGiftCardDAO_ListDueForExpiry_ReturnsCards(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "gift_cards" WHERE organization_id = $1 AND state = $2 AND balance > 0 AND expiry_date IS NOT NULL AND expiry_date < $3 AND deleted_at IS NULL ORDER BY expiry_date`)).
		WithArgs(10, GiftCardStateActive, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"id", "organization_id", "code", "balance", "state", "expiry_date"}).
			AddRow(42, 10, "GC-001", 100, GiftCardStateActive, asOf))

	cards := NewGiftCardDAO(db)

	due, err := cards.ListDueForExpiry(ctx, 10, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(due) != 1 || due[0].Code != "GC-001" || due[0].Balance != 100 {
		t.Errorf("due = %+v, want one card with balance 100", due)
	}

	query.AssertDBMockDone(t, mock)
}

func TestGiftCardDAO_ListDueForExpiry_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "gift_cards" WHERE organization_id = $1`)).
		WithArgs(10, GiftCardStateActive, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)).
		WillReturnError(errors.New("db down"))

	cards := NewGiftCardDAO(db)

	_, err := cards.ListDueForExpiry(ctx, 10, time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC))
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestGiftCardTransactionDAO_CreateTx_CreatesTransaction(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "gift_card_transactions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	transactions := NewGiftCardTransactionDAO(db)
	transaction := &GiftCardTransaction{GiftCardID: 42, Type: TransactionIssue, Amount: 100}

	created, err := transactions.CreateTx(ctx, db, transaction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("id = %d, want 1", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestGiftCardTransactionDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "gift_card_transactions"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	transactions := NewGiftCardTransactionDAO(db)

	_, err := transactions.CreateTx(ctx, db, &GiftCardTransaction{GiftCardID: 42})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestGiftCardTransactionDAO_ListByGiftCard_ReturnsTransactions(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "gift_card_transactions" WHERE gift_card_id = $1`)).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "gift_card_transactions" WHERE gift_card_id = $1`)).
		WithArgs(42).
		WillReturnRows(sqlmock.NewRows([]string{"id", "gift_card_id", "type", "amount"}).
			AddRow(11, 42, TransactionIssue, 100).
			AddRow(12, 42, TransactionRedeem, 30))

	transactions := NewGiftCardTransactionDAO(db)

	items, err := transactions.ListByGiftCard(ctx, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Type != TransactionIssue || items[1].Amount != 30 {
		t.Errorf("items = %+v, want two transactions", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestGiftCardTransactionDAO_ListByGiftCard_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "gift_card_transactions" WHERE gift_card_id = $1`)).
		WithArgs(42).
		WillReturnError(errors.New("db down"))

	transactions := NewGiftCardTransactionDAO(db)

	_, err := transactions.ListByGiftCard(ctx, 42)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
