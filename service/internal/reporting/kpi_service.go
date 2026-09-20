package reporting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/crm"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/project"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
)

type PipelineForecaster interface {
	Forecast(ctx context.Context, organizationID *uint64) (crm.PipelineForecast, error)
}

type SubscriptionMetricsProvider interface {
	Metrics(ctx context.Context, organizationID uint64) (subscription.Metrics, error)
}

type ProjectListReader interface {
	List(ctx context.Context, q *query.Query) (*query.Page[project.Project], error)
}

type ProjectSummarizer interface {
	Summary(ctx context.Context, projectID uint64) (project.ProjectSummary, error)
}

type KpiService struct {
	reports      ReportDAO
	pipeline     PipelineForecaster
	subscription SubscriptionMetricsProvider
	projects     ProjectListReader
	projectSvc   ProjectSummarizer
	classifier   accounting.AccountClassifier
}

func NewKpiService(
	reports ReportDAO,
	pipeline PipelineForecaster,
	subscription SubscriptionMetricsProvider,
	projects ProjectListReader,
	projectSvc ProjectSummarizer,
) KpiService {
	return KpiService{
		reports:      reports,
		pipeline:     pipeline,
		subscription: subscription,
		projects:     projects,
		projectSvc:   projectSvc,
		classifier:   accounting.GenericClassifier{},
	}
}

func (s KpiService) SetClassifier(classifier accounting.AccountClassifier) KpiService {
	s.classifier = classifier
	return s
}

type SalesKPI struct {
	Bookings       float64 `json:"bookings"`
	Revenue        float64 `json:"revenue"`
	COGS           float64 `json:"cogs"`
	GrossMargin    float64 `json:"gross_margin"`
	GrossMarginPct float64 `json:"gross_margin_pct"`
	WinRate        float64 `json:"win_rate"`
}

func (s KpiService) SalesKPI(ctx context.Context, organizationID uint64, start, end time.Time) (SalesKPI, error) {
	classifier := s.classifier
	bookings, err := s.reports.Bookings(ctx, organizationID, start, end)
	if err != nil {
		return SalesKPI{}, err
	}
	byType, err := s.reports.BalancesByType(ctx, organizationID, start, end)
	if err != nil {
		return SalesKPI{}, err
	}
	revenue := sumCreditSide(byType, classifier.IsIncome)
	cogs := sumDebitSide(byType, classifier.IsCOGS)
	forecast, err := s.pipeline.Forecast(ctx, &organizationID)
	if err != nil {
		return SalesKPI{}, err
	}
	kpi := SalesKPI{
		Bookings: bookings,
		Revenue:  revenue,
		COGS:     cogs,
		WinRate:  forecast.WinRate,
	}
	kpi.GrossMargin = revenue - cogs
	if revenue != 0 {
		kpi.GrossMarginPct = kpi.GrossMargin / revenue
	}
	return kpi, nil
}

type PipelineKPI struct {
	TotalExpectedRevenue float64 `json:"total_expected_revenue"`
	WeightedPipeline     float64 `json:"weighted_pipeline"`
	WinRate              float64 `json:"win_rate"`
	Stages               int     `json:"stages"`
}

func (s KpiService) PipelineKPI(ctx context.Context, organizationID uint64) (PipelineKPI, error) {
	forecast, err := s.pipeline.Forecast(ctx, &organizationID)
	if err != nil {
		return PipelineKPI{}, err
	}
	return PipelineKPI{
		TotalExpectedRevenue: forecast.TotalExpectedRevenue,
		WeightedPipeline:     forecast.WeightedPipeline,
		WinRate:              forecast.WinRate,
		Stages:               len(forecast.Stages),
	}, nil
}

type InventoryKPI struct {
	OnHandValue    float64 `json:"on_hand_value"`
	OnHandQuantity float64 `json:"on_hand_quantity"`
	ProductCount   int     `json:"product_count"`
}

func (s KpiService) InventoryKPI(ctx context.Context, organizationID uint64) (InventoryKPI, error) {
	rows, err := s.reports.InventoryValuation(ctx, organizationID)
	if err != nil {
		return InventoryKPI{}, err
	}
	kpi := InventoryKPI{}
	for _, row := range rows {
		kpi.OnHandValue += row.Value
		kpi.OnHandQuantity += row.Quantity
		kpi.ProductCount++
	}
	return kpi, nil
}

type SubscriptionKPI struct {
	MRR       float64 `json:"mrr"`
	ARR       float64 `json:"arr"`
	Churned   int     `json:"churned"`
	ChurnRate float64 `json:"churn_rate"`
	LTV       float64 `json:"ltv"`
}

