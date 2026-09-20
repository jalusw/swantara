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
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func laborTestService(productionOrder *ProductionOrder, moErr error, wo *ShopTask, woErr error, center *reference.WorkCenter, centerErr error, postErr error, balance float64, balanceErr error) ProductionService {
	orders := ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[ProductionOrder]{
			FindFunc:   func(_ context.Context, _ uint64) (*ProductionOrder, error) { return productionOrder, moErr },
			UpdateFunc: func(_ context.Context, m *ProductionOrder) (*ProductionOrder, error) { return m, nil },
		},
	}
	workOrders := ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[ShopTask]{
			FindFunc:   func(_ context.Context, _ uint64) (*ShopTask, error) { return wo, woErr },
			UpdateFunc: func(_ context.Context, w *ShopTask) (*ShopTask, error) { return w, nil },
		},
	}
	centers := dao.CRUDMock[reference.WorkCenter]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) { return center, centerErr },
	}
	poster := inventory.PosterMock{
		PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
			if postErr != nil {
				return nil, postErr
			}
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	lines := accounting.JournalLineDAOMock{
		BalanceByOriginAndAccountFunc: func(_ context.Context, _ string, _, _ uint64) (float64, error) {
			return balance, balanceErr
		},
	}
	seq := sequence.DAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 1, Number: "WO-1"}, nil
		},
	}
	return testProductionService(orders, ConsumedMaterialDAOMock{}, workOrders, ProductionStepDAOMock{}, centers, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, poster, lines, seq)
}

func inProgressMO() *ProductionOrder {
	productionOrder := plannedMO()
	productionOrder.State = ProductionOrderStateInProgress
	return productionOrder
}

