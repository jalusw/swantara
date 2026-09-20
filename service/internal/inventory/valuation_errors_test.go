package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type valFixture struct {
	movement   *StockMovement
	moveErr    error
	usage      string
	locErr     error
	resolved   ResolvedItem
	resolveErr error
}

func baseValMove() *StockMovement {
	orgID := uint64(10)
	return &StockMovement{Base: model.Base{ID: 1}, OrganizationID: &orgID, ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 40, State: MovementStateDraft}
}

func baseResolved() ResolvedItem {
	return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, StockInputAccountID: 1310, CogsAccountID: 1320}, CostMethod: "average"}
}

func valSvc(f valFixture) ValuationService {
	moveMock := StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[StockMovement]{
			FindFunc: func(_ context.Context, _ uint64) (*StockMovement, error) { return f.movement, f.moveErr },
		},
		FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ uint64) (*StockMovement, error) {
			return f.movement, f.moveErr
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, id uint64) (*reference.StockLocation, error) {
				if f.locErr != nil {
					return nil, f.locErr
				}
				return &reference.StockLocation{Base: model.Base{ID: id}, Usage: f.usage}, nil
			},
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) { return f.resolved, f.resolveErr },
	}
	return NewValuationService(moveMock, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})
}

func TestValuation_Receive_Errors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dbErr := errors.New("db down")

	newSvc := func(usage string, mutate func(*valFixture)) ValuationService {
		f := valFixture{movement: baseValMove(), usage: usage, resolved: baseResolved()}
		if mutate != nil {
			mutate(&f)
		}
		return valSvc(f)
	}

	cases := []struct {
		name    string
		usage   string
		mutate  func(*valFixture)
		cost    amount.Amount
		wantErr error
	}{
		{name: "movement error", usage: "supplier", mutate: func(f *valFixture) { f.moveErr = dbErr }, wantErr: dbErr},
		{name: "movement missing", usage: "supplier", mutate: func(f *valFixture) { f.movement = nil }, wantErr: ErrMovementNotFound},
		{name: "movement done", usage: "supplier", mutate: func(f *valFixture) { f.movement.State = MovementStateDone }, wantErr: ErrMovementState},
		{name: "negative cost", usage: "supplier", cost: amount.FromFloat64(-1), wantErr: ErrNegativeCost},
		{name: "location error", usage: "supplier", mutate: func(f *valFixture) { f.locErr = dbErr }, wantErr: dbErr},
		{name: "wrong usage", usage: "customer", wantErr: ErrNotReceipt},
		{name: "resolve error", usage: "supplier", mutate: func(f *valFixture) { f.resolveErr = dbErr }, wantErr: dbErr},
		{name: "batch required", usage: "supplier", mutate: func(f *valFixture) {
			f.resolved.Tracking = "batch"
		}, wantErr: ErrBatchRequired},
		{name: "org missing", usage: "supplier", mutate: func(f *valFixture) { f.movement.OrganizationID = nil }, wantErr: ErrOrganizationMissing},
		{name: "no valuation account", usage: "supplier", mutate: func(f *valFixture) {
			f.resolved.StockValuationAccountID = 0
		}, wantErr: ErrValuationAccount},
		{name: "no input account", usage: "supplier", mutate: func(f *valFixture) {
			f.resolved.StockInputAccountID = 0
		}, wantErr: ErrInputAccount},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.usage, tt.mutate)
			cost := amount.FromFloat64(25)
			if tt.name == "negative cost" {
				cost = tt.cost
			}
			_, err := svc.Receive(ctx, 1, cost, 500, now)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestValuation_ShipScrap_Errors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dbErr := errors.New("db down")

	t.Run("ship validations", func(t *testing.T) {
		newSvc := func(usage string, mutate func(*valFixture)) ValuationService {
			f := valFixture{movement: baseValMove(), usage: usage, resolved: baseResolved()}
			if mutate != nil {
				mutate(&f)
			}
			return valSvc(f)
		}
		cases := []struct {
			name    string
			usage   string
			mutate  func(*valFixture)
			wantErr error
		}{
			{name: "movement error", usage: "customer", mutate: func(f *valFixture) { f.moveErr = dbErr }, wantErr: dbErr},
			{name: "movement missing", usage: "customer", mutate: func(f *valFixture) { f.movement = nil }, wantErr: ErrMovementNotFound},
			{name: "movement done", usage: "customer", mutate: func(f *valFixture) { f.movement.State = MovementStateDone }, wantErr: ErrMovementState},
			{name: "location error", usage: "customer", mutate: func(f *valFixture) { f.locErr = dbErr }, wantErr: dbErr},
			{name: "wrong usage", usage: "supplier", wantErr: ErrNotShipment},
			{name: "resolve error", usage: "customer", mutate: func(f *valFixture) { f.resolveErr = dbErr }, wantErr: dbErr},
			{name: "no valuation account", usage: "customer", mutate: func(f *valFixture) {
				f.resolved.StockValuationAccountID = 0
			}, wantErr: ErrValuationAccount},
			{name: "no cogs account", usage: "customer", mutate: func(f *valFixture) {
				f.resolved.CogsAccountID = 0
			}, wantErr: ErrCogsAccount},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.usage, tt.mutate)
				_, err := svc.ShipTx(ctx, nil, 1, 500, now)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})

	t.Run("scrap validations", func(t *testing.T) {
		newSvc := func(usage string, mutate func(*valFixture)) ValuationService {
			f := valFixture{movement: baseValMove(), usage: usage, resolved: baseResolved()}
			if mutate != nil {
				mutate(&f)
			}
			return valSvc(f)
		}
		cases := []struct {
			name    string
			usage   string
			mutate  func(*valFixture)
			wantErr error
		}{
			{name: "movement error", usage: "scrap", mutate: func(f *valFixture) { f.moveErr = dbErr }, wantErr: dbErr},
			{name: "movement missing", usage: "scrap", mutate: func(f *valFixture) { f.movement = nil }, wantErr: ErrMovementNotFound},
			{name: "movement done", usage: "scrap", mutate: func(f *valFixture) { f.movement.State = MovementStateCancelled }, wantErr: ErrMovementState},
			{name: "location error", usage: "scrap", mutate: func(f *valFixture) { f.locErr = dbErr }, wantErr: dbErr},
			{name: "wrong usage", usage: "customer", wantErr: ErrNotScrap},
			{name: "resolve error", usage: "scrap", mutate: func(f *valFixture) { f.resolveErr = dbErr }, wantErr: dbErr},
			{name: "no valuation account", usage: "scrap", mutate: func(f *valFixture) {
				f.resolved.StockValuationAccountID = 0
			}, wantErr: ErrValuationAccount},
			{name: "no expense account", usage: "scrap", wantErr: ErrScrapAccount},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.usage, tt.mutate)
				_, err := svc.Scrap(ctx, 1, 500, 0, now)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})
}

