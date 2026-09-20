package reporting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"gorm.io/gorm"
)

type ReportDAOMock struct {
	TrialBalanceFn          func(ctx context.Context, organizationID uint64, start, end time.Time) ([]TrialBalanceRow, error)
	AgingFn                 func(ctx context.Context, organizationID uint64, asOf time.Time) ([]AgingRow, error)
	InventoryValuationFn    func(ctx context.Context, organizationID uint64) ([]InventoryValueRow, error)
	BalancesByTypeFn        func(ctx context.Context, organizationID uint64, start, end time.Time) ([]TypeBalanceRow, error)
	BookingsFn              func(ctx context.Context, organizationID uint64, start, end time.Time) (float64, error)
	PayrollCostFn           func(ctx context.Context, organizationID uint64, start, end time.Time) (float64, float64, error)
	OpenForeignPositionsFn  func(ctx context.Context, organizationID uint64) ([]ForeignPositionRow, error)
	StatementBalancesFn     func(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalanceRow, error)
	CashFlowFn              func(ctx context.Context, organizationID uint64, start, end time.Time) ([]CashFlowSectionRow, error)
	HasYearEndCloseFn       func(ctx context.Context, organizationID, periodID uint64) (bool, error)
	ProcurementMetricsFn    func(ctx context.Context, organizationID uint64, start, end time.Time) (ProcurementMetrics, error)
	ManufacturingMetricsFn  func(ctx context.Context, organizationID uint64, start, end time.Time) (ManufacturingMetrics, error)
	ArApMetricsFn           func(ctx context.Context, organizationID uint64, asOf time.Time) (ArApMetrics, error)
	CashMetricsFn           func(ctx context.Context, organizationID uint64, asOf time.Time) (CashMetrics, error)
	InventoryRatioMetricsFn func(ctx context.Context, organizationID uint64, start, end time.Time) (InventoryRatioMetrics, error)
}

func (m ReportDAOMock) TrialBalance(ctx context.Context, organizationID uint64, start, end time.Time) ([]TrialBalanceRow, error) {
	if m.TrialBalanceFn != nil {
		return m.TrialBalanceFn(ctx, organizationID, start, end)
	}
	return nil, nil
}

func (m ReportDAOMock) Aging(ctx context.Context, organizationID uint64, asOf time.Time) ([]AgingRow, error) {
	if m.AgingFn != nil {
		return m.AgingFn(ctx, organizationID, asOf)
	}
	return nil, nil
}

func (m ReportDAOMock) InventoryValuation(ctx context.Context, organizationID uint64) ([]InventoryValueRow, error) {
	if m.InventoryValuationFn != nil {
		return m.InventoryValuationFn(ctx, organizationID)
	}
	return nil, nil
}

func (m ReportDAOMock) BalancesByType(ctx context.Context, organizationID uint64, start, end time.Time) ([]TypeBalanceRow, error) {
	if m.BalancesByTypeFn != nil {
		return m.BalancesByTypeFn(ctx, organizationID, start, end)
	}
	return nil, nil
}

func (m ReportDAOMock) Bookings(ctx context.Context, organizationID uint64, start, end time.Time) (float64, error) {
	if m.BookingsFn != nil {
		return m.BookingsFn(ctx, organizationID, start, end)
	}
	return 0, nil
}

func (m ReportDAOMock) PayrollCost(ctx context.Context, organizationID uint64, start, end time.Time) (float64, float64, error) {
	if m.PayrollCostFn != nil {
		return m.PayrollCostFn(ctx, organizationID, start, end)
	}
	return 0, 0, nil
}

func (m ReportDAOMock) OpenForeignPositions(ctx context.Context, organizationID uint64) ([]ForeignPositionRow, error) {
	if m.OpenForeignPositionsFn != nil {
		return m.OpenForeignPositionsFn(ctx, organizationID)
	}
	return nil, nil
}

func (m ReportDAOMock) StatementBalances(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountBalanceRow, error) {
	if m.StatementBalancesFn != nil {
		return m.StatementBalancesFn(ctx, organizationID, start, end)
	}
	return nil, nil
}

