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
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func plannedMO() *ProductionOrder {
	orgID := uint64(10)
	src := uint64(30)
	dst := uint64(40)
	return &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &orgID, ItemID: 200, RecipeID: helper.Ptr(uint64(1)), QtyToProduce: 10, SrcLocationID: &src, DstLocationID: &dst, State: ProductionOrderStatePlanned}
}

func prodTestService(productionOrder *ProductionOrder, moErr error, operations []*ProductionStep, opsErr error, center *reference.WorkCenter, centerErr error, seqErr error) ProductionService {
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc:   func(_ context.Context, _ uint64) (*ProductionOrder, error) { return productionOrder, moErr },
			UpdateFunc: func(_ context.Context, m *ProductionOrder) (*ProductionOrder, error) { return m, nil },
		},
	}
	ops := ProductionStepDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*ProductionStep, error) { return operations, opsErr },
	}
	centers := dao.CRUDMock[reference.WorkCenter]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) { return center, centerErr },
	}
	seq := sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			if seqErr != nil {
				return nil, seqErr
			}
			return &sequence.Reservation{Value: 1, Number: "WO-1"}, nil
		},
	}
	workOrders := ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[ShopTask]{
			CreateFunc: func(_ context.Context, w *ShopTask) (*ShopTask, error) { return w, nil },
		},
	}
	return testProductionService(orders, ConsumedMaterialDAOMock{}, workOrders, ops, centers, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, seq)
}

func TestProductionService_Start_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("starts planned order", func(t *testing.T) {
		svc := prodTestService(plannedMO(), nil, nil, nil, nil, nil, nil)
		got, err := svc.Start(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != ProductionOrderStateInProgress {
			t.Errorf("state = %q", got.State)
		}
	})

	t.Run("start failures", func(t *testing.T) {
		svc := prodTestService(nil, dbErr, nil, nil, nil, nil, nil)
		_, err := svc.Start(ctx, 1)
		helper.AssertError(t, err, true, dbErr)

		svc = prodTestService(nil, nil, nil, nil, nil, nil, nil)
		_, err = svc.Start(ctx, 1)
		helper.AssertError(t, err, true, ErrProductionOrderNotFound)

		inProgress := plannedMO()
		inProgress.State = ProductionOrderStateInProgress
		svc = prodTestService(inProgress, nil, nil, nil, nil, nil, nil)
		_, err = svc.Start(ctx, 1)
		helper.AssertError(t, err, true, ErrProductionOrderState)
	})
}