func TestValuation_ProduceConsume_Errors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dbErr := errors.New("db down")

	t.Run("produce validations", func(t *testing.T) {
		newSvc := func(usage string, mutate func(*valFixture)) ValuationService {
			f := valFixture{movement: baseValMove(), usage: usage, resolved: baseResolved()}
			if mutate != nil {
				mutate(&f)
			}
			return valSvc(f)
		}
		cases := []struct {
			name    string
			usage   string
			cost    amount.Amount
			mutate  func(*valFixture)
			wantErr error
		}{
			{name: "negative cost", usage: "production", cost: amount.FromFloat64(-1), wantErr: ErrNegativeCost},
			{name: "movement error", usage: "production", mutate: func(f *valFixture) { f.moveErr = dbErr }, wantErr: dbErr},
			{name: "movement missing", usage: "production", mutate: func(f *valFixture) { f.movement = nil }, wantErr: ErrMovementNotFound},
			{name: "movement done", usage: "production", mutate: func(f *valFixture) { f.movement.State = MovementStateDone }, wantErr: ErrMovementState},
			{name: "location error", usage: "production", mutate: func(f *valFixture) { f.locErr = dbErr }, wantErr: dbErr},
			{name: "wrong usage", usage: "customer", wantErr: ErrNotProduction},
			{name: "resolve error", usage: "production", mutate: func(f *valFixture) { f.resolveErr = dbErr }, wantErr: dbErr},
			{name: "no valuation account", usage: "production", mutate: func(f *valFixture) {
				f.resolved.StockValuationAccountID = 0
			}, wantErr: ErrValuationAccount},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.usage, tt.mutate)
				cost := amount.FromFloat64(10)
				if tt.name == "negative cost" {
					cost = tt.cost
				}
				_, err := svc.Produce(ctx, 1, cost, 500, 600, now)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})

	t.Run("consume validations", func(t *testing.T) {
		newSvc := func(usage string, mutate func(*valFixture)) ValuationService {
			f := valFixture{movement: baseValMove(), usage: usage, resolved: baseResolved()}
			if mutate != nil {
				mutate(&f)
			}
			return valSvc(f)
		}
		cases := []struct {
			name    string
			usage   string
			mutate  func(*valFixture)
			wantErr error
		}{
			{name: "movement error", usage: "production", mutate: func(f *valFixture) { f.moveErr = dbErr }, wantErr: dbErr},
			{name: "movement missing", usage: "production", mutate: func(f *valFixture) { f.movement = nil }, wantErr: ErrMovementNotFound},
			{name: "movement done", usage: "production", mutate: func(f *valFixture) { f.movement.State = MovementStateDone }, wantErr: ErrMovementState},
			{name: "location error", usage: "production", mutate: func(f *valFixture) { f.locErr = dbErr }, wantErr: dbErr},
			{name: "wrong usage", usage: "customer", wantErr: ErrNotConsumption},
			{name: "resolve error", usage: "production", mutate: func(f *valFixture) { f.resolveErr = dbErr }, wantErr: dbErr},
			{name: "no valuation account", usage: "production", mutate: func(f *valFixture) {
				f.resolved.StockValuationAccountID = 0
			}, wantErr: ErrValuationAccount},
			{name: "no wip account", usage: "production", wantErr: ErrWIPAccount},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.usage, tt.mutate)
				wip := uint64(600)
				if tt.name == "no wip account" {
					wip = 0
				}
				_, err := svc.Consume(ctx, 1, 500, wip, now)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})

	t.Run("return to supplier validations", func(t *testing.T) {
		newSvc := func(usage string, mutate func(*valFixture)) ValuationService {
			f := valFixture{movement: baseValMove(), usage: usage, resolved: baseResolved()}
			if mutate != nil {
				mutate(&f)
			}
			return valSvc(f)
		}
		cases := []struct {
			name    string
			usage   string
			mutate  func(*valFixture)
			wantErr error
		}{
			{name: "movement error", usage: "supplier", mutate: func(f *valFixture) { f.moveErr = dbErr }, wantErr: dbErr},
			{name: "movement missing", usage: "supplier", mutate: func(f *valFixture) { f.movement = nil }, wantErr: ErrMovementNotFound},
			{name: "movement done", usage: "supplier", mutate: func(f *valFixture) { f.movement.State = MovementStateDone }, wantErr: ErrMovementState},
			{name: "location error", usage: "supplier", mutate: func(f *valFixture) { f.locErr = dbErr }, wantErr: dbErr},
			{name: "wrong usage", usage: "customer", wantErr: ErrNotSupplierReturn},
			{name: "resolve error", usage: "supplier", mutate: func(f *valFixture) { f.resolveErr = dbErr }, wantErr: dbErr},
			{name: "no valuation account", usage: "supplier", mutate: func(f *valFixture) {
				f.resolved.StockValuationAccountID = 0
			}, wantErr: ErrValuationAccount},
			{name: "no input account", usage: "supplier", mutate: func(f *valFixture) {
				f.resolved.StockInputAccountID = 0
			}, wantErr: ErrInputAccount},
		}
		for _, tt := range cases {
			t.Run(tt.name, func(t *testing.T) {
				svc := newSvc(tt.usage, tt.mutate)
				_, err := svc.ReturnToSupplier(ctx, 1, 500, now)
				helper.AssertError(t, err, true, tt.wantErr)
			})
		}
	})
}

