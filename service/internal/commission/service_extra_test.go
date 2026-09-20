package commission

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestCommissionService_CreatePlan(t *testing.T) {
	tests := []struct {
		name       string
		req        CreatePlanRequest
		wantErr    bool
		wantErrVal error
		wantName   string
		wantActive bool
		wantBasis  string
		wantOrgID  uint64
	}{
		{
			name:       "success",
			req:        CreatePlanRequest{OrganizationID: 10, Name: "Sales Plan", Basis: BasisMargin},
			wantName:   "Sales Plan",
			wantActive: true,
			wantBasis:  BasisMargin,
			wantOrgID:  10,
		},
		{
			name:       "empty name",
			req:        CreatePlanRequest{Name: ""},
			wantErr:    true,
			wantErrVal: ErrPlanNotFound,
		},
		{
			name:       "invalid basis",
			req:        CreatePlanRequest{Name: "X", Basis: "bogus"},
			wantErr:    true,
			wantErrVal: ErrPlanInvalidBasis,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *CommissionPlan
			svc := NewCommissionService(
				CommissionPlanDAOMock{CreateFunc: func(_ context.Context, plan *CommissionPlan) (*CommissionPlan, error) {
					created = plan
					return plan, nil
				}},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			plan, err := svc.CreatePlan(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if plan.Name != tt.wantName || !plan.Active || plan.Basis != tt.wantBasis || plan.OrganizationID == nil || *plan.OrganizationID != tt.wantOrgID {
				t.Errorf("plan = %+v", plan)
			}
			if created == nil || created.Active != true {
				t.Errorf("created = %+v, want default active", created)
			}
		})
	}
}

func TestCommissionService_UpdatePlan(t *testing.T) {
	tests := []struct {
		name       string
		req        UpdatePlanRequest
		wantErr    bool
		wantErrVal error
		wantActive bool
	}{
		{
			name:       "update active to false",
			req:        UpdatePlanRequest{PlanID: 1, Active: helper.Ptr(false)},
			wantActive: false,
		},
		{
			name:       "no active field leaves unchanged",
			req:        UpdatePlanRequest{PlanID: 1},
			wantActive: true,
		},
		{
			name:       "plan not found",
			req:        UpdatePlanRequest{PlanID: 1},
			wantErr:    true,
			wantErrVal: ErrPlanNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updated *CommissionPlan
			svc := NewCommissionService(
				CommissionPlanDAOMock{
					FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
						if tt.name == "plan not found" {
							return nil, nil
						}
						return &CommissionPlan{Base: baseID(1), Active: true}, nil
					},
					UpdateFunc: func(_ context.Context, plan *CommissionPlan) (*CommissionPlan, error) {
						updated = plan
						return plan, nil
					},
				},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			plan, err := svc.UpdatePlan(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if tt.wantActive {
				if !plan.Active {
					t.Error("plan.Active = false, want unchanged true")
				}
			} else {
				if plan.Active {
					t.Error("plan.Active = true, want false")
				}
				if updated == nil || updated.Active {
					t.Error("updated plan not persisted with Active false")
				}
			}
		})
	}
}

func TestCommissionService_CreateRule(t *testing.T) {
	tests := []struct {
		name       string
		req        CreateRuleRequest
		wantErr    bool
		wantErrVal error
		wantPlanID uint64
		wantMin    float64
		wantMax    float64
		wantRate   float64
	}{
		{
			name:       "success",
			req:        CreateRuleRequest{PlanID: 1, MinAmount: 100, MaxAmount: 500, RatePct: 5},
			wantPlanID: 1,
			wantMin:    100,
			wantMax:    500,
			wantRate:   5,
		},
		{
			name:       "bad range",
			req:        CreateRuleRequest{PlanID: 1, MaxAmount: 10, MinAmount: 20, RatePct: 5},
			wantErr:    true,
			wantErrVal: ErrRuleRange,
		},
		{
			name:       "no compensation",
			req:        CreateRuleRequest{PlanID: 1},
			wantErr:    true,
			wantErrVal: ErrRuleInvalid,
		},
		{
			name:       "plan not found",
			req:        CreateRuleRequest{PlanID: 99, RatePct: 5},
			wantErr:    true,
			wantErrVal: ErrPlanNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *CommissionRule
			svc := NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					if tt.name == "plan not found" {
						return nil, ErrPlanNotFound
					}
					return &CommissionPlan{Base: baseID(1), Active: true}, nil
				}},
				CommissionRuleDAOMock{CreateFunc: func(_ context.Context, rule *CommissionRule) (*CommissionRule, error) {
					created = rule
					return rule, nil
				}},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			rule, err := svc.CreateRule(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if rule.PlanID != tt.wantPlanID || rule.MinAmount != tt.wantMin || rule.MaxAmount != tt.wantMax || rule.RatePct != tt.wantRate {
				t.Errorf("rule = %+v", rule)
			}
			if created == nil {
				t.Error("rule not persisted")
			}
		})
	}
}

