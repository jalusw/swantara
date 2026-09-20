package commission

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/state"
	"gorm.io/gorm"
)

type CommissionService struct {
	plans        CommissionPlanDAO
	rules        CommissionRuleDAO
	assigns      CommissionAssignmentDAO
	entries      CommissionEntryDAO
	poster       accounting.Poster
	tx           db.Transactioner
	invoices     accounting.InvoiceDAO
	invoiceLines accounting.InvoiceLineDAO
	allocations  accounting.PaymentAllocationDAO
	resolver     inventory.ItemResolver
	now          func() time.Time
	entryState   state.Machine
}

func NewCommissionService(
	plans CommissionPlanDAO,
	rules CommissionRuleDAO,
	assigns CommissionAssignmentDAO,
	entries CommissionEntryDAO,
	poster accounting.Poster,
	invoices accounting.InvoiceDAO,
	invoiceLines accounting.InvoiceLineDAO,
	allocations accounting.PaymentAllocationDAO,
	resolver inventory.ItemResolver,
	tx db.Transactioner,
) CommissionService {
	return CommissionService{
		plans:        plans,
		rules:        rules,
		assigns:      assigns,
		entries:      entries,
		poster:       poster,
		tx:           tx,
		invoices:     invoices,
		invoiceLines: invoiceLines,
		allocations:  allocations,
		resolver:     resolver,
		now:          time.Now,
		entryState: state.NewMachine(
			state.Transition{From: model.Status(EntryStateDraft), To: model.Status(EntryStateConfirmed)},
			state.Transition{From: model.Status(EntryStateConfirmed), To: model.Status(EntryStatePaid)},
			state.Transition{From: model.Status(EntryStateDraft), To: model.Status(EntryStateCancelled)},
			state.Transition{From: model.Status(EntryStateConfirmed), To: model.Status(EntryStateCancelled)},
		),
	}
}

func (s CommissionService) plan(ctx context.Context, id uint64) (*CommissionPlan, error) {
	plan, err := s.plans.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, ErrPlanNotFound
	}
	return plan, nil
}

func validateBasis(basis string) error {
	switch basis {
	case BasisRevenue, BasisMargin, BasisCollected:
		return nil
	}
	return ErrPlanInvalidBasis
}

type CreatePlanRequest struct {
	OrganizationID uint64
	Name           string
	Basis          string
	Active         bool
}

func (s CommissionService) ListPlans(ctx context.Context, q *query.Query) (*query.Page[CommissionPlan], error) {
	return s.plans.List(ctx, q)
}

func (s CommissionService) FindPlan(ctx context.Context, id uint64) (*CommissionPlan, error) {
	return s.plans.Find(ctx, id)
}

func (s CommissionService) ListRulesByPlan(ctx context.Context, planID uint64) ([]*CommissionRule, error) {
	return s.rules.ListByPlan(ctx, planID)
}

func (s CommissionService) ListAssignments(ctx context.Context, q *query.Query) (*query.Page[CommissionAssignment], error) {
	return s.assigns.List(ctx, q)
}

func (s CommissionService) ListEntries(ctx context.Context, q *query.Query) (*query.Page[CommissionEntry], error) {
	return s.entries.List(ctx, q)
}

func (s CommissionService) CreatePlan(ctx context.Context, request CreatePlanRequest) (*CommissionPlan, error) {
	if request.Name == "" {
		return nil, ErrPlanNotFound
	}
	if err := validateBasis(request.Basis); err != nil {
		return nil, err
	}
	active := request.Active
	if !request.Active {
		active = true
	}
	return s.plans.Create(ctx, &CommissionPlan{
		OrganizationID: &request.OrganizationID,
		Name:           request.Name,
		Basis:          request.Basis,
		Active:         active,
	})
}

type UpdatePlanRequest struct {
	PlanID uint64
	Active *bool
}

func (s CommissionService) UpdatePlan(ctx context.Context, request UpdatePlanRequest) (*CommissionPlan, error) {
	plan, err := s.plan(ctx, request.PlanID)
	if err != nil {
		return nil, err
	}
	if request.Active != nil {
		plan.Active = *request.Active
	}
	return s.plans.Update(ctx, plan)
}

type CreateRuleRequest struct {
	PlanID         uint64
	ItemCategoryID *uint64
	MinAmount      float64
	MaxAmount      float64
	RatePct        float64
	FixedAmount    float64
}

