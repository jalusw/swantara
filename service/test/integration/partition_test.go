//go:build integration

package integration

import (
	"testing"
	"time"

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

type partitionFixture struct {
	organizationID   uint64
	accountID        uint64
	revenueAccount   uint64
	journalID        uint64
	postingSvc       accounting.PostingService
	itemID           uint64
	srcLocationID    uint64
	dstLocationID    uint64
	stockMovementDAO inventory.StockMovementDAO
}

func seedPartitionFixture(t *testing.T) *partitionFixture {
	t.Helper()
	testutil.CleanTables(t, testDB)
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "IDR", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}
	account, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1100", Name: "Cash", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create cash account failed: %v", err)
	}
	revenue, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create revenue account failed: %v", err)
	}
	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "General", Code: helper.Ptr("GEN"),
		Type: "general", DefaultAccountID: helper.Ptr(account.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}
	postingSvc := accounting.NewPostingService(accounting.NewJournalEntryDAO(testDB))

	warehouse, err := dao.NewBase[reference.Warehouse](testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(org.ID), Name: "Main",
	})
	if err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}
	src, err := dao.NewBase[reference.StockLocation](testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), WarehouseID: helper.Ptr(warehouse.ID),
		Name: "Stock", Usage: "internal",
	})
	if err != nil {
		t.Fatalf("create stock location failed: %v", err)
	}
	dst, err := dao.NewBase[reference.StockLocation](testDB).Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), WarehouseID: helper.Ptr(warehouse.ID),
		Name: "Production", Usage: "production",
	})
	if err != nil {
		t.Fatalf("create production location failed: %v", err)
	}
	template, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID), Name: "Widget", Type: "stockable",
		StandardCost: 50, IsPurchasable: true, IsSellable: true, Tracking: "none",
	})
	if err != nil {
		t.Fatalf("create item template failed: %v", err)
	}
	variant, err := products.NewItemVariantDAO(testDB).Create(ctx, &products.ItemVariant{
		ItemID: template.ID, Active: true,
	})
	if err != nil {
		t.Fatalf("create item variant failed: %v", err)
	}

	return &partitionFixture{
		organizationID:   org.ID,
		accountID:        account.ID,
		revenueAccount:   revenue.ID,
		journalID:        journal.ID,
		postingSvc:       postingSvc,
		itemID:           variant.ID,
		srcLocationID:    src.ID,
		dstLocationID:    dst.ID,
		stockMovementDAO: inventory.NewStockMovementDAO(testDB),
	}
}

func (f *partitionFixture) post(t *testing.T, date time.Time, value float64) *accounting.JournalEntry {
	t.Helper()
	movement, err := f.postingSvc.Post(testutil.SystemContext(), accounting.PostRequest{
		OrganizationID: f.organizationID,
		JournalID:      f.journalID,
		Date:           date,
		Ref:            gofakeit.LetterN(8),
		Description:    gofakeit.Sentence(4),
		Lines: []accounting.PostingLine{
			{AccountID: f.accountID, Debit: amount.FromFloat64(value)},
			{AccountID: f.revenueAccount, Credit: amount.FromFloat64(value)},
		},
	})
	if err != nil {
		t.Fatalf("post movement failed: %v", err)
	}
	return movement
}

func partitionRowCount(t *testing.T, table string) int64 {
	t.Helper()
	var count int64
	if err := testDB.WithContext(testutil.SystemContext()).Raw("SELECT count(*) FROM " + table).Scan(&count).Error; err != nil {
		t.Fatalf("count %s failed: %v", table, err)
	}
	return count
}

func TestJournalEntrysAndLinesRouteToMonthlyPartitions(t *testing.T) {
	f := seedPartitionFixture(t)

	july := f.post(t, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), 100)
	f.post(t, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC), 200)
	f.post(t, time.Date(2032, 1, 10, 0, 0, 0, 0, time.UTC), 300)

	if got := partitionRowCount(t, "journal_entrys_202607"); got != 1 {
		t.Fatalf("expected 1 movement in journal_entrys_202607, got %d", got)
	}
	if got := partitionRowCount(t, "journal_entrys_202608"); got != 1 {
		t.Fatalf("expected 1 movement in journal_entrys_202608, got %d", got)
	}
	if got := partitionRowCount(t, "journal_entrys_default"); got != 1 {
		t.Fatalf("expected 1 movement in journal_entrys_default, got %d", got)
	}
	if got := partitionRowCount(t, "journal_lines_202607"); got != 2 {
		t.Fatalf("expected 2 lines in journal_lines_202607, got %d", got)
	}
	if got := partitionRowCount(t, "journal_lines_202608"); got != 2 {
		t.Fatalf("expected 2 lines in journal_lines_202608, got %d", got)
	}
	if got := partitionRowCount(t, "journal_lines_default"); got != 2 {
		t.Fatalf("expected 2 lines in journal_lines_default, got %d", got)
	}

	lines, err := accounting.NewJournalLineDAO(testDB).ListByMovement(testutil.SystemContext(), july.ID)
	if err != nil {
		t.Fatalf("list lines by movement failed: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines on movement, got %d", len(lines))
	}
	for _, line := range lines {
		if !line.Date.Equal(july.Date) {
			t.Fatalf("expected line date %v to match movement date %v", line.Date, july.Date)
		}
	}
}

