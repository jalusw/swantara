package manufacturing

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func testOutsideProcessingService(
	outsideOrders OutsideProcessingOrderDAOMock,
	productionOrders ProductionOrderDAOMock,
	components ConsumedMaterialDAOMock,
	recipes RecipeDAOMock,
	purchases PurchaseOrderCreator,
	poDAO procurement.PurchaseOrderDAOMock,
	poLines procurement.PurchaseOrderLineDAOMock,
	locations inventory.StockLocationDAOMock,
	movements inventory.StockMovementDAOMock,
	layers inventory.CostLayerDAOMock,
	resolver inventory.ItemResolverMock,
	poster inventory.PosterMock,
) OutsideProcessingService {
	return NewTestOutsideProcessingService(outsideOrders, productionOrders, components, recipes, purchases, poDAO, poLines, locations, movements, layers, resolver, poster)
}

func supplierLocationMock() inventory.StockLocationDAOMock {
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 30}, OrganizationID: helper.Ptr(uint64(1)), Usage: "supplier"}, nil
			},
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field != "usage" {
					return &query.Page[reference.StockLocation]{}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 30}, OrganizationID: helper.Ptr(uint64(1)), Usage: "supplier"}}}, nil
			},
		},
	}
}

func outsideProcessingRecipeMock() RecipeDAOMock {
	return RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*Recipe, error) {
				return &Recipe{Base: model.Base{ID: 1}, Type: RecipeTypeSubcontract}, nil
			},
		},
	}
}

func TestOutsideProcessingService_Create_OpensOrderForSubcontractRecipe(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	order := &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, SupplierID: 7, State: OutsideProcessingStateDraft}
	productionOrders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) {
				return &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, RecipeID: helper.Ptr(uint64(1)), State: ProductionOrderStateConfirmed}, nil
			},
		},
	}
	outsideOrders := OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
			CreateFunc: func(_ context.Context, sub *OutsideProcessingOrder) (*OutsideProcessingOrder, error) {
				return order, nil
			},
		},
	}
	svc := testOutsideProcessingService(outsideOrders, productionOrders, ConsumedMaterialDAOMock{}, outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})

	created, err := svc.Create(ctx, 1, 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.State != OutsideProcessingStateDraft || created.ProductionOrderID != 1 || created.SupplierID != 7 {
		t.Errorf("created = %+v, want draft order for productionOrder 1 supplier 7", created)
	}
}

func TestOutsideProcessingService_Create_RejectsNonSubcontractRecipe(t *testing.T) {
	ctx := context.Background()
	productionOrders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) {
				return &ProductionOrder{Base: model.Base{ID: 1}, RecipeID: helper.Ptr(uint64(1)), State: ProductionOrderStateConfirmed}, nil
			},
		},
	}
	recipes := RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*Recipe, error) {
				return &Recipe{Base: model.Base{ID: 1}, Type: RecipeTypeManufacture}, nil
			},
		},
	}
	svc := testOutsideProcessingService(OutsideProcessingOrderDAOMock{}, productionOrders, ConsumedMaterialDAOMock{}, recipes, PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})

	_, err := svc.Create(ctx, 1, 7)
	helper.AssertError(t, err, true, ErrOutsideProcessingNotSubcontracted)
}

func TestOutsideProcessingService_Create_RejectsDuplicate(t *testing.T) {
	ctx := context.Background()
	productionOrders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) {
				return &ProductionOrder{Base: model.Base{ID: 1}, RecipeID: helper.Ptr(uint64(1)), State: ProductionOrderStateConfirmed}, nil
			},
		},
	}
	outsideOrders := OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[OutsideProcessingOrder], error) {
				return &query.Page[OutsideProcessingOrder]{Items: []*OutsideProcessingOrder{{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: OutsideProcessingStateDraft}}}, nil
			},
		},
	}
	svc := testOutsideProcessingService(outsideOrders, productionOrders, ConsumedMaterialDAOMock{}, outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})

	_, err := svc.Create(ctx, 1, 7)
	helper.AssertError(t, err, true, ErrOutsideProcessingDuplicate)
}

