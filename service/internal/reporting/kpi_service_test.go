package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

func TestKpiService_SalesKPI_ComputesMarginAndWinRate(t *testing.T) {
	ctx := context.Background()
	reportDAO := ReportDAOMock{
		BookingsFn: func(_ context.Context, _ uint64, _, _ time.Time) (float64, error) { return 5000, nil },
		BalancesByTypeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]TypeBalanceRow, error) {
			return []TypeBalanceRow{
				{AccountType: "income", Credit: 4000},
				{AccountType: "cogs", Debit: 2400},
			}, nil
		},
	}
	svc := NewKpiService(reportDAO, PipelineForecasterMock{
		ForecastFn: func(_ context.Context, _ *uint64) (crm.PipelineForecast, error) {
			return crm.PipelineForecast{WinRate: 40}, nil
		},
	}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.SalesKPI(ctx, 1, time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		t.Fatalf("sales kpi failed: %v", err)
	}
	if kpi.Bookings != 5000 || kpi.Revenue != 4000 || kpi.COGS != 2400 {
		t.Errorf("kpi = %+v, want bookings 5000 revenue 4000 cogs 2400", kpi)
	}
	if kpi.GrossMargin != 1600 || kpi.GrossMarginPct != 0.4 {
		t.Errorf("gross margin = %+v, want 1600 / 40%%", kpi)
	}
	if kpi.WinRate != 40 {
		t.Errorf("win rate = %v, want 40", kpi.WinRate)
	}
}

