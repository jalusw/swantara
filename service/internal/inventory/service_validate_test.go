package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestWarehouseService_Validate(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("warehouse validations", func(t *testing.T) {
		svc := NewWarehouseService(WarehouseDAOMock{}, StockLocationDAOMock{})
		_, err := svc.CreateWarehouse(ctx, &reference.Warehouse{})
		helper.AssertError(t, err, true, ErrWarehouseNameRequired)

		dup := WarehouseDAOMock{
			CRUDMock: dao.CRUDMock[reference.Warehouse]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
					return &reference.Warehouse{Base: model.Base{ID: 9}}, nil
				},
			},
		}
		svc = NewWarehouseService(dup, StockLocationDAOMock{})
		_, err = svc.CreateWarehouse(ctx, &reference.Warehouse{Name: "Main", Code: helper.Ptr("MAIN")})
		helper.AssertError(t, err, true, ErrWarehouseCodeTaken)

		failing := WarehouseDAOMock{
			CRUDMock: dao.CRUDMock[reference.Warehouse]{
				SearchFunc: func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
					return nil, dbErr
				},
			},
		}
		svc = NewWarehouseService(failing, StockLocationDAOMock{})
		_, err = svc.CreateWarehouse(ctx, &reference.Warehouse{Name: "Main", Code: helper.Ptr("MAIN")})
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("location validations", func(t *testing.T) {
		svc := NewWarehouseService(WarehouseDAOMock{}, StockLocationDAOMock{})
		_, err := svc.CreateLocation(ctx, &reference.StockLocation{})
		helper.AssertError(t, err, true, ErrLocationNameRequired)

		_, err = svc.CreateLocation(ctx, &reference.StockLocation{Name: "Shelf", Usage: "mystery"})
		helper.AssertError(t, err, true, ErrInvalidLocationType)

		parentErr := StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) { return nil, dbErr },
			},
		}
		svc = NewWarehouseService(WarehouseDAOMock{}, parentErr)
		_, err = svc.CreateLocation(ctx, &reference.StockLocation{Name: "Shelf", Usage: "internal", ParentID: helper.Ptr(uint64(1))})
		helper.AssertError(t, err, true, dbErr)

		noParent := StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) { return nil, nil },
			},
		}
		svc = NewWarehouseService(WarehouseDAOMock{}, noParent)
		_, err = svc.CreateLocation(ctx, &reference.StockLocation{Name: "Shelf", Usage: "internal", ParentID: helper.Ptr(uint64(1))})
		helper.AssertError(t, err, true, ErrLocationParent)
	})
}

func TestLedgerService_ValidateMove(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(locErr error, loc *reference.StockLocation) LedgerService {
		return NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) { return loc, locErr },
			},
		}, TransactionerMock{})
	}
	validMove := func() *StockMovement {
		return &StockMovement{ItemID: 100, Qty: 5, SrcLocationID: 30, DstLocationID: 40}
	}
	loc := &reference.StockLocation{Base: model.Base{ID: 30}}

	t.Run("rejects empty item and qty", func(t *testing.T) {
		svc := newSvc(nil, loc)
		_, err := svc.CreateMovement(ctx, &StockMovement{Qty: 5})
		helper.AssertError(t, err, true, ErrMovementItem)

		_, err = svc.CreateMovement(ctx, &StockMovement{ItemID: 100})
		helper.AssertError(t, err, true, ErrMovementQty)
	})

	t.Run("location errors", func(t *testing.T) {
		svc := newSvc(dbErr, nil)
		_, err := svc.CreateMovement(ctx, validMove())
		helper.AssertError(t, err, true, dbErr)

		svc = newSvc(nil, nil)
		_, err = svc.CreateMovement(ctx, validMove())
		helper.AssertError(t, err, true, ErrLocationNotFound)
	})

	t.Run("destination errors", func(t *testing.T) {
		calls := 0
		svc := NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{
			CRUDMock: dao.CRUDMock[reference.StockLocation]{
				FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
					calls++
					if calls == 1 {
						return loc, nil
					}
					return nil, dbErr
				},
			},
		}, TransactionerMock{})
		_, err := svc.CreateMovement(ctx, validMove())
		helper.AssertError(t, err, true, dbErr)
	})
}

var _ = query.Query{}
