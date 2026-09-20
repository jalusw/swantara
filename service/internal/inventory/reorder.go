package inventory

import (
	"context"
	"math"

	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type ReorderService struct {
	rules    ReorderRuleDAO
	ledger   LedgerService
	resolver ItemResolver
}

func NewReorderService(rules ReorderRuleDAO, ledger LedgerService, resolver ItemResolver) ReorderService {
	return ReorderService{rules: rules, ledger: ledger, resolver: resolver}
}

func (s ReorderService) ListInOrg(ctx context.Context, q *query.Query, organizationID uint64) (*query.Page[ReorderRule], error) {
	return s.rules.ListInOrg(ctx, q, organizationID)
}

func (s ReorderService) FindInOrg(ctx context.Context, id, organizationID uint64) (*ReorderRule, error) {
	return s.rules.FindInOrg(ctx, id, organizationID)
}

func (s ReorderService) Delete(ctx context.Context, id uint64) error {
	return s.rules.Delete(ctx, id)
}

func (s ReorderService) Create(ctx context.Context, organizationID uint64, rule *ReorderRule) (*ReorderRule, error) {
	if rule.LocationID == nil {
		return nil, ErrLocationRequired
	}
	if err := s.validateRule(ctx, organizationID, rule); err != nil {
		return nil, err
	}
	return s.rules.Create(ctx, rule)
}

func (s ReorderService) Update(ctx context.Context, organizationID uint64, rule *ReorderRule) (*ReorderRule, error) {
	if rule.LocationID == nil {
		return nil, ErrLocationRequired
	}
	if err := s.validateRule(ctx, organizationID, rule); err != nil {
		return nil, err
	}
	return s.rules.Update(ctx, rule)
}

func (s ReorderService) validateRule(ctx context.Context, organizationID uint64, rule *ReorderRule) error {
	if rule.MinQty < 0 || rule.MaxQty < rule.MinQty {
		return ErrInvalidRuleQty
	}
	if rule.QtyMultiple <= 0 {
		return ErrInvalidRuleQty
	}
	resolved, err := s.resolver.Resolve(ctx, rule.ItemID)
	if err != nil {
		return err
	}
	if resolved.OrganizationID == nil || *resolved.OrganizationID != organizationID {
		return ErrVariantNotFound
	}
	return nil
}

type ReplenishmentCandidate struct {
	RuleID         uint64  `json:"rule_id"`
	ItemID         uint64  `json:"item_id"`
	LocationID     uint64  `json:"location_id"`
	OnHand         float64 `json:"on_hand"`
	MinQty         float64 `json:"min_qty"`
	MaxQty         float64 `json:"max_qty"`
	RecommendedQty float64 `json:"recommended_qty"`
}

func (s ReorderService) Candidates(ctx context.Context, organizationID uint64) ([]ReplenishmentCandidate, error) {
	rules, err := s.rules.ListActiveInOrg(ctx, organizationID)
	if err != nil {
		return nil, err
	}

	candidates := make([]ReplenishmentCandidate, 0, len(rules))
	for _, rule := range rules {
		if rule.LocationID == nil {
			continue
		}
		onHand, err := s.ledger.OnHand(ctx, &organizationID, rule.ItemID, *rule.LocationID)
		if err != nil {
			return nil, err
		}
		if onHand >= rule.MinQty {
			continue
		}
		recommended := roundUpToMultiple(rule.MaxQty-onHand, rule.QtyMultiple)
		if recommended <= 0 {
			continue
		}
		candidates = append(candidates, ReplenishmentCandidate{
			RuleID:         rule.ID,
			ItemID:         rule.ItemID,
			LocationID:     *rule.LocationID,
			OnHand:         onHand,
			MinQty:         rule.MinQty,
			MaxQty:         rule.MaxQty,
			RecommendedQty: recommended,
		})
	}
	return candidates, nil
}

func roundUpToMultiple(value, multiple float64) float64 {
	if value <= 0 {
		return 0
	}
	return math.Ceil(value/multiple) * multiple
}
