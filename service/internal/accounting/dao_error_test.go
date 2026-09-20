package accounting

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestBankStatementDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "bank_statements" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	statements := NewBankStatementDAO(db)
	tx := db.Begin()

	_, err := statements.UpdateTx(ctx, tx, &BankStatement{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementLineDAO_ListByStatement_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "bank_statement_lines" WHERE statement_id = $1`)).
		WillReturnError(errors.New("db down"))

	lines := NewBankStatementLineDAO(db)

	_, err := lines.ListByStatement(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementLineDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "bank_statement_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	lines := NewBankStatementLineDAO(db)
	tx := db.Begin()

	_, err := lines.UpdateTx(ctx, tx, &BankStatementLine{Base: model.Base{ID: 11}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetLineDAO_ListByBudget_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "budget_lines" WHERE budget_id = $1`)).
		WillReturnError(errors.New("db down"))

	lines := NewBudgetLineDAO(db)

	_, err := lines.ListByBudget(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetLineDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "budget_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	lines := NewBudgetLineDAO(db)
	tx := db.Begin()

	_, err := lines.UpdateTx(ctx, tx, &BudgetLine{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxRuleDAO_CreateWithMapsTx_PropagatesTaxMapError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rule_tax_maps"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	positions := NewTaxRuleDAO(db)
	tx := db.Begin()

	_, err := positions.CreateWithMapsTx(ctx, tx, &TaxRule{}, []*TaxRuleTaxMap{{SrcTaxID: 5}}, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxRuleDAO_CreateWithMapsTx_PropagatesAccountMapError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rule_tax_maps"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rule_account_maps"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	positions := NewTaxRuleDAO(db)
	tx := db.Begin()

	_, err := positions.CreateWithMapsTx(ctx, tx, &TaxRule{}, []*TaxRuleTaxMap{{SrcTaxID: 5}}, []*TaxRuleAccountMap{{SrcAccountID: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxRuleTaxMapDAO_ListByPosition_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_rule_tax_maps" WHERE tax_rule_id = $1`)).
		WillReturnError(errors.New("db down"))

	maps := NewTaxRuleTaxMapDAO(db)

	_, err := maps.ListByPosition(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxRuleAccountMapDAO_ListByPosition_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_rule_account_maps" WHERE tax_rule_id = $1`)).
		WillReturnError(errors.New("db down"))

	maps := NewTaxRuleAccountMapDAO(db)

	_, err := maps.ListByPosition(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxReturnDAO_FindByPeriod_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "tax_returns" WHERE period_id = \$1 ORDER BY "tax_returns"\."id" LIMIT \$2`).
		WithArgs(3, 1).
		WillReturnError(errors.New("db down"))

	rows := NewTaxReturnDAO(db)

	_, err := rows.FindByPeriod(ctx, 3)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxReturnDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_returns" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	rows := NewTaxReturnDAO(db)
	tx := db.Begin()

	_, err := rows.UpdateTx(ctx, tx, &TaxReturn{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_ListByMovement_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journal_lines" WHERE entry_id = $1`)).
		WillReturnError(errors.New("db down"))

	lines := NewJournalLineDAO(db)

	_, err := lines.ListByMovement(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "journal_lines".*`).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "journal_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	lines := NewJournalLineDAO(db)
	tx := db.Begin()

	_, err := lines.UpdateTx(ctx, tx, &JournalLine{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_BalanceByOriginAndAccount_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(debit - credit\), 0\) FROM "journal_lines" JOIN journal_entrys ON`).
		WillReturnError(errors.New("db down"))

	lines := NewJournalLineDAO(db)

	_, err := lines.BalanceByOriginAndAccount(ctx, "invoice", 1, 100)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "deferred_schedules"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	schedules := NewDeferredScheduleDAO(db)
	tx := db.Begin()

	_, err := schedules.CreateTx(ctx, tx, &DeferredSchedule{})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "deferred_schedules" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	schedules := NewDeferredScheduleDAO(db)
	tx := db.Begin()

	_, err := schedules.UpdateTx(ctx, tx, &DeferredSchedule{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_ListBySchedule_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "deferred_schedule_lines" WHERE schedule_id = $1`)).
		WillReturnError(errors.New("db down"))

	lines := NewDeferredScheduleLineDAO(db)

	_, err := lines.ListBySchedule(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "deferred_schedule_lines"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	lines := NewDeferredScheduleLineDAO(db)
	tx := db.Begin()

	_, err := lines.CreateTx(ctx, tx, &DeferredScheduleLine{})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestDeferredScheduleLineDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "deferred_schedule_lines" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	lines := NewDeferredScheduleLineDAO(db)
	tx := db.Begin()

	_, err := lines.UpdateTx(ctx, tx, &DeferredScheduleLine{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "invoices" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	invoices := NewInvoiceDAO(db)
	tx := db.Begin()

	_, err := invoices.UpdateTx(ctx, tx, &Invoice{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_CreateWithLinesTx_PropagatesTaxError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoices"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoice_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "invoice_taxes"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	invoices := NewInvoiceDAO(db)
	tx := db.Begin()

	_, err := invoices.CreateWithLinesTx(ctx, tx, &Invoice{}, []*InvoiceLine{{Sequence: 10, Qty: 1, UnitPrice: amount.FromInt64(100)}}, []*InvoiceTax{{Amount: amount.FromInt64(10)}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceLineDAO_ListByInvoice_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "invoice_lines" WHERE invoice_id = $1`)).
		WillReturnError(errors.New("db down"))

	lines := NewInvoiceLineDAO(db)

	_, err := lines.ListByInvoice(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceTaxDAO_ListByInvoice_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "invoice_taxes" WHERE invoice_id = $1`)).
		WillReturnError(errors.New("db down"))

	taxes := NewInvoiceTaxDAO(db)

	_, err := taxes.ListByInvoice(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindByOrigin_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "invoices" WHERE origin_invoice_id = \$1 ORDER BY "invoices"\."id" LIMIT \$2`).
		WithArgs(5, 1).
		WillReturnError(errors.New("db down"))

	invoices := NewInvoiceDAO(db)

	_, err := invoices.FindByOrigin(ctx, 5)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_ListOverdue_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "invoices" WHERE state = $1 AND payment_state <> $2 AND due_date IS NOT NULL AND due_date < $3`)).
		WillReturnError(errors.New("db down"))

	invoices := NewInvoiceDAO(db)

	_, err := invoices.ListOverdue(ctx, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentDAO_CreateWithAllocationsTx_PropagatesPaymentError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "payments"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	payments := NewPaymentDAO(db)
	tx := db.Begin()

	_, err := payments.CreateWithAllocationsTx(ctx, tx, &Payment{}, []*PaymentAllocation{{InvoiceID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentAllocationDAO_ListByPayment_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_allocations" WHERE payment_id = $1`)).
		WillReturnError(errors.New("db down"))

	allocations := NewPaymentAllocationDAO(db)

	_, err := allocations.ListByPayment(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPaymentAllocationDAO_ListByInvoice_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "payment_allocations" WHERE invoice_id = $1`)).
		WillReturnError(errors.New("db down"))

	allocations := NewPaymentAllocationDAO(db)

	_, err := allocations.ListByInvoice(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountFullReconcileDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "account_full_reconciles"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	reconciles := NewAccountFullReconcileDAO(db)
	tx := db.Begin()

	_, err := reconciles.CreateTx(ctx, tx, &AccountFullReconcile{})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountPartialReconcileDAO_CreateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "account_partial_reconciles"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	reconciles := NewAccountPartialReconcileDAO(db)
	tx := db.Begin()

	_, err := reconciles.CreateTx(ctx, tx, &AccountPartialReconcile{})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountPartialReconcileDAO_UpdateTx_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "account_partial_reconciles" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()

	reconciles := NewAccountPartialReconcileDAO(db)
	tx := db.Begin()

	_, err := reconciles.UpdateTx(ctx, tx, &AccountPartialReconcile{Base: model.Base{ID: 1}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReminderActionDAO_FindByInvoiceLevel_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "reminder_actions" WHERE invoice_id = $1 AND level_id = $2`)).
		WithArgs(1, 2).
		WillReturnError(errors.New("db down"))

	actions := NewReminderActionDAO(db)

	_, err := actions.FindByInvoiceLevel(ctx, 1, 2)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestReminderActionDAO_ListByInvoice_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "reminder_actions" WHERE invoice_id = $1`)).
		WillReturnError(errors.New("db down"))

	actions := NewReminderActionDAO(db)

	_, err := actions.ListByInvoice(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBankStatementDAO_CreateWithLinesTx_PropagatesStatementError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "bank_statements"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	statements := NewBankStatementDAO(db)
	tx := db.Begin()

	_, err := statements.CreateWithLinesTx(ctx, tx, &BankStatement{}, []*BankStatementLine{{Amount: amount.FromFloat64(100)}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetDAO_CreateWithLinesTx_PropagatesBudgetError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "budgets"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	budgets := NewBudgetDAO(db)
	tx := db.Begin()

	_, err := budgets.CreateWithLinesTx(ctx, tx, &Budget{}, []*BudgetLine{{AccountID: 4100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindBySaleOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT "invoices"\."id","invoices"\."created_at","invoices"\."updated_at","invoices"\."deleted_at","invoices"\."organization_id","invoices"\."entry_id","invoices"\."type","invoices"\."contact_id","invoices"\."name","invoices"\."reference","invoices"\."invoice_date","invoices"\."due_date","invoices"\."currency_code","invoices"\."journal_id","invoices"\."payment_term_id","invoices"\."state","invoices"\."payment_state","invoices"\."amount_untaxed","invoices"\."amount_tax","invoices"\."amount_total","invoices"\."amount_residual","invoices"\."tax_rule","invoices"\."origin_invoice_id" FROM "invoices" JOIN invoice_lines ON invoice_lines\.invoice_id = invoices\.id JOIN sale_order_lines ON sale_order_lines\.id = invoice_lines\.sale_line_id WHERE sale_order_lines\.order_id = \$1 AND invoices\.state = \$2 ORDER BY invoices\.id DESC,"invoices"\."id" LIMIT \$3`).
		WithArgs(7, InvoiceStatePosted, 1).
		WillReturnError(errors.New("db down"))

	invoices := NewInvoiceDAO(db)

	_, err := invoices.FindBySaleOrder(ctx, 7)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestInvoiceDAO_FindByPurchaseOrder_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT "invoices"\."id","invoices"\."created_at","invoices"\."updated_at","invoices"\."deleted_at","invoices"\."organization_id","invoices"\."entry_id","invoices"\."type","invoices"\."contact_id","invoices"\."name","invoices"\."reference","invoices"\."invoice_date","invoices"\."due_date","invoices"\."currency_code","invoices"\."journal_id","invoices"\."payment_term_id","invoices"\."state","invoices"\."payment_state","invoices"\."amount_untaxed","invoices"\."amount_tax","invoices"\."amount_total","invoices"\."amount_residual","invoices"\."tax_rule","invoices"\."origin_invoice_id" FROM "invoices" JOIN invoice_lines ON invoice_lines\.invoice_id = invoices\.id JOIN purchase_order_lines ON purchase_order_lines\.id = invoice_lines\.purchase_line_id WHERE purchase_order_lines\.order_id = \$1 AND invoices\.state = \$2 ORDER BY invoices\.id DESC,"invoices"\."id" LIMIT \$3`).
		WithArgs(8, InvoiceStatePosted, 1).
		WillReturnError(errors.New("db down"))

	invoices := NewInvoiceDAO(db)

	_, err := invoices.FindByPurchaseOrder(ctx, 8)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
