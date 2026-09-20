package accounting

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type closingJournalMock struct {
	id  uint64
	err error
}

func (m closingJournalMock) ClosingJournalID(_ context.Context, _ uint64) (uint64, error) {
	return m.id, m.err
}

func TestPeriodCloseService_ClosesPeriodWithNetIncome(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{
			FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
				return &TaxPeriod{
					Base:           model.Base{ID: id},
					OrganizationID: 10,
					DateStart:      &start,
					DateEnd:        &end,
					State:          TaxPeriodStateOpen,
				}, nil
			},
			UpdateFunc: func(_ context.Context, p *TaxPeriod) (*TaxPeriod, error) {
				return p, nil
			},
		},
	}
	balances := PeriodAccountBalanceDAOMock{
		SumByAccountAndPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]PeriodAccountBalance, error) {
			return []PeriodAccountBalance{
				{AccountID: 100, AccountCode: "4100", AccountName: "Revenue", AccountType: "income", TotalDebit: 0, TotalCredit: 5000},
				{AccountID: 200, AccountCode: "5100", AccountName: "COGS", AccountType: "cogs", TotalDebit: 2000, TotalCredit: 0},
				{AccountID: 300, AccountCode: "6100", AccountName: "Rent", AccountType: "expense", TotalDebit: 500, TotalCredit: 0},
			}, nil
		},
	}
	movements := JournalEntryDAOMock{}
	lines := JournalLineDAOMock{}
	periodCloses := PeriodCloseDAOMock{
		CreateFunc: func(_ context.Context, e *PeriodCloseEntry) (*PeriodCloseEntry, error) {
			e.ID = 1
			return e, nil
		},
	}

	var capturedRequest PostRequest
	poster := JournalEntryDAOMock{}
	baseSvc := NewPeriodCloseService(periods, balances, movements, lines, periodCloses, NewPostingService(poster))
	svc := baseSvc.WithJournalResolver(closingJournalMock{id: 1})

	_, err := svc.Close(ctx, 10, 1, 900)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = capturedRequest
}

func TestPeriodCloseService_RejectsWrongOrg(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{
				Base:           model.Base{ID: id},
				OrganizationID: 10,
				DateStart:      &start,
				DateEnd:        &end,
				State:          TaxPeriodStateOpen,
			}, nil
		}},
	}
	svc := NewPeriodCloseService(periods, PeriodAccountBalanceDAOMock{}, JournalEntryDAOMock{}, JournalLineDAOMock{}, PeriodCloseDAOMock{}, NewPostingService(JournalEntryDAOMock{}))

	_, err := svc.Close(ctx, 99, 1, 900)
	if err != ErrPeriodNotFound {
		t.Errorf("err = %v, want ErrPeriodNotFound", err)
	}
}

func TestPeriodCloseService_RejectsClosedPeriod(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{
				Base:           model.Base{ID: id},
				OrganizationID: 10,
				DateStart:      &start,
				DateEnd:        &end,
				State:          TaxPeriodStateClosed,
			}, nil
		}},
	}
	svc := NewPeriodCloseService(periods, PeriodAccountBalanceDAOMock{}, JournalEntryDAOMock{}, JournalLineDAOMock{}, PeriodCloseDAOMock{}, NewPostingService(JournalEntryDAOMock{}))

	_, err := svc.Close(ctx, 10, 1, 900)
	if err != ErrPeriodNotOpen {
		t.Errorf("err = %v, want ErrPeriodNotOpen", err)
	}
}

func TestPeriodCloseService_RejectsAlreadyClosed(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
			return &TaxPeriod{
				Base:           model.Base{ID: id},
				OrganizationID: 10,
				DateStart:      &start,
				DateEnd:        &end,
				State:          TaxPeriodStateOpen,
			}, nil
		}},
	}
	periodCloses := PeriodCloseDAOMock{
		FindByPeriodFunc: func(_ context.Context, _ uint64) (*PeriodCloseEntry, error) {
			return &PeriodCloseEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	svc := NewPeriodCloseService(periods, PeriodAccountBalanceDAOMock{}, JournalEntryDAOMock{}, JournalLineDAOMock{}, periodCloses, NewPostingService(JournalEntryDAOMock{}))

	_, err := svc.Close(ctx, 10, 1, 900)
	if err != ErrPeriodAlreadyClosed {
		t.Errorf("err = %v, want ErrPeriodAlreadyClosed", err)
	}
}

func TestPeriodCloseService_ClosesEmptyPeriod(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	var updatedState string
	periods := TaxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[TaxPeriod]{
			FindFunc: func(_ context.Context, id uint64) (*TaxPeriod, error) {
				return &TaxPeriod{
					Base:           model.Base{ID: id},
					OrganizationID: 10,
					DateStart:      &start,
					DateEnd:        &end,
					State:          TaxPeriodStateOpen,
				}, nil
			},
			UpdateFunc: func(_ context.Context, p *TaxPeriod) (*TaxPeriod, error) {
				updatedState = p.State
				return p, nil
			},
		},
	}
	balances := PeriodAccountBalanceDAOMock{
		SumByAccountAndPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]PeriodAccountBalance, error) {
			return []PeriodAccountBalance{}, nil
		},
	}

	base := NewPeriodCloseService(periods, balances, JournalEntryDAOMock{}, JournalLineDAOMock{}, PeriodCloseDAOMock{}, NewPostingService(JournalEntryDAOMock{}))
	svc := base.WithJournalResolver(closingJournalMock{id: 1})

	entry, err := svc.Close(ctx, 10, 1, 900)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.State != PeriodCloseStatePosted {
		t.Errorf("entry state = %s, want posted", entry.State)
	}
	if updatedState != TaxPeriodStateClosed {
		t.Errorf("period state = %s, want closed", updatedState)
	}
}
