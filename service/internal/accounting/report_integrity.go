package accounting

import (
	"context"
	"math"
	"sort"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type IntegrityCheck struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail"`
}

type IntegrityReport struct {
	PeriodID uint64           `json:"period_id"`
	Passed   bool             `json:"passed"`
	Checks   []IntegrityCheck `json:"checks"`
}

type ReportIntegrityService struct {
	trialBalances TrialBalanceDAO
	cashFlow      CashFlowDAO
	equity        EquityDAO
	periods       TaxPeriodDAO
	accounts      AccountLookup
	classifier    AccountClassifier
}

func NewReportIntegrityService(trialBalances TrialBalanceDAO, cashFlow CashFlowDAO, equity EquityDAO, periods TaxPeriodDAO, accounts AccountLookup) ReportIntegrityService {
	return ReportIntegrityService{trialBalances: trialBalances, cashFlow: cashFlow, equity: equity, periods: periods, accounts: accounts, classifier: GenericClassifier{}}
}

func (s ReportIntegrityService) CheckPeriod(ctx context.Context, organizationID, periodID uint64) (*IntegrityReport, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return nil, err
	}
	if period == nil || period.OrganizationID != organizationID {
		return nil, ErrPeriodNotFound
	}
	if period.DateStart == nil || period.DateEnd == nil {
		return nil, ErrInvalidPeriod
	}
	asOfClosing := period.DateEnd.Format("2006-01-02")
	asOfOpening := period.DateStart.AddDate(0, 0, -1).Format("2006-01-02")

	closingTB, err := s.trialBalances.GenerateAsOf(ctx, organizationID, asOfClosing, true)
	if err != nil {
		return nil, err
	}
	openingTB, err := s.trialBalances.GenerateAsOf(ctx, organizationID, asOfOpening, true)
	if err != nil {
		return nil, err
	}
	periodLines, err := s.trialBalances.GenerateByPeriod(ctx, organizationID, periodID)
	if err != nil {
		return nil, err
	}
	cash, err := s.cashFlow.GenerateByPeriod(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}
	openingCash, err := s.cashFlow.GetOpeningCash(ctx, organizationID, *period.DateStart)
	if err != nil {
		return nil, err
	}
	movements, err := s.equity.EquityBalancesByPeriod(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}
	openingEquity, err := s.equity.OpeningEquityBalance(ctx, organizationID, *period.DateStart)
	if err != nil {
		return nil, err
	}
	netIncome, err := s.equity.NetIncomeForPeriod(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return nil, err
	}

	classifier := s.classifier

	checks := []IntegrityCheck{
		checkTrialBalanced(closingTB),
		checkCashArithmetic(openingCash, cash),
		checkEquityArithmetic(openingEquity, movements, netIncome, closingTB),
		checkCashToLedger(closingTB, openingCash, cash, classifier),
		checkEquityIncomeToLedger(periodLines, netIncome, classifier),
		checkOpeningPlusMovement(openingTB, periodLines, closingTB),
	}
	passed := true
	for _, c := range checks {
		if !c.Passed {
			passed = false
			break
		}
	}
	return &IntegrityReport{PeriodID: periodID, Passed: passed, Checks: checks}, nil
}

func checkTrialBalanced(tb []TrialBalanceLine) IntegrityCheck {
	debits := amount.Zero()
	credits := amount.Zero()
	for _, line := range tb {
		debits = debits.Add(line.Debit)
		credits = credits.Add(line.Credit)
	}
	passed := amount.IsBalanced(debits, credits, 2)
	return IntegrityCheck{Name: "trial_balance_balanced", Passed: passed, Detail: "closing debits equal credits"}
}

func checkCashArithmetic(opening float64, lines []CashFlowLine) IntegrityCheck {
	net := 0.0
	sections := map[string]float64{}
	for _, line := range lines {
		net += line.Amount
		sections[line.CashFlowSection] += line.Amount
	}
	sectionTotal := 0.0
	for _, total := range sections {
		sectionTotal += total
	}
	passed := absDiff(net, sectionTotal) <= 0.01 && !math.IsNaN(opening+net)
	return IntegrityCheck{Name: "cash_flow_arithmetic", Passed: passed, Detail: "sections foot to net movement"}
}

func checkEquityArithmetic(opening, movements []EquityMovement, _ float64, closingTB []TrialBalanceLine) IntegrityCheck {
	openingSum := 0.0
	for _, m := range opening {
		openingSum += m.Amount
	}
	movementSum := 0.0
	for _, m := range movements {
		movementSum += m.Amount
	}
	closingSum := 0.0
	for _, line := range closingTB {
		if line.AccountType == AccountTypeEquity {
			closingSum += line.Credit.Sub(line.Debit).Float64()
		}
	}
	passed := absDiff(closingSum, openingSum+movementSum) <= 0.01
	return IntegrityCheck{Name: "equity_arithmetic", Passed: passed, Detail: "closing equals opening plus movements"}
}

