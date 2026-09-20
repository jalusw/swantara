package accounting

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestInvoiceDAO_UpdateTx_SavesInvoice(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "invoices" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	invoices := NewInvoiceDAO(db)
	tx := db.Begin()

	invoice := &Invoice{Base: model.Base{ID: 1}, State: InvoiceStatePosted}
	updated, err := invoices.UpdateTx(ctx, tx, invoice)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != invoice {
		t.Errorf("UpdateTx returned a different invoice")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_ListOpenByContact_ReturnsInvoices(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invoices" WHERE contact_id = $1 AND state = $2 AND payment_state NOT IN ($3,$4)`)).
		WithArgs(10, InvoiceStatePosted, PaymentStatePaid, PaymentStateBadDebt).
		WillReturnRows(sqlmock.NewRows([]string{"id", "contact_id"}).AddRow(1, 10).AddRow(2, 10))

	invoices := NewInvoiceDAO(db)

	items, err := invoices.ListOpenByContact(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].ContactID != 10 {
		t.Errorf("items = %+v, want two invoices", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_ListOpenByContact_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invoices"`)).
		WillReturnError(errors.New("db down"))

	invoices := NewInvoiceDAO(db)

	_, err := invoices.ListOpenByContact(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindByOrigin_ReturnsInvoice(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE origin_invoice_id = \$1 ORDER BY "invoices"\."id" LIMIT \$2`).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "origin_invoice_id"}).AddRow(9, 5))

	invoices := NewInvoiceDAO(db)

	found, err := invoices.FindByOrigin(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.ID != 9 {
		t.Errorf("found = %+v, want invoice 9", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindByOrigin_ReturnsNilWhenNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE origin_invoice_id = \$1 ORDER BY "invoices"\."id" LIMIT \$2`).
		WithArgs(5, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	invoices := NewInvoiceDAO(db)

	found, err := invoices.FindByOrigin(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("found = %+v, want nil", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindBySaleOrder_ReturnsInvoice(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT "invoices"\."id","invoices"\."created_at","invoices"\."updated_at","invoices"\."deleted_at","invoices"\."organization_id","invoices"\."entry_id","invoices"\."type","invoices"\."contact_id","invoices"\."name","invoices"\."reference","invoices"\."invoice_date","invoices"\."due_date","invoices"\."currency_code","invoices"\."journal_id","invoices"\."payment_term_id","invoices"\."tax_rule_id","invoices"\."state","invoices"\."payment_state","invoices"\."amount_untaxed","invoices"\."amount_tax","invoices"\."amount_total","invoices"\."amount_residual","invoices"\."tax_rule","invoices"\."origin_invoice_id" FROM "invoices" JOIN invoice_lines ON invoice_lines\.invoice_id = invoices\.id JOIN sale_order_lines ON sale_order_lines\.id = invoice_lines\.sale_line_id WHERE sale_order_lines\.order_id = \$1 AND invoices\.state = \$2 ORDER BY invoices\.id DESC,"invoices"\."id" LIMIT \$3`).
		WithArgs(7, InvoiceStatePosted, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))

	invoices := NewInvoiceDAO(db)

	found, err := invoices.FindBySaleOrder(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.ID != 3 {
		t.Errorf("found = %+v, want invoice 3", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindBySaleOrder_ReturnsNilWhenNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT "invoices"\."id","invoices"\."created_at","invoices"\."updated_at","invoices"\."deleted_at","invoices"\."organization_id","invoices"\."entry_id","invoices"\."type","invoices"\."contact_id","invoices"\."name","invoices"\."reference","invoices"\."invoice_date","invoices"\."due_date","invoices"\."currency_code","invoices"\."journal_id","invoices"\."payment_term_id","invoices"\."tax_rule_id","invoices"\."state","invoices"\."payment_state","invoices"\."amount_untaxed","invoices"\."amount_tax","invoices"\."amount_total","invoices"\."amount_residual","invoices"\."tax_rule","invoices"\."origin_invoice_id" FROM "invoices" JOIN invoice_lines ON invoice_lines\.invoice_id = invoices\.id JOIN sale_order_lines ON sale_order_lines\.id = invoice_lines\.sale_line_id WHERE sale_order_lines\.order_id = \$1 AND invoices\.state = \$2 ORDER BY invoices\.id DESC,"invoices"\."id" LIMIT \$3`).
		WithArgs(7, InvoiceStatePosted, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	invoices := NewInvoiceDAO(db)

	found, err := invoices.FindBySaleOrder(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("found = %+v, want nil", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindByPurchaseOrder_ReturnsInvoice(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT "invoices"\."id","invoices"\."created_at","invoices"\."updated_at","invoices"\."deleted_at","invoices"\."organization_id","invoices"\."entry_id","invoices"\."type","invoices"\."contact_id","invoices"\."name","invoices"\."reference","invoices"\."invoice_date","invoices"\."due_date","invoices"\."currency_code","invoices"\."journal_id","invoices"\."payment_term_id","invoices"\."tax_rule_id","invoices"\."state","invoices"\."payment_state","invoices"\."amount_untaxed","invoices"\."amount_tax","invoices"\."amount_total","invoices"\."amount_residual","invoices"\."tax_rule","invoices"\."origin_invoice_id" FROM "invoices" JOIN invoice_lines ON invoice_lines\.invoice_id = invoices\.id JOIN purchase_order_lines ON purchase_order_lines\.id = invoice_lines\.purchase_line_id WHERE purchase_order_lines\.order_id = \$1 AND invoices\.state = \$2 ORDER BY invoices\.id DESC,"invoices"\."id" LIMIT \$3`).
		WithArgs(8, InvoiceStatePosted, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))

	invoices := NewInvoiceDAO(db)

	found, err := invoices.FindByPurchaseOrder(ctx, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.ID != 4 {
		t.Errorf("found = %+v, want invoice 4", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindByPurchaseOrder_ReturnsNilWhenNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT "invoices"\."id","invoices"\."created_at","invoices"\."updated_at","invoices"\."deleted_at","invoices"\."organization_id","invoices"\."entry_id","invoices"\."type","invoices"\."contact_id","invoices"\."name","invoices"\."reference","invoices"\."invoice_date","invoices"\."due_date","invoices"\."currency_code","invoices"\."journal_id","invoices"\."payment_term_id","invoices"\."tax_rule_id","invoices"\."state","invoices"\."payment_state","invoices"\."amount_untaxed","invoices"\."amount_tax","invoices"\."amount_total","invoices"\."amount_residual","invoices"\."tax_rule","invoices"\."origin_invoice_id" FROM "invoices" JOIN invoice_lines ON invoice_lines\.invoice_id = invoices\.id JOIN purchase_order_lines ON purchase_order_lines\.id = invoice_lines\.purchase_line_id WHERE purchase_order_lines\.order_id = \$1 AND invoices\.state = \$2 ORDER BY invoices\.id DESC,"invoices"\."id" LIMIT \$3`).
		WithArgs(8, InvoiceStatePosted, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	invoices := NewInvoiceDAO(db)

	found, err := invoices.FindByPurchaseOrder(ctx, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("found = %+v, want nil", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_ListOverdue_ReturnsInvoices(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	asOf := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invoices" WHERE state = $1 AND payment_state NOT IN ($2,$3) AND due_date IS NOT NULL AND due_date < $4`)).
		WithArgs(InvoiceStatePosted, PaymentStatePaid, PaymentStateBadDebt, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	invoices := NewInvoiceDAO(db)

	items, err := invoices.ListOverdue(ctx, &asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("items = %+v, want one invoice", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_ListOverdue_UsesNowWhenAsOfNil(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invoices" WHERE state = $1 AND payment_state NOT IN ($2,$3) AND due_date IS NOT NULL AND due_date < $4`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	invoices := NewInvoiceDAO(db)

	items, err := invoices.ListOverdue(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("items = %+v, want none", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_CreateWithLinesTx_CreatesInvoiceLinesAndTaxes(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoices"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoice_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoice_taxes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
	mock.ExpectCommit()

	invoices := NewInvoiceDAO(db)
	tx := db.Begin()

	invoice := &Invoice{State: InvoiceStateDraft}
	lines := []*InvoiceLine{{ItemID: helper.Ptr(uint64(100))}}
	taxes := []*InvoiceTax{{Amount: amount.FromInt64(10)}}

	created, err := invoices.CreateWithLinesTx(ctx, tx, invoice, lines, taxes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("invoice id = %d, want 1", created.ID)
	}
	if lines[0].InvoiceID != 1 || taxes[0].InvoiceID != 1 {
		t.Errorf("children invoice_id not set: %+v %+v", lines, taxes)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_CreateWithLinesTx_PropagatesLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoices"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoice_lines"`)).
		WillReturnError(errors.New("line insert failed"))
	mock.ExpectRollback()

	invoices := NewInvoiceDAO(db)
	tx := db.Begin()

	_, err := invoices.CreateWithLinesTx(ctx, tx, &Invoice{}, []*InvoiceLine{{ItemID: helper.Ptr(uint64(100))}}, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceLineDAO_ListByInvoice_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "invoice_lines" WHERE invoice_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invoice_lines" WHERE invoice_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "invoice_id"}).AddRow(11, 1))

	lines := NewInvoiceLineDAO(db)

	items, err := lines.ListByInvoice(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].InvoiceID != 1 {
		t.Errorf("items = %+v, want one line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceTaxDAO_ListByInvoice_ReturnsTaxes(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "invoice_taxes" WHERE invoice_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invoice_taxes" WHERE invoice_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount"}).AddRow(21, 10))

	taxes := NewInvoiceTaxDAO(db)

	items, err := taxes.ListByInvoice(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Amount.Float64() != 10 {
		t.Errorf("items = %+v, want one tax", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceTaxDAO_SumTaxByPeriod_ReturnsTotal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(invoice_taxes\.amount\), 0\) FROM "invoice_taxes" JOIN invoices ON invoices\.id = invoice_taxes\.invoice_id WHERE`).
		WithArgs(1, TaxReturnTypeSale, InvoiceStatePosted, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(25))

	taxes := NewInvoiceTaxDAO(db)

	total, err := taxes.SumTaxByPeriod(ctx, 1, TaxReturnTypeSale, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 25 {
		t.Errorf("total = %v, want 25", total)
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceTaxDAO_SumTaxByPeriod_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(invoice_taxes\.amount\), 0\)`).
		WillReturnError(errors.New("db down"))

	taxes := NewInvoiceTaxDAO(db)

	_, err := taxes.SumTaxByPeriod(ctx, 1, TaxReturnTypeSale, time.Now(), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentDAO_CreateWithAllocationsTx_CreatesPaymentAndAllocations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_allocations"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	payments := NewPaymentDAO(db)
	tx := db.Begin()

	payment := &Payment{State: PaymentStateDraft}
	allocations := []*PaymentAllocation{{InvoiceID: 7, Amount: 100}}

	created, err := payments.CreateWithAllocationsTx(ctx, tx, payment, allocations)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("payment id = %d, want 1", created.ID)
	}
	if allocations[0].PaymentID != 1 {
		t.Errorf("allocation payment_id not set: %+v", allocations)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentDAO_CreateWithAllocationsTx_PropagatesAllocationError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payments"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payment_allocations"`)).
		WillReturnError(errors.New("allocation insert failed"))
	mock.ExpectRollback()

	payments := NewPaymentDAO(db)
	tx := db.Begin()

	_, err := payments.CreateWithAllocationsTx(ctx, tx, &Payment{}, []*PaymentAllocation{{InvoiceID: 7}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentDAO_ListPostedByContact_ReturnsPayments(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payments" WHERE contact_id = $1 AND state = $2`)).
		WithArgs(10, PaymentStatePosted).
		WillReturnRows(sqlmock.NewRows([]string{"id", "contact_id"}).AddRow(1, 10))

	payments := NewPaymentDAO(db)

	items, err := payments.ListPostedByContact(ctx, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ContactID != 10 {
		t.Errorf("items = %+v, want one payment", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentDAO_ListPostedByContact_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payments"`)).
		WillReturnError(errors.New("db down"))

	payments := NewPaymentDAO(db)

	_, err := payments.ListPostedByContact(ctx, 10)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentAllocationDAO_ListByPayment_ReturnsAllocations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_allocations" WHERE payment_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_allocations" WHERE payment_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount"}).AddRow(11, 100))

	allocations := NewPaymentAllocationDAO(db)

	items, err := allocations.ListByPayment(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Amount != 100 {
		t.Errorf("items = %+v, want one allocation", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentAllocationDAO_ListByInvoice_ReturnsAllocations(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_allocations" WHERE invoice_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payment_allocations" WHERE invoice_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount"}).AddRow(11, 100))

	allocations := NewPaymentAllocationDAO(db)

	items, err := allocations.ListByInvoice(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Amount != 100 {
		t.Errorf("items = %+v, want one allocation", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementDAO_CreateWithLinesTx_CreatesStatementAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bank_statements"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bank_statement_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bank_statement_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))
	mock.ExpectCommit()

	statements := NewBankStatementDAO(db)
	tx := db.Begin()

	statement := &BankStatement{State: BankStatementStateDraft}
	lines := []*BankStatementLine{{Amount: amount.FromFloat64(100)}, {Amount: amount.FromFloat64(200)}}

	created, err := statements.CreateWithLinesTx(ctx, tx, statement, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("statement id = %d, want 1", created.ID)
	}
	if lines[0].StatementID != 1 || lines[1].StatementID != 1 {
		t.Errorf("lines statement_id not set: %+v", lines)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementDAO_CreateWithLinesTx_PropagatesLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bank_statements"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bank_statement_lines"`)).
		WillReturnError(errors.New("line insert failed"))
	mock.ExpectRollback()

	statements := NewBankStatementDAO(db)
	tx := db.Begin()

	_, err := statements.CreateWithLinesTx(ctx, tx, &BankStatement{}, []*BankStatementLine{{Amount: amount.FromFloat64(100)}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementDAO_UpdateTx_SavesStatement(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "bank_statements" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	statements := NewBankStatementDAO(db)
	tx := db.Begin()

	statement := &BankStatement{Base: model.Base{ID: 1}, State: BankStatementStateOpen}
	updated, err := statements.UpdateTx(ctx, tx, statement)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != statement {
		t.Errorf("UpdateTx returned a different statement")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementLineDAO_ListByStatement_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "bank_statement_lines" WHERE statement_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "bank_statement_lines" WHERE statement_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount"}).AddRow(11, 100))

	lines := NewBankStatementLineDAO(db)

	items, err := lines.ListByStatement(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Amount.Float64() != 100 {
		t.Errorf("items = %+v, want one line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementLineDAO_ListUnreconciledByStatement_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "bank_statement_lines" WHERE statement_id = $1 AND reconciled = $2`)).
		WithArgs(1, false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "amount"}).AddRow(11, 100))

	lines := NewBankStatementLineDAO(db)

	items, err := lines.ListUnreconciledByStatement(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Amount.Float64() != 100 {
		t.Errorf("items = %+v, want one line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementLineDAO_ListUnreconciledByStatement_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "bank_statement_lines"`)).
		WillReturnError(errors.New("db down"))

	lines := NewBankStatementLineDAO(db)

	_, err := lines.ListUnreconciledByStatement(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementLineDAO_UpdateTx_SavesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "bank_statement_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	lines := NewBankStatementLineDAO(db)
	tx := db.Begin()

	line := &BankStatementLine{Base: model.Base{ID: 11}, Reconciled: true}
	updated, err := lines.UpdateTx(ctx, tx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != line {
		t.Errorf("UpdateTx returned a different line")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountFullReconcileDAO_CreateTx_CreatesReconcile(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "account_full_reconciles"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	reconciles := NewAccountFullReconcileDAO(db)
	tx := db.Begin()

	reconcile := &AccountFullReconcile{Name: helper.Ptr("Reconcile 1")}
	created, err := reconciles.CreateTx(ctx, tx, reconcile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("reconcile id = %d, want 1", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountPartialReconcileDAO_CreateTx_CreatesReconcile(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "account_partial_reconciles"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	reconciles := NewAccountPartialReconcileDAO(db)
	tx := db.Begin()

	reconcile := &AccountPartialReconcile{DebitLineID: 1, CreditLineID: 2, Amount: 50}
	created, err := reconciles.CreateTx(ctx, tx, reconcile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("reconcile id = %d, want 1", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountPartialReconcileDAO_UpdateTx_SavesReconcile(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "account_partial_reconciles" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	reconciles := NewAccountPartialReconcileDAO(db)
	tx := db.Begin()

	reconcile := &AccountPartialReconcile{Base: model.Base{ID: 1}, Amount: 60}
	updated, err := reconciles.UpdateTx(ctx, tx, reconcile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != reconcile {
		t.Errorf("UpdateTx returned a different reconcile")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountPartialReconcileDAO_TotalByLine_ReturnsTotal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(amount), 0) FROM "account_partial_reconciles" WHERE debit_line_id = $1 OR credit_line_id = $2`)).
		WithArgs(5, 5).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(120))

	reconciles := NewAccountPartialReconcileDAO(db)

	total, err := reconciles.TotalByLine(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 120 {
		t.Errorf("total = %v, want 120", total)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountPartialReconcileDAO_TotalByLine_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(amount), 0)`)).
		WillReturnError(errors.New("db down"))

	reconciles := NewAccountPartialReconcileDAO(db)

	_, err := reconciles.TotalByLine(ctx, 5)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReminderActionDAO_FindByInvoiceLevel_ReturnsAction(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reminder_actions" WHERE invoice_id = $1 AND level_id = $2 LIMIT $3`)).
		WithArgs(7, 3, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	actions := NewReminderActionDAO(db)

	found, err := actions.FindByInvoiceLevel(ctx, 7, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.ID != 1 {
		t.Errorf("found = %+v, want action 1", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReminderActionDAO_FindByInvoiceLevel_ReturnsNilWhenNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reminder_actions" WHERE invoice_id = $1 AND level_id = $2 LIMIT $3`)).
		WithArgs(7, 3, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	actions := NewReminderActionDAO(db)

	found, err := actions.FindByInvoiceLevel(ctx, 7, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("found = %+v, want nil", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestReminderActionDAO_ListByInvoice_ReturnsActions(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "reminder_actions" WHERE invoice_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reminder_actions" WHERE invoice_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	actions := NewReminderActionDAO(db)

	items, err := actions.ListByInvoice(ctx, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("items = %+v, want one action", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleDAO_ListRunning_ReturnsSchedules(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "deferred_schedules" WHERE state = $1`)).
		WithArgs(DeferredStateRunning).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	schedules := NewDeferredScheduleDAO(db)

	items, err := schedules.ListRunning(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("items = %+v, want one schedule", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleDAO_ListRunning_WithOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "deferred_schedules" WHERE state = $1 AND organization_id = $2`)).
		WithArgs(DeferredStateRunning, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	schedules := NewDeferredScheduleDAO(db)

	items, err := schedules.ListRunning(ctx, helper.Ptr(uint64(1)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Errorf("items = %+v, want one schedule", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleDAO_ListRunning_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "deferred_schedules"`)).
		WillReturnError(errors.New("db down"))

	schedules := NewDeferredScheduleDAO(db)

	_, err := schedules.ListRunning(ctx, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleDAO_CreateTx_CreatesSchedule(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "deferred_schedules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	schedules := NewDeferredScheduleDAO(db)
	tx := db.Begin()

	schedule := &DeferredSchedule{State: DeferredStateDraft}
	created, err := schedules.CreateTx(ctx, tx, schedule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("schedule id = %d, want 1", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleDAO_UpdateTx_SavesSchedule(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "deferred_schedules" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	schedules := NewDeferredScheduleDAO(db)
	tx := db.Begin()

	schedule := &DeferredSchedule{Base: model.Base{ID: 1}, State: DeferredStateRunning}
	updated, err := schedules.UpdateTx(ctx, tx, schedule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != schedule {
		t.Errorf("UpdateTx returned a different schedule")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_ListBySchedule_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "deferred_schedule_lines" WHERE schedule_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "deferred_schedule_lines" WHERE schedule_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sequence"}).AddRow(11, 1))

	lines := NewDeferredScheduleLineDAO(db)

	items, err := lines.ListBySchedule(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Sequence != 1 {
		t.Errorf("items = %+v, want one line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_ListDue_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	asOf := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "deferred_schedule_lines" WHERE schedule_id = $1 AND posted = $2 AND recognition_date IS NOT NULL AND recognition_date <= $3 ORDER BY sequence ASC`)).
		WithArgs(1, false, asOf).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sequence"}).AddRow(11, 1))

	lines := NewDeferredScheduleLineDAO(db)

	items, err := lines.ListDue(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Sequence != 1 {
		t.Errorf("items = %+v, want one line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_ListDue_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "deferred_schedule_lines"`)).
		WillReturnError(errors.New("db down"))

	lines := NewDeferredScheduleLineDAO(db)

	_, err := lines.ListDue(ctx, 1, time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_CreateTx_CreatesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "deferred_schedule_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectCommit()

	lines := NewDeferredScheduleLineDAO(db)
	tx := db.Begin()

	line := &DeferredScheduleLine{ScheduleID: 1, Sequence: 1}
	created, err := lines.CreateTx(ctx, tx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 11 {
		t.Errorf("line id = %d, want 11", created.ID)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_UpdateTx_SavesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "deferred_schedule_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	lines := NewDeferredScheduleLineDAO(db)
	tx := db.Begin()

	line := &DeferredScheduleLine{Base: model.Base{ID: 11}, Posted: true}
	updated, err := lines.UpdateTx(ctx, tx, line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != line {
		t.Errorf("UpdateTx returned a different line")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}
