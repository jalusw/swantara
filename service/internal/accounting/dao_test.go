package accounting

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestJournalEntryDAO_CreateWithLinesTx_CreatesMoveAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journal_entrys"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journal_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11).AddRow(12))
	mock.ExpectCommit()

	movements := NewJournalEntryDAO(db)
	tx := db.Begin()

	entry := &JournalEntry{OrganizationID: 1, State: EntryStatePosted, Date: time.Now()}
	lines := []*JournalLine{{AccountID: 100, Debit: amount.FromInt64(50)}, {AccountID: 200, Credit: amount.FromInt64(50)}}

	created, err := movements.CreateWithLinesTx(ctx, tx, entry, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("entry id = %d, want 1", created.ID)
	}
	if lines[0].EntryID != 1 || lines[1].EntryID != 1 {
		t.Errorf("lines entry_id not set: %+v", lines)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalEntryDAO_CreateWithLinesTx_PropagatesMoveError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journal_entrys"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	movements := NewJournalEntryDAO(db)
	tx := db.Begin()

	_, err := movements.CreateWithLinesTx(ctx, tx, &JournalEntry{}, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalEntryDAO_CreateWithLines_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journal_entrys"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	movements := NewJournalEntryDAO(db)

	_, err := movements.CreateWithLines(ctx, &JournalEntry{}, nil)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalEntryDAO_FindByReversed_ReturnsMove(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "journal_entrys" WHERE reversed_entry_id = \$1 LIMIT \$2`).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(9))

	movements := NewJournalEntryDAO(db)

	found, err := movements.FindByReversed(ctx, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.ID != 9 {
		t.Errorf("found = %+v, want entry 9", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_ListByMovement_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journal_lines" WHERE entry_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journal_lines" WHERE entry_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(1, 100).AddRow(2, 200))

	lines := NewJournalLineDAO(db)

	items, err := lines.ListByMovement(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].AccountID != 100 {
		t.Errorf("items = %+v, want two lines", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_ListUnreconciledByAccount_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journal_lines" WHERE account_id = $1 AND reconciled = $2`)).
		WithArgs(100, false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(1, 100))

	lines := NewJournalLineDAO(db)

	items, err := lines.ListUnreconciledByAccount(ctx, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].AccountID != 100 {
		t.Errorf("items = %+v, want one line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_ListUnreconciledByAccount_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journal_lines" WHERE account_id = $1 AND reconciled = $2`)).
		WillReturnError(errors.New("db down"))

	lines := NewJournalLineDAO(db)

	_, err := lines.ListUnreconciledByAccount(ctx, 100)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_BalanceByOriginAndAccount_ReturnsBalance(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(debit - credit), 0) FROM "journal_lines" JOIN journal_entrys ON journal_entrys.id = journal_lines.entry_id WHERE journal_entrys.origin_type = $1 AND journal_entrys.origin_id = $2 AND journal_lines.account_id = $3`)).
		WithArgs("invoice", 7, 100).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(250))

	lines := NewJournalLineDAO(db)

	balance, err := lines.BalanceByOriginAndAccount(ctx, "invoice", 7, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 250 {
		t.Errorf("balance = %v, want 250", balance)
	}

	query.AssertDBMockDone(t, mock)
}