func (m ReportDAOMock) CashFlow(ctx context.Context, organizationID uint64, start, end time.Time) ([]CashFlowSectionRow, error) {
	if m.CashFlowFn != nil {
		return m.CashFlowFn(ctx, organizationID, start, end)
	}
	return nil, nil
}

func (m ReportDAOMock) HasYearEndClose(ctx context.Context, organizationID, periodID uint64) (bool, error) {
	if m.HasYearEndCloseFn != nil {
		return m.HasYearEndCloseFn(ctx, organizationID, periodID)
	}
	return false, nil
}

func (m ReportDAOMock) ProcurementMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (ProcurementMetrics, error) {
	if m.ProcurementMetricsFn != nil {
		return m.ProcurementMetricsFn(ctx, organizationID, start, end)
	}
	return ProcurementMetrics{}, nil
}

func (m ReportDAOMock) ManufacturingMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (ManufacturingMetrics, error) {
	if m.ManufacturingMetricsFn != nil {
		return m.ManufacturingMetricsFn(ctx, organizationID, start, end)
	}
	return ManufacturingMetrics{}, nil
}

func (m ReportDAOMock) ArApMetrics(ctx context.Context, organizationID uint64, asOf time.Time) (ArApMetrics, error) {
	if m.ArApMetricsFn != nil {
		return m.ArApMetricsFn(ctx, organizationID, asOf)
	}
	return ArApMetrics{}, nil
}

func (m ReportDAOMock) CashMetrics(ctx context.Context, organizationID uint64, asOf time.Time) (CashMetrics, error) {
	if m.CashMetricsFn != nil {
		return m.CashMetricsFn(ctx, organizationID, asOf)
	}
	return CashMetrics{}, nil
}

func (m ReportDAOMock) InventoryRatioMetrics(ctx context.Context, organizationID uint64, start, end time.Time) (InventoryRatioMetrics, error) {
	if m.InventoryRatioMetricsFn != nil {
		return m.InventoryRatioMetricsFn(ctx, organizationID, start, end)
	}
	return InventoryRatioMetrics{}, nil
}

type FxRevaluationDAOMock struct {
	CreateFn             func(ctx context.Context, entity *FxRevaluation) (*FxRevaluation, error)
	FindFn               func(ctx context.Context, id uint64) (*FxRevaluation, error)
	SearchFn             func(ctx context.Context, field string, value any) (*FxRevaluation, error)
	ListByOrganizationFn func(ctx context.Context, organizationID uint64) ([]*FxRevaluation, error)
}

func (m FxRevaluationDAOMock) Search(ctx context.Context, field string, value any) (*FxRevaluation, error) {
	if m.SearchFn != nil {
		return m.SearchFn(ctx, field, value)
	}
	return nil, nil
}

func (m FxRevaluationDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[FxRevaluation], error) {
	return &query.Page[FxRevaluation]{}, nil
}

func (m FxRevaluationDAOMock) Create(ctx context.Context, entity *FxRevaluation) (*FxRevaluation, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, entity)
	}
	entity.ID = 1
	return entity, nil
}

func (m FxRevaluationDAOMock) Find(ctx context.Context, id uint64) (*FxRevaluation, error) {
	if m.FindFn != nil {
		return m.FindFn(ctx, id)
	}
	return &FxRevaluation{State: FxRevaluationStatePosted}, nil
}

func (m FxRevaluationDAOMock) Update(ctx context.Context, entity *FxRevaluation) (*FxRevaluation, error) {
	return entity, nil
}
func (m FxRevaluationDAOMock) Delete(ctx context.Context, id uint64) error     { return nil }
func (m FxRevaluationDAOMock) HardDelete(ctx context.Context, id uint64) error { return nil }
func (m FxRevaluationDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*FxRevaluation, error) {
	if m.ListByOrganizationFn != nil {
		return m.ListByOrganizationFn(ctx, organizationID)
	}
	return nil, nil
}

