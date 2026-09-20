package accounting

import (
	"context"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

const postingPrecision int32 = 4

const SequenceMoveCode = "journal_entry"

type PostingLine struct {
	AccountID      uint64
	DimensionID    *uint64
	Name           string
	Debit          amount.Amount
	Credit         amount.Amount
	ContactID      *uint64
	TaxID          *uint64
	CurrencyCode   *string
	AmountCurrency float64
	DueDate        *time.Time
}

type PostRequest struct {
	OrganizationID  uint64
	JournalID       uint64
	Date            time.Time
	Ref             string
	OriginType      string
	OriginID        uint64
	Description     string
	CurrencyCode    *string
	Draft           bool
	AutoReverseDate *time.Time
	Lines           []PostingLine
}

type PostingService struct {
	movements  JournalEntryDAO
	lines      JournalLineDAO
	periods    PeriodLookup
	reverser   JournalEntryReverser
	dimensions DimensionDistributionDAO
	accounts   AccountFinder
	sequences  *sequence.Service
	now        func() time.Time
}

type AccountFinder interface {
	Find(ctx context.Context, id uint64) (*reference.Account, error)
}

func NewPostingService(movements JournalEntryDAO) PostingService {
	return PostingService{movements: movements, now: time.Now}
}

func (s PostingService) WithLines(lines JournalLineDAO) PostingService {
	s.lines = lines
	return s
}

func (s PostingService) SetPeriods(periods PeriodLookup) PostingService {
	s.periods = periods
	return s
}

func (s PostingService) SetReverser(reverser JournalEntryReverser) PostingService {
	s.reverser = reverser
	return s
}

func (s PostingService) WithDimensions(dimensions DimensionDistributionDAO) PostingService {
	s.dimensions = dimensions
	return s
}

func (s PostingService) WithAccounts(accounts AccountFinder) PostingService {
	s.accounts = accounts
	return s
}

func (s PostingService) WithSequences(sequences sequence.Service) PostingService {
	s.sequences = &sequences
	return s
}

func (s PostingService) Post(ctx context.Context, request PostRequest) (*JournalEntry, error) {
	if err := s.assertPeriodOpen(ctx, request); err != nil {
		return nil, err
	}
	if err := s.assignSequence(ctx, &request); err != nil {
		return nil, err
	}
	entry, lines, err := buildMove(ctx, request, s.now)
	if err != nil {
		return nil, err
	}
	if err := s.assertAccounts(ctx, request.OrganizationID, request.Lines); err != nil {
		return nil, err
	}
	created, err := s.movements.CreateWithLines(ctx, entry, lines)
	if err != nil {
		return nil, err
	}
	if s.dimensions != nil {
		if err := s.createDimensionDistributions(ctx, nil, lines); err != nil {
			return nil, err
		}
	}
	if request.AutoReverseDate != nil && !request.Draft {
		if s.reverser == nil {
			return nil, ErrReverserMissing
		}
		if _, err := s.reverser.Reverse(ctx, ReverseRequest{
			OrganizationID: request.OrganizationID,
			JournalID:      request.JournalID,
			Date:           *request.AutoReverseDate,
			Ref:            request.Ref + "/REV",
			Description:    "Auto-reverse " + request.Description,
			EntryID:        created.ID,
		}); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (s PostingService) PostTx(ctx context.Context, tx *gorm.DB, request PostRequest) (*JournalEntry, error) {
	if err := s.assertPeriodOpen(ctx, request); err != nil {
		return nil, err
	}
	if err := s.assignSequence(ctx, &request); err != nil {
		return nil, err
	}
	entry, lines, err := buildMove(ctx, request, s.now)
	if err != nil {
		return nil, err
	}
	if err := s.assertAccounts(ctx, request.OrganizationID, request.Lines); err != nil {
		return nil, err
	}
	created, err := s.movements.CreateWithLinesTx(ctx, tx, entry, lines)
	if err != nil {
		return nil, err
	}
	if s.dimensions != nil && tx != nil {
		if err := s.createDimensionDistributions(ctx, tx, lines); err != nil {
			return nil, err
		}
	}
	if request.AutoReverseDate != nil && !request.Draft {
		if s.reverser == nil {
			return nil, ErrReverserMissing
		}
		if _, err := s.reverser.ReverseTx(ctx, tx, ReverseRequest{
			OrganizationID: request.OrganizationID,
			JournalID:      request.JournalID,
			Date:           *request.AutoReverseDate,
			Ref:            request.Ref + "/REV",
			Description:    "Auto-reverse " + request.Description,
			EntryID:        created.ID,
		}); err != nil {
			return nil, err
		}
	}
	return created, nil
}

func (s PostingService) createDimensionDistributions(ctx context.Context, tx *gorm.DB, lines []*JournalLine) error {
	dists := make([]*DimensionDistribution, 0)
	for _, line := range lines {
		if line.DimensionID == nil {
			continue
		}
		amt := line.Debit
		if amt.IsZero() {
			amt = line.Credit
		}
		dists = append(dists, &DimensionDistribution{
			JournalLineID: line.ID,
			DimensionID:   *line.DimensionID,
			Percent:       100,
			Amount:        amt,
		})
	}
	if len(dists) == 0 {
		return nil
	}
	if tx != nil {
		return s.dimensions.CreateDistributions(ctx, tx, dists)
	}
	for _, dist := range dists {
		if _, err := s.dimensions.Create(ctx, dist); err != nil {
			return err
		}
	}
	return nil
}

func (s PostingService) Reverse(ctx context.Context, request ReverseRequest) (*JournalEntry, error) {
	if s.reverser == nil {
		return nil, ErrReverserMissing
	}
	return s.reverser.Reverse(ctx, request)
}
func (s PostingService) ReverseTx(ctx context.Context, tx *gorm.DB, request ReverseRequest) (*JournalEntry, error) {
	if s.reverser == nil {
		return nil, ErrReverserMissing
	}
	return s.reverser.ReverseTx(ctx, tx, request)
}

func (s PostingService) FindEntry(ctx context.Context, id uint64) (*JournalEntry, error) {
	return s.movements.Find(ctx, id)
}

func (s PostingService) Confirm(ctx context.Context, organizationID, entryID uint64) (*JournalEntry, error) {
	entry, err := s.movements.Find(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if entry == nil || entry.OrganizationID != organizationID {
		return nil, ErrEntryNotFound
	}
	if entry.State != EntryStateDraft {
		return nil, ErrEntryNotDraft
	}
	if s.periods != nil {
		if err := s.periods.AssertOpen(ctx, organizationID, entry.Date); err != nil {
			return nil, err
		}
	}
	postedAt := s.now().UTC()
	postedBy := model.ActorID(ctx)
	entry.State = EntryStatePosted
	entry.PostedAt = &postedAt
	entry.PostedBy = nullableUint(postedBy)
	return s.movements.Update(ctx, entry)
}

func (s PostingService) assignSequence(ctx context.Context, request *PostRequest) error {
	if s.sequences == nil || request.Ref != "" {
		return nil
	}
	number, err := s.sequences.Next(ctx, request.OrganizationID, SequenceMoveCode)
	if err != nil {
		return err
	}
	request.Ref = number
	if request.Description == "" {
		request.Description = number
	}
	return nil
}

func (s PostingService) ListEntries(ctx context.Context, q *query.Query) (*query.Page[JournalEntry], error) {
	return s.movements.List(ctx, q)
}

func (s PostingService) ListLinesByMove(ctx context.Context, entryID uint64) ([]*JournalLine, error) {
	return s.lines.ListByMovement(ctx, entryID)
}

func (s PostingService) assertPeriodOpen(ctx context.Context, request PostRequest) error {
	if s.periods == nil {
		return nil
	}
	return s.periods.AssertOpen(ctx, request.OrganizationID, request.Date)
}

func (s PostingService) assertAccounts(ctx context.Context, organizationID uint64, lines []PostingLine) error {
	if s.accounts == nil {
		return nil
	}
	seen := make(map[uint64]*reference.Account, len(lines))
	for _, line := range lines {
		if _, ok := seen[line.AccountID]; !ok {
			found, err := s.accounts.Find(ctx, line.AccountID)
			if err != nil {
				return err
			}
			if found == nil || found.OrganizationID != organizationID {
				return ErrAccountNotFound
			}
			if !found.Active {
				return ErrAccountInactive
			}
			seen[line.AccountID] = found
		}
	}
	return nil
}

func buildMove(ctx context.Context, request PostRequest, now func() time.Time) (*JournalEntry, []*JournalLine, error) {
	if request.OrganizationID == 0 || request.JournalID == 0 {
		return nil, nil, ErrInvalidLine
	}
	if err := validateLines(request.Lines); err != nil {
		return nil, nil, err
	}
	if request.OriginID != 0 && request.OriginType == "" {
		return nil, nil, ErrOriginIncomplete
	}

	name := request.Description
	if name == "" {
		name = request.Ref
	}
	ref := request.Ref
	state := EntryStatePosted
	var postedAt *time.Time
	var postedBy *uint64
	if request.Draft {
		state = EntryStateDraft
	} else {
		now := now().UTC()
		postedAt = &now
		actor := model.ActorID(ctx)
		postedBy = nullableUint(actor)
	}

	entry := &JournalEntry{
		OrganizationID: request.OrganizationID,
		JournalID:      request.JournalID,
		Name:           &name,
		Date:           request.Date,
		Ref:            &ref,
		State:          state,
		CurrencyCode:   request.CurrencyCode,
		OriginType:     helper.StringPtr(request.OriginType),
		OriginID:       nullableUint(request.OriginID),
		PostedAt:       postedAt,
		PostedBy:       postedBy,
	}

	lines := make([]*JournalLine, len(request.Lines))
	for i, line := range request.Lines {
		name := line.Name
		lines[i] = &JournalLine{
			AccountID:      line.AccountID,
			DimensionID:    line.DimensionID,
			Name:           &name,
			Debit:          line.Debit.Round(postingPrecision),
			Credit:         line.Credit.Round(postingPrecision),
			ContactID:      line.ContactID,
			TaxID:          line.TaxID,
			CurrencyCode:   line.CurrencyCode,
			AmountCurrency: line.AmountCurrency,
			DueDate:        line.DueDate,
		}
	}

	return entry, lines, nil
}

func validateLines(lines []PostingLine) error {
	if len(lines) == 0 {
		return ErrNoLines
	}

	debits := amount.Zero()
	credits := amount.Zero()
	for _, line := range lines {
		if line.Debit.IsNegative() || line.Credit.IsNegative() {
			return ErrInvalidLine
		}
		if line.Debit.IsPositive() && line.Credit.IsPositive() {
			return ErrInvalidLine
		}
		if line.Debit.IsZero() && line.Credit.IsZero() {
			return ErrInvalidLine
		}
		if line.AccountID == 0 {
			return ErrInvalidLine
		}
		if line.CurrencyCode != nil && *line.CurrencyCode == "" {
			return ErrInvalidLine
		}
		if line.CurrencyCode != nil && line.AmountCurrency == 0 {
			return ErrInvalidLine
		}
		if line.CurrencyCode == nil && line.AmountCurrency != 0 {
			return ErrInvalidLine
		}
		debits = debits.Add(line.Debit)
		credits = credits.Add(line.Credit)
	}
	if !amount.IsBalanced(debits, credits, postingPrecision) {
		return ErrUnbalanced
	}
	return nil
}

func nullableUint(value uint64) *uint64 {
	if value == 0 {
		return nil
	}
	return &value
}
