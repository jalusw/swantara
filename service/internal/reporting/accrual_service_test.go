package reporting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
)

func TestAccrualService_Create_PostsBalancedAccrual(t *testing.T) {
	ctx := context.Background()
	poster := &PosterMock{}
	var postedLines []accounting.PostingLine
	svc := NewAccrualService(AccrualDAOMock{}, AccrualLineDAOMock{}, poster, ConfigSourceMock{})
	poster.PostFn = func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
		postedLines = request.Lines
		movement := &accounting.JournalEntry{}
		movement.ID = 7
		return movement, nil
	}

	reversalDate := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	result, err := svc.Create(ctx, AccrualRequest{
		OrganizationID: 1,
		PeriodID:       7,
		Name:           "Utilities Accrual",
		Description:    "Accrued utilities for August",
		Date:           time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
		ReversalDate:   &reversalDate,
		Lines: []AccrualLineRequest{
			{AccountID: 5, Name: "Expense", Debit: 1000},
			{AccountID: 6, Name: "Accrued Expense", Credit: 1000},
		},
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if result.EntryID == nil || *result.EntryID != 7 || result.State != AccrualStatePosted {
		t.Errorf("result = %+v, want posted movement 7", result)
	}
	if len(postedLines) != 2 || postedLines[0].Debit.Float64() != 1000 {
		t.Errorf("lines = %+v, want balanced 1000", postedLines)
	}
}

func TestAccrualService_Create_RejectsUnbalancedLines(t *testing.T) {
	ctx := context.Background()
	svc := NewAccrualService(AccrualDAOMock{}, AccrualLineDAOMock{}, &PosterMock{}, ConfigSourceMock{})

	_, err := svc.Create(ctx, AccrualRequest{
		OrganizationID: 1,
		PeriodID:       7,
		Name:           "Bad Accrual",
		Lines: []AccrualLineRequest{
			{AccountID: 5, Name: "Expense", Debit: 1000},
			{AccountID: 6, Name: "Accrued Expense", Credit: 900},
		},
	})
	if err != ErrUnbalancedAccrual {
		t.Errorf("err = %v, want ErrUnbalancedAccrual", err)
	}
}

func TestAccrualService_ReverseDue_ReversesOnlyDueAccruals(t *testing.T) {
	ctx := context.Background()
	poster := &PosterMock{}
	dueMove := uint64(7)
	notDueMove := uint64(8)
	dueDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	futureDate := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	var reversedMoves []uint64
	var updated []*Accrual
	svc := NewAccrualService(AccrualDAOMock{
		ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*Accrual, error) {
			due := &Accrual{State: AccrualStatePosted, EntryID: &dueMove, ReversalDate: &dueDate}
			due.ID = 1
			notDue := &Accrual{State: AccrualStatePosted, EntryID: &notDueMove, ReversalDate: &futureDate}
			notDue.ID = 2
			return []*Accrual{due, notDue}, nil
		},
		UpdateFn: func(_ context.Context, entity *Accrual) (*Accrual, error) {
			updated = append(updated, entity)
			return entity, nil
		},
	}, AccrualLineDAOMock{}, poster, ConfigSourceMock{})
	poster.ReverseFn = func(_ context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
		reversedMoves = append(reversedMoves, request.EntryID)
		movement := &accounting.JournalEntry{}
		movement.ID = 100 + request.EntryID
		return movement, nil
	}

	asOf := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	reversed, err := svc.ReverseDue(ctx, 1, asOf)
	if err != nil {
		t.Fatalf("reverse due failed: %v", err)
	}
	if reversed != 1 {
		t.Errorf("reversed = %d, want 1", reversed)
	}
	if len(reversedMoves) != 1 || reversedMoves[0] != dueMove {
		t.Errorf("reversed movements = %v, want [7]", reversedMoves)
	}
	if len(updated) != 1 || updated[0].State != AccrualStateReversed {
		t.Errorf("updated = %+v, want one reversed accrual", updated)
	}
}
