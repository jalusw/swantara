package commission

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestCommissionService_Compute(t *testing.T) {
	tests := []struct {
		name       string
		svc        CommissionService
		req        ComputationRequest
		wantAmount float64
		wantErr    error
		checkErr   bool
	}{
		{
			name: "applies matching rule",
			svc: NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true, Basis: BasisRevenue}, nil
				}},
				CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
					return []*CommissionRule{{Base: baseID(11), MaxAmount: 1500, RatePct: 5}, {Base: baseID(12), MinAmount: 1000, RatePct: 7}}, nil
				}},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			),
			req:        ComputationRequest{PlanID: 1, BaseAmount: 800},
			wantAmount: 40,
		},
		{
			name: "applies matching rule higher bracket",
			svc: NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true, Basis: BasisRevenue}, nil
				}},
				CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
					return []*CommissionRule{{Base: baseID(11), MaxAmount: 1500, RatePct: 5}, {Base: baseID(12), MinAmount: 1000, RatePct: 7}}, nil
				}},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			),
			req:        ComputationRequest{PlanID: 1, BaseAmount: 2000},
			wantAmount: 140,
		},
		{
			name: "no rule matches",
			svc: NewCommissionService(
				CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
					return &CommissionPlan{Active: true}, nil
				}},
				CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
					return []*CommissionRule{{MaxAmount: 100}}, nil
				}},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			),
			req:      ComputationRequest{PlanID: 1, BaseAmount: 200},
			wantErr:  ErrNoRuleMatches,
			checkErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			computed, err := tt.svc.Compute(context.Background(), tt.req)
			if tt.checkErr {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if computed.CommissionAmount != tt.wantAmount {
				t.Errorf("computed = %+v, want %v", computed, tt.wantAmount)
			}
		})
	}
}

func TestCommissionService_Accrue(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(*accounting.PostRequest) CommissionService
		req             AccrueRequest
		wantState       string
		wantCommission  float64
		wantPostedLines int
		wantExpenseDr   amount.Amount
		wantExpenseAcct uint64
		wantPayableCr   amount.Amount
		wantPayableAcct uint64
		wantErr         error
		checkErr        bool
	}{
		{
			name: "posts expense and payable",
			setup: func(posted *accounting.PostRequest) CommissionService {
				return NewCommissionService(
					CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
						return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true, Basis: BasisRevenue}, nil
					}},
					CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
						return []*CommissionRule{{Base: baseID(11), RatePct: 5}}, nil
					}},
					CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
						start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
						return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{Base: baseID(21), PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
					}},
					CommissionEntryDAOMock{CreateTxFunc: func(_ context.Context, _ *gorm.DB, entry *CommissionEntry) (*CommissionEntry, error) {
						entry.ID = 99
						return entry, nil
					}},
					posterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
						*posted = request
						return &accounting.JournalEntry{Base: baseID(900)}, nil
					}},
					accounting.InvoiceDAOMock{},
					accounting.InvoiceLineDAOMock{},
					accounting.PaymentAllocationDAOMock{},
					inventory.ItemResolverMock{},
					txMock{},
				)
			},
			req: AccrueRequest{
				SalespersonID:    5,
				PlanID:           1,
				SourceType:       "invoice",
				SourceID:         7,
				BaseAmount:       1000,
				JournalID:        3,
				ExpenseAccountID: 200,
				PayableAccountID: 300,
				Date:             time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			wantState:       EntryStateConfirmed,
			wantCommission:  50,
			wantPostedLines: 2,
			wantExpenseDr:   amount.FromFloat64(50),
			wantExpenseAcct: 200,
			wantPayableCr:   amount.FromFloat64(50),
			wantPayableAcct: 300,
		},
		{
			name: "inactive plan rejected",
			setup: func(_ *accounting.PostRequest) CommissionService {
				return NewCommissionService(
					CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
						return &CommissionPlan{Active: false}, nil
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
			},
			req: AccrueRequest{
				PlanID: 1, SalespersonID: 5, SourceType: "invoice", SourceID: 7,
				BaseAmount: 1000, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300,
				Date: time.Now(),
			},
			wantErr:  ErrPlanInactive,
			checkErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var posted accounting.PostRequest
			svc := tt.setup(&posted)

			entry, err := svc.Accrue(context.Background(), tt.req)
			if tt.checkErr {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.State != tt.wantState || entry.CommissionAmount != tt.wantCommission {
				t.Errorf("entry = %+v, want state %s commission %v", entry, tt.wantState, tt.wantCommission)
			}
			if len(posted.Lines) != tt.wantPostedLines {
				t.Fatalf("posted = %+v, want %d lines", posted.Lines, tt.wantPostedLines)
			}
			if !posted.Lines[0].Debit.Equal(tt.wantExpenseDr) || posted.Lines[0].AccountID != tt.wantExpenseAcct {
				t.Errorf("expense line = %+v, want Dr %v on account %d", posted.Lines[0], tt.wantExpenseDr, tt.wantExpenseAcct)
			}
			if !posted.Lines[1].Credit.Equal(tt.wantPayableCr) || posted.Lines[1].AccountID != tt.wantPayableAcct {
				t.Errorf("payable line = %+v, want Cr %v on account %d", posted.Lines[1], tt.wantPayableCr, tt.wantPayableAcct)
			}
		})
	}
}

