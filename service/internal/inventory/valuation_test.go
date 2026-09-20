package inventory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestBatchService_Create(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "rejects when tracking disabled",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}
				svc := NewBatchService(BatchDAOMock{}, resolver)

				_, err := svc.Create(ctx, 1, &Batch{ItemID: 100, Name: "LOT-1"})
				if helper.AssertError(t, err, true, ErrBatchTrackingDisabled) {
					return
				}
			},
		},
		{
			name: "rejects duplicate name",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "batch", OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}
				lots := BatchDAOMock{
					ListByItemFunc: func(_ context.Context, _ uint64) ([]*Batch, error) {
						return []*Batch{{Base: model.Base{ID: 1}, ItemID: 100, Name: "LOT-1"}}, nil
					},
				}
				svc := NewBatchService(lots, resolver)

				_, err := svc.Create(ctx, 1, &Batch{ItemID: 100, Name: "LOT-1"})
				if helper.AssertError(t, err, true, ErrBatchNameTaken) {
					return
				}
			},
		},
		{
			name: "rejects item from other organization",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "batch", OrganizationID: helper.Ptr(uint64(2))}, nil
					},
				}
				svc := NewBatchService(BatchDAOMock{}, resolver)

				_, err := svc.Create(ctx, 1, &Batch{ItemID: 100, Name: "LOT-1"})
				if helper.AssertError(t, err, true, ErrItemNotFound) {
					return
				}
			},
		},
		{
			name: "succeeds",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "batch", OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}
				created := false
				lots := BatchDAOMock{
					CRUDMock: dao.CRUDMock[Batch]{
						CreateFunc: func(_ context.Context, batch *Batch) (*Batch, error) {
							created = true
							return batch, nil
						},
					},
				}
				svc := NewBatchService(lots, resolver)

				batch := &Batch{ItemID: 100, Name: "LOT-1"}
				got, err := svc.Create(ctx, 1, batch)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !created || got.Name != "LOT-1" {
					t.Errorf("created = %v, name = %s, want true/LOT-1", created, got.Name)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestBatchService_Update(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "succeeds",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "batch", OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}
				updated := false
				lots := BatchDAOMock{
					ListByItemFunc: func(_ context.Context, _ uint64) ([]*Batch, error) {
						return []*Batch{{Base: model.Base{ID: 2}, ItemID: 100, Name: "LOT-2"}}, nil
					},
					CRUDMock: dao.CRUDMock[Batch]{
						UpdateFunc: func(_ context.Context, batch *Batch) (*Batch, error) {
							updated = true
							return batch, nil
						},
					},
				}
				svc := NewBatchService(lots, resolver)

				batch := &Batch{Base: model.Base{ID: 1}, ItemID: 100, Name: "LOT-1"}
				got, err := svc.Update(ctx, 1, batch)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !updated || got.ID != 1 {
					t.Errorf("updated = %v, id = %d, want true/1", updated, got.ID)
				}
			},
		},
		{
			name: "rejects duplicate name",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "batch", OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}
				lots := BatchDAOMock{
					ListByItemFunc: func(_ context.Context, _ uint64) ([]*Batch, error) {
						return []*Batch{{Base: model.Base{ID: 2}, ItemID: 100, Name: "LOT-1"}}, nil
					},
				}
				svc := NewBatchService(lots, resolver)

				_, err := svc.Update(ctx, 1, &Batch{Base: model.Base{ID: 1}, ItemID: 100, Name: "LOT-1"})
				if helper.AssertError(t, err, true, ErrBatchNameTaken) {
					return
				}
			},
		},
		{
			name: "rejects empty name",
			run: func(t *testing.T) {
				ctx := context.Background()
				svc := NewBatchService(BatchDAOMock{}, ItemResolverMock{})

				_, err := svc.Update(ctx, 1, &Batch{ItemID: 100})
				if helper.AssertError(t, err, true, ErrBatchNameTaken) {
					return
				}
			},
		},
		{
			name: "propagates resolve error",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{}, ErrItemNotFound
					},
				}
				svc := NewBatchService(BatchDAOMock{}, resolver)

				_, err := svc.Update(ctx, 1, &Batch{ItemID: 100, Name: "LOT-1"})
				if helper.AssertError(t, err, true, ErrItemNotFound) {
					return
				}
			},
		},
		{
			name: "propagates list error",
			run: func(t *testing.T) {
				ctx := context.Background()
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "batch", OrganizationID: helper.Ptr(uint64(1))}, nil
					},
				}
				lots := BatchDAOMock{
					ListByItemFunc: func(_ context.Context, _ uint64) ([]*Batch, error) {
						return nil, ErrBatchNotFound
					},
				}
				svc := NewBatchService(lots, resolver)

				_, err := svc.Update(ctx, 1, &Batch{ItemID: 100, Name: "LOT-1"})
				if helper.AssertError(t, err, true, ErrBatchNotFound) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_Receive(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "posts GRNI with layer",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, StockInputAccountID: 1310,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				layer, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if layer.Value != 250 || layer.RemainingQty != 10 {
					t.Errorf("layer = value %v remaining %v, want 250/10", layer.Value, layer.RemainingQty)
				}
				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 1300 || got.Debit.Float64() != 250 {
					t.Errorf("line 0 = %+v, want debit 250 on valuation account 1300", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1310 || got.Credit.Float64() != 250 {
					t.Errorf("line 1 = %+v, want credit 250 on stock input account 1310", got)
				}
			},
		},
		{
			name: "post failure aborts before finalize",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, StockInputAccountID: 1310,
						}}, nil
					},
				}
				finalized := false
				layers := CostLayerDAOMock{
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *CostLayer) (*CostLayer, error) {
						finalized = true
						return nil, nil
					},
				}
				poster := PosterMock{
					PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
						return nil, errors.New("unbalanced posting")
					},
				}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
				if err == nil {
					t.Fatal("expected error when posting fails")
				}
				if finalized {
					t.Error("layer must not be finalized when posting fails inside the transaction")
				}
			},
		},
		{
			name: "rejects non-supplier source",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 1}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 20, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "internal"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(10), 50, time.Now())
				if helper.AssertError(t, err, true, ErrNotReceipt) {
					return
				}
			},
		},
		{
			name: "standard cost overrides actual",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 1}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "standard", StandardCost: 25, Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, StockInputAccountID: 1310,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				layer, err := svc.Receive(ctx, 1, amount.FromFloat64(30), 50, time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if layer.Value != 250 || layer.RemainingQty != 10 {
					t.Errorf("layer = value %v remaining %v, want 250/10 (10 x std 25)", layer.Value, layer.RemainingQty)
				}
				if *layer.UnitCost != 25 {
					t.Errorf("unit cost = %v, want 25", *layer.UnitCost)
				}
				if got := posted[0].Lines[0]; got.Debit.Float64() != 250 {
					t.Errorf("line 0 debit = %v, want 250 at standard cost", got.Debit.Float64())
				}
			},
		},
		{
			name: "propagates apply error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
						return &StockMovement{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft}, nil
					}),
					ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *StockMovement) (*StockMovement, error) {
						return nil, ErrMovementNotFound
					},
				}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, StockInputAccountID: 1310}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "propagates create layer error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, StockInputAccountID: 1310}}, nil
					},
				}
				layers := CostLayerDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *CostLayer) (*CostLayer, error) {
						return nil, ErrNoValuation
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrNoValuation) {
					return
				}
			},
		},
		{
			name: "propagates finalize error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, StockInputAccountID: 1310}}, nil
					},
				}
				poster := PosterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				layers := CostLayerDAOMock{
					UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *CostLayer) (*CostLayer, error) {
						return nil, ErrNoValuation
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrNoValuation) {
					return
				}
			},
		},
		{
			name: "rejects missing movement",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return nil, nil
				})}
				svc := NewValuationService(movements, CostLayerDAOMock{}, StockLocationDAOMock{}, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "rejects missing input account",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrInputAccount) {
					return
				}
			},
		},
		{
			name: "rejects negative standard cost",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 30, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "standard", StandardCost: -5, Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, StockInputAccountID: 1310,
						}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Receive(ctx, 1, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrNegativeCost) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_Ship(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "consumes FIFO",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 2}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 7, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 4, RemainingValue: 100, UnitCost: helper.Ptr(25.0)},
							{Base: model.Base{ID: 2}, RemainingQty: 6, RemainingValue: 180, UnitCost: helper.Ptr(30.0)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				updated := []*CostLayer{}
				layers.UpdateFunc = func(_ context.Context, layer *CostLayer) (*CostLayer, error) {
					updated = append(updated, layer)
					return layer, nil
				}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Ship(ctx, 2, 50, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 1320 || got.Debit.Float64() != 190 {
					t.Errorf("line 0 = %+v, want COGS debit 190 (4x25 + 3x30)", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 190 {
					t.Errorf("line 1 = %+v, want inventory credit 190", got)
				}

				remaining := map[uint64]float64{}
				remainingValue := map[uint64]float64{}
				for _, layer := range updated {
					remaining[layer.ID] = layer.RemainingQty
					remainingValue[layer.ID] = layer.RemainingValue
				}
				if remaining[1] != 0 || remainingValue[1] != 0 {
					t.Errorf("layer 1 = qty %v value %v, want fully consumed", remaining[1], remainingValue[1])
				}
				if remaining[2] != 3 || remainingValue[2] != 90 {
					t.Errorf("layer 2 = qty %v value %v, want 3/90", remaining[2], remainingValue[2])
				}
			},
		},
		{
			name: "consumes average",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 2}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 7, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "average", Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 4, RemainingValue: 100, UnitCost: helper.Ptr(25.0)},
							{Base: model.Base{ID: 2}, RemainingQty: 6, RemainingValue: 180, UnitCost: helper.Ptr(30.0)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				updated := []*CostLayer{}
				layers.UpdateFunc = func(_ context.Context, layer *CostLayer) (*CostLayer, error) {
					updated = append(updated, layer)
					return layer, nil
				}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Ship(ctx, 2, 50, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 1320 || got.Debit.Float64() != 196 {
					t.Errorf("line 0 = %+v, want COGS debit 196 (7 x avg 28)", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 196 {
					t.Errorf("line 1 = %+v, want inventory credit 196", got)
				}

				remaining := map[uint64]float64{}
				remainingValue := map[uint64]float64{}
				for _, layer := range updated {
					remaining[layer.ID] = layer.RemainingQty
					remainingValue[layer.ID] = layer.RemainingValue
				}
				if remaining[1] != 1.2 || remainingValue[1] != 30 {
					t.Errorf("layer 1 = qty %v value %v, want 1.2/30", remaining[1], remainingValue[1])
				}
				if remaining[2] != 1.8 || remainingValue[2] != 54 {
					t.Errorf("layer 2 = qty %v value %v, want 1.8/54", remaining[2], remainingValue[2])
				}
			},
		},
		{
			name: "average running cost on mixed receipts",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 2}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 8, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "average", Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 500, UnitCost: helper.Ptr(50.0)},
							{Base: model.Base{ID: 2}, RemainingQty: 10, RemainingValue: 600, UnitCost: helper.Ptr(60.0)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				updated := []*CostLayer{}
				layers.UpdateFunc = func(_ context.Context, layer *CostLayer) (*CostLayer, error) {
					updated = append(updated, layer)
					return layer, nil
				}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Ship(ctx, 2, 50, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if got := posted[0].Lines[0]; got.Debit.Float64() != 440 {
					t.Errorf("COGS debit = %v, want 440 (8 x avg 55)", got.Debit.Float64())
				}

				remaining := map[uint64]float64{}
				remainingValue := map[uint64]float64{}
				for _, layer := range updated {
					remaining[layer.ID] = layer.RemainingQty
					remainingValue[layer.ID] = layer.RemainingValue
				}
				if remaining[1] != 6 || remainingValue[1] != 300 {
					t.Errorf("layer 1 = qty %v value %v, want 6/300", remaining[1], remainingValue[1])
				}
				if remaining[2] != 6 || remainingValue[2] != 360 {
					t.Errorf("layer 2 = qty %v value %v, want 6/360", remaining[2], remainingValue[2])
				}
			},
		},
		{
			name: "rejects without stock",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 7, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 4, RemainingValue: 100, UnitCost: helper.Ptr(25.0)},
						}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Ship(ctx, 2, 50, time.Now())
				if helper.AssertError(t, err, true, ErrInsufficientStock) {
					return
				}
			},
		},
		{
			name: "rejects non-customer destination",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 20, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "internal"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Ship(ctx, 2, 50, time.Now())
				if helper.AssertError(t, err, true, ErrNotShipment) {
					return
				}
			},
		},
		{
			name: "full consumption zeroes value and COGS matches",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, Quantity: 10, RemainingQty: 10, RemainingValue: 33.3333, UnitCost: helper.Ptr(3.3333)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				updated := []*CostLayer{}
				layers.UpdateFunc = func(_ context.Context, layer *CostLayer) (*CostLayer, error) {
					updated = append(updated, layer)
					return layer, nil
				}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Ship(ctx, 2, 50, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				zeroed := 0
				for _, layer := range updated {
					if layer.Quantity > 0 && layer.RemainingQty == 0 {
						if layer.RemainingValue != 0 {
							t.Fatalf("fully consumed layer %d retains value %v, want 0", layer.ID, layer.RemainingValue)
						}
						zeroed++
					}
				}
				if zeroed != 1 {
					t.Fatalf("fully consumed layers = %d, want 1", zeroed)
				}
				if len(posted) != 1 || posted[0].Lines[0].Debit.Float64() != 33.3333 {
					t.Errorf("COGS = %v, want 33.3333 (no rounding drift)", posted[0].Lines[0].Debit.Float64())
				}
			},
		},
		{
			name: "standard cost",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 2}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 7, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "standard", StandardCost: 25, Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 4, RemainingValue: 100, UnitCost: helper.Ptr(25.0)},
							{Base: model.Base{ID: 2}, RemainingQty: 6, RemainingValue: 150, UnitCost: helper.Ptr(25.0)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				updated := []*CostLayer{}
				layers.UpdateFunc = func(_ context.Context, layer *CostLayer) (*CostLayer, error) {
					updated = append(updated, layer)
					return layer, nil
				}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Ship(ctx, 2, 50, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 1320 || got.Debit.Float64() != 175 {
					t.Errorf("line 0 = %+v, want COGS debit 175 (7 x std 25)", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 175 {
					t.Errorf("line 1 = %+v, want inventory credit 175", got)
				}

				remaining := map[uint64]float64{}
				remainingValue := map[uint64]float64{}
				for _, layer := range updated {
					remaining[layer.ID] = layer.RemainingQty
					remainingValue[layer.ID] = layer.RemainingValue
				}
				if remaining[1] != 0 || remainingValue[1] != 0 {
					t.Errorf("layer 1 = qty %v value %v, want 0/0", remaining[1], remainingValue[1])
				}
				if remaining[2] != 3 || remainingValue[2] != 75 {
					t.Errorf("layer 2 = qty %v value %v, want 3/75", remaining[2], remainingValue[2])
				}
			},
		},
		{
			name: "standard cost missing",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 2}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 7, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "standard", StandardCost: 0, Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 4, RemainingValue: 100, UnitCost: helper.Ptr(25.0)},
						}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Ship(ctx, 2, 50, time.Now())
				if helper.AssertError(t, err, true, ErrStandardCostMissing) {
					return
				}
			},
		},
		{
			name: "standard cost clamps decrement to layer value",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "standard", StandardCost: 25, Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 50, UnitCost: helper.Ptr(5.0)},
						}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Ship(ctx, 2, 50, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(posted) != 1 || posted[0].Lines[0].Debit.Float64() != 50 {
					t.Errorf("COGS = %v, want 50 clamped to remaining layer value", posted[0].Lines[0].Debit.Float64())
				}
			},
		},
		{
			name: "FIFO with nil unit cost",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 4, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, CogsAccountID: 1320}}, nil
					},
				}
				updated := []*CostLayer{}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 100, UnitCost: nil},
						}, nil
					},
				}
				layers.UpdateFunc = func(_ context.Context, layer *CostLayer) (*CostLayer, error) {
					updated = append(updated, layer)
					return layer, nil
				}
				poster := PosterMock{PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}
				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Ship(ctx, 2, 50, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if updated[0].RemainingQty != 6 || updated[0].RemainingValue != 100 {
					t.Errorf("layer after partial consume = qty %v value %v, want 6/100 (nil unit cost yields zero decrement)", updated[0].RemainingQty, updated[0].RemainingValue)
				}
			},
		},
		{
			name: "propagates apply error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
						return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
					}),
					ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *StockMovement) (*StockMovement, error) {
						return nil, ErrMovementNotFound
					},
				}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, CogsAccountID: 1320}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 100, UnitCost: helper.Ptr(20.0)}}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Ship(ctx, 2, 50, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "propagates layer list error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, CogsAccountID: 1320}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return nil, ErrNoValuation
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Ship(ctx, 2, 50, time.Now())
				if helper.AssertError(t, err, true, ErrNoValuation) {
					return
				}
			},
		},
		{
			name: "rejects missing batch",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "batch", StockAccounts: StockAccounts{StockValuationAccountID: 1300, CogsAccountID: 1320}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Ship(ctx, 2, 50, time.Now())
				if helper.AssertError(t, err, true, ErrBatchRequired) {
					return
				}
			},
		},
		{
			name: "rejects missing COGS account",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 2}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 40, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Ship(ctx, 2, 50, time.Now())
				if helper.AssertError(t, err, true, ErrCogsAccount) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_Scrap(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "consumes layers and posts expense",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 3}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 60, State: MovementStateConfirmed,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "scrap"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 125, UnitCost: helper.Ptr(25.0)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				if _, err := svc.Scrap(ctx, 3, 50, 6000, time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 6000 || got.Debit.Float64() != 125 {
					t.Errorf("line 0 = %+v, want expense debit 125 on account 6000", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 125 {
					t.Errorf("line 1 = %+v, want inventory credit 125 on account 1300", got)
				}
			},
		},
		{
			name: "rejects non-scrap destination",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 3}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 20, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "internal"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Scrap(ctx, 3, 50, 6000, time.Now())
				if helper.AssertError(t, err, true, ErrNotScrap) {
					return
				}
			},
		},
		{
			name: "rejects missing expense account",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 3}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 60, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "scrap"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300,
						}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Scrap(ctx, 3, 50, 0, time.Now())
				if helper.AssertError(t, err, true, ErrScrapAccount) {
					return
				}
			},
		},
		{
			name: "propagates apply error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
						return &StockMovement{Base: model.Base{ID: 3}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 60, State: MovementStateConfirmed}, nil
					}),
					ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *StockMovement) (*StockMovement, error) {
						return nil, ErrMovementNotFound
					},
				}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "scrap"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 100, UnitCost: helper.Ptr(20.0)}}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Scrap(ctx, 3, 50, 6000, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "propagates create layer error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 3}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 60, State: MovementStateConfirmed}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "scrap"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 5, RemainingValue: 100, UnitCost: helper.Ptr(20.0)}}, nil
					},
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *CostLayer) (*CostLayer, error) {
						return nil, ErrNoValuation
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Scrap(ctx, 3, 50, 6000, time.Now())
				if helper.AssertError(t, err, true, ErrNoValuation) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_Restock(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "posts inventory and reverses COGS",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 4}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 10, SrcLocationID: 40, DstLocationID: 10, State: MovementStateConfirmed,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, CogsAccountID: 1320,
						}}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}

				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, poster, TransactionerMock{})

				layer, err := svc.Restock(ctx, 4, amount.FromFloat64(25), 50, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if layer.Value != 250 || layer.RemainingQty != 10 {
					t.Errorf("layer = value %v remaining %v, want 250/10", layer.Value, layer.RemainingQty)
				}
				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 1300 || got.Debit.Float64() != 250 {
					t.Errorf("line 0 = %+v, want inventory debit 250 on account 1300", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1320 || got.Credit.Float64() != 250 {
					t.Errorf("line 1 = %+v, want COGS reversal credit 250 on account 1320", got)
				}
			},
		},
		{
			name: "rejects non-customer source",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 4}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 20, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "internal"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Restock(ctx, 4, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrNotReturn) {
					return
				}
			},
		},
		{
			name: "propagates apply error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
						return &StockMovement{Base: model.Base{ID: 4}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 40, DstLocationID: 10, State: MovementStateConfirmed}, nil
					}),
					ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *StockMovement) (*StockMovement, error) {
						return nil, ErrMovementNotFound
					},
				}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, CogsAccountID: 1320}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Restock(ctx, 4, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "rejects missing COGS account",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 4}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 40, DstLocationID: 10, State: MovementStateConfirmed}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "customer"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Restock(ctx, 4, amount.FromFloat64(25), 50, time.Now())
				if helper.AssertError(t, err, true, ErrCogsAccount) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_ReturnToSupplier(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "posts reversal and consumes layers",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 5}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 30, State: MovementStateConfirmed,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300, StockInputAccountID: 1310,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 250, UnitCost: helper.Ptr(25.0)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				layer, err := svc.ReturnToSupplier(ctx, 5, 50, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if layer.Value != -125 || layer.Quantity != -5 {
					t.Errorf("layer = value %v qty %v, want -125/-5", layer.Value, layer.Quantity)
				}
				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 1310 || got.Debit.Float64() != 125 {
					t.Errorf("line 0 = %+v, want stock input debit 125 on account 1310", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 125 {
					t.Errorf("line 1 = %+v, want inventory credit 125 on account 1300", got)
				}
			},
		},
		{
			name: "rejects non-supplier destination",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 5}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 20, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "internal"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.ReturnToSupplier(ctx, 5, 50, time.Now())
				if helper.AssertError(t, err, true, ErrNotSupplierReturn) {
					return
				}
			},
		},
		{
			name: "propagates apply error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
						return &StockMovement{Base: model.Base{ID: 5}, OrganizationID: &organizationID, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 30, State: MovementStateConfirmed}, nil
					}),
					ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *StockMovement) (*StockMovement, error) {
						return nil, ErrMovementNotFound
					},
				}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300, StockInputAccountID: 1310}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 250, UnitCost: helper.Ptr(25.0)}}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.ReturnToSupplier(ctx, 5, 50, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "rejects missing input account",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 5}, OrganizationID: &organizationID, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 30, State: MovementStateConfirmed}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "supplier"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 250, UnitCost: helper.Ptr(25.0)}}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.ReturnToSupplier(ctx, 5, 50, time.Now())
				if helper.AssertError(t, err, true, ErrInputAccount) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_Consume(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "posts WIP consumption",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 6}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 50, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300,
						}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 200, UnitCost: helper.Ptr(20.0)},
						}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}

				svc := NewValuationService(movements, layers, locations, resolver, poster, TransactionerMock{})

				layer, err := svc.Consume(ctx, 6, 50, 7000, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if layer.Value != -100 || layer.Quantity != -5 {
					t.Errorf("layer = value %v qty %v, want -100/-5", layer.Value, layer.Quantity)
				}
				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 7000 || got.Debit.Float64() != 100 {
					t.Errorf("line 0 = %+v, want WIP debit 100 on account 7000", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 100 {
					t.Errorf("line 1 = %+v, want inventory credit 100 on account 1300", got)
				}
			},
		},
		{
			name: "rejects non-production destination",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 6}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 20, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "internal"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Consume(ctx, 6, 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrNotConsumption) {
					return
				}
			},
		},
		{
			name: "rejects missing WIP account",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 6}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 50, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Consume(ctx, 6, 50, 0, time.Now())
				if helper.AssertError(t, err, true, ErrWIPAccount) {
					return
				}
			},
		},
		{
			name: "rejects insufficient stock",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 6}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 15, SrcLocationID: 10, DstLocationID: 50, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{
							{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 200, UnitCost: helper.Ptr(20.0)},
						}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Consume(ctx, 6, 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrInsufficientStock) {
					return
				}
			},
		},
		{
			name: "rejects missing organization",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 6}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 50, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Consume(ctx, 6, 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrOrganizationMissing) {
					return
				}
			},
		},
		{
			name: "rejects done movement",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 6}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 50, State: MovementStateDone}, nil
				})}
				svc := NewValuationService(movements, CostLayerDAOMock{}, StockLocationDAOMock{}, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Consume(ctx, 6, 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrMovementState) {
					return
				}
			},
		},
		{
			name: "propagates apply error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
						return &StockMovement{Base: model.Base{ID: 6}, OrganizationID: &organizationID, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 50, State: MovementStateDraft}, nil
					}),
					ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *StockMovement) (*StockMovement, error) {
						return nil, ErrMovementNotFound
					},
				}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 200, UnitCost: helper.Ptr(20.0)}}, nil
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Consume(ctx, 6, 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "propagates create layer error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 6}, OrganizationID: &organizationID, ItemID: 100, Qty: 5, SrcLocationID: 10, DstLocationID: 50, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				layers := CostLayerDAOMock{
					ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
						return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 200, UnitCost: helper.Ptr(20.0)}}, nil
					},
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *CostLayer) (*CostLayer, error) {
						return nil, ErrNoValuation
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Consume(ctx, 6, 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrNoValuation) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_Produce(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "posts finished goods",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movement := &StockMovement{
					Base: model.Base{ID: 7}, OrganizationID: &organizationID,
					ItemID: 100, Qty: 10, SrcLocationID: 50, DstLocationID: 10, State: MovementStateDraft,
				}
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return movement, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{
							StockValuationAccountID: 1300,
						}}, nil
					},
				}
				var posted []accounting.PostRequest
				poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
					posted = append(posted, request)
					return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
				}}

				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, poster, TransactionerMock{})

				layer, err := svc.Produce(ctx, 7, amount.FromFloat64(25), 50, 7000, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if layer.Value != 250 || layer.RemainingQty != 10 {
					t.Errorf("layer = value %v remaining %v, want 250/10", layer.Value, layer.RemainingQty)
				}
				if len(posted) != 1 {
					t.Fatalf("postings = %d, want 1", len(posted))
				}
				if got := posted[0].Lines[0]; got.AccountID != 1300 || got.Debit.Float64() != 250 {
					t.Errorf("line 0 = %+v, want inventory debit 250 on account 1300", got)
				}
				if got := posted[0].Lines[1]; got.AccountID != 7000 || got.Credit.Float64() != 250 {
					t.Errorf("line 1 = %+v, want WIP credit 250 on account 7000", got)
				}
			},
		},
		{
			name: "rejects non-production source",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 7}, ItemID: 100, Qty: 1, SrcLocationID: 10, DstLocationID: 20, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "internal"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Produce(ctx, 7, amount.FromFloat64(25), 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrNotProduction) {
					return
				}
			},
		},
		{
			name: "rejects negative cost",
			run: func(t *testing.T) {
				ctx := context.Background()
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 7}, ItemID: 100, Qty: 1, SrcLocationID: 50, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Produce(ctx, 7, amount.FromFloat64(-1), 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrNegativeCost) {
					return
				}
			},
		},
		{
			name: "rejects missing valuation account",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 7}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 50, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.Produce(ctx, 7, amount.FromFloat64(25), 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrValuationAccount) {
					return
				}
			},
		},
		{
			name: "missing standard cost rejected",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 7}, OrganizationID: &organizationID, ItemID: 100, Qty: 1, SrcLocationID: 50, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{CostMethod: "standard", Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Produce(ctx, 7, amount.FromFloat64(25), 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrStandardCostMissing) {
					return
				}
			},
		},
		{
			name: "propagates apply error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
						return &StockMovement{Base: model.Base{ID: 7}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 50, DstLocationID: 10, State: MovementStateDraft}, nil
					}),
					ApplyTxFunc: func(_ context.Context, _ *gorm.DB, _ *StockMovement) (*StockMovement, error) {
						return nil, ErrMovementNotFound
					},
				}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				svc := NewValuationService(movements, CostLayerDAOMock{}, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Produce(ctx, 7, amount.FromFloat64(25), 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrMovementNotFound) {
					return
				}
			},
		},
		{
			name: "propagates create layer error",
			run: func(t *testing.T) {
				ctx := context.Background()
				organizationID := uint64(1)
				movements := StockMovementDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockMovement, error) {
					return &StockMovement{Base: model.Base{ID: 7}, OrganizationID: &organizationID, ItemID: 100, Qty: 10, SrcLocationID: 50, DstLocationID: 10, State: MovementStateDraft}, nil
				})}
				locations := StockLocationDAOMock{
					CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
						return &reference.StockLocation{Usage: "production"}, nil
					}),
				}
				resolver := ItemResolverMock{
					ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
						return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
					},
				}
				layers := CostLayerDAOMock{
					CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *CostLayer) (*CostLayer, error) {
						return nil, ErrNoValuation
					},
				}
				svc := NewValuationService(movements, layers, locations, resolver, PosterMock{}, TransactionerMock{})

				_, err := svc.Produce(ctx, 7, amount.FromFloat64(25), 50, 7000, time.Now())
				if helper.AssertError(t, err, true, ErrNoValuation) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestValuationService_OnHandValue(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "returns rounded value",
			run: func(t *testing.T) {
				ctx := context.Background()
				layers := CostLayerDAOMock{
					ValueForItemFunc: func(_ context.Context, _ uint64) (float64, error) {
						return 123.45678, nil
					},
				}
				svc := NewValuationService(StockMovementDAOMock{}, layers, StockLocationDAOMock{}, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				got, err := svc.OnHandValue(ctx, 100)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got.Float64() != 123.4568 {
					t.Errorf("OnHandValue() = %v, want 123.4568", got.Float64())
				}
			},
		},
		{
			name: "propagates error",
			run: func(t *testing.T) {
				ctx := context.Background()
				layers := CostLayerDAOMock{
					ValueForItemFunc: func(_ context.Context, _ uint64) (float64, error) {
						return 0, ErrNoValuation
					},
				}
				svc := NewValuationService(StockMovementDAOMock{}, layers, StockLocationDAOMock{}, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

				_, err := svc.OnHandValue(ctx, 100)
				if helper.AssertError(t, err, true, ErrNoValuation) {
					return
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
