package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

func TestReportDAOMock_All(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	end := start.AddDate(0, 1, 0)

	bare := ReportDAOMock{}
	if _, err := bare.TrialBalance(ctx, 1, start, end); err != nil {
		t.Errorf("TrialBalance = %v", err)
	}
	if _, err := bare.Aging(ctx, 1, end); err != nil {
		t.Errorf("Aging = %v", err)
	}
	if _, err := bare.InventoryValuation(ctx, 1); err != nil {
		t.Errorf("InventoryValuation = %v", err)
	}
	if _, err := bare.BalancesByType(ctx, 1, start, end); err != nil {
		t.Errorf("BalancesByType = %v", err)
	}
	if _, err := bare.Bookings(ctx, 1, start, end); err != nil {
		t.Errorf("Bookings = %v", err)
	}
	if _, _, err := bare.PayrollCost(ctx, 1, start, end); err != nil {
		t.Errorf("PayrollCost = %v", err)
	}
	if _, err := bare.OpenForeignPositions(ctx, 1); err != nil {
		t.Errorf("OpenForeignPositions = %v", err)
	}
	if _, err := bare.StatementBalances(ctx, 1, start, end); err != nil {
		t.Errorf("StatementBalances = %v", err)
	}
	if _, err := bare.CashFlow(ctx, 1, start, end); err != nil {
		t.Errorf("CashFlow = %v", err)
	}
	if _, err := bare.HasYearEndClose(ctx, 1, 2); err != nil {
		t.Errorf("HasYearEndClose = %v", err)
	}
	if _, err := bare.ProcurementMetrics(ctx, 1, start, end); err != nil {
		t.Errorf("ProcurementMetrics = %v", err)
	}
	if _, err := bare.ManufacturingMetrics(ctx, 1, start, end); err != nil {
		t.Errorf("ManufacturingMetrics = %v", err)
	}
	if _, err := bare.ArApMetrics(ctx, 1, end); err != nil {
		t.Errorf("ArApMetrics = %v", err)
	}
	if _, err := bare.CashMetrics(ctx, 1, end); err != nil {
		t.Errorf("CashMetrics = %v", err)
	}
	if _, err := bare.InventoryRatioMetrics(ctx, 1, start, end); err != nil {
		t.Errorf("InventoryRatioMetrics = %v", err)
	}

	wired := ReportDAOMock{
		TrialBalanceFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]TrialBalanceRow, error) {
			return []TrialBalanceRow{{}}, nil
		},
		AgingFn:              func(_ context.Context, _ uint64, _ time.Time) ([]AgingRow, error) { return []AgingRow{{}}, nil },
		InventoryValuationFn: func(_ context.Context, _ uint64) ([]InventoryValueRow, error) { return []InventoryValueRow{{}}, nil },
		BalancesByTypeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]TypeBalanceRow, error) {
			return []TypeBalanceRow{{}}, nil
		},
		BookingsFn:             func(_ context.Context, _ uint64, _, _ time.Time) (float64, error) { return 1, nil },
		PayrollCostFn:          func(_ context.Context, _ uint64, _, _ time.Time) (float64, float64, error) { return 1, 2, nil },
		OpenForeignPositionsFn: func(_ context.Context, _ uint64) ([]ForeignPositionRow, error) { return []ForeignPositionRow{{}}, nil },
		StatementBalancesFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]AccountBalanceRow, error) {
			return []AccountBalanceRow{{}}, nil
		},
		CashFlowFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]CashFlowSectionRow, error) {
			return []CashFlowSectionRow{{}}, nil
		},
		HasYearEndCloseFn: func(_ context.Context, _, _ uint64) (bool, error) { return true, nil },
		ProcurementMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (ProcurementMetrics, error) {
			return ProcurementMetrics{}, nil
		},
		ManufacturingMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (ManufacturingMetrics, error) {
			return ManufacturingMetrics{}, nil
		},
		ArApMetricsFn: func(_ context.Context, _ uint64, _ time.Time) (ArApMetrics, error) { return ArApMetrics{}, nil },
		CashMetricsFn: func(_ context.Context, _ uint64, _ time.Time) (CashMetrics, error) { return CashMetrics{}, nil },
		InventoryRatioMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (InventoryRatioMetrics, error) {
			return InventoryRatioMetrics{}, nil
		},
	}
	if _, err := wired.TrialBalance(ctx, 1, start, end); err != nil {
		t.Errorf("TrialBalance = %v", err)
	}
	if _, err := wired.Aging(ctx, 1, end); err != nil {
		t.Errorf("Aging = %v", err)
	}
	if _, err := wired.InventoryValuation(ctx, 1); err != nil {
		t.Errorf("InventoryValuation = %v", err)
	}
	if _, err := wired.BalancesByType(ctx, 1, start, end); err != nil {
		t.Errorf("BalancesByType = %v", err)
	}
	if _, err := wired.Bookings(ctx, 1, start, end); err != nil {
		t.Errorf("Bookings = %v", err)
	}
	if _, _, err := wired.PayrollCost(ctx, 1, start, end); err != nil {
		t.Errorf("PayrollCost = %v", err)
	}
	if _, err := wired.OpenForeignPositions(ctx, 1); err != nil {
		t.Errorf("OpenForeignPositions = %v", err)
	}
	if _, err := wired.StatementBalances(ctx, 1, start, end); err != nil {
		t.Errorf("StatementBalances = %v", err)
	}
	if _, err := wired.CashFlow(ctx, 1, start, end); err != nil {
		t.Errorf("CashFlow = %v", err)
	}
	if _, err := wired.HasYearEndClose(ctx, 1, 2); err != nil {
		t.Errorf("HasYearEndClose = %v", err)
	}
	if _, err := wired.ProcurementMetrics(ctx, 1, start, end); err != nil {
		t.Errorf("ProcurementMetrics = %v", err)
	}
	if _, err := wired.ManufacturingMetrics(ctx, 1, start, end); err != nil {
		t.Errorf("ManufacturingMetrics = %v", err)
	}
	if _, err := wired.ArApMetrics(ctx, 1, end); err != nil {
		t.Errorf("ArApMetrics = %v", err)
	}
	if _, err := wired.CashMetrics(ctx, 1, end); err != nil {
		t.Errorf("CashMetrics = %v", err)
	}
	if _, err := wired.InventoryRatioMetrics(ctx, 1, start, end); err != nil {
		t.Errorf("InventoryRatioMetrics = %v", err)
	}
}

