package inventory

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestInventoryMock_All(t *testing.T) {
	ctx := context.Background()
	q := &query.Query{}

	t.Run("transactioner", func(t *testing.T) {
		bare := TransactionerMock{}
		if err := bare.Run(ctx, func(_ *gorm.DB) error { return nil }); err != nil {
			t.Errorf("Run = %v", err)
		}
		wired := TransactionerMock{
			RunFunc: func(_ context.Context, fn func(tx *gorm.DB) error) error { return fn(nil) },
		}
		if err := wired.Run(ctx, func(_ *gorm.DB) error { return nil }); err != nil {
			t.Errorf("Run = %v", err)
		}
	})

	t.Run("shipment dao", func(t *testing.T) {
		bare := ShipmentDAOMock{}
		if _, err := bare.CreateWithMovements(ctx, &Shipment{}, nil); err != nil {
			t.Errorf("CreateWithMovements = %v", err)
		}
		if _, err := bare.CreateWithMovementsTx(ctx, nil, &Shipment{}, nil); err != nil {
			t.Errorf("CreateWithMovementsTx = %v", err)
		}
		if _, err := bare.UpdateTx(ctx, nil, &Shipment{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
		wired := ShipmentDAOMock{
			CreateWithMovementsFunc: func(_ context.Context, p *Shipment, _ []*StockMovement) (*Shipment, error) {
				return p, nil
			},
			CreateWithMovementsTxFunc: func(_ context.Context, _ *gorm.DB, p *Shipment, _ []*StockMovement) (*Shipment, error) {
				return p, nil
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, p *Shipment) (*Shipment, error) { return p, nil },
		}
		if _, err := wired.CreateWithMovements(ctx, &Shipment{}, nil); err != nil {
			t.Errorf("CreateWithMovements = %v", err)
		}
		if _, err := wired.CreateWithMovementsTx(ctx, nil, &Shipment{}, nil); err != nil {
			t.Errorf("CreateWithMovementsTx = %v", err)
		}
		if _, err := wired.UpdateTx(ctx, nil, &Shipment{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
	})

	t.Run("movement dao", func(t *testing.T) {
		bare := StockMovementDAOMock{}
		if _, err := bare.CreateTx(ctx, nil, &StockMovement{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := bare.FindForUpdateTx(ctx, nil, 1); err != nil {
			t.Errorf("FindForUpdateTx = %v", err)
		}
		if _, err := bare.ApplyTx(ctx, nil, &StockMovement{}); err != nil {
			t.Errorf("ApplyTx = %v", err)
		}
		if err := bare.ApplyAllTx(ctx, nil, nil); err != nil {
			t.Errorf("ApplyAllTx = %v", err)
		}
		if _, err := bare.LedgerTotals(ctx, nil); err != nil {
			t.Errorf("LedgerTotals = %v", err)
		}
		if _, err := bare.ListByOrigin(ctx, "shipment", 1); err != nil {
			t.Errorf("ListByOrigin = %v", err)
		}
		wired := StockMovementDAOMock{
			CreateTxFunc:        func(_ context.Context, _ *gorm.DB, m *StockMovement) (*StockMovement, error) { return m, nil },
			FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64) (*StockMovement, error) { return &StockMovement{}, nil },
			ApplyTxFunc:         func(_ context.Context, _ *gorm.DB, m *StockMovement) (*StockMovement, error) { return m, nil },
			ApplyAllTxFunc:      func(_ context.Context, _ *gorm.DB, _ []*StockMovement) error { return nil },
			LedgerTotalsFunc:    func(_ context.Context, _ *uint64) ([]LedgerTotal, error) { return []LedgerTotal{{}}, nil },
			ListByOriginFunc: func(_ context.Context, _ string, _ uint64) ([]*StockMovement, error) {
				return []*StockMovement{{}}, nil
			},
			ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*StockMovement, error) { return []*StockMovement{{}}, nil },
		}
		if _, err := wired.CreateTx(ctx, nil, &StockMovement{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := wired.FindForUpdateTx(ctx, nil, 1); err != nil {
			t.Errorf("FindForUpdateTx = %v", err)
		}
		if _, err := wired.ApplyTx(ctx, nil, &StockMovement{}); err != nil {
			t.Errorf("ApplyTx = %v", err)
		}
		if err := wired.ApplyAllTx(ctx, nil, nil); err != nil {
			t.Errorf("ApplyAllTx = %v", err)
		}
		if _, err := wired.LedgerTotals(ctx, nil); err != nil {
			t.Errorf("LedgerTotals = %v", err)
		}
		if _, err := wired.ListByOrigin(ctx, "shipment", 1); err != nil {
			t.Errorf("ListByOrigin = %v", err)
		}
	})

	t.Run("quant dao", func(t *testing.T) {
		bare := StockBalanceDAOMock{}
		if _, err := bare.FindByKey(ctx, 1, 2, nil); err != nil {
			t.Errorf("FindByKey = %v", err)
		}
		if _, err := bare.Upsert(ctx, nil, 1, 2, nil, 5); err != nil {
			t.Errorf("Upsert = %v", err)
		}
		if _, err := bare.ListByItem(ctx, 1); err != nil {
			t.Errorf("ListByItem = %v", err)
		}
		if _, err := bare.ListAll(ctx, nil); err != nil {
			t.Errorf("ListAll = %v", err)
		}
		wired := StockBalanceDAOMock{
			FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) { return &StockBalance{}, nil },
			UpsertFunc: func(_ context.Context, _ *uint64, _, _ uint64, _ *uint64, _ float64) (*StockBalance, error) {
				return &StockBalance{}, nil
			},
			UpsertTxFunc: func(_ context.Context, _ *gorm.DB, _ *uint64, _, _ uint64, _ *uint64, _ float64) (*StockBalance, error) {
				return &StockBalance{}, nil
			},
			ListByItemFunc: func(_ context.Context, _ uint64) ([]*StockBalance, error) { return []*StockBalance{{}}, nil },
			ListAllFunc:    func(_ context.Context, _ *uint64) ([]*StockBalance, error) { return []*StockBalance{{}}, nil },
		}
		if _, err := wired.FindByKey(ctx, 1, 2, nil); err != nil {
			t.Errorf("FindByKey = %v", err)
		}
		if _, err := wired.Upsert(ctx, nil, 1, 2, nil, 5); err != nil {
			t.Errorf("Upsert = %v", err)
		}
		if _, err := wired.UpsertTx(ctx, nil, nil, 1, 2, nil, 5); err != nil {
			t.Errorf("UpsertTx = %v", err)
		}
		if _, err := wired.ListByItem(ctx, 1); err != nil {
			t.Errorf("ListByItem = %v", err)
		}
		if _, err := wired.ListAll(ctx, nil); err != nil {
			t.Errorf("ListAll = %v", err)
		}
	})

	t.Run("batch dao", func(t *testing.T) {
		bare := BatchDAOMock{}
		if _, err := bare.ListByItem(ctx, 1); err != nil {
			t.Errorf("ListByItem = %v", err)
		}
		if _, err := bare.ListInOrg(ctx, q, 10); err != nil {
			t.Errorf("ListInOrg = %v", err)
		}
		if _, err := bare.FindInOrg(ctx, 1, 10); err != nil {
			t.Errorf("FindInOrg = %v", err)
		}
	})

	t.Run("reservation dao", func(t *testing.T) {
		bare := StockHoldDAOMock{}
		if _, err := bare.ListByMovement(ctx, 1); err != nil {
			t.Errorf("ListByMovement = %v", err)
		}
		if err := bare.DeleteByMove(ctx, 1); err != nil {
			t.Errorf("DeleteByMove = %v", err)
		}
		if _, err := bare.Reserve(ctx, 1, nil, 5); err != nil {
			t.Errorf("Reserve = %v", err)
		}
		if err := bare.Release(ctx, 1); err != nil {
			t.Errorf("Release = %v", err)
		}
		if err := bare.ReleaseByMovement(ctx, 1); err != nil {
			t.Errorf("ReleaseByMovement = %v", err)
		}
		if _, err := bare.ListInOrg(ctx, q, 10); err != nil {
			t.Errorf("ListInOrg = %v", err)
		}
		if _, err := bare.FindInOrg(ctx, 1, 10); err != nil {
			t.Errorf("FindInOrg = %v", err)
		}
		if err := bare.ReleaseByMovementInOrg(ctx, 10, 1); err != nil {
			t.Errorf("ReleaseByMovementInOrg = %v", err)
		}
		wired := StockHoldDAOMock{
			CRUDMock: dao.CRUDMock[StockHold]{
				ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[StockHold], error) {
					return &query.Page[StockHold]{}, nil
				},
				FindFunc: func(_ context.Context, _ uint64) (*StockHold, error) {
					return &StockHold{}, nil
				},
			},
			ListByMovementFunc: func(_ context.Context, _ uint64) ([]*StockHold, error) { return []*StockHold{{}}, nil },
			ReserveFunc: func(_ context.Context, _ uint64, _ *uint64, _ float64) (*StockHold, error) {
				return &StockHold{}, nil
			},
			ReleaseByMovementFunc: func(_ context.Context, _ uint64) error { return nil },
		}
		if _, err := wired.ListByMovement(ctx, 1); err != nil {
			t.Errorf("ListByMovement = %v", err)
		}
		if _, err := wired.Reserve(ctx, 1, nil, 5); err != nil {
			t.Errorf("Reserve = %v", err)
		}
		if _, err := wired.ListInOrg(ctx, q, 10); err != nil {
			t.Errorf("ListInOrg = %v", err)
		}
		if _, err := wired.FindInOrg(ctx, 1, 10); err != nil {
			t.Errorf("FindInOrg = %v", err)
		}
		if err := wired.ReleaseByMovementInOrg(ctx, 10, 1); err != nil {
			t.Errorf("ReleaseByMovementInOrg = %v", err)
		}
	})

	t.Run("layer dao", func(t *testing.T) {
		bare := CostLayerDAOMock{}
		if _, err := bare.ListOpenByItem(ctx, 1); err != nil {
			t.Errorf("ListOpenByItem = %v", err)
		}
		if _, err := bare.ListOpenByItemForUpdateTx(ctx, nil, 1); err != nil {
			t.Errorf("ListOpenByItemForUpdateTx = %v", err)
		}
		if _, err := bare.ListOpenByItemInOrg(ctx, 1, 10); err != nil {
			t.Errorf("ListOpenByItemInOrg = %v", err)
		}
		if _, err := bare.ValueForItem(ctx, 1); err != nil {
			t.Errorf("ValueForItem = %v", err)
		}
		if _, err := bare.ListByMovement(ctx, 1); err != nil {
			t.Errorf("ListByMovement = %v", err)
		}
		if _, err := bare.CreateTx(ctx, nil, &CostLayer{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := bare.UpdateTx(ctx, nil, &CostLayer{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
		wired := CostLayerDAOMock{
			ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
				return []*CostLayer{{}}, nil
			},
			ListOpenByItemForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64) ([]*CostLayer, error) {
				return []*CostLayer{{}}, nil
			},
			ListOpenByItemInOrgFunc: func(_ context.Context, _, _ uint64) ([]*CostLayer, error) {
				return []*CostLayer{{}}, nil
			},
			ValueForItemFunc: func(_ context.Context, _ uint64) (float64, error) { return 1, nil },
			ListByMovementFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
				return []*CostLayer{{}}, nil
			},
			CreateTxFunc: func(_ context.Context, _ *gorm.DB, l *CostLayer) (*CostLayer, error) {
				return l, nil
			},
			UpdateTxFunc: func(_ context.Context, _ *gorm.DB, l *CostLayer) (*CostLayer, error) {
				return l, nil
			},
		}
		if _, err := wired.ListOpenByItem(ctx, 1); err != nil {
			t.Errorf("ListOpenByItem = %v", err)
		}
		if _, err := wired.ListOpenByItemForUpdateTx(ctx, nil, 1); err != nil {
			t.Errorf("ListOpenByItemForUpdateTx = %v", err)
		}
		if _, err := wired.ListOpenByItemInOrg(ctx, 1, 10); err != nil {
			t.Errorf("ListOpenByItemInOrg = %v", err)
		}
		if _, err := wired.ValueForItem(ctx, 1); err != nil {
			t.Errorf("ValueForItem = %v", err)
		}
		if _, err := wired.ListByMovement(ctx, 1); err != nil {
			t.Errorf("ListByMovement = %v", err)
		}
		if _, err := wired.CreateTx(ctx, nil, &CostLayer{}); err != nil {
			t.Errorf("CreateTx = %v", err)
		}
		if _, err := wired.UpdateTx(ctx, nil, &CostLayer{}); err != nil {
			t.Errorf("UpdateTx = %v", err)
		}
	})
}

func TestInventoryFixtures(t *testing.T) {
	if ShipmentFixture() == nil {
		t.Error("shipment = nil")
	}
	if StockMovementFixture() == nil {
		t.Error("movement = nil")
	}
	if StockBalanceFixture() == nil {
		t.Error("quant = nil")
	}
	if BatchFixture() == nil {
		t.Error("batch = nil")
	}
}
