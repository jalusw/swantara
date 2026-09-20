package reporting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"gorm.io/gorm"
)

type TrialBalanceRow struct {
	AccountID     uint64
	Code          string
	Name          string
	AccountType   string
	OpeningDebit  float64
	OpeningCredit float64
	PeriodDebit   float64
	PeriodCredit  float64
	ClosingDebit  float64
	ClosingCredit float64
}

type AgingRow struct {
	Type         string
	CurrencyCode string
	Bucket       string
	Amount       float64
}

type InventoryValueRow struct {
	ItemID      uint64
	ProductName string
	Quantity    float64
	Value       float64
}

type TypeBalanceRow struct {
	AccountType string
	Debit       float64
	Credit      float64
}

type ForeignPositionRow struct {
	InvoiceID      uint64
	Type           string
	CurrencyCode   string
	AmountTotal    float64
	AmountResidual float64
	AccountID      uint64
	AccountType    string
	BaseBalance    float64
}

type AccountBalanceRow struct {
	AccountID   uint64
	Code        string
	Name        string
	AccountType string
	Balance     float64
}

type CashFlowSectionRow struct {
	Section string
	Amount  float64
}

type ReportDAO interface {
	TrialBalance(ctx context.Context, organizationID uint64, start, end time.Time) ([]TrialBalanceRow, error)
	Aging(ctx context.Context, organizationID uint64, asOf time.Time) ([]AgingRow, error)
	InventoryValuation(ctx context.Context, organizationID uint64) ([]InventoryValueRow, error)
	BalancesByType(ctx context.Context, organizationID uint64, start, end time.Time) ([]TypeBalanceRow, error)
	Bookings(ctx context.Context, organizationID uint64, start, end time.Time) (float64, error)
	PayrollCost(ctx context.Context, organizationID uint64, start, end time.Time) (float64, float64, error)
	OpenForeignPositions(ctx context.Context, organizationID uint64) ([]ForeignPositionRow, error)
	StatementBalances(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalanceRow, error)
	CashFlow(ctx context.Context, organizationID uint64, start, end time.Time) ([]CashFlowSectionRow, error)
	HasYearEndClose(ctx context.Context, organizationID, periodID uint64) (bool, error)
	ProcurementMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (ProcurementMetrics, error)
	ManufacturingMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (ManufacturingMetrics, error)
	ArApMetrics(ctx context.Context, organizationID uint64, asOf time.Time) (ArApMetrics, error)
	CashMetrics(ctx context.Context, organizationID uint64, asOf time.Time) (CashMetrics, error)
	InventoryRatioMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (InventoryRatioMetrics, error)
}

type reportDAO struct {
	db *gorm.DB
}

func NewReportDAO(db *gorm.DB) ReportDAO {
	return reportDAO{db: db}
}

