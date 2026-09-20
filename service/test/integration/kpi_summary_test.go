//go:build integration

package integration

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/asset"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/reporting"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type summaryFixture struct {
	organizationID uint64
	cashAccountID  uint64
	revenueAccount uint64
	equityAccount  uint64
	journalID      uint64
	taxYearID      uint64
	taxPeriodSvc   accounting.TaxPeriodService
	postingSvc     accounting.PostingService
	reportDAO      reporting.ReportDAO
	summaryDAO     reporting.KpiSummaryDAO
	summarySvc     reporting.KpiSummaryService
	reportSvc      reporting.ReportService
}

func seedSummaryFixture(t *testing.T) *summaryFixture {
	t.Helper()
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	cash, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1100", Name: "Cash", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create cash account failed: %v", err)
	}
	revenueAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create revenue account failed: %v", err)
	}
	equity, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "3000", Name: "Equity", Type: "equity", Active: true,
	})
	if err != nil {
		t.Fatalf("create equity account failed: %v", err)
	}
	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "General", Code: helper.Ptr("GEN"),
		Type: "general", DefaultAccountID: helper.Ptr(cash.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(testDB), dao.NewBase[reference.TaxYear](testDB))
	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc)

	reportDAO := reporting.NewReportDAO(testDB)
	summaryDAO := reporting.NewKpiSummaryDAO(testDB)
	summarySvc := reporting.NewKpiSummaryService(summaryDAO, accounting.NewTaxPeriodDAO(testDB))
	reportSvc := reporting.NewReportService(reportDAO, accounting.NewTaxPeriodDAO(testDB), summaryDAO)

	return &summaryFixture{
		organizationID: org.ID,
		cashAccountID:  cash.ID,
		revenueAccount: revenueAccount.ID,
		equityAccount:  equity.ID,
		journalID:      journal.ID,
		taxPeriodSvc:   taxPeriodSvc,
		postingSvc:     postingSvc,
		reportDAO:      reportDAO,
		summaryDAO:     summaryDAO,
		summarySvc:     summarySvc,
		reportSvc:      reportSvc,
	}
}

func (f *summaryFixture) createPeriod(t *testing.T, name string, start, end time.Time) *accounting.TaxPeriod {
	t.Helper()
	ctx := testutil.SystemContext()
	if f.taxYearID == 0 {
		year, err := dao.NewBase[reference.TaxYear](testDB).Create(ctx, &reference.TaxYear{
			OrganizationID: helper.Ptr(f.organizationID), Name: "2026", State: helper.Ptr("open"),
		})
		if err != nil {
			t.Fatalf("create tax year failed: %v", err)
		}
		f.taxYearID = year.ID
	}
	period, err := f.taxPeriodSvc.Create(ctx, &accounting.TaxPeriod{
		OrganizationID: f.organizationID, TaxYearID: f.taxYearID, Name: name,
		DateStart: helper.Ptr(start), DateEnd: helper.Ptr(end),
	})
	if err != nil {
		t.Fatalf("create tax period %s failed: %v", name, err)
	}
	return period
}

func (f *summaryFixture) post(t *testing.T, date time.Time, description string, lines []accounting.PostingLine) {
	t.Helper()
	if _, err := f.postingSvc.Post(testutil.SystemContext(), accounting.PostRequest{
		OrganizationID: f.organizationID,
		JournalID:      f.journalID,
		Date:           date,
		Ref:            gofakeit.LetterN(8),
		Description:    description,
		Lines:          lines,
	}); err != nil {
		t.Fatalf("post movement failed: %v", err)
	}
}