func TestProductionService_GenerateShopTasks_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	center := &reference.WorkCenter{Base: model.Base{ID: 5}}
	op := &ProductionStep{Base: model.Base{ID: 1}, WorkCenterID: helper.Ptr(uint64(5)), SetupMinutes: 10, TimeMinutes: 20}

	t.Run("generates orders", func(t *testing.T) {
		svc := prodTestService(plannedMO(), nil, []*ProductionStep{op}, nil, center, nil, nil)
		got, err := svc.GenerateShopTasks(ctx, 1)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if len(got) != 1 || got[0].PlannedMinutes != 30 {
			t.Errorf("orders = %+v", got)
		}
	})

	cases := []struct {
		name            string
		productionOrder *ProductionOrder
		moErr           error
		ops             []*ProductionStep
		opsErr          error
		center          *reference.WorkCenter
		centerE         error
		seqErr          error
		wantErr         error
	}{
		{name: "lookup error", moErr: dbErr, wantErr: dbErr},
		{name: "missing", wantErr: ErrProductionOrderNotFound},
		{name: "bad state", productionOrder: prodOrder(ProductionOrderStateDraft), wantErr: ErrProductionOrderState},
		{name: "no recipe", productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)), State: ProductionOrderStatePlanned}, wantErr: ErrProductionOrderRecipe},
		{name: "no org", productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, RecipeID: helper.Ptr(uint64(1)), State: ProductionOrderStatePlanned}, wantErr: ErrProductionOrderOrganization},
		{name: "operations error", productionOrder: plannedMO(), opsErr: dbErr, wantErr: dbErr},
		{name: "operation without center", productionOrder: plannedMO(), ops: []*ProductionStep{{Base: model.Base{ID: 1}}}, wantErr: ErrShopTaskWorkCenter},
		{name: "center lookup error", productionOrder: plannedMO(), ops: []*ProductionStep{op}, centerE: dbErr, wantErr: dbErr},
		{name: "center missing", productionOrder: plannedMO(), ops: []*ProductionStep{op}, wantErr: ErrShopTaskWorkCenter},
		{name: "sequence error", productionOrder: plannedMO(), ops: []*ProductionStep{op}, center: center, seqErr: dbErr, wantErr: dbErr},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := prodTestService(tt.productionOrder, tt.moErr, tt.ops, tt.opsErr, tt.center, tt.centerE, tt.seqErr)
			_, err := svc.GenerateShopTasks(ctx, 1)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionService_Consume_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	now := time.Now()

	newSvc := func(productionOrder *ProductionOrder, moErr error, comp *ConsumedMaterial, compErr error) ProductionService {
		svc := prodTestService(productionOrder, moErr, nil, nil, nil, nil, nil)
		svc.components = ConsumedMaterialDAOMock{
			CRUDMock: dao.CRUDMock[ConsumedMaterial]{
				FindFunc: func(_ context.Context, _ uint64) (*ConsumedMaterial, error) { return comp, compErr },
			},
		}
		return svc
	}
	inProgress := func() *ProductionOrder {
		productionOrder := plannedMO()
		productionOrder.State = ProductionOrderStateInProgress
		return productionOrder
	}
	comp := &ConsumedMaterial{Base: model.Base{ID: 3}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 10}

	cases := []struct {
		name            string
		qty             float64
		wip             uint64
		productionOrder *ProductionOrder
		moErr           error
		comp            *ConsumedMaterial
		compErr         error
		wantErr         error
	}{
		{name: "zero qty", qty: 0, wip: 1, productionOrder: inProgress(), comp: comp, wantErr: ErrConsumeQuantity},
		{name: "wip required", qty: 1, wip: 0, productionOrder: inProgress(), comp: comp, wantErr: ErrWIPAccount},
		{name: "productionOrder lookup error", qty: 1, wip: 1, moErr: dbErr, wantErr: dbErr},
		{name: "productionOrder missing", qty: 1, wip: 1, wantErr: ErrProductionOrderNotFound},
		{name: "productionOrder bad state", qty: 1, wip: 1, productionOrder: plannedMO(), comp: comp, wantErr: ErrProductionOrderState},
		{name: "productionOrder no location", qty: 1, wip: 1, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateInProgress}, comp: comp, wantErr: ErrProductionOrderLocation},
		{name: "component error", qty: 1, wip: 1, productionOrder: inProgress(), compErr: dbErr, wantErr: dbErr},
		{name: "component missing", qty: 1, wip: 1, productionOrder: inProgress(), wantErr: ErrConsumeComponent},
		{name: "over qty", qty: 99, wip: 1, productionOrder: inProgress(), comp: comp, wantErr: ErrConsumeQuantity},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.productionOrder, tt.moErr, tt.comp, tt.compErr)
			_, err := svc.Consume(ctx, 1, 3, tt.qty, 500, tt.wip, now)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionService_Produce_Errors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	now := time.Now()

	newSvc := func(productionOrder *ProductionOrder, moErr error, resolveErr error, noValuation bool) ProductionService {
		svc := prodTestService(productionOrder, moErr, nil, nil, nil, nil, nil)
		svc.resolver = inventory.ItemResolverMock{
			ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
				if resolveErr != nil {
					return inventory.ResolvedItem{}, resolveErr
				}
				return inventory.ResolvedItem{StandardCost: 5, Tracking: "none"}, nil
			},
		}
		_ = noValuation
		return svc
	}
	inProgress := func() *ProductionOrder {
		productionOrder := plannedMO()
		productionOrder.State = ProductionOrderStateInProgress
		return productionOrder
	}

	cases := []struct {
		name            string
		qty             float64
		wip             uint64
		productionOrder *ProductionOrder
		moErr           error
		resolveErr      error
		wantErr         error
	}{
		{name: "zero qty", qty: 0, wip: 1, productionOrder: inProgress(), wantErr: ErrProduceQuantity},
		{name: "wip required", qty: 1, wip: 0, productionOrder: inProgress(), wantErr: ErrWIPAccount},
		{name: "productionOrder lookup error", qty: 1, wip: 1, moErr: dbErr, wantErr: dbErr},
		{name: "productionOrder missing", qty: 1, wip: 1, wantErr: ErrProductionOrderNotFound},
		{name: "productionOrder bad state", qty: 1, wip: 1, productionOrder: plannedMO(), wantErr: ErrProductionOrderState},
		{name: "productionOrder no location", qty: 1, wip: 1, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateInProgress}, wantErr: ErrProductionOrderLocation},
		{name: "over qty", qty: 99, wip: 1, productionOrder: inProgress(), wantErr: ErrProduceQuantity},
		{name: "resolve error", qty: 1, wip: 1, productionOrder: inProgress(), resolveErr: dbErr, wantErr: dbErr},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.productionOrder, tt.moErr, tt.resolveErr, false)
			_, err := svc.Produce(ctx, 1, tt.qty, 500, tt.wip, now)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}
