package manufacturing

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func testProductionOrderService(
	orders ProductionOrderDAOMock,
	components ConsumedMaterialDAOMock,
	recipes RecipeDAOMock,
	lines RecipeLineDAOMock,
	variants products.ItemVariantDAOMock,
	locations inventory.StockLocationDAOMock,
	reservations inventory.StockHoldDAOMock,
	quants inventory.StockBalanceDAOMock,
	sequences sequence.DAOMock,
) ProductionOrderService {
	return NewTestProductionOrderService(orders, components, recipes, lines, variants, locations, reservations, quants, sequences)
}

func moLocationMock() inventory.StockLocationDAOMock {
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1)), Usage: "internal"}, nil
			},
		},
	}
}

func moForeignLocationMock() inventory.StockLocationDAOMock {
	foreignOrg := uint64(2)
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: &foreignOrg, Usage: "internal"}, nil
			},
		},
	}
}

func moVariantMock() products.ItemVariantDAOMock {
	return products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 200}, ItemID: 1}, nil
			},
		},
	}
}

func moRecipeMock() RecipeDAOMock {
	return RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*Recipe, error) {
				return &Recipe{Base: model.Base{ID: 1}, ItemID: 200, Qty: 1, Type: RecipeTypeManufacture}, nil
			},
		},
	}
}

func moRecipeLinesMock() RecipeLineDAOMock {
	return RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*RecipeLine, error) {
			return []*RecipeLine{
				{Base: model.Base{ID: 1}, RecipeID: 1, ComponentID: 300, Qty: 2, ScrapPct: 10},
				{Base: model.Base{ID: 2}, RecipeID: 1, ComponentID: 301, Qty: 0.5, ScrapPct: 0},
			}, nil
		},
	}
}

func moSequenceMock() sequence.DAOMock {
	return DefaultOrderSequenceMock()
}

func TestProductionOrderService_Create_ExplodesRecipeAndSequencesName(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	dst := uint64(20)
	var createdComponents []*ConsumedMaterial
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{},
		CreateWithComponentsFunc: func(_ context.Context, productionOrder *ProductionOrder, components []*ConsumedMaterial) (*ProductionOrder, error) {
			productionOrder.ID = 5
			createdComponents = components
			return productionOrder, nil
		},
	}
	svc := testProductionOrderService(orders, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())

	productionOrder, err := svc.Create(ctx, &ProductionOrder{
		OrganizationID: &organizationID,
		ItemID:         200,
		RecipeID:       helper.Ptr(uint64(1)),
		QtyToProduce:   10,
		SrcLocationID:  &src,
		DstLocationID:  &dst,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if productionOrder.Name == nil || *productionOrder.Name != "SEQ/00001" {
		t.Errorf("name = %v, want SEQ/00001", productionOrder.Name)
	}
	if productionOrder.State != ProductionOrderStateDraft {
		t.Errorf("state = %s, want draft", productionOrder.State)
	}
	if len(createdComponents) != 2 {
		t.Fatalf("components = %d, want 2", len(createdComponents))
	}
	if createdComponents[0].ItemID != 300 || createdComponents[0].QtyPlanned != 22 {
		t.Errorf("component 0 = item %d qty %v, want 300/22", createdComponents[0].ItemID, createdComponents[0].QtyPlanned)
	}
	if createdComponents[1].ItemID != 301 || createdComponents[1].QtyPlanned != 5 {
		t.Errorf("component 1 = item %d qty %v, want 301/5", createdComponents[1].ItemID, createdComponents[1].QtyPlanned)
	}
}

func TestProductionOrderService_Create_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	dst := uint64(20)
	validMO := func() *ProductionOrder {
		return &ProductionOrder{
			OrganizationID: &organizationID,
			ItemID:         200,
			RecipeID:       helper.Ptr(uint64(1)),
			QtyToProduce:   10,
			SrcLocationID:  &src,
			DstLocationID:  &dst,
		}
	}

	mismatchedRecipe := RecipeDAOMock{
		CRUDMock: dao.CRUDMock[Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*Recipe, error) {
				return &Recipe{Base: model.Base{ID: 1}, ItemID: 999, Qty: 1, Type: RecipeTypeManufacture}, nil
			},
		},
	}

	tests := []struct {
		name            string
		svc             ProductionOrderService
		productionOrder *ProductionOrder
		wantErr         error
	}{
		{
			name: "missing organization",
			svc:  testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock()),
			productionOrder: &ProductionOrder{
				ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst,
			},
			wantErr: ErrProductionOrderOrganization,
		},
		{
			name:            "non positive qty",
			svc:             testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock()),
			productionOrder: &ProductionOrder{OrganizationID: &organizationID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 0, SrcLocationID: &src, DstLocationID: &dst},
			wantErr:         ErrProductionOrderQty,
		},
		{
			name:            "missing recipe",
			svc:             testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock()),
			productionOrder: &ProductionOrder{OrganizationID: &organizationID, ItemID: 200, QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst},
			wantErr:         ErrProductionOrderRecipe,
		},
		{
			name:            "recipe item mismatch",
			svc:             testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, mismatchedRecipe, moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock()),
			productionOrder: validMO(),
			wantErr:         ErrProductionOrderRecipe,
		},
		{
			name:            "missing locations",
			svc:             testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock()),
			productionOrder: &ProductionOrder{OrganizationID: &organizationID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10},
			wantErr:         ErrProductionOrderLocation,
		},
		{
			name:            "foreign location",
			svc:             testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moForeignLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock()),
			productionOrder: validMO(),
			wantErr:         ErrProductionOrderLocation,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.svc.Create(ctx, tt.productionOrder)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionOrderService_Create_RejectsUnknownVariantAndMissingSequence(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	dst := uint64(20)

	tests := []struct {
		name     string
		variants products.ItemVariantDAOMock
		wantErr  error
	}{
		{
			name:     "unknown item",
			variants: products.ItemVariantDAOMock{},
			wantErr:  ErrProductionOrderItem,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), tt.variants, moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())
			_, err := svc.Create(ctx, &ProductionOrder{OrganizationID: &organizationID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst})
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}

	t.Run("missing sequence", func(t *testing.T) {
		sequences := sequence.DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
				return nil, sequence.ErrSequenceNotFound
			},
		}
		svc := testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, sequences)
		_, err := svc.Create(ctx, &ProductionOrder{OrganizationID: &organizationID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst})
		helper.AssertError(t, err, true, sequence.ErrSequenceNotFound)
	})
}