func (s KpiService) SubscriptionKPI(ctx context.Context, organizationID uint64) (SubscriptionKPI, error) {
	metrics, err := s.subscription.Metrics(ctx, organizationID)
	if err != nil {
		return SubscriptionKPI{}, err
	}
	return SubscriptionKPI{
		MRR:       metrics.MRR,
		ARR:       metrics.ARR,
		Churned:   metrics.Churned,
		ChurnRate: metrics.ChurnRate,
		LTV:       metrics.LTV,
	}, nil
}

type ProjectKPI struct {
	ProjectCount int     `json:"project_count"`
	TotalMargin  float64 `json:"total_margin"`
	TotalCost    float64 `json:"total_cost"`
	TotalBilled  float64 `json:"total_billed"`
	Utilization  float64 `json:"utilization"`
}

func (s KpiService) ProjectKPI(ctx context.Context, organizationID uint64) (ProjectKPI, error) {
	projectPage, err := s.projects.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return ProjectKPI{}, err
	}
	kpi := ProjectKPI{}
	utilized := 0
	for _, p := range projectPage.Items {
		summary, err := s.projectSvc.Summary(ctx, p.ID)
		if err != nil {
			return ProjectKPI{}, err
		}
		kpi.ProjectCount++
		kpi.TotalMargin += summary.MarginAmount
		kpi.TotalCost += summary.CostAmount
		kpi.TotalBilled += summary.BilledAmount
		kpi.Utilization += summary.Utilization
		if summary.Utilization > 0 {
			utilized++
		}
	}
	if utilized > 0 {
		kpi.Utilization /= float64(utilized)
	}
	return kpi, nil
}

type PayrollKPI struct {
	GrossCost float64 `json:"gross_cost"`
	NetCost   float64 `json:"net_cost"`
}

func (s KpiService) PayrollKPI(ctx context.Context, organizationID uint64, start, end time.Time) (PayrollKPI, error) {
	gross, net, err := s.reports.PayrollCost(ctx, organizationID, start, end)
	if err != nil {
		return PayrollKPI{}, err
	}
	return PayrollKPI{GrossCost: gross, NetCost: net}, nil
}

type FinanceKPI struct {
	Revenue        float64 `json:"revenue"`
	Expenses       float64 `json:"expenses"`
	GrossMarginPct float64 `json:"gross_margin_pct"`
	NetMarginPct   float64 `json:"net_margin_pct"`
	EBITDA         float64 `json:"ebitda"`
	CurrentRatio   float64 `json:"current_ratio"`
}

func (s KpiService) FinanceKPI(ctx context.Context, organizationID uint64, start, end time.Time) (FinanceKPI, error) {
	classifier := s.classifier
	byType, err := s.reports.BalancesByType(ctx, organizationID, start, end)
	if err != nil {
		return FinanceKPI{}, err
	}
	revenue := sumCreditSide(byType, classifier.IsIncome)
	cogs := sumDebitSide(byType, classifier.IsCOGS)
	expense := sumDebitSide(byType, classifier.IsExpense)
	depreciation := sumDebitSide(byType, classifier.IsDepreciation)

	kpi := FinanceKPI{
		Revenue:  revenue,
		Expenses: cogs + expense + depreciation,
	}
	if revenue != 0 {
		kpi.GrossMarginPct = (revenue - cogs) / revenue
		kpi.NetMarginPct = (revenue - kpi.Expenses) / revenue
	}
	kpi.EBITDA = revenue - cogs - expense

	currentAssets := sumCreditSide(byType, classifier.IsCurrentAsset)
	currentLiabilities := sumDebitSide(byType, classifier.IsCurrentLiability)
	if currentLiabilities != 0 {
		kpi.CurrentRatio = currentAssets / currentLiabilities
	}
	return kpi, nil
}

type ProcurementKPI struct {
	AvgCycleDays      float64 `json:"avg_cycle_days"`
	OnTimeDeliveryPct float64 `json:"on_time_delivery_pct"`
	PriceVariancePct  float64 `json:"price_variance_pct"`
	PurchaseCount     int     `json:"purchase_count"`
}

func (s KpiService) ProcurementKPI(ctx context.Context, organizationID uint64, start, end time.Time) (ProcurementKPI, error) {
	metrics, err := s.reports.ProcurementMetrics(ctx, organizationID, start, end)
	if err != nil {
		return ProcurementKPI{}, err
	}
	kpi := ProcurementKPI{
		AvgCycleDays:  metrics.AvgCycleDays,
		PurchaseCount: metrics.Count,
	}
	if metrics.ReceivedCount > 0 {
		kpi.OnTimeDeliveryPct = float64(metrics.OnTimeCount) / float64(metrics.ReceivedCount)
	}
	if metrics.StandardCost != 0 {
		kpi.PriceVariancePct = (metrics.ActualCost - metrics.StandardCost) / metrics.StandardCost
	}
	return kpi, nil
}

