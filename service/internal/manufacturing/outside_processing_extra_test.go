package manufacturing

import (
	"context"
	"errors"
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
)

func outsideOrder(state string, poID *uint64) *OutsideProcessingOrder {
	return &OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, SupplierID: 7, State: state, PurchaseOrderID: poID}
}

func prodOrder(state string) *ProductionOrder {
	orgID := uint64(1)
	src := uint64(10)
	dst := uint64(11)
	return &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &orgID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst, State: state}
}

func outsideProcessingTestService(outsideOrders OutsideProcessingOrderDAOMock, productionOrders ProductionOrderDAOMock, components ConsumedMaterialDAOMock, recipes RecipeDAOMock) OutsideProcessingService {
	return testOutsideProcessingService(outsideOrders, productionOrders, components, recipes, PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, supplierLocationMock(), inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
}

func TestOutsideProcessingService_Create_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(productionOrder *ProductionOrder, moErr error, recipe *Recipe, recipeErr error, listErr error, existing []*OutsideProcessingOrder) OutsideProcessingService {
		productionOrders := ProductionOrderDAOMock{
			CRUDMock: dao.CRUDMock[ProductionOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return productionOrder, moErr },
			},
		}
		recipes := RecipeDAOMock{
			CRUDMock: dao.CRUDMock[Recipe]{
				FindFunc: func(_ context.Context, _ uint64) (*Recipe, error) { return recipe, recipeErr },
			},
		}
		outsideOrders := OutsideProcessingOrderDAOMock{
			CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[OutsideProcessingOrder], error) {
					return &query.Page[OutsideProcessingOrder]{Items: existing}, listErr
				},
			},
		}
		return outsideProcessingTestService(outsideOrders, productionOrders, ConsumedMaterialDAOMock{}, recipes)
	}
	confirmed := prodOrder(ProductionOrderStateConfirmed)
	subRecipe := &Recipe{Base: model.Base{ID: 1}, Type: RecipeTypeSubcontract}

	cases := []struct {
		name            string
		productionOrder *ProductionOrder
		moErr           error
		recipe          *Recipe
		recipeErr       error
		listErr         error
		wantErr         error
	}{
		{name: "productionOrder lookup error", moErr: dbErr, wantErr: dbErr},
		{name: "productionOrder missing", wantErr: ErrProductionOrderNotFound},
		{name: "productionOrder draft", productionOrder: prodOrder(ProductionOrderStateDraft), recipe: subRecipe, wantErr: ErrOutsideProcessingOrderState},
		{name: "productionOrder without recipe", productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateConfirmed}, wantErr: ErrProductionOrderRecipe},
		{name: "recipe lookup error", productionOrder: confirmed, recipeErr: dbErr, wantErr: dbErr},
		{name: "recipe missing", productionOrder: confirmed, wantErr: ErrProductionOrderRecipe},
		{name: "list error", productionOrder: confirmed, recipe: subRecipe, listErr: dbErr, wantErr: dbErr},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.productionOrder, tt.moErr, tt.recipe, tt.recipeErr, tt.listErr, nil)
			_, err := svc.Create(ctx, 1, 7)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestOutsideProcessingService_Send_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(order *OutsideProcessingOrder, orderErr error, productionOrder *ProductionOrder, moErr error, components []*ConsumedMaterial, compErr error) OutsideProcessingService {
		outsideOrders := OutsideProcessingOrderDAOMock{
			CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return order, orderErr },
			},
		}
		productionOrders := ProductionOrderDAOMock{
			CRUDMock: dao.CRUDMock[ProductionOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return productionOrder, moErr },
			},
		}
		componentsMock := ConsumedMaterialDAOMock{
			ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) { return components, compErr },
		}
		return outsideProcessingTestService(outsideOrders, productionOrders, componentsMock, outsideProcessingRecipeMock())
	}
	draft := outsideOrder(OutsideProcessingStateDraft, nil)
	confirmed := prodOrder(ProductionOrderStateConfirmed)
	comp := []*ConsumedMaterial{{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 10}}

	cases := []struct {
		name            string
		order           *OutsideProcessingOrder
		orderErr        error
		productionOrder *ProductionOrder
		moErr           error
		components      []*ConsumedMaterial
		compErr         error
		wantErr         error
	}{
		{name: "order lookup error", orderErr: dbErr, wantErr: dbErr},
		{name: "order missing", wantErr: ErrOutsideProcessingNotFound},
		{name: "productionOrder lookup error", order: draft, moErr: dbErr, wantErr: dbErr},
		{name: "productionOrder missing", order: draft, wantErr: ErrProductionOrderNotFound},
		{name: "components error", order: draft, productionOrder: confirmed, compErr: dbErr, wantErr: dbErr},
		{name: "no components", order: draft, productionOrder: confirmed, wantErr: ErrOutsideProcessingComponents},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.order, tt.orderErr, tt.productionOrder, tt.moErr, tt.components, tt.compErr)
			if tt.name == "no components" {
				svc = newSvc(tt.order, tt.orderErr, tt.productionOrder, tt.moErr, []*ConsumedMaterial{}, tt.compErr)
			}
			if tt.name == "components error" {
				svc = newSvc(tt.order, tt.orderErr, tt.productionOrder, tt.moErr, nil, tt.compErr)
			}
			_ = comp
			_, err := svc.Send(ctx, 1, 500, 600, time.Now())
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}

	t.Run("supplier location error", func(t *testing.T) {
		svc := testOutsideProcessingService(
			OutsideProcessingOrderDAOMock{
				CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return draft, nil },
				},
			},
			ProductionOrderDAOMock{
				CRUDMock: dao.CRUDMock[ProductionOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return confirmed, nil },
				},
			},
			ConsumedMaterialDAOMock{
				ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) { return comp, nil },
			},
			outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{},
			inventory.StockLocationDAOMock{
				CRUDMock: dao.CRUDMock[reference.StockLocation]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
						return nil, dbErr
					},
				},
			}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
		_, err := svc.Send(ctx, 1, 500, 600, time.Now())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("purchase create error", func(t *testing.T) {
		svc := testOutsideProcessingService(
			OutsideProcessingOrderDAOMock{
				CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return draft, nil },
				},
			},
			ProductionOrderDAOMock{
				CRUDMock: dao.CRUDMock[ProductionOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return confirmed, nil },
				},
			},
			ConsumedMaterialDAOMock{
				ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) { return comp, nil },
			},
			outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{
				CreateFunc: func(_ context.Context, _ *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
					return nil, dbErr
				},
			}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, supplierLocationMock(), inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
		_, err := svc.Send(ctx, 1, 500, 600, time.Now())
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestOutsideProcessingService_Receive_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	poID := uint64(55)

	newSvc := func(order *OutsideProcessingOrder, orderErr error, productionOrder *ProductionOrder, moErr error) OutsideProcessingService {
		outsideOrders := OutsideProcessingOrderDAOMock{
			CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return order, orderErr },
			},
		}
		productionOrders := ProductionOrderDAOMock{
			CRUDMock: dao.CRUDMock[ProductionOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return productionOrder, moErr },
			},
		}
		return outsideProcessingTestService(outsideOrders, productionOrders, ConsumedMaterialDAOMock{}, outsideProcessingRecipeMock())
	}
	sent := outsideOrder(OutsideProcessingStateSent, &poID)

	cases := []struct {
		name            string
		order           *OutsideProcessingOrder
		orderErr        error
		productionOrder *ProductionOrder
		moErr           error
		wip             uint64
		ap              uint64
		wantErr         error
	}{
		{name: "wip required", wip: 0, ap: 1, order: sent, productionOrder: prodOrder(ProductionOrderStateConfirmed), wantErr: ErrWIPAccount},
		{name: "payable required", wip: 1, ap: 0, order: sent, productionOrder: prodOrder(ProductionOrderStateConfirmed), wantErr: ErrOutsideProcessingOperation},
		{name: "order lookup error", wip: 1, ap: 1, orderErr: dbErr, wantErr: dbErr},
		{name: "order missing", wip: 1, ap: 1, wantErr: ErrOutsideProcessingNotFound},
		{name: "wrong state", wip: 1, ap: 1, order: outsideOrder(OutsideProcessingStateDraft, &poID), wantErr: ErrOutsideProcessingState},
		{name: "missing po", wip: 1, ap: 1, order: outsideOrder(OutsideProcessingStateSent, nil), wantErr: ErrOutsideProcessingPurchaseOrder},
		{name: "productionOrder lookup error", wip: 1, ap: 1, order: sent, moErr: dbErr, wantErr: dbErr},
		{name: "productionOrder missing", wip: 1, ap: 1, order: sent, wantErr: ErrProductionOrderNotFound},
		{name: "productionOrder bad state", wip: 1, ap: 1, order: sent, productionOrder: prodOrder(ProductionOrderStateDraft), wantErr: ErrOutsideProcessingOrderState},
		{name: "productionOrder no location", wip: 1, ap: 1, order: sent, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateConfirmed}, wantErr: ErrProductionOrderLocation},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.order, tt.orderErr, tt.productionOrder, tt.moErr)
			_, err := svc.Receive(ctx, 1, 500, tt.wip, tt.ap, time.Now())
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestOutsideProcessingService_Done_Cancel(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(order *OutsideProcessingOrder, orderErr error, updateErr error) OutsideProcessingService {
		outsideOrders := OutsideProcessingOrderDAOMock{
			CRUDMock: dao.CRUDMock[OutsideProcessingOrder]{
				FindFunc: func(_ context.Context, _ uint64) (*OutsideProcessingOrder, error) { return order, orderErr },
				UpdateFunc: func(_ context.Context, o *OutsideProcessingOrder) (*OutsideProcessingOrder, error) {
					return o, updateErr
				},
			},
		}
		return outsideProcessingTestService(outsideOrders, ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, outsideProcessingRecipeMock())
	}

	t.Run("done transitions", func(t *testing.T) {
		svc := newSvc(outsideOrder(OutsideProcessingStateReceived, nil), nil, nil)
		got, err := svc.Done(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != OutsideProcessingStateDone {
			t.Errorf("state = %q", got.State)
		}
	})

	t.Run("done failures", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		_, err := svc.Done(ctx, 1)
		helper.AssertError(t, err, true, dbErr)

		svc = newSvc(nil, nil, nil)
		_, err = svc.Done(ctx, 1)
		helper.AssertError(t, err, true, ErrOutsideProcessingNotFound)

		svc = newSvc(outsideOrder(OutsideProcessingStateDraft, nil), nil, nil)
		_, err = svc.Done(ctx, 1)
		helper.AssertError(t, err, true, ErrOutsideProcessingState)

		svc = newSvc(outsideOrder(OutsideProcessingStateReceived, nil), nil, dbErr)
		_, err = svc.Done(ctx, 1)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("cancel transitions", func(t *testing.T) {
		svc := newSvc(outsideOrder(OutsideProcessingStateDraft, nil), nil, nil)
		got, err := svc.Cancel(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != OutsideProcessingStateCancelled {
			t.Errorf("state = %q", got.State)
		}
	})

	t.Run("cancel failures", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil)
		_, err := svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, dbErr)

		svc = newSvc(nil, nil, nil)
		_, err = svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, ErrOutsideProcessingNotFound)

		svc = newSvc(outsideOrder(OutsideProcessingStateDone, nil), nil, nil)
		_, err = svc.Cancel(ctx, 1)
		helper.AssertError(t, err, true, ErrOutsideProcessingState)
	})
}

func TestOutsideProcessingService_FindSupplierLocation(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(locations inventory.StockLocationDAOMock) OutsideProcessingService {
		return testOutsideProcessingService(OutsideProcessingOrderDAOMock{}, ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, outsideProcessingRecipeMock(), PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, locations, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	}

	t.Run("matches global location", func(t *testing.T) {
		svc := newSvc(inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 30}}}}, nil
				},
			},
		})
		got, err := svc.findSupplierLocation(ctx, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != 30 {
			t.Errorf("location = %d, want 30", got)
		}
	})

	t.Run("skips foreign and returns zero", func(t *testing.T) {
		svc := newSvc(inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 31}, OrganizationID: helper.Ptr(uint64(99))}}}, nil
				},
			},
		})
		got, err := svc.findSupplierLocation(ctx, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != 0 {
			t.Errorf("location = %d, want 0", got)
		}
	})

	t.Run("propagates error", func(t *testing.T) {
		svc := newSvc(inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
					return nil, dbErr
				},
			},
		})
		_, err := svc.findSupplierLocation(ctx, 10)
		helper.AssertError(t, err, true, dbErr)
	})
}

var _ = accounting.OriginTypeOutsideProcessingOrder
