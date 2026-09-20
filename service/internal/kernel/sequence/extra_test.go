package sequence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestSequenceService_NextTx(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates with tx", func(t *testing.T) {
		svc := NewSequenceService(DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*Reservation, error) {
				return &Reservation{Value: 1, Number: "X/1"}, nil
			},
		})
		got, err := svc.NextTx(ctx, nil, 1, "SO")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != "X/1" {
			t.Errorf("number = %q", got)
		}
	})

	t.Run("propagates tx reserve error", func(t *testing.T) {
		dbErr := errors.New("db down")
		svc := NewSequenceService(DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*Reservation, error) {
				return nil, dbErr
			},
		})
		_, err := svc.NextTx(ctx, nil, 1, "SO")
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("next propagates reserve error", func(t *testing.T) {
		dbErr := errors.New("db down")
		svc := NewSequenceService(DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*Reservation, error) {
				return nil, dbErr
			},
		})
		_, err := svc.Next(ctx, 1, "SO")
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("next with tx uses reserve tx", func(t *testing.T) {
		db, _ := query.NewMockDB(t)
		svc := NewSequenceService(DAOMock{
			ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*Reservation, error) {
				return &Reservation{Value: 2, Number: "X/2"}, nil
			},
		})
		got, err := svc.NextTx(ctx, db, 1, "SO")
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != "X/2" {
			t.Errorf("number = %q", got)
		}
	})

	t.Run("mock fallbacks", func(t *testing.T) {
		if _, err := (DAOMock{}).Reserve(ctx, 1, "SO", time.Now()); err != nil {
			t.Errorf("Reserve = %v", err)
		}
		if _, err := (DAOMock{}).ReserveTx(ctx, nil, 1, "SO", time.Now()); err != nil {
			t.Errorf("ReserveTx = %v", err)
		}
	})
}
