package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

func TestReportService_TrialBalance_ReconcilesRows(t *testing.T) {
	ctx := context.Background()
	reportDAO := ReportDAOMock{
		TrialBalanceFn: func(_ context.Context, _ uint64, start, end time.Time) ([]TrialBalanceRow, error) {
			return []TrialBalanceRow{
				{AccountID: 11, Code: "1100", Name: "Cash", AccountType: "asset", OpeningDebit: 100, PeriodDebit: 50, ClosingDebit: 150},
				{AccountID: 12, Code: "4000", Name: "Revenue", AccountType: "income", ClosingCredit: 200},
			}, nil
		},
	}
	svc := NewReportService(reportDAO, TaxPeriodFinderMock{}, KpiSummaryDAOMock{})

	trial, err := svc.TrialBalance(ctx, 1, 7)
	if err != nil {
		t.Fatalf("trial balance failed: %v", err)
	}
	if len(trial.Rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(trial.Rows))
	}
	if trial.Rows[0].OpeningDebit != 100 || trial.Rows[0].ClosingDebit != 150 {
		t.Errorf("row 0 = %+v, want opening 100 closing 150", trial.Rows[0])
	}
}

func TestReportService_TrialBalance_ReadsSummaryWhenAvailable(t *testing.T) {
	ctx := context.Background()
	svc := NewReportService(
		ReportDAOMock{},
		TaxPeriodFinderMock{},
		KpiSummaryDAOMock{
			HasPeriodFn: func(_ context.Context, _, _ uint64) (bool, error) { return true, nil },
			TrialBalanceFn: func(_ context.Context, _, _ uint64) ([]TrialBalanceRow, error) {
				return []TrialBalanceRow{
					{AccountID: 11, Code: "1100", Name: "Cash", AccountType: "asset", OpeningDebit: 100, PeriodDebit: 50, ClosingDebit: 150},
				}, nil
			},
		},
	)

	trial, err := svc.TrialBalance(ctx, 1, 7)
	if err != nil {
		t.Fatalf("trial balance failed: %v", err)
	}
	if len(trial.Rows) != 1 || trial.Rows[0].ClosingDebit != 150 {
		t.Errorf("rows = %+v, want summary row with closing 150", trial.Rows)
	}
}

func TestReportService_TrialBalance_PeriodNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewReportService(ReportDAOMock{}, TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		},
	}, KpiSummaryDAOMock{})

	if _, err := svc.TrialBalance(ctx, 1, 7); err != ErrPeriodNotFound {
		t.Errorf("err = %v, want ErrPeriodNotFound", err)
	}
}
