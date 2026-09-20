package inventory

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func TestHoldService_Reserve_DecrementsAvailable(t *testing.T) {
	ctx := context.Background()
	quant := &StockBalance{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(1)), ItemID: 100, LocationID: 10, Quantity: 10, ReservedQty: 2}
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return quant, nil
		},
	}
	svc := NewHoldService(StockHoldDAOMock{}, quants)

	reservation, err := svc.Reserve(ctx, 1, 100, 10, nil, amount.FromFloat64(5), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reservation.Qty != 5 || reservation.BalanceID != 3 {
		t.Errorf("reservation = %+v, want qty 5 on quant 3", reservation)
	}
}

func TestHoldService_Reserve_RejectsOverflow(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return &StockBalance{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(1)), Quantity: 10, ReservedQty: 8}, nil
		},
	}
	svc := NewHoldService(StockHoldDAOMock{}, quants)

	_, err := svc.Reserve(ctx, 1, 100, 10, nil, amount.FromFloat64(5), nil)
	if helper.AssertError(t, err, true, ErrHoldOverflow) {
		return
	}
}

func TestHoldService_Reserve_RejectsMissingQuant(t *testing.T) {
	ctx := context.Background()
	svc := NewHoldService(StockHoldDAOMock{}, StockBalanceDAOMock{})

	_, err := svc.Reserve(ctx, 1, 100, 10, nil, amount.FromFloat64(5), nil)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}

func TestHoldService_Reserve_RejectsQuantFromOtherOrganization(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return &StockBalance{Base: model.Base{ID: 3}, OrganizationID: helper.Ptr(uint64(2)), Quantity: 10}, nil
		},
	}
	svc := NewHoldService(StockHoldDAOMock{}, quants)

	_, err := svc.Reserve(ctx, 1, 100, 10, nil, amount.FromFloat64(5), nil)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}

func TestHoldService_Reserve_RejectsNonPositiveQty(t *testing.T) {
	ctx := context.Background()
	svc := NewHoldService(StockHoldDAOMock{}, StockBalanceDAOMock{})

	_, err := svc.Reserve(ctx, 1, 100, 10, nil, amount.Zero(), nil)
	if helper.AssertError(t, err, true, ErrMovementQty) {
		return
	}
}

func TestHoldService_Reserve_PropagatesFindByKeyError(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return nil, ErrBalanceNotFound
		},
	}
	svc := NewHoldService(StockHoldDAOMock{}, quants)

	_, err := svc.Reserve(ctx, 1, 100, 10, nil, amount.FromFloat64(5), nil)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}