func (s CommissionService) CreateRule(ctx context.Context, request CreateRuleRequest) (*CommissionRule, error) {
	if _, err := s.plan(ctx, request.PlanID); err != nil {
		return nil, err
	}
	if request.MinAmount < 0 || request.MaxAmount < 0 || (request.MaxAmount > 0 && request.MaxAmount < request.MinAmount) {
		return nil, ErrRuleRange
	}
	if request.RatePct == 0 && request.FixedAmount == 0 {
		return nil, ErrRuleInvalid
	}
	return s.rules.Create(ctx, &CommissionRule{
		PlanID:         request.PlanID,
		ItemCategoryID: request.ItemCategoryID,
		MinAmount:      request.MinAmount,
		MaxAmount:      request.MaxAmount,
		RatePct:        request.RatePct,
		FixedAmount:    request.FixedAmount,
	})
}

type AssignRequest struct {
	PlanID        uint64
	SalespersonID uint64
	DateStart     time.Time
	DateEnd       *time.Time
}

func (s CommissionService) Assign(ctx context.Context, request AssignRequest) (*CommissionAssignment, error) {
	if _, err := s.plan(ctx, request.PlanID); err != nil {
		return nil, err
	}
	overlapPage, err := s.assigns.List(ctx, &query.Query{Filters: []query.Filter{{Field: "plan_id", Operator: query.Equal, Value: request.PlanID}}})
	if err != nil {
		return nil, err
	}
	start := request.DateStart
	end := request.DateEnd
	for _, existing := range overlapPage.Items {
		if existing.SalespersonID != request.SalespersonID {
			continue
		}
		if periodsOverlap(existing.DateStart, existing.DateEnd, &start, end) {
			return nil, ErrAssignmentOverlap
		}
	}
	return s.assigns.Create(ctx, &CommissionAssignment{
		PlanID:        request.PlanID,
		SalespersonID: request.SalespersonID,
		DateStart:     &start,
		DateEnd:       end,
	})
}

func periodsOverlap(aStart, aEnd, bStart, bEnd *time.Time) bool {
	if aStart == nil || bStart == nil {
		return true
	}
	if aEnd == nil && bEnd == nil {
		return true
	}
	leftEnd := aEnd
	if leftEnd == nil {
		leftEnd = bStart
	}
	rightEnd := bEnd
	if rightEnd == nil {
		rightEnd = aStart
	}
	if leftEnd.Before(*bStart) || rightEnd.Before(*aStart) {
		return false
	}
	return true
}

type ComputationRequest struct {
	PlanID         uint64
	ItemCategoryID *uint64
	BaseAmount     float64
}

type Computation struct {
	RuleID           uint64
	ItemCategoryID   *uint64
	CommissionAmount float64
}

func (s CommissionService) Compute(ctx context.Context, request ComputationRequest) (Computation, error) {
	if _, err := s.plan(ctx, request.PlanID); err != nil {
		return Computation{}, err
	}
	rules, err := s.rules.ListByPlan(ctx, request.PlanID)
	if err != nil {
		return Computation{}, err
	}
	base := amount.FromFloat64(request.BaseAmount)
	for _, rule := range rules {
		if request.ItemCategoryID != nil && rule.ItemCategoryID != nil && *request.ItemCategoryID != *rule.ItemCategoryID {
			continue
		}
		if rule.MinAmount > 0 && base.LessThan(amount.FromFloat64(rule.MinAmount)) {
			continue
		}
		if rule.MaxAmount > 0 && base.GreaterThan(amount.FromFloat64(rule.MaxAmount)) {
			continue
		}
		computed := ruleCommission(rule, base)
		return Computation{RuleID: rule.ID, ItemCategoryID: rule.ItemCategoryID, CommissionAmount: computed.Float64()}, nil
	}
	return Computation{}, ErrNoRuleMatches
}

func ruleCommission(rule *CommissionRule, base amount.Amount) amount.Amount {
	if rule.FixedAmount > 0 {
		return amount.FromFloat64(rule.FixedAmount)
	}
	computed, err := base.Mul(amount.FromFloat64(rule.RatePct)).Div(amount.FromInt64(100))
	if err != nil {
		return amount.Zero()
	}
	return computed.Round(4)
}

type AccrueRequest struct {
	SalespersonID    uint64
	PlanID           uint64
	SourceType       string
	SourceID         uint64
	ItemCategoryID   *uint64
	BaseAmount       float64
	JournalID        uint64
	ExpenseAccountID uint64
	PayableAccountID uint64
	Date             time.Time
	PeriodID         *uint64
}

