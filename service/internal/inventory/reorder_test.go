package inventory

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func TestReorderService_Candidates_RecommendsReplenishment(t *testing.T) {
	ctx := context.Background()
	rules := ReorderRuleDAOMock{
		ListActiveFunc: func(_ context.Context) ([]*ReorderRule, error) {
			return []*ReorderRule{
				{Base: model.Base{ID: 1}, ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 10, MaxQty: 50, QtyMultiple: 10},
				{Base: model.Base{ID: 2}, ItemID: 101, LocationID: helper.Ptr(uint64(10)), MinQty: 5, MaxQty: 20, QtyMultiple: 5},
			}, nil
		},
	}
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, itemID, _ uint64, _ *uint64) (*StockBalance, error) {
			switch itemID {
			case 100:
				return &StockBalance{ItemID: 100, Quantity: 7}, nil
			default:
				return &StockBalance{ItemID: 101, Quantity: 30}, nil
			}
		},
	}
	svc := NewReorderService(rules, NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}, TransactionerMock{}), ItemResolverMock{})

	candidates, err := svc.Candidates(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("candidates = %d, want 1", len(candidates))
	}
	if candidates[0].ItemID != 100 || candidates[0].RecommendedQty != 50 {
		t.Errorf("candidate = %+v, want item 100 recommended 50 (43 rounded up to multiple of 10)", candidates[0])
	}
}

func TestReorderService_Create_RejectsMaxBelowMin(t *testing.T) {
	ctx := context.Background()
	svc := NewReorderService(ReorderRuleDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{})

	_, err := svc.Create(ctx, 1, &ReorderRule{ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 10, MaxQty: 5, QtyMultiple: 1})
	if helper.AssertError(t, err, true, ErrInvalidRuleQty) {
		return
	}
}

func TestReorderService_Create_RejectsMissingLocation(t *testing.T) {
	ctx := context.Background()
	svc := NewReorderService(ReorderRuleDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{})

	_, err := svc.Create(ctx, 1, &ReorderRule{ItemID: 100, MinQty: 0, MaxQty: 10, QtyMultiple: 1})
	if helper.AssertError(t, err, true, ErrLocationRequired) {
		return
	}
}

func TestReorderService_Update_Succeeds(t *testing.T) {
	ctx := context.Background()
	updated := false
	rules := ReorderRuleDAOMock{
		CRUDMock: dao.CRUDMock[ReorderRule]{
			UpdateFunc: func(_ context.Context, rule *ReorderRule) (*ReorderRule, error) {
				updated = true
				return rule, nil
			},
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", OrganizationID: helper.Ptr(uint64(1))}, nil
		},
	}
	svc := NewReorderService(rules, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver)

	rule := &ReorderRule{Base: model.Base{ID: 5}, ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 10, MaxQty: 50, QtyMultiple: 10}
	got, err := svc.Update(ctx, 1, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !updated || got.ID != 5 {
		t.Errorf("updated = %v, id = %d, want true/5", updated, got.ID)
	}
}

func TestReorderService_Update_RejectsInvalidQuantities(t *testing.T) {
	ctx := context.Background()
	svc := NewReorderService(ReorderRuleDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{})

	_, err := svc.Update(ctx, 1, &ReorderRule{ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 20, MaxQty: 10, QtyMultiple: 1})
	if helper.AssertError(t, err, true, ErrInvalidRuleQty) {
		return
	}
}

func TestReorderService_Update_RejectsZeroMultiple(t *testing.T) {
	ctx := context.Background()
	svc := NewReorderService(ReorderRuleDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{})

	_, err := svc.Update(ctx, 1, &ReorderRule{ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 0, MaxQty: 10, QtyMultiple: 0})
	if helper.AssertError(t, err, true, ErrInvalidRuleQty) {
		return
	}
}

func TestReorderService_Create_Succeeds(t *testing.T) {
	ctx := context.Background()
	created := false
	rules := ReorderRuleDAOMock{
		CRUDMock: dao.CRUDMock[ReorderRule]{
			CreateFunc: func(_ context.Context, rule *ReorderRule) (*ReorderRule, error) {
				created = true
				return rule, nil
			},
		},
	}
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", OrganizationID: helper.Ptr(uint64(1))}, nil
		},
	}
	svc := NewReorderService(rules, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver)

	rule := &ReorderRule{ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 10, MaxQty: 50, QtyMultiple: 10}
	got, err := svc.Create(ctx, 1, rule)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created || got.ItemID != 100 {
		t.Errorf("created = %v, item = %d, want true/100", created, got.ItemID)
	}
}

