package interorganization_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func TestConsolidationService_CreateRun_CreatesRun(t *testing.T) {
	var created *interorganization.ConsolidationRun
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: baseID(5)}, nil
		}},
	}
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			CreateFunc: func(_ context.Context, run *interorganization.ConsolidationRun) (*interorganization.ConsolidationRun, error) {
				created = run
				return run, nil
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	run, err := svc.CreateRun(context.Background(), interorganization.CreateConsolidationRunRequest{GroupOrganizationID: 10, PeriodID: 5, ReportingCurrency: "IDR"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created == nil || created != run {
		t.Errorf("created run not recorded")
	}
	if *created.GroupOrganizationID != 10 || *created.PeriodID != 5 || created.ReportingCurrency != "IDR" || created.State != interorganization.RunStateDraft {
		t.Errorf("run = %+v, want org 10 period 5 draft", created)
	}
}

func TestConsolidationService_CreateRun_PropagatesOrgFindError(t *testing.T) {
	svc := newConsolidationServiceForTest(
		interorganization.ConsolidationRunDAOMock{},
		interorganization.ConsolidationEliminationDAOMock{},
		interorganization.InterorganizationTransactionDAOMock{},
		organizationLookupMock{FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, errors.New("db down")
		}},
		taxPeriodDAOMock{},
		accountLookupMock{},
		accounting.AccountBalanceDAOMock{},
		procurement.PurchaseOrderLineDAOMock{},
		inventory.ItemResolverMock{},
		inventory.CostLayerDAOMock{},
		rateSourceMock{},
		txMock{},
	)

	_, err := svc.CreateRun(context.Background(), interorganization.CreateConsolidationRunRequest{GroupOrganizationID: 10, PeriodID: 5})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_CreateRun_RejectsMissingOrg(t *testing.T) {
	svc := newConsolidationServiceForTest(
		interorganization.ConsolidationRunDAOMock{},
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

	_, err := svc.CreateRun(context.Background(), interorganization.CreateConsolidationRunRequest{GroupOrganizationID: 10, PeriodID: 5})
	if err != interorganization.ErrConsolidationNoOrg {
		t.Errorf("err = %v, want ErrConsolidationNoOrg", err)
	}
}

func TestConsolidationService_CreateRun_PropagatesPeriodFindError(t *testing.T) {
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10)}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, errors.New("db down")
		}},
	}
	svc := newConsolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.CreateRun(context.Background(), interorganization.CreateConsolidationRunRequest{GroupOrganizationID: 10, PeriodID: 5})
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_CreateRun_RejectsMissingPeriod(t *testing.T) {
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10)}, nil
		},
	}
	svc := newConsolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.CreateRun(context.Background(), interorganization.CreateConsolidationRunRequest{GroupOrganizationID: 10, PeriodID: 5})
	if err != interorganization.ErrConsolidationNoPeriod {
		t.Errorf("err = %v, want ErrConsolidationNoPeriod", err)
	}
}

func TestConsolidationService_ListRuns_ListsRuns(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.ConsolidationRun], error) {
				return &query.Page[interorganization.ConsolidationRun]{Items: []*interorganization.ConsolidationRun{{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10))}}}, nil
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	items, err := svc.ListRuns(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || *items[0].GroupOrganizationID != 10 {
		t.Errorf("items = %+v, want one run for org 10", items)
	}
}

func TestConsolidationService_ListRuns_PropagatesError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.ConsolidationRun], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.ListRuns(context.Background(), 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_GetRun_ReturnsRun(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), State: interorganization.RunStateDraft}, nil
			},
		},
	}
	eliminations := interorganization.ConsolidationEliminationDAOMock{}
	svc := newConsolidationServiceForTest(runs, eliminations, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	run, items, err := svc.GetRun(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run == nil || run.ID != 1 {
		t.Errorf("run = %+v, want run 1", run)
	}
	if len(items) != 0 {
		t.Errorf("items = %+v, want none", items)
	}
}

func TestConsolidationService_GetRun_PropagatesFindError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, _, err := svc.GetRun(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_GetRun_ReturnsNotFound(t *testing.T) {
	svc := newConsolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, _, err := svc.GetRun(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationRunNotFound {
		t.Errorf("err = %v, want ErrConsolidationRunNotFound", err)
	}
}

func TestConsolidationService_GetRun_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(99))}, nil
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, _, err := svc.GetRun(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationRunNotFound {
		t.Errorf("err = %v, want ErrConsolidationRunNotFound", err)
	}
}

func TestConsolidationService_GetRun_PropagatesEliminationError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10))}, nil
			},
		},
	}
	eliminations := interorganization.ConsolidationEliminationDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*interorganization.ConsolidationElimination, error) {
			return nil, errors.New("db down")
		},
	}
	svc := newConsolidationServiceForTest(runs, eliminations, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, _, err := svc.GetRun(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_PropagatesFindError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_ReturnsNotFound(t *testing.T) {
	svc := newConsolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationRunNotFound {
		t.Errorf("err = %v, want ErrConsolidationRunNotFound", err)
	}
}

func TestConsolidationService_Run_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(99)), PeriodID: helper.Ptr(uint64(5)), State: interorganization.RunStateDraft}, nil
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationRunNotFound {
		t.Errorf("err = %v, want ErrConsolidationRunNotFound", err)
	}
}

