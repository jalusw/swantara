//go:build integration

package integration

import (
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestStandardCostValuationReconcilesToControlAccount(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t, "standard")
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
		t.Fatalf("create receive movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Receive(ctx, receive.ID, amount.FromFloat64(40), fx.journalID, fx.date); err != nil {
		t.Fatalf("receive 10@40 failed: %v", err)
	}

	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 500 {
		t.Fatalf("expected inventory GL 500 (10 x std 50, actual 40 ignored), got %v", balance)
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

	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 300 {
		t.Fatalf("expected inventory GL 300 after ship 4 at std 50, got %v", balance)
	}
	if balance := glBalance(t, fx.orgID, fx.cogsID); balance != 200 {
		t.Fatalf("expected COGS 200, got %v", balance)
	}

	rows, err := fx.reportDAO.InventoryValuation(ctx, fx.orgID)
	if err != nil {
		t.Fatalf("inventory valuation failed: %v", err)
	}
	if len(rows) != 1 || rows[0].Value != 300 || rows[0].Quantity != 6 {
		t.Fatalf("expected valuation 6 qty / 300 value, got %+v", rows)
	}

	layerDAO := inventory.NewCostLayerDAO(testDB)
	openLayers, err := layerDAO.ListOpenByItem(ctx, fx.variantID)
	if err != nil {
		t.Fatalf("list open layers failed: %v", err)
	}
	if len(openLayers) != 1 {
		t.Fatalf("expected 1 open layer, got %d", len(openLayers))
	}
	if openLayers[0].RemainingQty != 6 || openLayers[0].RemainingValue != 300 {
		t.Fatalf("open layer = qty %v / value %v, want 6 / 300", openLayers[0].RemainingQty, openLayers[0].RemainingValue)
	}
}
