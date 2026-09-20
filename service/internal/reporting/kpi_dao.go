package reporting

import (
	"context"
	"time"
)

type ProcurementMetrics struct {
	Count         int
	AvgCycleDays  float64
	ReceivedCount int
	OnTimeCount   int
	ActualCost    float64
	StandardCost  float64
}

type ManufacturingMetrics struct {
	OrderCount       int
	PlannedMinutes   float64
	ActualMinutes    float64
	QtyToProduce     float64
	QtyProduced      float64
	PlannedMaterial  float64
	ConsumedMaterial float64
	ActualValue      float64
	StandardValue    float64
}

type ArApMetrics struct {
	OpenAR    float64
	OpenAP    float64
	OverdueAR float64
	OverdueAP float64
	Revenue   float64
	Purchases float64
}

type CashMetrics struct {
	BankBalance float64
	Outflows    float64
	Inflows     float64
	OpenAR      float64
	OpenAP      float64
}

type InventoryRatioMetrics struct {
	COGS          float64
	AvgInventory  float64
	StockoutCount int
}

func (d reportDAO) ProcurementMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (ProcurementMetrics, error) {
	var orders struct {
		Count        int
		AvgCycleDays float64
	}
	err := d.db.WithContext(ctx).Raw(`
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
		) t`, organizationID, start, end).Scan(&orders).Error
	if err != nil {
		return ProcurementMetrics{}, err
	}

	var delivery struct {
		ReceivedCount int
		OnTimeCount   int
	}
	err = d.db.WithContext(ctx).Raw(`
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
		  AND po.order_date >= ? AND po.order_date <= ?`, organizationID, start, end).Scan(&delivery).Error
	if err != nil {
		return ProcurementMetrics{}, err
	}

	var cost struct {
		ActualCost   float64
		StandardCost float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(po_line.price_subtotal), 0) AS actual_cost,
		       COALESCE(SUM(sp.price * po_line.qty_received), 0) AS standard_cost
		FROM purchase_order_lines po_line
		JOIN purchase_orders po ON po.id = po_line.order_id
		JOIN supplier_products sp ON sp.item_id = po_line.item_id AND sp.supplier_id = po.supplier_id
		WHERE po.organization_id = ? AND po.state IN ('confirmed', 'done') AND po_line.deleted_at IS NULL AND po.deleted_at IS NULL
		  AND po.order_date >= ? AND po.order_date <= ?`, organizationID, start, end).Scan(&cost).Error
	if err != nil {
		return ProcurementMetrics{}, err
	}
	return ProcurementMetrics{
		Count:         orders.Count,
		AvgCycleDays:  orders.AvgCycleDays,
		ReceivedCount: delivery.ReceivedCount,
		OnTimeCount:   delivery.OnTimeCount,
		ActualCost:    cost.ActualCost,
		StandardCost:  cost.StandardCost,
	}, nil
}

func (d reportDAO) ManufacturingMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (ManufacturingMetrics, error) {
	var orders struct {
		OrderCount   int
		QtyToProduce float64
		QtyProduced  float64
	}
	err := d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) AS order_count,
		       COALESCE(SUM(qty_to_produce), 0) AS qty_to_produce,
		       COALESCE(SUM(qty_produced), 0) AS qty_produced
		FROM manufacturing_orders mo
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, organizationID, start, end).Scan(&orders).Error
	if err != nil {
		return ManufacturingMetrics{}, err
	}

	var work struct {
		PlannedMinutes float64
		ActualMinutes  float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(wo.planned_minutes), 0) AS planned_minutes,
		       COALESCE(SUM(wo.actual_minutes), 0) AS actual_minutes
		FROM work_orders wo
		JOIN manufacturing_orders mo ON mo.id = wo.production_order_id
		WHERE mo.organization_id = ? AND wo.state = 'done' AND wo.deleted_at IS NULL AND mo.deleted_at IS NULL
		  AND wo.date_finished >= ? AND wo.date_finished <= ?`, organizationID, start, end).Scan(&work).Error
	if err != nil {
		return ManufacturingMetrics{}, err
	}

	var material struct {
		PlannedMaterial  float64
		ConsumedMaterial float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(mc.qty_planned), 0) AS planned_material,
		       COALESCE(SUM(mc.qty_consumed), 0) AS consumed_material
		FROM mo_components mc
		JOIN manufacturing_orders mo ON mo.id = mc.production_order_id
		WHERE mo.organization_id = ? AND mo.state = 'done' AND mc.deleted_at IS NULL AND mo.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, organizationID, start, end).Scan(&material).Error
	if err != nil {
		return ManufacturingMetrics{}, err
	}

	var value struct {
		ActualValue   float64
		StandardValue float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(m.unit_cost * m.qty), 0) AS actual_value,
		       COALESCE(SUM(pt.standard_cost * m.qty), 0) AS standard_value
		FROM stock_movements m
		JOIN manufacturing_orders mo ON mo.id = m.origin_id AND m.origin_type = 'manufacturing_order'
		JOIN item_variants pv ON pv.id = m.item_id
		JOIN items pt ON pt.id = pv.item_id
		WHERE mo.organization_id = ? AND mo.state = 'done' AND m.state = 'done'
		  AND m.deleted_at IS NULL AND mo.deleted_at IS NULL AND pv.deleted_at IS NULL AND pt.deleted_at IS NULL
		  AND mo.date_finished >= ? AND mo.date_finished <= ?`, organizationID, start, end).Scan(&value).Error
	if err != nil {
		return ManufacturingMetrics{}, err
	}
	return ManufacturingMetrics{
		OrderCount:       orders.OrderCount,
		QtyToProduce:     orders.QtyToProduce,
		QtyProduced:      orders.QtyProduced,
		PlannedMinutes:   work.PlannedMinutes,
		ActualMinutes:    work.ActualMinutes,
		PlannedMaterial:  material.PlannedMaterial,
		ConsumedMaterial: material.ConsumedMaterial,
		ActualValue:      value.ActualValue,
		StandardValue:    value.StandardValue,
	}, nil
}

func (d reportDAO) ArApMetrics(ctx context.Context, organizationID uint64, asOf time.Time) (ArApMetrics, error) {
	var open struct {
		OpenAR    float64
		OpenAP    float64
		OverdueAR float64
		OverdueAP float64
	}
	err := d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice'), 0) AS open_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill'), 0) AS open_ap,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill' AND due_date IS NOT NULL AND due_date < ?), 0) AS overdue_ap
		FROM invoices
		WHERE organization_id = ? AND state = 'posted' AND payment_state <> 'paid' AND deleted_at IS NULL`,
		asOf, asOf, organizationID).Scan(&open).Error
	if err != nil {
		return ArApMetrics{}, err
	}

	yearStart := time.Date(asOf.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	var revenue struct {
		Revenue float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(l.credit - l.debit), 0) AS revenue
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type = 'income'
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, organizationID, yearStart, asOf).Scan(&revenue).Error
	if err != nil {
		return ArApMetrics{}, err
	}

	var purchases struct {
		Purchases float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS purchases
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type IN ('cogs', 'expense')
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, organizationID, yearStart, asOf).Scan(&purchases).Error
	if err != nil {
		return ArApMetrics{}, err
	}
	return ArApMetrics{
		OpenAR:    open.OpenAR,
		OpenAP:    open.OpenAP,
		OverdueAR: open.OverdueAR,
		OverdueAP: open.OverdueAP,
		Revenue:   revenue.Revenue,
		Purchases: purchases.Purchases,
	}, nil
}

