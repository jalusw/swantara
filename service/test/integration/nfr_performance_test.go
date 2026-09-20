//go:build integration

package integration

import (
	"sort"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

const (
	stockLookupP95Target  = 300 * time.Millisecond
	postingP95Target      = 2 * time.Second
	stockLookupIterations = 500
	postingIterations     = 100
)

func p95(durations []time.Duration) time.Duration {
	sorted := append([]time.Duration(nil), durations...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	index := int(float64(len(sorted))*0.95) - 1
	if index < 0 {
		index = 0
	}
	return sorted[index]
}

func TestStockLookupP95WithinTarget(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedNfrInventoryFixture(t)
	ctx := testutil.SystemContext()

	quantDAO := inventory.NewStockBalanceDAO(testDB)
	if _, err := quantDAO.Create(ctx, &inventory.StockBalance{
		OrganizationID: helper.Ptr(fx.orgID), ItemID: fx.variantID, LocationID: fx.locationID, Quantity: 10,
	}); err != nil {
		t.Fatalf("create quant failed: %v", err)
	}

	for i := 0; i < 20; i++ {
		if _, err := quantDAO.FindByKey(ctx, fx.variantID, fx.locationID, nil); err != nil {
			t.Fatalf("warmup lookup failed: %v", err)
		}
	}

	durations := make([]time.Duration, 0, stockLookupIterations)
	for i := 0; i < stockLookupIterations; i++ {
		start := time.Now()
		quant, err := quantDAO.FindByKey(ctx, fx.variantID, fx.locationID, nil)
		durations = append(durations, time.Since(start))
		if err != nil {
			t.Fatalf("lookup failed: %v", err)
		}
		if quant == nil {
			t.Fatal("quant not found")
		}
	}

	measured := p95(durations)
	if measured > stockLookupP95Target {
		t.Errorf("stock lookup p95 = %v, want < %v (NFR-PER-001)", measured, stockLookupP95Target)
	}
}

func TestPostingPathP95WithinTarget(t *testing.T) {
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	debit, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "5100", Name: "Expense", Type: "expense", Active: true,
	})
	if err != nil {
		t.Fatalf("create debit account failed: %v", err)
	}
	credit, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2100", Name: "Accrued", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create credit account failed: %v", err)
	}
	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Perf Journal", Code: helper.Ptr("PERF"),
		Type: "general", DefaultAccountID: helper.Ptr(debit.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	postingSvc := accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB))
	durations := make([]time.Duration, 0, postingIterations)
	for i := 0; i < postingIterations; i++ {
		start := time.Now()
		movement, err := postingSvc.Post(ctx, accounting.PostRequest{
			OrganizationID: org.ID,
			JournalID:      journal.ID,
			Date:           time.Now().UTC(),
			Ref:            "PERF",
			Description:    "Performance posting",
			Lines: []accounting.PostingLine{
				{AccountID: debit.ID, Name: "Expense", Debit: amount.FromFloat64(10)},
				{AccountID: credit.ID, Name: "Accrued", Credit: amount.FromFloat64(10)},
			},
		})
		durations = append(durations, time.Since(start))
		if err != nil {
			t.Fatalf("post failed: %v", err)
		}
		if movement.ID == 0 {
			t.Fatal("post produced no movement")
		}
	}

	measured := p95(durations)
	if measured > postingP95Target {
		t.Errorf("posting p95 = %v, want < %v (NFR-PER-003)", measured, postingP95Target)
	}
}
