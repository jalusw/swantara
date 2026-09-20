package accounting

import (
	"context"
	"sort"
	"strings"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

type ReconcileRuleEngine struct {
	rules    ReconcileRuleDAO
	matches  ReconcileRuleMatchDAO
	lines    JournalLineDAO
	partials AccountPartialReconcileDAO
	tx       *gorm.DB
}

func NewReconcileRuleEngine(
	rules ReconcileRuleDAO,
	matches ReconcileRuleMatchDAO,
	lines JournalLineDAO,
	partials AccountPartialReconcileDAO,
	db *gorm.DB,
) ReconcileRuleEngine {
	return ReconcileRuleEngine{
		rules:    rules,
		matches:  matches,
		lines:    lines,
		partials: partials,
		tx:       db,
	}
}

type ReconcileCandidate struct {
	Rule         *ReconcileRule
	DebitLineID  uint64
	CreditLineID uint64
	Score        int
}

func (e ReconcileRuleEngine) FindMatches(ctx context.Context, organizationID uint64) ([]ReconcileCandidate, error) {
	rules, err := e.rules.ListByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return nil, nil
	}

	var candidates []ReconcileCandidate
	for _, rule := range rules {
		ruleCandidates, err := e.matchRule(ctx, rule)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, ruleCandidates...)
	}
	return candidates, nil
}

type ReconcileProposal struct {
	RuleID       uint64
	RuleName     string
	DebitLineID  uint64
	CreditLineID uint64
	Score        int
	Amount       amount.Amount
}

