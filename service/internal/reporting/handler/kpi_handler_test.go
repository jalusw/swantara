package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

func TestKpiHandler_ReturnsKpi(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		reports      reporting.ReportDAOMock
		pipeline     reporting.PipelineForecasterMock
		subscription reporting.SubscriptionMetricsProviderMock
		projects     reporting.ProjectListReaderMock
	}{
		{name: "sales", path: "/kpis/sales"},
		{name: "pipeline", path: "/kpis/pipeline"},
		{name: "inventory", path: "/kpis/inventory"},
		{name: "subscription", path: "/kpis/subscription"},
		{name: "projects", path: "/kpis/projects"},
		{name: "payroll", path: "/kpis/payroll"},
		{name: "finance", path: "/kpis/finance"},
		{name: "procurement", path: "/kpis/procurement"},
		{name: "manufacturing", path: "/kpis/manufacturing"},
		{name: "ar-ap", path: "/kpis/ar-ap"},
		{name: "cash", path: "/kpis/cash"},
		{name: "inventory-ratio", path: "/kpis/inventory-ratio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := kpiTestSvc(tt.reports, tt.pipeline, tt.subscription, tt.projects, reporting.ProjectSummarizerMock{})
			app := handlerTestApp(t, NewKpiHandler(svc).Register, true)

			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want 200", resp.StatusCode)
			}
		})
	}
}

func TestKpiHandler_ReturnsKpiWithExplicitDates(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "sales", path: "/kpis/sales?start=2026-08-01&end=2026-08-31"},
		{name: "payroll", path: "/kpis/payroll?start=2026-08-01&end=2026-08-31"},
		{name: "finance", path: "/kpis/finance?start=2026-08-01&end=2026-08-31"},
		{name: "procurement", path: "/kpis/procurement?start=2026-08-01&end=2026-08-31"},
		{name: "manufacturing", path: "/kpis/manufacturing?start=2026-08-01&end=2026-08-31"},
		{name: "inventory-ratio", path: "/kpis/inventory-ratio?start=2026-08-01&end=2026-08-31"},
		{name: "ar-ap", path: "/kpis/ar-ap?as_of=2026-08-01"},
		{name: "cash", path: "/kpis/cash?as_of=2026-08-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := kpiTestSvc(reporting.ReportDAOMock{}, reporting.PipelineForecasterMock{}, reporting.SubscriptionMetricsProviderMock{}, reporting.ProjectListReaderMock{}, reporting.ProjectSummarizerMock{})
			app := handlerTestApp(t, NewKpiHandler(svc).Register, true)

			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status = %d, want 200", resp.StatusCode)
			}
		})
	}
}

func TestKpiHandler_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "sales", path: "/kpis/sales"},
		{name: "pipeline", path: "/kpis/pipeline"},
		{name: "inventory", path: "/kpis/inventory"},
		{name: "subscription", path: "/kpis/subscription"},
		{name: "projects", path: "/kpis/projects"},
		{name: "payroll", path: "/kpis/payroll"},
		{name: "finance", path: "/kpis/finance"},
		{name: "procurement", path: "/kpis/procurement"},
		{name: "manufacturing", path: "/kpis/manufacturing"},
		{name: "ar-ap", path: "/kpis/ar-ap"},
		{name: "cash", path: "/kpis/cash"},
		{name: "inventory-ratio", path: "/kpis/inventory-ratio"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := kpiTestSvc(reporting.ReportDAOMock{}, reporting.PipelineForecasterMock{}, reporting.SubscriptionMetricsProviderMock{}, reporting.ProjectListReaderMock{}, reporting.ProjectSummarizerMock{})
			app := handlerTestApp(t, NewKpiHandler(svc).Register, false)

			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", resp.StatusCode)
			}
		})
	}
}