func (d reportDAO) TrialBalance(ctx context.Context, organizationID uint64, start, end time.Time) ([]TrialBalanceRow, error) {
	rows := []TrialBalanceRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT l.account_id,
		       a.code,
		       a.name,
		       a.type AS account_type,
		       COALESCE(SUM(CASE WHEN m.date < ? THEN l.debit END), 0) AS opening_debit,
		       COALESCE(SUM(CASE WHEN m.date < ? THEN l.credit END), 0) AS opening_credit,
		       COALESCE(SUM(CASE WHEN m.date >= ? AND m.date <= ? THEN l.debit END), 0) AS period_debit,
		       COALESCE(SUM(CASE WHEN m.date >= ? AND m.date <= ? THEN l.credit END), 0) AS period_credit,
		       COALESCE(SUM(l.debit), 0) AS closing_debit,
		       COALESCE(SUM(l.credit), 0) AS closing_credit
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date <= ? AND m.deleted_at IS NULL AND a.deleted_at IS NULL
		GROUP BY l.account_id, a.code, a.name, a.type
		ORDER BY a.code`, start, start, start, end, start, end, organizationID, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d reportDAO) Aging(ctx context.Context, organizationID uint64, asOf time.Time) ([]AgingRow, error) {
	d30 := asOf.AddDate(0, 0, -30)
	d60 := asOf.AddDate(0, 0, -60)
	d90 := asOf.AddDate(0, 0, -90)
	rows := []AgingRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT i.type,
		       COALESCE(i.currency_code, '') AS currency_code,
		       CASE
		           WHEN i.due_date IS NULL OR i.due_date >= ? THEN 'current'
		           WHEN i.due_date >= ? THEN '1-30'
		           WHEN i.due_date >= ? THEN '31-60'
		           WHEN i.due_date >= ? THEN '61-90'
		           ELSE '90+'
		       END AS bucket,
		       COALESCE(SUM(i.amount_residual), 0) AS amount
		FROM invoices i
		WHERE i.organization_id = ? AND i.state = 'posted' AND i.payment_state <> 'paid'
		  AND i.type IN ('customer_invoice', 'supplier_bill') AND i.deleted_at IS NULL
		GROUP BY i.type, i.currency_code, bucket
		ORDER BY i.type, i.currency_code, bucket`, asOf, d30, d60, d90, organizationID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d reportDAO) InventoryValuation(ctx context.Context, organizationID uint64) ([]InventoryValueRow, error) {
	rows := []InventoryValueRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT l.item_id,
		       p.name AS product_name,
		       COALESCE(SUM(l.remaining_qty), 0) AS quantity,
		       COALESCE(SUM(l.remaining_value), 0) AS value
		FROM cost_layers l
		JOIN items p ON p.id = l.item_id
		WHERE l.remaining_qty <> 0 AND l.deleted_at IS NULL AND p.organization_id = ? AND p.deleted_at IS NULL
		GROUP BY l.item_id, p.name
		ORDER BY p.name`, organizationID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d reportDAO) BalancesByType(ctx context.Context, organizationID uint64, start, end time.Time) ([]TypeBalanceRow, error) {
	rows := []TypeBalanceRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT a.type AS account_type,
		       COALESCE(SUM(l.debit), 0) AS debit,
		       COALESCE(SUM(l.credit), 0) AS credit
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND m.deleted_at IS NULL AND a.deleted_at IS NULL
		GROUP BY a.type`, organizationID, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d reportDAO) OpenForeignPositions(ctx context.Context, organizationID uint64) ([]ForeignPositionRow, error) {
	rows := []ForeignPositionRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT i.id AS invoice_id,
		       i.type,
		       i.currency_code AS currency_code,
		       i.amount_total AS amount_total,
		       i.amount_residual AS amount_residual,
		       l.account_id,
		       a.type AS account_type,
		       COALESCE(SUM(l.debit - l.credit), 0) AS base_balance
		FROM invoices i
		JOIN journal_lines l ON l.entry_id = i.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE i.organization_id = ? AND i.state = 'posted' AND i.payment_state <> 'paid'
		  AND i.type IN ('customer_invoice', 'supplier_bill') AND i.currency_code IS NOT NULL
		  AND a.type IN ('receivable', 'payable') AND i.deleted_at IS NULL AND a.deleted_at IS NULL
		GROUP BY i.id, i.type, i.currency_code, i.amount_total, i.amount_residual, l.account_id, a.type`,
		organizationID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d reportDAO) Bookings(ctx context.Context, organizationID uint64, start, end time.Time) (float64, error) {
	var total float64
	err := d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount_total), 0)
		FROM sale_orders
		WHERE organization_id = ? AND state IN ('confirmed', 'done') AND order_date >= ? AND order_date <= ? AND deleted_at IS NULL`,
		organizationID, start, end).Scan(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (d reportDAO) PayrollCost(ctx context.Context, organizationID uint64, start, end time.Time) (float64, float64, error) {
	var row struct {
		Gross float64
		Net   float64
	}
	err := d.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(p.gross), 0) AS gross, COALESCE(SUM(p.net), 0) AS net
		FROM payslips p
		JOIN payroll_runs r ON r.id = p.run_id
		WHERE r.organization_id = ? AND r.state = 'paid' AND r.period_start >= ? AND r.period_end <= ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL`,
		organizationID, start, end).Scan(&row).Error
	if err != nil {
		return 0, 0, err
	}
	return row.Gross, row.Net, nil
}

func (d reportDAO) StatementBalances(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalanceRow, error) {
	rows := []AccountBalanceRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT l.account_id,
		       a.code,
		       a.name,
		       a.type AS account_type,
		       COALESCE(SUM(l.debit - l.credit), 0) AS balance
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ?
		  AND m.deleted_at IS NULL AND a.deleted_at IS NULL
		GROUP BY l.account_id, a.code, a.name, a.type
		ORDER BY a.code`, organizationID, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d reportDAO) CashFlow(ctx context.Context, organizationID uint64, start, end time.Time) ([]CashFlowSectionRow, error) {
	rows := []CashFlowSectionRow{}
	err := d.db.WithContext(ctx).Raw(`
		SELECT section, COALESCE(SUM(debit - credit), 0) AS amount
		FROM (
			SELECT l.id,
			       CASE
			           WHEN EXISTS (SELECT 1 FROM journal_lines c JOIN accounts ca ON ca.id = c.account_id
			                        WHERE c.entry_id = l.entry_id AND c.id <> l.id AND ca.type IN ('fixed_asset', 'depreciation') AND ca.deleted_at IS NULL) THEN 'investing'
			           WHEN EXISTS (SELECT 1 FROM journal_lines c JOIN accounts ca ON ca.id = c.account_id
			                        WHERE c.entry_id = l.entry_id AND c.id <> l.id AND ca.type = 'equity' AND ca.deleted_at IS NULL) THEN 'financing'
			           ELSE 'operating'
			       END AS section,
			       l.debit,
			       l.credit
			FROM journal_lines l
			JOIN journal_entrys m ON m.id = l.entry_id
			JOIN accounts a ON a.id = l.account_id
			WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ?
			  AND a.type IN ('cash', 'bank') AND m.deleted_at IS NULL AND a.deleted_at IS NULL
		) t
		GROUP BY section`, organizationID, start, end).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (d reportDAO) HasYearEndClose(ctx context.Context, organizationID, periodID uint64) (bool, error) {
	var count int64
	err := d.db.WithContext(ctx).
		Model(&accounting.JournalEntry{}).
		Where("organization_id = ? AND origin_type = ? AND origin_id = ?", organizationID, yearEndCloseOrigin, periodID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
