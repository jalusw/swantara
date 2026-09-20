package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
	"gorm.io/gorm"
)

func interorganizationTestApp(t *testing.T, register func(fiber.Router, httpx.RouteGuards)) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(model.ActorKey, uint64(5))
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	register(app, passthroughGuards())
	return app
}

func interorganizationTestAppNoTenant(t *testing.T, register func(fiber.Router, httpx.RouteGuards)) *fiber.App {
	t.Helper()
	app := fiber.New()
	register(app, passthroughGuards())
	return app
}

func consolidationServiceForTest(
	runs interorganization.ConsolidationRunDAOMock,
	eliminations interorganization.ConsolidationEliminationDAOMock,
	trans interorganization.InterorganizationTransactionDAOMock,
	orgs organizationLookupMock,
	periods taxPeriodDAOMock,
	accounts accountLookupMock,
	balances accounting.AccountBalanceDAOMock,
	poLines procurement.PurchaseOrderLineDAOMock,
	resolver inventory.ItemResolverMock,
	layers inventory.CostLayerDAOMock,
	rates rateSourceMock,
	tx txMock,
) interorganization.ConsolidationService {
	return interorganization.NewConsolidationService(
		runs, eliminations, trans, orgs, periods, accounts, balances, poLines, resolver, layers, rates, tx,
	)
}

func dropshipServiceForTest(
	links interorganization.DropshipLinkDAOMock,
	poCreate purchaseOrderCreatorMock,
	poOrders procurement.PurchaseOrderDAOMock,
	poLines procurement.PurchaseOrderLineDAOMock,
	soOrders sales.SaleOrderDAOMock,
	soLines sales.SaleOrderLineDAOMock,
	movements inventory.StockMovementDAOMock,
	locations inventory.StockLocationDAOMock,
	resolver inventory.ItemResolverMock,
	poster inventory.PosterMock,
	tx txMock,
) interorganization.DropShipService {
	return interorganization.NewDropShipService(
		links, poCreate, poOrders, poLines, soOrders, soLines, movements, locations, resolver, poster, tx,
	)
}

func interorganizationServiceForTest(
	rules interorganization.InterorganizationRuleDAOMock,
	trans interorganization.InterorganizationTransactionDAOMock,
	poCreate purchaseOrderCreatorMock,
	soOrders sales.SaleOrderDAOMock,
	soLines sales.SaleOrderLineDAOMock,
) interorganization.InterorganizationService {
	return interorganization.NewInterorganizationService(rules, trans, poCreate, soOrders, soLines)
}

func consolidationRunDraft() *interorganization.ConsolidationRun {
	return &interorganization.ConsolidationRun{
		Base:                model.Base{ID: 1},
		GroupOrganizationID: helper.Ptr(uint64(10)),
		PeriodID:            helper.Ptr(uint64(1)),
		ReportingCurrency:   "USD",
		State:               interorganization.RunStateDraft,
	}
}

func consolidationRunDone() *interorganization.ConsolidationRun {
	run := consolidationRunDraft()
	run.State = interorganization.RunStateDone
	return run
}

func organizationWithID(id uint64, currency string) *reference.Organization {
	return &reference.Organization{Base: model.Base{ID: id}, BaseCurrency: currency}
}

func taxPeriodOpen() *accounting.TaxPeriod {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	return &accounting.TaxPeriod{
		Base: model.Base{ID: 1}, DateStart: &start, DateEnd: &end,
	}
}

func saleOrderConfirmed(id uint64) *sales.SaleOrder {
	return &sales.SaleOrder{
		Base: model.Base{ID: id}, OrganizationID: helper.Ptr(uint64(10)),
		State: sales.OrderStateConfirmed, CurrencyCode: helper.Ptr("USD"),
	}
}

