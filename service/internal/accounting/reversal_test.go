package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"gorm.io/gorm"
)

func TestReversalEngine_Reverse_MirrorsLines(t *testing.T) {
	ctx := context.Background()
	var capturedMovement *JournalEntry
	var capturedLines []*JournalLine

	original := &JournalEntry{
		Base:           model.Base{ID: 7},
		OrganizationID: 1,
		JournalID:      2,
		State:          EntryStatePosted,
	}
	movements := JournalEntryDAOMock{
		CRUDMock: dao.CRUDMock[JournalEntry]{
			FindFunc: func(_ context.Context, id uint64) (*JournalEntry, error) {
				if id == 7 {
					return original, nil
				}
				return nil, nil
			},
		},
		FindByReversedFunc: func(_ context.Context, moveID uint64) (*JournalEntry, error) {
			return nil, nil
		},
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			capturedMovement = entry
			capturedLines = lines
			entry.ID = 8
			return entry, nil
		},
	}
	lines := JournalLineDAOMock{
		ListByMovementFunc: func(_ context.Context, moveID uint64) ([]*JournalLine, error) {
			return []*JournalLine{
				{AccountID: 1300, Debit: amount.FromInt64(100), Credit: amount.FromInt64(0)},
				{AccountID: 1350, Debit: amount.FromInt64(0), Credit: amount.FromInt64(100)},
			}, nil
		},
	}
	svc := NewReversalEngine(movements, lines)

	date := time.Date(2026, 8, 2, 0, 0, 0, 0, time.UTC)
	reversal, err := svc.Reverse(ctx, ReverseRequest{
		OrganizationID: 1,
		JournalID:      3,
		Date:           date,
		Ref:            "REV/0001",
		Description:    "Reversal of REC/0001",
		EntryID:        7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reversal.ID != 8 {
		t.Errorf("reversal id = %d, want 8", reversal.ID)
	}
	if capturedMovement.ReversedEntryID == nil || *capturedMovement.ReversedEntryID != 7 {
		t.Errorf("reversed_entry_id = %v, want 7", capturedMovement.ReversedEntryID)
	}
	if capturedMovement.OriginID == nil || *capturedMovement.OriginID != 7 {
		t.Errorf("origin_id = %v, want 7", capturedMovement.OriginID)
	}
	if len(capturedLines) != 2 {
		t.Fatalf("lines = %d, want 2", len(capturedLines))
	}
	if !capturedLines[0].Debit.IsZero() || !capturedLines[0].Credit.Equal(amount.FromInt64(100)) {
		t.Errorf("line 0 = debit %v credit %v, want credit 100", capturedLines[0].Debit, capturedLines[0].Credit)
	}
	if !capturedLines[1].Debit.Equal(amount.FromInt64(100)) || !capturedLines[1].Credit.IsZero() {
		t.Errorf("line 1 = debit %v credit %v, want debit 100", capturedLines[1].Debit, capturedLines[1].Credit)
	}
}

func TestReversalEngine_Reverse_RejectsNotPosted(t *testing.T) {
	ctx := context.Background()
	movements := JournalEntryDAOMock{
		CRUDMock: dao.CRUDMock[JournalEntry]{
			FindFunc: func(_ context.Context, id uint64) (*JournalEntry, error) {
				return &JournalEntry{Base: model.Base{ID: 7}, OrganizationID: 1, State: EntryStateDraft}, nil
			},
		},
	}
	svc := NewReversalEngine(movements, JournalLineDAOMock{})

	_, err := svc.Reverse(ctx, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if helper.AssertError(t, err, true, ErrEntryNotPosted) {
		return
	}
}

func TestReversalEngine_Reverse_RejectsAlreadyReversed(t *testing.T) {
	ctx := context.Background()
	movements := JournalEntryDAOMock{
		CRUDMock: dao.CRUDMock[JournalEntry]{
			FindFunc: func(_ context.Context, id uint64) (*JournalEntry, error) {
				return &JournalEntry{Base: model.Base{ID: 7}, OrganizationID: 1, State: EntryStatePosted}, nil
			},
		},
		FindByReversedFunc: func(_ context.Context, moveID uint64) (*JournalEntry, error) {
			return &JournalEntry{Base: model.Base{ID: 8}}, nil
		},
	}
	svc := NewReversalEngine(movements, JournalLineDAOMock{})

	_, err := svc.Reverse(ctx, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if helper.AssertError(t, err, true, ErrEntryReversed) {
		return
	}
}

func TestReversalEngine_Reverse_RejectsMissingMove(t *testing.T) {
	ctx := context.Background()
	movements := JournalEntryDAOMock{
		CRUDMock: dao.CRUDMock[JournalEntry]{
			FindFunc: func(_ context.Context, id uint64) (*JournalEntry, error) {
				return nil, nil
			},
		},
	}
	svc := NewReversalEngine(movements, JournalLineDAOMock{})

	_, err := svc.Reverse(ctx, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if helper.AssertError(t, err, true, ErrEntryNotFound) {
		return
	}
}

func TestReversalEngine_Reverse_RejectsWrongOrganization(t *testing.T) {
	ctx := context.Background()
	movements := JournalEntryDAOMock{
		CRUDMock: dao.CRUDMock[JournalEntry]{
			FindFunc: func(_ context.Context, id uint64) (*JournalEntry, error) {
				return &JournalEntry{Base: model.Base{ID: 7}, OrganizationID: 2, State: EntryStatePosted}, nil
			},
		},
	}
	svc := NewReversalEngine(movements, JournalLineDAOMock{})

	_, err := svc.Reverse(ctx, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if helper.AssertError(t, err, true, ErrEntryNotFound) {
		return
	}
}

func TestReversalEngine_ReverseTx_DelegatesCreateWithLinesTx(t *testing.T) {
	ctx := context.Background()
	var capturedMovement *JournalEntry
	var capturedLines []*JournalLine

	original := &JournalEntry{
		Base:           model.Base{ID: 7},
		OrganizationID: 1,
		JournalID:      2,
		State:          EntryStatePosted,
	}
	movements := JournalEntryDAOMock{
		CRUDMock: dao.CRUDMock[JournalEntry]{
			FindFunc: func(_ context.Context, id uint64) (*JournalEntry, error) {
				return original, nil
			},
		},
		CreateWithLinesTxFunc: func(_ context.Context, tx *gorm.DB, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			capturedMovement = entry
			capturedLines = lines
			entry.ID = 8
			return entry, nil
		},
	}
	lines := JournalLineDAOMock{
		ListByMovementFunc: func(_ context.Context, moveID uint64) ([]*JournalLine, error) {
			return []*JournalLine{
				{AccountID: 1300, Debit: amount.FromInt64(100), Credit: amount.FromInt64(0)},
				{AccountID: 1350, Debit: amount.FromInt64(0), Credit: amount.FromInt64(100)},
			}, nil
		},
	}
	svc := NewReversalEngine(movements, lines)

	reversal, err := svc.ReverseTx(ctx, &gorm.DB{}, ReverseRequest{OrganizationID: 1, JournalID: 3, EntryID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reversal.ID != 8 {
		t.Errorf("reversal id = %d, want 8", reversal.ID)
	}
	if capturedMovement.ReversedEntryID == nil || *capturedMovement.ReversedEntryID != 7 {
		t.Errorf("reversed_entry_id = %v, want 7", capturedMovement.ReversedEntryID)
	}
	if len(capturedLines) != 2 || !capturedLines[0].Debit.IsZero() || !capturedLines[0].Credit.Equal(amount.FromInt64(100)) {
		t.Errorf("mirrored lines not swapped: %+v", capturedLines)
	}
}

func TestPostingService_Reverse_RequiresEngine(t *testing.T) {
	ctx := context.Background()
	svc := NewPostingService(JournalEntryDAOMock{})

	_, err := svc.Reverse(ctx, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if helper.AssertError(t, err, true, ErrReverserMissing) {
		return
	}
}

func TestPostingService_Reverse_DelegatesToEngine(t *testing.T) {
	ctx := context.Background()
	engine := ReversalEngineMock{Reversal: &JournalEntry{Base: model.Base{ID: 8}}}
	svc := NewPostingService(JournalEntryDAOMock{}).SetReverser(engine)

	reversal, err := svc.Reverse(ctx, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reversal.ID != 8 {
		t.Errorf("reversal id = %d, want 8", reversal.ID)
	}
}

func TestReversalEngine_ReverseTx_RejectsMissingMove(t *testing.T) {
	ctx := context.Background()
	movements := JournalEntryDAOMock{
		CRUDMock: dao.CRUDMock[JournalEntry]{
			FindFunc: func(_ context.Context, _ uint64) (*JournalEntry, error) { return nil, nil },
		},
	}
	svc := NewReversalEngine(movements, JournalLineDAOMock{})

	_, err := svc.ReverseTx(ctx, &gorm.DB{}, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if helper.AssertError(t, err, true, ErrEntryNotFound) {
		return
	}
}

func TestJournalEntry_IsPosted_ReturnsState(t *testing.T) {
	if !(JournalEntry{State: EntryStatePosted}).IsPosted() {
		t.Error("posted entry not recognized as posted")
	}
	if (JournalEntry{State: EntryStateDraft}).IsPosted() {
		t.Error("draft entry recognized as posted")
	}
	if (JournalEntry{}).IsPosted() {
		t.Error("empty entry recognized as posted")
	}
}

type ReversalEngineMock struct {
	Reversal *JournalEntry
}

func (m ReversalEngineMock) Reverse(ctx context.Context, request ReverseRequest) (*JournalEntry, error) {
	return m.Reversal, nil
}

func (m ReversalEngineMock) ReverseTx(ctx context.Context, tx *gorm.DB, request ReverseRequest) (*JournalEntry, error) {
	return m.Reversal, nil
}