func TestConsolidationService_Run_RejectsMissingPeriod(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), State: interorganization.RunStateDraft}, nil
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNoPeriod {
		t.Errorf("err = %v, want ErrConsolidationNoPeriod", err)
	}
}

func TestConsolidationService_Run_PropagatesPeriodFindError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), State: interorganization.RunStateDraft}, nil
			},
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return nil, errors.New("db down")
		}},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_RejectsMissingPeriodDates(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), State: interorganization.RunStateDraft}, nil
			},
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return &accounting.TaxPeriod{Base: baseID(5)}, nil
		}},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationPeriodMiss {
		t.Errorf("err = %v, want ErrConsolidationPeriodMiss", err)
	}
}

func TestConsolidationService_Run_RejectsMissingGroupOrg(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "IDR", State: interorganization.RunStateDraft}, nil
			},
		},
	}
	period := &accounting.TaxPeriod{Base: baseID(5), DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)), DateEnd: ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return period, nil
		}},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNoOrg {
		t.Errorf("err = %v, want ErrConsolidationNoOrg", err)
	}
}

func TestConsolidationService_Run_PropagatesMemberListError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "IDR", State: interorganization.RunStateDraft}, nil
			},
		},
	}
	period := &accounting.TaxPeriod{Base: baseID(5), DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)), DateEnd: ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Organization], error) {
			return nil, errors.New("db down")
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return period, nil
		}},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_PropagatesBalanceError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "IDR", State: interorganization.RunStateDraft}, nil
			},
		},
	}
	period := &accounting.TaxPeriod{Base: baseID(5), DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)), DateEnd: ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return period, nil
		}},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return nil, errors.New("db down")
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, balances, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_ReturnsNoFxRate(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "USD", State: interorganization.RunStateDraft}, nil
			},
		},
	}
	period := &accounting.TaxPeriod{Base: baseID(5), DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)), DateEnd: ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return period, nil
		}},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return []accounting.AccountBalance{{AccountID: 1000, Debit: 500}}, nil
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, balances, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err != interorganization.ErrConsolidationNoFx {
		t.Errorf("err = %v, want ErrConsolidationNoFx", err)
	}
}

func TestConsolidationService_Run_PropagatesTransactionListError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "IDR", State: interorganization.RunStateDraft}, nil
			},
		},
	}
	period := &accounting.TaxPeriod{Base: baseID(5), DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)), DateEnd: ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return period, nil
		}},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, trans, orgs, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "db down" {
		t.Errorf("err = %v, want db down", err)
	}
}

func TestConsolidationService_Run_PropagatesTxError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return &interorganization.ConsolidationRun{Base: baseID(1), GroupOrganizationID: helper.Ptr(uint64(10)), PeriodID: helper.Ptr(uint64(5)), ReportingCurrency: "IDR", State: interorganization.RunStateDraft}, nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, run *interorganization.ConsolidationRun) (*interorganization.ConsolidationRun, error) {
			return run, nil
		},
	}
	period := &accounting.TaxPeriod{Base: baseID(5), DateStart: ptrTime(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)), DateEnd: ptrTime(time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC))}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return &reference.Organization{Base: baseID(10), BaseCurrency: "IDR"}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
			return period, nil
		}},
	}
	tx := txMock{
		RunFunc: func(_ context.Context, _ func(tx *gorm.DB) error) error {
			return errors.New("tx failed")
		},
	}
	svc := newConsolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, tx)

	_, err := svc.Run(context.Background(), 1, 10)
	if err == nil || err.Error() != "tx failed" {
		t.Errorf("err = %v, want tx failed", err)
	}
}