type FxRevaluationLineDAOMock struct {
	CreateFn            func(ctx context.Context, entity *FxRevaluationLine) (*FxRevaluationLine, error)
	UpdateFn            func(ctx context.Context, entity *FxRevaluationLine) (*FxRevaluationLine, error)
	SearchFn            func(ctx context.Context, field string, value any) (*FxRevaluationLine, error)
	ListByRevaluationFn func(ctx context.Context, revaluationID uint64) ([]*FxRevaluationLine, error)
	ListOpenByOrgFn     func(ctx context.Context, organizationID uint64) ([]*FxRevaluationLine, error)
}

func (m FxRevaluationLineDAOMock) Search(ctx context.Context, field string, value any) (*FxRevaluationLine, error) {
	if m.SearchFn != nil {
		return m.SearchFn(ctx, field, value)
	}
	return nil, nil
}

func (m FxRevaluationLineDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[FxRevaluationLine], error) {
	return &query.Page[FxRevaluationLine]{}, nil
}

func (m FxRevaluationLineDAOMock) Create(ctx context.Context, entity *FxRevaluationLine) (*FxRevaluationLine, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, entity)
	}
	entity.ID = 1
	return entity, nil
}

func (m FxRevaluationLineDAOMock) Update(ctx context.Context, entity *FxRevaluationLine) (*FxRevaluationLine, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, entity)
	}
	return entity, nil
}

func (m FxRevaluationLineDAOMock) Find(ctx context.Context, id uint64) (*FxRevaluationLine, error) {
	return nil, nil
}
func (m FxRevaluationLineDAOMock) Delete(ctx context.Context, id uint64) error     { return nil }
func (m FxRevaluationLineDAOMock) HardDelete(ctx context.Context, id uint64) error { return nil }
func (m FxRevaluationLineDAOMock) ListByRevaluation(ctx context.Context, revaluationID uint64) ([]*FxRevaluationLine, error) {
	if m.ListByRevaluationFn != nil {
		return m.ListByRevaluationFn(ctx, revaluationID)
	}
	return nil, nil
}
func (m FxRevaluationLineDAOMock) ListOpenByOrg(ctx context.Context, organizationID uint64) ([]*FxRevaluationLine, error) {
	if m.ListOpenByOrgFn != nil {
		return m.ListOpenByOrgFn(ctx, organizationID)
	}
	return nil, nil
}

type AccrualDAOMock struct {
	CreateFn             func(ctx context.Context, entity *Accrual) (*Accrual, error)
	UpdateFn             func(ctx context.Context, entity *Accrual) (*Accrual, error)
	SearchFn             func(ctx context.Context, field string, value any) (*Accrual, error)
	ListByOrganizationFn func(ctx context.Context, organizationID uint64) ([]*Accrual, error)
}

func (m AccrualDAOMock) Search(ctx context.Context, field string, value any) (*Accrual, error) {
	if m.SearchFn != nil {
		return m.SearchFn(ctx, field, value)
	}
	return nil, nil
}

func (m AccrualDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[Accrual], error) {
	return &query.Page[Accrual]{}, nil
}

func (m AccrualDAOMock) Create(ctx context.Context, entity *Accrual) (*Accrual, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, entity)
	}
	entity.ID = 1
	return entity, nil
}

func (m AccrualDAOMock) Update(ctx context.Context, entity *Accrual) (*Accrual, error) {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, entity)
	}
	return entity, nil
}

func (m AccrualDAOMock) Find(ctx context.Context, id uint64) (*Accrual, error) { return nil, nil }
func (m AccrualDAOMock) Delete(ctx context.Context, id uint64) error           { return nil }
func (m AccrualDAOMock) HardDelete(ctx context.Context, id uint64) error       { return nil }
func (m AccrualDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*Accrual, error) {
	if m.ListByOrganizationFn != nil {
		return m.ListByOrganizationFn(ctx, organizationID)
	}
	return nil, nil
}

type AccrualLineDAOMock struct {
	CreateFn        func(ctx context.Context, entity *AccrualLine) (*AccrualLine, error)
	SearchFn        func(ctx context.Context, field string, value any) (*AccrualLine, error)
	ListByAccrualFn func(ctx context.Context, accrualID uint64) ([]*AccrualLine, error)
}

