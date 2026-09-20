package interorganization_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
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

func consolidationRunForTest(groupOrganizationID, periodID uint64) *interorganization.ConsolidationRun {
	return &interorganization.ConsolidationRun{
		Base:                baseID(1),
		GroupOrganizationID: helper.Ptr(groupOrganizationID),
		PeriodID:            helper.Ptr(periodID),
		ReportingCurrency:   "IDR",
		State:               interorganization.RunStateDraft,
	}
}

func consolidationPeriodForTest() *accounting.TaxPeriod {
	return &accounting.TaxPeriod{
		Base:      baseID(5),
		DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)),
		DateEnd:   ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)),
	}
}

func consolidationServiceForTestWith(
	runs interorganization.ConsolidationRunDAO,
	eliminations interorganization.ConsolidationEliminationDAO,
	trans interorganization.InterorganizationTransactionDAO,
	orgs interorganization.OrganizationLookup,
	balances accounting.AccountBalanceDAO,
	accounts interorganization.AccountLookup,
	poLines procurement.PurchaseOrderLineDAO,
	resolver inventory.ItemResolver,
	layers inventory.CostLayerDAO,
) interorganization.ConsolidationService {
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return consolidationPeriodForTest(), nil
		}},
	}
	return newConsolidationServiceForTest(runs, eliminations, trans, orgs, periods, accounts, balances, poLines, resolver, layers, rateSourceMock{}, txMock{})
}

func TestConsolidationService_Run_SkipsZeroNetBalance(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return []accounting.AccountBalance{{AccountID: 1000, Debit: 500, Credit: 500}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, balances, accountLookupMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	result, err := svc.Run(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.MemberBalances) != 0 {
		t.Errorf("member balances = %+v, want none", result.MemberBalances)
	}
	if result.Run.State != interorganization.RunStateDone {
		t.Errorf("run state = %s, want done", result.Run.State)
	}
}

func TestConsolidationService_Run_PropagatesNonFxConversionError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				run := consolidationRunForTest(10, 5)
				run.ReportingCurrency = "USD"
				return run, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return []accounting.AccountBalance{{AccountID: 1000, Debit: 500}}, nil
		},
	}
	rates := rateSourceMock{
		RateFunc: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.Amount{}, errors.New("rate down")
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return consolidationPeriodForTest(), nil
		}},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, balances, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rates, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "rate down" {
		t.Errorf("err = %v, want rate down", err)
	}
}

func TestConsolidationService_Run_ConsolidatesTransactionsAndCreatesEliminations(t *testing.T) {
	var created []*interorganization.ConsolidationElimination
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	eliminations := interorganization.ConsolidationEliminationDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, elimination *interorganization.ConsolidationElimination) (*interorganization.ConsolidationElimination, error) {
			created = append(created, elimination)
			return elimination, nil
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{
					{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, MirrorType: "sale_order", State: interorganization.TransactionStateDone},
					{Base: baseID(402), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(31)), MirrorID: helper.Ptr(uint64(201)), Amount: 50, State: interorganization.TransactionStateDone},
					{Base: baseID(403), SourceOrganizationID: helper.Ptr(uint64(99)), MirrorOrganizationID: helper.Ptr(uint64(98)), SourceID: helper.Ptr(uint64(32)), MirrorID: helper.Ptr(uint64(202)), Amount: 50, State: interorganization.TransactionStateDone},
					{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(33)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone},
				}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			switch id {
			case 10:
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			case 11:
				return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
			}
			return nil, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(1002), Type: "income", Active: false}}}, nil
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
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 100}, {Base: baseID(211), QtyOrdered: 1}}, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 2003, CogsAccountID: 2004}, StandardCost: 40}, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		ListOpenByItemInOrgFunc: func(_ context.Context, _, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: baseID(500), RemainingQty: 2}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, eliminations, trans, orgs, balances, accounts, poLines, resolver, layers)

	result, err := svc.Run(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.MemberBalances) != 2 {
		t.Errorf("member balances = %+v, want 2", result.MemberBalances)
	}
	expected := map[uint64]float64{1001: -160, 2001: 160, 2002: -160, 2003: -120, 2004: 120}
	actual := map[uint64]float64{}
	for _, elimination := range created {
		actual[elimination.AccountID] += elimination.Amount
	}
	if len(actual) != len(expected) {
		t.Fatalf("created eliminations = %+v, want %d accounts", created, len(expected))
	}
	for accountID, want := range expected {
		if actual[accountID] != want {
			t.Errorf("elimination account %d amount = %v, want %v", accountID, actual[accountID], want)
		}
	}
}

