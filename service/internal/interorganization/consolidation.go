package interorganization

import (
	"context"
	"errors"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

type OrganizationLookup interface {
	Find(ctx context.Context, id uint64) (*reference.Organization, error)
	List(ctx context.Context, q *query.Query) (*query.Page[reference.Organization], error)
}

type AccountLookup interface {
	List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

type ConsolidationService struct {
	runs         ConsolidationRunDAO
	eliminations ConsolidationEliminationDAO
	trans        InterorganizationTransactionDAO
	orgs         OrganizationLookup
	periods      accounting.TaxPeriodDAO
	accounts     AccountLookup
	balances     accounting.AccountBalanceDAO
	poLines      procurement.PurchaseOrderLineDAO
	resolver     inventory.ItemResolver
	layers       inventory.CostLayerDAO
	rates        amount.RateSource
	tx           db.Transactioner
}

func NewConsolidationService(
	runs ConsolidationRunDAO,
	eliminations ConsolidationEliminationDAO,
	trans InterorganizationTransactionDAO,
	orgs OrganizationLookup,
	periods accounting.TaxPeriodDAO,
	accounts AccountLookup,
	balances accounting.AccountBalanceDAO,
	poLines procurement.PurchaseOrderLineDAO,
	resolver inventory.ItemResolver,
	layers inventory.CostLayerDAO,
	rates amount.RateSource,
	tx db.Transactioner,
) ConsolidationService {
	return ConsolidationService{
		runs:         runs,
		eliminations: eliminations,
		trans:        trans,
		orgs:         orgs,
		periods:      periods,
		accounts:     accounts,
		balances:     balances,
		poLines:      poLines,
		resolver:     resolver,
		layers:       layers,
		rates:        rates,
		tx:           tx,
	}
}

type CreateConsolidationRunRequest struct {
	GroupOrganizationID uint64
	PeriodID            uint64
	ReportingCurrency   string
}

func (s ConsolidationService) CreateRun(ctx context.Context, request CreateConsolidationRunRequest) (*ConsolidationRun, error) {
	group, err := s.orgs.Find(ctx, request.GroupOrganizationID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrConsolidationNoOrg
	}
	period, err := s.periods.Find(ctx, request.PeriodID)
	if err != nil {
		return nil, err
	}
	if period == nil {
		return nil, ErrConsolidationNoPeriod
	}
	return s.runs.Create(ctx, &ConsolidationRun{
		GroupOrganizationID: helper.Ptr(request.GroupOrganizationID),
		PeriodID:            helper.Ptr(request.PeriodID),
		ReportingCurrency:   request.ReportingCurrency,
		State:               RunStateDraft,
	})
}

func (s ConsolidationService) ListRuns(ctx context.Context, organizationID uint64) ([]*ConsolidationRun, error) {
	page, err := s.runs.List(ctx, &query.Query{Filters: []query.Filter{{Field: "group_organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (s ConsolidationService) GetRun(ctx context.Context, runID, organizationID uint64) (*ConsolidationRun, []*ConsolidationElimination, error) {
	run, err := s.runs.Find(ctx, runID)
	if err != nil {
		return nil, nil, err
	}
	if run == nil || run.GroupOrganizationID == nil || *run.GroupOrganizationID != organizationID {
		return nil, nil, ErrConsolidationRunNotFound
	}
	eliminations, err := s.eliminations.ListByRun(ctx, run.ID)
	if err != nil {
		return nil, nil, err
	}
	return run, eliminations, nil
}

type ConsolidatedBalance struct {
	OrganizationID uint64
	AccountID      uint64
	Amount         float64
}

type ConsolidationResult struct {
	Run            *ConsolidationRun
	MemberBalances []ConsolidatedBalance
	Eliminations   []*ConsolidationElimination
}

func (s ConsolidationService) Run(ctx context.Context, runID, organizationID uint64) (*ConsolidationResult, error) {
	run, err := s.runs.Find(ctx, runID)
	if err != nil {
		return nil, err
	}
	if run == nil || run.GroupOrganizationID == nil || *run.GroupOrganizationID != organizationID {
		return nil, ErrConsolidationRunNotFound
	}
	if run.State != RunStateDraft {
		return nil, ErrConsolidationNotDraft
	}
	if run.PeriodID == nil {
		return nil, ErrConsolidationNoPeriod
	}
	groupOrganizationID := *run.GroupOrganizationID

	period, err := s.periods.Find(ctx, *run.PeriodID)
	if err != nil {
		return nil, err
	}
	if period == nil || period.DateStart == nil || period.DateEnd == nil {
		return nil, ErrConsolidationPeriodMiss
	}

	members, err := s.collectGroup(ctx, groupOrganizationID)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, ErrConsolidationNoMember
	}

	converter := amount.NewConverter(run.ReportingCurrency, s.rates)
	periodEnd := *period.DateEnd

	memberBalances := []ConsolidatedBalance{}
	for _, member := range members {
		balances, err := s.balances.ListBalancesByPeriod(ctx, member.ID, *period.DateStart, periodEnd)
		if err != nil {
			return nil, err
		}
		for _, balance := range balances {
			net := balance.Debit - balance.Credit
			if net == 0 {
				continue
			}
			converted, err := converter.Convert(ctx, amount.FromFloat64(net), member.BaseCurrency, run.ReportingCurrency, groupOrganizationID, amount.RateAverage, periodEnd)
			if err != nil {
				if errors.Is(err, amount.ErrInvalidRate) {
					return nil, ErrConsolidationNoFx
				}
				return nil, err
			}
			memberBalances = append(memberBalances, ConsolidatedBalance{
				OrganizationID: member.ID,
				AccountID:      balance.AccountID,
				Amount:         converted.Float64(),
			})
		}
	}

	eliminations, err := s.eliminationsFor(ctx, run, periodEnd, converter, members)
	if err != nil {
		return nil, err
	}

	err = s.tx.Run(ctx, func(tx *gorm.DB) error {
		for _, elimination := range eliminations {
			if _, err := s.eliminations.CreateTx(ctx, tx, elimination); err != nil {
				return err
			}
		}
		run.State = RunStateDone
		if _, err := s.runs.UpdateTx(ctx, tx, run); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &ConsolidationResult{Run: run, MemberBalances: memberBalances, Eliminations: eliminations}, nil
}

func (s ConsolidationService) collectGroup(ctx context.Context, groupOrganizationID uint64) ([]*reference.Organization, error) {
	group, err := s.orgs.Find(ctx, groupOrganizationID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrConsolidationNoOrg
	}
	members := []*reference.Organization{group}
	seen := map[uint64]bool{group.ID: true}
	frontier := []uint64{group.ID}
	for len(frontier) > 0 {
		var next []uint64
		for _, parentID := range frontier {
			page, err := s.orgs.List(ctx, &query.Query{Filters: []query.Filter{{Field: "parent_id", Operator: query.Equal, Value: parentID}}})
			if err != nil {
				return nil, err
			}
			children := page.Items
			for _, child := range children {
				if seen[child.ID] {
					continue
				}
				seen[child.ID] = true
				members = append(members, child)
				next = append(next, child.ID)
			}
		}
		frontier = next
	}
	return members, nil
}

func (s ConsolidationService) eliminationsFor(ctx context.Context, run *ConsolidationRun, periodEnd time.Time, converter amount.Converter, members []*reference.Organization) ([]*ConsolidationElimination, error) {
	memberSet := make(map[uint64]bool, len(members))
	for _, member := range members {
		memberSet[member.ID] = true
	}
	transactions, err := s.trans.List(ctx, &query.Query{Filters: []query.Filter{{Field: "state", Operator: query.Equal, Value: TransactionStateDone}}})
	if err != nil {
		return nil, err
	}
	groupOrganizationID := *run.GroupOrganizationID
	eliminations := []*ConsolidationElimination{}
	for _, transaction := range transactions.Items {
		if transaction.SourceOrganizationID == nil || transaction.MirrorOrganizationID == nil ||
			transaction.SourceID == nil || transaction.MirrorID == nil {
			continue
		}
		sellerOrgID, buyerOrgID := *transaction.SourceOrganizationID, *transaction.MirrorOrganizationID
		if !memberSet[sellerOrgID] || !memberSet[buyerOrgID] {
			continue
		}
		seller, err := s.orgs.Find(ctx, sellerOrgID)
		if err != nil {
			return nil, err
		}
		if seller == nil {
			return nil, ErrConsolidationNoMember
		}
		value, err := converter.Convert(ctx, amount.FromFloat64(transaction.Amount), seller.BaseCurrency, run.ReportingCurrency, groupOrganizationID, amount.RateAverage, periodEnd)
		if err != nil {
			if errors.Is(err, amount.ErrInvalidRate) {
				return nil, ErrConsolidationNoFx
			}
			return nil, err
		}
		amountValue := value.Float64()

		sellerReceivable, err := s.findAccount(ctx, sellerOrgID, "receivable")
		if err != nil {
			return nil, err
		}
		buyerPayable, err := s.findAccount(ctx, buyerOrgID, "payable")
		if err != nil {
			return nil, err
		}
		sellerIncome, err := s.findAccount(ctx, sellerOrgID, "income")
		if err != nil {
			return nil, err
		}
		buyerExpense, err := s.findAccount(ctx, buyerOrgID, "expense")
		if err != nil {
			return nil, err
		}

		if sellerReceivable != nil {
			eliminations = append(eliminations, &ConsolidationElimination{
				ConsolidationRunID:         run.ID,
				AccountID:                  sellerReceivable.ID,
				CounterpartyOrganizationID: helper.Ptr(buyerOrgID),
				Amount:                     -amountValue,
				Description:                "AR elimination",
			})
		}
		if buyerPayable != nil {
			eliminations = append(eliminations, &ConsolidationElimination{
				ConsolidationRunID:         run.ID,
				AccountID:                  buyerPayable.ID,
				CounterpartyOrganizationID: helper.Ptr(sellerOrgID),
				Amount:                     amountValue,
				Description:                "AP elimination",
			})
		}
		if sellerIncome != nil {
			eliminations = append(eliminations, &ConsolidationElimination{
				ConsolidationRunID:         run.ID,
				AccountID:                  sellerIncome.ID,
				CounterpartyOrganizationID: helper.Ptr(buyerOrgID),
				Amount:                     amountValue,
				Description:                "Sales elimination",
			})
		}
		if buyerExpense != nil {
			eliminations = append(eliminations, &ConsolidationElimination{
				ConsolidationRunID:         run.ID,
				AccountID:                  buyerExpense.ID,
				CounterpartyOrganizationID: helper.Ptr(sellerOrgID),
				Amount:                     -amountValue,
				Description:                "Cost elimination",
			})
		}

		if transaction.MirrorType == "purchase_order" {
			unrealized, err := s.unrealizedProfit(ctx, run, buyerOrgID, sellerOrgID, *transaction.MirrorID, periodEnd, converter)
			if err != nil {
				return nil, err
			}
			eliminations = append(eliminations, unrealized...)
		}
	}
	return eliminations, nil
}

func (s ConsolidationService) unrealizedProfit(ctx context.Context, run *ConsolidationRun, buyerOrgID, sellerOrgID, purchaseOrderID uint64, periodEnd time.Time, converter amount.Converter) ([]*ConsolidationElimination, error) {
	lines, err := s.poLines.ListByOrder(ctx, purchaseOrderID)
	if err != nil {
		return nil, err
	}
	buyer, err := s.orgs.Find(ctx, buyerOrgID)
	if err != nil {
		return nil, err
	}
	if buyer == nil {
		return nil, ErrConsolidationNoMember
	}
	groupOrganizationID := *run.GroupOrganizationID
	eliminations := []*ConsolidationElimination{}
	for _, line := range lines {
		if line.ItemID == nil || line.QtyOrdered <= 0 {
			continue
		}
		resolved, err := s.resolver.Resolve(ctx, *line.ItemID)
		if err != nil {
			return nil, err
		}
		layers, err := s.layers.ListOpenByItemInOrg(ctx, *line.ItemID, buyerOrgID)
		if err != nil {
			return nil, err
		}
		var remainingQty float64
		for _, layer := range layers {
			remainingQty += layer.RemainingQty
		}
		profitPerUnit := line.UnitPrice - resolved.StandardCost
		if profitPerUnit <= 0 || remainingQty <= 0 {
			continue
		}
		unrealized, err := converter.Convert(ctx, amount.FromFloat64(profitPerUnit*remainingQty), buyer.BaseCurrency, run.ReportingCurrency, groupOrganizationID, amount.RateAverage, periodEnd)
		if err != nil {
			if errors.Is(err, amount.ErrInvalidRate) {
				return nil, ErrConsolidationNoFx
			}
			return nil, err
		}
		unrealizedValue := unrealized.Float64()
		if resolved.StockValuationAccountID != 0 {
			eliminations = append(eliminations, &ConsolidationElimination{
				ConsolidationRunID:         run.ID,
				AccountID:                  resolved.StockValuationAccountID,
				CounterpartyOrganizationID: helper.Ptr(sellerOrgID),
				Amount:                     -unrealizedValue,
				Description:                "Unrealized profit elimination (inventory)",
			})
		}
		if resolved.CogsAccountID != 0 {
			eliminations = append(eliminations, &ConsolidationElimination{
				ConsolidationRunID:         run.ID,
				AccountID:                  resolved.CogsAccountID,
				CounterpartyOrganizationID: helper.Ptr(sellerOrgID),
				Amount:                     unrealizedValue,
				Description:                "Unrealized profit elimination (COGS)",
			})
		}
	}
	return eliminations, nil
}

func (s ConsolidationService) findAccount(ctx context.Context, organizationID uint64, accountType string) (*reference.Account, error) {
	page, err := s.accounts.List(ctx, &query.Query{Filters: []query.Filter{{Field: "organization_id", Operator: query.Equal, Value: organizationID}}})
	if err != nil {
		return nil, err
	}
	for _, account := range page.Items {
		if account.Type == accountType && account.Active {
			return account, nil
		}
	}
	return nil, nil
}
