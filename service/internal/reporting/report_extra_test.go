package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
)

func TestReportService_TrialBalance_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	start := time.Now()
	end := start.AddDate(0, 1, 0)

	periodWithDates := TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{DateStart: &start, DateEnd: &end}, nil
		},
	}
	periodNoDates := TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{}, nil
		},
	}

	t.Run("period lookup error", func(t *testing.T) {
		svc := NewReportService(ReportDAOMock{}, TaxPeriodFinderMock{
			FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) { return nil, dbErr },
		}, KpiSummaryDAOMock{})
		_, err := svc.TrialBalance(ctx, 1, 7)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("period without dates", func(t *testing.T) {
		svc := NewReportService(ReportDAOMock{}, periodNoDates, KpiSummaryDAOMock{})
		_, err := svc.TrialBalance(ctx, 1, 7)
		helper.AssertError(t, err, true, ErrPeriodNotFound)
	})

	t.Run("summary check error", func(t *testing.T) {
		svc := NewReportService(ReportDAOMock{}, periodWithDates, KpiSummaryDAOMock{
			HasPeriodFn: func(_ context.Context, _, _ uint64) (bool, error) { return false, dbErr },
		})
		_, err := svc.TrialBalance(ctx, 1, 7)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("summary read error", func(t *testing.T) {
		svc := NewReportService(ReportDAOMock{}, periodWithDates, KpiSummaryDAOMock{
			HasPeriodFn: func(_ context.Context, _, _ uint64) (bool, error) { return true, nil },
			TrialBalanceFn: func(_ context.Context, _, _ uint64) ([]TrialBalanceRow, error) {
				return nil, dbErr
			},
		})
		_, err := svc.TrialBalance(ctx, 1, 7)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("report read error", func(t *testing.T) {
		svc := NewReportService(ReportDAOMock{
			TrialBalanceFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]TrialBalanceRow, error) {
				return nil, dbErr
			},
		}, periodWithDates, KpiSummaryDAOMock{})
		_, err := svc.TrialBalance(ctx, 1, 7)
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestReportService_Aging_Valuation(t *testing.T) {
	ctx := context.Background()
	asOf := time.Now()

	t.Run("delegates aging and valuation", func(t *testing.T) {
		svc := NewReportService(ReportDAOMock{
			AgingFn: func(_ context.Context, _ uint64, _ time.Time) ([]AgingRow, error) {
				return []AgingRow{{}}, nil
			},
			InventoryValuationFn: func(_ context.Context, _ uint64) ([]InventoryValueRow, error) {
				return []InventoryValueRow{{}}, nil
			},
		}, TaxPeriodFinderMock{}, KpiSummaryDAOMock{})
		rows, err := svc.AgingReport(ctx, 1, asOf)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(rows) != 1 {
			t.Errorf("rows = %d", len(rows))
		}
		values, err := svc.InventoryValuation(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(values) != 1 {
			t.Errorf("values = %d", len(values))
		}
	})
}

func TestSetClassifiers(t *testing.T) {
	kpi := KpiService{}.SetClassifier(accounting.GenericClassifier{})
	_ = kpi
	statement := StatementService{}.SetClassifier(accounting.GenericClassifier{}).SetFormatter(accounting.GenericFormatter{})
	_ = statement
}

func TestReportingFixtures_Opts(t *testing.T) {
	if FxRevaluationFixture(func(r *FxRevaluation) *FxRevaluation { return r }) == nil {
		t.Error("revaluation = nil")
	}
	if FxRevaluationLineFixture(func(l *FxRevaluationLine) *FxRevaluationLine { return l }) == nil {
		t.Error("line = nil")
	}
	if AccrualFixture(func(a *Accrual) *Accrual { return a }) == nil {
		t.Error("accrual = nil")
	}
	if AccrualLineFixture(func(l *AccrualLine) *AccrualLine { return l }) == nil {
		t.Error("accrual line = nil")
	}
}
