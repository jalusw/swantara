package reporting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

type TaxPeriodFinder interface {
	Find(ctx context.Context, id uint64) (*accounting.TaxPeriod, error)
}

type ReportService struct {
	reports   ReportDAO
	periods   TaxPeriodFinder
	summaries KpiSummaryDAO
}

func NewReportService(reports ReportDAO, periods TaxPeriodFinder, summaries KpiSummaryDAO) ReportService {
	return ReportService{reports: reports, periods: periods, summaries: summaries}
}

type TrialBalance struct {
	OrganizationID uint64            `json:"organization_id"`
	PeriodID       uint64            `json:"period_id"`
	Start          time.Time         `json:"start"`
	End            time.Time         `json:"end"`
	Rows           []TrialBalanceRow `json:"rows"`
}

func (s ReportService) TrialBalance(ctx context.Context, organizationID, periodID uint64) (TrialBalance, error) {
	period, err := s.periods.Find(ctx, periodID)
	if err != nil {
		return TrialBalance{}, err
	}
	if period == nil {
		return TrialBalance{}, ErrPeriodNotFound
	}
	if period.DateStart == nil || period.DateEnd == nil {
		return TrialBalance{}, ErrPeriodNotFound
	}

	var rows []TrialBalanceRow
	hasSummary, err := s.summaries.HasPeriod(ctx, organizationID, periodID)
	if err != nil {
		return TrialBalance{}, err
	}
	if hasSummary {
		rows, err = s.summaries.TrialBalance(ctx, organizationID, periodID)
	} else {
		rows, err = s.reports.TrialBalance(ctx, organizationID, *period.DateStart, *period.DateEnd)
	}
	if err != nil {
		return TrialBalance{}, err
	}
	return TrialBalance{
		OrganizationID: organizationID,
		PeriodID:       periodID,
		Start:          *period.DateStart,
		End:            *period.DateEnd,
		Rows:           rows,
	}, nil
}

func (s ReportService) AgingReport(ctx context.Context, organizationID uint64, asOf time.Time) ([]AgingRow, error) {
	return s.reports.Aging(ctx, organizationID, asOf)
}

func (s ReportService) InventoryValuation(ctx context.Context, organizationID uint64) ([]InventoryValueRow, error) {
	return s.reports.InventoryValuation(ctx, organizationID)
}