func (m AccrualLineDAOMock) Search(ctx context.Context, field string, value any) (*AccrualLine, error) {
	if m.SearchFn != nil {
		return m.SearchFn(ctx, field, value)
	}
	return nil, nil
}

func (m AccrualLineDAOMock) List(ctx context.Context, q *query.Query) (*query.Page[AccrualLine], error) {
	return &query.Page[AccrualLine]{}, nil
}

func (m AccrualLineDAOMock) Create(ctx context.Context, entity *AccrualLine) (*AccrualLine, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, entity)
	}
	return entity, nil
}

func (m AccrualLineDAOMock) Update(ctx context.Context, entity *AccrualLine) (*AccrualLine, error) {
	return entity, nil
}
func (m AccrualLineDAOMock) Find(ctx context.Context, id uint64) (*AccrualLine, error) {
	return nil, nil
}
func (m AccrualLineDAOMock) Delete(ctx context.Context, id uint64) error     { return nil }
func (m AccrualLineDAOMock) HardDelete(ctx context.Context, id uint64) error { return nil }
func (m AccrualLineDAOMock) ListByAccrual(ctx context.Context, accrualID uint64) ([]*AccrualLine, error) {
	if m.ListByAccrualFn != nil {
		return m.ListByAccrualFn(ctx, accrualID)
	}
	return nil, nil
}

type ConfigLookupMock struct {
	ListFn func(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error)
}

func (m ConfigLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.SystemConfig], error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, q)
	}
	return &query.Page[reference.SystemConfig]{}, nil
}

type ConfigSourceMock struct {
	JournalIDFn           func(ctx context.Context, organizationID uint64) (uint64, error)
	FxGainLossAccountIDFn func(ctx context.Context, organizationID uint64) (uint64, error)
}

func (m ConfigSourceMock) JournalID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.JournalIDFn != nil {
		return m.JournalIDFn(ctx, organizationID)
	}
	return 1, nil
}

func (m ConfigSourceMock) FxGainLossAccountID(ctx context.Context, organizationID uint64) (uint64, error) {
	if m.FxGainLossAccountIDFn != nil {
		return m.FxGainLossAccountIDFn(ctx, organizationID)
	}
	return 9, nil
}

type RateResolverMock struct {
	RateFn func(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error)
}

func (m RateResolverMock) Rate(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error) {
	if m.RateFn != nil {
		return m.RateFn(ctx, currencyCode, organizationID, rateType, date)
	}
	return amount.FromInt64(1), nil
}

type OrgReaderMock struct {
	FindFn func(ctx context.Context, id uint64) (*reference.Organization, error)
}

func (m OrgReaderMock) Find(ctx context.Context, id uint64) (*reference.Organization, error) {
	if m.FindFn != nil {
		return m.FindFn(ctx, id)
	}
	return &reference.Organization{BaseCurrency: "IDR"}, nil
}

type PosterMock struct {
	PostFn       func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	ReverseFn    func(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
	postCount    int
	reverseCount int
}

func (m *PosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	m.postCount++
	if m.PostFn != nil {
		return m.PostFn(ctx, request)
	}
	movement := &accounting.JournalEntry{}
	movement.ID = uint64(m.postCount)
	return movement, nil
}

func (m *PosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	m.reverseCount++
	if m.ReverseFn != nil {
		return m.ReverseFn(ctx, request)
	}
	movement := &accounting.JournalEntry{}
	movement.ID = uint64(m.reverseCount + 100)
	return movement, nil
}

type TaxPeriodFinderMock struct {
	FindFn func(ctx context.Context, id uint64) (*accounting.TaxPeriod, error)
}

func (m TaxPeriodFinderMock) Find(ctx context.Context, id uint64) (*accounting.TaxPeriod, error) {
	if m.FindFn != nil {
		return m.FindFn(ctx, id)
	}
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	return &accounting.TaxPeriod{State: accounting.TaxPeriodStateOpen, DateStart: &start, DateEnd: &end}, nil
}

type TaxPeriodByDateFinderMock struct {
	FindByDateFn func(ctx context.Context, organizationID uint64, date time.Time) (*accounting.TaxPeriod, error)
}

func (m TaxPeriodByDateFinderMock) FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*accounting.TaxPeriod, error) {
	if m.FindByDateFn != nil {
		return m.FindByDateFn(ctx, organizationID, date)
	}
	end := date
	start := date.AddDate(0, -1, 0)
	return &accounting.TaxPeriod{State: accounting.TaxPeriodStateOpen, DateStart: &start, DateEnd: &end}, nil
}