func TestProductionService_RecordLabor(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dbErr := errors.New("db down")
	cost := 60.0
	center := &reference.WorkCenter{Base: model.Base{ID: 5}, CostPerHour: &cost}
	wo := &ShopTask{Base: model.Base{ID: 2}, ProductionOrderID: 1, WorkCenterID: 5}

	t.Run("records labor", func(t *testing.T) {
		svc := laborTestService(inProgressMO(), nil, wo, nil, center, nil, nil, 0, nil)
		got, err := svc.RecordLabor(ctx, 1, 2, 60, 500, 600, 700, now)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != ShopTaskStateDone || got.ActualMinutes != 60 {
			t.Errorf("shop task = %+v", got)
		}
	})

	cases := []struct {
		name            string
		minutes         float64
		wip             uint64
		applied         uint64
		productionOrder *ProductionOrder
		moErr           error
		wo              *ShopTask
		woErr           error
		center          *reference.WorkCenter
		centerErr       error
		postErr         error
		wantErr         error
	}{
		{name: "zero minutes", minutes: 0, wip: 600, applied: 700, productionOrder: inProgressMO(), wo: wo, center: center, wantErr: ErrShopTaskLabor},
		{name: "wip required", minutes: 60, applied: 700, productionOrder: inProgressMO(), wo: wo, center: center, wantErr: ErrWIPAccount},
		{name: "labor account required", minutes: 60, wip: 600, productionOrder: inProgressMO(), wo: wo, center: center, wantErr: ErrLaborAccount},
		{name: "productionOrder lookup error", minutes: 60, wip: 600, applied: 700, moErr: dbErr, wantErr: dbErr},
		{name: "productionOrder missing", minutes: 60, wip: 600, applied: 700, wantErr: ErrProductionOrderNotFound},
		{name: "productionOrder bad state", minutes: 60, wip: 600, applied: 700, productionOrder: plannedMO(), wo: wo, center: center, wantErr: ErrProductionOrderState},
		{name: "wo lookup error", minutes: 60, wip: 600, applied: 700, productionOrder: inProgressMO(), woErr: dbErr, wantErr: dbErr},
		{name: "wo missing", minutes: 60, wip: 600, applied: 700, productionOrder: inProgressMO(), wantErr: ErrShopTaskNotFound},
		{name: "wo mismatch", minutes: 60, wip: 600, applied: 700, productionOrder: inProgressMO(), wo: &ShopTask{Base: model.Base{ID: 2}, ProductionOrderID: 9, WorkCenterID: 5}, center: center, wantErr: ErrShopTaskMismatch},
		{name: "center lookup error", minutes: 60, wip: 600, applied: 700, productionOrder: inProgressMO(), wo: wo, centerErr: dbErr, wantErr: dbErr},
		{name: "center missing", minutes: 60, wip: 600, applied: 700, productionOrder: inProgressMO(), wo: wo, wantErr: ErrShopTaskWorkCenter},
		{name: "post error", minutes: 60, wip: 600, applied: 700, productionOrder: inProgressMO(), wo: wo, center: center, postErr: dbErr, wantErr: dbErr},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := laborTestService(tt.productionOrder, tt.moErr, tt.wo, tt.woErr, tt.center, tt.centerErr, tt.postErr, 0, nil)
			_, err := svc.RecordLabor(ctx, 1, 2, tt.minutes, 500, tt.wip, tt.applied, now)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionService_SettleVariance(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dbErr := errors.New("db down")

	newSvc := func(productionOrder *ProductionOrder, moErr error, balance float64, balanceErr error, postErr error) ProductionService {
		return laborTestService(productionOrder, moErr, nil, nil, nil, nil, postErr, balance, balanceErr)
	}

	t.Run("settles positive and negative", func(t *testing.T) {
		svc := newSvc(inProgressMO(), nil, 100, nil, nil)
		got, err := svc.SettleVariance(ctx, 1, 500, 600, 700, now)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != ProductionOrderStateDone {
			t.Errorf("state = %q", got.State)
		}

		svc = newSvc(inProgressMO(), nil, -50, nil, nil)
		got, err = svc.SettleVariance(ctx, 1, 500, 600, 700, now)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != ProductionOrderStateDone {
			t.Errorf("state = %q", got.State)
		}
	})

	t.Run("skips zero balance", func(t *testing.T) {
		svc := newSvc(inProgressMO(), nil, 0, nil, nil)
		got, err := svc.SettleVariance(ctx, 1, 500, 600, 700, now)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.State != ProductionOrderStateDone {
			t.Errorf("state = %q", got.State)
		}
	})

	cases := []struct {
		name            string
		wip             uint64
		variance        uint64
		productionOrder *ProductionOrder
		moErr           error
		balanceErr      error
		postErr         error
		wantErr         error
	}{
		{name: "wip required", variance: 700, productionOrder: inProgressMO(), wantErr: ErrWIPAccount},
		{name: "variance required", wip: 600, productionOrder: inProgressMO(), wantErr: ErrVarianceAccount},
		{name: "productionOrder lookup error", wip: 600, variance: 700, moErr: dbErr, wantErr: dbErr},
		{name: "productionOrder missing", wip: 600, variance: 700, wantErr: ErrProductionOrderNotFound},
		{name: "productionOrder bad state", wip: 600, variance: 700, productionOrder: plannedMO(), wantErr: ErrProductionOrderState},
		{name: "productionOrder no org", wip: 600, variance: 700, productionOrder: &ProductionOrder{Base: model.Base{ID: 1}, State: ProductionOrderStateInProgress}, wantErr: ErrProductionOrderOrganization},
		{name: "balance error", wip: 600, variance: 700, productionOrder: inProgressMO(), balanceErr: dbErr, wantErr: dbErr},
		{name: "post error", wip: 600, variance: 700, productionOrder: inProgressMO(), postErr: dbErr, wantErr: dbErr},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.productionOrder, tt.moErr, 100, tt.balanceErr, tt.postErr)
			_, err := svc.SettleVariance(ctx, 1, 500, tt.wip, tt.variance, now)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestProductionService_FindProductionLocation(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(locations inventory.StockLocationDAOMock) ProductionService {
		return testProductionService(ProductionOrderDAOMock{}, ConsumedMaterialDAOMock{}, ShopTaskDAOMock{}, ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, locations, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, sequence.DAOMock{})
	}

	t.Run("finds org location", func(t *testing.T) {
		svc := newSvc(inventory.StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
					return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 40}, OrganizationID: helper.Ptr(uint64(10))}}}, nil
				},
			},
		})
		got, err := svc.findProductionLocation(ctx, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.ID != 40 {
			t.Errorf("location = %d", got.ID)
		}
	})

	t.Run("returns nil when none match", func(t *testing.T) {
		svc := newSvc(inventory.StockLocationDAOMock{})
		got, err := svc.findProductionLocation(ctx, 10)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != nil {
			t.Errorf("location = %+v, want nil", got)
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
		_, err := svc.findProductionLocation(ctx, 10)
		helper.AssertError(t, err, true, dbErr)
	})
}
