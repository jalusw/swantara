package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
)

func reportTestSvc(reports reporting.ReportDAOMock, periods reporting.TaxPeriodFinderMock, summaries reporting.KpiSummaryDAOMock) reporting.ReportService {
	return reporting.NewReportService(reports, periods, summaries)
}

func statementTestSvc(
	reports reporting.ReportDAOMock,
	periods reporting.TaxPeriodFinderMock,
	periodByDate reporting.TaxPeriodByDateFinderMock,
	years reporting.TaxYearFinderMock,
	poster reporting.StatementPosterMock,
) reporting.StatementService {
	return reporting.NewStatementService(reports, periods, periodByDate, years, poster)
}

func kpiTestSvc(
	reports reporting.ReportDAOMock,
	pipeline reporting.PipelineForecasterMock,
	subscription reporting.SubscriptionMetricsProviderMock,
	projects reporting.ProjectListReaderMock,
	projectSvc reporting.ProjectSummarizerMock,
) reporting.KpiService {
	return reporting.NewKpiService(reports, pipeline, subscription, projects, projectSvc)
}

func fxRevaluationTestSvc(
	reports reporting.ReportDAOMock,
	revaluations reporting.FxRevaluationDAOMock,
	lines reporting.FxRevaluationLineDAOMock,
	poster *reporting.PosterMock,
	config reporting.ConfigSourceMock,
	rates reporting.RateResolverMock,
	orgs reporting.OrgReaderMock,
) reporting.FxRevaluationService {
	return reporting.NewFxRevaluationService(reports, revaluations, lines, poster, config, rates, orgs)
}

func accrualTestSvc(
	accruals reporting.AccrualDAOMock,
	lines reporting.AccrualLineDAOMock,
	poster *reporting.PosterMock,
	config reporting.ConfigSourceMock,
) reporting.AccrualService {
	return reporting.NewAccrualService(accruals, lines, poster, config)
}

func periodCloseTestSvc(
	periods *reporting.PeriodCloserMock,
	finder reporting.TaxPeriodFinderMock,
	config reporting.ConfigSourceMock,
	assets reporting.AssetListReaderMock,
	depreciation *reporting.DepreciationPosterMock,
	fx reporting.FxRevaluationService,
	accruals reporting.AccrualService,
	deferrals reporting.DeferralRecognizerMock,
	summaries reporting.KpiSummarizerMock,
) reporting.PeriodCloseService {
	return reporting.NewPeriodCloseService(periods, finder, config, assets, depreciation, fx, accruals, deferrals, summaries)
}

func handlerTestApp(t *testing.T, register func(fiber.Router, httpx.RouteGuards), withTenant bool) *fiber.App {
	t.Helper()
	app := fiber.New()
	if withTenant {
		app.Use(func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(10))
			c.Locals(model.ActorKey, uint64(5))
			return c.Next()
		})
	}
	register(app, passthroughGuards())
	return app
}

func sampleTaxYear() *reference.TaxYear {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	return &reference.TaxYear{Base: model.Base{ID: 1}, Name: "FY2026", DateStart: &start, DateEnd: &end}
}

func sampleForeignPosition() []reporting.ForeignPositionRow {
	return []reporting.ForeignPositionRow{
		{InvoiceID: 1, Type: "customer_invoice", CurrencyCode: "USD", AmountTotal: 100, AmountResidual: 100, AccountID: 100, AccountType: "receivable", BaseBalance: 14000},
	}
}

func samplePostedAccrual() *reporting.Accrual {
	reversalDate := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	moveID := uint64(5)
	return &reporting.Accrual{Base: model.Base{ID: 1}, OrganizationID: 10, PeriodID: 1, State: reporting.AccrualStatePosted, EntryID: &moveID, ReversalDate: &reversalDate}
}

