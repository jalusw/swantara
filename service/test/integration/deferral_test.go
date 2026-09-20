//go:build integration

package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestDeferralScheduleCreateAndRecognizeIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	bsAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2400", Name: "Deferred Revenue", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create bs account failed: %v", err)
	}
	plAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Sales Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create pl account failed: %v", err)
	}

	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Deferral Journal", Code: helper.Ptr("DEF"),
		Type: "general", DefaultAccountID: helper.Ptr(bsAccount.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	raw, err := json.Marshal(journal.ID)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(org.ID), Key: "subscription.journal_id", Value: raw,
	}); err != nil {
		t.Fatalf("create system config failed: %v", err)
	}

	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	journalEntryLineDAO := accounting.NewJournalLineDAO(testDB)
	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(testDB), dao.NewBase[reference.TaxYear](testDB))
	reversalEngine := accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO)
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine)

	configSource := subscription.NewSubscriptionConfigSource(dao.NewBase[reference.SystemConfig](testDB))
	deferralSvc := accounting.NewDeferralService(
		accounting.NewDeferredScheduleDAO(testDB),
		accounting.NewDeferredScheduleLineDAO(testDB),
		postingSvc,
		configSource,
		db.NewDBTransactioner(testDB),
	)

	dateStart := time.Now().UTC().Truncate(24 * time.Hour)
	schedule, err := deferralSvc.Create(ctx, accounting.CreateScheduleRequest{
		OrganizationID:        org.ID,
		Type:                  accounting.DeferredTypeDeferredRevenue,
		SourceType:            "invoice_line",
		SourceID:              101,
		ContactID:             nil,
		ItemID:                nil,
		TotalAmount:           300,
		BalanceSheetAccountID: bsAccount.ID,
		PLAccountID:           plAccount.ID,
		Method:                accounting.DeferredMethodLinear,
		DateStart:             dateStart,
		DateEnd:               dateStart.AddDate(0, 3, 0),
		Periods:               3,
	})
	if err != nil {
		t.Fatalf("create schedule failed: %v", err)
	}
	if schedule.State != accounting.DeferredStateRunning || schedule.TotalAmount != 300 {
		t.Errorf("schedule = %+v, want running of 300", schedule)
	}

	lines, err := deferralSvc.ListLines(ctx, schedule.ID)
	if err != nil {
		t.Fatalf("list lines failed: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(lines))
	}
	total := 0.0
	for _, line := range lines {
		total += line.Amount
		if line.Posted {
			t.Errorf("line %d should not be posted yet", line.ID)
		}
	}
	if total != 300 {
		t.Errorf("lines total = %v, want 300", total)
	}

	posted, err := deferralSvc.RecognizeDue(ctx, nil, dateStart.AddDate(0, 4, 0))
	if err != nil {
		t.Fatalf("recognize failed: %v", err)
	}
	if posted != 3 {
		t.Fatalf("recognized = %d, want 3", posted)
	}

	updated, err := deferralSvc.Get(ctx, org.ID, schedule.ID)
	if err != nil {
		t.Fatalf("get schedule failed: %v", err)
	}
	if updated.State != accounting.DeferredStateDone || updated.RecognizedAmount != 300 {
		t.Errorf("schedule = %+v, want done with 300 recognized", updated)
	}

	postedLines, err := deferralSvc.ListLines(ctx, schedule.ID)
	if err != nil {
		t.Fatalf("list posted lines failed: %v", err)
	}
	for _, line := range postedLines {
		if !line.Posted || line.MovementID == nil {
			t.Errorf("line %d should be posted with a movement", line.ID)
		}
		moveLines, err := journalEntryLineDAO.ListByMovement(ctx, *line.MovementID)
		if err != nil {
			t.Fatalf("list movement lines failed: %v", err)
		}
		if len(moveLines) != 2 {
			t.Errorf("movement %d lines = %d, want 2", *line.MovementID, len(moveLines))
		}
	}
}
