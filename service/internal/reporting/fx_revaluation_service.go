package reporting

import (
	"context"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type RateResolver interface {
	Rate(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error)
}

type Poster interface {
	Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
}

type OrgReader interface {
	Find(ctx context.Context, id uint64) (*reference.Organization, error)
}

type FxRevaluationService struct {
	reports      ReportDAO
	revaluations FxRevaluationDAO
	lines        FxRevaluationLineDAO
	poster       Poster
	config       ConfigSource
	rates        RateResolver
	orgs         OrgReader
}

type position struct {
	currency    string
	accountID   uint64
	accountType string
	foreign     amount.Amount
	base        amount.Amount
	delta       amount.Amount
	rate        amount.Amount
}

func NewFxRevaluationService(
	reports ReportDAO,
	revaluations FxRevaluationDAO,
	lines FxRevaluationLineDAO,
	poster Poster,
	config ConfigSource,
	rates RateResolver,
	orgs OrgReader,
) FxRevaluationService {
	return FxRevaluationService{
		reports:      reports,
		revaluations: revaluations,
		lines:        lines,
		poster:       poster,
		config:       config,
		rates:        rates,
		orgs:         orgs,
	}
}

func (s FxRevaluationService) ListRunsByOrganization(ctx context.Context, organizationID uint64) ([]*FxRevaluation, error) {
	return s.revaluations.ListByOrganization(ctx, organizationID)
}

type RevaluationResult struct {
	RevaluationID uint64            `json:"revaluation_id"`
	PeriodID      uint64            `json:"period_id"`
	Date          time.Time         `json:"date"`
	Reversed      int               `json:"reversed_lines"`
	Posted        int               `json:"posted_lines"`
	TotalGainLoss float64           `json:"total_gain_loss"`
	Lines         []RevaluationLine `json:"lines"`
}

type RevaluationLine struct {
	AccountID      uint64  `json:"account_id"`
	CurrencyCode   string  `json:"currency_code"`
	ForeignBalance float64 `json:"foreign_balance"`
	BaseBalance    float64 `json:"base_balance"`
	ClosingRate    float64 `json:"closing_rate"`
	GainLoss       float64 `json:"gain_loss"`
	EntryID        *uint64 `json:"entry_id"`
}

func (s FxRevaluationService) Revalue(ctx context.Context, organizationID, periodID uint64, date time.Time) (RevaluationResult, error) {
	journalID, err := s.config.JournalID(ctx, organizationID)
	if err != nil {
		return RevaluationResult{}, err
	}
	gainLossAccount, err := s.config.FxGainLossAccountID(ctx, organizationID)
	if err != nil {
		return RevaluationResult{}, err
	}
	org, err := s.orgs.Find(ctx, organizationID)
	if err != nil {
		return RevaluationResult{}, err
	}
	if org == nil {
		return RevaluationResult{}, ErrBaseCurrencyMissing
	}

	reversed, err := s.reverseOpen(ctx, organizationID, journalID, date)
	if err != nil {
		return RevaluationResult{}, err
	}

	positions, err := s.reports.OpenForeignPositions(ctx, organizationID)
	if err != nil {
		return RevaluationResult{}, err
	}

	groups := map[string]*position{}
	for _, row := range positions {
		if row.CurrencyCode == org.BaseCurrency {
			continue
		}
		if row.AmountTotal == 0 || row.AmountResidual == 0 {
			continue
		}
		rate, err := s.rates.Rate(ctx, row.CurrencyCode, organizationID, amount.RateClosing, date)
		if err != nil {
			return RevaluationResult{}, ErrClosingRateMissing
		}
		foreignResidual := amount.FromFloat64(row.AmountResidual)
		baseResidual := foreignResidual.Mul(amount.FromFloat64(row.BaseBalance / row.AmountTotal)).Round(4)
		closing := foreignResidual.Mul(rate).Round(4)
		delta := closing.Sub(baseResidual).Round(4)
		if delta.IsZero() {
			continue
		}

		key := fmt.Sprintf("%d:%s", row.AccountID, row.CurrencyCode)
		existing, ok := groups[key]
		if !ok {
			existing = &position{currency: row.CurrencyCode, accountID: row.AccountID, accountType: row.AccountType}
			groups[key] = existing
		}
		existing.foreign = existing.foreign.Add(foreignResidual)
		existing.base = existing.base.Add(baseResidual)
		existing.delta = existing.delta.Add(delta)
		if existing.rate.IsZero() {
			existing.rate = rate
		}
	}

	if len(groups) == 0 {
		return RevaluationResult{PeriodID: periodID, Date: date, Reversed: reversed}, nil
	}

	run := &FxRevaluation{
		OrganizationID: organizationID,
		PeriodID:       periodID,
		Name:           helper.Ptr(fmt.Sprintf("FX Revaluation %s", date.Format("2006-01-02"))),
		Date:           date,
		State:          FxRevaluationStatePosted,
	}
	created, err := s.revaluations.Create(ctx, run)
	if err != nil {
		return RevaluationResult{}, err
	}

	result := RevaluationResult{RevaluationID: created.ID, PeriodID: periodID, Date: date, Reversed: reversed}
	for _, group := range groups {
		line, err := s.postRevaluation(ctx, created.ID, organizationID, journalID, gainLossAccount, date, group)
		if err != nil {
			return RevaluationResult{}, err
		}
		result.Posted++
		result.TotalGainLoss += line.GainLoss
		result.Lines = append(result.Lines, line)
	}
	return result, nil
}

func (s FxRevaluationService) postRevaluation(
	ctx context.Context,
	runID, organizationID, journalID, gainLossAccount uint64,
	date time.Time,
	group *position,
) (RevaluationLine, error) {
	gainLoss := group.delta
	if group.accountType == "payable" {
		gainLoss = group.delta.Neg()
	}

	debitAccount, creditAccount := s.legs(group, gainLossAccount)
	accountName := "AR/AP Revaluation"
	gainLossName := "FX Gain/Loss (unrealized)"

	movement, err := s.poster.Post(ctx, accounting.PostRequest{
		OrganizationID: organizationID,
		JournalID:      journalID,
		Date:           date,
		Ref:            fmt.Sprintf("FXREVAL/%d", runID),
		OriginType:     accounting.OriginTypeFxRevaluation,
		OriginID:       runID,
		Description:    fmt.Sprintf("FX revaluation %s %s", group.currency, date.Format("2006-01-02")),
		Lines: []accounting.PostingLine{
			{AccountID: debitAccount, Name: accountName, Debit: group.delta.Abs()},
			{AccountID: creditAccount, Name: gainLossName, Credit: group.delta.Abs()},
		},
	})
	if err != nil {
		return RevaluationLine{}, err
	}

	entity := &FxRevaluationLine{
		RevaluationID:  runID,
		AccountID:      group.accountID,
		CurrencyCode:   group.currency,
		ForeignBalance: group.foreign.Round(4).Float64(),
		BaseBalance:    group.base.Round(4).Float64(),
		ClosingRate:    group.rate.Round(8).Float64(),
		GainLoss:       gainLoss.Round(4).Float64(),
		EntryID:        &movement.ID,
	}
	if _, err := s.lines.Create(ctx, entity); err != nil {
		return RevaluationLine{}, err
	}

	return RevaluationLine{
		AccountID:      group.accountID,
		CurrencyCode:   group.currency,
		ForeignBalance: entity.ForeignBalance,
		BaseBalance:    entity.BaseBalance,
		ClosingRate:    entity.ClosingRate,
		GainLoss:       entity.GainLoss,
		EntryID:        &movement.ID,
	}, nil
}

func (s FxRevaluationService) legs(group *position, gainLossAccount uint64) (uint64, uint64) {
	if group.accountType == "payable" {
		if group.delta.IsPositive() {
			return gainLossAccount, group.accountID
		}
		return group.accountID, gainLossAccount
	}
	if group.delta.IsPositive() {
		return group.accountID, gainLossAccount
	}
	return gainLossAccount, group.accountID
}

func (s FxRevaluationService) reverseOpen(ctx context.Context, organizationID, journalID uint64, date time.Time) (int, error) {
	open, err := s.lines.ListOpenByOrg(ctx, organizationID)
	if err != nil {
		return 0, err
	}
	affected := map[uint64]bool{}
	reversed := 0
	for _, line := range open {
		if line.EntryID == nil {
			continue
		}
		_, err := s.poster.Reverse(ctx, accounting.ReverseRequest{
			OrganizationID: organizationID,
			JournalID:      journalID,
			Date:           date,
			Ref:            fmt.Sprintf("FXREVAL-RV/%d", line.RevaluationID),
			Description:    "FX revaluation reversal",
			EntryID:        *line.EntryID,
		})
		if err != nil {
			return reversed, err
		}
		line.Reversed = true
		if _, err := s.lines.Update(ctx, line); err != nil {
			return reversed, err
		}
		affected[line.RevaluationID] = true
		reversed++
	}
	for runID := range affected {
		lines, err := s.lines.ListByRevaluation(ctx, runID)
		if err != nil {
			return reversed, err
		}
		allReversed := true
		for _, line := range lines {
			if !line.Reversed {
				allReversed = false
				break
			}
		}
		if !allReversed {
			continue
		}
		run, err := s.revaluations.Find(ctx, runID)
		if err != nil {
			return reversed, err
		}
		if run == nil {
			continue
		}
		run.State = FxRevaluationStateReversed
		if _, err := s.revaluations.Update(ctx, run); err != nil {
			return reversed, err
		}
	}
	return reversed, nil
}
