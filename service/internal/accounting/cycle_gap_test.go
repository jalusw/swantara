package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestPostingService_DraftConfirmCycle(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	movements := JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			entry.ID = 1
			return entry, nil
		},
	}
	svc := NewPostingService(movements)
	draft, err := svc.Post(ctx, PostRequest{
		OrganizationID: 10,
		JournalID:      20,
		Date:           date,
		Ref:            "MAN/0001",
		Draft:          true,
		Lines: []PostingLine{
			{AccountID: 1000, Name: "Cash", Debit: amount.FromFloat64(100)},
			{AccountID: 4000, Name: "Revenue", Credit: amount.FromFloat64(100)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if draft.State != EntryStateDraft {
		t.Fatalf("state = %s, want draft", draft.State)
	}
	if draft.PostedAt != nil {
		t.Fatalf("draft posted_at must be nil")
	}
	confirmMoves := JournalEntryDAOMock{}
	confirmMoves.FindFunc = func(_ context.Context, id uint64) (*JournalEntry, error) {
		return &JournalEntry{Base: model.Base{ID: id}, OrganizationID: 10, JournalID: 20, Date: date, State: EntryStateDraft}, nil
	}
	confirmMoves.UpdateFunc = func(_ context.Context, e *JournalEntry) (*JournalEntry, error) {
		return e, nil
	}
	confirmSvc := NewPostingService(confirmMoves)
	posted, err := confirmSvc.Confirm(ctx, 10, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posted.State != EntryStatePosted || posted.PostedAt == nil {
		t.Errorf("confirm did not post")
	}
	_, err = confirmSvc.Confirm(ctx, 99, 1)
	if err != ErrEntryNotFound {
		t.Errorf("err = %v, want ErrEntryNotFound", err)
	}
}

func TestPostingService_Post_PreservesFullFidelity(t *testing.T) {
	ctx := context.Background()
	var captured []*JournalLine
	movements := JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			captured = lines
			return entry, nil
		},
	}
	svc := NewPostingService(movements)
	due := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	cc := "USD"
	_, err := svc.Post(ctx, PostRequest{
		OrganizationID: 10,
		JournalID:      20,
		Date:           time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		CurrencyCode:   &cc,
		Lines: []PostingLine{
			{AccountID: 1100, Name: "AR", Debit: amount.FromFloat64(100), ContactID: helper.Ptr(uint64(7)), TaxID: helper.Ptr(uint64(9)), CurrencyCode: &cc, AmountCurrency: 1600000, DueDate: &due},
			{AccountID: 4000, Name: "Revenue", Credit: amount.FromFloat64(100), CurrencyCode: &cc, AmountCurrency: 1600000},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(captured) != 2 {
		t.Fatalf("lines = %d, want 2", len(captured))
	}
	if captured[0].ContactID == nil || *captured[0].ContactID != 7 {
		t.Errorf("contact not preserved")
	}
	if captured[0].TaxID == nil || *captured[0].TaxID != 9 {
		t.Errorf("tax not preserved")
	}
	if captured[0].CurrencyCode == nil || *captured[0].CurrencyCode != "USD" || captured[0].AmountCurrency != 1600000 {
		t.Errorf("currency not preserved")
	}
	if captured[0].DueDate == nil {
		t.Errorf("due date not preserved")
	}
}

func TestPostingService_Post_AutoReverse(t *testing.T) {
	ctx := context.Background()
	reversed := 0
	reverseDate := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	countingReverser := countingReverser{count: &reversed}
	svc := NewPostingService(JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			entry.ID = 3
			return entry, nil
		},
	}).SetReverser(countingReverser)
	_, err := svc.Post(ctx, PostRequest{
		OrganizationID:  10,
		JournalID:       20,
		Date:            time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Ref:             "ACR/1",
		Description:     "accrual",
		AutoReverseDate: &reverseDate,
		Lines: []PostingLine{
			{AccountID: 6000, Debit: amount.FromFloat64(50)},
			{AccountID: 2000, Credit: amount.FromFloat64(50)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reversed != 1 {
		t.Errorf("reversals = %d, want 1", reversed)
	}
	_, err = NewPostingService(JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			return entry, nil
		},
	}).Post(ctx, PostRequest{
		OrganizationID:  10,
		JournalID:       20,
		Date:            time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Ref:             "ACR/2",
		AutoReverseDate: &reverseDate,
		Lines: []PostingLine{
			{AccountID: 6000, Debit: amount.FromFloat64(50)},
			{AccountID: 2000, Credit: amount.FromFloat64(50)},
		},
	})
	if err != ErrReverserMissing {
		t.Errorf("err = %v, want ErrReverserMissing", err)
	}
}

type countingReverser struct {
	count *int
}

func (m countingReverser) Reverse(ctx context.Context, r ReverseRequest) (*JournalEntry, error) {
	*m.count++
	return &JournalEntry{}, nil
}

func (m countingReverser) ReverseTx(ctx context.Context, _ *gorm.DB, r ReverseRequest) (*JournalEntry, error) {
	return m.Reverse(ctx, r)
}

func TestPostingService_Post_AssignsSequence(t *testing.T) {
	ctx := context.Background()
	var captured *JournalEntry
	movements := JournalEntryDAOMock{
		CreateWithLinesFunc: func(_ context.Context, entry *JournalEntry, lines []*JournalLine) (*JournalEntry, error) {
			captured = entry
			return entry, nil
		},
	}
	seq := sequence.NewSequenceService(SequenceDAOMock{
		ReserveFunc: func(_ context.Context, _ uint64, _ string, _ time.Time) (*sequence.Reservation, error) {
			return &sequence.Reservation{Value: 7, Number: "MOVE/00007"}, nil
		},
	})
	svc := NewPostingService(movements).WithSequences(seq)
	_, err := svc.Post(ctx, PostRequest{
		OrganizationID: 10,
		JournalID:      20,
		Date:           time.Now(),
		Lines: []PostingLine{
			{AccountID: 1000, Debit: amount.FromFloat64(10)},
			{AccountID: 4000, Credit: amount.FromFloat64(10)},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if captured.Ref == nil || *captured.Ref != "MOVE/00007" {
		t.Errorf("ref not sequenced")
	}
}

func TestPeriodCloseService_ValidatesRetainedEarnings(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{
			FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
				return &TaxPeriod{Base: model.Base{ID: id}, OrganizationID: 10, DateStart: &start, DateEnd: &end, State: TaxPeriodStateOpen}, nil
			},
		},
	}
	tests := []struct {
		name    string
		account *reference.Account
		wantErr error
	}{
		{"zero id rejected", nil, ErrInvalidRetainedEarnings},
		{"non equity rejected", &reference.Account{OrganizationID: 10, Active: true, Type: AccountTypeExpense}, ErrInvalidRetainedEarnings},
		{"inactive rejected", &reference.Account{OrganizationID: 10, Active: false, Type: AccountTypeEquity}, ErrAccountInactive},
		{"foreign org rejected", &reference.Account{OrganizationID: 99, Active: true, Type: AccountTypeEquity}, ErrInvalidRetainedEarnings},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewPeriodCloseService(periods, PeriodAccountBalanceDAOMock{}, JournalEntryDAOMock{}, JournalLineDAOMock{}, PeriodCloseDAOMock{}, NewPostingService(JournalEntryDAOMock{}), TransactionerMock{})
			if tt.account != nil {
				id := uint64(900)
				acc := tt.account
				acc.ID = id
				svc = svc.WithAccounts(accountFinderStub{accounts: map[uint64]*reference.Account{id: acc}})
				_, err := svc.Close(ctx, 10, 1, id)
				if err != tt.wantErr {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			_, err := svc.Close(ctx, 10, 1, 0)
			if err != tt.wantErr {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestTaxPeriodService_CloseRequiresEntries(t *testing.T) {
	ctx := context.Background()
	period := &TaxPeriod{Base: model.Base{ID: 1}, OrganizationID: 10, State: TaxPeriodStateOpen}
	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{
			FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
				return period, nil
			},
			UpdateFunc: func(_ context.Context, p *TaxPeriod) (*TaxPeriod, error) {
				return p, nil
			},
		},
	}
	guard := PeriodCloseDAOMock{}
	svc := NewTaxPeriodService(periods, dao.CRUDMock[reference.TaxYear]{}).WithCloseGuard(guard)
	if _, err := svc.Close(ctx, 1); err != ErrCloseRequiresEntries {
		t.Errorf("err = %v, want ErrCloseRequiresEntries", err)
	}
}

type generalLedgerDAOMock struct {
	lines []GeneralLedgerLine
	err   error
}

func (m generalLedgerDAOMock) ListByAccount(_ context.Context, _, _ uint64, _, _ time.Time) ([]GeneralLedgerLine, error) {
	return m.lines, m.err
}

func TestGeneralLedgerService_RunningTotal(t *testing.T) {
	ctx := context.Background()
	svc := NewGeneralLedgerService(generalLedgerDAOMock{lines: []GeneralLedgerLine{
		{Debit: amount.FromFloat64(100)},
		{Credit: amount.FromFloat64(40)},
	}})
	lines, err := svc.ListByAccount(ctx, 10, 1000, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(lines))
	}
	if _, err := svc.ListByAccount(ctx, 10, 0, time.Now(), time.Now()); err != ErrInvalidLine {
		t.Errorf("err = %v, want ErrInvalidLine", err)
	}
}
