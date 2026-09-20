package accounting

import (
	"context"
	"testing"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
)

type reconcileRuleDAOFake struct {
	dao.CRUDMock[ReconcileRule]
	listByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*ReconcileRule, error)
}

func (f reconcileRuleDAOFake) ListByOrganization(ctx context.Context, organizationID uint64) ([]*ReconcileRule, error) {
	if f.listByOrganizationFunc != nil {
		return f.listByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}

func reconcileRuleEngineFixture(rules reconcileRuleDAOFake) ReconcileRuleEngine {
	debit := &JournalLine{Base: model.Base{ID: 11}, AccountID: 1100, ContactID: helper.Ptr(uint64(5)), Debit: amount.FromInt64(100)}
	credit := &JournalLine{Base: model.Base{ID: 22}, AccountID: 1100, ContactID: helper.Ptr(uint64(5)), Credit: amount.FromInt64(100)}
	lines := JournalLineDAOMock{
		ListUnreconciledByAccountFunc: func(_ context.Context, _ uint64) ([]*JournalLine, error) {
			return []*JournalLine{debit, credit}, nil
		},
		CRUDMock: dao.CRUDMock[JournalLine]{
			FindFunc: func(_ context.Context, id uint64) (*JournalLine, error) {
				if id == 11 {
					return debit, nil
				}
				return credit, nil
			},
		},
	}
	return NewReconcileRuleEngine(rules, dao.CRUDMock[ReconcileRuleMatch]{}, lines, AccountPartialReconcileDAOMock{}, nil)
}

func TestReconcileRuleEngine_ApplyMatchesFiltersByMinScore(t *testing.T) {
	rule := &ReconcileRule{Base: model.Base{ID: 3}, Name: helper.Ptr("Contact+Amount"), AccountID: helper.Ptr(uint64(1100)), MatchContact: true, MatchAmount: true, AmountTolerance: 0.01}
	tests := []struct {
		name     string
		minScore int
		want     int
	}{
		{name: "applies candidate above threshold", minScore: 0, want: 1},
		{name: "skips candidate below threshold", minScore: 71, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := reconcileRuleDAOFake{
				listByOrganizationFunc: func(_ context.Context, _ uint64) ([]*ReconcileRule, error) {
					return []*ReconcileRule{rule}, nil
				},
			}
			engine := reconcileRuleEngineFixture(rules)
			candidates, err := engine.FindMatches(context.Background(), 10)
			if err != nil {
				t.Fatal(err)
			}
			if len(candidates) != 1 {
				t.Fatalf("candidates = %d, want 1", len(candidates))
			}

			applied, err := engine.ApplyMatches(context.Background(), candidates, tt.minScore)
			if err != nil {
				t.Fatal(err)
			}
			if applied != tt.want {
				t.Fatalf("applied = %d, want %d", applied, tt.want)
			}
		})
	}
}

func TestReconcileRuleEngine_CreateRule(t *testing.T) {
	var created *ReconcileRule
	rules := reconcileRuleDAOFake{}
	rules.CreateFunc = func(_ context.Context, rule *ReconcileRule) (*ReconcileRule, error) {
		created = rule
		return rule, nil
	}
	engine := reconcileRuleEngineFixture(rules)

	rule := &ReconcileRule{OrganizationID: 10, Name: helper.Ptr("Contact match"), AccountID: helper.Ptr(uint64(1100)), MatchContact: true, Active: true}
	result, err := engine.CreateRule(context.Background(), rule)
	if err != nil {
		t.Fatal(err)
	}
	if result != rule || created != rule {
		t.Fatalf("created = %+v, want %+v", created, rule)
	}
}

func TestReconcileRuleEngine_ListRules(t *testing.T) {
	rules := reconcileRuleDAOFake{}
	rules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[ReconcileRule], error) {
		return &query.Page[ReconcileRule]{Items: []*ReconcileRule{{Base: model.Base{ID: 1}}}, Count: 1}, nil
	}
	engine := reconcileRuleEngineFixture(rules)

	page, err := engine.ListRules(context.Background(), &query.Query{})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Count != 1 {
		t.Fatalf("page = %+v, want 1 item", page)
	}
}
