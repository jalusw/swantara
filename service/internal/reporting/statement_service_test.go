package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestStatementService_ProfitAndLoss_TotalsByAccountType(t *testing.T) {
	ctx := context.Background()
	svc := NewStatementService(
		ReportDAOMock{
			StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]AccountBalanceRow, error) {
				return []AccountBalanceRow{
					{AccountID: 41, Code: "4100", Name: "Sales", AccountType: "income", Balance: -1000},
					{AccountID: 50, Code: "5000", Name: "COGS", AccountType: "cogs", Balance: 600},
					{AccountID: 60, Code: "6000", Name: "OpEx", AccountType: "expense", Balance: 200},
				}, nil
			},
		},
		TaxPeriodFinderMock{},
		TaxPeriodByDateFinderMock{},
		TaxYearFinderMock{},
		StatementPosterMock{},
	)

	statement, err := svc.ProfitAndLoss(ctx, 1, 7)
	if err != nil {
		t.Fatalf("profit and loss failed: %v", err)
	}
	if statement.Revenue != 1000 || statement.COGS != 600 || statement.Expenses != 200 {
		t.Errorf("statement = %+v, want revenue 1000 cogs 600 expenses 200", statement)
	}
	if statement.GrossProfit != 400 || statement.NetIncome != 200 {
		t.Errorf("gross = %v net = %v, want 400 and 200", statement.GrossProfit, statement.NetIncome)
	}
	if len(statement.Rows) != 3 {
		t.Errorf("rows = %d, want 3", len(statement.Rows))
	}
}

func TestStatementService_BalanceSheet_SectionsAndCurrentEarnings(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	svc := NewStatementService(
		ReportDAOMock{
			StatementBalancesFn: func(_ context.Context, _ uint64, start, end time.Time) ([]AccountBalanceRow, error) {
				if start.Equal(epochDate) {
					return []AccountBalanceRow{
						{AccountID: 11, Code: "1100", Name: "Cash", AccountType: "cash", Balance: 1500},
						{AccountID: 15, Code: "1500", Name: "Fixed Assets", AccountType: "fixed_asset", Balance: 8000},
						{AccountID: 15, Code: "1510", Name: "Accum Dep", AccountType: "depreciation", Balance: -2000},
						{AccountID: 20, Code: "2000", Name: "AP", AccountType: "payable", Balance: 500},
						{AccountID: 30, Code: "3000", Name: "Equity", AccountType: "equity", Balance: 7000},
					}, nil
				}
				return nil, nil
			},
		},
		TaxPeriodFinderMock{},
		TaxPeriodByDateFinderMock{},
		TaxYearFinderMock{},
		StatementPosterMock{},
	)

	sheet, err := svc.BalanceSheet(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("balance sheet failed: %v", err)
	}
	if sheet.TotalAssets != 7500 {
		t.Errorf("total assets = %v, want 7500", sheet.TotalAssets)
	}
	if sheet.TotalLiabilities != 500 {
		t.Errorf("total liabilities = %v, want 500", sheet.TotalLiabilities)
	}
	if sheet.TotalEquity != 7000 {
		t.Errorf("total equity = %v, want 7000 (no current earnings)", sheet.TotalEquity)
	}
	if len(sheet.Assets) != 3 || len(sheet.Liabilities) != 1 || len(sheet.Equity) != 1 {
		t.Errorf("sections = assets %d liabilities %d equity %d, want 3/1/1", len(sheet.Assets), len(sheet.Liabilities), len(sheet.Equity))
	}
}

func TestStatementService_CashFlow_ClassifiesAndReconciles(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	svc := NewStatementService(
		ReportDAOMock{
			CashFlowFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]CashFlowSectionRow, error) {
				return []CashFlowSectionRow{
					{Section: "operating", Amount: 300},
					{Section: "investing", Amount: -1500},
					{Section: "financing", Amount: 2000},
				}, nil
			},
			StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]AccountBalanceRow, error) {
				return []AccountBalanceRow{
					{AccountID: 11, Code: "1100", Name: "Cash", AccountType: "cash", Balance: 1000},
					{AccountID: 12, Code: "1110", Name: "Bank", AccountType: "bank", Balance: 500},
				}, nil
			},
		},
		TaxPeriodFinderMock{},
		TaxPeriodByDateFinderMock{},
		TaxYearFinderMock{},
		StatementPosterMock{},
	)

	flow, err := svc.CashFlow(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("cash flow failed: %v", err)
	}
	if flow.Operating != 300 || flow.Investing != -1500 || flow.Financing != 2000 {
		t.Errorf("flow = %+v, want operating 300 investing -1500 financing 2000", flow)
	}
	if flow.OpeningCash != 1500 {
		t.Errorf("opening cash = %v, want 1500", flow.OpeningCash)
	}
	if flow.ClosingCash != 2300 {
		t.Errorf("closing cash = %v, want 2300", flow.ClosingCash)
	}
}