func TestValuation_RestockTx_Errors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	dbErr := errors.New("db down")

	newSvc := func(usage string, mutate func(*valFixture)) ValuationService {
		f := valFixture{movement: baseValMove(), usage: usage, resolved: baseResolved()}
		if mutate != nil {
			mutate(&f)
		}
		return valSvc(f)
	}
	cases := []struct {
		name    string
		usage   string
		mutate  func(*valFixture)
		wantErr error
	}{
		{name: "movement error", usage: "customer", mutate: func(f *valFixture) { f.moveErr = dbErr }, wantErr: dbErr},
		{name: "movement missing", usage: "customer", mutate: func(f *valFixture) { f.movement = nil }, wantErr: ErrMovementNotFound},
		{name: "movement done", usage: "customer", mutate: func(f *valFixture) { f.movement.State = MovementStateDone }, wantErr: ErrMovementState},
		{name: "negative cost", usage: "customer", wantErr: ErrNegativeCost},
		{name: "location error", usage: "customer", mutate: func(f *valFixture) { f.locErr = dbErr }, wantErr: dbErr},
		{name: "wrong usage", usage: "supplier", wantErr: ErrNotReturn},
		{name: "resolve error", usage: "customer", mutate: func(f *valFixture) { f.resolveErr = dbErr }, wantErr: dbErr},
		{name: "no valuation account", usage: "customer", mutate: func(f *valFixture) {
			f.resolved.StockValuationAccountID = 0
		}, wantErr: ErrValuationAccount},
		{name: "no cogs account", usage: "customer", mutate: func(f *valFixture) {
			f.resolved.CogsAccountID = 0
		}, wantErr: ErrCogsAccount},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			svc := newSvc(tt.usage, tt.mutate)
			cost := amount.FromFloat64(10)
			if tt.name == "negative cost" {
				cost = amount.FromFloat64(-1)
			}
			_, err := svc.RestockTx(ctx, nil, 1, cost, 500, now)
			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestEnsureMovementApplicable(t *testing.T) {
	helper.AssertError(t, EnsureMovementApplicable(nil), true, ErrMovementNotFound)
	done := &StockMovement{State: MovementStateDone}
	helper.AssertError(t, EnsureMovementApplicable(done), true, ErrMovementState)
	if err := EnsureMovementApplicable(baseValMove()); err != nil {
		t.Errorf("applicable = %v", err)
	}
}
