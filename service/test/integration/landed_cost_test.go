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
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

type inboundCostFixture struct {
	orgID           uint64
	shipmentID      uint64
	moveID          uint64
	variantID       uint64
	stockValAcc     uint64
	clearingAcc     uint64
	journalID       uint64
	inboundCostSvc  inventory.InboundCostService
	adjustmentDAO   inventory.InboundCostAdjustmentDAO
	layerDAO        inventory.CostLayerDAO
	journalEntryDAO accounting.JournalEntryDAO
	moveLineDAO     accounting.JournalLineDAO
}

func TestInboundCostPostIntegration(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedInboundCostFixture(t)
	ctx := testutil.SystemContext()

	cost, err := fx.inboundCostSvc.Create(ctx, inventory.CreateInboundCostRequest{
		OrganizationID:    fx.orgID,
		Name:              "Ocean Freight",
		TargetShipmentIDs: helper.Int64Array{int64(fx.shipmentID)},
		Lines: []inventory.CreateInboundCostLineRequest{
			{
				ItemID:      fx.variantID,
				Description: "Freight cost",
				Amount:      100,
				SplitMethod: inventory.SplitMethodQuantity,
				AccountID:   helper.Ptr(fx.clearingAcc),
			},
		},
	})
	if err != nil {
		t.Fatalf("create landed cost failed: %v", err)
	}
	if cost.State != inventory.InboundCostStateDraft {
		t.Errorf("cost = %+v, want draft", cost)
	}

	posted, err := fx.inboundCostSvc.Post(ctx, fx.orgID, cost.ID)
	if err != nil {
		t.Fatalf("post landed cost failed: %v", err)
	}
	if posted.State != inventory.InboundCostStatePosted || posted.MovementID == nil {
		t.Fatalf("posted = %+v, want posted with movement", posted)
	}

	adjustments, err := fx.adjustmentDAO.ListByInboundCost(ctx, cost.ID)
	if err != nil {
		t.Fatalf("list adjustments failed: %v", err)
	}
	if len(adjustments) != 1 {
		t.Fatalf("expected 1 adjustment, got %d", len(adjustments))
	}
	if adjustments[0].AdditionalCost != 100 || adjustments[0].StockMovementID != fx.moveID {
		t.Errorf("adjustment = %+v, want +100 on movement %d", adjustments[0], fx.moveID)
	}

	layers, err := fx.layerDAO.ListByMovement(ctx, fx.moveID)
	if err != nil {
		t.Fatalf("list layers failed: %v", err)
	}
	if len(layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(layers))
	}
	if layers[0].RemainingValue != 600 {
		t.Errorf("layer remaining value = %v, want 600", layers[0].RemainingValue)
	}

	postLines, err := fx.moveLineDAO.ListByMovement(ctx, *posted.MovementID)
	if err != nil {
		t.Fatalf("list post lines failed: %v", err)
	}
	if len(postLines) != 2 {
		t.Fatalf("expected 2 posting lines, got %d", len(postLines))
	}
	assertJournalLine(t, postLines, fx.stockValAcc, 100, 0)
	assertJournalLine(t, postLines, fx.clearingAcc, 0, 100)
}

