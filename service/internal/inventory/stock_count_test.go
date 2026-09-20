package inventory

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestStockCountService_Create_ComputesTheoreticalQty(t *testing.T) {
	ctx := context.Background()
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return &StockBalance{ItemID: 100, Quantity: 15}, nil
		},
	}
	ledger := NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}}, nil
			},
		},
	}, TransactionerMock{})
	var captured []*StockCountLine
	counts := StockCountDAOMock{
		CreateWithLinesFunc: func(_ context.Context, count *StockCount, lines []*StockCountLine) (*StockCount, error) {
			captured = lines
			return count, nil
		},
	}
	svc := NewStockCountService(counts, StockCountLineDAOMock{}, quants, CostLayerDAOMock{}, ledger, ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	count, err := svc.Create(ctx, &StockCount{LocationID: helper.Ptr(uint64(10))}, []StockCountLineRequest{
		{ItemID: 100, CountedQty: 17},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.State != CountStateDraft {
		t.Errorf("state = %s, want draft", count.State)
	}
	if len(captured) != 1 {
		t.Fatalf("lines = %d, want 1", len(captured))
	}
	if captured[0].TheoreticalQty != 15 || captured[0].DiffQty != 2 {
		t.Errorf("line = theoretical %v diff %v, want 15/2", captured[0].TheoreticalQty, captured[0].DiffQty)
	}
}

func TestStockCountService_Create_RejectsMissingLocation(t *testing.T) {
	ctx := context.Background()
	svc := NewStockCountService(StockCountDAOMock{}, StockCountLineDAOMock{}, StockBalanceDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Create(ctx, &StockCount{}, nil)
	if helper.AssertError(t, err, true, ErrLocationRequired) {
		return
	}
}

func TestStockCountService_Post_PostsBalancedAdjustment(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	count := &StockCount{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		LocationID: helper.Ptr(uint64(10)), State: CountStateDraft,
	}
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return count, nil
	})}
	lines := StockCountLineDAOMock{
		ListByCountFunc: func(_ context.Context, _ uint64) ([]*StockCountLine, error) {
			return []*StockCountLine{
				{Base: model.Base{ID: 1}, StockCountID: 1, ItemID: 100, TheoreticalQty: 10, CountedQty: 13, DiffQty: 3},
			}, nil
		},
	}
	var upserted *StockBalance
	quants := StockBalanceDAOMock{
		UpsertFunc: func(_ context.Context, _ *uint64, _, _ uint64, _ *uint64, _ float64) (*StockBalance, error) {
			upserted = &StockBalance{ItemID: 100, Quantity: 3}
			return upserted, nil
		},
	}
	layers := CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
			return []*CostLayer{
				{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 250, UnitCost: helper.Ptr(25.0)},
			}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
		posted = append(posted, request)
		return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
	}}
	ledger := NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{}, TransactionerMock{})
	svc := NewStockCountService(counts, lines, quants, layers, ledger, resolver, poster, TransactionerMock{})

	updated, err := svc.Post(ctx, 1, 50, 1400, time.Date(2026, 8, 3, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.State != CountStatePosted {
		t.Errorf("state = %s, want posted", updated.State)
	}
	if upserted.Quantity != 3 {
		t.Errorf("quant delta = %v, want 3", upserted.Quantity)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if got := posted[0].Lines[0]; got.AccountID != 1300 || got.Debit.Float64() != 75 {
		t.Errorf("line 0 = %+v, want inventory debit 75 (3 x 25)", got)
	}
	if got := posted[0].Lines[1]; got.AccountID != 1400 || got.Credit.Float64() != 75 {
		t.Errorf("line 1 = %+v, want gain/loss credit 75", got)
	}
}

func TestStockCountService_Post_RejectsWhenPosted(t *testing.T) {
	ctx := context.Background()
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return &StockCount{Base: model.Base{ID: 1}, State: CountStatePosted}, nil
	})}
	svc := NewStockCountService(counts, StockCountLineDAOMock{}, StockBalanceDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Post(ctx, 1, 50, 1400, time.Now())
	if helper.AssertError(t, err, true, ErrStockCountState) {
		return
	}
}