func TestOutsideProcessingService_Send_CreatesPoAndDispatchesComponents(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	order := &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, SupplierID: 7, State: OutsideProcessingStateDraft}
	outsideOrders := OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, sub *OutsideProcessingOrder) (*OutsideProcessingOrder, error) {
				return sub, nil
			},
		},
	}
	productionOrders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) {
				return &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10, SrcLocationID: &src, State: ProductionOrderStateConfirmed}, nil
			},
		},
	}
	components := ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 10}}, nil
		},
	}
	purchases := PurchaseOrderCreatorMock{
		CreateFunc: func(_ context.Context, po *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			po.ID = 55
			po.AmountUntaxed = 120
			return po, nil
		},
	}
	movements := inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 100
				movement.State = inventory.MovementStateDraft
				return movement, nil
			},
			FindFunc: func(_ context.Context, moveID uint64) (*inventory.StockMovement, error) {
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, ItemID: 300, Qty: 10, SrcLocationID: src, DstLocationID: 30, State: inventory.MovementStateDraft, OrganizationID: &organizationID}, nil
			},
		},
		ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement.State = inventory.MovementStateDone
			return movement, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: model.Base{ID: 1}, Quantity: 50, UnitCost: helper.Ptr(2.0), RemainingQty: 50, RemainingValue: 100}}, nil
		},
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			layer.ID = 90
			return layer, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			return layer, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400}, Tracking: "none"}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := inventory.PosterMock{
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = append(posted, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := testOutsideProcessingService(outsideOrders, productionOrders, components, outsideProcessingRecipeMock(), purchases, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, supplierLocationMock(), movements, layers, resolver, poster)

	sent, err := svc.Send(ctx, 1, 500, 600, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent.State != OutsideProcessingStateSent || sent.PurchaseOrderID == nil || *sent.PurchaseOrderID != 55 {
		t.Errorf("sent = %+v, want sent order linked to po 55", sent)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1 material issue", len(posted))
	}
	if posted[0].Lines[0].AccountID != 600 || posted[0].Lines[1].AccountID != 400 {
		t.Errorf("material issue accounts = %d/%d, want 600/400", posted[0].Lines[0].AccountID, posted[0].Lines[1].AccountID)
	}
}

func TestOutsideProcessingService_Send_RejectsInvalidInputs(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)

	tests := []struct {
		name            string
		order           *OutsideProcessingOrder
		productionOrder *ProductionOrder
		components      []*ConsumedMaterial
		wantErr         error
	}{
		{name: "wrong order state", order: &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: OutsideProcessingStateSent}, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateConfirmed}, wantErr: ErrOutsideProcessingState},
		{name: "productionOrder not confirmed or planned", order: &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: OutsideProcessingStateDraft}, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateDraft}, wantErr: ErrOutsideProcessingOrderState},
		{name: "missing source location", order: &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: OutsideProcessingStateDraft}, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateConfirmed}, wantErr: ErrProductionOrderLocation},
		{name: "no components", order: &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: OutsideProcessingStateDraft}, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, SrcLocationID: &src, State: ProductionOrderStateConfirmed}, components: []*ConsumedMaterial{}, wantErr: ErrOutsideProcessingComponents},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outsideOrders := OutsideProcessingOrderDAOMock{
				CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return tt.order, nil },
				},
			}
			productionOrders := ProductionOrderDAOMock{
				CRUDMock: dao.CRUDMock[ProductionOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return tt.productionOrder, nil },
				},
			}
			components := ConsumedMaterialDAOMock{
				ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) { return tt.components, nil },
			}
			svc := testOutsideProcessingService(outsideOrders, productionOrders, components, outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})

			_, err := svc.Send(ctx, 1, 500, 600, time.Now())
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestOutsideProcessingService_Receive_ConsumesPoAndProducesFinished(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	dst := uint64(20)
	order := &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, SupplierID: 7, PurchaseOrderID: helper.Ptr(uint64(55)), State: OutsideProcessingStateSent}
	outsideOrders := OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, sub *OutsideProcessingOrder) (*OutsideProcessingOrder, error) {
				return sub, nil
			},
		},
	}
	var updatedMO *ProductionOrder
	productionOrders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) {
				return &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 200, QtyToProduce: 10, DstLocationID: &dst, State: ProductionOrderStatePlanned}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				updatedMO = productionOrder
				return productionOrder, nil
			},
		},
	}
	components := ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 10}}, nil
		},
	}
	poDAO := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return &procurement.PurchaseOrder{Base: model.Base{ID: 55}, OrganizationID: &organizationID, AmountUntaxed: 120}, nil
			},
		},
	}
	line := &procurement.PurchaseOrderLine{Base: model.Base{ID: 9}, OrderID: 55, QtyOrdered: 10, QtyReceived: 0}
	poLines := procurement.PurchaseOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrderLine]{
			UpdateFunc: func(_ context.Context, l *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
				line = l
				return l, nil
			},
		},
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{line}, nil
		},
	}
	movements := inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 200
				movement.State = inventory.MovementStateDraft
				return movement, nil
			},
			FindFunc: func(_ context.Context, moveID uint64) (*inventory.StockMovement, error) {
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, ItemID: 200, Qty: 10, SrcLocationID: 30, DstLocationID: dst, State: inventory.MovementStateDraft, OrganizationID: &organizationID}, nil
			},
		},
		ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement.State = inventory.MovementStateDone
			return movement, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			layer.ID = 90
			return layer, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			return layer, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, itemID uint64) (inventory.ResolvedItem, error) {
			if itemID == 300 {
				return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400}, StandardCost: 5, Tracking: "none"}, nil
			}
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400}, Tracking: "none"}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := inventory.PosterMock{
		PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = append(posted, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			posted = append(posted, request)
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := testOutsideProcessingService(outsideOrders, productionOrders, components, outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{}, poDAO, poLines, supplierLocationMock(), movements, layers, resolver, poster)

	received, err := svc.Receive(ctx, 1, 500, 600, 700, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.State != OutsideProcessingStateReceived {
		t.Errorf("state = %v, want received", received.State)
	}
	if line.QtyReceived != 10 {
		t.Errorf("po line qty_received = %v, want 10", line.QtyReceived)
	}
	if updatedMO == nil || updatedMO.State != ProductionOrderStateDone || updatedMO.QtyProduced != 10 {
		t.Errorf("productionOrder = %+v, want done with qty produced 10", updatedMO)
	}
	if len(posted) != 2 {
		t.Fatalf("postings = %d, want 2 (ap + produce)", len(posted))
	}
	if posted[0].Lines[0].AccountID != 600 || posted[0].Lines[1].AccountID != 700 || posted[0].Lines[0].Debit.Float64() != 120 {
		t.Errorf("ap posting = %+v, want wip debit 120 / ap credit 700", posted[0].Lines)
	}
	if posted[1].Lines[0].AccountID != 400 || posted[1].Lines[0].Debit.Float64() != 170 {
		t.Errorf("produce posting = %+v, want inventory debit 170", posted[1].Lines)
	}
}

func TestOutsideProcessingService_DoneAndCancel_EnforceState(t *testing.T) {
	ctx := context.Background()

	done, err := testOutsideProcessingService(
		OutsideProcessingOrderDAOMock{CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) {
			return &OutsideProcessingOrder{Base: model.Base{ID: 1}, State: OutsideProcessingStateReceived}, nil
		}}},
		ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, RecipeDAOMock{}, PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{},
	).Done(ctx, 1)
	if err != nil {
		t.Fatalf("done failed: %v", err)
	}
	if done.State != OutsideProcessingStateDone {
		t.Errorf("state = %v, want done", done.State)
	}

	_, err = testOutsideProcessingService(
		OutsideProcessingOrderDAOMock{CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) {
			return &OutsideProcessingOrder{Base: model.Base{ID: 1}, State: OutsideProcessingStateSent}, nil
		}}},
		ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, RecipeDAOMock{}, PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{},
	).Done(ctx, 1)
	helper.AssertError(t, err, true, ErrOutsideProcessingState)

	cancelled, err := testOutsideProcessingService(
		OutsideProcessingOrderDAOMock{CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) {
			return &OutsideProcessingOrder{Base: model.Base{ID: 1}, State: OutsideProcessingStateDraft}, nil
		}}},
		ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, RecipeDAOMock{}, PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{},
	).Cancel(ctx, 1)
	if err != nil {
		t.Fatalf("cancel failed: %v", err)
	}
	if cancelled.State != OutsideProcessingStateCancelled {
		t.Errorf("state = %v, want cancelled", cancelled.State)
	}
}
