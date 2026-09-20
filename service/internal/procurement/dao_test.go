package procurement

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

func TestPurchaseRequestDAO_CreateWithLines_CreatesRequisitionAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_requests"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_request_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	requisitions := NewPurchaseRequestDAO(db)
	requisition := &PurchaseRequest{RequesterID: 5}
	lines := []*PurchaseRequestLine{{ItemID: helper.Ptr(uint64(200)), Qty: 3}}

	created, err := requisitions.CreateWithLines(ctx, requisition, lines)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("requisition id = %d, want 1", created.ID)
	}
	if lines[0].RequestID != 1 {
		t.Errorf("line request_id = %d, want 1", lines[0].RequestID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseRequestDAO_CreateWithLines_RollsBackOnLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_requests"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_request_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	requisitions := NewPurchaseRequestDAO(db)
	_, err := requisitions.CreateWithLines(ctx, &PurchaseRequest{}, []*PurchaseRequestLine{{Qty: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseRequestDAO_UpdateTx_SavesRequisition(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "purchase_requests" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	requisitions := NewPurchaseRequestDAO(db)
	requisition := &PurchaseRequest{Base: model.Base{ID: 1}, State: RequestStateConfirmed}

	updated, err := requisitions.UpdateTx(ctx, db, requisition)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != requisition {
		t.Errorf("UpdateTx returned a different requisition")
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseRequestDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "purchase_requests" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	requisitions := NewPurchaseRequestDAO(db)
	_, err := requisitions.UpdateTx(ctx, db, &PurchaseRequest{Base: model.Base{ID: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseRequestLineDAO_ListByRequest_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "purchase_request_lines" WHERE request_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "purchase_request_lines" WHERE request_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "request_id", "qty"}).
			AddRow(1, 7, 3).
			AddRow(2, 7, 4))

	lines := NewPurchaseRequestLineDAO(db)

	items, err := lines.ListByRequest(ctx, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Qty != 3 || items[1].Qty != 4 {
		t.Errorf("items = %+v, want two lines", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseRequestLineDAO_ListByRequest_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "purchase_request_lines"`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	lines := NewPurchaseRequestLineDAO(db)

	_, err := lines.ListByRequest(ctx, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseRequestLineDAO_ReplaceLines_DeletesAndCreates(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "purchase_request_lines" WHERE request_id = $1`)).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_request_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
	mock.ExpectCommit()

	lines := NewPurchaseRequestLineDAO(db)
	line := &PurchaseRequestLine{Base: model.Base{ID: 99}, Qty: 2}

	err := lines.ReplaceLines(ctx, 7, []*PurchaseRequestLine{line})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line.RequestID != 7 || line.ID != 21 {
		t.Errorf("line = %+v, want request_id 7 and recreated id 21", line)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseRequestLineDAO_ReplaceLines_RollsBackOnDeleteError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "purchase_request_lines"`)).
		WithArgs(7).
		WillReturnError(errors.New("delete failed"))
	mock.ExpectRollback()

	lines := NewPurchaseRequestLineDAO(db)

	err := lines.ReplaceLines(ctx, 7, []*PurchaseRequestLine{{Qty: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderDAO_CreateWithLines_CreatesOrderAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_order_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	orders := NewPurchaseOrderDAO(db)
	order := &PurchaseOrder{SupplierID: supplierID}
	lines := []*PurchaseOrderLine{{QtyOrdered: 5}}

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

func TestPurchaseOrderDAO_CreateWithLines_RollsBackOnLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_orders"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_order_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	orders := NewPurchaseOrderDAO(db)
	_, err := orders.CreateWithLines(ctx, &PurchaseOrder{}, []*PurchaseOrderLine{{QtyOrdered: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderDAO_UpdateTx_SavesOrder(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "purchase_orders" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	orders := NewPurchaseOrderDAO(db)
	order := &PurchaseOrder{Base: model.Base{ID: 1}, State: PurchaseOrderStateSent}

	updated, err := orders.UpdateTx(ctx, db, order)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != order {
		t.Errorf("UpdateTx returned a different order")
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "purchase_orders" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	orders := NewPurchaseOrderDAO(db)
	_, err := orders.UpdateTx(ctx, db, &PurchaseOrder{Base: model.Base{ID: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderLineDAO_ListByOrder_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "purchase_order_lines" WHERE order_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "purchase_order_lines" WHERE order_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id", "qty_ordered"}).
			AddRow(1, 3, 10))

	lines := NewPurchaseOrderLineDAO(db)

	items, err := lines.ListByOrder(ctx, 3)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].QtyOrdered != 10 {
		t.Errorf("items = %+v, want one line qty 10", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderLineDAO_ListByOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "purchase_order_lines"`)).
		WithArgs(3).
		WillReturnError(errors.New("db down"))

	lines := NewPurchaseOrderLineDAO(db)

	_, err := lines.ListByOrder(ctx, 3)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderLineDAO_ReplaceLines_DeletesAndCreates(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "purchase_order_lines" WHERE order_id = $1`)).
		WithArgs(3).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "purchase_order_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
	mock.ExpectCommit()

	lines := NewPurchaseOrderLineDAO(db)
	line := &PurchaseOrderLine{Base: model.Base{ID: 99}, QtyOrdered: 4}

	err := lines.ReplaceLines(ctx, 3, []*PurchaseOrderLine{line})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if line.OrderID != 3 || line.ID != 21 {
		t.Errorf("line = %+v, want order_id 3 and recreated id 21", line)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderLineDAO_UpdateTx_SavesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "purchase_order_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	lines := NewPurchaseOrderLineDAO(db)
	line := &PurchaseOrderLine{Base: model.Base{ID: 1}, QtyReceived: 5}

	updated, err := lines.UpdateTx(ctx, db, line)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != line {
		t.Errorf("UpdateTx returned a different line")
	}

	query.AssertDBMockDone(t, mock)
}

func TestPurchaseOrderLineDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "purchase_order_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	lines := NewPurchaseOrderLineDAO(db)
	_, err := lines.UpdateTx(ctx, db, &PurchaseOrderLine{Base: model.Base{ID: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteRequestDAO_CreateWithLines_CreatesRFQAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quoteRequests"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quoteRequest_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	quote_requests := NewSupplierQuoteRequestDAO(db)
	quoteRequest := &SupplierQuoteRequest{RequesterID: 5}
	lines := []*SupplierQuoteRequestLine{{Qty: 2}}

	created, err := quote_requests.CreateWithLines(ctx, quoteRequest, lines)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("quoteRequest id = %d, want 1", created.ID)
	}
	if lines[0].QuoteRequestID != 1 {
		t.Errorf("line quoteRequest_id = %d, want 1", lines[0].QuoteRequestID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteRequestDAO_CreateWithLines_RollsBackOnLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quoteRequests"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quoteRequest_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	quote_requests := NewSupplierQuoteRequestDAO(db)
	_, err := quote_requests.CreateWithLines(ctx, &SupplierQuoteRequest{}, []*SupplierQuoteRequestLine{{Qty: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteRequestDAO_UpdateTx_SavesRFQ(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "supplier_quoteRequests" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	quote_requests := NewSupplierQuoteRequestDAO(db)
	quoteRequest := &SupplierQuoteRequest{Base: model.Base{ID: 1}, State: QuoteRequestStateSent}

	updated, err := quote_requests.UpdateTx(ctx, db, quoteRequest)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != quoteRequest {
		t.Errorf("UpdateTx returned a different quoteRequest")
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteRequestDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "supplier_quoteRequests" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	quote_requests := NewSupplierQuoteRequestDAO(db)
	_, err := quote_requests.UpdateTx(ctx, db, &SupplierQuoteRequest{Base: model.Base{ID: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteRequestLineDAO_ListByRFQ_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_quoteRequest_lines" WHERE quoteRequest_id = $1`)).
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_quoteRequest_lines" WHERE quoteRequest_id = $1`)).
		WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "quoteRequest_id", "qty"}).
			AddRow(1, 4, 6))

	lines := NewSupplierQuoteRequestLineDAO(db)

	items, err := lines.ListByRFQ(ctx, 4)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Qty != 6 {
		t.Errorf("items = %+v, want one line qty 6", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteRequestLineDAO_ListByRFQ_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_quoteRequest_lines"`)).
		WithArgs(4).
		WillReturnError(errors.New("db down"))

	lines := NewSupplierQuoteRequestLineDAO(db)

	_, err := lines.ListByRFQ(ctx, 4)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteDAO_CreateWithLines_CreatesQuoteAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quotes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quote_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	quotes := NewSupplierQuoteDAO(db)
	quote := &SupplierQuote{SupplierID: supplierID}
	lines := []*SupplierQuoteLine{{Qty: 2, UnitPrice: 25}}

	created, err := quotes.CreateWithLines(ctx, quote, lines)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("quote id = %d, want 1", created.ID)
	}
	if lines[0].SupplierQuoteID != 1 {
		t.Errorf("line supplier_quote_id = %d, want 1", lines[0].SupplierQuoteID)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteDAO_CreateWithLines_RollsBackOnLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quotes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "supplier_quote_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	quotes := NewSupplierQuoteDAO(db)
	_, err := quotes.CreateWithLines(ctx, &SupplierQuote{}, []*SupplierQuoteLine{{Qty: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteDAO_UpdateTx_SavesQuote(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "supplier_quotes" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	quotes := NewSupplierQuoteDAO(db)
	quote := &SupplierQuote{Base: model.Base{ID: 1}, State: SupplierQuoteStateSubmitted}

	updated, err := quotes.UpdateTx(ctx, db, quote)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != quote {
		t.Errorf("UpdateTx returned a different quote")
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "supplier_quotes" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	quotes := NewSupplierQuoteDAO(db)
	_, err := quotes.UpdateTx(ctx, db, &SupplierQuote{Base: model.Base{ID: 1}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteLineDAO_ListByQuote_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_quote_lines" WHERE supplier_quote_id = $1`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_quote_lines" WHERE supplier_quote_id = $1`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "supplier_quote_id", "qty"}).
			AddRow(1, 5, 8))

	lines := NewSupplierQuoteLineDAO(db)

	items, err := lines.ListByQuote(ctx, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Qty != 8 {
		t.Errorf("items = %+v, want one line qty 8", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierQuoteLineDAO_ListByQuote_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_quote_lines"`)).
		WithArgs(5).
		WillReturnError(errors.New("db down"))

	lines := NewSupplierQuoteLineDAO(db)

	_, err := lines.ListByQuote(ctx, 5)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
