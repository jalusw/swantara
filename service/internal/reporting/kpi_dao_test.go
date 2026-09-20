package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestReportDAO_ProcurementMetrics_ReturnsMetrics(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COUNT(*) AS count, COALESCE(AVG(cycle_days), 0) AS avg_cycle_days
		FROM (
			SELECT po.id,
			       EXTRACT(EPOCH FROM (COALESCE((
			           SELECT MIN(al.changed_at) FROM audit_logs al
			           WHERE al.table_name = 'purchase_orders' AND al.record_id = po.id AND al.action = 'update'
			       ), po.created_at) - po.created_at)) / 86400.0 AS cycle_days
			FROM purchase_orders po
			WHERE po.organization_id = ? AND po.state IN ('confirmed', 'done') AND po.deleted_at IS NULL
			  AND po.order_date >= ? AND po.order_date <= ?
		) t`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count", "avg_cycle_days"}).AddRow(4, 3.5))
	expectRawQuery(mock, `
		SELECT COALESCE(COUNT(*) FILTER (WHERE sm.max_done IS NOT NULL), 0) AS received_count,
		       COALESCE(COUNT(*) FILTER (WHERE sm.max_done IS NOT NULL AND sm.max_done <= po.expected_date), 0) AS on_time_count
		FROM purchase_orders po
		LEFT JOIN (
			SELECT sm.origin_id AS po_id, MAX(sm.date_done) AS max_done
			FROM stock_movements sm
			WHERE sm.origin_type = 'purchase_order' AND sm.state = 'done' AND sm.deleted_at IS NULL
			GROUP BY sm.origin_id
		) sm ON sm.po_id = po.id
		WHERE po.organization_id = ? AND po.state IN ('confirmed', 'done') AND po.expected_date IS NOT NULL AND po.deleted_at IS NULL
		  AND po.order_date >= ? AND po.order_date <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"received_count", "on_time_count"}).AddRow(4, 3))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(po_line.price_subtotal), 0) AS actual_cost,
		       COALESCE(SUM(sp.price * po_line.qty_received), 0) AS standard_cost
		FROM purchase_order_lines po_line
		JOIN purchase_orders po ON po.id = po_line.order_id
		JOIN supplier_products sp ON sp.item_id = po_line.item_id AND sp.supplier_id = po.supplier_id
		WHERE po.organization_id = ? AND po.state IN ('confirmed', 'done') AND po_line.deleted_at IS NULL AND po.deleted_at IS NULL
		  AND po.order_date >= ? AND po.order_date <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"actual_cost", "standard_cost"}).AddRow(4500, 4000))

	reports := NewReportDAO(db)

	metrics, err := reports.ProcurementMetrics(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.Count != 4 || metrics.AvgCycleDays != 3.5 || metrics.ReceivedCount != 4 || metrics.OnTimeCount != 3 {
		t.Errorf("metrics = %+v, want 4 orders with 3 on time", metrics)
	}
	if metrics.ActualCost != 4500 || metrics.StandardCost != 4000 {
		t.Errorf("costs = %+v, want 4500/4000", metrics)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ProcurementMetrics_PropagatesOrdersError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ProcurementMetrics(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ProcurementMetrics_PropagatesDeliveryError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COUNT(*) AS count, COALESCE(AVG(cycle_days), 0) AS avg_cycle_days
		FROM (
			SELECT po.id,
			       EXTRACT(EPOCH FROM (COALESCE((
			           SELECT MIN(al.changed_at) FROM audit_logs al
			           WHERE al.table_name = 'purchase_orders' AND al.record_id = po.id AND al.action = 'update'
			       ), po.created_at) - po.created_at)) / 86400.0 AS cycle_days
			FROM purchase_orders po
			WHERE po.organization_id = ? AND po.state IN ('confirmed', 'done') AND po.deleted_at IS NULL
			  AND po.order_date >= ? AND po.order_date <= ?
		) t`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count", "avg_cycle_days"}).AddRow(0, 0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ProcurementMetrics(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ProcurementMetrics_PropagatesCostError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COUNT(*) AS count, COALESCE(AVG(cycle_days), 0) AS avg_cycle_days
		FROM (
			SELECT po.id,
			       EXTRACT(EPOCH FROM (COALESCE((
			           SELECT MIN(al.changed_at) FROM audit_logs al
			           WHERE al.table_name = 'purchase_orders' AND al.record_id = po.id AND al.action = 'update'
			       ), po.created_at) - po.created_at)) / 86400.0 AS cycle_days
			FROM purchase_orders po
			WHERE po.organization_id = ? AND po.state IN ('confirmed', 'done') AND po.deleted_at IS NULL
			  AND po.order_date >= ? AND po.order_date <= ?
		) t`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count", "avg_cycle_days"}).AddRow(0, 0))
	expectRawQuery(mock, `
		SELECT COALESCE(COUNT(*) FILTER (WHERE sm.max_done IS NOT NULL), 0) AS received_count,
		       COALESCE(COUNT(*) FILTER (WHERE sm.max_done IS NOT NULL AND sm.max_done <= po.expected_date), 0) AS on_time_count
		FROM purchase_orders po
		LEFT JOIN (
			SELECT sm.origin_id AS po_id, MAX(sm.date_done) AS max_done
			FROM stock_movements sm
			WHERE sm.origin_type = 'purchase_order' AND sm.state = 'done' AND sm.deleted_at IS NULL
			GROUP BY sm.origin_id
		) sm ON sm.po_id = po.id
		WHERE po.organization_id = ? AND po.state IN ('confirmed', 'done') AND po.expected_date IS NOT NULL AND po.deleted_at IS NULL
		  AND po.order_date >= ? AND po.order_date <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"received_count", "on_time_count"}).AddRow(0, 0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ProcurementMetrics(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ManufacturingMetrics_ReturnsMetrics(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COUNT(*) AS order_count,
		       COALESCE(SUM(qty_to_produce), 0) AS qty_to_produce,
		       COALESCE(SUM(qty_produced), 0) AS qty_produced
		FROM manufacturing_orders mo
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"order_count", "qty_to_produce", "qty_produced"}).AddRow(2, 10, 9))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(wo.planned_minutes), 0) AS planned_minutes,
		       COALESCE(SUM(wo.actual_minutes), 0) AS actual_minutes
		FROM work_orders wo
		JOIN manufacturing_orders mo ON mo.id = wo.production_order_id
		WHERE mo.organization_id = ? AND wo.state = 'done' AND wo.deleted_at IS NULL AND mo.deleted_at IS NULL
		  AND wo.date_finished >= ? AND wo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"planned_minutes", "actual_minutes"}).AddRow(100, 80))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(mc.qty_planned), 0) AS planned_material,
		       COALESCE(SUM(mc.qty_consumed), 0) AS consumed_material
		FROM mo_components mc
		JOIN manufacturing_orders mo ON mo.id = mc.production_order_id
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mc.deleted_at IS NULL AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"planned_material", "consumed_material"}).AddRow(10, 11))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(m.unit_cost * m.qty), 0) AS actual_value,
		       COALESCE(SUM(pt.standard_cost * m.qty), 0) AS standard_value
		FROM stock_movements m
		JOIN manufacturing_orders mo ON mo.id = m.origin_id AND m.origin_type = 'manufacturing_order'
		JOIN item_variants pv ON pv.id = m.item_id
		JOIN items pt ON pt.id = pv.item_id
		WHERE mo.organization_id = ? AND mo.state = 'done' AND m.state = 'done'
		  AND m.deleted_at IS NULL AND mo.deleted_at IS NULL AND pv.deleted_at IS NULL AND pt.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"actual_value", "standard_value"}).AddRow(1100, 1000))

	reports := NewReportDAO(db)

	metrics, err := reports.ManufacturingMetrics(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.OrderCount != 2 || metrics.QtyToProduce != 10 || metrics.QtyProduced != 9 {
		t.Errorf("orders = %+v, want 2 orders", metrics)
	}
	if metrics.PlannedMinutes != 100 || metrics.ActualMinutes != 80 || metrics.PlannedMaterial != 10 || metrics.ConsumedMaterial != 11 {
		t.Errorf("work/material = %+v, want 100/80 and 10/11", metrics)
	}
	if metrics.ActualValue != 1100 || metrics.StandardValue != 1000 {
		t.Errorf("values = %+v, want 1100/1000", metrics)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ManufacturingMetrics_PropagatesOrdersError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ManufacturingMetrics(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ManufacturingMetrics_PropagatesWorkError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COUNT(*) AS order_count,
		       COALESCE(SUM(qty_to_produce), 0) AS qty_to_produce,
		       COALESCE(SUM(qty_produced), 0) AS qty_produced
		FROM manufacturing_orders mo
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"order_count", "qty_to_produce", "qty_produced"}).AddRow(0, 0, 0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ManufacturingMetrics(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ManufacturingMetrics_PropagatesMaterialError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COUNT(*) AS order_count,
		       COALESCE(SUM(qty_to_produce), 0) AS qty_to_produce,
		       COALESCE(SUM(qty_produced), 0) AS qty_produced
		FROM manufacturing_orders mo
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"order_count", "qty_to_produce", "qty_produced"}).AddRow(0, 0, 0))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(wo.planned_minutes), 0) AS planned_minutes,
		       COALESCE(SUM(wo.actual_minutes), 0) AS actual_minutes
		FROM work_orders wo
		JOIN manufacturing_orders mo ON mo.id = wo.production_order_id
		WHERE mo.organization_id = ? AND wo.state = 'done' AND wo.deleted_at IS NULL AND mo.deleted_at IS NULL
		  AND wo.date_finished >= ? AND wo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"planned_minutes", "actual_minutes"}).AddRow(0, 0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ManufacturingMetrics(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ManufacturingMetrics_PropagatesValueError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COUNT(*) AS order_count,
		       COALESCE(SUM(qty_to_produce), 0) AS qty_to_produce,
		       COALESCE(SUM(qty_produced), 0) AS qty_produced
		FROM manufacturing_orders mo
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"order_count", "qty_to_produce", "qty_produced"}).AddRow(0, 0, 0))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(wo.planned_minutes), 0) AS planned_minutes,
		       COALESCE(SUM(wo.actual_minutes), 0) AS actual_minutes
		FROM work_orders wo
		JOIN manufacturing_orders mo ON mo.id = wo.production_order_id
		WHERE mo.organization_id = ? AND wo.state = 'done' AND wo.deleted_at IS NULL AND mo.deleted_at IS NULL
		  AND wo.date_finished >= ? AND wo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"planned_minutes", "actual_minutes"}).AddRow(0, 0))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(mc.qty_planned), 0) AS planned_material,
		       COALESCE(SUM(mc.qty_consumed), 0) AS consumed_material
		FROM mo_components mc
		JOIN manufacturing_orders mo ON mo.id = mc.production_order_id
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mc.deleted_at IS NULL AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"planned_material", "consumed_material"}).AddRow(0, 0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ManufacturingMetrics(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ArApMetrics_ReturnsMetrics(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	yearStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice'), 0) AS open_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill'), 0) AS open_ap,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ap
		FROM invoices
		WHERE organization_id = ? AND state = 'posted' AND payment_state <> 'paid' AND deleted_at IS NULL`, asOf, asOf, 1).
		WillReturnRows(sqlmock.NewRows([]string{"open_ar", "open_ap", "overdue_ar", "overdue_ap"}).AddRow(1000, 500, 100, 50))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.credit - l.debit), 0) AS revenue
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type = 'income'
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, yearStart, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"revenue"}).AddRow(7300))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS purchases
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type IN ('cogs', 'expense')
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, yearStart, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"purchases"}).AddRow(3650))

	reports := NewReportDAO(db)

	metrics, err := reports.ArApMetrics(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.OpenAR != 1000 || metrics.OpenAP != 500 || metrics.OverdueAR != 100 || metrics.OverdueAP != 50 {
		t.Errorf("open = %+v, want 1000/500/100/50", metrics)
	}
	if metrics.Revenue != 7300 || metrics.Purchases != 3650 {
		t.Errorf("flows = %+v, want 7300/3650", metrics)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ArApMetrics_PropagatesOpenError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ArApMetrics(ctx, 1, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ArApMetrics_PropagatesRevenueError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice'), 0) AS open_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill'), 0) AS open_ap,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ap
		FROM invoices
		WHERE organization_id = ? AND state = 'posted' AND payment_state <> 'paid' AND deleted_at IS NULL`, asOf, asOf, 1).
		WillReturnRows(sqlmock.NewRows([]string{"open_ar", "open_ap", "overdue_ar", "overdue_ap"}).AddRow(0, 0, 0, 0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ArApMetrics(ctx, 1, asOf)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_ArApMetrics_PropagatesPurchasesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	yearStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice'), 0) AS open_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill'), 0) AS open_ap,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ap
		FROM invoices
		WHERE organization_id = ? AND state = 'posted' AND payment_state <> 'paid' AND deleted_at IS NULL`, asOf, asOf, 1).
		WillReturnRows(sqlmock.NewRows([]string{"open_ar", "open_ap", "overdue_ar", "overdue_ap"}).AddRow(0, 0, 0, 0))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.credit - l.debit), 0) AS revenue
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type = 'income'
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, yearStart, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"revenue"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.ArApMetrics(ctx, 1, asOf)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_CashMetrics_ReturnsMetrics(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS bank_balance
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date <= ? AND a.type IN ('cash', 'bank')
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"bank_balance"}).AddRow(2000))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(amount) FILTER (WHERE type = 'outbound'), 0) AS outflows,
		       COALESCE(SUM(amount) FILTER (WHERE type = 'inbound'), 0) AS inflows
		FROM payments
		WHERE organization_id = ? AND state = 'posted' AND date >= ? AND date <= ? AND deleted_at IS NULL`, 1, monthStart, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"outflows", "inflows"}).AddRow(1500, 500))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice'), 0) AS open_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill'), 0) AS open_ap
		FROM invoices
		WHERE organization_id = ? AND state = 'posted' AND payment_state <> 'paid' AND deleted_at IS NULL`, 1).
		WillReturnRows(sqlmock.NewRows([]string{"open_ar", "open_ap"}).AddRow(1000, 700))

	reports := NewReportDAO(db)

	metrics, err := reports.CashMetrics(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.BankBalance != 2000 || metrics.Outflows != 1500 || metrics.Inflows != 500 {
		t.Errorf("position = %+v, want 2000/1500/500", metrics)
	}
	if metrics.OpenAR != 1000 || metrics.OpenAP != 700 {
		t.Errorf("open = %+v, want 1000/700", metrics)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_CashMetrics_PropagatesPositionError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.CashMetrics(ctx, 1, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_CashMetrics_PropagatesFlowError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS bank_balance
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date <= ? AND a.type IN ('cash', 'bank')
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"bank_balance"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.CashMetrics(ctx, 1, asOf)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_CashMetrics_PropagatesOpenError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS bank_balance
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date <= ? AND a.type IN ('cash', 'bank')
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"bank_balance"}).AddRow(0))
	expectRawQuery(mock, `
		SELECT COALESCE(SUM(amount) FILTER (WHERE type = 'outbound'), 0) AS outflows,
		       COALESCE(SUM(amount) FILTER (WHERE type = 'inbound'), 0) AS inflows
		FROM payments
		WHERE organization_id = ? AND state = 'posted' AND date >= ? AND date <= ? AND deleted_at IS NULL`, 1, monthStart, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"outflows", "inflows"}).AddRow(0, 0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.CashMetrics(ctx, 1, asOf)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_InventoryRatioMetrics_ReturnsMetrics(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS cogs
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type = 'cogs'
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"cogs"}).AddRow(3000))
	expectRawQuery(mock, `
		SELECT COALESCE(AVG(value), 0) AS avg_inventory
		FROM (
			SELECT COALESCE(SUM(remaining_value), 0) AS value
			FROM cost_layers l
			WHERE l.remaining_qty <> 0 AND l.deleted_at IS NULL AND EXISTS (
				SELECT 1 FROM items pt WHERE pt.id = l.item_id AND pt.organization_id = ? AND pt.deleted_at IS NULL
			)
			GROUP BY l.item_id
		) t`, 1).
		WillReturnRows(sqlmock.NewRows([]string{"avg_inventory"}).AddRow(1000))
	expectRawQuery(mock, `
		SELECT COUNT(*) AS stockout_count
		FROM reorder_rules rr
		JOIN items pt ON pt.id = rr.item_id
		WHERE rr.active = true AND rr.deleted_at IS NULL AND pt.organization_id = ? AND pt.deleted_at IS NULL AND EXISTS (
			SELECT 1 FROM stock_balances sq
			JOIN stock_locations sl ON sl.id = sq.location_id
			WHERE sq.item_id = rr.item_id AND sl.usage = 'internal' AND sq.deleted_at IS NULL
			GROUP BY sq.item_id
			HAVING COALESCE(SUM(sq.quantity), 0) < rr.min_qty
		)`, 1).
		WillReturnRows(sqlmock.NewRows([]string{"stockout_count"}).AddRow(2))

	reports := NewReportDAO(db)

	metrics, err := reports.InventoryRatioMetrics(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if metrics.COGS != 3000 || metrics.AvgInventory != 1000 || metrics.StockoutCount != 2 {
		t.Errorf("metrics = %+v, want 3000/1000/2", metrics)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_InventoryRatioMetrics_PropagatesCOGSError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.InventoryRatioMetrics(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_InventoryRatioMetrics_PropagatesValueError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS cogs
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type = 'cogs'
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"cogs"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.InventoryRatioMetrics(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_InventoryRatioMetrics_PropagatesStockoutError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS cogs
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type = 'cogs'
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"cogs"}).AddRow(0))
	expectRawQuery(mock, `
		SELECT COALESCE(AVG(value), 0) AS avg_inventory
		FROM (
			SELECT COALESCE(SUM(remaining_value), 0) AS value
			FROM cost_layers l
			WHERE l.remaining_qty <> 0 AND l.deleted_at IS NULL AND EXISTS (
				SELECT 1 FROM items pt WHERE pt.id = l.item_id AND pt.organization_id = ? AND pt.deleted_at IS NULL
			)
			GROUP BY l.item_id
		) t`, 1).
		WillReturnRows(sqlmock.NewRows([]string{"avg_inventory"}).AddRow(0))
	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.InventoryRatioMetrics(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