func TestReorderService_Update_PropagatesResolveError(t *testing.T) {
	ctx := context.Background()
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{}, ErrVariantNotFound
		},
	}
	svc := NewReorderService(ReorderRuleDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver)

	_, err := svc.Update(ctx, 1, &ReorderRule{ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 0, MaxQty: 10, QtyMultiple: 1})
	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestReorderService_Create_RejectsProductFromOtherOrganization(t *testing.T) {
	ctx := context.Background()
	resolver := ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (ResolvedItem, error) {
			return ResolvedItem{Tracking: "none", OrganizationID: helper.Ptr(uint64(2))}, nil
		},
	}
	svc := NewReorderService(ReorderRuleDAOMock{}, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), resolver)

	_, err := svc.Create(ctx, 1, &ReorderRule{ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 0, MaxQty: 10, QtyMultiple: 1})
	if helper.AssertError(t, err, true, ErrVariantNotFound) {
		return
	}
}

func TestReorderService_Candidates_SkipsUnqualifiedRules(t *testing.T) {
	ctx := context.Background()
	rules := ReorderRuleDAOMock{
		ListActiveFunc: func(_ context.Context) ([]*ReorderRule, error) {
			return []*ReorderRule{
				{Base: model.Base{ID: 1}, ItemID: 100, MinQty: 10, MaxQty: 50, QtyMultiple: 10},
				{Base: model.Base{ID: 2}, ItemID: 101, LocationID: helper.Ptr(uint64(10)), MinQty: 5, MaxQty: 20, QtyMultiple: 5},
				{Base: model.Base{ID: 3}, ItemID: 102, LocationID: helper.Ptr(uint64(10)), MinQty: 5, MaxQty: 20, QtyMultiple: 5},
			}, nil
		},
	}
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, itemID, _ uint64, _ *uint64) (*StockBalance, error) {
			switch itemID {
			case 101:
				return &StockBalance{ItemID: 101, Quantity: 30}, nil
			default:
				return &StockBalance{ItemID: 102, Quantity: 30}, nil
			}
		},
	}
	svc := NewReorderService(rules, NewLedgerService(StockMovementDAOMock{}, quants, StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}, TransactionerMock{}), ItemResolverMock{})

	candidates, err := svc.Candidates(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(candidates) != 0 {
		t.Errorf("candidates = %d, want 0 (missing location, above minimum, or zero recommendation)", len(candidates))
	}
}

func TestReorderService_Candidates_PropagatesLedgerError(t *testing.T) {
	ctx := context.Background()
	rules := ReorderRuleDAOMock{
		ListActiveFunc: func(_ context.Context) ([]*ReorderRule, error) {
			return []*ReorderRule{{Base: model.Base{ID: 1}, ItemID: 100, LocationID: helper.Ptr(uint64(10)), MinQty: 10, MaxQty: 50, QtyMultiple: 10}}, nil
		},
	}
	quants := StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*StockBalance, error) {
			return nil, ErrBalanceNotFound
		},
	}
	locations := StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 10}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}
	svc := NewReorderService(rules, NewLedgerService(StockMovementDAOMock{}, quants, locations, TransactionerMock{}), ItemResolverMock{})

	_, err := svc.Candidates(ctx, 1)
	if helper.AssertError(t, err, true, ErrBalanceNotFound) {
		return
	}
}

func TestReorderService_Candidates_PropagatesListError(t *testing.T) {
	ctx := context.Background()
	rules := ReorderRuleDAOMock{
		ListActiveFunc: func(_ context.Context) ([]*ReorderRule, error) {
			return nil, ErrRuleNotFound
		},
	}
	svc := NewReorderService(rules, NewLedgerService(StockMovementDAOMock{}, StockBalanceDAOMock{}, StockLocationDAOMock{}, TransactionerMock{}), ItemResolverMock{})

	_, err := svc.Candidates(ctx, 1)
	if helper.AssertError(t, err, true, ErrRuleNotFound) {
		return
	}
}

func TestReorderService_roundUpToMultiple_ReturnsZeroForNonPositiveValue(t *testing.T) {
	if got := roundUpToMultiple(0, 10); got != 0 {
		t.Errorf("roundUpToMultiple(0, 10) = %v, want 0", got)
	}
	if got := roundUpToMultiple(-5, 10); got != 0 {
		t.Errorf("roundUpToMultiple(-5, 10) = %v, want 0", got)
	}
}