func TestCommissionService_Assign(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		req        AssignRequest
		listFunc   func(context.Context, *query.Query) (*query.Page[CommissionAssignment], error)
		wantErr    bool
		wantErrVal error
		wantSPID   uint64
	}{
		{
			name: "non overlapping date",
			req:  AssignRequest{PlanID: 1, SalespersonID: 5, DateStart: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{
					{Base: baseID(1), PlanID: 1, SalespersonID: 5, DateStart: &start, DateEnd: &end},
				}}, nil
			},
			wantSPID: 5,
		},
		{
			name: "overlapping date same salesperson",
			req:  AssignRequest{PlanID: 1, SalespersonID: 5, DateStart: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{
					{Base: baseID(1), PlanID: 1, SalespersonID: 5, DateStart: &start, DateEnd: &end},
				}}, nil
			},
			wantErr:    true,
			wantErrVal: ErrAssignmentOverlap,
		},
		{
			name: "overlapping date different salesperson",
			req:  AssignRequest{PlanID: 1, SalespersonID: 9, DateStart: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{
					{Base: baseID(1), PlanID: 1, SalespersonID: 5, DateStart: &start, DateEnd: &end},
				}}, nil
			},
			wantSPID: 9,
		},
		{
			name: "list error",
			req:  AssignRequest{PlanID: 1, SalespersonID: 5, DateStart: time.Now()},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				return nil, errors.New("boom")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var created *CommissionAssignment
			svc := NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					return &CommissionPlan{Base: baseID(1), Active: true}, nil
				}},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{
					ListFunc: tt.listFunc,
					CreateFunc: func(_ context.Context, assignment *CommissionAssignment) (*CommissionAssignment, error) {
						created = assignment
						return assignment, nil
					},
				},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			assignment, err := svc.Assign(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if assignment.SalespersonID != tt.wantSPID || assignment.DateStart == nil {
				t.Errorf("assignment = %+v", assignment)
			}
			if created == nil {
				t.Error("assignment not persisted")
			}
		})
	}
}