func TestStatementService_YearEndRoll_PostsBalancedAndGuardsDoubleRun(t *testing.T) {
	ctx := context.Background()
	posted := false
	svc := NewStatementService(
		ReportDAOMock{
			StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]AccountBalanceRow, error) {
				return []AccountBalanceRow{
					{AccountID: 41, Code: "4100", Name: "Sales", AccountType: "income", Balance: -1000},
					{AccountID: 50, Code: "5000", Name: "COGS", AccountType: "cogs", Balance: 600},
				}, nil
			},
			HasYearEndCloseFn: func(_ context.Context, _, _ uint64) (bool, error) { return posted, nil },
		},
		TaxPeriodFinderMock{},
		TaxPeriodByDateFinderMock{},
		TaxYearFinderMock{
			FindFn: func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
				return &reference.TaxYear{Name: "FY2026", DateStart: &start, DateEnd: &end}, nil
			},
		},
		StatementPosterMock{
			PostFn: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
				if request.OriginType != "year_end_close" {
					t.Errorf("origin type = %v, want year_end_close", request.OriginType)
				}
				return &accounting.JournalEntry{Base: model.Base{ID: 900}}, nil
			},
		},
	)

	first, err := svc.YearEndRoll(ctx, YearEndRollRequest{
		OrganizationID: 1, TaxYearID: 3, JournalID: 2, RetainedEarningsAccountID: 32,
	})
	if err != nil {
		t.Fatalf("first roll failed: %v", err)
	}
	if first.NetIncome != 400 || first.ZeroedAccounts != 2 || first.EntryID != 900 {
		t.Errorf("first roll = %+v, want net 400 zeroed 2 movement 900", first)
	}

	posted = true
	if _, err := svc.YearEndRoll(ctx, YearEndRollRequest{
		OrganizationID: 1, TaxYearID: 3, JournalID: 2, RetainedEarningsAccountID: 32,
	}); err != ErrYearEndAlreadyClosed {
		t.Errorf("second roll err = %v, want ErrYearEndAlreadyClosed", err)
	}
}

func TestStatementService_YearEndRoll_BalancedLines(t *testing.T) {
	ctx := context.Background()
	var lines []accounting.PostingLine
	svc := NewStatementService(
		ReportDAOMock{
			StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]AccountBalanceRow, error) {
				return []AccountBalanceRow{
					{AccountID: 41, Code: "4100", Name: "Sales", AccountType: "income", Balance: -1000},
					{AccountID: 50, Code: "5000", Name: "COGS", AccountType: "cogs", Balance: 600},
					{AccountID: 60, Code: "6000", Name: "OpEx", AccountType: "expense", Balance: 100},
				}, nil
			},
			HasYearEndCloseFn: func(_ context.Context, _, _ uint64) (bool, error) { return false, nil },
		},
		TaxPeriodFinderMock{},
		TaxPeriodByDateFinderMock{},
		TaxYearFinderMock{
			FindFn: func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
				return &reference.TaxYear{Name: "FY2026", DateStart: &start, DateEnd: &end}, nil
			},
		},
		StatementPosterMock{
			PostFn: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
				lines = request.Lines
				return &accounting.JournalEntry{Base: model.Base{ID: 901}}, nil
			},
		},
	)

	_, err := svc.YearEndRoll(ctx, YearEndRollRequest{
		OrganizationID: 1, TaxYearID: 3, JournalID: 2, RetainedEarningsAccountID: 32,
	})
	if err != nil {
		t.Fatalf("roll failed: %v", err)
	}
	if len(lines) != 4 {
		t.Fatalf("lines = %d, want 4", len(lines))
	}
	var debit, credit amount.Amount
	for _, line := range lines {
		debit = debit.Add(line.Debit)
		credit = credit.Add(line.Credit)
	}
	if !debit.Equal(credit) {
		t.Errorf("unbalanced roll: debit %v credit %v", debit, credit)
	}
	if !debit.Equal(amount.FromFloat64(1000)) {
		t.Errorf("debit total = %v, want 1000", debit)
	}
}
