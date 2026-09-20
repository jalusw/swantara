package inventory

import (
	"context"
	"errors"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"gorm.io/gorm"
)

func TestHoldService_Release(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	t.Run("releases reservation", func(t *testing.T) {
		svc := NewHoldService(StockHoldDAOMock{}, StockBalanceDAOMock{})
		if err := svc.Release(ctx, 1); helper.AssertError(t, err, false, nil) {
			return
		}
	})

	t.Run("maps not found", func(t *testing.T) {
		svc := NewHoldService(StockHoldDAOMock{
			ReleaseFunc: func(_ context.Context, _ uint64) error { return gorm.ErrRecordNotFound },
		}, StockBalanceDAOMock{})
		helper.AssertError(t, svc.Release(ctx, 1), true, ErrHoldNotFound)
	})

	t.Run("propagates error", func(t *testing.T) {
		svc := NewHoldService(StockHoldDAOMock{
			ReleaseFunc: func(_ context.Context, _ uint64) error { return dbErr },
		}, StockBalanceDAOMock{})
		helper.AssertError(t, svc.Release(ctx, 1), true, dbErr)
	})

	t.Run("maps reservation conflict", func(t *testing.T) {
		orgID := uint64(10)
		svc := NewHoldService(StockHoldDAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ *uint64, _ float64) (*StockHold, error) {
				return nil, ErrHoldConflict
			},
		}, StockBalanceDAOMock{
			FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
				return &StockBalance{OrganizationID: &orgID, Quantity: 100}, nil
			},
		})
		_, err := svc.Reserve(ctx, 10, 1, 2, nil, amount.FromFloat64(5), nil)
		helper.AssertError(t, err, true, ErrHoldOverflow)
	})
}

func TestReorderService_Update_Location(t *testing.T) {
	ctx := context.Background()
	svc := NewReorderService(ReorderRuleDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{})
	_, err := svc.Update(ctx, 10, &ReorderRule{})
	helper.AssertError(t, err, true, ErrLocationRequired)
}