type TaxYearFinderMock struct {
	FindFn func(ctx context.Context, id uint64) (*reference.TaxYear, error)
}

func (m TaxYearFinderMock) Find(ctx context.Context, id uint64) (*reference.TaxYear, error) {
	if m.FindFn != nil {
		return m.FindFn(ctx, id)
	}
	return &reference.TaxYear{Name: "FY2026", State: &[]string{"open"}[0]}, nil
}

type StatementPosterMock struct {
	PostFn func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
}

func (m StatementPosterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostFn != nil {
		return m.PostFn(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m StatementPosterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	return m.Post(ctx, request)
}

func (m StatementPosterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return &accounting.JournalEntry{}, nil
}

func (m StatementPosterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	return m.Reverse(ctx, request)
}

type PeriodCloserMock struct {
	CloseFn    func(ctx context.Context, periodID uint64) (*accounting.TaxPeriod, error)
	LockFn     func(ctx context.Context, periodID uint64) (*accounting.TaxPeriod, error)
	closeCount int
	lockCount  int
}

func (m *PeriodCloserMock) Close(ctx context.Context, periodID uint64) (*accounting.TaxPeriod, error) {
	m.closeCount++
	if m.CloseFn != nil {
		return m.CloseFn(ctx, periodID)
	}
	return &accounting.TaxPeriod{State: accounting.TaxPeriodStateClosed}, nil
}

func (m *PeriodCloserMock) Lock(ctx context.Context, periodID uint64) (*accounting.TaxPeriod, error) {
	m.lockCount++
	if m.LockFn != nil {
		return m.LockFn(ctx, periodID)
	}
	return &accounting.TaxPeriod{State: accounting.TaxPeriodStateLocked}, nil
}

type AssetListReaderMock struct {
	ListFn func(ctx context.Context, q *query.Query) (*query.Page[asset.FixedAsset], error)
}

func (m AssetListReaderMock) List(ctx context.Context, q *query.Query) (*query.Page[asset.FixedAsset], error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, q)
	}
	return &query.Page[asset.FixedAsset]{}, nil
}

type DepreciationPosterMock struct {
	PostFn    func(ctx context.Context, request asset.PostDepreciationRequest) (*asset.AssetDepreciationLine, error)
	postCount int
}

func (m *DepreciationPosterMock) PostDepreciation(ctx context.Context, request asset.PostDepreciationRequest) (*asset.AssetDepreciationLine, error) {
	m.postCount++
	if m.PostFn != nil {
		return m.PostFn(ctx, request)
	}
	if m.postCount == 1 {
		return &asset.AssetDepreciationLine{Posted: true}, nil
	}
	return nil, asset.ErrAssetNothingToPost
}

type DeferralRecognizerMock struct {
	RecognizeDueFn func(ctx context.Context, organizationID *uint64, asOf time.Time) (int, error)
}

func (m DeferralRecognizerMock) RecognizeDue(ctx context.Context, organizationID *uint64, asOf time.Time) (int, error) {
	if m.RecognizeDueFn != nil {
		return m.RecognizeDueFn(ctx, organizationID, asOf)
	}
	return 0, nil
}

type PipelineForecasterMock struct {
	ForecastFn func(ctx context.Context, organizationID *uint64) (crm.PipelineForecast, error)
}

func (m PipelineForecasterMock) Forecast(ctx context.Context, organizationID *uint64) (crm.PipelineForecast, error) {
	if m.ForecastFn != nil {
		return m.ForecastFn(ctx, organizationID)
	}
	return crm.PipelineForecast{}, nil
}