func TestConsolidationHandler_CreateRun_CreatesRun(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			CreateFunc: func(_ context.Context, run *interorganization.ConsolidationRun) (*interorganization.ConsolidationRun, error) {
				run.ID = 1
				return run, nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return organizationWithID(10, "USD"), nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return taxPeriodOpen(), nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"period_id":1,"reporting_currency":"USD"}`
	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestConsolidationHandler_CreateRun_MapsNoOrg(t *testing.T) {
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, nil
		},
	}
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"period_id":1,"reporting_currency":"USD"}`
	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_CreateRun_MapsNoPeriod(t *testing.T) {
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return organizationWithID(10, "USD"), nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return nil, nil
			},
		},
	}
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"period_id":1,"reporting_currency":"USD"}`
	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_CreateRun_RejectsValidation(t *testing.T) {
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"period_id":0,"reporting_currency":"USD"}`
	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_CreateRun_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	body := `{"period_id":1,"reporting_currency":"USD"}`
	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_CreateRun_ReturnsServerError(t *testing.T) {
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return nil, errors.New("db down")
		},
	}
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"period_id":1,"reporting_currency":"USD"}`
	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestConsolidationHandler_ListRuns_ListsRuns(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.ConsolidationRun], error) {
				return &query.Page[interorganization.ConsolidationRun]{Items: []*interorganization.ConsolidationRun{consolidationRunDraft()}, Count: 1}, nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestConsolidationHandler_ListRuns_ReturnsServerError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.ConsolidationRun], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestConsolidationHandler_ListRuns_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_GetRun_ReturnsRun(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunDraft(), nil
			},
		},
	}
	eliminations := interorganization.ConsolidationEliminationDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*interorganization.ConsolidationElimination, error) {
			return []*interorganization.ConsolidationElimination{{Base: model.Base{ID: 1}, AccountID: 7, Amount: 100}}, nil
		},
	}
	svc := consolidationServiceForTest(runs, eliminations, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestConsolidationHandler_GetRun_ReturnsNotFound(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return nil, nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestConsolidationHandler_GetRun_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	run := consolidationRunDraft()
	run.GroupOrganizationID = helper.Ptr(uint64(99))
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return run, nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestConsolidationHandler_GetRun_ReturnsServerError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestConsolidationHandler_GetRun_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/consolidation-runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_RunsConsolidation(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunDraft(), nil
			},
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, run *interorganization.ConsolidationRun) (*interorganization.ConsolidationRun, error) {
			return run, nil
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return organizationWithID(10, "USD"), nil
		},
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Organization], error) {
			return &query.Page[reference.Organization]{Items: []*reference.Organization{organizationWithID(20, "USD")}}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return taxPeriodOpen(), nil
			},
		},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return []accounting.AccountBalance{{AccountID: 1, Debit: 500}}, nil
		},
	}
	accounts := accountLookupMock{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Account], error) {
			return &query.Page[reference.Account]{Items: []*reference.Account{
				{Base: model.Base{ID: 1}, Type: "receivable", Active: true},
				{Base: model.Base{ID: 2}, Type: "payable", Active: true},
				{Base: model.Base{ID: 3}, Type: "income", Active: true},
				{Base: model.Base{ID: 4}, Type: "expense", Active: true},
			}}, nil
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{
					SourceOrganizationID: helper.Ptr(uint64(10)),
					MirrorOrganizationID: helper.Ptr(uint64(20)),
					SourceID:             helper.Ptr(uint64(1)),
					MirrorID:             helper.Ptr(uint64(2)),
					Amount:               100,
					MirrorType:           "sale_order",
				}}}, nil
			},
		},
	}
	eliminations := interorganization.ConsolidationEliminationDAOMock{
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, e *interorganization.ConsolidationElimination) (*interorganization.ConsolidationElimination, error) {
			return e, nil
		},
	}
	svc := consolidationServiceForTest(runs, eliminations, trans, orgs, periods, accounts, balances, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_ReturnsNotFound(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return nil, nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_MapsNotDraft(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunDone(), nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_MapsMissingPeriod(t *testing.T) {
	run := consolidationRunDraft()
	run.PeriodID = nil
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return run, nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_MapsPeriodMiss(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunDraft(), nil
			},
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return nil, nil
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, periods, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_MapsNoFxRate(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return consolidationRunDraft(), nil
			},
		},
	}
	orgs := organizationLookupMock{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Organization, error) {
			return organizationWithID(10, "IDR"), nil
		},
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Organization], error) {
			return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
		},
	}
	periods := taxPeriodDAOMock{
		CRUDMock: dao.CRUDMock[accounting.TaxPeriod]{
			FindFunc: func(_ context.Context, _ uint64) (*accounting.TaxPeriod, error) {
				return taxPeriodOpen(), nil
			},
		},
	}
	balances := accounting.AccountBalanceDAOMock{
		ListBalancesByPeriodFunc: func(_ context.Context, _ uint64, _, _ time.Time) ([]accounting.AccountBalance, error) {
			return []accounting.AccountBalance{{AccountID: 1, Debit: 100, Credit: 0}}, nil
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, orgs, periods, accountLookupMock{}, balances, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_ReturnsServerError(t *testing.T) {
	runs := interorganization.ConsolidationRunDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.ConsolidationRun]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.ConsolidationRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := consolidationServiceForTest(runs, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestConsolidationHandler_Run_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := consolidationServiceForTest(interorganization.ConsolidationRunDAOMock{}, interorganization.ConsolidationEliminationDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, organizationLookupMock{}, taxPeriodDAOMock{}, accountLookupMock{}, accounting.AccountBalanceDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.ItemResolverMock{}, inventory.CostLayerDAOMock{}, rateSourceMock{}, txMock{})
	h := NewConsolidationHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	resp, err := doRequest(app, http.MethodPost, "/consolidation-runs/1/run", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_CreatesOrder(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(7)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	poCreate := purchaseOrderCreatorMock{
		CreateFunc: func(_ context.Context, order *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			order.ID = 1
			return order, nil
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: model.Base{ID: 1}, OrderID: 1}}, nil
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, poCreate, procurement.PurchaseOrderDAOMock{}, poLines, soOrders, soLines, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20,"dest_location_id":30,"date":"2026-01-15"}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_ReturnsNotFound(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, soOrders, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_MapsSourceNotActive(t *testing.T) {
	so := saleOrderConfirmed(1)
	so.State = sales.OrderStateCancelled
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return so, nil
			},
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, soOrders, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_MapsNoLines(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}}}, nil
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, soOrders, soLines, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_MapsDestinationMiss(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(7)), QtyOrdered: 2}}, nil
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, soOrders, soLines, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_RejectsValidation(t *testing.T) {
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":0,"supplier_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_RejectsInvalidDate(t *testing.T) {
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20,"date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Create_ReturnsServerError(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, soOrders, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"sale_order_id":1,"supplier_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_ReceivesOrder(t *testing.T) {
	po := &procurement.PurchaseOrder{
		Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)),
		State: procurement.PurchaseOrderStateConfirmed, DestLocationID: helper.Ptr(uint64(30)),
	}
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return po, nil
			},
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: model.Base{ID: 1}, OrderID: 1, ItemID: helper.Ptr(uint64(7)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	links := interorganization.DropshipLinkDAOMock{
		ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
			return []*interorganization.DropshipLink{{Base: model.Base{ID: 1}, SaleOrderLineID: 1, PurchaseOrderLineID: 1}}, nil
		},
	}
	so := saleOrderConfirmed(1)
	soLines := sales.SaleOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
				return &sales.SaleOrderLine{Base: model.Base{ID: 1}, OrderID: 1, QtyOrdered: 2}, nil
			},
		},
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, OrderID: 1, QtyOrdered: 2}}, nil
		},
	}
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return so, nil
			},
		},
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 20}, Usage: "supplier"}}}, nil
			},
		},
	}
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{CogsAccountID: 100, StockInputAccountID: 200}}, nil
		},
	}
	svc := dropshipServiceForTest(links, purchaseOrderCreatorMock{}, poOrders, poLines, soOrders, soLines, inventory.StockMovementDAOMock{
		FindForUpdateTxFunc: func(_ context.Context, _ *gorm.DB, moveID uint64) (*inventory.StockMovement, error) {
			return &inventory.StockMovement{Base: model.Base{ID: moveID}, State: inventory.MovementStateConfirmed}, nil
		},
	}, locations, resolver, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5,"date":"2026-01-15"}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_ReturnsNotFound(t *testing.T) {
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, nil
			},
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, poOrders, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_MapsAlreadyReceived(t *testing.T) {
	po := &procurement.PurchaseOrder{
		Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)),
		State: procurement.PurchaseOrderStateDone,
	}
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return po, nil
			},
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, poOrders, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_MapsNoLines(t *testing.T) {
	po := &procurement.PurchaseOrder{
		Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)),
		State: procurement.PurchaseOrderStateConfirmed, DestLocationID: helper.Ptr(uint64(30)),
	}
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return po, nil
			},
		},
	}
	links := interorganization.DropshipLinkDAOMock{
		ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
			return []*interorganization.DropshipLink{}, nil
		},
	}
	svc := dropshipServiceForTest(links, purchaseOrderCreatorMock{}, poOrders, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_MapsSourceNotFound(t *testing.T) {
	po := &procurement.PurchaseOrder{
		Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)),
		State: procurement.PurchaseOrderStateConfirmed, DestLocationID: helper.Ptr(uint64(30)),
	}
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return po, nil
			},
		},
	}
	links := interorganization.DropshipLinkDAOMock{
		ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
			return []*interorganization.DropshipLink{{Base: model.Base{ID: 1}, SaleOrderLineID: 1, PurchaseOrderLineID: 1}}, nil
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
				return nil, nil
			},
		},
	}
	svc := dropshipServiceForTest(links, purchaseOrderCreatorMock{}, poOrders, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, soLines, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_MapsSupplierMissing(t *testing.T) {
	po := &procurement.PurchaseOrder{
		Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)),
		State: procurement.PurchaseOrderStateConfirmed, DestLocationID: helper.Ptr(uint64(30)),
	}
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return po, nil
			},
		},
	}
	links := interorganization.DropshipLinkDAOMock{
		ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
			return []*interorganization.DropshipLink{{Base: model.Base{ID: 1}, SaleOrderLineID: 1, PurchaseOrderLineID: 1}}, nil
		},
	}
	so := saleOrderConfirmed(1)
	soLines := sales.SaleOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
				return &sales.SaleOrderLine{Base: model.Base{ID: 1}, OrderID: 1}, nil
			},
		},
	}
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return so, nil
			},
		},
	}
	svc := dropshipServiceForTest(links, purchaseOrderCreatorMock{}, poOrders, procurement.PurchaseOrderLineDAOMock{}, soOrders, soLines, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_MapsDestinationMiss(t *testing.T) {
	po := &procurement.PurchaseOrder{
		Base: model.Base{ID: 1}, OrganizationID: helper.Ptr(uint64(10)),
		State: procurement.PurchaseOrderStateConfirmed,
	}
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return po, nil
			},
		},
	}
	links := interorganization.DropshipLinkDAOMock{
		ListByPurchaseOrderFunc: func(_ context.Context, _ uint64) ([]*interorganization.DropshipLink, error) {
			return []*interorganization.DropshipLink{{Base: model.Base{ID: 1}, SaleOrderLineID: 1, PurchaseOrderLineID: 1}}, nil
		},
	}
	so := saleOrderConfirmed(1)
	soLines := sales.SaleOrderLineDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrderLine]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrderLine, error) {
				return &sales.SaleOrderLine{Base: model.Base{ID: 1}, OrderID: 1}, nil
			},
		},
	}
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return so, nil
			},
		},
	}
	locations := inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 20}, Usage: "supplier"}}}, nil
			},
		},
	}
	svc := dropshipServiceForTest(links, purchaseOrderCreatorMock{}, poOrders, procurement.PurchaseOrderLineDAOMock{}, soOrders, soLines, inventory.StockMovementDAOMock{}, locations, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_RejectsValidation(t *testing.T) {
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_RejectsInvalidDate(t *testing.T) {
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5,"date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDropShipHandler_Receive_ReturnsServerError(t *testing.T) {
	poOrders := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := dropshipServiceForTest(interorganization.DropshipLinkDAOMock{}, purchaseOrderCreatorMock{}, poOrders, procurement.PurchaseOrderLineDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, txMock{})
	h := NewDropShipHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"journal_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dropship-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_CreatesRule(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{}}, nil
			},
			CreateFunc: func(_ context.Context, rule *interorganization.InterorganizationRule) (*interorganization.InterorganizationRule, error) {
				rule.ID = 1
				return rule, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20,"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_AllowsExplicitAutoMirror(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{}}, nil
			},
			CreateFunc: func(_ context.Context, rule *interorganization.InterorganizationRule) (*interorganization.InterorganizationRule, error) {
				rule.ID = 1
				return rule, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20,"auto_mirror":false}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_ReturnsBadRequest(t *testing.T) {
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_MapsRequired(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_MapsSameOrganization(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":10,"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_MapsContactsRequired(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_MapsDuplicate(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{ToOrganizationID: helper.Ptr(uint64(20))}}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20,"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	body := `{"to_organization_id":20,"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_CreateRule_ReturnsServerError(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20,"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-rules/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_ListRules_ListsRules(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: model.Base{ID: 1}}}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/interorganization-rules/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestInterorganizationHandler_ListRules_ReturnsServerError(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/interorganization-rules/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_ListRules_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/interorganization-rules/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_UpdateRule_UpdatesRule(t *testing.T) {
	existing := &interorganization.InterorganizationRule{
		Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(10)), AutoMirror: true,
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return existing, nil
			},
			UpdateFunc: func(_ context.Context, rule *interorganization.InterorganizationRule) (*interorganization.InterorganizationRule, error) {
				return rule, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPut, "/interorganization-rules/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestInterorganizationHandler_UpdateRule_OverridesAutoMirror(t *testing.T) {
	existing := &interorganization.InterorganizationRule{
		Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(10)), AutoMirror: true,
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return existing, nil
			},
			UpdateFunc: func(_ context.Context, rule *interorganization.InterorganizationRule) (*interorganization.InterorganizationRule, error) {
				return rule, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"auto_mirror":false,"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPut, "/interorganization-rules/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestInterorganizationHandler_UpdateRule_ReturnsBadRequest(t *testing.T) {
	existing := &interorganization.InterorganizationRule{
		Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(10)), AutoMirror: true,
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return existing, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{`
	resp, err := doRequest(app, http.MethodPut, "/interorganization-rules/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestInterorganizationHandler_UpdateRule_ReturnsNotFound(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return nil, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/interorganization-rules/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestInterorganizationHandler_UpdateRule_ReturnsNotFoundForForeignOrganization(t *testing.T) {
	existing := &interorganization.InterorganizationRule{
		Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(99)), AutoMirror: true,
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return existing, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/interorganization-rules/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestInterorganizationHandler_UpdateRule_ReturnsServerError(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/interorganization-rules/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_UpdateRule_ReturnsServerErrorOnUpdate(t *testing.T) {
	existing := &interorganization.InterorganizationRule{
		Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(10)), AutoMirror: true,
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return existing, nil
			},
			UpdateFunc: func(_ context.Context, _ *interorganization.InterorganizationRule) (*interorganization.InterorganizationRule, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"supplier_contact_id":5,"customer_contact_id":6}`
	resp, err := doRequest(app, http.MethodPut, "/interorganization-rules/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_DeleteRule_DeletesRule(t *testing.T) {
	existing := &interorganization.InterorganizationRule{
		Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(10)), AutoMirror: true,
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return existing, nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodDelete, "/interorganization-rules/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestInterorganizationHandler_DeleteRule_ReturnsNotFound(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return nil, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodDelete, "/interorganization-rules/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestInterorganizationHandler_DeleteRule_ReturnsServerError(t *testing.T) {
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodDelete, "/interorganization-rules/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_DeleteRule_ReturnsServerErrorOnDelete(t *testing.T) {
	existing := &interorganization.InterorganizationRule{
		Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(10)), AutoMirror: true,
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			FindFunc: func(_ context.Context, _ uint64) (*interorganization.InterorganizationRule, error) {
				return existing, nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodDelete, "/interorganization-rules/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_MirrorsOrder(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: model.Base{ID: 1}, FromOrganizationID: helper.Ptr(uint64(10)), ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true, SupplierContactID: helper.Ptr(uint64(5)), CustomerContactID: helper.Ptr(uint64(6))}}}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}, ItemID: helper.Ptr(uint64(7)), QtyOrdered: 2, UnitPrice: 100}}, nil
		},
	}
	poCreate := purchaseOrderCreatorMock{
		CreateFunc: func(_ context.Context, order *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			order.ID = 2
			order.AmountTotal = 200
			return order, nil
		},
	}
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			CreateFunc: func(_ context.Context, transaction *interorganization.InterorganizationTransaction) (*interorganization.InterorganizationTransaction, error) {
				transaction.ID = 1
				return transaction, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, trans, poCreate, soOrders, soLines)
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_RejectsValidation(t *testing.T) {
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_ReturnsNotFound(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, nil
			},
		},
	}
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, soOrders, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_MapsNotPosted(t *testing.T) {
	so := saleOrderConfirmed(1)
	so.State = sales.OrderStateDraft
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return so, nil
			},
		},
	}
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, soOrders, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_MapsRuleMissing(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, soOrders, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_MapsRuleDisabled(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: model.Base{ID: 1}, ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: false, SupplierContactID: helper.Ptr(uint64(5))}}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, soOrders, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_MapsContactsRequired(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: model.Base{ID: 1}, ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true}}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, soOrders, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_MapsNoLines(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return saleOrderConfirmed(1), nil
			},
		},
	}
	rules := interorganization.InterorganizationRuleDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationRule]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationRule], error) {
				return &query.Page[interorganization.InterorganizationRule]{Items: []*interorganization.InterorganizationRule{{Base: model.Base{ID: 1}, ToOrganizationID: helper.Ptr(uint64(20)), AutoMirror: true, SupplierContactID: helper.Ptr(uint64(5)), CustomerContactID: helper.Ptr(uint64(6))}}}, nil
			},
		},
	}
	soLines := sales.SaleOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*sales.SaleOrderLine, error) {
			return []*sales.SaleOrderLine{{Base: model.Base{ID: 1}}}, nil
		},
	}
	svc := interorganizationServiceForTest(rules, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, soOrders, soLines)
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestInterorganizationHandler_MirrorSaleOrder_ReturnsServerError(t *testing.T) {
	soOrders := sales.SaleOrderDAOMock{
		CRUDMock: dao.CRUDMock[sales.SaleOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*sales.SaleOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, soOrders, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	body := `{"to_organization_id":20}`
	resp, err := doRequest(app, http.MethodPost, "/interorganization-transactions/mirror-sale-order/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_ListTransactions_ListsTransactions(t *testing.T) {
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return &query.Page[interorganization.InterorganizationTransaction]{Items: []*interorganization.InterorganizationTransaction{{Base: model.Base{ID: 1}}}}, nil
			},
		},
	}
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, trans, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/interorganization-transactions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestInterorganizationHandler_ListTransactions_ReturnsServerError(t *testing.T) {
	trans := interorganization.InterorganizationTransactionDAOMock{
		CRUDMock: dao.CRUDMock[interorganization.InterorganizationTransaction]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[interorganization.InterorganizationTransaction], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, trans, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestApp(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/interorganization-transactions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestInterorganizationHandler_ListTransactions_ReturnsUnprocessableWithoutTenant(t *testing.T) {
	svc := interorganizationServiceForTest(interorganization.InterorganizationRuleDAOMock{}, interorganization.InterorganizationTransactionDAOMock{}, purchaseOrderCreatorMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{})
	h := NewInterorganizationHandler(svc)
	app := interorganizationTestAppNoTenant(t, h.Register)

	resp, err := doRequest(app, http.MethodGet, "/interorganization-transactions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

type purchaseOrderCreatorMock struct {
	CreateFunc func(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error)
}

func (m purchaseOrderCreatorMock) Create(ctx context.Context, order *procurement.PurchaseOrder, lines []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, order, lines)
	}
	return order, nil
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

type rateSourceMock struct {
	RateFunc func(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error)
}

func (m rateSourceMock) Rate(ctx context.Context, currencyCode string, organizationID uint64, rateType amount.RateType, date time.Time) (amount.Amount, error) {
	if m.RateFunc != nil {
		return m.RateFunc(ctx, currencyCode, organizationID, rateType, date)
	}
	return amount.Zero(), amount.ErrInvalidRate
}

type organizationLookupMock struct {
	FindFunc func(ctx context.Context, id uint64) (*reference.Organization, error)
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.Organization], error)
}

func (m organizationLookupMock) Find(ctx context.Context, id uint64) (*reference.Organization, error) {
	if m.FindFunc != nil {
		return m.FindFunc(ctx, id)
	}
	return nil, nil
}

func (m organizationLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.Organization], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.Organization]{Items: []*reference.Organization{}}, nil
}

type accountLookupMock struct {
	ListFunc func(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error)
}

func (m accountLookupMock) List(ctx context.Context, q *query.Query) (*query.Page[reference.Account], error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, q)
	}
	return &query.Page[reference.Account]{Items: []*reference.Account{}}, nil
}