func checkCashToLedger(closingTB []TrialBalanceLine, opening float64, lines []CashFlowLine, classifier AccountClassifier) IntegrityCheck {
	ledgerCash := 0.0
	for _, line := range closingTB {
		if classifier.IsCashEquivalent(line.AccountType) {
			ledgerCash += line.Balance.Float64()
		}
	}
	net := 0.0
	for _, line := range lines {
		net += line.Amount
	}
	passed := absDiff(ledgerCash, opening+net) <= 0.01
	return IntegrityCheck{Name: "cash_to_ledger", Passed: passed, Detail: "TB cash and bank equals cash closing"}
}

func checkEquityIncomeToLedger(periodLines []TrialBalanceLine, netIncome float64, classifier AccountClassifier) IntegrityCheck {
	income := amount.Zero()
	for _, line := range periodLines {
		switch {
		case classifier.IsIncome(line.AccountType):
			income = income.Add(line.Credit.Sub(line.Debit))
		case classifier.IsCOGS(line.AccountType), classifier.IsExpense(line.AccountType), classifier.IsDepreciation(line.AccountType):
			income = income.Sub(line.Debit.Sub(line.Credit))
		case classifier.ProfitLossSection(line.AccountType) == SectionTax:
			income = income.Sub(line.Debit.Sub(line.Credit))
		}
	}
	passed := absDiff(income.Float64(), netIncome) <= 0.01
	return IntegrityCheck{Name: "equity_income_to_ledger", Passed: passed, Detail: "equity net income equals TB P&L"}
}

func checkOpeningPlusMovement(opening, movement, closing []TrialBalanceLine) IntegrityCheck {
	openMap := map[uint64]amount.Amount{}
	for _, line := range opening {
		openMap[line.AccountID] = line.Balance
	}
	for _, line := range movement {
		openMap[line.AccountID] = openMap[line.AccountID].Add(line.Balance)
	}
	closingMap := map[uint64]amount.Amount{}
	for _, line := range closing {
		closingMap[line.AccountID] = line.Balance
	}
	for id, expected := range openMap {
		if actual, ok := closingMap[id]; ok {
			if !actual.Round(2).Equal(expected.Round(2)) {
				return IntegrityCheck{Name: "opening_plus_movement", Passed: false, Detail: "opening plus movement equals closing"}
			}
		} else if !expected.Round(2).IsZero() {
			return IntegrityCheck{Name: "opening_plus_movement", Passed: false, Detail: "opening plus movement equals closing"}
		}
	}
	for id, actual := range closingMap {
		if _, ok := openMap[id]; !ok {
			if !actual.Round(2).IsZero() {
				return IntegrityCheck{Name: "opening_plus_movement", Passed: false, Detail: "opening plus movement equals closing"}
			}
		}
	}
	return IntegrityCheck{Name: "opening_plus_movement", Passed: true, Detail: "opening plus movement equals closing"}
}

func absDiff(a, b float64) float64 {
	d := a - b
	if d < 0 {
		return -d
	}
	return d
}

func RollupTrialBalance(lines []TrialBalanceLine, accounts []*reference.Account) []TrialBalanceLine {
	byID := map[uint64]*reference.Account{}
	for _, acc := range accounts {
		byID[acc.ID] = acc
	}
	rolled := map[uint64]*TrialBalanceLine{}
	byAccount := map[uint64]TrialBalanceLine{}
	for _, line := range lines {
		cp := line
		byAccount[line.AccountID] = cp
		rolled[line.AccountID] = &cp
	}
	for _, line := range lines {
		acc := byID[line.AccountID]
		for acc != nil && acc.ParentID != nil {
			parent := byID[*acc.ParentID]
			if parent == nil {
				break
			}
			existing, ok := rolled[parent.ID]
			if !ok {
				created := TrialBalanceLine{AccountID: parent.ID, AccountCode: parent.Code, AccountName: parent.Name, AccountType: parent.Type}
				rolled[parent.ID] = &created
				existing = &created
			}
			existing.Debit = existing.Debit.Add(line.Debit)
			existing.Credit = existing.Credit.Add(line.Credit)
			existing.Balance = existing.Balance.Add(line.Balance)
			acc = parent
		}
	}
	out := make([]TrialBalanceLine, 0, len(rolled))
	for _, line := range rolled {
		out = append(out, *line)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].AccountCode < out[j].AccountCode })
	return out
}