func TestJournalLineDAO_UpdateTx_SavesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT .* FROM "journal_lines".*`).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "journal_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	lines := NewJournalLineDAO(db)
	tx := db.Begin()

	line := &JournalLine{Base: model.Base{ID: 1}, Reconciled: true}
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

func TestAccountBalanceDAO_ListBalancesByPeriod_ReturnsBalances(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT journal_lines\.account_id AS account_id, COALESCE\(SUM\(journal_lines\.debit\), 0\) AS debit, COALESCE\(SUM\(journal_lines\.credit\), 0\) AS credit FROM "journal_lines" JOIN journal_entrys ON journal_entrys\.id = journal_lines\.entry_id WHERE`).
		WithArgs(1, EntryStatePosted, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "debit", "credit"}).AddRow(100, 500, 0))

	balances := NewAccountBalanceDAO(db)

	items, err := balances.ListBalancesByPeriod(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].AccountID != 100 || items[0].Debit != 500 {
		t.Errorf("items = %+v, want one balance", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestAccountBalanceDAO_ListBalancesByPeriod_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT journal_lines.account_id`)).
		WillReturnError(errors.New("db down"))

	balances := NewAccountBalanceDAO(db)

	_, err := balances.ListBalancesByPeriod(ctx, 1, time.Now().AddDate(0, -1, 0), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetDAO_CreateWithLinesTx_CreatesBudgetAndLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "budgets"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "budget_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "budget_lines"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))
	mock.ExpectCommit()

	budgets := NewBudgetDAO(db)
	tx := db.Begin()

	budget := &Budget{State: BudgetStateDraft}
	lines := []*BudgetLine{{AccountID: 100, PlannedAmount: 500}, {AccountID: 200, PlannedAmount: 300}}

	created, err := budgets.CreateWithLinesTx(ctx, tx, budget, lines)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("budget id = %d, want 1", created.ID)
	}
	if lines[0].BudgetID != 1 || lines[1].BudgetID != 1 {
		t.Errorf("lines budget_id not set: %+v", lines)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetDAO_CreateWithLinesTx_PropagatesLineError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "budgets"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "budget_lines"`)).
		WillReturnError(errors.New("line insert failed"))
	mock.ExpectRollback()

	budgets := NewBudgetDAO(db)
	tx := db.Begin()

	_, err := budgets.CreateWithLinesTx(ctx, tx, &Budget{}, []*BudgetLine{{AccountID: 100}})
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetLineDAO_ListByBudget_ReturnsLines(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "budget_lines" WHERE budget_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "budget_lines" WHERE budget_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "account_id"}).AddRow(11, 100))

	lines := NewBudgetLineDAO(db)

	items, err := lines.ListByBudget(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].AccountID != 100 {
		t.Errorf("items = %+v, want one line", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetLineDAO_UpdateTx_SavesLine(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "budget_lines" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	lines := NewBudgetLineDAO(db)
	tx := db.Begin()

	line := &BudgetLine{Base: model.Base{ID: 11}, PracticalAmount: 400}
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

func TestBudgetLineDAO_SumPracticalByBudget_ReturnsTotal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(practical_amount), 0) FROM "budget_lines" WHERE budget_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(800))

	lines := NewBudgetLineDAO(db)

	total, err := lines.SumPracticalByBudget(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 800 {
		t.Errorf("total = %v, want 800", total)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetLineDAO_SumPracticalByBudget_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COALESCE(SUM(practical_amount), 0)`)).
		WillReturnError(errors.New("db down"))

	lines := NewBudgetLineDAO(db)

	_, err := lines.SumPracticalByBudget(ctx, 1)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxRuleDAO_CreateWithMapsTx_CreatesPositionAndMaps(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rules"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rule_tax_maps"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_rule_account_maps"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(21))
	mock.ExpectCommit()

	positions := NewTaxRuleDAO(db)
	tx := db.Begin()

	position := &TaxRule{Name: helper.Ptr("EU")}
	taxMaps := []*TaxRuleTaxMap{{SrcTaxID: 5, DestTaxID: helper.Ptr(uint64(6))}}
	accountMaps := []*TaxRuleAccountMap{{SrcAccountID: 100, DestAccountID: 200}}

	created, err := positions.CreateWithMapsTx(ctx, tx, position, taxMaps, accountMaps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("position id = %d, want 1", created.ID)
	}
	if taxMaps[0].TaxRuleID != 1 || accountMaps[0].TaxRuleID != 1 {
		t.Errorf("maps tax_rule_id not set: %+v %+v", taxMaps, accountMaps)
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxRuleTaxMapDAO_ListByPosition_ReturnsMaps(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_rule_tax_maps" WHERE tax_rule_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_rule_tax_maps" WHERE tax_rule_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "src_tax_id"}).AddRow(11, 5))

	maps := NewTaxRuleTaxMapDAO(db)

	items, err := maps.ListByPosition(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].SrcTaxID != 5 {
		t.Errorf("items = %+v, want one map", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxRuleAccountMapDAO_ListByPosition_ReturnsMaps(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_rule_account_maps" WHERE tax_rule_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_rule_account_maps" WHERE tax_rule_id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "src_account_id"}).AddRow(21, 100))

	maps := NewTaxRuleAccountMapDAO(db)

	items, err := maps.ListByPosition(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].SrcAccountID != 100 {
		t.Errorf("items = %+v, want one map", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestWithholdingTaxDAO_ListActiveByScope_ReturnsTaxes(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "withholding_taxes" WHERE organization_id = $1 AND scope = $2 AND active = $3`)).
		WithArgs(1, WithholdingScopeSale, true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "rate_pct"}).AddRow(1, 2.0))

	taxes := NewWithholdingTaxDAO(db)

	items, err := taxes.ListActiveByScope(ctx, 1, WithholdingScopeSale)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].RatePct != 2.0 {
		t.Errorf("items = %+v, want one tax", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestWithholdingTaxDAO_ListActiveByScope_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "withholding_taxes"`)).
		WillReturnError(errors.New("db down"))

	taxes := NewWithholdingTaxDAO(db)

	_, err := taxes.ListActiveByScope(ctx, 1, WithholdingScopeSale)
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxReturnDAO_FindByPeriod_ReturnsTaxReturn(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "tax_returns" WHERE period_id = \$1 ORDER BY "tax_returns"\."id" LIMIT \$2`).
		WithArgs(3, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "period_id"}).AddRow(1, 3))

	rows := NewTaxReturnDAO(db)

	found, err := rows.FindByPeriod(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.PeriodID != 3 {
		t.Errorf("found = %+v, want tax return 3", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxReturnDAO_FindByPeriod_ReturnsNilWhenNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT \* FROM "tax_returns" WHERE period_id = \$1 ORDER BY "tax_returns"\."id" LIMIT \$2`).
		WithArgs(3, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	rows := NewTaxReturnDAO(db)

	found, err := rows.FindByPeriod(ctx, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found != nil {
		t.Errorf("found = %+v, want nil", found)
	}

	query.AssertDBMockDone(t, mock)
}

func TestTaxReturnDAO_UpdateTx_SavesTaxReturn(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_returns" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rows := NewTaxReturnDAO(db)
	tx := db.Begin()

	taxReturn := &TaxReturn{Base: model.Base{ID: 1}, State: TaxReturnStateFiled}
	updated, err := rows.UpdateTx(ctx, tx, taxReturn)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != taxReturn {
		t.Errorf("UpdateTx returned a different tax return")
	}
	if err := tx.Commit().Error; err != nil {
		t.Fatalf("commit: %v", err)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetQueryDAO_SumPracticalByPeriod_ReturnsTotal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(journal_lines\.debit - journal_lines\.credit\), 0\) FROM "journal_lines" JOIN journal_entrys ON journal_entrys\.id = journal_lines\.entry_id WHERE`).
		WithArgs(1, EntryStatePosted, 100, start, end).
		WillReturnRows(sqlmock.NewRows([]string{"total"}).AddRow(1200))

	queries := NewBudgetQueryDAO(db)

	total, err := queries.SumPracticalByPeriod(ctx, 1, 100, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1200 {
		t.Errorf("total = %v, want 1200", total)
	}

	query.AssertDBMockDone(t, mock)
}

func TestBudgetQueryDAO_SumPracticalByPeriod_Error(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(journal_lines\.debit - journal_lines\.credit\), 0\)`).
		WillReturnError(errors.New("db down"))

	queries := NewBudgetQueryDAO(db)

	_, err := queries.SumPracticalByPeriod(ctx, 1, 100, time.Now().AddDate(0, -1, 0), time.Now())
	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}
