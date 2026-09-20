package pos

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestPOSSessionDAO_Find_ReturnsSession(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pos_sessions" WHERE id = $1 ORDER BY "pos_sessions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "config_id", "cashier_id", "state"}).
			AddRow(1, 10, 7, SessionStateOpened))

	sessions := NewPOSSessionDAO(db)

	session, err := sessions.Find(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session == nil || session.ConfigID != 10 || session.CashierID != 7 || session.State != SessionStateOpened {
		t.Errorf("session = %+v, want opened config 10 cashier 7", session)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSSessionDAO_Create_ReturnsSession(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_sessions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	mock.ExpectCommit()

	sessions := NewPOSSessionDAO(db)

	created, err := sessions.Create(ctx, &POSSession{ConfigID: 10, State: SessionStateOpened})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 3 {
		t.Errorf("id = %d, want 3", created.ID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSOrderDAO_CreateWithLinesAndPayments_CreatesAll(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_order_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_payments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
	mock.ExpectCommit()

	orders := NewPOSOrderDAO(db)
	order := &POSOrder{SessionID: 1, AmountTotal: amount.FromInt64(220)}
	lines := []*POSOrderLine{{ItemID: helper.Ptr(uint64(100)), Qty: 2}}
	payments := []*POSPayment{{Method: "cash", Amount: 220}}

	created, err := orders.CreateWithLinesAndPayments(ctx, order, lines, payments)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("order id = %d, want 1", created.ID)
	}
	if lines[0].OrderID != 1 {
		t.Errorf("line order_id = %d, want 1", lines[0].OrderID)
	}
	if payments[0].OrderID != 1 {
		t.Errorf("payment order_id = %d, want 1", payments[0].OrderID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSOrderDAO_CreateWithLinesAndPayments_PropagatesOrderError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_orders"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	orders := NewPOSOrderDAO(db)

	_, err := orders.CreateWithLinesAndPayments(ctx, &POSOrder{}, nil, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSOrderDAO_CreateWithLinesAndPayments_PropagatesLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "pos_order_lines"`)).
		WillReturnError(errors.New("line failed"))
	mock.ExpectRollback()

	orders := NewPOSOrderDAO(db)

	_, err := orders.CreateWithLinesAndPayments(ctx, &POSOrder{}, []*POSOrderLine{{}}, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSOrderLineDAO_ListByOrder_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_order_lines" WHERE order_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pos_order_lines" WHERE order_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "item_id", "qty"}).
			AddRow(11, 7, 100, 2).
			AddRow(12, 7, 200, 1))

	lines := NewPOSOrderLineDAO(db)

	items, err := lines.ListByOrder(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].ItemID == nil || *items[0].ItemID != 100 || items[1].Qty != 1 {
		t.Errorf("items = %+v, want two lines", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSOrderLineDAO_ListByOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_order_lines" WHERE order_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	lines := NewPOSOrderLineDAO(db)

	_, err := lines.ListByOrder(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSPaymentDAO_ListByOrder_ReturnsPayments(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_payments" WHERE order_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "pos_payments" WHERE order_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "method", "amount"}).
			AddRow(21, 7, "cash", 100).
			AddRow(22, 7, "card", 120))

	payments := NewPOSPaymentDAO(db)

	items, err := payments.ListByOrder(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Method != "cash" || items[1].Amount != 120 {
		t.Errorf("items = %+v, want two payments", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSPaymentDAO_ListByOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "pos_payments" WHERE order_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	payments := NewPOSPaymentDAO(db)

	_, err := payments.ListByOrder(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPOSModels_TableNames(t *testing.T) {
	session := POSSession{Base: model.Base{ID: 1}}
	order := POSOrder{Base: model.Base{ID: 2}}
	line := POSOrderLine{Base: model.Base{ID: 3}}
	payment := POSPayment{Base: model.Base{ID: 4}}

	if session.TableName() != "pos_sessions" {
		t.Errorf("session table = %s, want pos_sessions", session.TableName())
	}
	if order.TableName() != "pos_orders" {
		t.Errorf("order table = %s, want pos_orders", order.TableName())
	}
	if line.TableName() != "pos_order_lines" {
		t.Errorf("line table = %s, want pos_order_lines", line.TableName())
	}
	if payment.TableName() != "pos_payments" {
		t.Errorf("payment table = %s, want pos_payments", payment.TableName())
	}
}