func TestReportHandler_TrialBalance_ReturnsTrialBalance(t *testing.T) {
	reports := reporting.ReportDAOMock{
		TrialBalanceFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.TrialBalanceRow, error) {
			return []reporting.TrialBalanceRow{{AccountID: 1, Code: "1000", AccountType: "asset"}}, nil
		},
	}
	svc := reportTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/trial-balance?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestReportHandler_TrialBalance_ReturnsTrialBalanceFromSummary(t *testing.T) {
	summaries := reporting.KpiSummaryDAOMock{
		HasPeriodFn: func(_ context.Context, _, _ uint64) (bool, error) {
			return true, nil
		},
		TrialBalanceFn: func(_ context.Context, _, _ uint64) ([]reporting.TrialBalanceRow, error) {
			return []reporting.TrialBalanceRow{{AccountID: 1, Code: "1000", AccountType: "asset"}}, nil
		},
	}
	svc := reportTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, summaries)
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/trial-balance?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestReportHandler_TrialBalance_RejectsInvalidPeriod(t *testing.T) {
	svc := reportTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/trial-balance?period_id=abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestReportHandler_TrialBalance_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := reportTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/reports/trial-balance?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestReportHandler_TrialBalance_ReturnsNotFound(t *testing.T) {
	periods := reporting.TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, nil
		},
	}
	svc := reportTestSvc(reporting.ReportDAOMock{}, periods, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/trial-balance?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestReportHandler_TrialBalance_ReturnsServerErrorOnPeriodLookup(t *testing.T) {
	periods := reporting.TaxPeriodFinderMock{
		FindFn: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, errors.New("db down")
		},
	}
	svc := reportTestSvc(reporting.ReportDAOMock{}, periods, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/trial-balance?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestReportHandler_TrialBalance_ReturnsServerErrorOnReportQuery(t *testing.T) {
	reports := reporting.ReportDAOMock{
		TrialBalanceFn: func(_ context.Context, _ uint64, _, _ time.Time) ([]reporting.TrialBalanceRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := reportTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/trial-balance?period_id=1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestReportHandler_Aging_ReturnsAging(t *testing.T) {
	reports := reporting.ReportDAOMock{
		AgingFn: func(_ context.Context, _ uint64, _ time.Time) ([]reporting.AgingRow, error) {
			return []reporting.AgingRow{{Type: "customer_invoice", Bucket: "current", Amount: 100}}, nil
		},
	}
	svc := reportTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/aging?as_of=2026-08-01", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestReportHandler_Aging_RejectsInvalidAsOf(t *testing.T) {
	svc := reportTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/aging?as_of=bogus", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestReportHandler_Aging_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := reportTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/reports/aging", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestReportHandler_Aging_ReturnsServerError(t *testing.T) {
	reports := reporting.ReportDAOMock{
		AgingFn: func(_ context.Context, _ uint64, _ time.Time) ([]reporting.AgingRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := reportTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/aging", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestReportHandler_InventoryValuation_ReturnsValuation(t *testing.T) {
	reports := reporting.ReportDAOMock{
		InventoryValuationFn: func(_ context.Context, _ uint64) ([]reporting.InventoryValueRow, error) {
			return []reporting.InventoryValueRow{{ItemID: 1, ProductName: "Widget", Quantity: 10, Value: 500}}, nil
		},
	}
	svc := reportTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/inventory-valuation", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestReportHandler_InventoryValuation_ReturnsUnprocessableWhenTenantMissing(t *testing.T) {
	svc := reportTestSvc(reporting.ReportDAOMock{}, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, false)

	resp, err := doRequest(app, http.MethodGet, "/reports/inventory-valuation", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestReportHandler_InventoryValuation_ReturnsServerError(t *testing.T) {
	reports := reporting.ReportDAOMock{
		InventoryValuationFn: func(_ context.Context, _ uint64) ([]reporting.InventoryValueRow, error) {
			return nil, errors.New("db down")
		},
	}
	svc := reportTestSvc(reports, reporting.TaxPeriodFinderMock{}, reporting.KpiSummaryDAOMock{})
	app := handlerTestApp(t, NewReportHandler(svc).Register, true)

	resp, err := doRequest(app, http.MethodGet, "/reports/inventory-valuation", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestReportHandler_WriteReportError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "period not found", err: reporting.ErrPeriodNotFound, want: http.StatusNotFound},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeReportError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}

func writeErrorStatus(path string, fn func(c fiber.Ctx) error) int {
	app := fiber.New()
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}
