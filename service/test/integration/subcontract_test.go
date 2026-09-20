//go:build integration

package integration

import (
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/test/testutil"
)

func TestSubcontract_MoFlowsVendorPurchaseDispatchAndReceipt(t *testing.T) {
	testutil.CleanTables(t, testDB)
	fx := seedIntegrityFixture(t)
	ctx := testutil.SystemContext()
	date := fx.date

	stockIn, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "1501S", Name: "Component Stock Input", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create stock input account failed: %v", err)
	}
	stockOut, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "1502S", Name: "Component Stock Output", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create stock output account failed: %v", err)
	}
	wipAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "1600", Name: "Subcontracted WIP", Type: "asset", Active: true,
	})
	if err != nil {
		t.Fatalf("create wip account failed: %v", err)
	}
	apAccount, err := dao.NewBase[reference.Account](testDB).Create(ctx, &reference.Account{
		OrganizationID: fx.orgID, Code: "2401", Name: "AP Payable", Type: "liability", Active: true,
	})
	if err != nil {
		t.Fatalf("create ap payable account failed: %v", err)
	}

	componentCategory, err := dao.NewBase[reference.ItemCategory](testDB).Create(ctx, &reference.ItemCategory{
		Name:                    "Subcontract Component",
		CostMethod:              helper.Ptr("standard"),
		IncomeAccountID:         helper.Ptr(fx.incomeID),
		StockValuationAccountID: helper.Ptr(fx.inventoryID),
		StockInputAccountID:     helper.Ptr(stockIn.ID),
		StockOutputAccountID:    helper.Ptr(stockOut.ID),
		CogsAccountID:           helper.Ptr(fx.cogsID),
	})
	if err != nil {
		t.Fatalf("create component category failed: %v", err)
	}

	componentTemplate, err := products.NewItemDAO(testDB).Create(ctx, &products.Item{
		OrganizationID: helper.Ptr(fx.orgID), Name: "Subcomponent", CategoryID: helper.Ptr(componentCategory.ID),
		Type: "stockable", ListPrice: 30, StandardCost: 10, Tracking: "none", IsSellable: true, Active: true,
	})
	if err != nil {
		t.Fatalf("create component template failed: %v", err)
	}
	componentVariant, err := products.NewItemVariantDAO(testDB).Create(ctx, &products.ItemVariant{
		ItemID: componentTemplate.ID, Active: true,
	})
	if err != nil {
		t.Fatalf("create component variant failed: %v", err)
	}

	supplier, err := contacts.NewContactDAO(testDB).Create(ctx, &contacts.Contact{
		OrganizationID: &fx.orgID, Name: "Contract Manufacturer",
	})
	if err != nil {
		t.Fatalf("create supplier failed: %v", err)
	}
	if _, err := contacts.NewSupplierProfileDAO(testDB).Create(ctx, &contacts.SupplierProfile{
		ContactID: supplier.ID, Active: true,
	}); err != nil {
		t.Fatalf("create supplier supplier failed: %v", err)
	}
	if _, err := products.NewSupplierProductDAO(testDB).Create(ctx, &products.SupplierProduct{
		ItemID: fx.variantID, SupplierID: supplier.ID, MinQty: 1, Price: helper.Ptr(25.0), Priority: 10,
	}); err != nil {
		t.Fatalf("create supplier item failed: %v", err)
	}

	if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
		OrganizationID: fx.orgID, Code: manufacturing.SequenceProductionOrderCode, NextNumber: 1, Padding: 5,
	}).Error; err != nil {
		t.Fatalf("create mo sequence failed: %v", err)
	}
	if err := testDB.WithContext(ctx).Create(&sequence.DocumentSequence{
		OrganizationID: fx.orgID, Code: procurement.SequencePurchaseOrderCode, NextNumber: 1, Padding: 5,
	}).Error; err != nil {
		t.Fatalf("create po sequence failed: %v", err)
	}

	itemVariantDAO := products.NewItemVariantDAO(testDB)
	itemDAO := products.NewItemDAO(testDB)
	productResolver := inventory.NewProductResolver(itemVariantDAO, itemDAO, dao.NewBase[reference.ItemCategory](testDB))

	supplierProductSvc := products.NewSupplierProductService(
		itemVariantDAO,
		itemDAO,
		products.NewSupplierProductDAO(testDB),
		contacts.NewContactDAO(testDB),
		contacts.NewSupplierProfileDAO(testDB),
	)
	poSvc := procurement.NewPurchaseOrderService(
		procurement.NewPurchaseOrderDAO(testDB),
		procurement.NewPurchaseOrderLineDAO(testDB),
		procurement.NewPurchaseRequestDAO(testDB),
		procurement.NewPurchaseRequestLineDAO(testDB),
		sequence.NewSequenceService(sequence.NewDAO(testDB)),
		procurement.NewSystemConfigSource(dao.NewBase[reference.SystemConfig](testDB)),
		supplierProductSvc,
		contacts.NewContactDAO(testDB),
		contacts.NewSupplierProfileDAO(testDB),
		inventory.NewWarehouseDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		inventory.NewShipmentDAO(testDB),
		inventory.NewStockMovementDAO(testDB),
		productResolver,
		phase37ExpenseEngine{},
		dao.NewBase[reference.Tax](testDB),
		phase37ReceiveEngine{},
		procurement.ApprovalEngineMock{},
		procurement.BillEngineMock{},
		procurement.OpenInvoiceLookupMock{},
		procurement.OutboundPaymentEngineMock{},
		procurement.QualityEngineMock{},
	)

	bomSvc := manufacturing.NewRecipeService(itemVariantDAO, manufacturing.NewRecipeDAO(testDB), manufacturing.NewRecipeLineDAO(testDB))
	recipe, err := bomSvc.CreateRecipe(ctx, &manufacturing.Recipe{
		OrganizationID: &fx.orgID, ItemID: fx.variantID, Type: manufacturing.RecipeTypeSubcontract, Qty: 1,
	}, []*manufacturing.RecipeLine{
		{ComponentID: componentVariant.ID, Qty: 2},
	})
	if err != nil {
		t.Fatalf("create recipe failed: %v", err)
	}

	moDAO := manufacturing.NewProductionOrderDAO(testDB)
	consumedMaterialDAO := manufacturing.NewConsumedMaterialDAO(testDB)
	recipeDAO := manufacturing.NewRecipeDAO(testDB)
	moSvc := manufacturing.NewProductionOrderService(
		moDAO,
		consumedMaterialDAO,
		recipeDAO,
		bomSvc,
		itemVariantDAO,
		inventory.NewStockLocationDAO(testDB),
		inventory.NewHoldService(inventory.NewStockHoldDAO(testDB), inventory.NewStockBalanceDAO(testDB)),
		sequence.NewSequenceService(sequence.NewDAO(testDB)),
	)
	mo, err := moSvc.Create(ctx, &manufacturing.ProductionOrder{
		OrganizationID: &fx.orgID, ItemID: fx.variantID, RecipeID: &recipe.ID, QtyToProduce: 5,
		SrcLocationID: &fx.internalLocID, DstLocationID: &fx.internalLocID,
	})
	if err != nil {
		t.Fatalf("create mo failed: %v", err)
	}
	if _, err := moSvc.Confirm(ctx, mo.ID); err != nil {
		t.Fatalf("confirm mo failed: %v", err)
	}

	componentReceive, err := fx.ledgerSvc.CreateMovement(ctx, &inventory.StockMovement{
		OrganizationID: &fx.orgID, ItemID: componentVariant.ID, Qty: 20,
		SrcLocationID: fx.supplierLocID, DstLocationID: fx.internalLocID,
		OriginType: helper.Ptr("test"), OriginID: helper.Ptr(uint64(1)),
	})
	if err != nil {
		t.Fatalf("create component receive movement failed: %v", err)
	}
	if _, err := fx.valuationSvc.Receive(ctx, componentReceive.ID, amount.FromFloat64(10), fx.journalID, date); err != nil {
		t.Fatalf("receive component into stock failed: %v", err)
	}

	subSvc := manufacturing.NewOutsideProcessingService(
		manufacturing.NewOutsideProcessingOrderDAO(testDB),
		moDAO,
		consumedMaterialDAO,
		recipeDAO,
		poSvc,
		procurement.NewPurchaseOrderDAO(testDB),
		procurement.NewPurchaseOrderLineDAO(testDB),
		inventory.NewStockLocationDAO(testDB),
		fx.ledgerSvc,
		fx.valuationSvc,
		productResolver,
		fx.postingSvc,
	)

	order, err := subSvc.Create(ctx, mo.ID, supplier.ID)
	if err != nil {
		t.Fatalf("create subcontract order failed: %v", err)
	}
	if _, err := subSvc.Create(ctx, mo.ID, supplier.ID); err == nil {
		t.Fatal("duplicate subcontract order accepted")
	}

	order, err = subSvc.Send(ctx, order.ID, fx.journalID, wipAccount.ID, date)
	if err != nil {
		t.Fatalf("send subcontract order failed: %v", err)
	}
	if order.PurchaseOrderID == nil {
		t.Fatal("subcontract order not linked to a purchase order")
	}
	po, err := procurement.NewPurchaseOrderDAO(testDB).Find(ctx, *order.PurchaseOrderID)
	if err != nil {
		t.Fatalf("find po failed: %v", err)
	}
	if po.AmountUntaxed != 125 {
		t.Errorf("po amount = %v, want 125", po.AmountUntaxed)
	}

	order, err = subSvc.Receive(ctx, order.ID, fx.journalID, wipAccount.ID, apAccount.ID, date)
	if err != nil {
		t.Fatalf("receive subcontract order failed: %v", err)
	}
	if _, err := subSvc.Done(ctx, order.ID); err != nil {
		t.Fatalf("done subcontract order failed: %v", err)
	}

	if balance := glBalance(t, fx.orgID, wipAccount.ID); balance != 0 {
		t.Errorf("wip balance = %v, want 0", balance)
	}
	if balance := glBalance(t, fx.orgID, apAccount.ID); balance != -125 {
		t.Errorf("ap payable balance = %v, want -125", balance)
	}
	if balance := glBalance(t, fx.orgID, fx.inventoryID); balance != 325 {
		t.Errorf("inventory balance = %v, want 325", balance)
	}

	onHand, err := fx.valuationSvc.OnHandValue(ctx, fx.variantID)
	if err != nil {
		t.Fatalf("on hand value failed: %v", err)
	}
	if onHand.Float64() != 225 {
		t.Errorf("finished goods value = %v, want 225", onHand.Float64())
	}

	poLines, err := procurement.NewPurchaseOrderLineDAO(testDB).ListByOrder(ctx, po.ID)
	if err != nil {
		t.Fatalf("list po lines failed: %v", err)
	}
	if len(poLines) != 1 || poLines[0].QtyReceived != 5 {
		t.Errorf("po line qty_received = %+v, want 5", poLines)
	}

	mo, err = moDAO.Find(ctx, mo.ID)
	if err != nil {
		t.Fatalf("find mo failed: %v", err)
	}
	if mo.State != manufacturing.ProductionOrderStateDone || mo.QtyProduced != 5 {
		t.Errorf("mo = %+v, want done with 5 produced", mo)
	}

	movements, err := inventory.NewStockMovementDAO(testDB).ListByOrigin(ctx, "outside_processing_order", order.ID)
	if err != nil {
		t.Fatalf("list subcontract movements failed: %v", err)
	}
	if len(movements) != 2 {
		t.Errorf("subcontract stock movements = %d, want 2 (dispatch + receipt)", len(movements))
	}

	if err := testDB.WithContext(ctx).Create(&manufacturing.OutsideProcessingOrder{
		ProductionOrderID: 999999, SupplierID: supplier.ID, State: manufacturing.OutsideProcessingStateDraft,
	}).Error; err == nil {
		t.Fatal("subcontract order accepted without an existing manufacturing order")
	}
	if err := testDB.WithContext(ctx).Create(&manufacturing.OutsideProcessingOrder{
		ProductionOrderID: mo.ID, SupplierID: supplier.ID, PurchaseOrderID: helper.Ptr(uint64(999999)), State: manufacturing.OutsideProcessingStateDraft,
	}).Error; err == nil {
		t.Fatal("subcontract order accepted without an existing purchase order")
	}
}
