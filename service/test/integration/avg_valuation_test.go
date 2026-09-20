//go:build integration

package integration

import (
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestAverageValuationReconcilesToControlAccount(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t, "average")
	ctx := testutil.SystemContext()

	receive, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            10,
		SrcLocationID:  fx.supplierLocID,
		DstLocationID:  fx.internalLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create first receive movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Receive(ctx, receive.ID, amount.FromFloat64(50), fx.journalID, fx.date); err != nil {
		t.Fatalf("receive 10@50 failed: %v", err)
	}

	receive2, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            10,
		SrcLocationID:  fx.supplierLocID,
		DstLocationID:  fx.internalLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create second receive movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Receive(ctx, receive2.ID, amount.FromFloat64(60), fx.journalID, fx.date); err != nil {
		t.Fatalf("receive 10@60 failed: %v", err)
	}

	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 1100 {
		t.Fatalf("expected inventory GL 1100 after mixed receipts, got %v", balance)
	}

	ship, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            4,
		SrcLocationID:  fx.internalLocID,
		DstLocationID:  fx.customerLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create ship movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Ship(ctx, ship.ID, fx.journalID, fx.date); err != nil {
		t.Fatalf("ship stock failed: %v", err)
	}

	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 880 {
		t.Fatalf("expected inventory GL 880 after ship 4 at avg 55, got %v", balance)
	}
	if balance := glBalance(t, fx.orgID, fx.cogsID); balance != 220 {
		t.Fatalf("expected COGS 220 after ship 4, got %v", balance)
	}

	rows, err := fx.reportDAO.InventoryValuation(ctx, fx.orgID)
	if err != nil {
		t.Fatalf("inventory valuation failed: %v", err)
	}
	if len(rows) != 1 || rows[0].Value != 880 || rows[0].Quantity != 16 {
		t.Fatalf("expected valuation 16 qty / 880 value, got %+v", rows)
	}

	layerDAO := inventory.NewCostLayerDAO(testDB)
	openLayers, err := layerDAO.ListOpenByItem(ctx, fx.variantID)
	if err != nil {
		t.Fatalf("list open layers failed: %v", err)
	}
	if len(openLayers) != 2 {
		t.Fatalf("expected 2 open layers, got %d", len(openLayers))
	}
	totalQty := 0.0
	totalValue := 0.0
	for _, layer := range openLayers {
		totalQty += layer.RemainingQty
		totalValue += layer.RemainingValue
	}
	if totalQty != 16 || totalValue != 880 {
		t.Fatalf("open layers = qty %v / value %v, want 16 / 880", totalQty, totalValue)
	}

	ship2, err := fx.stockMovementDAO.Create(ctx, &inventory.StockMovement{
		OrganizationID: helper.Ptr(fx.orgID),
		ItemID:         fx.variantID,
		Qty:            6,
		SrcLocationID:  fx.internalLocID,
		DstLocationID:  fx.customerLocID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(fx.date),
	})
	if err != nil {
		t.Fatalf("create second ship movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Ship(ctx, ship2.ID, fx.journalID, fx.date); err != nil {
		t.Fatalf("second ship failed: %v", err)
	}

	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 550 {
		t.Fatalf("expected inventory GL 550 after second ship at avg 55, got %v", balance)
	}
	if balance := glBalance(t, fx.orgID, fx.cogsID); balance != 550 {
		t.Fatalf("expected cumulative COGS 550, got %v", balance)
	}

	rows, err = fx.reportDAO.InventoryValuation(ctx, fx.orgID)
	if err != nil {
		t.Fatalf("inventory valuation after second ship failed: %v", err)
	}
	if len(rows) != 1 || rows[0].Value != 550 || rows[0].Quantity != 10 {
		t.Fatalf("expected valuation 10 qty / 550 value, got %+v", rows)
	}
}