func TestPeriodsOverlap(t *testing.T) {
	jan := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	feb := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	mar := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		a    *time.Time
		b    *time.Time
		c    *time.Time
		d    *time.Time
		want bool
	}{
		{
			name: "adjacent periods overlap",
			a:    &jan, b: &feb, c: &feb, d: &mar,
			want: true,
		},
		{
			name: "disjoint periods do not overlap",
			a:    &jan, b: &jan, c: &feb, d: &feb,
			want: false,
		},
		{
			name: "open end overlaps",
			a:    &jan, b: &feb, c: &jan, d: nil,
			want: true,
		},
		{
			name: "nil start overlaps",
			a:    nil, b: nil, c: &jan, d: &feb,
			want: true,
		},
		{
			name: "both open periods overlap",
			a:    &jan, b: nil, c: &feb, d: nil,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := periodsOverlap(tt.a, tt.b, tt.c, tt.d)
			if got != tt.want {
				t.Errorf("periodsOverlap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCommissionService_Accrue_Extra(t *testing.T) {
	tests := []struct {
		name         string
		req          AccrueRequest
		planFunc     func(context.Context, uint64) (*CommissionPlan, error)
		rulesFunc    func(context.Context, uint64) ([]*CommissionRule, error)
		listFunc     func(context.Context, *query.Query) (*query.Page[CommissionAssignment], error)
		createTxFunc func(context.Context, *gorm.DB, *CommissionEntry) (*CommissionEntry, error)
		postTxFunc   func(context.Context, *gorm.DB, accounting.PostRequest) (*accounting.JournalEntry, error)
		txRunFunc    func(context.Context, func(tx *gorm.DB) error) error
		wantErr      bool
		wantErrVal   error
		wantTxCalled bool
	}{
		{
			name: "empty source",
			req:  AccrueRequest{PlanID: 1, ExpenseAccountID: 200, PayableAccountID: 300, JournalID: 3},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true}, nil
			},
			wantErr:    true,
			wantErrVal: ErrSourceMissing,
		},
		{
			name: "no accounts",
			req:  AccrueRequest{PlanID: 1, SalespersonID: 5, SourceType: "invoice", SourceID: 7, BaseAmount: 1000, JournalID: 3, PayableAccountID: 300, Date: time.Now()},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true}, nil
			},
			wantErr:    true,
			wantErrVal: ErrAccountsRequired,
		},
		{
			name: "no assignment",
			req:  AccrueRequest{PlanID: 1, SalespersonID: 5, SourceType: "invoice", SourceID: 7, BaseAmount: 1000, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300, Date: time.Now()},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true}, nil
			},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{}}, nil
			},
			wantErr:    true,
			wantErrVal: ErrAssignmentNotFound,
		},
		{
			name: "plan not found",
			req:  AccrueRequest{PlanID: 99, SalespersonID: 5, SourceType: "invoice", SourceID: 7},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return nil, ErrPlanNotFound
			},
			wantErr:    true,
			wantErrVal: ErrPlanNotFound,
		},
		{
			name: "assignment plan mismatch",
			req:  AccrueRequest{PlanID: 1, SalespersonID: 5, SourceType: "invoice", SourceID: 7, BaseAmount: 1000, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300, Date: time.Now()},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: true, OrganizationID: helper.Ptr(uint64(10))}, nil
			},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{PlanID: 2, SalespersonID: 5, DateStart: &start}}}, nil
			},
			wantErr:    true,
			wantErrVal: ErrAssignmentNotFound,
		},
		{
			name: "no rule match",
			req:  AccrueRequest{PlanID: 1, SalespersonID: 5, SourceType: "invoice", SourceID: 7, BaseAmount: 100, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300, Date: time.Now()},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: true, OrganizationID: helper.Ptr(uint64(10))}, nil
			},
			rulesFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
				return []*CommissionRule{{MinAmount: 1000, RatePct: 5}}, nil
			},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
			},
			wantErr:    true,
			wantErrVal: ErrNoRuleMatches,
		},
		{
			name: "entry create error",
			req:  AccrueRequest{PlanID: 1, SalespersonID: 5, SourceType: "invoice", SourceID: 7, BaseAmount: 1000, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300, Date: time.Now()},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: true, OrganizationID: helper.Ptr(uint64(10))}, nil
			},
			rulesFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
				return []*CommissionRule{{RatePct: 5}}, nil
			},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
			},
			createTxFunc: func(_ context.Context, _ *gorm.DB, _ *CommissionEntry) (*CommissionEntry, error) {
				return nil, errors.New("entry insert failed")
			},
			wantErr: true,
		},
		{
			name: "poster error rolls back",
			req:  AccrueRequest{PlanID: 1, SalespersonID: 5, SourceType: "invoice", SourceID: 7, BaseAmount: 1000, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300, Date: time.Now()},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true}, nil
			},
			rulesFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
				return []*CommissionRule{{Base: baseID(11), RatePct: 5}}, nil
			},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{Base: baseID(21), PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
			},
			createTxFunc: func(_ context.Context, _ *gorm.DB, entry *CommissionEntry) (*CommissionEntry, error) {
				entry.ID = 99
				return entry, nil
			},
			postTxFunc: func(_ context.Context, _ *gorm.DB, _ accounting.PostRequest) (*accounting.JournalEntry, error) {
				return nil, errors.New("post failed")
			},
			txRunFunc: func(_ context.Context, fn func(tx *gorm.DB) error) error {
				return fn(&gorm.DB{})
			},
			wantErr:      true,
			wantTxCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var txInvoked bool
			svc := NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					if tt.planFunc != nil {
						return tt.planFunc(context.Background(), 0)
					}
					return nil, nil
				}},
				CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
					if tt.rulesFunc != nil {
						return tt.rulesFunc(context.Background(), 0)
					}
					return nil, nil
				}},
				CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
					if tt.listFunc != nil {
						return tt.listFunc(context.Background(), nil)
					}
					return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{}}, nil
				}},
				CommissionEntryDAOMock{CreateTxFunc: tt.createTxFunc},
				posterMock{PostTxFunc: tt.postTxFunc},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{RunFunc: func(_ context.Context, fn func(tx *gorm.DB) error) error {
					txInvoked = true
					if tt.txRunFunc != nil {
						return tt.txRunFunc(context.Background(), fn)
					}
					return fn(nil)
				}},
			)

			_, err := svc.Accrue(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantTxCalled && !txInvoked {
				t.Error("expected tx to be invoked")
			}
		})
	}
}