func (s CommissionService) Accrue(ctx context.Context, request AccrueRequest) (*CommissionEntry, error) {
	plan, err := s.plan(ctx, request.PlanID)
	if err != nil {
		return nil, err
	}
	if !plan.Active {
		return nil, ErrPlanInactive
	}
	if request.SourceType == "" || request.SourceID == 0 {
		return nil, ErrSourceMissing
	}
	if request.ExpenseAccountID == 0 || request.PayableAccountID == 0 {
		return nil, ErrAccountsRequired
	}
	if _, err := s.activeAssignment(ctx, request.PlanID, request.SalespersonID, request.Date); err != nil {
		return nil, err
	}
	if err := s.ensureNoDuplicate(ctx, request.PlanID, request.SalespersonID, request.SourceType, request.SourceID); err != nil {
		return nil, err
	}
	computed, err := s.Compute(ctx, ComputationRequest{PlanID: request.PlanID, ItemCategoryID: request.ItemCategoryID, BaseAmount: request.BaseAmount})
	if err != nil {
		return nil, err
	}
	if computed.CommissionAmount <= 0 {
		return nil, ErrNoRuleMatches
	}

	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	return s.accrue(ctx, accrueParams{
		organizationID: *plan.OrganizationID,
		salespersonID:  request.SalespersonID,
		planID:         plan.ID,
		planName:       plan.Name,
		sourceType:     request.SourceType,
		sourceID:       request.SourceID,
		baseAmount:     request.BaseAmount,
		commission:     computed.CommissionAmount,
		journalID:      request.JournalID,
		expenseAccount: request.ExpenseAccountID,
		payableAccount: request.PayableAccountID,
		date:           date,
		periodID:       request.PeriodID,
	})
}

type AccrueFromInvoiceRequest struct {
	InvoiceID        uint64
	SalespersonID    uint64
	JournalID        uint64
	ExpenseAccountID uint64
	PayableAccountID uint64
	Date             time.Time
	PeriodID         *uint64
}