func (e ReconcileRuleEngine) SuggestMatches(ctx context.Context, organizationID uint64, minScore int) ([]ReconcileProposal, error) {
	candidates, err := e.FindMatches(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	proposals := make([]ReconcileProposal, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.Score < minScore {
			continue
		}
		remaining, err := e.remaining(ctx, candidate.DebitLineID, true)
		if err != nil {
			return nil, err
		}
		creditRemaining, err := e.remaining(ctx, candidate.CreditLineID, false)
		if err != nil {
			return nil, err
		}
		if remaining.IsZero() || creditRemaining.IsZero() {
			continue
		}
		proposalAmount := remaining
		if creditRemaining.LessThan(remaining) {
			proposalAmount = creditRemaining
		}
		name := ""
		if candidate.Rule != nil && candidate.Rule.Name != nil {
			name = *candidate.Rule.Name
		}
		ruleID := uint64(0)
		if candidate.Rule != nil {
			ruleID = candidate.Rule.ID
		}
		proposals = append(proposals, ReconcileProposal{
			RuleID:       ruleID,
			RuleName:     name,
			DebitLineID:  candidate.DebitLineID,
			CreditLineID: candidate.CreditLineID,
			Score:        candidate.Score,
			Amount:       proposalAmount,
		})
	}
	sort.Slice(proposals, func(i, j int) bool {
		if proposals[i].Score != proposals[j].Score {
			return proposals[i].Score > proposals[j].Score
		}
		return proposals[i].Amount.GreaterThan(proposals[j].Amount)
	})
	return proposals, nil
}

func (e ReconcileRuleEngine) ListRules(ctx context.Context, q *query.Query) (*query.Page[ReconcileRule], error) {
	return e.rules.List(ctx, q)
}

func (e ReconcileRuleEngine) CreateRule(ctx context.Context, rule *ReconcileRule) (*ReconcileRule, error) {
	return e.rules.Create(ctx, rule)
}

func (e ReconcileRuleEngine) ApplyMatches(ctx context.Context, candidates []ReconcileCandidate, minScore int) (int, error) {
	filtered := candidates[:0]
	for _, candidate := range candidates {
		if candidate.Score >= minScore {
			filtered = append(filtered, candidate)
		}
	}
	applied := 0
	for _, candidate := range filtered {
		remaining, err := e.remaining(ctx, candidate.DebitLineID, true)
		if err != nil {
			return applied, err
		}
		if remaining.IsZero() {
			continue
		}

		creditRemaining, err := e.remaining(ctx, candidate.CreditLineID, false)
		if err != nil {
			return applied, err
		}
		if creditRemaining.IsZero() {
			continue
		}

		reconcileAmount := remaining
		if creditRemaining.LessThan(remaining) {
			reconcileAmount = creditRemaining
		}

		partial := &AccountPartialReconcile{
			DebitLineID:  candidate.DebitLineID,
			CreditLineID: candidate.CreditLineID,
			Amount:       reconcileAmount.Float64(),
		}
		if _, err := e.partials.Create(ctx, partial); err != nil {
			return applied, err
		}

		match := &ReconcileRuleMatch{
			RuleID:       candidate.Rule.ID,
			DebitLineID:  candidate.DebitLineID,
			CreditLineID: candidate.CreditLineID,
			Score:        candidate.Score,
		}
		if _, err := e.matches.Create(ctx, match); err != nil {
			return applied, err
		}
		if remaining.Equal(reconcileAmount) {
			line, err := e.lines.Find(ctx, candidate.DebitLineID)
			if err != nil {
				return applied, err
			}
			if line != nil {
				line.Reconciled = true
				if _, err := e.lines.Update(ctx, line); err != nil {
					return applied, err
				}
			}
		}
		if creditRemaining.Equal(reconcileAmount) {
			line, err := e.lines.Find(ctx, candidate.CreditLineID)
			if err != nil {
				return applied, err
			}
			if line != nil {
				line.Reconciled = true
				if _, err := e.lines.Update(ctx, line); err != nil {
					return applied, err
				}
			}
		}
		applied++
	}
	return applied, nil
}

func (e ReconcileRuleEngine) matchRule(ctx context.Context, rule *ReconcileRule) ([]ReconcileCandidate, error) {
	if rule.AccountID == nil {
		return nil, ErrReconcileRuleNoAccount
	}
	var candidates []ReconcileCandidate

	lines, err := e.lines.ListUnreconciledByAccount(ctx, *rule.AccountID)
	if err != nil {
		return nil, err
	}
	debits := make([]*JournalLine, 0)
	credits := make([]*JournalLine, 0)
	for _, line := range lines {
		if line.Debit.IsPositive() {
			debits = append(debits, line)
		} else if line.Credit.IsPositive() {
			credits = append(credits, line)
		}
	}

	for _, debit := range debits {
		for _, credit := range credits {
			if debit.ID == credit.ID {
				continue
			}
			score := e.scoreMatch(rule, debit, credit)
			if score > 0 {
				candidates = append(candidates, ReconcileCandidate{
					Rule:         rule,
					DebitLineID:  debit.ID,
					CreditLineID: credit.ID,
					Score:        score,
				})
			}
		}
	}
	return candidates, nil
}

func (e ReconcileRuleEngine) scoreMatch(rule *ReconcileRule, debit, credit *JournalLine) int {
	score := 0

	if rule.MatchContact {
		if debit.ContactID != nil && credit.ContactID != nil && *debit.ContactID == *credit.ContactID {
			score += 40
		} else {
			return 0
		}
	}

	if rule.MatchAmount {
		diff := debit.Debit.Sub(credit.Credit).Abs()
		tolerance := amount.FromFloat64(rule.AmountTolerance)
		if diff.LessThan(tolerance) || diff.Equal(tolerance) {
			score += 30
		} else {
			return 0
		}
	}

	if rule.MatchRef {
		if debit.Name != nil && credit.Name != nil {
			if strings.EqualFold(*debit.Name, *credit.Name) {
				score += 20
			} else if strings.Contains(*debit.Name, *credit.Name) || strings.Contains(*credit.Name, *debit.Name) {
				score += 10
			} else {
				return 0
			}
		}
	}

	return score
}

func (e ReconcileRuleEngine) remaining(ctx context.Context, lineID uint64, debit bool) (amount.Amount, error) {
	line, err := e.lines.Find(ctx, lineID)
	if err != nil {
		return amount.Zero(), err
	}
	if line == nil {
		return amount.Zero(), ErrLineNotFound
	}

	var nominal amount.Amount
	if debit {
		nominal = line.Debit
	} else {
		nominal = line.Credit
	}

	total, err := e.partials.TotalByLine(ctx, lineID)
	if err != nil {
		return amount.Zero(), err
	}
	return nominal.Sub(amount.FromFloat64(total)), nil
}