func TestStockCountService_Post_PostsInventoryLoss(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	count := &StockCount{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		LocationID: helper.Ptr(uint64(10)), State: CountStateDraft,
	}
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return count, nil
	})}
	lines := StockCountLineDAOMock{
		ListByCountFunc: func(_ context.Context, _ uint64) ([]*StockCountLine, error) {
			return []*StockCountLine{
				{Base: model.Base{ID: 1}, StockCountID: 1, ItemID: 100, DiffQty: -3},
			}, nil
		},
	}
	quants := StockBalanceDAOMock{
		UpsertFunc: func(_ context.Context, _ *uint64, _, _ uint64, _ *uint64, delta float64) (*StockBalance, error) {
			return &StockBalance{ItemID: 100, Quantity: delta}, nil
		},
	}
	layers := CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
			return []*CostLayer{{Base: model.Base{ID: 1}, RemainingQty: 10, RemainingValue: 250, UnitCost: helper.Ptr(25.0)}}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	var posted []accounting.PostRequest
	poster := PosterMock{PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
		posted = append(posted, request)
		return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
	}}
	svc := NewStockCountService(counts, lines, quants, layers, NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{}, TransactionerMock{}), resolver, poster, TransactionerMock{})

	if _, err := svc.Post(ctx, 1, 50, 1400, time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posted) != 1 {
		t.Fatalf("postings = %d, want 1", len(posted))
	}
	if got := posted[0].Lines[0]; got.AccountID != 1400 || got.Debit.Float64() != 75 {
		t.Errorf("line 0 = %+v, want gain/loss debit 75 on account 1400", got)
	}
	if got := posted[0].Lines[1]; got.AccountID != 1300 || got.Credit.Float64() != 75 {
		t.Errorf("line 1 = %+v, want inventory credit 75 on account 1300", got)
	}
}

func TestStockCountService_Post_SkipsZeroDiffAndZeroValueLines(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	count := &StockCount{
		Base: model.Base{ID: 1}, OrganizationID: &organizationID,
		LocationID: helper.Ptr(uint64(10)), State: CountStateDraft,
	}
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return count, nil
	})}
	lines := StockCountLineDAOMock{
		ListByCountFunc: func(_ context.Context, _ uint64) ([]*StockCountLine, error) {
			return []*StockCountLine{
				{Base: model.Base{ID: 1}, StockCountID: 1, ItemID: 100, DiffQty: 0},
				{Base: model.Base{ID: 2}, StockCountID: 1, ItemID: 101, DiffQty: 2},
			}, nil
		},
	}
	quants := StockBalanceDAOMock{
		UpsertFunc: func(_ context.Context, _ *uint64, _, _ uint64, _ *uint64, delta float64) (*StockBalance, error) {
			return &StockBalance{ItemID: 101, Quantity: delta}, nil
		},
	}
	layers := CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*CostLayer, error) {
			return []*CostLayer{}, nil
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", StockAccounts: StockAccounts{StockValuationAccountID: 1300}}, nil
		},
	}
	posted := false
	poster := PosterMock{PostFunc: func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
		posted = true
		return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
	}}
	svc := NewStockCountService(counts, lines, quants, layers, NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{}, TransactionerMock{}), resolver, poster, TransactionerMock{})

	if _, err := svc.Post(ctx, 1, 50, 1400, time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posted {
		t.Error("no posting expected for a zero diff and a zero value line")
	}
}

func TestStockCountService_Post_RejectsMissingGainLossAccount(t *testing.T) {
	ctx := context.Background()
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return &StockCount{Base: model.Base{ID: 1}, State: CountStateDraft}, nil
	})}
	svc := NewStockCountService(counts, StockCountLineDAOMock{}, StockBalanceDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Post(ctx, 1, 50, 0, time.Now())
	if helper.AssertError(t, err, true, ErrGainLossAccount) {
		return
	}
}

func TestStockCountService_Post_RejectsMissingOrganization(t *testing.T) {
	ctx := context.Background()
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return &StockCount{Base: model.Base{ID: 1}, State: CountStateDraft}, nil
	})}
	svc := NewStockCountService(counts, StockCountLineDAOMock{}, StockBalanceDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Post(ctx, 1, 50, 1400, time.Now())
	if helper.AssertError(t, err, true, ErrOrganizationMissing) {
		return
	}
}

func TestStockCountService_Post_RejectsWithoutLines(t *testing.T) {
	ctx := context.Background()
	organizationID := uint64(1)
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return &StockCount{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: CountStateDraft}, nil
	})}
	svc := NewStockCountService(counts, StockCountLineDAOMock{}, StockBalanceDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Post(ctx, 1, 50, 1400, time.Now())
	if helper.AssertError(t, err, true, ErrStockCountNoLines) {
		return
	}
}

func TestStockCountService_Post_RejectsMissingCount(t *testing.T) {
	ctx := context.Background()
	counts := StockCountDAOMock{CRUDMock: daoCRUDFind(func(_ context.Context, _ uint64) (*StockCount, error) {
		return nil, nil
	})}
	svc := NewStockCountService(counts, StockCountLineDAOMock{}, StockBalanceDAOMock{}, CostLayerDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{}, PosterMock{}, TransactionerMock{})

	_, err := svc.Post(ctx, 1, 50, 1400, time.Now())
	if helper.AssertError(t, err, true, ErrStockCountNotFound) {
		return
	}
}