func TestKpiHandler_ReturnsServerError(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		reports      reporting.ReportDAOMock
		pipeline     reporting.PipelineForecasterMock
		subscription reporting.SubscriptionMetricsProviderMock
		projects     reporting.ProjectListReaderMock
	}{
		{
			name: "sales",
			path: "/kpis/sales",
			reports: reporting.ReportDAOMock{
				BookingsFn: func(_ context.Context, _ uint64, _, _ time.Time) (float64, error) {
					return 0, errors.New("db down")
				},
			},
		},
		{
			name: "pipeline",
			path: "/kpis/pipeline",
			pipeline: reporting.PipelineForecasterMock{
				ForecastFn: func(_ context.Context, _ *uint64) (crm.PipelineForecast, error) {
					return crm.PipelineForecast{}, errors.New("db down")
				},
			},
		},
		{
			name: "inventory",
			path: "/kpis/inventory",
			reports: reporting.ReportDAOMock{
				InventoryValuationFn: func(_ context.Context, _ uint64) ([]reporting.InventoryValueRow, error) {
					return nil, errors.New("db down")
				},
			},
		},
		{
			name: "subscription",
			path: "/kpis/subscription",
			subscription: reporting.SubscriptionMetricsProviderMock{
				MetricsFn: func(_ context.Context, _ uint64) (subscription.Metrics, error) {
					return subscription.Metrics{}, errors.New("db down")
				},
			},
		},
		{
			name: "projects",
			path: "/kpis/projects",
			projects: reporting.ProjectListReaderMock{
				ListFn: func(_ context.Context, _ *query.Query) (*query.Page[project.Project], error) {
					return nil, errors.New("db down")
				},
			},
		},
		{
			name: "payroll",
			path: "/kpis/payroll",
			reports: reporting.ReportDAOMock{
				PayrollCostFn: func(_ context.Context, _ uint64, _, _ time.Time) (float64, float64, error) {
					return 0, 0, errors.New("db down")
				},
			},
		},
		{
			name: "finance",
			path: "/kpis/finance",
			reports: reporting.ReportDAOMock{
				BalancesByTypeFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.TypeBalanceRow, error) {
					return nil, errors.New("db down")
				},
			},
		},
		{
			name: "procurement",
			path: "/kpis/procurement",
			reports: reporting.ReportDAOMock{
				ProcurementMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (reporting.ProcurementMetrics, error) {
					return reporting.ProcurementMetrics{}, errors.New("db down")
				},
			},
		},
		{
			name: "manufacturing",
			path: "/kpis/manufacturing",
			reports: reporting.ReportDAOMock{
				ManufacturingMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (reporting.ManufacturingMetrics, error) {
					return reporting.ManufacturingMetrics{}, errors.New("db down")
				},
			},
		},
		{
			name: "ar-ap",
			path: "/kpis/ar-ap",
			reports: reporting.ReportDAOMock{
				ArApMetricsFn: func(_ context.Context, _ uint64, _ time.Time) (reporting.ArApMetrics, error) {
					return reporting.ArApMetrics{}, errors.New("db down")
				},
			},
		},
		{
			name: "cash",
			path: "/kpis/cash",
			reports: reporting.ReportDAOMock{
				CashMetricsFn: func(_ context.Context, _ uint64, _ time.Time) (reporting.CashMetrics, error) {
					return reporting.CashMetrics{}, errors.New("db down")
				},
			},
		},
		{
			name: "inventory-ratio",
			path: "/kpis/inventory-ratio",
			reports: reporting.ReportDAOMock{
				InventoryRatioMetricsFn: func(_ context.Context, _ uint64, _, _ time.Time) (reporting.InventoryRatioMetrics, error) {
					return reporting.InventoryRatioMetrics{}, errors.New("db down")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := kpiTestSvc(tt.reports, tt.pipeline, tt.subscription, tt.projects, reporting.ProjectSummarizerMock{})
			app := handlerTestApp(t, NewKpiHandler(svc).Register, true)

			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusInternalServerError {
				t.Fatalf("status = %d, want 500", resp.StatusCode)
			}
		})
	}
}

func TestKpiHandler_RejectsInvalidDates(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "bad start", path: "/kpis/sales?start=bogus"},
		{name: "bad end", path: "/kpis/payroll?end=bogus"},
		{name: "bad finance start", path: "/kpis/finance?start=bogus"},
		{name: "bad procurement end", path: "/kpis/procurement?end=bogus"},
		{name: "bad manufacturing start", path: "/kpis/manufacturing?start=bogus"},
		{name: "bad inventory-ratio end", path: "/kpis/inventory-ratio?end=bogus"},
		{name: "bad as-of", path: "/kpis/ar-ap?as_of=bogus"},
		{name: "bad cash as-of", path: "/kpis/cash?as_of=bogus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := kpiTestSvc(reporting.ReportDAOMock{}, reporting.PipelineForecasterMock{}, reporting.SubscriptionMetricsProviderMock{}, reporting.ProjectListReaderMock{}, reporting.ProjectSummarizerMock{})
			app := handlerTestApp(t, NewKpiHandler(svc).Register, true)

			resp, err := doRequest(app, http.MethodGet, tt.path, "")
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", resp.StatusCode)
			}
		})
	}
}