func TestKpiService_PipelineKPI_ReturnsWeightedPipeline(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(ReportDAOMock{}, PipelineForecasterMock{
		ForecastFn: func(_ context.Context, _ *uint64) (crm.PipelineForecast, error) {
			return crm.PipelineForecast{TotalExpectedRevenue: 1000, WeightedPipeline: 400, WinRate: 50}, nil
		},
	}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.PipelineKPI(ctx, 1)
	if err != nil {
		t.Fatalf("pipeline kpi failed: %v", err)
	}
	if kpi.WeightedPipeline != 400 || kpi.WinRate != 50 {
		t.Errorf("kpi = %+v, want weighted 400 win rate 50", kpi)
	}
}

func TestKpiService_InventoryKPI_SumsValuation(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(ReportDAOMock{
		InventoryValuationFn: func(_ context.Context, _ uint64) ([]InventoryValueRow, error) {
			return []InventoryValueRow{
				{ItemID: 1, Quantity: 10, Value: 150},
				{ItemID: 2, Quantity: 5, Value: 50},
			}, nil
		},
	}, PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.InventoryKPI(ctx, 1)
	if err != nil {
		t.Fatalf("inventory kpi failed: %v", err)
	}
	if kpi.OnHandValue != 200 || kpi.OnHandQuantity != 15 || kpi.ProductCount != 2 {
		t.Errorf("kpi = %+v, want value 200 qty 15 count 2", kpi)
	}
}

func TestKpiService_SubscriptionKPI_DelegatesMetrics(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(ReportDAOMock{}, PipelineForecasterMock{}, SubscriptionMetricsProviderMock{
		MetricsFn: func(_ context.Context, _ uint64) (subscription.Metrics, error) {
			return subscription.Metrics{MRR: 1000, ARR: 12000, ChurnRate: 0.05, LTV: 240000}, nil
		},
	}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.SubscriptionKPI(ctx, 1)
	if err != nil {
		t.Fatalf("subscription kpi failed: %v", err)
	}
	if kpi.MRR != 1000 || kpi.ARR != 12000 || kpi.LTV != 240000 {
		t.Errorf("kpi = %+v, want MRR 1000 ARR 12000 LTV 240000", kpi)
	}
}

func TestKpiService_ProjectKPI_AggregatesMarginAndUtilization(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(ReportDAOMock{}, PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{
		ListFn: func(_ context.Context, _ *query.Query) (*query.Page[project.Project], error) {
			return &query.Page[project.Project]{Items: []*project.Project{{}, {}}}, nil
		},
	}, ProjectSummarizerMock{
		SummaryFn: func(_ context.Context, projectID uint64) (project.ProjectSummary, error) {
			return project.ProjectSummary{MarginAmount: 1000, CostAmount: 500, BilledAmount: 1500, Utilization: 0.5}, nil
		},
	})

	kpi, err := svc.ProjectKPI(ctx, 1)
	if err != nil {
		t.Fatalf("project kpi failed: %v", err)
	}
	if kpi.ProjectCount != 2 || kpi.TotalMargin != 2000 || kpi.Utilization != 0.5 {
		t.Errorf("kpi = %+v, want count 2 margin 2000 utilization 0.5", kpi)
	}
}

func TestKpiService_PayrollKPI_ReturnsCosts(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(ReportDAOMock{
		PayrollCostFn: func(_ context.Context, _ uint64, _, _ time.Time) (float64, float64, error) {
			return 5000, 3800, nil
		},
	}, PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.PayrollKPI(ctx, 1, time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		t.Fatalf("payroll kpi failed: %v", err)
	}
	if kpi.GrossCost != 5000 || kpi.NetCost != 3800 {
		t.Errorf("kpi = %+v, want gross 5000 net 3800", kpi)
	}
}

func TestKpiService_FinanceKPI_ComputesMarginsAndCurrentRatio(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(ReportDAOMock{
		BalancesByTypeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]TypeBalanceRow, error) {
			return []TypeBalanceRow{
				{AccountType: "income", Credit: 10000},
				{AccountType: "cogs", Debit: 6000},
				{AccountType: "expense", Debit: 1000},
				{AccountType: "depreciation", Debit: 500},
				{AccountType: "bank", Debit: 5000},
				{AccountType: "payable", Credit: 2000},
			}, nil
		},
	}, PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.FinanceKPI(ctx, 1, time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		t.Fatalf("finance kpi failed: %v", err)
	}
	if kpi.Revenue != 10000 || kpi.Expenses != 7500 {
		t.Errorf("revenue/expenses = %+v, want 10000/7500", kpi)
	}
	if kpi.GrossMarginPct != 0.4 || kpi.NetMarginPct != 0.25 {
		t.Errorf("margins = %+v, want 40%%/25%%", kpi)
	}
	if kpi.EBITDA != 3000 {
		t.Errorf("ebitda = %v, want 3000", kpi.EBITDA)
	}
	if kpi.CurrentRatio != 2.5 {
		t.Errorf("current ratio = %v, want 2.5", kpi.CurrentRatio)
	}
}

func TestKpiService_ProcurementKPI_ComputesRatios(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(
		ReportDAOMock{
			ProcurementMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (ProcurementMetrics, error) {
				return ProcurementMetrics{Count: 4, AvgCycleDays: 3.5, ReceivedCount: 4, OnTimeCount: 3, ActualCost: 4500, StandardCost: 4000}, nil
			},
		},
		PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.ProcurementKPI(ctx, 1, time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		t.Fatalf("procurement kpi failed: %v", err)
	}
	if kpi.AvgCycleDays != 3.5 || kpi.PurchaseCount != 4 {
		t.Errorf("cycle = %v count = %d, want 3.5 / 4", kpi.AvgCycleDays, kpi.PurchaseCount)
	}
	if kpi.OnTimeDeliveryPct != 0.75 {
		t.Errorf("on time = %v, want 0.75", kpi.OnTimeDeliveryPct)
	}
	if kpi.PriceVariancePct != 0.125 {
		t.Errorf("price variance = %v, want 0.125", kpi.PriceVariancePct)
	}
}

func TestKpiService_ManufacturingKPI_ComputesRatios(t *testing.T) {
	ctx := context.Background()
	svc := NewKpiService(
		ReportDAOMock{
			ManufacturingMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (ManufacturingMetrics, error) {
				return ManufacturingMetrics{
					OrderCount: 2, PlannedMinutes: 100, ActualMinutes: 80,
					QtyToProduce: 10, QtyProduced: 9, PlannedMaterial: 10, ConsumedMaterial: 11,
					ActualValue: 1100, StandardValue: 1000,
				}, nil
			},
		},
		PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.ManufacturingKPI(ctx, 1, time.Now().AddDate(0, 0, -30), time.Now())
	if err != nil {
		t.Fatalf("manufacturing kpi failed: %v", err)
	}
	if kpi.OrderCount != 2 || kpi.OEEPct != 1.25 {
		t.Errorf("oee = %+v, want 1.25", kpi)
	}
	if kpi.YieldPct != 0.9 || kpi.ScrapPct != 0.1 || kpi.CostVariancePct != 0.1 {
		t.Errorf("yield/scrap/variance = %+v, want 0.9/0.1/0.1", kpi)
	}
}

func TestKpiService_ArApKPI_ComputesDSOAndDPO(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	svc := NewKpiService(
		ReportDAOMock{
			ArApMetricsFn: func(_ context.Context, _ uint64, _ time.Time) (ArApMetrics, error) {
				return ArApMetrics{OpenAR: 1000, OpenAP: 500, OverdueAR: 100, OverdueAP: 50, Revenue: 7300, Purchases: 3650}, nil
			},
		},
		PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.ArApKPI(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("ar/ap kpi failed: %v", err)
	}
	if kpi.DSO != 50 || kpi.DPO != 50 {
		t.Errorf("dso/dpo = %+v, want 50/50", kpi)
	}
	if kpi.OverdueARPct != 0.1 || kpi.OverdueAPPct != 0.1 {
		t.Errorf("overdue = %+v, want 0.1/0.1", kpi)
	}
}

func TestKpiService_CashKPI_ComputesPositionAndForecast(t *testing.T) {
	ctx := context.Background()
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	svc := NewKpiService(
		ReportDAOMock{
			CashMetricsFn: func(_ context.Context, _ uint64, _ time.Time) (CashMetrics, error) {
				return CashMetrics{BankBalance: 2000, Outflows: 1500, Inflows: 500, OpenAR: 1000, OpenAP: 700}, nil
			},
		},
		PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.CashKPI(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("cash kpi failed: %v", err)
	}
	if kpi.Position != 2000 || kpi.Burn != 1000 || kpi.Forecast != 2300 {
		t.Errorf("cash = %+v, want 2000/1000/2300", kpi)
	}
}

func TestKpiService_InventoryRatioKPI_ComputesTurnover(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	svc := NewKpiService(
		ReportDAOMock{
			InventoryRatioMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (InventoryRatioMetrics, error) {
				return InventoryRatioMetrics{COGS: 3000, AvgInventory: 1000, StockoutCount: 2}, nil
			},
		},
		PipelineForecasterMock{}, SubscriptionMetricsProviderMock{}, ProjectListReaderMock{}, ProjectSummarizerMock{})

	kpi, err := svc.InventoryRatioKPI(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("inventory ratio kpi failed: %v", err)
	}
	if kpi.StockoutCount != 2 {
		t.Errorf("stockouts = %d, want 2", kpi.StockoutCount)
	}
	if kpi.Turnover != 3 || kpi.DaysOnHand != 10 {
		t.Errorf("turnover/days = %+v, want 3/10", kpi)
	}
}