type taxPeriodDAOMock struct {
	dao.CRUDMock[accounting.TaxPeriod]
	FindByDateFunc         func(ctx context.Context, organizationID uint64, date time.Time) (*accounting.TaxPeriod, error)
	ListByOrganizationFunc func(ctx context.Context, organizationID uint64) ([]*accounting.TaxPeriod, error)
	UpdateTxFunc           func(ctx context.Context, tx *gorm.DB, period *accounting.TaxPeriod) (*accounting.TaxPeriod, error)
}

func (m taxPeriodDAOMock) UpdateTx(ctx context.Context, tx *gorm.DB, period *accounting.TaxPeriod) (*accounting.TaxPeriod, error) {
	if m.UpdateTxFunc != nil {
		return m.UpdateTxFunc(ctx, tx, period)
	}
	return period, nil
}

func (m taxPeriodDAOMock) FindByDate(ctx context.Context, organizationID uint64, date time.Time) (*accounting.TaxPeriod, error) {
	if m.FindByDateFunc != nil {
		return m.FindByDateFunc(ctx, organizationID, date)
	}
	return nil, nil
}

func (m taxPeriodDAOMock) ListByOrganization(ctx context.Context, organizationID uint64) ([]*accounting.TaxPeriod, error) {
	if m.ListByOrganizationFunc != nil {
		return m.ListByOrganizationFunc(ctx, organizationID)
	}
	return nil, nil
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}
