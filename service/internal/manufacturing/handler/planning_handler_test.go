package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"github.com/jalusw/swantara/apps/service/internal/sales"
)

func mrpTestApp(t *testing.T, svc manufacturing.PlanningService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewPlanningHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func mrpLedger() inventory.LedgerService {
	return inventory.NewLedgerService(inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.TransactionerMock{})
}

func mrpReorder() inventory.ReorderService {
	return inventory.NewReorderService(inventory.ReorderRuleDAOMock{}, mrpLedger(), inventory.ItemResolverMock{})
}

func planningServiceFor(
	runs manufacturing.PlanningRunDAOMock,
	demands manufacturing.PlanningNeedDAOMock,
	planned manufacturing.PlannedSupplyDAOMock,
	forecasts manufacturing.DemandPlanDAOMock,
	soOrders sales.SaleOrderDAOMock,
	soLines sales.SaleOrderLineDAOMock,
	poOrders procurement.PurchaseOrderDAOMock,
	poLines procurement.PurchaseOrderLineDAOMock,
	orders manufacturing.ProductionOrderDAOMock,
	recipes manufacturing.RecipeDAOMock,
	recipeLines manufacturing.RecipeLineDAOMock,
	variants products.ItemVariantDAOMock,
	templates products.ItemDAOMock,
	locations inventory.StockLocationDAOMock,
) manufacturing.PlanningService {
	return manufacturing.NewTestPlanningService(runs, demands, planned, forecasts, soOrders, soLines, poOrders, poLines, orders, mrpReorder(), mrpLedger(), recipes, recipeLines, variants, templates, locations, manufacturing.PurchasePlannerMock{}, manufacturing.ManufacturePlannerMock{}, manufacturing.TransferPlannerMock{})
}

func planningServiceWith(
	runs manufacturing.PlanningRunDAOMock,
	demands manufacturing.PlanningNeedDAOMock,
	planned manufacturing.PlannedSupplyDAOMock,
	forecasts manufacturing.DemandPlanDAOMock,
) manufacturing.PlanningService {
	return planningServiceFor(runs, demands, planned, forecasts, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
}

func mrpDefaultService() manufacturing.PlanningService {
	return planningServiceWith(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{})
}

func mrpDefaultLocations() inventory.StockLocationDAOMock {
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{
					referenceLocation(30, "production", 10),
					referenceLocation(31, "internal", 10),
				}}, nil
			},
		},
	}
}

