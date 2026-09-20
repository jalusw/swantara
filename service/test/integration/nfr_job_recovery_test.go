//go:build integration

package integration

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/jobs"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/subscription"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestScheduledJobInterruptedRunResumesWithoutDuplicates(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	orgA, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization A failed: %v", err)
	}
	orgB, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization B failed: %v", err)
	}

	bsAccountA, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: orgA.ID, Code: "2400", Name: "Deferred Revenue", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create bs account A failed: %v", err)
	}
	plAccountA, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: orgA.ID, Code: "4100", Name: "Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create pl account A failed: %v", err)
	}
	bsAccountB, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: orgB.ID, Code: "2400", Name: "Deferred Revenue", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create bs account B failed: %v", err)
	}
	plAccountB, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: orgB.ID, Code: "4100", Name: "Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create pl account B failed: %v", err)
	}

	journalA, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: orgA.ID, Name: "Deferral Journal A", Code: helper.Ptr("DEFA"),
		Type: "general", DefaultAccountID: helper.Ptr(bsAccountA.ID),
	})
	if err != nil {
		t.Fatalf("create journal A failed: %v", err)
	}
	raw, err := json.Marshal(journalA.ID)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(orgA.ID), Key: "subscription.journal_id", Value: raw,
	}); err != nil {
		t.Fatalf("create config A failed: %v", err)
	}

	postingSvc := accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB))
	deferralSvc := accounting.NewDeferralService(
		accounting.NewDeferredScheduleDAO(testDB),
		accounting.NewDeferredScheduleLineDAO(testDB),
		postingSvc,
		subscription.NewSubscriptionConfigSource(dao.NewBase[reference.SystemConfig](testDB)),
		db.NewDBTransactioner(testDB),
	)

	asOf := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)
	dateStart := asOf.AddDate(0, -2, 0)
	totalDue := 0
	createOrgSchedules := func(orgID uint64, bsAccountID, plAccountID uint64, count int) {
		for i := 0; i < count; i++ {
			if _, err := deferralSvc.Create(ctx, accounting.CreateScheduleRequest{
				OrganizationID:        orgID,
				Type:                  accounting.DeferredTypeDeferredRevenue,
				SourceType:            "recovery_test",
				SourceID:              uint64(i + 1),
				TotalAmount:           200,
				BalanceSheetAccountID: bsAccountID,
				PLAccountID:           plAccountID,
				Method:                accounting.DeferredMethodLinear,
				DateStart:             dateStart,
				DateEnd:               dateStart.AddDate(0, 2, 0),
				Periods:               2,
			}); err != nil {
				t.Fatalf("create schedule for org failed: %v", err)
			}
			totalDue += 2
		}
	}
	createOrgSchedules(orgA.ID, bsAccountA.ID, plAccountA.ID, 3)
	createOrgSchedules(orgB.ID, bsAccountB.ID, plAccountB.ID, 2)

	jobRuns := jobs.NewService(jobs.NewJobRunDAO(testDB))

	firstRun, err := jobRuns.Start(ctx, jobs.TypeDeferralRecognition)
	if err != nil {
		t.Fatalf("start job run failed: %v", err)
	}
	postedFirst, firstErr := deferralSvc.RecognizeDue(ctx, nil, asOf)
	if !errors.Is(firstErr, accounting.ErrScheduleNoJournal) {
		t.Fatalf("first run error = %v, want ErrScheduleNoJournal (org B has no journal config)", firstErr)
	}
	if err := jobRuns.Fail(ctx, firstRun, firstErr.Error()); err != nil {
		t.Fatalf("fail job run failed: %v", err)
	}

	journalB, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: orgB.ID, Name: "Deferral Journal B", Code: helper.Ptr("DEFB"),
		Type: "general", DefaultAccountID: helper.Ptr(bsAccountB.ID),
	})
	if err != nil {
		t.Fatalf("create journal B failed: %v", err)
	}
	rawB, err := json.Marshal(journalB.ID)
	if err != nil {
		t.Fatalf("marshal config B failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(orgB.ID), Key: "subscription.journal_id", Value: rawB,
	}); err != nil {
		t.Fatalf("create config B failed: %v", err)
	}

	secondRun, err := jobRuns.Start(ctx, jobs.TypeDeferralRecognition)
	if err != nil {
		t.Fatalf("start job run failed: %v", err)
	}
	postedSecond, secondErr := deferralSvc.RecognizeDue(ctx, nil, asOf)
	if secondErr != nil {
		t.Fatalf("resumed run failed: %v", secondErr)
	}
	if err := jobRuns.Complete(ctx, secondRun, postedSecond); err != nil {
		t.Fatalf("complete job run failed: %v", err)
	}

	if postedFirst+postedSecond != totalDue {
		t.Errorf("posted across both runs = %d + %d = %d, want %d", postedFirst, postedSecond, postedFirst+postedSecond, totalDue)
	}

	var movements int64
	if err := testDB.WithContext(ctx).Model(&accounting.JournalEntry{}).Where("origin_type = ?", accounting.DeferredTypeDeferredRevenue).Count(&movements).Error; err != nil {
		t.Fatalf("count movements failed: %v", err)
	}
	if movements != int64(totalDue) {
		t.Errorf("account movements = %d, want %d (one per due line, no duplicates)", movements, totalDue)
	}

	replayed, err := deferralSvc.RecognizeDue(ctx, nil, asOf)
	if err != nil {
		t.Fatalf("third run failed: %v", err)
	}
	if replayed != 0 {
		t.Errorf("third run posted = %d, want 0", replayed)
	}

	runsPage, err := jobs.NewJobRunDAO(testDB).List(ctx, &query.Query{
		Filters: []query.Filter{{Field: "job_type", Operator: query.Equal, Value: jobs.TypeDeferralRecognition}},
	})
	if err != nil {
		t.Fatalf("list job runs failed: %v", err)
	}
	if len(runsPage.Items) != 2 {
		t.Fatalf("job runs = %d, want 2", len(runsPage.Items))
	}
	statuses := map[string]int{}
	for _, run := range runsPage.Items {
		statuses[run.Status]++
	}
	if statuses[jobs.StatusFailed] != 1 || statuses[jobs.StatusCompleted] != 1 {
		t.Errorf("job run statuses = %v, want one failed + one completed", statuses)
	}
}