func TestConsolidationService_Run_PropagatesSellerFindError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(11)), MirrorOrganizationID: helper.Ptr(uint64(10)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return nil, errors.New("db down")
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accountLookupMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_RejectsMissingSeller(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(11)), MirrorOrganizationID: helper.Ptr(uint64(10)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return nil, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accountLookupMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNoMember {
		t.Errorf("err = %v, want ErrConsolidationNoMember", err)
	}
}

func TestConsolidationService_Run_RejectsNoFxInEliminations(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "USD"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accountLookupMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNoFx {
		t.Errorf("err = %v, want ErrConsolidationNoFx", err)
	}
}

func TestConsolidationService_Run_PropagatesAccountListError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return nil, errors.New("db down")
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_PropagatesPoLineError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return nil, errors.New("db down")
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, poLines, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_PropagatesBuyerFindError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			if id == 11 {
				return nil, errors.New("db down")
			}
			return nil, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, poLines, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_RejectsMissingBuyer(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return nil, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, poLines, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNoMember {
		t.Errorf("err = %v, want ErrConsolidationNoMember", err)
	}
}

func TestConsolidationService_Run_PropagatesResolverError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{}, errors.New("resolve failed")
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, poLines, resolver, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "resolve failed" {
		t.Errorf("err = %v, want resolve failed", err)
	}
}

func TestConsolidationService_Run_PropagatesLayersError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
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
		ListOpenByItemInOrgFunc: func(_ context.Context, _, _ uint64) ([]*inventory.CostLayer, error) {
			return nil, errors.New("layers down")
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, poLines, resolver, layers)

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "layers down" {
		t.Errorf("err = %v, want layers down", err)
	}
}

func TestConsolidationService_Run_SkipsNonProfitableLines(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 30}}, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StandardCost: 40}, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		ListOpenByItemInOrgFunc: func(_ context.Context, _, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: baseID(500), RemainingQty: 2}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, poLines, resolver, layers)

	result, err := svc.Run(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Eliminations) != 4 {
		t.Errorf("eliminations = %+v, want the 4 ar/ap/income/expense entries", result.Eliminations)
	}
}

func TestConsolidationService_Run_RejectsNoFxInUnrealizedProfit(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "USD"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "USD"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StandardCost: 40}, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		ListOpenByItemInOrgFunc: func(_ context.Context, _, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: baseID(500), RemainingQty: 2}}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, poLines, resolver, layers)

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNoFx {
		t.Errorf("err = %v, want ErrConsolidationNoFx", err)
	}
}

func TestConsolidationService_Run_PropagatesEliminationCreateError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	eliminations := interorganization.ConsolidationEliminationDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, _ *interorganization.ConsolidationElimination) (*interorganization.ConsolidationElimination, error) {
			return nil, errors.New("insert failed")
		},
	}
	svc := consolidationServiceForTestWith(runs, eliminations, trans, orgs, accounting.AccountBalanceDAOMock{}, accounts, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "insert failed" {
		t.Errorf("err = %v, want insert failed", err)
	}
}

func TestConsolidationService_Run_PropagatesRunUpdateError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, _ *interorganization.ConsolidationRun) (*interorganization.ConsolidationRun, error) {
			return nil, errors.New("update failed")
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, accounting.AccountBalanceDAOMock{}, accountLookupMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "update failed" {
		t.Errorf("err = %v, want update failed", err)
	}
}

func TestConsolidationService_Run_SkipsAlreadySeenMember(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(id), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(10), BaseCurrency: "IDR"}}}, nil
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
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, balances, accountLookupMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	result, err := svc.Run(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.MemberBalances) != 2 {
		t.Errorf("member balances = %+v, want 2 members without duplication", result.MemberBalances)
	}
}

