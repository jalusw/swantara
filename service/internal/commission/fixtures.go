package commission

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func CommissionPlanFixture(opts ...func(*CommissionPlan) *CommissionPlan) *CommissionPlan {
	now := time.Now()
	plan := &CommissionPlan{
		Base:           model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		OrganizationID: helper.Ptr(uint64(gofakeit.Number(1, 10000))),
		Name:           gofakeit.AppName(),
		Basis:          BasisRevenue,
		Active:         true,
	}
	for _, opt := range opts {
		opt(plan)
	}
	return plan
}

func CommissionRuleFixture(opts ...func(*CommissionRule) *CommissionRule) *CommissionRule {
	now := time.Now()
	rule := &CommissionRule{
		Base:      model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		PlanID:    uint64(gofakeit.Number(1, 10000)),
		MinAmount: gofakeit.Float64Range(1, 5000),
		MaxAmount: gofakeit.Float64Range(5000, 10000),
		RatePct:   gofakeit.Float64Range(1, 50),
	}
	for _, opt := range opts {
		opt(rule)
	}
	return rule
}

func CommissionAssignmentFixture(opts ...func(*CommissionAssignment) *CommissionAssignment) *CommissionAssignment {
	now := time.Now()
	assignment := &CommissionAssignment{
		Base:          model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		PlanID:        uint64(gofakeit.Number(1, 10000)),
		SalespersonID: uint64(gofakeit.Number(1, 10000)),
		DateStart:     helper.Ptr(time.Now()),
	}
	for _, opt := range opts {
		opt(assignment)
	}
	return assignment
}

func CommissionEntryFixture(opts ...func(*CommissionEntry) *CommissionEntry) *CommissionEntry {
	now := time.Now()
	entry := &CommissionEntry{
		Base:             model.Base{ID: uint64(gofakeit.Number(1, 10000)), CreatedAt: now, UpdatedAt: now},
		SalespersonID:    uint64(gofakeit.Number(1, 10000)),
		PlanID:           uint64(gofakeit.Number(1, 10000)),
		SourceType:       "journal_entry",
		SourceID:         uint64(gofakeit.Number(1, 10000)),
		BaseAmount:       gofakeit.Float64Range(1, 10000),
		CommissionAmount: gofakeit.Float64Range(1, 10000),
		State:            EntryStateDraft,
	}
	for _, opt := range opts {
		opt(entry)
	}
	return entry
}
