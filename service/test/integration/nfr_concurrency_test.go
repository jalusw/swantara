//go:build integration

package integration

import (
	"errors"
	"sync"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type nfrInventoryFixture struct {
	orgID      uint64
	locationID uint64
	variantID  uint64
}

func seedNfrInventoryFixture(t *testing.T) nfrInventoryFixture {
	t.Helper()
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	warehouse, err := dao.NewBase[reference.Warehouse](testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(org.ID), Name: "Main",
	})
	if err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}
	location, err := dao.NewBase[reference.StockLocation](testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), WarehouseID: helper.Ptr(warehouse.ID),
		Name: "Stock", Usage: "internal",
	})
	if err != nil {
		t.Fatalf("create location failed: %v", err)
	}
	template, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID), Name: "Concurrent Item", CategoryID: nil,
		Type: "stockable", StandardCost: 10, IsPurchasable: true, IsSellable: true, Tracking: "none",
	})
	if err != nil {
		t.Fatalf("create template failed: %v", err)
	}
	variant, err := products.NewItemVariantDAO(testDB).Create(ctx, &products.ItemVariant{
		ItemID: template.ID, Active: true,
	})
	if err != nil {
		t.Fatalf("create variant failed: %v", err)
	}

	return nfrInventoryFixture{orgID: org.ID, locationID: location.ID, variantID: variant.ID}
}

func TestConcurrentStockHoldNoLostUpdates(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedNfrInventoryFixture(t)
	ctx := testutil.SystemContext()

	quantDAO := inventory.NewStockBalanceDAO(testDB)
	if _, err := quantDAO.Create(ctx, &inventory.StockBalance{
		OrganizationID: helper.Ptr(fx.orgID), ItemID: fx.variantID, LocationID: fx.locationID, Quantity: 100,
	}); err != nil {
		t.Fatalf("create quant failed: %v", err)
	}

	reservationSvc := inventory.NewHoldService(
		inventory.NewStockHoldDAO(testDB),
		quantDAO,
	)

	workers := 10
	reservePerWorker := amount.FromFloat64(12)
	var successes int
	var overflows int
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := reservationSvc.Reserve(ctx, fx.orgID, fx.variantID, fx.locationID, nil, reservePerWorker, nil)
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successes++
			} else if errors.Is(err, inventory.ErrHoldOverflow) {
				overflows++
			} else {
				t.Errorf("reserve failed with unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if successes != 8 || overflows != 2 {
		t.Errorf("successes/overflows = %d/%d, want 8/2", successes, overflows)
	}
	quant, err := quantDAO.FindByKey(ctx, fx.variantID, fx.locationID, nil)
	if err != nil {
		t.Fatalf("find quant failed: %v", err)
	}
	if quant.ReservedQty != 96 {
		t.Errorf("reserved qty = %v, want 96 (no lost updates)", quant.ReservedQty)
	}
}

func TestConcurrentStockBalanceUpsertNoLostUpdates(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedNfrInventoryFixture(t)
	ctx := testutil.SystemContext()

	quantDAO := inventory.NewStockBalanceDAO(testDB)
	if _, err := quantDAO.Create(ctx, &inventory.StockBalance{
		OrganizationID: helper.Ptr(fx.orgID), ItemID: fx.variantID, LocationID: fx.locationID, Quantity: 0,
	}); err != nil {
		t.Fatalf("create quant failed: %v", err)
	}

	workers := 10
	delta := 5.0
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := quantDAO.Upsert(ctx, helper.Ptr(fx.orgID), fx.variantID, fx.locationID, nil, delta); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("upsert failed: %v", err)
	}

	quant, err := quantDAO.FindByKey(ctx, fx.variantID, fx.locationID, nil)
	if err != nil {
		t.Fatalf("find quant failed: %v", err)
	}
	if quant.Quantity != delta*float64(workers) {
		t.Errorf("quantity = %v, want %v (no lost updates)", quant.Quantity, delta*float64(workers))
	}
}

func TestConcurrentPostingAtomicAndComplete(t *testing.T) {
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
		OrganizationID: org.ID, Name: "NFR Journal", Code: helper.Ptr("NFR"),
		Type: "general", DefaultAccountID: helper.Ptr(debit.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	postingSvc := accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB))

	workers := 5
	postErr := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			_, err := postingSvc.Post(ctx, accounting.PostRequest{
				OrganizationID: org.ID,
				JournalID:      journal.ID,
				Date:           org.CreatedAt.UTC(),
				Ref:            "NFR",
				Description:    "Concurrent posting",
				Lines: []accounting.PostingLine{
					{AccountID: debit.ID, Name: "Expense", Debit: amount.FromFloat64(100)},
					{AccountID: credit.ID, Name: "Accrued", Credit: amount.FromFloat64(100)},
				},
			})
			postErr <- err
		}(i)
	}
	wg.Wait()
	close(postErr)
	for err := range postErr {
		if err != nil {
			t.Errorf("concurrent post failed: %v", err)
		}
	}

	var posted int64
	if err := testDB.WithContext(ctx).Model(&accounting.JournalEntry{}).Count(&posted).Error; err != nil {
		t.Fatalf("count movements failed: %v", err)
	}
	if posted != int64(workers) {
		t.Errorf("posted movements = %d, want %d", posted, workers)
	}
	var lineCount int64
	if err := testDB.WithContext(ctx).Model(&accounting.JournalLine{}).Count(&lineCount).Error; err != nil {
		t.Fatalf("count movement lines failed: %v", err)
	}
	if lineCount != posted*2 {
		t.Errorf("movement lines = %d, want %d (no partial writes under concurrency)", lineCount, posted*2)
	}
}
