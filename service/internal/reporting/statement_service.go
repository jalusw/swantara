package reporting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

const yearEndCloseOrigin = "year_end_close"

var epochDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)

type TaxYearFinder interface {
	Find(ctx context.Context, id uint64) (*reference.TaxYear, error)
}

type TaxPeriodByDateFinder interface {
	FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*accounting.TaxPeriod, error)
}

type StatementService struct {
	reports      ReportDAO
	periods      TaxPeriodFinder
	periodByDate TaxPeriodByDateFinder
	years        TaxYearFinder
	poster       accounting.Poster
	classifier   accounting.AccountClassifier
	formatter    accounting.StatementFormatter
}

func NewStatementService(
	reports ReportDAO,
	periods TaxPeriodFinder,
	periodByDate TaxPeriodByDateFinder,
	years TaxYearFinder,
	poster accounting.Poster,
) StatementService {
	return StatementService{
		reports:      reports,
		periods:      periods,
		periodByDate: periodByDate,
		years:        years,
		poster:       poster,
		classifier:   accounting.GenericClassifier{},
		formatter:    accounting.GenericFormatter{},
	}
}

func (s StatementService) SetClassifier(classifier accounting.AccountClassifier) StatementService {
	s.classifier = classifier
	return s
}

func (s StatementService) SetFormatter(formatter accounting.StatementFormatter) StatementService {
	s.formatter = formatter
	return s
}

type ProfitAndLossRow struct {
	AccountID   uint64  `json:"account_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	AccountType string  `json:"account_type"`
	Amount      float64 `json:"amount"`
}

type ProfitAndLoss struct {
	OrganizationID uint64             `json:"organization_id"`
	PeriodID       uint64             `json:"period_id"`
	Start          time.Time          `json:"start"`
	End            time.Time          `json:"end"`
	Rows           []ProfitAndLossRow `json:"rows"`
	Revenue        float64            `json:"revenue"`
	COGS           float64            `json:"cogs"`
	GrossProfit    float64            `json:"gross_profit"`
	Expenses       float64            `json:"expenses"`
	NetIncome      float64            `json:"net_income"`
}

func (s StatementService) ProfitAndLoss(ctx context.Context, organizationID, periodID uint64) (ProfitAndLoss, error) {
	classifier := s.classifier
	formatter := s.formatter
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return ProfitAndLoss{}, err
	}
	if period == nil || period.DateStart == nil || period.DateEnd == nil {
		return ProfitAndLoss{}, ErrPeriodNotFound
	}
	balances, err := s.reports.StatementBalances(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return ProfitAndLoss{}, err
	}

	lines := make([]accounting.StatementLine, 0, len(balances))
	for _, row := range balances {
		section := classifier.ProfitLossSection(row.AccountType)
		amount := row.Balance
		if section == accounting.SectionRevenue || section == accounting.SectionOtherIncome {
			amount = -row.Balance
		}
		lines = append(lines, accounting.StatementLine{
			Name:    row.Name,
			Code:    row.Code,
			Amount:  amount,
			Section: section,
		})
	}
	formatted := formatter.FormatProfitLoss(lines)

	byCode := make(map[string]AccountBalanceRow, len(balances))
	for _, row := range balances {
		byCode[row.Code] = row
	}
	result := ProfitAndLoss{
		OrganizationID: organizationID,
		PeriodID:       periodID,
		Start:          *period.DateStart,
		End:            *period.DateEnd,
		GrossProfit:    formatted.GrossProfit,
		NetIncome:      formatted.NetIncome,
	}
	appendProfitAndLossRows := func(sectionLines []accounting.StatementLine, total *float64) {
		for _, line := range sectionLines {
			row, ok := byCode[line.Code]
			if !ok {
				continue
			}
			*total += line.Amount
			result.Rows = append(result.Rows, profitAndLossRow(row, line.Amount))
		}
	}
	appendProfitAndLossRows(formatted.Revenue, &result.Revenue)
	appendProfitAndLossRows(formatted.COGS, &result.COGS)
	appendProfitAndLossRows(formatted.OperatingExpenses, &result.Expenses)
	appendProfitAndLossRows(formatted.OtherIncome, &result.Revenue)
	appendProfitAndLossRows(formatted.OtherExpenses, &result.Expenses)
	appendProfitAndLossRows(formatted.TaxExpense, &result.Expenses)
	return result, nil
}

func profitAndLossRow(row AccountBalanceRow, amount float64) ProfitAndLossRow {
	return ProfitAndLossRow{
		AccountID:   row.AccountID,
		Code:        row.Code,
		Name:        row.Name,
		AccountType: row.AccountType,
		Amount:      amount,
	}
}

type BalanceSheetAccount struct {
	AccountID   uint64  `json:"account_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	AccountType string  `json:"account_type"`
	Balance     float64 `json:"balance"`
}