func TestCommissionService_AccrueFromInvoice(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(*accounting.PostRequest) CommissionService
		req             AccrueFromInvoiceRequest
		wantState       string
		wantBase        float64
		wantCommission  float64
		wantExpenseDr   amount.Amount
		wantExpenseAcct uint64
		wantErr         error
		checkErr        bool
	}{
		{
			name: "revenue basis uses invoice untaxed",
			setup: func(posted *accounting.PostRequest) CommissionService {
				return NewCommissionService(
					CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
						return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true, Basis: BasisRevenue}, nil
					}},
					CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
						return []*CommissionRule{{Base: baseID(11), RatePct: 5}}, nil
					}},
					CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
						start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
						return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{Base: baseID(21), PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
					}},
					CommissionEntryDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionEntry], error) {
						return &query.Page[CommissionEntry]{Items: []*CommissionEntry{}}, nil
					}, CreateTxFunc: func(_ context.Context, _ *gorm.DB, entry *CommissionEntry) (*CommissionEntry, error) {
						entry.ID = 99
						return entry, nil
					}},
					posterMock{PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
						*posted = request
						return &accounting.JournalEntry{Base: baseID(900)}, nil
					}},
					accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: baseID(7), OrganizationID: helper.Ptr(uint64(10)), AmountUntaxed: amount.FromInt64(1000)}, nil
					}}},
					accounting.InvoiceLineDAOMock{},
					accounting.PaymentAllocationDAOMock{},
					inventory.ItemResolverMock{},
					txMock{},
				)
			},
			req: AccrueFromInvoiceRequest{
				InvoiceID: 7, SalespersonID: 5, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300,
				Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			wantState:       EntryStateConfirmed,
			wantBase:        1000,
			wantCommission:  50,
			wantExpenseDr:   amount.FromFloat64(50),
			wantExpenseAcct: 200,
		},
		{
			name: "margin basis deducts standard cost",
			setup: func(_ *accounting.PostRequest) CommissionService {
				return NewCommissionService(
					CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
						return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true, Basis: BasisMargin}, nil
					}},
					CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
						return []*CommissionRule{{Base: baseID(11), RatePct: 10}}, nil
					}},
					CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
						start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
						return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{Base: baseID(21), PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
					}},
					CommissionEntryDAOMock{CreateTxFunc: func(_ context.Context, _ *gorm.DB, entry *CommissionEntry) (*CommissionEntry, error) {
						entry.ID = 99
						return entry, nil
					}},
					posterMock{},
					accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: baseID(7), OrganizationID: helper.Ptr(uint64(10)), AmountUntaxed: amount.FromInt64(1000)}, nil
					}}},
					accounting.InvoiceLineDAOMock{ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*accounting.InvoiceLine, error) {
						return []*accounting.InvoiceLine{
							{ItemID: helper.Ptr(uint64(1)), Qty: 2},
							{ItemID: helper.Ptr(uint64(2)), Qty: 1},
							{ItemID: nil, Qty: 3},
						}, nil
					}},
					accounting.PaymentAllocationDAOMock{},
					inventory.ItemResolverMock{ResolveFunc: func(_ context.Context, variantID uint64) (inventory.ResolvedItem, error) {
						if variantID == 1 {
							return inventory.ResolvedItem{StandardCost: 100}, nil
						}
						return inventory.ResolvedItem{StandardCost: 50}, nil
					}},
					txMock{},
				)
			},
			req: AccrueFromInvoiceRequest{
				InvoiceID: 7, SalespersonID: 5, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300,
				Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			wantState:      EntryStateConfirmed,
			wantBase:       750,
			wantCommission: 75,
		},
		{
			name: "collected basis sums allocations",
			setup: func(_ *accounting.PostRequest) CommissionService {
				return NewCommissionService(
					CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
						return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true, Basis: BasisCollected}, nil
					}},
					CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
						return []*CommissionRule{{Base: baseID(11), RatePct: 5}}, nil
					}},
					CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
						start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
						return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{Base: baseID(21), PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
					}},
					CommissionEntryDAOMock{CreateTxFunc: func(_ context.Context, _ *gorm.DB, entry *CommissionEntry) (*CommissionEntry, error) {
						entry.ID = 99
						return entry, nil
					}},
					posterMock{},
					accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: baseID(7), OrganizationID: helper.Ptr(uint64(10)), AmountUntaxed: amount.FromInt64(1000)}, nil
					}}},
					accounting.InvoiceLineDAOMock{},
					accounting.PaymentAllocationDAOMock{ListByInvoiceFunc: func(_ context.Context, _ uint64) ([]*accounting.PaymentAllocation, error) {
						return []*accounting.PaymentAllocation{
							{Amount: 300},
							{Amount: 200},
						}, nil
					}},
					inventory.ItemResolverMock{},
					txMock{},
				)
			},
			req: AccrueFromInvoiceRequest{
				InvoiceID: 7, SalespersonID: 5, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300,
				Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			wantState:      EntryStateConfirmed,
			wantBase:       500,
			wantCommission: 25,
		},
		{
			name: "duplicate entry rejected",
			setup: func(_ *accounting.PostRequest) CommissionService {
				return NewCommissionService(
					CommissionPlanDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionPlan, error) {
						return &CommissionPlan{Base: baseID(1), OrganizationID: helper.Ptr(uint64(10)), Active: true, Basis: BasisRevenue}, nil
					}},
					CommissionRuleDAOMock{ListByPlanFunc: func(_ context.Context, _ uint64) ([]*CommissionRule, error) {
						return []*CommissionRule{{Base: baseID(11), RatePct: 5}}, nil
					}},
					CommissionAssignmentDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionAssignment], error) {
						start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
						return &query.Page[CommissionAssignment]{Items: []*CommissionAssignment{{Base: baseID(21), PlanID: 1, SalespersonID: 5, DateStart: &start}}}, nil
					}},
					CommissionEntryDAOMock{ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[CommissionEntry], error) {
						return &query.Page[CommissionEntry]{Items: []*CommissionEntry{{SourceType: "invoice", SourceID: 7, PlanID: 1, SalespersonID: 5, State: EntryStateConfirmed}}}, nil
					}},
					posterMock{},
					accounting.InvoiceDAOMock{CRUDMock: dao.CRUDMock[accounting.Invoice]{FindFunc: func(_ context.Context, _ uint64) (*accounting.Invoice, error) {
						return &accounting.Invoice{Base: baseID(7), OrganizationID: helper.Ptr(uint64(10)), AmountUntaxed: amount.FromInt64(1000)}, nil
					}}},
					accounting.InvoiceLineDAOMock{},
					accounting.PaymentAllocationDAOMock{},
					inventory.ItemResolverMock{},
					txMock{},
				)
			},
			req: AccrueFromInvoiceRequest{
				InvoiceID: 7, SalespersonID: 5, JournalID: 3, ExpenseAccountID: 200, PayableAccountID: 300,
				Date: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr:  ErrEntryDuplicate,
			checkErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var posted accounting.PostRequest
			svc := tt.setup(&posted)

			entry, err := svc.AccrueFromInvoice(context.Background(), tt.req)
			if tt.checkErr {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if entry.State != tt.wantState || entry.BaseAmount != tt.wantBase || entry.CommissionAmount != tt.wantCommission {
				t.Errorf("entry = %+v, want state %s base %v commission %v", entry, tt.wantState, tt.wantBase, tt.wantCommission)
			}
			if len(posted.Lines) > 0 {
				if !posted.Lines[0].Debit.Equal(tt.wantExpenseDr) || posted.Lines[0].AccountID != tt.wantExpenseAcct {
					t.Errorf("expense line = %+v, want Dr %v on account %d", posted.Lines[0], tt.wantExpenseDr, tt.wantExpenseAcct)
				}
			}
		})
	}
}