func TestCommissionService_PayExtra(t *testing.T) {
	tests := []struct {
		name       string
		req        PayRequest
		findFunc   func(context.Context, uint64) (*CommissionEntry, error)
		wantErr    bool
		wantErrVal error
		wantState  string
		wantPayID  uint64
	}{
		{
			name: "success",
			req:  PayRequest{EntryID: 1, PayslipID: helper.Ptr(uint64(77))},
			findFunc: func(_ context.Context, _ uint64) (*CommissionEntry, error) {
				return &CommissionEntry{Base: baseID(1), State: EntryStateConfirmed}, nil
			},
			wantState: EntryStatePaid,
			wantPayID: 77,
		},
		{
			name: "entry not found",
			req:  PayRequest{EntryID: 1},
			findFunc: func(_ context.Context, _ uint64) (*CommissionEntry, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrEntryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updated *CommissionEntry
			svc := NewCommissionService(
				CommissionPlanDAOMock{},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{
					FindFunc: tt.findFunc,
					UpdateFunc: func(_ context.Context, entry *CommissionEntry) (*CommissionEntry, error) {
						updated = entry
						return entry, nil
					},
				},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			entry, err := svc.Pay(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if entry.State != tt.wantState || (tt.wantPayID != 0 && (entry.PayslipID == nil || *entry.PayslipID != tt.wantPayID)) {
				t.Errorf("entry = %+v, want state %v with payslip %d", entry, tt.wantState, tt.wantPayID)
			}
			if updated == nil || updated.State != tt.wantState {
				t.Error("entry not persisted")
			}
		})
	}
}

func TestCommissionService_Cancel(t *testing.T) {
	tests := []struct {
		name       string
		entryID    uint64
		findFunc   func(context.Context, uint64) (*CommissionEntry, error)
		wantErr    bool
		wantErrVal error
		wantState  string
	}{
		{
			name:    "success",
			entryID: 1,
			findFunc: func(_ context.Context, _ uint64) (*CommissionEntry, error) {
				return &CommissionEntry{Base: baseID(1), State: EntryStateDraft}, nil
			},
			wantState: EntryStateCancelled,
		},
		{
			name:    "paid entry rejected",
			entryID: 1,
			findFunc: func(_ context.Context, _ uint64) (*CommissionEntry, error) {
				return &CommissionEntry{Base: baseID(1), State: EntryStatePaid}, nil
			},
			wantErr:    true,
			wantErrVal: ErrEntryNotOpen,
		},
		{
			name:    "entry not found",
			entryID: 1,
			findFunc: func(_ context.Context, _ uint64) (*CommissionEntry, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrEntryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var updated *CommissionEntry
			svc := NewCommissionService(
				CommissionPlanDAOMock{},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{
					FindFunc: tt.findFunc,
					UpdateFunc: func(_ context.Context, entry *CommissionEntry) (*CommissionEntry, error) {
						updated = entry
						return entry, nil
					},
				},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			entry, err := svc.Cancel(context.Background(), tt.entryID)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if entry.State != tt.wantState {
				t.Errorf("entry = %+v, want %v", entry, tt.wantState)
			}
			if updated == nil || updated.State != tt.wantState {
				t.Error("entry not persisted")
			}
		})
	}
}

func TestCommissionService_ComputeExtra(t *testing.T) {
	tests := []struct {
		name           string
		req            ComputationRequest
		planFunc       func(context.Context, uint64) (*CommissionPlan, error)
		rulesFunc      func(context.Context, uint64) ([]*CommissionRule, error)
		wantErr        bool
		wantErrVal     error
		wantAmount     float64
		wantCategoryID *uint64
	}{
		{
			name: "skips mismatched category and range",
			req:  ComputationRequest{PlanID: 1, ItemCategoryID: helper.Ptr(uint64(5)), BaseAmount: 500},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: true}, nil
			},
			rulesFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
				return []*CommissionRule{
					{ItemCategoryID: helper.Ptr(uint64(5)), RatePct: 10},
					{ItemCategoryID: helper.Ptr(uint64(6)), RatePct: 20},
					{MaxAmount: 100, RatePct: 30},
					{MinAmount: 1000, RatePct: 40},
				}, nil
			},
			wantAmount:     50,
			wantCategoryID: helper.Ptr(uint64(5)),
		},
		{
			name: "category 6 matches second rule",
			req:  ComputationRequest{PlanID: 1, ItemCategoryID: helper.Ptr(uint64(6)), BaseAmount: 500},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: true}, nil
			},
			rulesFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
				return []*CommissionRule{
					{ItemCategoryID: helper.Ptr(uint64(5)), RatePct: 10},
					{ItemCategoryID: helper.Ptr(uint64(6)), RatePct: 20},
					{MaxAmount: 100, RatePct: 30},
					{MinAmount: 1000, RatePct: 40},
				}, nil
			},
			wantAmount: 100,
		},
		{
			name: "no category uses range only",
			req:  ComputationRequest{PlanID: 1, ItemCategoryID: helper.Ptr(uint64(6)), BaseAmount: 500},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: true}, nil
			},
			rulesFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
				return []*CommissionRule{{ItemCategoryID: helper.Ptr(uint64(5)), RatePct: 10}, {FixedAmount: 250}}, nil
			},
			wantAmount: 250,
		},
		{
			name: "list by plan error",
			req:  ComputationRequest{PlanID: 1, BaseAmount: 500},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: true}, nil
			},
			rulesFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
				return nil, errors.New("rules down")
			},
			wantErr: true,
		},
		{
			name: "plan not found",
			req:  ComputationRequest{PlanID: 99, BaseAmount: 500},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return nil, nil
			},
			wantErr:    true,
			wantErrVal: ErrPlanNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					if tt.planFunc != nil {
						return tt.planFunc(context.Background(), 0)
					}
					return nil, nil
				}},
				CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
					if tt.rulesFunc != nil {
						return tt.rulesFunc(context.Background(), 0)
					}
					return nil, nil
				}},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			computed, err := svc.Compute(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if computed.CommissionAmount != tt.wantAmount {
				t.Errorf("commission amount = %v, want %v", computed.CommissionAmount, tt.wantAmount)
			}
			if tt.wantCategoryID != nil && (computed.ItemCategoryID == nil || *computed.ItemCategoryID != *tt.wantCategoryID) {
				t.Errorf("category = %v, want %v", computed.ItemCategoryID, *tt.wantCategoryID)
			}
		})
	}
}

