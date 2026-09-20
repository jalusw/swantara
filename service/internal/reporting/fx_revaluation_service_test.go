package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

func TestFxRevaluationService_Revalue_PostsBalancedUnrealizedEntry(t *testing.T) {
	ctx := context.Background()
	poster := &PosterMock{}
	var postedLines []accounting.PostingLine
	svc := NewFxRevaluationService(
		ReportDAOMock{OpenForeignPositionsFn: func(_ context.Context, _ uint64) ([]ForeignPositionRow, error) {
			return []ForeignPositionRow{
				{InvoiceID: 1, Type: "customer_invoice", CurrencyCode: "USD", AmountTotal: 1000, AmountResidual: 1000, AccountID: 2, AccountType: "receivable", BaseBalance: 15000},
			}, nil
		}},
		FxRevaluationDAOMock{},
		FxRevaluationLineDAOMock{},
		poster,
		ConfigSourceMock{},
		RateResolverMock{RateFn: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.FromFloat64(16), nil
		}},
		OrgReaderMock{},
	)
	poster.PostFn = func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
		postedLines = request.Lines
		move5 := &accounting.JournalEntry{}
		move5.ID = 5
		return move5, nil
	}

	result, err := svc.Revalue(ctx, 1, 7, time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("revalue failed: %v", err)
	}
	if result.Posted != 1 {
		t.Fatalf("posted = %d, want 1", result.Posted)
	}
	if len(postedLines) != 2 {
		t.Fatalf("movement lines = %d, want 2", len(postedLines))
	}
	if postedLines[0].Debit.Float64() != postedLines[1].Credit.Float64() {
		t.Errorf("debit %v != credit %v, unbalanced", postedLines[0].Debit, postedLines[1].Credit)
	}
	expected := amount.FromFloat64(1000*16 - 15000).Round(4).Float64()
	if result.TotalGainLoss != expected {
		t.Errorf("gain/loss = %v, want %v", result.TotalGainLoss, expected)
	}
}

func TestFxRevaluationService_Revalue_ReversesPriorUnrealized(t *testing.T) {
	ctx := context.Background()
	poster := &PosterMock{}
	svc := NewFxRevaluationService(
		ReportDAOMock{OpenForeignPositionsFn: func(_ context.Context, _ uint64) ([]ForeignPositionRow, error) {
			return nil, nil
		}},
		FxRevaluationDAOMock{},
		FxRevaluationLineDAOMock{
			ListOpenByOrgFn: func(_ context.Context, _ uint64) ([]*FxRevaluationLine, error) {
				moveID := uint64(41)
				return []*FxRevaluationLine{{RevaluationID: 3, EntryID: &moveID}}, nil
			},
			ListByRevaluationFn: func(_ context.Context, _ uint64) ([]*FxRevaluationLine, error) {
				return []*FxRevaluationLine{{RevaluationID: 3, Reversed: true}}, nil
			},
		},
		poster,
		ConfigSourceMock{},
		RateResolverMock{},
		OrgReaderMock{},
	)
	poster.ReverseFn = func(_ context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
		if request.EntryID != 41 {
			t.Errorf("reversed movement id = %d, want 41", request.EntryID)
		}
		move99 := &accounting.JournalEntry{}
		move99.ID = 99
		return move99, nil
	}

	result, err := svc.Revalue(ctx, 1, 7, time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("revalue failed: %v", err)
	}
	if result.Reversed != 1 {
		t.Errorf("reversed = %d, want 1", result.Reversed)
	}
	if poster.reverseCount != 1 {
		t.Errorf("reverse calls = %d, want 1", poster.reverseCount)
	}
}

func TestFxRevaluationService_Revalue_SkipsBaseCurrencyPositions(t *testing.T) {
	ctx := context.Background()
	poster := &PosterMock{}
	svc := NewFxRevaluationService(
		ReportDAOMock{OpenForeignPositionsFn: func(_ context.Context, _ uint64) ([]ForeignPositionRow, error) {
			return []ForeignPositionRow{
				{InvoiceID: 1, Type: "customer_invoice", CurrencyCode: "IDR", AmountTotal: 100, AmountResidual: 100, AccountID: 2, AccountType: "receivable", BaseBalance: 100},
			}, nil
		}},
		FxRevaluationDAOMock{},
		FxRevaluationLineDAOMock{},
		poster,
		ConfigSourceMock{},
		RateResolverMock{},
		OrgReaderMock{},
	)

	result, err := svc.Revalue(ctx, 1, 7, time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("revalue failed: %v", err)
	}
	if result.Posted != 0 || poster.postCount != 0 {
		t.Errorf("posted = %d calls = %d, want none for base currency", result.Posted, poster.postCount)
	}
}
