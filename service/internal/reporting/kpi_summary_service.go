package reporting

import (
	"context"
)

type KpiSummaryService struct {
	summaries KpiSummaryDAO
	periods   TaxPeriodFinder
}

func NewKpiSummaryService(summaries KpiSummaryDAO, periods TaxPeriodFinder) KpiSummaryService {
	return KpiSummaryService{summaries: summaries, periods: periods}
}

func (s KpiSummaryService) RefreshPeriod(ctx context.Context, organizationID, periodID uint64) (int, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return 0, err
	}
	if period == nil || period.DateStart == nil || period.DateEnd == nil {
		return 0, ErrPeriodNotFound
	}

	opening, err := s.summaries.OpeningBalances(ctx, organizationID, *period.DateStart)
	if err != nil {
		return 0, err
	}
	activity, err := s.summaries.PeriodActivity(ctx, organizationID, *period.DateStart, *period.DateEnd)
	if err != nil {
		return 0, err
	}

	rows := make([]KpiAccountSummary, 0, len(opening)+len(activity))
	byAccount := map[uint64]*KpiAccountSummary{}
	for _, balance := range opening {
		row := &KpiAccountSummary{
			OrganizationID: organizationID,
			TaxPeriodID:    periodID,
			AccountID:      balance.AccountID,
			OpeningDebit:   balance.Debit,
			OpeningCredit:  balance.Credit,
		}
		byAccount[balance.AccountID] = row
		rows = append(rows, *row)
	}
	for _, balance := range activity {
		row, ok := byAccount[balance.AccountID]
		if !ok {
			row = &KpiAccountSummary{
				OrganizationID: organizationID,
				TaxPeriodID:    periodID,
				AccountID:      balance.AccountID,
			}
			byAccount[balance.AccountID] = row
			rows = append(rows, *row)
		}
		row.PeriodDebit = balance.Debit
		row.PeriodCredit = balance.Credit
	}
	for i := range rows {
		row := byAccount[rows[i].AccountID]
		rows[i] = *row
	}

	if err := s.summaries.ReplacePeriod(ctx, organizationID, periodID, rows); err != nil {
		return 0, err
	}
	return len(rows), nil
}