func TestKpiSummaryReconcilesToSourceLedger(t *testing.T) {
	fixture := seedSummaryFixture(t)
	ctx := testutil.SystemContext()

	period0 := fixture.createPeriod(t, "Jan 2026",
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC))
	period1 := fixture.createPeriod(t, "Feb 2026",
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC))

	fixture.post(t, time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), "opening", []accounting.PostingLine{
		{AccountID: fixture.cashAccountID, Name: "Opening cash", Debit: amount.FromFloat64(500)},
		{AccountID: fixture.equityAccount, Name: "Opening equity", Credit: amount.FromFloat64(500)},
	})
	fixture.post(t, time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC), "sale", []accounting.PostingLine{
		{AccountID: fixture.cashAccountID, Name: "Cash receipt", Debit: amount.FromFloat64(100)},
		{AccountID: fixture.revenueAccount, Name: "Revenue", Credit: amount.FromFloat64(100)},
	})

	processed, err := fixture.summarySvc.RefreshPeriod(ctx, fixture.organizationID, period1.ID)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if processed != 3 {
		t.Errorf("processed accounts = %d, want 3 (cash, equity, revenue)", processed)
	}

	has, err := fixture.summaryDAO.HasPeriod(ctx, fixture.organizationID, period1.ID)
	if err != nil {
		t.Fatalf("has period failed: %v", err)
	}
	if !has {
		t.Fatalf("period %d has no summary", period1.ID)
	}

	summaries, err := fixture.summaryDAO.ListPeriod(ctx, fixture.organizationID, period1.ID)
	if err != nil {
		t.Fatalf("list summaries failed: %v", err)
	}
	byAccount := map[uint64]reporting.KpiAccountSummary{}
	for _, row := range summaries {
		byAccount[row.AccountID] = row
	}
	if got := byAccount[fixture.cashAccountID]; got.OpeningDebit != 500 || got.PeriodDebit != 100 {
		t.Errorf("cash summary = %+v, want opening 500 period 100", got)
	}
	if got := byAccount[fixture.equityAccount]; got.OpeningCredit != 500 || got.PeriodDebit != 0 {
		t.Errorf("equity summary = %+v, want opening credit 500", got)
	}
	if got := byAccount[fixture.revenueAccount]; got.OpeningCredit != 0 || got.PeriodCredit != 100 {
		t.Errorf("revenue summary = %+v, want period credit 100", got)
	}

	start1, end1 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC)
	rawRows, err := fixture.reportDAO.TrialBalance(ctx, fixture.organizationID, start1, end1)
	if err != nil {
		t.Fatalf("raw trial balance failed: %v", err)
	}
	summaryRows, err := fixture.summaryDAO.TrialBalance(ctx, fixture.organizationID, period1.ID)
	if err != nil {
		t.Fatalf("summary trial balance failed: %v", err)
	}
	if len(rawRows) != len(summaryRows) {
		t.Fatalf("trial balance rows = raw %d vs summary %d, want equal", len(rawRows), len(summaryRows))
	}
	rawByAccount := map[uint64]reporting.TrialBalanceRow{}
	for _, row := range rawRows {
		rawByAccount[row.AccountID] = row
	}
	for _, row := range summaryRows {
		raw, ok := rawByAccount[row.AccountID]
		if !ok {
			t.Errorf("summary row for account %d missing from raw trial balance", row.AccountID)
			continue
		}
		if !closeFloats(raw.OpeningDebit, row.OpeningDebit) || !closeFloats(raw.OpeningCredit, row.OpeningCredit) ||
			!closeFloats(raw.PeriodDebit, row.PeriodDebit) || !closeFloats(raw.PeriodCredit, row.PeriodCredit) ||
			!closeFloats(raw.ClosingDebit, row.ClosingDebit) || !closeFloats(raw.ClosingCredit, row.ClosingCredit) {
			t.Errorf("account %d summary = %+v, raw = %+v, want reconcile", row.AccountID, row, raw)
		}
	}

	reportRows, err := fixture.reportSvc.TrialBalance(ctx, fixture.organizationID, period1.ID)
	if err != nil {
		t.Fatalf("report trial balance failed: %v", err)
	}
	if len(reportRows.Rows) != len(summaryRows) {
		t.Errorf("report trial balance rows = %d, want %d (summary-backed)", len(reportRows.Rows), len(summaryRows))
	}

	if _, err := fixture.summarySvc.RefreshPeriod(ctx, fixture.organizationID, period1.ID); err != nil {
		t.Fatalf("second refresh failed: %v", err)
	}
	afterRedo, err := fixture.summaryDAO.ListPeriod(ctx, fixture.organizationID, period1.ID)
	if err != nil {
		t.Fatalf("list summaries after redo failed: %v", err)
	}
	if len(afterRedo) != len(summaries) {
		t.Errorf("summaries after refresh = %d, want %d (idempotent)", len(afterRedo), len(summaries))
	}

	if _, err := fixture.postingSvc.Post(testutil.SystemContext(), accounting.PostRequest{
		OrganizationID: fixture.organizationID,
		JournalID:      fixture.journalID,
		Date:           time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
		Ref:            gofakeit.LetterN(8),
		Description:    "late movement",
		Lines: []accounting.PostingLine{
			{AccountID: fixture.cashAccountID, Name: "Cash", Debit: amount.FromFloat64(25)},
			{AccountID: fixture.revenueAccount, Name: "Revenue", Credit: amount.FromFloat64(25)},
		},
	}); err != nil {
		t.Fatalf("post late movement failed: %v", err)
	}
	if _, err := fixture.summarySvc.RefreshPeriod(ctx, fixture.organizationID, period1.ID); err != nil {
		t.Fatalf("refresh after new movement failed: %v", err)
	}
	updated, err := fixture.summaryDAO.ListPeriod(ctx, fixture.organizationID, period1.ID)
	if err != nil {
		t.Fatalf("list updated summaries failed: %v", err)
	}
	for _, row := range updated {
		if row.AccountID == fixture.cashAccountID && row.PeriodDebit != 125 {
			t.Errorf("cash period debit after refresh = %v, want 125", row.PeriodDebit)
		}
	}
	_ = period0
}