func seedInboundCostFixture(t *testing.T) inboundCostFixture {
	t.Helper()
	ctx := testutil.SystemContext()

	org, err := dao.NewBase[reference.Organization](testDB).Create(ctx, &reference.Organization{
		Name: gofakeit.Company(), BaseCurrency: "USD", Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("create organization failed: %v", err)
	}

	stockVal, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1500", Name: "Inventory", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create inventory account failed: %v", err)
	}
	stockIn, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1501", Name: "Stock Input", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create stock input account failed: %v", err)
	}
	stockOut, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "1502", Name: "Stock Output", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create stock output account failed: %v", err)
	}
	cogs, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "5000", Name: "Cost of Goods Sold", Type: "cogs", Active: true,
	})
	if err != nil {
		t.Fatalf("create cogs account failed: %v", err)
	}
	income, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "4000", Name: "Sales Revenue", Type: "income", Active: true,
	})
	if err != nil {
		t.Fatalf("create income account failed: %v", err)
	}
	clearing, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: org.ID, Code: "2100", Name: "Freight Payable", Type: "expense", Active: true,
	})
	if err != nil {
		t.Fatalf("create clearing account failed: %v", err)
	}

	journal, err := dao.NewBase[reference.Journal](testDB).Create(ctx, &reference.Journal{
		OrganizationID: org.ID, Name: "Stock Journal", Code: helper.Ptr("STK"),
		Type: "general", DefaultAccountID: helper.Ptr(stockVal.ID),
	})
	if err != nil {
		t.Fatalf("create journal failed: %v", err)
	}

	raw, err := json.Marshal(journal.ID)
	if err != nil {
		t.Fatalf("marshal config failed: %v", err)
	}
	if _, err := dao.NewBase[reference.SystemConfig](testDB).Create(ctx, &reference.SystemConfig{
		OrganizationID: helper.Ptr(org.ID), Key: "inventory.inbound_cost_journal_id", Value: raw,
	}); err != nil {
		t.Fatalf("create system config failed: %v", err)
	}

	category, err := dao.NewBase[reference.ItemCategory](testDB).Create(ctx, &reference.ItemCategory{
		Name:                    "General",
		IncomeAccountID:         helper.Ptr(income.ID),
		StockValuationAccountID: helper.Ptr(stockVal.ID),
		StockInputAccountID:     helper.Ptr(stockIn.ID),
		StockOutputAccountID:    helper.Ptr(stockOut.ID),
		CogsAccountID:           helper.Ptr(cogs.ID),
	})
	if err != nil {
		t.Fatalf("create item category failed: %v", err)
	}

	itemDAO := products.NewItemDAO(testDB)
	itemVariantDAO := products.NewItemVariantDAO(testDB)
	template, err := itemDAO.CreateWithVariants(ctx, &products.Item{
		OrganizationID: helper.Ptr(org.ID),
		Name:           "Test Item",
		CategoryID:     helper.Ptr(category.ID),
		Type:           "stockable",
		ListPrice:      100,
		StandardCost:   50,
		Tracking:       "none",
		IsPurchasable:  true,
		Active:         true,
	}, []*products.ItemVariant{{Active: true}})
	if err != nil {
		t.Fatalf("create item template failed: %v", err)
	}
	variants, err := itemVariantDAO.ListByTemplate(ctx, template.ID)
	if err != nil {
		t.Fatalf("list variants failed: %v", err)
	}
	if len(variants) == 0 {
		t.Fatal("expected at least one variant")
	}

	warehouse, err := inventory.NewWarehouseDAO(testDB).Create(ctx, &reference.Warehouse{
		OrganizationID: helper.Ptr(org.ID), Name: "Main Warehouse", Code: helper.Ptr("WH"),
	})
	if err != nil {
		t.Fatalf("create warehouse failed: %v", err)
	}

	stockLocationDAO := inventory.NewStockLocationDAO(testDB)
	internal, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), WarehouseID: helper.Ptr(warehouse.ID),
		Name: "Stock", Code: helper.Ptr("STK"), Usage: "internal",
	})
	if err != nil {
		t.Fatalf("create internal location failed: %v", err)
	}
	supplier, err := stockLocationDAO.Create(ctx, &reference.StockLocation{
		OrganizationID: helper.Ptr(org.ID), Name: "Suppliers", Usage: "supplier",
	})
	if err != nil {
		t.Fatalf("create supplier location failed: %v", err)
	}

	date := time.Now().UTC()
	shipmentDAO := inventory.NewShipmentDAO(testDB)
	shipment, err := shipmentDAO.CreateWithMovements(ctx, &inventory.Shipment{
		OrganizationID: helper.Ptr(org.ID), Name: helper.Ptr("PO-001"),
		Type: "incoming", SrcLocationID: helper.Ptr(supplier.ID), DstLocationID: helper.Ptr(internal.ID),
		State: "confirmed", ScheduledDate: helper.Ptr(date),
	}, []*inventory.StockMovement{{
		OrganizationID: helper.Ptr(org.ID),
		ItemID:         variants[0].ID,
		Qty:            10,
		SrcLocationID:  supplier.ID,
		DstLocationID:  internal.ID,
		State:          inventory.MovementStateConfirmed,
		ScheduledDate:  helper.Ptr(date),
	}})
	if err != nil {
		t.Fatalf("create shipment failed: %v", err)
	}
	movements, err := inventory.NewStockMovementDAO(testDB).ListByShipment(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("list movements failed: %v", err)
	}
	if len(movements) != 1 {
		t.Fatalf("expected 1 movement, got %d", len(movements))
	}

	journalEntryDAO := accounting.NewJournalEntryDAO(testDB)
	journalEntryLineDAO := accounting.NewJournalLineDAO(testDB)
	taxPeriodSvc := accounting.NewTaxPeriodService(accounting.NewTaxPeriodDAO(testDB), dao.NewBase[reference.TaxYear](testDB))
	reversalEngine := accounting.NewReversalEngine(journalEntryDAO, journalEntryLineDAO)
	postingSvc := accounting.NewPostingService(journalEntryDAO).SetPeriods(taxPeriodSvc).SetReverser(reversalEngine)

	stockMovementDAO := inventory.NewStockMovementDAO(testDB)
	layerDAO := inventory.NewCostLayerDAO(testDB)
	productResolver := inventory.NewProductResolver(itemVariantDAO, itemDAO, dao.NewBase[reference.ItemCategory](testDB))
	valuationSvc := inventory.NewValuationService(
		stockMovementDAO,
		layerDAO,
		stockLocationDAO,
		productResolver,
		postingSvc,
		db.NewDBTransactioner(testDB),
	)

	if _, err := valuationSvc.Receive(ctx, movements[0].ID, amount.FromFloat64(50), journal.ID, date); err != nil {
		t.Fatalf("receive stock failed: %v", err)
	}

	inboundCostSvc := inventory.NewInboundCostService(
		inventory.NewInboundCostDAO(testDB),
		inventory.NewInboundCostLineDAO(testDB),
		inventory.NewInboundCostAdjustmentDAO(testDB),
		stockMovementDAO,
		layerDAO,
		productResolver,
		postingSvc,
		inventory.NewInboundCostConfigSource(dao.NewBase[reference.SystemConfig](testDB)),
		accounting.NewInvoiceLineDAO(testDB),
		db.NewDBTransactioner(testDB),
	)

	return inboundCostFixture{
		orgID:           org.ID,
		shipmentID:      shipment.ID,
		moveID:          movements[0].ID,
		variantID:       variants[0].ID,
		stockValAcc:     stockVal.ID,
		clearingAcc:     clearing.ID,
		journalID:       journal.ID,
		inboundCostSvc:  inboundCostSvc,
		adjustmentDAO:   inventory.NewInboundCostAdjustmentDAO(testDB),
		layerDAO:        layerDAO,
		journalEntryDAO: journalEntryDAO,
		moveLineDAO:     journalEntryLineDAO,
	}
}