func TestCommissionService_AccrueFromInvoiceExtra(t *testing.T) {
	tests := []struct {
		name       string
		req        AccrueFromInvoiceRequest
		planFunc   func(context.Context, uint64) (*CommissionPlan, error)
		listFunc   func(context.Context, *query.Query) (*query.Page[CommissionAssignment], error)
		invoiceDAO accounting.InvoiceDAOMock
		wantErr    bool
		wantErrVal error
	}{
		{
			name:       "requires accounts",
			req:        AccrueFromInvoiceRequest{InvoiceID: 7, SalespersonID: 5},
			wantErr:    true,
			wantErrVal: ErrAccountsRequired,
		},
		{
			name: "invoice not found",
			req:  AccrueFromInvoiceRequest{InvoiceID: 7, SalespersonID: 5, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300},
			invoiceDAO: accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
				return nil, nil
			}}},
			wantErr:    true,
			wantErrVal: ErrInvoiceNotFound,
		},
		{
			name: "no assignment",
			req:  AccrueFromInvoiceRequest{InvoiceID: 7, SalespersonID: 5, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{}}, nil
			},
			invoiceDAO: accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
				return &accounting.Invoice{Base: baseID(7), OrganizationID: helper.Ptr(uint64(10))}, nil
			}}},
			wantErr:    true,
			wantErrVal: ErrAssignmentNotFound,
		},
		{
			name: "inactive plan",
			req:  AccrueFromInvoiceRequest{InvoiceID: 7, SalespersonID: 5, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300, Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
			planFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
				return &CommissionPlan{Active: false}, nil
			},
			listFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
				start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
			},
			invoiceDAO: accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
				return &accounting.Invoice{Base: baseID(7), OrganizationID: helper.Ptr(uint64(10))}, nil
			}}},
			wantErr:    true,
			wantErrVal: ErrPlanInactive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					if tt.planFunc != nil {
						return tt.planFunc(context.Background(), 0)
					}
					return nil, nil
				}},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
					if tt.listFunc != nil {
						return tt.listFunc(context.Background(), nil)
					}
					return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{}}, nil
				}},
				CommissionEntryDAOMock{},
				posterMock{},
				tt.invoiceDAO,
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			)

			_, err := svc.AccrueFromInvoice(context.Background(), tt.req)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
		})
	}
}