func TestPeriodCloseRefreshesKpiSummary(t *testing.T) {
	fixture := seedSummaryFixture(t)
	ctx := testutil.SystemContext()

	period := fixture.createPeriod(t, "Feb 2026",
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 28, 0, 0, 0, 0, time.UTC))

	raw, err := json.Marshal(fixture.journalID)
	if err != nil {
		t.Fatalf("marshal journal config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(fixture.organizationID), Key: "reporting.journal_id", Value: raw,
	}); err != nil {
		t.Fatalf("create journal config failed: %v", err)
	}
	fxRaw, err := json.Marshal(fixture.revenueAccount)
	if err != nil {
		t.Fatalf("marshal fx config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(fixture.organizationID), Key: "reporting.fx_gain_loss_account_id", Value: fxRaw,
	}); err != nil {
		t.Fatalf("create fx config failed: %v", err)
	}

	fixture.post(t, time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC), "sale", []accounting.PostingLine{
		{AccountID: fixture.cashAccountID, Name: "Cash receipt", Debit: amount.FromFloat64(100)},
		{AccountID: fixture.revenueAccount, Name: "Revenue", Credit: amount.FromFloat64(100)},
	})

	reportConfigSource := reporting.NewConfigSource(dao.NewBase[reference.SystemConfig](testDB))
	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	journalEntryLineDAO := accounting.NewJournalLineDAO(testDB)
	taxPeriodDAO := accounting.NewTaxPeriodDAO(testDB)
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(fixture.taxPeriodSvc).
		SetReverser(accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO))

	fxSvc := reporting.NewFxRevaluationService(
		fixture.reportDAO,
		reporting.NewFxRevaluationDAO(testDB),
		reporting.NewFxRevaluationLineDAO(testDB),
		postingSvc,
		reportConfigSource,
		reference.NewFxRateSource(dao.NewBase[reference.FxRate](testDB)),
		dao.NewBase[reference.Organization](testDB),
	)
	accrualSvc := reporting.NewAccrualService(
		reporting.NewAccrualDAO(testDB),
		reporting.NewAccrualLineDAO(testDB),
		postingSvc,
		reportConfigSource,
	)
	deferralSvc := accounting.NewDeferralService(
		accounting.NewDeferredScheduleDAO(testDB),
		accounting.NewDeferredScheduleLineDAO(testDB),
		postingSvc,
		subscription.NewSubscriptionConfigSource(dao.NewBase[reference.SystemConfig](testDB)),
		db.NewDBTransactioner(testDB),
	)
	summarySvc := reporting.NewKpiSummaryService(fixture.summaryDAO, taxPeriodDAO)
	closeSvc := reporting.NewPeriodCloseService(
		fixture.taxPeriodSvc,
		taxPeriodDAO,
		reportConfigSource,
		asset.NewFixedAssetDAO(testDB),
		asset.NewAssetService(
			asset.NewFixedAssetDAO(testDB),
			asset.NewAssetDepreciationLineDAO(testDB),
			dao.NewBase[reference.AssetCategory](testDB),
			accounting.NewInvoiceLineDAO(testDB),
			accounting.NewInvoiceDAO(testDB),
			postingSvc,
			db.NewDBTransactioner(testDB),
		),
		fxSvc,
		accrualSvc,
		deferralSvc,
		summarySvc,
	)

	result, err := closeSvc.Close(ctx, fixture.organizationID, period.ID)
	if err != nil {
		t.Fatalf("period close failed: %v", err)
	}
	if result.SummarizedAccounts != 2 {
		t.Errorf("summarized accounts = %d, want 2 (cash, revenue)", result.SummarizedAccounts)
	}

	has, err := fixture.summaryDAO.HasPeriod(ctx, fixture.organizationID, period.ID)
	if err != nil {
		t.Fatalf("has period failed: %v", err)
	}
	if !has {
		t.Fatalf("no summary after period close")
	}

	closed, err := taxPeriodDAO.Find(ctx, period.ID)
	if err != nil {
		t.Fatalf("find period failed: %v", err)
	}
	if closed == nil || closed.State != accounting.TaxPeriodStateLocked {
		t.Fatalf("period state = %v, want locked", closed)
	}

	reportRows, err := fixture.reportSvc.TrialBalance(ctx, fixture.organizationID, period.ID)
	if err != nil {
		t.Fatalf("report trial balance failed: %v", err)
	}
	if len(reportRows.Rows) != 2 {
		t.Errorf("trial balance rows = %d, want 2 (summary-backed after close)", len(reportRows.Rows))
	}

	if _, err := closeSvc.Close(ctx, fixture.organizationID, period.ID); err == nil {
		t.Fatal("second close should fail on locked period")
	}
}

func closeFloats(a, b float64) bool {
	return math.Abs(a-b) < 0.0001
}
