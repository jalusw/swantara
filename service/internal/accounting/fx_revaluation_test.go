package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

func TestFxRevaluationService_Revalue_SkipsAlreadyPosted(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	positions := FxRevaluationDAOMock{
		OpenForeignPositionsFunc: func(_ context.Context, _ uint64, _ string) ([]FxPosition, error) {
			return []FxPosition{
				{AccountID: 1100, AccountType: AccountTypeReceivable, CurrencyCode: "USD", ForeignBalance: amount.FromFloat64(100), BaseBalance: amount.FromFloat64(90)},
				{AccountID: 1100, AccountType: AccountTypeReceivable, CurrencyCode: "EUR", ForeignBalance: amount.FromFloat64(100), BaseBalance: amount.FromFloat64(90)},
			}, nil
		},
	}
	var postedRefs []string
	posterMoves := JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, _ []*JournalLine) (*JournalEntry, error) {
			if entry.Ref != nil {
				postedRefs = append(postedRefs, *entry.Ref)
			}
			return entry, nil
		},
	}
	movements := JournalEntryDAOMock{}
	movements.SearchFunc = func(_ context.Context, _ string, value any) (*JournalEntry, error) {
		if value == fxRevaluationRef(date, "USD") {
			return &JournalEntry{}, nil
		}
		return nil, nil
	}
	svc := NewFxRevaluationService(
		positions,
		NewPostingService(posterMoves),
		rateMock{rates: map[string]float64{"USD": 1, "EUR": 1}},
		fxMock{gain: 8000, loss: 8100},
		closingJournalMock{id: 20},
		func(_ context.Context, _ uint64) (string, error) { return "IDR", nil },
	).WithMoves(movements)

	posted, err := svc.Revalue(ctx, 10, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posted != 1 {
		t.Errorf("posted = %d, want 1", posted)
	}
	if len(postedRefs) != 1 || postedRefs[0] != fxRevaluationRef(date, "EUR") {
		t.Errorf("posted refs = %v, want [EUR ref]", postedRefs)
	}
}

func TestFxRevaluationService_ReverseRevaluation_ReversesPosted(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	ref := fxRevaluationRef(date, "USD")
	journalID := uint64(20)
	entry := &JournalEntry{Base: model.Base{ID: 5}, OrganizationID: 10, JournalID: journalID, Date: date, Ref: &ref, State: EntryStatePosted}
	movements := JournalEntryDAOMock{
		FindByReversedFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) {
			return nil, nil
		},
	}
	movements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[JournalEntry], error) {
		return &query.Page[JournalEntry]{Items: []*JournalEntry{entry}, Count: 1}, nil
	}
	var reversed []ReverseRequest
	poster := NewPostingService(JournalEntryDAOMock{}).SetReverser(PosterMock{
		ReverseFunc: func(_ context.Context, request ReverseRequest) (*JournalEntry, error) {
			reversed = append(reversed, request)
			return &JournalEntry{}, nil
		},
	})
	svc := NewFxRevaluationService(
		FxRevaluationDAOMock{},
		poster,
		rateMock{},
		fxMock{},
		closingJournalMock{id: 20},
		func(_ context.Context, _ uint64) (string, error) { return "IDR", nil },
	).WithMoves(movements)

	count, err := svc.ReverseRevaluation(ctx, 10, date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("reversed = %d, want 1", count)
	}
	if len(reversed) != 1 || reversed[0].EntryID != 5 || reversed[0].JournalID != journalID {
		t.Errorf("reverse request = %+v, want entry 5 journal 20", reversed)
	}
}
