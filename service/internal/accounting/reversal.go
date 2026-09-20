package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"gorm.io/gorm"
)

type ReverseRequest struct {
	OrganizationID uint64
	JournalID      uint64
	Date           time.Time
	Ref            string
	Description    string
	EntryID        uint64
}

type JournalEntryReverser interface {
	Reverse(ctx context.Context, request ReverseRequest) (*JournalEntry, error)
	ReverseTx(ctx context.Context, tx *gorm.DB, request ReverseRequest) (*JournalEntry, error)
}

type ReversalEngine struct {
	movements JournalEntryDAO
	lines     JournalLineDAO
	periods   PeriodLookup
	now       func() time.Time
}

func NewReversalEngine(movements JournalEntryDAO, lines JournalLineDAO) ReversalEngine {
	return ReversalEngine{movements: movements, lines: lines, now: time.Now}
}

func (s ReversalEngine) SetPeriods(periods PeriodLookup) ReversalEngine {
	s.periods = periods
	return s
}

func (s ReversalEngine) Reverse(ctx context.Context, request ReverseRequest) (*JournalEntry, error) {
	entry, lines, err := s.buildReversal(ctx, request)
	if err != nil {
		return nil, err
	}
	return s.movements.CreateWithLines(ctx, entry, lines)
}

func (s ReversalEngine) ReverseTx(ctx context.Context, tx *gorm.DB, request ReverseRequest) (*JournalEntry, error) {
	entry, lines, err := s.buildReversal(ctx, request)
	if err != nil {
		return nil, err
	}
	return s.movements.CreateWithLinesTx(ctx, tx, entry, lines)
}

func (s ReversalEngine) buildReversal(ctx context.Context, request ReverseRequest) (*JournalEntry, []*JournalLine, error) {
	original, err := s.movements.Find(ctx, request.EntryID)
	if err != nil {
		return nil, nil, err
	}
	if original == nil || original.OrganizationID != request.OrganizationID {
		return nil, nil, ErrEntryNotFound
	}
	if original.State != EntryStatePosted {
		return nil, nil, ErrEntryNotPosted
	}

	existing, err := s.movements.FindByReversed(ctx, original.ID)
	if err != nil {
		return nil, nil, err
	}
	if existing != nil {
		return nil, nil, ErrEntryReversed
	}

	originalLines, err := s.lines.ListByMovement(ctx, original.ID)
	if err != nil {
		return nil, nil, err
	}

	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	if s.periods != nil {
		if err := s.periods.AssertOpen(ctx, request.OrganizationID, date); err != nil {
			return nil, nil, err
		}
	}
	name := request.Description
	ref := request.Ref
	journalID := request.JournalID
	if journalID == 0 {
		journalID = original.JournalID
	}
	reversal := &JournalEntry{
		OrganizationID:  request.OrganizationID,
		JournalID:       journalID,
		Name:            &name,
		Date:            date,
		Ref:             &ref,
		State:           EntryStatePosted,
		CurrencyCode:    original.CurrencyCode,
		OriginType:      helper.Ptr(OriginTypeReversal),
		OriginID:        &original.ID,
		ReversedEntryID: &original.ID,
	}

	lines := make([]*JournalLine, len(originalLines))
	for i, line := range originalLines {
		lines[i] = &JournalLine{
			AccountID:      line.AccountID,
			DimensionID:    line.DimensionID,
			Name:           line.Name,
			Debit:          line.Credit,
			Credit:         line.Debit,
			ContactID:      line.ContactID,
			TaxID:          line.TaxID,
			CurrencyCode:   line.CurrencyCode,
			AmountCurrency: line.AmountCurrency,
			DueDate:        line.DueDate,
			Reconciled:     false,
		}
	}

	return reversal, lines, nil
}