func TestConsolidationService_Run_PropagatesCollectGroupFindError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunForTest(10, 5), nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, errors.New("db down")
		},
	}
	svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, accounting.AccountBalanceDAOMock{}, accountLookupMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_PropagatesEliminationConversionError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				run := consolidationRunForTest(10, 5)
				run.ReportingCurrency = "USD"
				return run, nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return []accounting.AccountBalance{}, nil
		},
	}
	rates := rateSourceMock{
		RateFunc: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.Amount{}, errors.New("rate down")
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return consolidationPeriodForTest(), nil
		}},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, periods, accountLookupMock{}, balances, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rates, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "rate down" {
		t.Errorf("err = %v, want rate down", err)
	}
}

func TestConsolidationService_Run_PropagatesFindAccountErrors(t *testing.T) {
	for _, failOnCall := range []int{2, 3, 4} {
		t.Run(fmt.Sprintf("call %d", failOnCall), func(t *testing.T) {
			runs := interorganization.ConsolidationRunDAOMock{
				CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
					FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
						return consolidationRunForTest(10, 5), nil
					},
				},
			}
			trans := interorganization.InterorganizationTransactionDAOMock{
				CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
					ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
						return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(401), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(200)), Amount: 100, State: interorganization.TransactionStateDone}}}, nil
					},
				},
			}
			orgs := organizationLookupMock{
				FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
					if id == 10 {
						return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
					}
					return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
				},
				ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
					if q.Filters[0].Value.(uint64) == 10 {
						return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
					}
					return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
				},
			}
			balances := accounting.AccountBalanceDAOMock{
				ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
					return []accounting.AccountBalance{}, nil
				},
			}
			calls := 0
			accounts := accountLookupMock{
				ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
					calls++
					if calls == failOnCall {
						return nil, errors.New("db down")
					}
					if q.Filters[0].Value.(uint64) == 10 {
						return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}}}, nil
					}
					return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
				},
			}
			svc := consolidationServiceForTestWith(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, balances, accounts, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{})

			_, err := svc.Run(context.Background(), 1, 10)
			if err == nil || err.Error() != "db down" {
				t.Errorf("err = %v, want db down", err)
			}
		})
	}
}

func TestConsolidationService_Run_PropagatesUnrealizedProfitConversionError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				run := consolidationRunForTest(10, 5)
				run.ReportingCurrency = "USD"
				return run, nil
			},
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: baseID(404), SourceOrganizationID: helper.Ptr(uint64(10)), MirrorOrganizationID: helper.Ptr(uint64(11)), SourceID: helper.Ptr(uint64(30)), MirrorID: helper.Ptr(uint64(203)), Amount: 60, MirrorType: "purchase_order", State: interorganization.TransactionStateDone}}}, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, id uint64) (*reference.Organization, error) {
			if id == 10 {
				return &reference.Organization{Base: baseID(10), BaseCurrency: "USD"}, nil
			}
			return &reference.Organization{Base: baseID(11), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Organization]{Items: []*reference.Organization{{Base: baseID(11), BaseCurrency: "IDR"}}}, nil
			}
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return []accounting.AccountBalance{}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.Account], error) {
			if q.Filters[0].Value.(uint64) == 10 {
				return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(1001), Type: "receivable", Active: true}, {Base: baseID(1002), Type: "income", Active: true}}}, nil
			}
			return &query.Page[reference.Account]{Items: []*reference.Account{{Base: baseID(2001), Type: "payable", Active: true}, {Base: baseID(2002), Type: "expense", Active: true}}}, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: baseID(210), ItemID: helper.Ptr(uint64(300)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StandardCost: 40}, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
		ListOpenByItemInOrgFunc: func(_ context.Context, _, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: baseID(500), RemainingQty: 2}}, nil
		},
	}
	rates := rateSourceMock{
		RateFunc: func(_ context.Context, _ string, _ uint64, _ amount.RateType, _ time.Time) (amount.Amount, error) {
			return amount.Amount{}, errors.New("rate down")
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return consolidationPeriodForTest(), nil
		}},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, periods, accounts, balances, poLines, resolver, layers, rates, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "rate down" {
		t.Errorf("err = %v, want rate down", err)
	}
}
