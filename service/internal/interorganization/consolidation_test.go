package interorganization_test

import (
	"context"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/db"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func newConsolidationServiceForTest(
	runs interorganization.ConsolidationRunDAO,
	eliminations interorganization.ConsolidationEliminationDAO,
	trans interorganization.InterorganizationTransactionDAO,
	orgs interorganization.OrganizationLookup,
	periods accounting.TaxPeriodDAO,
	accounts interorganization.AccountLookup,
	balances accounting.AccountBalanceDAO,
	poLines procurement.PurchaseOrderLineDAO,
	resolver inventory.ItemResolver,
	layers inventory.CostLayerDAO,
	rates amount.RateSource,
	tx db.Transactioner,
) interorganization.ConsolidationService {
	return interorganization.NewConsolidationService(runs, eliminations, trans, orgs, periods, accounts, balances, poLines, resolver, layers, rates, tx)
}

func TestConsolidationService_Run_TranslatesBalancesAndCreatesEliminations(t *testing.T) {
	var runState string
	period := &accounting.TaxPeriod{Base: baseID(5), Name: "2026-07", DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)), DateEnd: ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))}
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "IDR", State: interorganization.RunStateDraft}, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, run *interorganization.ConsolidationRun) (*interorganization.ConsolidationRun, error) {
			runState = run.State
			return run, nil
		},
	}
	eliminations := interorganization.ConsolidationEliminationDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, elimination *interorganization.ConsolidationElimination) (*interorganization.ConsolidationElimination, error) {
			return elimination, nil
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(400), SourceOrganizationID: helper.Ptr(uint64(10)), SourceType: "sale_order", SourceID: helper.Ptr(uint64(30)), MirrorOrganizationID: helper.Ptr(uint64(11)), MirrorType: "purchase_order", MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			switch id {
			case 10:
				return &reference.Organization{Base: baseID(10), Name: "Parent", BaseCurrency: "IDR"}, nil
			case 11:
				return &reference.Organization{Base: baseID(11), Name: "Child", ParentID: helper.Ptr(uint64(10)), BaseCurrency: "IDR"}, nil
			}
			return nil, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			value := q.Filters[0].Value.(uint64)
			if value == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), ParentID: helper.Ptr(uint64(10)), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return period, nil
		}},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}}}, nil
			}
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, organizationID uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			if organizationID == 10 {
				return []accounting.AccountBalance{{AccountID: 1000, Debit: 500}}, nil
			}
			return []accounting.AccountBalance{{AccountID: 2000, Debit: 100}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 2003, CogsAccountID: 2004}, StandardCost: 40}, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		ListOpenByItemInOrgFunc: func(_ context.Context, _ uint64, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: baseID(500), RemainingQty: 2}}, nil
		},
	}
	svc := newConsolidationServiceForTest(runs, eliminations, trans, orgs, periods, accounts, balances, poLines, resolver, layers, rateSourceMock{}, txMock{})

	result, err := svc.Run(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Run.State != interorganization.RunStateDone {
		t.Errorf("run state = %s, want done", result.Run.State)
	}
	if runState != interorganization.RunStateDone {
		t.Errorf("persisted run state = %s, want done", runState)
	}
	if len(result.MemberBalances) != 2 {
		t.Fatalf("member balances = %+v, want 2", result.MemberBalances)
	}
	if result.MemberBalances[0].OrganizationID != 10 || result.MemberBalances[0].Amount != 500 {
		t.Errorf("balance = %+v, want org 10 amount 500", result.MemberBalances[0])
	}
	expected := map[uint64]float64{1001: -100, 2001: 100, 1002: 100, 2002: -100, 2003: -120, 2004: 120}
	if len(result.Eliminations) != len(expected) {
		t.Fatalf("eliminations = %+v, want %d", result.Eliminations, len(expected))
	}
	for _, elimination := range result.Eliminations {
		if want := expected[elimination.AccountID]; elimination.Amount != want {
			t.Errorf("elimination account %d amount = %v, want %v", elimination.AccountID, elimination.Amount, want)
		}
	}
}

func TestConsolidationService_Run_CompletedRunRejected(t *testing.T) {
	svc := newConsolidationServiceForTest(
		interorganization.ConsolidationRunDAOMock{CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
			return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "IDR", State: interorganization.RunStateDone}, nil
		}}},
		interorganization.ConsolidationEliminationDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		organizationLookupMock{},
		taxPeriodDAOMock{},
		accountLookupMock{},
		accounting.AccountBalanceDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		inventory.ItemResolverMock{},
		inventory.CostLayerDAOMock{},
		rateSourceMock{},
		txMock{},
	)

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNotDraft {
		t.Errorf("err = %v, want ErrConsolidationNotDraft", err)
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