func TestStockMovementsRouteToMonthlyPartitions(t *testing.T) {
	f := seedPartitionFixture(t)
	ctx := testutil.SystemContext()

	dated := time.Date(2026, 9, 10, 10, 0, 0, 0, time.UTC)
	if _, err := f.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(f.organizationID),
		ItemID:         f.itemID,
		Qty:            1,
		SrcLocationID:  f.srcLocationID,
		DstLocationID:  f.dstLocationID,
		State:          "done",
		DateDone:       &dated,
	}); err != nil {
		t.Fatalf("create dated stock movement failed: %v", err)
	}
	if _, err := f.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(f.organizationID),
		ItemID:         f.itemID,
		Qty:            2,
		SrcLocationID:  f.srcLocationID,
		DstLocationID:  f.dstLocationID,
		State:          "done",
	}); err != nil {
		t.Fatalf("create undated stock movement failed: %v", err)
	}

	if got := partitionRowCount(t, "stock_movements_202609"); got != 1 {
		t.Fatalf("expected 1 movement in stock_movements_202609, got %d", got)
	}
	if got := partitionRowCount(t, "stock_movements_default"); got != 1 {
		t.Fatalf("expected 1 movement in stock_movements_default, got %d", got)
	}
}

func TestPartitionTriggersEnforceLineInvariants(t *testing.T) {
	f := seedPartitionFixture(t)

	july := f.post(t, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC), 100)

	ctx := testutil.SystemContext()
	mismatched := testDB.WithContext(ctx).Exec("INSERT INTO journal_lines (movement_id, date, account_id, debit, credit) VALUES (?, '2026-08-01', ?, 0, 50)", july.ID, f.accountID)
	if mismatched.Error == nil {
		t.Fatal("expected line date mismatch to be rejected")
	}
	missing := testDB.WithContext(ctx).Exec("INSERT INTO journal_lines (movement_id, date, account_id, debit, credit) VALUES (999999, '2026-07-01', ?, 0, 50)", f.accountID)
	if missing.Error == nil {
		t.Fatal("expected missing movement to be rejected")
	}

	before := partitionRowCount(t, "journal_lines")
	if err := testDB.WithContext(ctx).Exec("DELETE FROM journal_entrys WHERE id = ?", july.ID).Error; err != nil {
		t.Fatalf("delete movement failed: %v", err)
	}
	if after := partitionRowCount(t, "journal_lines"); after != before-2 {
		t.Fatalf("expected deleting the movement to cascade its 2 lines, got %d before and %d after", before, after)
	}
}

func TestStockMovementPartitionRoutingIsNotTreatedAsDelete(t *testing.T) {
	f := seedPartitionFixture(t)
	ctx := testutil.SystemContext()

	movement, err := f.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(f.organizationID),
		ItemID:         f.itemID,
		Qty:            3,
		SrcLocationID:  f.srcLocationID,
		DstLocationID:  f.dstLocationID,
		State:          "confirmed",
	})
	if err != nil {
		t.Fatalf("create stock movement failed: %v", err)
	}

	quant, err := inventory.NewStockBalanceDAO(testDB).Create(ctx, &inventory.StockBalance{
		OrganizationID: helper.Ptr(f.organizationID),
		ItemID:         f.itemID,
		LocationID:     f.srcLocationID,
		Quantity:       10,
	})
	if err != nil {
		t.Fatalf("create stock quant failed: %v", err)
	}
	if _, err := inventory.NewStockHoldDAO(testDB).Create(ctx, &inventory.StockHold{
		MovementID: helper.Ptr(movement.ID),
		BalanceID:  quant.ID,
		Qty:        3,
	}); err != nil {
		t.Fatalf("create stock reservation failed: %v", err)
	}

	dated := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	if err := testDB.WithContext(ctx).Exec("UPDATE stock_movements SET date_done = ? WHERE id = ?", dated, movement.ID).Error; err != nil {
		t.Fatalf("routing movement into a monthly partition failed: %v", err)
	}
	if err := testDB.WithContext(ctx).Exec("UPDATE stock_movements SET date_done = NULL WHERE id = ?", movement.ID).Error; err != nil {
		t.Fatalf("routing movement back to the default partition failed: %v", err)
	}
	if got := partitionRowCount(t, "stock_movements_default"); got != 1 {
		t.Fatalf("expected 1 movement in stock_movements_default after routing back, got %d", got)
	}

	deleteErr := testDB.WithContext(ctx).Exec("DELETE FROM stock_movements WHERE id = ?", movement.ID).Error
	if deleteErr == nil {
		t.Fatal("expected deleting a referenced movement to be rejected")
	}
}
