package reporting

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func validAccrualRequest() AccrualRequest {
	return AccrualRequest{
		OrganizationID: 10,
		PeriodID:       3,
		Name:           "Year-end bonus",
		Date:           time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		Lines: []AccrualLineRequest{
			{AccountID: 600, Name: "Bonus", Debit: 1000},
			{AccountID: 200, Name: "Payable", Credit: 1000},
		},
	}
}

func TestAccrualService_Create_Branches(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")

	newSvc := func(journalErr error, postErr error, createErr error, lineErr error) AccrualService {
		poster := &PosterMock{}
		if postErr != nil || true {
			poster.PostFn = func(_ context.Context, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
				if postErr != nil {
					return nil, postErr
				}
				return &accounting.JournalEntry{Base: model.Base{ID: 9}}, nil
			}
		}
		config := ConfigSourceMock{}
		if journalErr != nil {
			config.JournalIDFn = func(_ context.Context, _ uint64) (uint64, error) { return 0, journalErr }
		}
		accruals := AccrualDAOMock{
			CreateFn: func(_ context.Context, a *Accrual) (*Accrual, error) {
				if createErr != nil {
					return nil, createErr
				}
				a.ID = 4
				return a, nil
			},
		}
		lines := AccrualLineDAOMock{
			CreateFn: func(_ context.Context, l *AccrualLine) (*AccrualLine, error) {
				if lineErr != nil {
					return nil, lineErr
				}
				return l, nil
			},
		}
		return NewAccrualService(accruals, lines, poster, config)
	}

	t.Run("creates with default date", func(t *testing.T) {
		svc := newSvc(nil, nil, nil, nil)
		req := validAccrualRequest()
		req.Date = time.Time{}
		got, err := svc.Create(ctx, req)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got.AccrualID != 4 || got.State != AccrualStatePosted {
			t.Errorf("result = %+v", got)
		}
	})

	t.Run("journal error", func(t *testing.T) {
		svc := newSvc(dbErr, nil, nil, nil)
		_, err := svc.Create(ctx, validAccrualRequest())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("post error", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil, nil)
		_, err := svc.Create(ctx, validAccrualRequest())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("accrual persist error", func(t *testing.T) {
		svc := newSvc(nil, nil, dbErr, nil)
		_, err := svc.Create(ctx, validAccrualRequest())
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("line persist error", func(t *testing.T) {
		svc := newSvc(nil, nil, nil, dbErr)
		_, err := svc.Create(ctx, validAccrualRequest())
		helper.AssertError(t, err, true, dbErr)
	})
}

func TestValidateAccrualLines(t *testing.T) {
	cases := []struct {
		name    string
		lines   []AccrualLineRequest
		wantErr bool
	}{
		{name: "empty", lines: nil, wantErr: true},
		{name: "negative", lines: []AccrualLineRequest{{AccountID: 1, Debit: -5}}, wantErr: true},
		{name: "both sides", lines: []AccrualLineRequest{{AccountID: 1, Debit: 5, Credit: 5}}, wantErr: true},
		{name: "unbalanced", lines: []AccrualLineRequest{{AccountID: 1, Debit: 5}}, wantErr: true},
		{name: "balanced", lines: []AccrualLineRequest{{AccountID: 1, Debit: 5}, {AccountID: 2, Credit: 5}}, wantErr: false},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAccrualLines(tt.lines)
			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestAccrualService_ReverseDue_Branches(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("db down")
	asOf := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	past := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	moveID := uint64(9)

	dueAccrual := func() *Accrual {
		return &Accrual{Base: model.Base{ID: 4}, State: AccrualStatePosted, EntryID: &moveID, ReversalDate: &past}
	}

	newSvc := func(accruals []*Accrual, listErr error, journalErr error, reverseErr error, updateErr error) AccrualService {
		accrualMock := AccrualDAOMock{
			ListByOrganizationFn: func(_ context.Context, _ uint64) ([]*Accrual, error) { return accruals, listErr },
			UpdateFn: func(_ context.Context, a *Accrual) (*Accrual, error) {
				if updateErr != nil {
					return nil, updateErr
				}
				return a, nil
			},
		}
		config := ConfigSourceMock{}
		if journalErr != nil {
			config.JournalIDFn = func(_ context.Context, _ uint64) (uint64, error) { return 0, journalErr }
		}
		poster := &PosterMock{
			ReverseFn: func(_ context.Context, _ accounting.ReverseRequest) (*accounting.JournalEntry, error) {
				if reverseErr != nil {
					return nil, reverseErr
				}
				return &accounting.JournalEntry{Base: model.Base{ID: 10}}, nil
			},
		}
		return NewAccrualService(accrualMock, AccrualLineDAOMock{}, poster, config)
	}

	t.Run("reverses due accrual", func(t *testing.T) {
		svc := newSvc([]*Accrual{dueAccrual()}, nil, nil, nil, nil)
		got, err := svc.ReverseDue(ctx, 10, asOf)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != 1 {
			t.Errorf("reversed = %d, want 1", got)
		}
	})

	t.Run("skips ineligible", func(t *testing.T) {
		future := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
		draft := &Accrual{Base: model.Base{ID: 5}, State: "draft", EntryID: &moveID, ReversalDate: &past}
		noMove := &Accrual{Base: model.Base{ID: 6}, State: AccrualStatePosted, ReversalDate: &past}
		notDue := &Accrual{Base: model.Base{ID: 7}, State: AccrualStatePosted, EntryID: &moveID, ReversalDate: &future}
		svc := newSvc([]*Accrual{draft, noMove, notDue}, nil, nil, nil, nil)
		got, err := svc.ReverseDue(ctx, 10, asOf)
		if helper.AssertError(t, err, false, nil) {
			return
		}
		if got != 0 {
			t.Errorf("reversed = %d, want 0", got)
		}
	})

	t.Run("list error", func(t *testing.T) {
		svc := newSvc(nil, dbErr, nil, nil, nil)
		_, err := svc.ReverseDue(ctx, 10, asOf)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("journal error", func(t *testing.T) {
		svc := newSvc([]*Accrual{dueAccrual()}, nil, dbErr, nil, nil)
		_, err := svc.ReverseDue(ctx, 10, asOf)
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("reverse error", func(t *testing.T) {
		svc := newSvc([]*Accrual{dueAccrual()}, nil, nil, dbErr, nil)
		got, err := svc.ReverseDue(ctx, 10, asOf)
		if got != 0 {
			t.Errorf("reversed = %d, want 0", got)
		}
		helper.AssertError(t, err, true, dbErr)
	})

	t.Run("update error", func(t *testing.T) {
		svc := newSvc([]*Accrual{dueAccrual()}, nil, nil, nil, dbErr)
		got, err := svc.ReverseDue(ctx, 10, asOf)
		if got != 0 {
			t.Errorf("reversed = %d, want 0", got)
		}
		helper.AssertError(t, err, true, dbErr)
	})
}