type SubscriptionMetricsProviderMock struct {
	MetricsFn func(ctx context.Context, organizationID uint64) (subscription.Metrics, error)
}

func (m SubscriptionMetricsProviderMock) Metrics(ctx context.Context, organizationID uint64) (subscription.Metrics, error) {
	if m.MetricsFn != nil {
		return m.MetricsFn(ctx, organizationID)
	}
	return subscription.Metrics{}, nil
}

type ProjectListReaderMock struct {
	ListFn func(ctx context.Context, q *query.Query) (*query.Page[project.Project], error)
}

func (m ProjectListReaderMock) List(ctx context.Context, q *query.Query) (*query.Page[project.Project], error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, q)
	}
	return &query.Page[project.Project]{}, nil
}

type ProjectSummarizerMock struct {
	SummaryFn func(ctx context.Context, projectID uint64) (project.ProjectSummary, error)
}

func (m ProjectSummarizerMock) Summary(ctx context.Context, projectID uint64) (project.ProjectSummary, error) {
	if m.SummaryFn != nil {
		return m.SummaryFn(ctx, projectID)
	}
	return project.ProjectSummary{}, nil
}

type KpiSummarizerMock struct {
	RefreshPeriodFn func(ctx context.Context, organizationID, periodID uint64) (int, error)
}

func (m KpiSummarizerMock) RefreshPeriod(ctx context.Context, organizationID, periodID uint64) (int, error) {
	if m.RefreshPeriodFn != nil {
		return m.RefreshPeriodFn(ctx, organizationID, periodID)
	}
	return 0, nil
}

type KpiSummaryDAOMock struct {
	OpeningBalancesFn func(ctx context.Context, organizationID uint64, before time.Time) ([]AccountPeriodBalance, error)
	PeriodActivityFn  func(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountPeriodBalance, error)
	ReplacePeriodFn   func(ctx context.Context, organizationID, periodID uint64, rows []KpiAccountSummary) error
	HasPeriodFn       func(ctx context.Context, organizationID, periodID uint64) (bool, error)
	TrialBalanceFn    func(ctx context.Context, organizationID, periodID uint64) ([]TrialBalanceRow, error)
	ListPeriodFn      func(ctx context.Context, organizationID, periodID uint64) ([]KpiAccountSummary, error)
}

func (m KpiSummaryDAOMock) OpeningBalances(ctx context.Context, organizationID uint64, before time.Time) ([]AccountPeriodBalance, error) {
	if m.OpeningBalancesFn != nil {
		return m.OpeningBalancesFn(ctx, organizationID, before)
	}
	return nil, nil
}

func (m KpiSummaryDAOMock) PeriodActivity(ctx context.Context, organizationID uint64, start, end time.Time) ([]AccountPeriodBalance, error) {
	if m.PeriodActivityFn != nil {
		return m.PeriodActivityFn(ctx, organizationID, start, end)
	}
	return nil, nil
}

func (m KpiSummaryDAOMock) ReplacePeriod(ctx context.Context, organizationID, periodID uint64, rows []KpiAccountSummary) error {
	if m.ReplacePeriodFn != nil {
		return m.ReplacePeriodFn(ctx, organizationID, periodID, rows)
	}
	return nil
}

func (m KpiSummaryDAOMock) HasPeriod(ctx context.Context, organizationID, periodID uint64) (bool, error) {
	if m.HasPeriodFn != nil {
		return m.HasPeriodFn(ctx, organizationID, periodID)
	}
	return false, nil
}

func (m KpiSummaryDAOMock) TrialBalance(ctx context.Context, organizationID, periodID uint64) ([]TrialBalanceRow, error) {
	if m.TrialBalanceFn != nil {
		return m.TrialBalanceFn(ctx, organizationID, periodID)
	}
	return nil, nil
}

func (m KpiSummaryDAOMock) ListPeriod(ctx context.Context, organizationID, periodID uint64) ([]KpiAccountSummary, error) {
	if m.ListPeriodFn != nil {
		return m.ListPeriodFn(ctx, organizationID, periodID)
	}
	return nil, nil
}