func TestFxRevaluationMocks_All(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("revaluation dao", func(t *testing.T) {
		bare := FxRevaluationDAOMock{}
		if _, err := bare.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Create(ctx, &FxRevaluation{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := bare.Update(ctx, &FxRevaluation{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.ListByOrganization(ctx, 1); err != nil {
			t.Errorf("ListByOrganization = %v", err)
		}

		wired := FxRevaluationDAOMock{
			SearchFn:             func(_ context.Context, _ string, _ any) (*FxRevaluation, error) { return &FxRevaluation{}, nil },
			CreateFn:             func(_ context.Context, e *FxRevaluation) (*FxRevaluation, error) { return e, nil },
			FindFn:               func(_ context.Context, _ uint64) (*FxRevaluation, error) { return &FxRevaluation{}, nil },
			ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*FxRevaluation, error) { return []*FxRevaluation{{}}, nil },
		}
		if _, err := wired.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := wired.Create(ctx, &FxRevaluation{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if _, err := wired.ListByOrganization(ctx, 1); err != nil {
			t.Errorf("ListByOrganization = %v", err)
		}
	})

	t.Run("revaluation line dao", func(t *testing.T) {
		bare := FxRevaluationLineDAOMock{}
		if _, err := bare.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Create(ctx, &FxRevaluationLine{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &FxRevaluationLine{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.ListByRevaluation(ctx, 1); err != nil {
			t.Errorf("ListByRevaluation = %v", err)
		}
		if _, err := bare.ListOpenByOrg(ctx, 1); err != nil {
			t.Errorf("ListOpenByOrg = %v", err)
		}

		wired := FxRevaluationLineDAOMock{
			SearchFn:            func(_ context.Context, _ string, _ any) (*FxRevaluationLine, error) { return &FxRevaluationLine{}, nil },
			CreateFn:            func(_ context.Context, e *FxRevaluationLine) (*FxRevaluationLine, error) { return e, nil },
			UpdateFn:            func(_ context.Context, e *FxRevaluationLine) (*FxRevaluationLine, error) { return e, nil },
			ListByRevaluationFn: func(_ context.Context, _ uint64) ([]*FxRevaluationLine, error) { return []*FxRevaluationLine{{}}, nil },
			ListOpenByOrgFn:     func(_ context.Context, _ uint64) ([]*FxRevaluationLine, error) { return []*FxRevaluationLine{{}}, nil },
		}
		if _, err := wired.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := wired.Create(ctx, &FxRevaluationLine{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Update(ctx, &FxRevaluationLine{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if _, err := wired.ListByRevaluation(ctx, 1); err != nil {
			t.Errorf("ListByRevaluation = %v", err)
		}
		if _, err := wired.ListOpenByOrg(ctx, 1); err != nil {
			t.Errorf("ListOpenByOrg = %v", err)
		}
	})
}

func TestAccrualMocks_All(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("accrual dao", func(t *testing.T) {
		bare := AccrualDAOMock{}
		if _, err := bare.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Create(ctx, &Accrual{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &Accrual{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.ListByOrganization(ctx, 1); err != nil {
			t.Errorf("ListByOrganization = %v", err)
		}

		wired := AccrualDAOMock{
			SearchFn:             func(_ context.Context, _ string, _ any) (*Accrual, error) { return &Accrual{}, nil },
			CreateFn:             func(_ context.Context, e *Accrual) (*Accrual, error) { return e, nil },
			UpdateFn:             func(_ context.Context, e *Accrual) (*Accrual, error) { return e, nil },
			ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*Accrual, error) { return []*Accrual{{}}, nil },
		}
		if _, err := wired.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := wired.Create(ctx, &Accrual{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.Update(ctx, &Accrual{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if _, err := wired.ListByOrganization(ctx, 1); err != nil {
			t.Errorf("ListByOrganization = %v", err)
		}
	})

	t.Run("accrual line dao", func(t *testing.T) {
		bare := AccrualLineDAOMock{}
		if _, err := bare.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := bare.List(ctx, q); err != nil {
			t.Errorf("List = %v", err)
		}
		if _, err := bare.Create(ctx, &AccrualLine{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := bare.Update(ctx, &AccrualLine{}); err != nil {
			t.Errorf("Update = %v", err)
		}
		if _, err := bare.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		if err := bare.Delete(ctx, 1); err != nil {
			t.Errorf("Delete = %v", err)
		}
		if err := bare.HardDelete(ctx, 1); err != nil {
			t.Errorf("HardDelete = %v", err)
		}
		if _, err := bare.ListByAccrual(ctx, 1); err != nil {
			t.Errorf("ListByAccrual = %v", err)
		}

		wired := AccrualLineDAOMock{
			SearchFn:        func(_ context.Context, _ string, _ any) (*AccrualLine, error) { return &AccrualLine{}, nil },
			CreateFn:        func(_ context.Context, e *AccrualLine) (*AccrualLine, error) { return e, nil },
			ListByAccrualFn: func(_ context.Context, _ uint64) ([]*AccrualLine, error) { return []*AccrualLine{{}}, nil },
		}
		if _, err := wired.Search(ctx, "f", 1); err != nil {
			t.Errorf("Search = %v", err)
		}
		if _, err := wired.Create(ctx, &AccrualLine{}); err != nil {
			t.Errorf("Create = %v", err)
		}
		if _, err := wired.ListByAccrual(ctx, 1); err != nil {
			t.Errorf("ListByAccrual = %v", err)
		}
	})
}

func TestReportingInfraMocks_All(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	t.Run("config and rate", func(t *testing.T) {
		if _, err := (ConfigLookupMock{}).List(ctx, &query.Query{}); err != nil {
			t.Errorf("List = %v", err)
		}
		wired := ConfigLookupMock{
			ListFn: func(_ context.Context, _ *query.Query) (*query.Page[reference.SystemConfig], error) {
				return &query.Page[reference.SystemConfig]{}, nil
			},
		}
		if _, err := wired.List(ctx, &query.Query{}); err != nil {
			t.Errorf("List = %v", err)
		}

		if _, err := (OrgReaderMock{}).Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		orgWired := OrgReaderMock{
			FindFn: func(_ context.Context, _ uint64) (*reference.Organization, error) {
				return &reference.Organization{}, nil
			},
		}
		if _, err := orgWired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
	})

	t.Run("poster and reverse", func(t *testing.T) {
		if _, err := (&PosterMock{}).Post(ctx, accounting.PostRequest{}); err != nil {
			t.Errorf("Post = %v", err)
		}
		if _, err := (&PosterMock{}).Reverse(ctx, accounting.ReverseRequest{}); err != nil {
			t.Errorf("Reverse = %v", err)
		}
		wired := &PosterMock{
			PostFn: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
				return &accounting.JournalEntry{}, nil
			},
			ReverseFn: func(_ context.Context, _ accounting.ReverseRequest) (*accounting.JournalEntry, error) {
				return &accounting.JournalEntry{}, nil
			},
		}
		if _, err := wired.Post(ctx, accounting.PostRequest{}); err != nil {
			t.Errorf("Post = %v", err)
		}
		if _, err := wired.Reverse(ctx, accounting.ReverseRequest{}); err != nil {
			t.Errorf("Reverse = %v", err)
		}
	})

	t.Run("finders", func(t *testing.T) {
		if _, err := (TaxPeriodFinderMock{}).Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		fpWired := TaxPeriodFinderMock{
			FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return &accounting.TaxPeriod{}, nil
			},
		}
		if _, err := fpWired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}

		if _, err := (TaxPeriodByDateFinderMock{}).FindByDate(ctx, 1, now); err != nil {
			t.Errorf("FindByDate = %v", err)
		}
		fpdWired := TaxPeriodByDateFinderMock{
			FindByDateFn: func(_ context.Context, _ uint64, _ time.Time) (*accounting.TaxPeriod, error) {
				return &accounting.TaxPeriod{}, nil
			},
		}
		if _, err := fpdWired.FindByDate(ctx, 1, now); err != nil {
			t.Errorf("FindByDate = %v", err)
		}

		if _, err := (TaxYearFinderMock{}).Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
		fyWired := TaxYearFinderMock{
			FindFn: func(_ context.Context, _ uint64) (*reference.TaxYear, error) {
				return &reference.TaxYear{}, nil
			},
		}
		if _, err := fyWired.Find(ctx, 1); err != nil {
			t.Errorf("Find = %v", err)
		}
	})

	t.Run("statement poster", func(t *testing.T) {
		bare := StatementPosterMock{}
		if _, err := bare.Post(ctx, accounting.PostRequest{}); err != nil {
			t.Errorf("Post = %v", err)
		}
		if _, err := bare.PostTx(ctx, nil, accounting.PostRequest{}); err != nil {
			t.Errorf("PostTx = %v", err)
		}
		if _, err := bare.Reverse(ctx, accounting.ReverseRequest{}); err != nil {
			t.Errorf("Reverse = %v", err)
		}
		if _, err := bare.ReverseTx(ctx, nil, accounting.ReverseRequest{}); err != nil {
			t.Errorf("ReverseTx = %v", err)
		}
	})

	t.Run("period closer", func(t *testing.T) {
		bare := &PeriodCloserMock{}
		if _, err := bare.Close(ctx, 1); err != nil {
			t.Errorf("Close = %v", err)
		}
		if _, err := bare.Lock(ctx, 1); err != nil {
			t.Errorf("Lock = %v", err)
		}
		wired := &PeriodCloserMock{
			CloseFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return &accounting.TaxPeriod{}, nil
			},
			LockFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return &accounting.TaxPeriod{}, nil
			},
		}
		if _, err := wired.Close(ctx, 1); err != nil {
			t.Errorf("Close = %v", err)
		}
		if _, err := wired.Lock(ctx, 1); err != nil {
			t.Errorf("Lock = %v", err)
		}
	})

	t.Run("domain readers", func(t *testing.T) {
		if _, err := (AssetListReaderMock{}).List(ctx, &query.Query{}); err != nil {
			t.Errorf("assets = %v", err)
		}
		assetsWired := AssetListReaderMock{
			ListFn: func(_ context.Context, _ *query.Query) (*query.Page[asset.FixedAsset], error) {
				return &query.Page[asset.FixedAsset]{}, nil
			},
		}
		if _, err := assetsWired.List(ctx, &query.Query{}); err != nil {
			t.Errorf("assets = %v", err)
		}

		if _, err := (&DepreciationPosterMock{}).PostDepreciation(ctx, asset.PostDepreciationRequest{}); err != nil {
			t.Errorf("depreciation = %v", err)
		}
		depWired := &DepreciationPosterMock{
			PostFn: func(_ context.Context, _ asset.PostDepreciationRequest) (*asset.AssetDepreciationLine, error) {
				return &asset.AssetDepreciationLine{}, nil
			},
		}
		if _, err := depWired.PostDepreciation(ctx, asset.PostDepreciationRequest{}); err != nil {
			t.Errorf("depreciation = %v", err)
		}

		if _, err := (DeferralRecognizerMock{}).RecognizeDue(ctx, nil, now); err != nil {
			t.Errorf("deferral = %v", err)
		}
		defWired := DeferralRecognizerMock{
			RecognizeDueFn: func(_ context.Context, _ *uint64, _ time.Time) (int, error) { return 1, nil },
		}
		if _, err := defWired.RecognizeDue(ctx, nil, now); err != nil {
			t.Errorf("deferral = %v", err)
		}

		if _, err := (PipelineForecasterMock{}).Forecast(ctx, nil); err != nil {
			t.Errorf("forecast = %v", err)
		}
		fcWired := PipelineForecasterMock{
			ForecastFn: func(_ context.Context, _ *uint64) (crm.PipelineForecast, error) { return crm.PipelineForecast{}, nil },
		}
		if _, err := fcWired.Forecast(ctx, nil); err != nil {
			t.Errorf("forecast = %v", err)
		}

		if _, err := (SubscriptionMetricsProviderMock{}).Metrics(ctx, 1); err != nil {
			t.Errorf("metrics = %v", err)
		}
		subWired := SubscriptionMetricsProviderMock{
			MetricsFn: func(_ context.Context, _ uint64) (subscription.Metrics, error) { return subscription.Metrics{}, nil },
		}
		if _, err := subWired.Metrics(ctx, 1); err != nil {
			t.Errorf("metrics = %v", err)
		}

		if _, err := (ProjectListReaderMock{}).List(ctx, &query.Query{}); err != nil {
			t.Errorf("projects = %v", err)
		}
		projWired := ProjectListReaderMock{
			ListFn: func(_ context.Context, _ *query.Query) (*query.Page[project.Project], error) {
				return &query.Page[project.Project]{}, nil
			},
		}
		if _, err := projWired.List(ctx, &query.Query{}); err != nil {
			t.Errorf("projects = %v", err)
		}

		if _, err := (ProjectSummarizerMock{}).Summary(ctx, 1); err != nil {
			t.Errorf("summary = %v", err)
		}
		sumWired := ProjectSummarizerMock{
			SummaryFn: func(_ context.Context, _ uint64) (project.ProjectSummary, error) {
				return project.ProjectSummary{}, nil
			},
		}
		if _, err := sumWired.Summary(ctx, 1); err != nil {
			t.Errorf("summary = %v", err)
		}
	})

	t.Run("kpi summary", func(t *testing.T) {
		bare := KpiSummarizerMock{}
		if _, err := bare.RefreshPeriod(ctx, 1, 2); err != nil {
			t.Errorf("RefreshPeriod = %v", err)
		}
		wired := KpiSummarizerMock{
			RefreshPeriodFn: func(_ context.Context, _, _ uint64) (int, error) { return 1, nil },
		}
		if _, err := wired.RefreshPeriod(ctx, 1, 2); err != nil {
			t.Errorf("RefreshPeriod = %v", err)
		}

		summary := KpiSummaryDAOMock{}
		if _, err := summary.OpeningBalances(ctx, 1, now); err != nil {
			t.Errorf("OpeningBalances = %v", err)
		}
		if _, err := summary.PeriodActivity(ctx, 1, now, now); err != nil {
			t.Errorf("PeriodActivity = %v", err)
		}
		if err := summary.ReplacePeriod(ctx, 1, 2, nil); err != nil {
			t.Errorf("ReplacePeriod = %v", err)
		}
		if _, err := summary.TrialBalance(ctx, 1, 2); err != nil {
			t.Errorf("TrialBalance = %v", err)
		}
		if _, err := summary.ListPeriod(ctx, 1, 2); err != nil {
			t.Errorf("ListPeriod = %v", err)
		}
	})
}

func TestReportingFixtures(t *testing.T) {
	if FxRevaluationFixture() == nil {
		t.Error("revaluation = nil")
	}
	if FxRevaluationLineFixture() == nil {
		t.Error("line = nil")
	}
	if AccrualFixture() == nil {
		t.Error("accrual = nil")
	}
	if AccrualLineFixture() == nil {
		t.Error("accrual line = nil")
	}
}

var _ = amount.Zero