func TestProductionOrderService_Confirm_TransitionsDraftToConfirmed(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateDraft}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	svc := testProductionOrderService(orders, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())

	updated, err := svc.Confirm(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.State != ProductionOrderStateConfirmed {
		t.Errorf("state = %s, want confirmed", updated.State)
	}
}

func TestProductionOrderService_Confirm_RejectsNonDraft(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateDone}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
		},
	}
	svc := testProductionOrderService(orders, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())

	_, err := svc.Confirm(ctx, 1)
	helper.AssertError(t, err, true, ErrProductionOrderState)
}

func TestProductionOrderService_Plan_ReservesComponents(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateConfirmed, SrcLocationID: &src}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
			UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	components := ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{
				{Base: model.Base{ID: 1}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 22},
				{Base: model.Base{ID: 2}, ProductionOrderID: 1, ItemID: 301, QtyPlanned: 5},
			}, nil
		},
	}
	reserved := make([]float64, 0)
	reservations := inventory.StockHoldDAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ *uint64, qty float64) (*inventory.StockHold, error) {
			reserved = append(reserved, qty)
			return &inventory.StockHold{Base: model.Base{ID: 1}, Qty: qty}, nil
		},
	}
	quants := inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Quantity: 100}, nil
		},
	}
	svc := testProductionOrderService(orders, components, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), reservations, quants, moSequenceMock())

	updated, err := svc.Plan(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.State != ProductionOrderStatePlanned {
		t.Errorf("state = %s, want planned", updated.State)
	}
	if len(reserved) != 2 || reserved[0] != 22 || reserved[1] != 5 {
		t.Errorf("reserved quantities = %v, want [22 5]", reserved)
	}
}

