package sales

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestNewSaleOrderDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewSaleOrderDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestSaleOrderDAO_CreateWithLines_CreatesOrderAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_order_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	orders := NewSaleOrderDAO(db)
	order := &SaleOrder{ContactID: 5}
	lines := []*SaleOrderLine{{ItemID: helper.Ptr(uint64(100)), QtyOrdered: 2}}

	created, err := orders.CreateWithLines(ctx, order, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("order id = %d, want 1", created.ID)
	}
	if lines[0].OrderID != 1 {
		t.Errorf("line order_id = %d, want 1", lines[0].OrderID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderDAO_CreateWithLines_RollsBackOnOrderError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_orders"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	orders := NewSaleOrderDAO(db)
	_, err := orders.CreateWithLines(ctx, &SaleOrder{}, []*SaleOrderLine{{QtyOrdered: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderDAO_CreateWithLines_RollsBackOnLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_order_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	orders := NewSaleOrderDAO(db)
	_, err := orders.CreateWithLines(ctx, &SaleOrder{}, []*SaleOrderLine{{QtyOrdered: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderDAO_UpdateTx_SavesOrder(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "sale_orders" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	orders := NewSaleOrderDAO(db)
	order := &SaleOrder{Base: model.Base{ID: 1}, State: OrderStateSent}

	updated, err := orders.UpdateTx(ctx, db, order)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != order {
		t.Errorf("UpdateTx returned a different order")
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "sale_orders" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	orders := NewSaleOrderDAO(db)
	_, err := orders.UpdateTx(ctx, db, &SaleOrder{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewSaleOrderLineDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewSaleOrderLineDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestSaleOrderLineDAO_ListByOrder_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "sale_order_lines" WHERE order_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "sale_order_lines" WHERE order_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "qty_ordered"}).
			AddRow(1, 3, 10))

	lines := NewSaleOrderLineDAO(db)

	items, err := lines.ListByOrder(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].QtyOrdered != 10 {
		t.Errorf("items = %+v, want one line qty 10", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderLineDAO_ListByOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "sale_order_lines"`)).
		WithArgs(3).
		WillReturnError(errors.New("db down"))

	lines := NewSaleOrderLineDAO(db)

	_, err := lines.ListByOrder(ctx, 3)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderLineDAO_ReplaceLines_DeletesAndCreates(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sale_order_lines" WHERE order_id = $1`)).
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_order_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
	mock.ExpectCommit()

	lines := NewSaleOrderLineDAO(db)
	line := &SaleOrderLine{Base: model.Base{ID: 99}, QtyOrdered: 4}

	err := lines.ReplaceLines(ctx, 3, []*SaleOrderLine{line})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line.OrderID != 3 || line.ID != 21 {
		t.Errorf("line = %+v, want order_id 3 and recreated id 21", line)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderLineDAO_ReplaceLines_RollsBackOnDeleteError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sale_order_lines"`)).
		WithArgs(3).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	lines := NewSaleOrderLineDAO(db)

	err := lines.ReplaceLines(ctx, 3, []*SaleOrderLine{{QtyOrdered: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderLineDAO_ReplaceLines_RollsBackOnInsertError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "sale_order_lines"`)).
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "sale_order_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	lines := NewSaleOrderLineDAO(db)

	err := lines.ReplaceLines(ctx, 3, []*SaleOrderLine{{QtyOrdered: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderLineDAO_UpdateTx_SavesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "sale_order_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	lines := NewSaleOrderLineDAO(db)
	line := &SaleOrderLine{Base: model.Base{ID: 1}, QtyDelivered: 5}

	updated, err := lines.UpdateTx(ctx, db, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != line {
		t.Errorf("UpdateTx returned a different line")
	}

	query.AssertDBMockDone(t, mock)
}

func TestSaleOrderLineDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "sale_order_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	lines := NewSaleOrderLineDAO(db)
	_, err := lines.UpdateTx(ctx, db, &SaleOrderLine{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