func TestPlanningHandler_Run_ReturnsRun(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			CreateFunc: func(_ context.Context, run *manufacturing.PlanningRun) (*manufacturing.PlanningRun, error) {
				run.ID = 1
				return run, nil
			},
			UpdateFunc: func(_ context.Context, run *manufacturing.PlanningRun) (*manufacturing.PlanningRun, error) {
				return run, nil
			},
		},
	}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/runs", `{"horizon_days":30,"run_date":"2026-08-01"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_Run_RejectsInvalidBody(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			CreateFunc: func(_ context.Context, run *manufacturing.PlanningRun) (*manufacturing.PlanningRun, error) {
				run.ID = 1
				return run, nil
			},
		},
	}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/runs", `{"horizon_days":`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestPlanningHandler_Run_RejectsZeroHorizon(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			CreateFunc: func(_ context.Context, run *manufacturing.PlanningRun) (*manufacturing.PlanningRun, error) {
				run.ID = 1
				return run, nil
			},
		},
	}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/runs", `{"horizon_days":0}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_Run_RejectsMissingOrganization(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := fiber.New()
	h := NewPlanningHandler(svc)
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/planning/runs", `{"horizon_days":30}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_Run_ReturnsServerError(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			CreateFunc: func(_ context.Context, run *manufacturing.PlanningRun) (*manufacturing.PlanningRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/runs", `{"horizon_days":30}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPlanningHandler_Get_ReturnsRun(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlanningRun, error) {
				return &manufacturing.PlanningRun{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), HorizonDays: 30, State: manufacturing.PlanningRunStateDone}, nil
			},
		},
	}
	demands := manufacturing.PlanningNeedDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*manufacturing.PlanningNeed, error) {
			return []*manufacturing.PlanningNeed{{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 100, Qty: 5}}, nil
		},
	}
	planned := manufacturing.PlannedSupplyDAOMock{
		ListByRunFunc: func(_ context.Context, _ uint64) ([]*manufacturing.PlannedSupply, error) {
			return []*manufacturing.PlannedSupply{{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 100, Type: manufacturing.PlannedSupplyTypePurchase, Qty: 5}}, nil
		},
	}
	app := mrpTestApp(t, planningServiceWith(runs, demands, planned, manufacturing.DemandPlanDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/planning/runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_Get_ReturnsNotFound(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{}
	app := mrpTestApp(t, planningServiceWith(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/planning/runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPlanningHandler_Get_RejectsInvalidID(t *testing.T) {
	app := mrpTestApp(t, mrpDefaultService())

	resp, err := doRequest(app, http.MethodGet, "/planning/runs/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_Get_ReturnsServerError(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlanningRun, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := mrpTestApp(t, planningServiceWith(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/planning/runs/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPlanningHandler_List_ReturnsRuns(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.PlanningRun], error) {
				return &query.Page[manufacturing.PlanningRun]{Items: []*manufacturing.PlanningRun{{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), HorizonDays: 30, State: manufacturing.PlanningRunStateDone}}, Count: 1}, nil
			},
		},
	}
	app := mrpTestApp(t, planningServiceWith(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/planning/runs", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := mrpTestApp(t, mrpDefaultService())

	resp, err := doRequest(app, http.MethodGet, "/planning/runs?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{}
	app := fiber.New()
	h := NewPlanningHandler(planningServiceWith(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}))
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/planning/runs", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestPlanningHandler_List_ReturnsServerError(t *testing.T) {
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.PlanningRun], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := mrpTestApp(t, planningServiceWith(runs, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, manufacturing.DemandPlanDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/planning/runs", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_ConfirmsPurchase(t *testing.T) {
	warehouseID := uint64(1)
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlanningRun, error) {
				return &manufacturing.PlanningRun{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10)}, nil
			},
		},
	}
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 100, WarehouseID: &warehouseID, Type: manufacturing.PlannedSupplyTypePurchase, Qty: 5}, nil
			},
			UpdateFunc: func(_ context.Context, order *manufacturing.PlannedSupply) (*manufacturing.PlannedSupply, error) {
				return order, nil
			},
		},
	}
	variants := products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, _ uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: 100}, ItemID: 1}, nil
			},
		},
	}
	templates := products.ItemDAOMock{
		CRUDMock: dao.CRUDMock[products.Item]{
			FindFunc: func(_ context.Context, _ uint64) (*products.Item, error) {
				return &products.Item{Base: model.Base{ID: 1}, StandardCost: 5}, nil
			},
		},
	}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, variants, templates, mrpDefaultLocations())
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{"supplier_id":5}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_ConfirmsManufacture(t *testing.T) {
	warehouseID := uint64(1)
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlanningRun, error) {
				return &manufacturing.PlanningRun{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10)}, nil
			},
		},
	}
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 200, WarehouseID: &warehouseID, Type: manufacturing.PlannedSupplyTypeManufacture, Qty: 5}, nil
			},
			UpdateFunc: func(_ context.Context, order *manufacturing.PlannedSupply) (*manufacturing.PlannedSupply, error) {
				return order, nil
			},
		},
	}
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return &manufacturing.Recipe{Base: model.Base{ID: 1}, ItemID: 200, Type: manufacturing.RecipeTypeManufacture, Active: true}, nil
			},
		},
		ListByItemFunc: func(_ context.Context, _ uint64) ([]*manufacturing.Recipe, error) {
			return []*manufacturing.Recipe{{Base: model.Base{ID: 1}, ItemID: 200, Type: manufacturing.RecipeTypeManufacture, Active: true}}, nil
		},
	}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, recipes, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, mrpDefaultLocations())
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_ConfirmsTransfer(t *testing.T) {
	srcWarehouseID := uint64(2)
	warehouseID := uint64(1)
	runs := manufacturing.PlanningRunDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlanningRun]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlanningRun, error) {
				return &manufacturing.PlanningRun{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10)}, nil
			},
		},
	}
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, PlanningRunID: 1, ItemID: 300, SrcWarehouseID: &srcWarehouseID, WarehouseID: &warehouseID, Type: manufacturing.PlannedSupplyTypeTransfer, Qty: 5}, nil
			},
			UpdateFunc: func(_ context.Context, order *manufacturing.PlannedSupply) (*manufacturing.PlannedSupply, error) {
				return order, nil
			},
		},
	}
	svc := planningServiceFor(runs, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_MapsAlreadyConfirmed(t *testing.T) {
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, Type: manufacturing.PlannedSupplyTypePurchase, Confirmed: true}, nil
			},
		},
	}
	svc := planningServiceFor(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{"supplier_id":5}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_ReturnsNotFound(t *testing.T) {
	planned := manufacturing.PlannedSupplyDAOMock{}
	svc := planningServiceFor(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{"supplier_id":5}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_RequiresVendor(t *testing.T) {
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, Type: manufacturing.PlannedSupplyTypePurchase}, nil
			},
		},
	}
	svc := planningServiceFor(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{"supplier_id":0}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_RequiresRecipe(t *testing.T) {
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, Type: manufacturing.PlannedSupplyTypeManufacture}, nil
			},
		},
	}
	svc := planningServiceFor(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_MapsUnsupportedType(t *testing.T) {
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, Type: "bogus"}, nil
			},
		},
	}
	svc := planningServiceFor(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_RejectsInvalidID(t *testing.T) {
	app := mrpTestApp(t, mrpDefaultService())

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/abc/confirm", `{"supplier_id":5}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_RejectsInvalidBody(t *testing.T) {
	app := mrpTestApp(t, mrpDefaultService())

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{"supplier_id":`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestPlanningHandler_Confirm_MapsTransferWarehouseError(t *testing.T) {
	planned := manufacturing.PlannedSupplyDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.PlannedSupply]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.PlannedSupply, error) {
				return &manufacturing.PlannedSupply{Base: model.Base{ID: 1}, Type: manufacturing.PlannedSupplyTypeTransfer}, nil
			},
		},
	}
	svc := planningServiceFor(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, planned, manufacturing.DemandPlanDAOMock{}, sales.SaleOrderDAOMock{}, sales.SaleOrderLineDAOMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, manufacturing.ProductionOrderDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.RecipeLineDAOMock{}, products.ItemVariantDAOMock{}, products.ItemDAOMock{}, inventory.StockLocationDAOMock{})
	app := mrpTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/planning/planned-orders/1/confirm", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPlanningHandler_CreateForecast_CreatesForecast(t *testing.T) {
	forecasts := manufacturing.DemandPlanDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.DemandPlan]{
			CreateFunc: func(_ context.Context, forecast *manufacturing.DemandPlan) (*manufacturing.DemandPlan, error) {
				forecast.ID = 1
				return forecast, nil
			},
		},
	}
	app := mrpTestApp(t, planningServiceWith(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, forecasts))

	body := `{"item_id":100,"period_start":"2026-08-01","period_end":"2026-08-31","forecast_qty":50}`
	resp, err := doRequest(app, http.MethodPost, "/planning/forecasts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_CreateForecast_RejectsInvalidBody(t *testing.T) {
	app := mrpTestApp(t, mrpDefaultService())

	resp, err := doRequest(app, http.MethodPost, "/planning/forecasts", `{"item_id":`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestPlanningHandler_CreateForecast_ReturnsServerError(t *testing.T) {
	forecasts := manufacturing.DemandPlanDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.DemandPlan]{
			CreateFunc: func(_ context.Context, forecast *manufacturing.DemandPlan) (*manufacturing.DemandPlan, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := mrpTestApp(t, planningServiceWith(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, forecasts))

	body := `{"item_id":100,"period_start":"2026-08-01","period_end":"2026-08-31","forecast_qty":50}`
	resp, err := doRequest(app, http.MethodPost, "/planning/forecasts", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestPlanningHandler_ListForecasts_ReturnsForecasts(t *testing.T) {
	forecasts := manufacturing.DemandPlanDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.DemandPlan]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.DemandPlan], error) {
				return &query.Page[manufacturing.DemandPlan]{Items: []*manufacturing.DemandPlan{{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), ItemID: 100, ForecastQty: 50}}, Count: 1}, nil
			},
		},
	}
	app := mrpTestApp(t, planningServiceWith(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, forecasts))

	resp, err := doRequest(app, http.MethodGet, "/planning/forecasts", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestPlanningHandler_ListForecasts_RejectsInvalidQuery(t *testing.T) {
	app := mrpTestApp(t, mrpDefaultService())

	resp, err := doRequest(app, http.MethodGet, "/planning/forecasts?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestPlanningHandler_ListForecasts_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	forecasts := manufacturing.DemandPlanDAOMock{}
	app := fiber.New()
	h := NewPlanningHandler(planningServiceWith(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, forecasts))
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/planning/forecasts", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestPlanningHandler_ListForecasts_ReturnsServerError(t *testing.T) {
	forecasts := manufacturing.DemandPlanDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.DemandPlan]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.DemandPlan], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := mrpTestApp(t, planningServiceWith(manufacturing.PlanningRunDAOMock{}, manufacturing.PlanningNeedDAOMock{}, manufacturing.PlannedSupplyDAOMock{}, forecasts))

	resp, err := doRequest(app, http.MethodGet, "/planning/forecasts", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

var _ = accounting.JournalLineDAOMock{}