func (s CommissionService) AccrueFromInvoice(ctx context.Context, request AccrueFromInvoiceRequest) (*CommissionEntry, error) {
	if request.JournalID == 0 || request.ExpenseAccountID == 0 || request.PayableAccountID == 0 {
		return nil, ErrAccountsRequired
	}
	invoice, err := s.invoices.Find(ctx, request.InvoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil || invoice.OrganizationID == nil {
		return nil, ErrInvoiceNotFound
	}
	date := request.Date
	if date.IsZero() {
		date = s.now().UTC()
	}
	assign, err := s.assignmentForSalesperson(ctx, request.SalespersonID, date)
	if err != nil {
		return nil, err
	}
	plan, err := s.plan(ctx, assign.PlanID)
	if err != nil {
		return nil, err
	}
	if !plan.Active {
		return nil, ErrPlanInactive
	}
	if err := s.ensureNoDuplicate(ctx, plan.ID, request.SalespersonID, "invoice", request.InvoiceID); err != nil {
		return nil, err
	}
	base, err := s.baseFromInvoice(ctx, invoice, plan.Basis)
	if err != nil {
		return nil, err
	}
	computed, err := s.Compute(ctx, ComputationRequest{PlanID: plan.ID, BaseAmount: base})
	if err != nil {
		return nil, err
	}
	if computed.CommissionAmount <= 0 {
		return nil, ErrNoRuleMatches
	}
	return s.accrue(ctx, accrueParams{
		organizationID: *invoice.OrganizationID,
		salespersonID:  request.SalespersonID,
		planID:         plan.ID,
		planName:       plan.Name,
		sourceType:     "invoice",
		sourceID:       request.InvoiceID,
		baseAmount:     base,
		commission:     computed.CommissionAmount,
		journalID:      request.JournalID,
		expenseAccount: request.ExpenseAccountID,
		payableAccount: request.PayableAccountID,
		date:           date,
		periodID:       request.PeriodID,
	})
}

func (s CommissionService) baseFromInvoice(ctx context.Context, invoice *accounting.Invoice, basis string) (float64, error) {
	switch basis {
	case BasisRevenue:
		return invoice.AmountUntaxed.Float64(), nil
	case BasisMargin:
		lines, err := s.invoiceLines.ListByInvoice(ctx, invoice.ID)
		if err != nil {
			return 0, err
		}
		cost := amount.Zero()
		for _, line := range lines {
			if line.ItemID == nil {
				continue
			}
			resolved, err := s.resolver.Resolve(ctx, *line.ItemID)
			if err != nil {
				if errors.Is(err, inventory.ErrVariantNotFound) {
					continue
				}
				return 0, err
			}
			cost = cost.Add(amount.FromFloat64(line.Qty).Mul(amount.FromFloat64(resolved.StandardCost)).Round(4))
		}
		margin := invoice.AmountUntaxed.Sub(cost)
		if margin.IsNegative() {
			margin = amount.Zero()
		}
		return margin.Float64(), nil
	case BasisCollected:
		allocations, err := s.allocations.ListByInvoice(ctx, invoice.ID)
		if err != nil {
			return 0, err
		}
		collected := amount.Zero()
		for _, allocation := range allocations {
			collected = collected.Add(amount.FromFloat64(allocation.Amount))
		}
		return collected.Float64(), nil
	}
	return 0, ErrPlanInvalidBasis
}

func (s CommissionService) ensureNoDuplicate(ctx context.Context, planID, salespersonID uint64, sourceType string, sourceID uint64) error {
	existing, err := s.entries.List(ctx, &query.Query{Filters: []query.Filter{{Field: "source_id", Operator: query.Equal, Value: sourceID}}})
	if err != nil {
		return err
	}
	for _, entry := range existing.Items {
		if entry.SourceType == sourceType && entry.PlanID == planID && entry.SalespersonID == salespersonID && entry.State != EntryStateCancelled {
			return ErrEntryDuplicate
		}
	}
	return nil
}

type accrueParams struct {
	organizationID uint64
	salespersonID  uint64
	planID         uint64
	planName       string
	sourceType     string
	sourceID       uint64
	baseAmount     float64
	commission     float64
	journalID      uint64
	expenseAccount uint64
	payableAccount uint64
	date           time.Time
	periodID       *uint64
}

func (s CommissionService) accrue(ctx context.Context, p accrueParams) (*CommissionEntry, error) {
	commission := amount.FromFloat64(p.commission)

	var entry *CommissionEntry
	err := s.tx.Run(ctx, func(tx *gorm.DB) error {
		_, err := s.poster.PostTx(ctx, tx, accounting.PostRequest{
			OrganizationID: p.organizationID,
			JournalID:      p.journalID,
			Date:           p.date,
			Ref:            fmt.Sprintf("COMM/%d", p.sourceID),
			OriginType:     accounting.OriginTypeCommission,
			OriginID:       p.sourceID,
			Description:    fmt.Sprintf("Commission for %s", p.planName),
			Lines: []accounting.PostingLine{
				{AccountID: p.expenseAccount, Name: "Commission Expense", Debit: commission},
				{AccountID: p.payableAccount, Name: "Commission Payable", Credit: commission},
			},
		})
		if err != nil {
			return err
		}
		created, err := s.entries.CreateTx(ctx, tx, &CommissionEntry{
			SalespersonID:    p.salespersonID,
			PlanID:           p.planID,
			SourceType:       p.sourceType,
			SourceID:         p.sourceID,
			BaseAmount:       p.baseAmount,
			CommissionAmount: commission.Float64(),
			State:            EntryStateConfirmed,
			PeriodID:         p.periodID,
		})
		if err != nil {
			return err
		}
		entry = created
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (s CommissionService) assignmentForSalesperson(ctx context.Context, salespersonID uint64, date time.Time) (*CommissionAssignment, error) {
	assigns, err := s.assigns.List(ctx, &query.Query{Filters: []query.Filter{{Field: "salesperson_id", Operator: query.Equal, Value: salespersonID}}})
	if err != nil {
		return nil, err
	}
	for _, a := range assigns.Items {
		if a.DateStart != nil && date.Before(*a.DateStart) {
			continue
		}
		if a.DateEnd != nil && date.After(*a.DateEnd) {
			continue
		}
		return a, nil
	}
	return nil, ErrAssignmentNotFound
}

func (s CommissionService) activeAssignment(ctx context.Context, planID, salespersonID uint64, date time.Time) (*CommissionAssignment, error) {
	assign, err := s.assignmentForSalesperson(ctx, salespersonID, date)
	if err != nil {
		return nil, err
	}
	if assign.PlanID != planID {
		return nil, ErrAssignmentNotFound
	}
	return assign, nil
}

type PayRequest struct {
	EntryID   uint64
	PayslipID *uint64
}

func (s CommissionService) Pay(ctx context.Context, request PayRequest) (*CommissionEntry, error) {
	entry, err := s.entry(ctx, request.EntryID)
	if err != nil {
		return nil, err
	}
	if err := s.entryState.TryTransition(model.Status(entry.State), model.Status(EntryStatePaid)); err != nil {
		return nil, ErrEntryNotConfirmed
	}
	entry.State = EntryStatePaid
	entry.PayslipID = request.PayslipID
	return s.entries.Update(ctx, entry)
}

func (s CommissionService) Cancel(ctx context.Context, entryID uint64) (*CommissionEntry, error) {
	entry, err := s.entry(ctx, entryID)
	if err != nil {
		return nil, err
	}
	if err := s.entryState.TryTransition(model.Status(entry.State), model.Status(EntryStateCancelled)); err != nil {
		return nil, ErrEntryNotOpen
	}
	entry.State = EntryStateCancelled
	return s.entries.Update(ctx, entry)
}

func (s CommissionService) entry(ctx context.Context, id uint64) (*CommissionEntry, error) {
	entry, err := s.entries.Find(ctx, id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, ErrEntryNotFound
	}
	return entry, nil
}