func TestCommissionService_BaseFromInvoice(t *testing.T) {
	tests := []struct {
		name       string
		svc        CommissionService
		invoice    *accounting.Invoice
		basis      string
		wantErr    bool
		wantErrVal error
		wantBase   float64
	}{
		{
			name: "margin ignores missing variant",
			svc: NewCommissionService(
				CommissionPlanDAOMock{},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
					return []*accounting.InvoiceLine{{ItemID: helper.Ptr(uint64(1)), Qty: 2}, {ItemID: helper.Ptr(uint64(2)), Qty: 1}}, nil
				}},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{ResolveFunc: func(_ context.Context, variantID uint64) (inventory.ResolvedItem, error) {
					if variantID == 1 {
						return inventory.ResolvedItem{}, inventory.ErrVariantNotFound
					}
					return inventory.ResolvedItem{StandardCost: 200}, nil
				}},
				txMock{},
			),
			invoice:  &accounting.Invoice{Base: baseID(7), AmountUntaxed: amount.FromFloat64(1000)},
			basis:    BasisMargin,
			wantBase: 800,
		},
		{
			name: "invalid basis",
			svc: NewCommissionService(
				CommissionPlanDAOMock{},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			),
			invoice:    &accounting.Invoice{Base: baseID(7), AmountUntaxed: amount.FromFloat64(1000)},
			basis:      "bogus",
			wantErr:    true,
			wantErrVal: ErrPlanInvalidBasis,
		},
		{
			name: "margin negative clamps to zero",
			svc: NewCommissionService(
				CommissionPlanDAOMock{},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
					return []*accounting.InvoiceLine{{ItemID: helper.Ptr(uint64(1)), Qty: 10}}, nil
				}},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
					return inventory.ResolvedItem{StandardCost: 200}, nil
				}},
				txMock{},
			),
			invoice:  &accounting.Invoice{Base: baseID(7), AmountUntaxed: amount.FromFloat64(100)},
			basis:    BasisMargin,
			wantBase: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, err := tt.svc.baseFromInvoice(context.Background(), tt.invoice, tt.basis)
			helper.AssertError(t, err, tt.wantErr, tt.wantErrVal)
			if tt.wantErr {
				return
			}
			if base != tt.wantBase {
				t.Errorf("base = %v, want %v", base, tt.wantBase)
			}
		})
	}
}
