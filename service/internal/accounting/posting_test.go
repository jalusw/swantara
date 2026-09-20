package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestPostingService_Post_CreatesBalancedPostedMove(t *testing.T) {
	ctx := context.Background()
	var capturedMovement *JournalEntry
	var capturedLines []*JournalLine
	movements := JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			capturedMovement = entry
			capturedLines = lines
			entry.ID = 1
			return entry, nil
		},
	}
	svc := NewPostingService(movements)

	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	entry, err := svc.Post(ctx, PostRequest{
		OrganizationID: 10,
		JournalID:      20,
		Date:           date,
		Ref:            "REC/0001",
		OriginType:     OriginTypeStockMovement,
		OriginID:       99,
		Description:    "Goods receipt",
		Lines: []PostingLine{
			{AccountID: 1300, Name: "Inventory", Debit: amount.FromFloat64(1000)},
			{AccountID: 1350, Name: "Goods Received Not Billed", Credit: amount.FromFloat64(1000)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if entry.ID != 1 {
		t.Errorf("entry id = %d, want 1", entry.ID)
	}
	if capturedMovement.State != EntryStatePosted {
		t.Errorf("entry state = %s, want posted", capturedMovement.State)
	}
	if capturedMovement.OrganizationID != 10 || capturedMovement.JournalID != 20 {
		t.Errorf("entry scoping = %d/%d, want 10/20", capturedMovement.OrganizationID, capturedMovement.JournalID)
	}
	if len(capturedLines) != 2 {
		t.Fatalf("lines = %d, want 2", len(capturedLines))
	}
	if !capturedLines[0].Debit.Equal(amount.FromInt64(1000)) || !capturedLines[0].Credit.IsZero() {
		t.Errorf("line 0 = debit %v credit %v, want debit 1000", capturedLines[0].Debit, capturedLines[0].Credit)
	}
	if !capturedLines[1].Credit.Equal(amount.FromInt64(1000)) || !capturedLines[1].Debit.IsZero() {
		t.Errorf("line 1 = debit %v credit %v, want credit 1000", capturedLines[1].Debit, capturedLines[1].Credit)
	}
}

func TestPostingService_Post_RejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	svc := NewPostingService(JournalEntryDAOMock{})

	tests := []struct {
		name    string
		request PostRequest
		wantErr error
	}{
		{
			name: "unbalanced entry",
			request: PostRequest{
				OrganizationID: 10,
				JournalID:      20,
				Date:           time.Now(),
				Lines: []PostingLine{
					{AccountID: 1300, Name: "Inventory", Debit: amount.FromFloat64(100)},
					{AccountID: 1350, Name: "Goods Received Not Billed", Credit: amount.FromFloat64(99.99)},
				},
			},
			wantErr: ErrUnbalanced,
		},
		{
			name:    "no lines",
			request: PostRequest{OrganizationID: 10, JournalID: 20, Date: time.Now()},
			wantErr: ErrNoLines,
		},
		{
			name: "double sided line",
			request: PostRequest{
				OrganizationID: 10,
				JournalID:      20,
				Date:           time.Now(),
				Lines: []PostingLine{
					{AccountID: 1300, Name: "Inventory", Debit: amount.FromFloat64(100), Credit: amount.FromFloat64(100)},
				},
			},
			wantErr: ErrInvalidLine,
		},
		{
			name: "negative amount",
			request: PostRequest{
				OrganizationID: 10,
				JournalID:      20,
				Date:           time.Now(),
				Lines: []PostingLine{
					{AccountID: 1300, Name: "Inventory", Debit: amount.FromFloat64(-100)},
				},
			},
			wantErr: ErrInvalidLine,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Post(ctx, tt.request)

			helper.AssertError(t, err, true, tt.wantErr)
		})
	}
}

func TestPostingService_PostTx_CreatesBalancedPostedMove(t *testing.T) {
	ctx := context.Background()
	var capturedMovement *JournalEntry
	var capturedLines []*JournalLine
	movements := JournalEntryDAOMock{
		CreateWithLinesTxFunc: func(_ context.Context, tx *gorm.DB, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			capturedMovement = entry
			capturedLines = lines
			entry.ID = 5
			return entry, nil
		},
	}
	svc := NewPostingService(movements)

	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	entry, err := svc.PostTx(ctx, &gorm.DB{}, PostRequest{
		OrganizationID: 10,
		JournalID:      20,
		Date:           date,
		Description:    "Goods receipt tx",
		Lines: []PostingLine{
			{AccountID: 1300, Name: "Inventory", Debit: amount.FromFloat64(1000)},
			{AccountID: 1350, Name: "Goods Received Not Billed", Credit: amount.FromFloat64(1000)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.ID != 5 {
		t.Errorf("entry id = %d, want 5", entry.ID)
	}
	if capturedMovement.State != EntryStatePosted || len(capturedLines) != 2 {
		t.Errorf("entry state = %s lines = %d, want posted/2", capturedMovement.State, len(capturedLines))
	}
}

func TestPostingService_ReverseTx_RequiresEngine(t *testing.T) {
	ctx := context.Background()
	svc := NewPostingService(JournalEntryDAOMock{})

	_, err := svc.ReverseTx(ctx, &gorm.DB{}, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if helper.AssertError(t, err, true, ErrReverserMissing) {
		return
	}
}

func TestPostingService_ReverseTx_DelegatesToEngine(t *testing.T) {
	ctx := context.Background()
	engine := ReversalEngineMock{Reversal: &JournalEntry{Base: model.Base{ID: 8}}}
	svc := NewPostingService(JournalEntryDAOMock{}).SetReverser(engine)

	reversal, err := svc.ReverseTx(ctx, &gorm.DB{}, ReverseRequest{OrganizationID: 1, EntryID: 7})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reversal.ID != 8 {
		t.Errorf("reversal id = %d, want 8", reversal.ID)
	}
}

func TestPostingService_PostTx_RejectsClosedPeriod(t *testing.T) {
	ctx := context.Background()
	svc := NewPostingService(JournalEntryDAOMock{}).SetPeriods(PeriodLookupMock{AssertOpenErr: ErrPeriodLocked})

	_, err := svc.PostTx(ctx, &gorm.DB{}, PostRequest{
		OrganizationID: 1,
		Date:           time.Now(),
		Lines: []PostingLine{
			{AccountID: 1300, Debit: amount.FromFloat64(10)},
			{AccountID: 1350, Credit: amount.FromFloat64(10)},
		},
	})
	if helper.AssertError(t, err, true, ErrPeriodLocked) {
		return
	}
}

func TestPostingService_PostTx_RejectsOriginIncomplete(t *testing.T) {
	ctx := context.Background()
	svc := NewPostingService(JournalEntryDAOMock{})

	_, err := svc.PostTx(ctx, &gorm.DB{}, PostRequest{
		OrganizationID: 1,
		JournalID:      20,
		Date:           time.Now(),
		OriginID:       5,
		Lines: []PostingLine{
			{AccountID: 1300, Name: "A", Debit: amount.FromFloat64(10)},
			{AccountID: 1350, Name: "B", Credit: amount.FromFloat64(10)},
		},
	})
	if helper.AssertError(t, err, true, ErrOriginIncomplete) {
		return
	}
}

func TestPostingService_Post_RejectsOriginIncomplete(t *testing.T) {
	ctx := context.Background()
	svc := NewPostingService(JournalEntryDAOMock{})

	_, err := svc.Post(ctx, PostRequest{
		OrganizationID: 1,
		JournalID:      20,
		Date:           time.Now(),
		OriginID:       5,
		Lines: []PostingLine{
			{AccountID: 1300, Name: "A", Debit: amount.FromFloat64(10)},
			{AccountID: 1350, Name: "B", Credit: amount.FromFloat64(10)},
		},
	})
	if helper.AssertError(t, err, true, ErrOriginIncomplete) {
		return
	}
}

type accountFinderStub struct {
	accounts map[uint64]*reference.Account
}

func (s accountFinderStub) Find(_ context.Context, id uint64) (*reference.Account, error) {
	return s.accounts[id], nil
}

func TestPostingService_Post_ValidatesAccounts(t *testing.T) {
	ctx := context.Background()
	active := &reference.Account{OrganizationID: 10, Active: true}
	active.ID = 1300
	inactive := &reference.Account{OrganizationID: 10, Active: false}
	foreign := &reference.Account{OrganizationID: 99, Active: true}

	tests := []struct {
		name     string
		accounts map[uint64]*reference.Account
		wantErr  error
	}{
		{"valid accounts post", map[uint64]*reference.Account{1300: active, 1350: active}, nil},
		{"missing account rejected", map[uint64]*reference.Account{1300: active}, ErrAccountNotFound},
		{"inactive account rejected", map[uint64]*reference.Account{1300: active, 1350: inactive}, ErrAccountInactive},
		{"foreign org account rejected", map[uint64]*reference.Account{1300: active, 1350: foreign}, ErrAccountNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewPostingService(JournalEntryDAOMock{
				CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
					return entry, nil
				},
			}).WithAccounts(accountFinderStub{accounts: tt.accounts})
			_, err := svc.Post(ctx, PostRequest{
				OrganizationID: 10,
				JournalID:      20,
				Date:           time.Now(),
				Ref:            "REC/0001",
				Lines: []PostingLine{
					{AccountID: 1300, Name: "A", Debit: amount.FromFloat64(10)},
					{AccountID: 1350, Name: "B", Credit: amount.FromFloat64(10)},
				},
			})
			if tt.wantErr == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != nil && helper.AssertError(t, err, true, tt.wantErr) {
				return
			}
		})
	}
}