func (d reportDAO) CashMetrics(ctx context.Context, organizationID uint64, asOf time.Time) (CashMetrics, error) {
	var position struct {
		BankBalance float64
	}
	err := d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS bank_balance
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date <= ? AND a.type IN ('cash', 'bank')
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, organizationID, asOf).Scan(&position).Error
	if err != nil {
		return CashMetrics{}, err
	}

	monthStart := time.Date(asOf.Year(), asOf.Month(), 1, 0, 0, 0, 0, time.UTC)
	var flow struct {
		Outflows float64
		Inflows  float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount) FILTER (WHERE type = 'outbound'), 0) AS outflows,
		       COALESCE(SUM(amount) FILTER (WHERE type = 'inbound'), 0) AS inflows
		FROM payments
		WHERE organization_id = ? AND state = 'posted' AND date >= ? AND date <= ? AND deleted_at IS NULL`,
		organizationID, monthStart, asOf).Scan(&flow).Error
	if err != nil {
		return CashMetrics{}, err
	}

	var open struct {
		OpenAR float64
		OpenAP float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount_residual) FILTER (WHERE type = 'customer_invoice'), 0) AS open_ar,
		       COALESCE(SUM(amount_residual) FILTER (WHERE type = 'supplier_bill'), 0) AS open_ap
		FROM invoices
		WHERE organization_id = ? AND state = 'posted' AND payment_state <> 'paid' AND deleted_at IS NULL`,
		organizationID).Scan(&open).Error
	if err != nil {
		return CashMetrics{}, err
	}
	return CashMetrics{
		BankBalance: position.BankBalance,
		Outflows:    flow.Outflows,
		Inflows:     flow.Inflows,
		OpenAR:      open.OpenAR,
		OpenAP:      open.OpenAP,
	}, nil
}

func (d reportDAO) InventoryRatioMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (InventoryRatioMetrics, error) {
	var cogs struct {
		COGS float64
	}
	err := d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(l.debit - l.credit), 0) AS cogs
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND a.type = 'cogs'
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL`, organizationID, start, end).Scan(&cogs).Error
	if err != nil {
		return InventoryRatioMetrics{}, err
	}

	var value struct {
		AvgInventory float64
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(AVG(value), 0) AS avg_inventory
		FROM (
			SELECT COALESCE(SUM(remaining_value), 0) AS value
			FROM cost_layers l
			WHERE l.remaining_qty <> 0 AND l.deleted_at IS NULL AND EXISTS (
				SELECT 1 FROM items pt WHERE pt.id = l.item_id AND pt.organization_id = ? AND pt.deleted_at IS NULL
			)
			GROUP BY l.item_id
		) t`, organizationID).Scan(&value).Error
	if err != nil {
		return InventoryRatioMetrics{}, err
	}

	var stockouts struct {
		StockoutCount int
	}
	err = d.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) AS stockout_count
		FROM reorder_rules rr
		JOIN items pt ON pt.id = rr.item_id
		WHERE rr.active = true AND rr.deleted_at IS NULL AND pt.organization_id = ? AND pt.deleted_at IS NULL AND EXISTS (
			SELECT 1 FROM stock_balances sq
			JOIN stock_locations sl ON sl.id = sq.location_id
			WHERE sq.item_id = rr.item_id AND sl.usage = 'internal' AND sq.deleted_at IS NULL
			GROUP BY sq.item_id
			HAVING COALESCE(SUM(sq.quantity), 0) < rr.min_qty
		)`, organizationID).Scan(&stockouts).Error
	if err != nil {
		return InventoryRatioMetrics{}, err
	}
	return InventoryRatioMetrics{
		COGS:          cogs.COGS,
		AvgInventory:  value.AvgInventory,
		StockoutCount: stockouts.StockoutCount,
	}, nil
}