func TestProductionOrderService_Plan_RejectsNonConfirmedAndMissingComponents(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)

	tests := []struct {
		name       string
		state      string
		hasSrc     bool
		components ConsumedMaterialDAOMock
		wantErr    error
	}{
		{name: "reject non confirmed", state: ProductionOrderStateDraft, components: ConsumedMaterialDAOMock{ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{{ItemID: 300, QtyPlanned: 1}}, nil
		}}, wantErr: ErrProductionOrderState},
		{name: "missing source location", state: ProductionOrderStateConfirmed, hasSrc: false, components: ConsumedMaterialDAOMock{ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{{ItemID: 300, QtyPlanned: 1}}, nil
		}}, wantErr: ErrProductionOrderLocation},
		{name: "no components", state: ProductionOrderStateConfirmed, hasSrc: true, wantErr: ErrConsumedMaterial},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: tt.state}
			if tt.hasSrc {
				order.SrcLocationID = &src
			}
			orders := ProductionOrderDAOMock{
				CRUDMock: dao.CRUDMock[ProductionOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
				},
			}
			svc := testProductionOrderService(orders, tt.components, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())

			_, err := svc.Plan(ctx, 1)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionOrderService_Plan_PropagatesReservationFailure(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateConfirmed, SrcLocationID: &src}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
		},
	}
	components := ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{{Base: model.Base{ID: 1}, ItemID: 300, QtyPlanned: 22}}, nil
		},
	}
	reservations := inventory.StockHoldDAOMock{}
	quants := inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Base: model.Base{ID: 3}, OrganizationID: &organizationID, Quantity: 10, ReservedQty: 0}, nil
		},
	}
	svc := testProductionOrderService(orders, components, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), reservations, quants, moSequenceMock())

	_, err := svc.Plan(ctx, 1)
	helper.AssertError(t, err, true, inventory.ErrHoldOverflow)
}

func TestProductionOrderService_Cancel_AllowsDraftConfirmedPlanned(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	tests := []struct {
		state string
	}{
		{state: ProductionOrderStateDraft},
		{state: ProductionOrderStateConfirmed},
		{state: ProductionOrderStatePlanned},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: tt.state}
			orders := ProductionOrderDAOMock{
				CRUDMock: dao.CRUDMock[ProductionOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
					UpdateFunc: func(_ context.Context, productionOrder *ProductionOrder) (*ProductionOrder, error) {
						return productionOrder, nil
					},
				},
			}
			svc := testProductionOrderService(orders, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())

			updated, err := svc.Cancel(ctx, 1)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if updated.State != ProductionOrderStateCancelled {
				t.Errorf("state = %s, want cancelled", updated.State)
			}
		})
	}
}

func TestProductionOrderService_Cancel_RejectsDoneAndCancelled(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	tests := []struct {
		state string
	}{
		{state: ProductionOrderStateDone},
		{state: ProductionOrderStateCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: tt.state}
			orders := ProductionOrderDAOMock{
				CRUDMock: dao.CRUDMock[ProductionOrder]{
					FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
				},
			}
			svc := testProductionOrderService(orders, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())

			_, err := svc.Cancel(ctx, 1)
			helper.AssertError(t, err, true, ErrProductionOrderState)
		})
	}
}

func TestProductionOrderService_RejectsNotFound(t *testing.T) {
	ctx := context.Background()
	svc := testProductionOrderService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, moSequenceMock())

	for _, action := range []string{"confirm", "plan", "cancel"} {
		t.Run(action, func(t *testing.T) {
			var err error
			switch action {
			case "confirm":
				_, err = svc.Confirm(ctx, 1)
			case "plan":
				_, err = svc.Plan(ctx, 1)
			case "cancel":
				_, err = svc.Cancel(ctx, 1)
			}
			helper.AssertError(t, err, true, ErrProductionOrderNotFound)
		})
	}
}

func TestProductionOrderService_Plan_RejectsUnknownQuant(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	src := uint64(10)
	order := &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: ProductionOrderStateConfirmed, SrcLocationID: &src}
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*ProductionOrder, error) { return order, nil },
		},
	}
	components := ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*ConsumedMaterial, error) {
			return []*ConsumedMaterial{{Base: model.Base{ID: 1}, ItemID: 300, QtyPlanned: 22}}, nil
		},
	}
	reservations := inventory.StockHoldDAOMock{}
	quants := inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return nil, nil
		},
	}
	svc := testProductionOrderService(orders, components, moRecipeMock(), moRecipeLinesMock(), moVariantMock(), moLocationMock(), reservations, quants, moSequenceMock())

	_, err := svc.Plan(ctx, 1)
	helper.AssertError(t, err, true, inventory.ErrBalanceNotFound)
}