type BalanceSheet struct {
	OrganizationID   uint64                `json:"organization_id"`
	AsOf             time.Time             `json:"as_of"`
	Assets           []BalanceSheetAccount `json:"assets"`
	Liabilities      []BalanceSheetAccount `json:"liabilities"`
	Equity           []BalanceSheetAccount `json:"equity"`
	CurrentEarnings  float64               `json:"current_earnings"`
	TotalAssets      float64               `json:"total_assets"`
	TotalLiabilities float64               `json:"total_liabilities"`
	TotalEquity      float64               `json:"total_equity"`
}

func (s StatementService) BalanceSheet(ctx context.Context, organizationID uint64, asOf time.Time) (BalanceSheet, error) {
	classifier := s.classifier
	formatter := s.formatter
	balances, err := s.reports.StatementBalances(ctx, organizationID, epochDate, asOf)
	if err != nil {
		return BalanceSheet{}, err
	}

	lines := make([]accounting.StatementLine, 0, len(balances))
	for _, row := range balances {
		lines = append(lines, accounting.StatementLine{
			Name:    row.Name,
			Code:    row.Code,
			Amount:  row.Balance,
			Section: classifier.BalanceSheetSection(row.AccountType),
		})
	}
	formatted := formatter.FormatBalanceSheet(lines)

	byCode := make(map[string]AccountBalanceRow, len(balances))
	for _, row := range balances {
		byCode[row.Code] = row
	}
	sheet := BalanceSheet{
		OrganizationID:   organizationID,
		AsOf:             asOf,
		TotalAssets:      formatted.TotalAssets,
		TotalLiabilities: formatted.TotalLiabilities,
		TotalEquity:      formatted.TotalEquity,
	}
	for _, line := range formatted.CurrentAssets {
		if row, ok := byCode[line.Code]; ok {
			sheet.Assets = append(sheet.Assets, balanceSheetAccount(row))
		}
	}
	for _, line := range formatted.NonCurrentAssets {
		if row, ok := byCode[line.Code]; ok {
			sheet.Assets = append(sheet.Assets, balanceSheetAccount(row))
		}
	}
	for _, line := range formatted.CurrentLiabilities {
		if row, ok := byCode[line.Code]; ok {
			sheet.Liabilities = append(sheet.Liabilities, balanceSheetAccount(row))
		}
	}
	for _, line := range formatted.NonCurrentLiabilities {
		if row, ok := byCode[line.Code]; ok {
			sheet.Liabilities = append(sheet.Liabilities, balanceSheetAccount(row))
		}
	}
	for _, line := range formatted.Equity {
		if row, ok := byCode[line.Code]; ok {
			sheet.Equity = append(sheet.Equity, balanceSheetAccount(row))
		}
	}

	yearStart := time.Date(asOf.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	earnings, err := s.netIncome(ctx, organizationID, yearStart, asOf, classifier)
	if err != nil {
		return BalanceSheet{}, err
	}
	sheet.CurrentEarnings = earnings
	sheet.TotalEquity += earnings
	return sheet, nil
}

func balanceSheetAccount(row AccountBalanceRow) BalanceSheetAccount {
	return BalanceSheetAccount(row)
}

func (s StatementService) netIncome(ctx context.Context, organizationID uint64, start, end time.Time, classifier accounting.AccountClassifier) (float64, error) {
	balances, err := s.reports.StatementBalances(ctx, organizationID, start, end)
	if err != nil {
		return 0, err
	}
	revenue, cogs, expense := 0.0, 0.0, 0.0
	for _, row := range balances {
		section := classifier.ProfitLossSection(row.AccountType)
		switch section {
		case accounting.SectionRevenue:
			revenue += -row.Balance
		case accounting.SectionCOGS:
			cogs += row.Balance
		case accounting.SectionOperatingExpense:
			expense += row.Balance
		}
	}
	return revenue - cogs - expense, nil
}

type CashFlow struct {
	OrganizationID uint64    `json:"organization_id"`
	Start          time.Time `json:"start"`
	End            time.Time `json:"end"`
	Operating      float64   `json:"operating"`
	Investing      float64   `json:"investing"`
	Financing      float64   `json:"financing"`
	NetChange      float64   `json:"net_change"`
	OpeningCash    float64   `json:"opening_cash"`
	ClosingCash    float64   `json:"closing_cash"`
}

func (s StatementService) CashFlow(ctx context.Context, organizationID uint64, start, end time.Time) (CashFlow, error) {
	classifier := s.classifier
	formatter := s.formatter
	rows, err := s.reports.CashFlow(ctx, organizationID, start, end)
	if err != nil {
		return CashFlow{}, err
	}

	lines := make([]accounting.StatementLine, 0, len(rows))
	for _, row := range rows {
		section, ok := cashFlowSectionFromRow(row.Section)
		if !ok {
			continue
		}
		lines = append(lines, accounting.StatementLine{
			Amount:          row.Amount,
			CashFlowSection: section,
		})
	}

	opening, err := s.cashBalance(ctx, organizationID, start.AddDate(0, 0, -1), classifier)
	if err != nil {
		return CashFlow{}, err
	}
	formatted := formatter.FormatCashFlow(lines, opening)
	return CashFlow{
		OrganizationID: organizationID,
		Start:          start,
		End:            end,
		Operating:      sumCashFlowLines(formatted.Operating),
		Investing:      sumCashFlowLines(formatted.Investing),
		Financing:      sumCashFlowLines(formatted.Financing),
		NetChange:      formatted.NetChange,
		OpeningCash:    formatted.OpeningCash,
		ClosingCash:    formatted.ClosingCash,
	}, nil
}

func cashFlowSectionFromRow(section string) (string, bool) {
	switch section {
	case accounting.CashFlowOperating:
		return accounting.CashFlowOperating, true
	case accounting.CashFlowInvesting:
		return accounting.CashFlowInvesting, true
	case accounting.CashFlowFinancing:
		return accounting.CashFlowFinancing, true
	default:
		return "", false
	}
}

func sumCashFlowLines(lines []accounting.StatementLine) float64 {
	var total float64
	for _, line := range lines {
		total += line.Amount
	}
	return total
}

func (s StatementService) cashBalance(ctx context.Context, organizationID uint64, asOf time.Time, classifier accounting.AccountClassifier) (float64, error) {
	balances, err := s.reports.StatementBalances(ctx, organizationID, epochDate, asOf)
	if err != nil {
		return 0, err
	}
	var total float64
	for _, row := range balances {
		if classifier.IsCashEquivalent(row.AccountType) {
			total += row.Balance
		}
	}
	return total, nil
}

type YearEndRollRequest struct {
	OrganizationID            uint64
	TaxYearID                 uint64
	JournalID                 uint64
	RetainedEarningsAccountID uint64
}

type YearEndRollResult struct {
	PeriodID       uint64    `json:"period_id"`
	Date           time.Time `json:"date"`
	NetIncome      float64   `json:"net_income"`
	ZeroedAccounts int       `json:"zeroed_accounts"`
	EntryID        uint64    `json:"entry_id"`
}

func (s StatementService) YearEndRoll(ctx context.Context, request YearEndRollRequest) (YearEndRollResult, error) {
	classifier := s.classifier
	year, err := s.years.Find(ctx, request.TaxYearID)
	if err != nil {
		return YearEndRollResult{}, err
	}
	if year == nil || year.DateStart == nil || year.DateEnd == nil {
		return YearEndRollResult{}, ErrPeriodNotFound
	}
	yearEnd := *year.DateEnd

	period, err := s.periodByDate.FindByDate(ctx, request.OrganizationID, yearEnd)
	if err != nil {
		return YearEndRollResult{}, err
	}
	if period == nil {
		return YearEndRollResult{}, ErrPeriodNotFound
	}

	closed, err := s.reports.HasYearEndClose(ctx, request.OrganizationID, period.ID)
	if err != nil {
		return YearEndRollResult{}, err
	}
	if closed {
		return YearEndRollResult{}, ErrYearEndAlreadyClosed
	}

	balances, err := s.reports.StatementBalances(ctx, request.OrganizationID, *year.DateStart, yearEnd)
	if err != nil {
		return YearEndRollResult{}, err
	}

	lines := []accounting.PostingLine{}
	netIncome := 0.0
	zeroed := 0
	for _, row := range balances {
		if classifier.IsIncome(row.AccountType) {
			amt := -row.Balance
			if amt == 0 {
				continue
			}
			lines = append(lines, accounting.PostingLine{AccountID: row.AccountID, Name: "Year-end close " + row.Code, Debit: amount.FromFloat64(amt)})
			netIncome += amt
			zeroed++
		} else if classifier.IsCOGS(row.AccountType) || classifier.IsExpense(row.AccountType) || classifier.IsDepreciation(row.AccountType) {
			if row.Balance == 0 {
				continue
			}
			lines = append(lines, accounting.PostingLine{AccountID: row.AccountID, Name: "Year-end close " + row.Code, Credit: amount.FromFloat64(row.Balance)})
			netIncome -= row.Balance
			zeroed++
		}
	}

	result := YearEndRollResult{PeriodID: period.ID, Date: yearEnd, NetIncome: netIncome, ZeroedAccounts: zeroed}
	if netIncome == 0 && zeroed == 0 {
		return result, nil
	}
	if netIncome > 0 {
		lines = append(lines, accounting.PostingLine{AccountID: request.RetainedEarningsAccountID, Name: "Retained earnings", Credit: amount.FromFloat64(netIncome)})
	} else {
		lines = append(lines, accounting.PostingLine{AccountID: request.RetainedEarningsAccountID, Name: "Retained earnings", Debit: amount.FromFloat64(-netIncome)})
	}

	movement, err := s.poster.Post(ctx, accounting.PostRequest{
		OrganizationID: request.OrganizationID,
		JournalID:      request.JournalID,
		Date:           yearEnd,
		Ref:            "YE-" + year.Name,
		OriginType:     yearEndCloseOrigin,
		OriginID:       period.ID,
		Description:    "Year-end retained earnings roll",
		Lines:          lines,
	})
	if err != nil {
		return YearEndRollResult{}, err
	}
	result.EntryID = movement.ID
	return result, nil
}
