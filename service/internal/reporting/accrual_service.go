package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type AccrualRequest struct {
	OrganizationID uint64
	PeriodID       uint64
	Name           string
	Description    string
	Date           time.Time
	ReversalDate   *time.Time
	Lines          []AccrualLineRequest
}

type AccrualLineRequest struct {
	AccountID uint64
	Name      string
	Debit     float64
	Credit    float64
}

type AccrualResult struct {
	AccrualID    uint64     `json:"accrual_id"`
	PeriodID     uint64     `json:"period_id"`
	State        string     `json:"state"`
	EntryID      *uint64    `json:"entry_id"`
	ReversalDate *time.Time `json:"reversal_date"`
}

type AccrualService struct {
	accruals AccrualDAO
	lines    AccrualLineDAO
	poster   Poster
	config   ConfigSource
}

func NewAccrualService(accruals AccrualDAO, lines AccrualLineDAO, poster Poster, config ConfigSource) AccrualService {
	return AccrualService{accruals: accruals, lines: lines, poster: poster, config: config}
}

func (s AccrualService) ListByOrganization(ctx context.Context, organizationID uint64) ([]*Accrual, error) {
	return s.accruals.ListByOrganization(ctx, organizationID)
}

func (s AccrualService) Create(ctx context.Context, request AccrualRequest) (AccrualResult, error) {
	if err := validateAccrualLines(request.Lines); err != nil {
		return AccrualResult{}, err
	}
	journalID, err := s.config.JournalID(ctx, request.OrganizationID)
	if err != nil {
		return AccrualResult{}, err
	}

	postLines := make([]accounting.PostingLine, len(request.Lines))
	for i, line := range request.Lines {
		postLines[i] = accounting.PostingLine{
			AccountID: line.AccountID,
			Name:      line.Name,
			Debit:     amount.FromFloat64(line.Debit),
			Credit:    amount.FromFloat64(line.Credit),
		}
	}

	movement, err := s.poster.Post(ctx, accounting.PostRequest{
		OrganizationID: request.OrganizationID,
		JournalID:      journalID,
		Date:           accrualDate(request),
		Ref:            fmt.Sprintf("ACCRUAL/%s", request.Name),
		OriginType:     accounting.OriginTypeAccrual,
		Description:    request.Description,
		Lines:          postLines,
	})
	if err != nil {
		return AccrualResult{}, err
	}

	accrual := &Accrual{
		OrganizationID: request.OrganizationID,
		PeriodID:       request.PeriodID,
		Name:           &request.Name,
		Description:    &request.Description,
		ReversalDate:   request.ReversalDate,
		State:          AccrualStatePosted,
		EntryID:        &movement.ID,
	}
	created, err := s.accruals.Create(ctx, accrual)
	if err != nil {
		return AccrualResult{}, err
	}
	for _, line := range request.Lines {
		name := line.Name
		if _, err := s.lines.Create(ctx, &AccrualLine{
			AccrualID: created.ID,
			AccountID: line.AccountID,
			Name:      &name,
			Debit:     line.Debit,
			Credit:    line.Credit,
		}); err != nil {
			return AccrualResult{}, err
		}
	}

	return AccrualResult{
		AccrualID:    created.ID,
		PeriodID:     created.PeriodID,
		State:        created.State,
		EntryID:      created.EntryID,
		ReversalDate: created.ReversalDate,
	}, nil
}

func (s AccrualService) ReverseDue(ctx context.Context, organizationID uint64, asOf time.Time) (int, error) {
	accruals, err := s.accruals.ListByOrganization(ctx, organizationID)
	if err != nil {
		return 0, err
	}
	journalID, err := s.config.JournalID(ctx, organizationID)
	if err != nil {
		return 0, err
	}

	reversed := 0
	for _, accrual := range accruals {
		if accrual.State != AccrualStatePosted || accrual.EntryID == nil {
			continue
		}
		if accrual.ReversalDate == nil || accrual.ReversalDate.After(asOf) {
			continue
		}
		reversal, err := s.poster.Reverse(ctx, accounting.ReverseRequest{
			OrganizationID: organizationID,
			JournalID:      journalID,
			Date:           asOf,
			Ref:            fmt.Sprintf("ACCRUAL-RV/%d", accrual.ID),
			Description:    "Accrual auto-reversal",
			EntryID:        *accrual.EntryID,
		})
		if err != nil {
			return reversed, err
		}
		accrual.State = AccrualStateReversed
		accrual.ReversalEntryID = &reversal.ID
		if _, err := s.accruals.Update(ctx, accrual); err != nil {
			return reversed, err
		}
		reversed++
	}
	return reversed, nil
}

func validateAccrualLines(lines []AccrualLineRequest) error {
	if len(lines) == 0 {
		return ErrUnbalancedAccrual
	}
	debits := amount.Zero()
	credits := amount.Zero()
	for _, line := range lines {
		if line.Debit < 0 || line.Credit < 0 {
			return ErrUnbalancedAccrual
		}
		if line.Debit > 0 && line.Credit > 0 {
			return ErrUnbalancedAccrual
		}
		debits = debits.Add(amount.FromFloat64(line.Debit))
		credits = credits.Add(amount.FromFloat64(line.Credit))
	}
	if !amount.IsBalanced(debits, credits, 4) {
		return ErrUnbalancedAccrual
	}
	return nil
}

func accrualDate(request AccrualRequest) time.Time {
	if !request.Date.IsZero() {
		return request.Date
	}
	return time.Now().UTC()
}
