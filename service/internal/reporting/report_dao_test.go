package reporting

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestReportDAO_TrialBalance_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{
		"account_id", "code", "name", "account_type",
		"opening_debit", "opening_credit", "period_debit", "period_credit", "closing_debit", "closing_credit",
	}).AddRow(11, "1100", "Cash", "asset", 100, 0, 50, 0, 150, 0))

	reports := NewReportDAO(db)

	rows, err := reports.TrialBalance(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].AccountID != 11 || rows[0].ClosingDebit != 150 {
		t.Errorf("rows = %+v, want one closing debit row", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_TrialBalance_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.TrialBalance(ctx, 1, start, end)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_Aging_ReturnsBuckets(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
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
		ORDER BY i.type, i.currency_code, bucket`,
		asOf, asOf.AddDate(0, 0, -30), asOf.AddDate(0, 0, -60), asOf.AddDate(0, 0, -90), 1).
		WillReturnRows(sqlmock.NewRows([]string{"type", "currency_code", "bucket", "amount"}).
			AddRow("customer_invoice", "USD", "current", 1000).
			AddRow("supplier_bill", "", "90+", 200))

	reports := NewReportDAO(db)

	rows, err := reports.Aging(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 || rows[0].Bucket != "current" || rows[1].Amount != 200 {
		t.Errorf("rows = %+v, want current and 90+ buckets", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_Aging_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.Aging(ctx, 1, asOf)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_InventoryValuation_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	expectRawQuery(mock, `
		SELECT l.item_id,
		       p.name AS product_name,
		       COALESCE(SUM(l.remaining_qty), 0) AS quantity,
		       COALESCE(SUM(l.remaining_value), 0) AS value
		FROM cost_layers l
		JOIN items p ON p.id = l.item_id
		WHERE l.remaining_qty <> 0 AND l.deleted_at IS NULL AND p.organization_id = ? AND p.deleted_at IS NULL
		GROUP BY l.item_id, p.name
		ORDER BY p.name`, 1).
		WillReturnRows(sqlmock.NewRows([]string{"item_id", "product_name", "quantity", "value"}).
			AddRow(5, "Widget", 10, 150))

	reports := NewReportDAO(db)

	rows, err := reports.InventoryValuation(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].ProductName != "Widget" || rows[0].Value != 150 {
		t.Errorf("rows = %+v, want one widget row", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_InventoryValuation_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.InventoryValuation(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_BalancesByType_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT a.type AS account_type,
		       COALESCE(SUM(l.debit), 0) AS debit,
		       COALESCE(SUM(l.credit), 0) AS credit
		FROM journal_lines l
		JOIN journal_entrys m ON m.id = l.entry_id
		JOIN accounts a ON a.id = l.account_id
		WHERE m.organization_id = ? AND m.state = 'posted' AND m.date >= ? AND m.date <= ? AND m.deleted_at IS NULL AND a.deleted_at IS NULL
		GROUP BY a.type`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"account_type", "debit", "credit"}).
			AddRow("income", 0, 4000).
			AddRow("cogs", 2400, 0))

	reports := NewReportDAO(db)

	rows, err := reports.BalancesByType(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 || rows[0].AccountType != "income" || rows[1].Debit != 2400 {
		t.Errorf("rows = %+v, want income and cogs rows", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_BalancesByType_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.BalancesByType(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_OpenForeignPositions_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	expectRawQuery(mock, `
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
		GROUP BY i.id, i.type, i.currency_code, i.amount_total, i.amount_residual, l.account_id, a.type`, 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"invoice_id", "type", "currency_code", "amount_total", "amount_residual", "account_id", "account_type", "base_balance",
		}).AddRow(9, "customer_invoice", "USD", 1000, 300, 22, "receivable", 300))

	reports := NewReportDAO(db)

	rows, err := reports.OpenForeignPositions(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].InvoiceID != 9 || rows[0].BaseBalance != 300 {
		t.Errorf("rows = %+v, want one receivable position", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_OpenForeignPositions_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.OpenForeignPositions(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_Bookings_ReturnsTotal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(amount_total), 0)
		FROM sale_orders
		WHERE organization_id = ? AND state IN ('confirmed', 'done') AND order_date >= ? AND order_date <= ? AND deleted_at IS NULL`,
		1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(5000))

	reports := NewReportDAO(db)

	total, err := reports.Bookings(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5000 {
		t.Errorf("total = %v, want 5000", total)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_Bookings_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.Bookings(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_PayrollCost_ReturnsGrossAndNet(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
		SELECT COALESCE(SUM(p.gross), 0) AS gross, COALESCE(SUM(p.net), 0) AS net
		FROM payslips p
		JOIN payroll_runs r ON r.id = p.run_id
		WHERE r.organization_id = ? AND r.state = 'paid' AND r.period_start >= ? AND r.period_end <= ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL`,
		1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"gross", "net"}).AddRow(5000, 3800))

	reports := NewReportDAO(db)

	gross, net, err := reports.PayrollCost(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gross != 5000 || net != 3800 {
		t.Errorf("gross/net = %v/%v, want 5000/3800", gross, net)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_PayrollCost_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, _, err := reports.PayrollCost(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_StatementBalances_ReturnsRows(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
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
		ORDER BY a.code`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "code", "name", "account_type", "balance"}).
			AddRow(11, "1100", "Cash", "asset", 150))

	reports := NewReportDAO(db)

	rows, err := reports.StatementBalances(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].Code != "1100" || rows[0].Balance != 150 {
		t.Errorf("rows = %+v, want one cash row", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_StatementBalances_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.StatementBalances(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_CashFlow_ReturnsSections(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	expectRawQuery(mock, `
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
		GROUP BY section`, 1, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"section", "amount"}).
			AddRow("operating", 500).
			AddRow("investing", -200))

	reports := NewReportDAO(db)

	rows, err := reports.CashFlow(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 || rows[0].Section != "operating" || rows[1].Amount != -200 {
		t.Errorf("rows = %+v, want operating and investing sections", rows)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_CashFlow_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery("SELECT").WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.CashFlow(ctx, 1, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_HasYearEndClose_ReturnsTrueWhenPresent(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journal_entrys" WHERE organization_id = $1 AND origin_type = $2 AND origin_id = $3`)).
		WithArgs(1, yearEndCloseOrigin, 7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	reports := NewReportDAO(db)

	closed, err := reports.HasYearEndClose(ctx, 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !closed {
		t.Errorf("closed = false, want true")
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_HasYearEndClose_ReturnsFalseWhenAbsent(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journal_entrys"`)).
		WithArgs(1, yearEndCloseOrigin, 7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	reports := NewReportDAO(db)

	closed, err := reports.HasYearEndClose(ctx, 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if closed {
		t.Errorf("closed = true, want false")
	}

	query.AssertDBMockDone(t, mock)
}

func TestReportDAO_HasYearEndClose_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journal_entrys"`)).
		WithArgs(1, yearEndCloseOrigin, 7).
		WillReturnError(errors.New("db down"))

	reports := NewReportDAO(db)

	_, err := reports.HasYearEndClose(ctx, 1, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