func TestCommissionService_Pay(t *testing.T) {
	tests := []struct {
		name    string
		svc     CommissionService
		req     PayRequest
		wantErr error
	}{
		{
			name: "requires confirmed entry",
			svc: NewCommissionService(
				CommissionPlanDAOMock{},
				CommissionRuleDAOMock{},
				CommissionAssignmentDAOMock{},
				CommissionEntryDAOMock{FindFunc: func(_ context.Context, _ uint64) (*CommissionEntry, error) {
					return &CommissionEntry{State: EntryStateDraft}, nil
				}},
				posterMock{},
				accounting.InvoiceDAOMock{},
				accounting.InvoiceLineDAOMock{},
				accounting.PaymentAllocationDAOMock{},
				inventory.ItemResolverMock{},
				txMock{},
			),
			req:     PayRequest{EntryID: 1},
			wantErr: ErrEntryNotConfirmed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.svc.Pay(context.Background(), tt.req)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

type posterMock struct {
	PostFunc      func(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error)
	PostTxFunc    func(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error)
	ReverseFunc   func(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
	ReverseTxFunc func(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error)
}

func (m posterMock) Post(ctx context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostFunc != nil {
		return m.PostFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m posterMock) PostTx(ctx context.Context, tx *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
	if m.PostTxFunc != nil {
		return m.PostTxFunc(ctx, tx, request)
	}
	return m.Post(ctx, request)
}

func (m posterMock) Reverse(ctx context.Context, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseFunc != nil {
		return m.ReverseFunc(ctx, request)
	}
	return &accounting.JournalEntry{}, nil
}

func (m posterMock) ReverseTx(ctx context.Context, tx *gorm.DB, request accounting.ReverseRequest) (*accounting.JournalEntry, error) {
	if m.ReverseTxFunc != nil {
		return m.ReverseTxFunc(ctx, tx, request)
	}
	return m.Reverse(ctx, request)
}

type txMock struct {
	RunFunc func(ctx context.Context, fn func(tx *gorm.DB) error) error
}

func (m txMock) Run(ctx context.Context, fn func(tx *gorm.DB) error) error {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, fn)
	}
	return fn(nil)
}

func baseID(id uint64) model.Base { return model.Base{ID: id} }
