//go:build integration

package integration

import (
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestStatements_ReconcileToTrialBalance(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()

	expense, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "6000", Name: "Operating Expenses", Type: "expense", Active: true,
	})
	if err != nil {
		t.Fatalf("create expense account failed: %v", err)
	}
	retained, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "3200", Name: "Retained Earnings", Type: "equity", Active: true,
	})
	if err != nil {
		t.Fatalf("create retained earnings account failed: %v", err)
	}

	date := fx.date
	if _, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID, JournalID: fx.journalID, Date: date, Ref: "STMT-001",
		Description: "revenue",
		Lines: []accounting.PostingLine{
			{AccountID: fx.bankID, Debit: amount.FromFloat64(1000)},
			{AccountID: fx.incomeID, Credit: amount.FromFloat64(1000)},
		},
	}); err != nil {
		t.Fatalf("post revenue failed: %v", err)
	}
	if _, err := fx.postingSvc.Post(ctx, accounting.PostRequest{
		OrganizationID: fx.orgID, JournalID: fx.journalID, Date: date, Ref: "STMT-002",
		Description: "expense",
		Lines: []accounting.PostingLine{
			{AccountID: expense.ID, Debit: amount.FromFloat64(300)},
			{AccountID: fx.bankID, Credit: amount.FromFloat64(300)},
		},
	}); err != nil {
		t.Fatalf("post expense failed: %v", err)
	}

	year, err := dao.NewBase[reference.TaxYear](testDB).Search(ctx, "organization_id", fx.orgID)
	if err != nil {
		t.Fatalf("find tax year failed: %v", err)
	}
	statementSvc := reporting.NewStatementService(
		fx.reportDAO,
		accounting.NewTaxPeriodDAO(testDB),
		accounting.NewTaxPeriodDAO(testDB),
		dao.NewBase[reference.TaxYear](testDB),
		fx.postingSvc,
	)

	pl, err := statementSvc.ProfitAndLoss(ctx, fx.orgID, fx.periodID)
	if err != nil {
		t.Fatalf("profit and loss failed: %v", err)
	}
	if pl.Revenue != 1000 || pl.Expenses != 300 || pl.NetIncome != 700 {
		t.Errorf("pnl = %+v, want revenue 1000 expenses 300 net 700", pl)
	}

	asOf := time.Date(date.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	sheet, err := statementSvc.BalanceSheet(ctx, fx.orgID, asOf)
	if err != nil {
		t.Fatalf("balance sheet failed: %v", err)
	}
	if sheet.TotalAssets != 700 {
		t.Errorf("total assets = %v, want 700", sheet.TotalAssets)
	}
	if sheet.CurrentEarnings != 700 || sheet.TotalEquity != 700 {
		t.Errorf("earnings = %v equity = %v, want 700 and 700", sheet.CurrentEarnings, sheet.TotalEquity)
	}
	if sheet.TotalAssets != sheet.TotalLiabilities+sheet.TotalEquity {
		t.Errorf("balance sheet unbalanced: assets %v vs liabilities+equity %v", sheet.TotalAssets, sheet.TotalLiabilities+sheet.TotalEquity)
	}

	flow, err := statementSvc.CashFlow(ctx, fx.orgID, time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC), asOf)
	if err != nil {
		t.Fatalf("cash flow failed: %v", err)
	}
	if flow.Operating != 700 || flow.ClosingCash != 700 {
		t.Errorf("cash flow = %+v, want operating 700 closing 700", flow)
	}

	first, err := statementSvc.YearEndRoll(ctx, reporting.YearEndRollRequest{
		OrganizationID: fx.orgID, TaxYearID: year.ID, JournalID: fx.journalID, RetainedEarningsAccountID: retained.ID,
	})
	if err != nil {
		t.Fatalf("year-end roll failed: %v", err)
	}
	if first.NetIncome != 700 || first.ZeroedAccounts != 2 {
		t.Errorf("roll = %+v, want net 700 zeroed 2", first)
	}

	trial, err := fx.reportDAO.TrialBalance(ctx, fx.orgID, time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC), asOf)
	if err != nil {
		t.Fatalf("trial balance failed: %v", err)
	}
	for _, row := range trial {
		switch row.AccountID {
		case fx.incomeID:
			if row.ClosingCredit-row.ClosingDebit != 0 {
				t.Errorf("income account not zeroed: %+v", row)
			}
		case retained.ID:
			if row.ClosingCredit-row.ClosingDebit != 700 {
				t.Errorf("retained earnings = %+v, want 700 credit", row)
			}
		}
	}

	if _, err := statementSvc.YearEndRoll(ctx, reporting.YearEndRollRequest{
		OrganizationID: fx.orgID, TaxYearID: year.ID, JournalID: fx.journalID, RetainedEarningsAccountID: retained.ID,
	}); !errors.Is(err, reporting.ErrYearEndAlreadyClosed) {
		t.Fatalf("second roll err = %v, want ErrYearEndAlreadyClosed", err)
	}
}