type ManufacturingKPI struct {
	OEEPct          float64 `json:"oee_pct"`
	YieldPct        float64 `json:"yield_pct"`
	ScrapPct        float64 `json:"scrap_pct"`
	CostVariancePct float64 `json:"cost_variance_pct"`
	OrderCount      int     `json:"order_count"`
}

func (s KpiService) ManufacturingKPI(ctx context.Context, organizationID uint64, start, end time.Time) (ManufacturingKPI, error) {
	metrics, err := s.reports.ManufacturingMetrics(ctx, organizationID, start, end)
	if err != nil {
		return ManufacturingKPI{}, err
	}
	kpi := ManufacturingKPI{OrderCount: metrics.OrderCount}
	if metrics.ActualMinutes != 0 {
		kpi.OEEPct = metrics.PlannedMinutes / metrics.ActualMinutes
	}
	if metrics.QtyToProduce != 0 {
		kpi.YieldPct = metrics.QtyProduced / metrics.QtyToProduce
	}
	if metrics.PlannedMaterial != 0 {
		kpi.ScrapPct = (metrics.ConsumedMaterial - metrics.PlannedMaterial) / metrics.PlannedMaterial
	}
	if metrics.StandardValue != 0 {
		kpi.CostVariancePct = (metrics.ActualValue - metrics.StandardValue) / metrics.StandardValue
	}
	return kpi, nil
}

type ArApKPI struct {
	DSO          float64 `json:"dso"`
	DPO          float64 `json:"dpo"`
	OverdueARPct float64 `json:"overdue_ar_pct"`
	OverdueAPPct float64 `json:"overdue_ap_pct"`
}

func (s KpiService) ArApKPI(ctx context.Context, organizationID uint64, asOf time.Time) (ArApKPI, error) {
	metrics, err := s.reports.ArApMetrics(ctx, organizationID, asOf)
	if err != nil {
		return ArApKPI{}, err
	}
	kpi := ArApKPI{}
	days := float64(daysInYear(asOf))
	if metrics.Revenue != 0 {
		kpi.DSO = metrics.OpenAR / (metrics.Revenue / days)
	}
	if metrics.Purchases != 0 {
		kpi.DPO = metrics.OpenAP / (metrics.Purchases / days)
	}
	if metrics.OpenAR != 0 {
		kpi.OverdueARPct = metrics.OverdueAR / metrics.OpenAR
	}
	if metrics.OpenAP != 0 {
		kpi.OverdueAPPct = metrics.OverdueAP / metrics.OpenAP
	}
	return kpi, nil
}

type CashKPI struct {
	Position float64 `json:"position"`
	Burn     float64 `json:"burn"`
	Forecast float64 `json:"forecast"`
}

func (s KpiService) CashKPI(ctx context.Context, organizationID uint64, asOf time.Time) (CashKPI, error) {
	metrics, err := s.reports.CashMetrics(ctx, organizationID, asOf)
	if err != nil {
		return CashKPI{}, err
	}
	return CashKPI{
		Position: metrics.BankBalance,
		Burn:     metrics.Outflows - metrics.Inflows,
		Forecast: metrics.BankBalance + metrics.OpenAR - metrics.OpenAP,
	}, nil
}

type InventoryRatioKPI struct {
	Turnover      float64 `json:"turnover"`
	DaysOnHand    float64 `json:"days_on_hand"`
	StockoutCount int     `json:"stockout_count"`
}

func (s KpiService) InventoryRatioKPI(ctx context.Context, organizationID uint64, start, end time.Time) (InventoryRatioKPI, error) {
	metrics, err := s.reports.InventoryRatioMetrics(ctx, organizationID, start, end)
	if err != nil {
		return InventoryRatioKPI{}, err
	}
	kpi := InventoryRatioKPI{StockoutCount: metrics.StockoutCount}
	days := daysBetween(start, end)
	if metrics.AvgInventory != 0 && days > 0 {
		kpi.Turnover = metrics.COGS / metrics.AvgInventory
		kpi.DaysOnHand = days / (metrics.COGS / metrics.AvgInventory)
	}
	return kpi, nil
}

func daysBetween(start, end time.Time) float64 {
	return end.Sub(start).Hours() / 24
}

func daysInYear(asOf time.Time) float64 {
	if asOf.Month() == time.February && asOf.Day() == 29 {
		return 366
	}
	start := time.Date(asOf.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(asOf.Year(), 12, 31, 0, 0, 0, 0, time.UTC)
	return daysBetween(start, end) + 1
}

func sumCreditSide(rows []TypeBalanceRow, matches func(string) bool) float64 {
	var balance float64
	for _, row := range rows {
		if matches(row.AccountType) {
			balance += row.Credit - row.Debit
		}
	}
	return balance
}

func sumDebitSide(rows []TypeBalanceRow, matches func(string) bool) float64 {
	var balance float64
	for _, row := range rows {
		if matches(row.AccountType) {
			balance += row.Debit - row.Credit
		}
	}
	return balance
}
