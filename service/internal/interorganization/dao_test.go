package interorganization

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

func TestNewDropshipLinkDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewDropshipLinkDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewInterorganizationRuleDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewInterorganizationRuleDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewInterorganizationTransactionDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewInterorganizationTransactionDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewConsolidationRunDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewConsolidationRunDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestNewConsolidationEliminationDAO(t *testing.T) {
	db, _ := query.NewMockDB(t)
	if NewConsolidationEliminationDAO(db) == nil {
		t.Fatal("expected a dao, got nil")
	}
}

func TestDropshipLinkDAO_ListByPurchaseOrder_ReturnsLinks(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "dropship_links"."id","dropship_links"."created_at","dropship_links"."updated_at","dropship_links"."deleted_at","dropship_links"."sale_order_line_id","dropship_links"."purchase_order_line_id","dropship_links"."stock_movement_id" FROM "dropship_links" JOIN purchase_order_lines ON purchase_order_lines.id = dropship_links.purchase_order_line_id WHERE purchase_order_lines.order_id = $1 AND dropship_links.deleted_at IS NULL`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sale_order_line_id", "purchase_order_line_id"}).
			AddRow(1, 300, 210))

	links := NewDropshipLinkDAO(db)

	items, err := links.ListByPurchaseOrder(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].SaleOrderLineID != 300 || items[0].PurchaseOrderLineID != 210 {
		t.Errorf("items = %+v, want one link", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDropshipLinkDAO_ListByPurchaseOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "dropship_links"."id","dropship_links"."created_at","dropship_links"."updated_at","dropship_links"."deleted_at","dropship_links"."sale_order_line_id","dropship_links"."purchase_order_line_id","dropship_links"."stock_movement_id" FROM "dropship_links"`)).
		WillReturnError(errors.New("db down"))

	links := NewDropshipLinkDAO(db)

	_, err := links.ListByPurchaseOrder(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDropshipLinkDAO_UpdateTx_SavesLink(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "dropship_links" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	links := NewDropshipLinkDAO(db)
	link := &DropshipLink{Base: model.Base{ID: 1}, StockMovementID: helper.Ptr(uint64(700))}

	updated, err := links.UpdateTx(ctx, db, link)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != link {
		t.Errorf("UpdateTx returned a different link")
	}

	query.AssertDBMockDone(t, mock)
}

func TestDropshipLinkDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "dropship_links" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	links := NewDropshipLinkDAO(db)

	_, err := links.UpdateTx(ctx, db, &DropshipLink{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestConsolidationRunDAO_UpdateTx_SavesRun(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consolidation_runs" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	runs := NewConsolidationRunDAO(db)
	run := &ConsolidationRun{Base: model.Base{ID: 1}, State: RunStateDone}

	updated, err := runs.UpdateTx(ctx, db, run)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != run {
		t.Errorf("UpdateTx returned a different run")
	}

	query.AssertDBMockDone(t, mock)
}

func TestConsolidationRunDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "consolidation_runs" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	runs := NewConsolidationRunDAO(db)

	_, err := runs.UpdateTx(ctx, db, &ConsolidationRun{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestConsolidationEliminationDAO_ListByRun_ReturnsEliminations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "consolidation_eliminations" WHERE consolidation_run_id = $1`)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "consolidation_eliminations" WHERE consolidation_run_id = $1`)).
		WithArgs(9).
		WillReturnRows(sqlmock.NewRows([]string{"id", "consolidation_run_id", "account_id"}).
			AddRow(1, 9, 1001))

	eliminations := NewConsolidationEliminationDAO(db)

	items, err := eliminations.ListByRun(ctx, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].AccountID != 1001 {
		t.Errorf("items = %+v, want one elimination", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestConsolidationEliminationDAO_ListByRun_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "consolidation_eliminations"`)).
		WithArgs(9).
		WillReturnError(errors.New("db down"))

	eliminations := NewConsolidationEliminationDAO(db)

	_, err := eliminations.ListByRun(ctx, 9)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestConsolidationEliminationDAO_CreateTx_CreatesElimination(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consolidation_eliminations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	eliminations := NewConsolidationEliminationDAO(db)
	elimination := &ConsolidationElimination{ConsolidationRunID: 9, AccountID: 1001, Amount: 100}

	created, err := eliminations.CreateTx(ctx, db, elimination)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created != elimination || created.ID != 1 {
		t.Errorf("created = %+v, want original with id 1", created)
	}

	query.AssertDBMockDone(t, mock)
}

func TestConsolidationEliminationDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "consolidation_eliminations"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	eliminations := NewConsolidationEliminationDAO(db)

	_, err := eliminations.CreateTx(ctx, db, &ConsolidationElimination{AccountID: 1001})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInterorganizationRuleDAO_Create_CreatesRule(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "interorganization_rules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	rules := NewInterorganizationRuleDAO(db)
	rule := &InterorganizationRule{FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(11))}

	created, err := rules.Create(ctx, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created != rule || created.ID != 1 {
		t.Errorf("created = %+v, want original with id 1", created)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInterorganizationTransactionDAO_Create_CreatesTransaction(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "interorganization_transactions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	trans := NewInterorganizationTransactionDAO(db)
	transaction := &InterorganizationTransaction{SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), Amount: 100, State: TransactionStateDone}

	created, err := trans.Create(ctx, transaction)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created != transaction || created.ID != 1 {
		t.Errorf("created = %+v, want original with id 1", created)
	}

	query.AssertDBMockDone(t, mock)
}

func TestRecomputeDeliveryStatus_Pending(t *testing.T) {
	status := recomputeDeliveryStatus([]*sales.SaleOrderLine{{QtyOrdered: 5}})
	if status != sales.DeliveryStatusPending {
		t.Errorf("status = %s, want pending", status)
	}
}

func TestRecomputeDeliveryStatus_Done(t *testing.T) {
	status := recomputeDeliveryStatus([]*sales.SaleOrderLine{{QtyOrdered: 5, QtyDelivered: 5}})
	if status != sales.DeliveryStatusDone {
		t.Errorf("status = %s, want done", status)
	}
}

func TestRecomputeDeliveryStatus_Partial(t *testing.T) {
	status := recomputeDeliveryStatus([]*sales.SaleOrderLine{{QtyOrdered: 5, QtyDelivered: 3}})
	if status != sales.DeliveryStatusPartial {
		t.Errorf("status = %s, want partial", status)
	}
}
