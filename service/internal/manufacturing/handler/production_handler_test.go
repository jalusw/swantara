package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func productionTestApp(t *testing.T, svc manufacturing.ProductionService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewProductionHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func productionLocationMock() inventory.StockLocationDAOMock {
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return referenceLocation(30, "production", 10), nil
			},
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{referenceLocation(30, "production", 10)}}, nil
			},
		},
	}
}

func productionServiceFor(
	orders manufacturing.ProductionOrderDAOMock,
	components manufacturing.ConsumedMaterialDAOMock,
	workOrders manufacturing.ShopTaskDAOMock,
	operations manufacturing.ProductionStepDAOMock,
	workCenters dao.CRUDMock[reference.WorkCenter],
	locations inventory.StockLocationDAOMock,
	movements inventory.StockMovementDAOMock,
	layers inventory.CostLayerDAOMock,
	resolver inventory.ItemResolverMock,
	poster inventory.PosterMock,
	lines accounting.JournalLineDAOMock,
	sequences sequence.DAOMock,
) manufacturing.ProductionService {
	return manufacturing.NewTestProductionService(orders, components, workOrders, operations, workCenters, locations, movements, layers, resolver, poster, lines, sequences)
}

func productionResolvedItem() inventory.ResolvedItem {
	return inventory.ResolvedItem{
		StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400, StockInputAccountID: 401, CogsAccountID: 402, StockOutputAccountID: 403},
		Tracking:      "none",
		StandardCost:  5,
	}
}

func productionMoveMock(organizationID, src, dst uint64, itemID, qty float64) inventory.StockMovementDAOMock {
	return inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 100
				movement.State = inventory.MovementStateDraft
				return movement, nil
			},
			FindFunc: func(_ context.Context, moveID uint64) (*inventory.StockMovement, error) {
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, ItemID: uint64(itemID), Qty: qty, SrcLocationID: src, DstLocationID: dst, State: inventory.MovementStateDraft, OrganizationID: &organizationID}, nil
			},
		},
		ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement.State = inventory.MovementStateDone
			return movement, nil
		},
	}
}

func productionLayerMock() inventory.CostLayerDAOMock {
	return inventory.CostLayerDAOMock{
		ListOpenByItemFunc: func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: model.Base{ID: 1}, Quantity: 50, UnitCost: helper.Ptr(2.0), RemainingQty: 50, RemainingValue: 100}}, nil
		},
		CreateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			layer.ID = 90
			return layer, nil
		},
		UpdateTxFunc: func(_ context.Context, _ *gorm.DB, layer *inventory.CostLayer) (*inventory.CostLayer, error) {
			return layer, nil
		},
	}
}

func productionPosterMock() inventory.PosterMock {
	return inventory.PosterMock{
		PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
}

func TestProductionHandler_Start_StartsOrder(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), State: manufacturing.ProductionOrderStatePlanned}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_Start_ReturnsNotFound(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return nil, nil
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductionHandler_Start_MapsStateConflict(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), State: manufacturing.ProductionOrderStateDraft}, nil
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestProductionHandler_Start_RejectsInvalidID(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_Start_ReturnsServerError(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), State: manufacturing.ProductionOrderStatePlanned}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/start", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionHandler_GenerateShopTasks_Generates(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), RecipeID: uint64Ptr(1), State: manufacturing.ProductionOrderStatePlanned}, nil
			},
		},
	}
	operations := manufacturing.ProductionStepDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ProductionStep, error) {
			return []*manufacturing.ProductionStep{{Base: model.Base{ID: 1}, WorkCenterID: uint64Ptr(7), Sequence: 1, SetupMinutes: 5, TimeMinutes: 10}}, nil
		},
	}
	workCenters := dao.CRUDMock[reference.WorkCenter]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) {
			return &reference.WorkCenter{Base: model.Base{ID: 7}}, nil
		},
	}
	workOrders := manufacturing.ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ShopTask]{
			CreateFunc: func(_ context.Context, wo *manufacturing.ShopTask) (*manufacturing.ShopTask, error) {
				wo.ID = 1
				return wo, nil
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, workOrders, operations, workCenters, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/shop-tasks", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_GenerateShopTasks_ReturnsServerError(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), RecipeID: uint64Ptr(1), State: manufacturing.ProductionOrderStatePlanned}, nil
			},
		},
	}
	operations := manufacturing.ProductionStepDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ProductionStep, error) {
			return []*manufacturing.ProductionStep{{Base: model.Base{ID: 1}, WorkCenterID: uint64Ptr(7), Sequence: 1, SetupMinutes: 5, TimeMinutes: 10}}, nil
		},
	}
	workCenters := dao.CRUDMock[reference.WorkCenter]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) {
			return &reference.WorkCenter{Base: model.Base{ID: 7}}, nil
		},
	}
	workOrders := manufacturing.ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ShopTask]{
			CreateFunc: func(_ context.Context, wo *manufacturing.ShopTask) (*manufacturing.ShopTask, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, workOrders, operations, workCenters, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/shop-tasks", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionHandler_GenerateShopTasks_RejectsInvalidID(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/shop-tasks", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_ListShopTasks_ReturnsShopTasks(t *testing.T) {
	workOrders := manufacturing.ShopTaskDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ShopTask, error) {
			return []*manufacturing.ShopTask{{Base: model.Base{ID: 1}, ProductionOrderID: 1, WorkCenterID: 7, State: manufacturing.ShopTaskStatePlanned}}, nil
		},
	}
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, workOrders, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1/shop-tasks", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_ListShopTasks_RejectsInvalidID(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/production-orders/abc/shop-tasks", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_ListShopTasks_ReturnsServerError(t *testing.T) {
	workOrders := manufacturing.ShopTaskDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ShopTask, error) {
			return nil, errors.New("db down")
		},
	}
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, workOrders, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1/shop-tasks", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionHandler_ListShopTasks_ExportsCSV(t *testing.T) {
	workOrders := manufacturing.ShopTaskDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ShopTask, error) {
			return []*manufacturing.ShopTask{{Base: model.Base{ID: 1}, ProductionOrderID: 1, WorkCenterID: 7, State: manufacturing.ShopTaskStatePlanned}}, nil
		},
	}
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, workOrders, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1/shop-tasks?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_Consume_ConsumesMaterial(t *testing.T) {
	organizationID := uint64(10)
	src := uint64(10)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, SrcLocationID: &src, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ConsumedMaterial]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ConsumedMaterial, error) {
				return &manufacturing.ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 22, QtyConsumed: 0}, nil
			},
			UpdateFunc: func(_ context.Context, comp *manufacturing.ConsumedMaterial) (*manufacturing.ConsumedMaterial, error) {
				return comp, nil
			},
		},
	}
	movements := productionMoveMock(organizationID, src, 30, 300, 2)
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return productionResolvedItem(), nil
		},
	}
	svc := productionServiceFor(orders, components, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), movements, productionLayerMock(), resolver, productionPosterMock(), accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"component_id":5,"qty":2,"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/consume", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_Consume_RejectsValidation(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"component_id":5,"qty":2}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/consume", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_Consume_RejectsInvalidID(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"component_id":5,"qty":2,"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/consume", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_Consume_ReturnsServerError(t *testing.T) {
	organizationID := uint64(10)
	src := uint64(10)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, SrcLocationID: &src, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ConsumedMaterial]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ConsumedMaterial, error) {
				return &manufacturing.ConsumedMaterial{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 22, QtyConsumed: 0}, nil
			},
			UpdateFunc: func(_ context.Context, comp *manufacturing.ConsumedMaterial) (*manufacturing.ConsumedMaterial, error) {
				return nil, errors.New("db down")
			},
		},
	}
	movements := productionMoveMock(organizationID, src, 30, 300, 2)
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return productionResolvedItem(), nil
		},
	}
	svc := productionServiceFor(orders, components, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), movements, productionLayerMock(), resolver, productionPosterMock(), accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"component_id":5,"qty":2,"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/consume", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionHandler_Produce_ProducesGoods(t *testing.T) {
	organizationID := uint64(10)
	dst := uint64(40)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 200, QtyToProduce: 10, QtyProduced: 0, DstLocationID: &dst, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	movements := productionMoveMock(organizationID, 30, dst, 200, 4)
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return productionResolvedItem(), nil
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), movements, productionLayerMock(), resolver, productionPosterMock(), accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"qty":4,"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/produce", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_Produce_RejectsValidation(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"qty":4}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/produce", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_Produce_RejectsInvalidID(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"qty":4,"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/produce", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_Produce_ReturnsServerError(t *testing.T) {
	organizationID := uint64(10)
	dst := uint64(40)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, ItemID: 200, QtyToProduce: 10, QtyProduced: 0, DstLocationID: &dst, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	movements := productionMoveMock(organizationID, 30, dst, 200, 4)
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return productionResolvedItem(), nil
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, productionLocationMock(), movements, productionLayerMock(), resolver, productionPosterMock(), accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"qty":4,"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/produce", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionHandler_RecordLabor_RecordsLabor(t *testing.T) {
	organizationID := uint64(10)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
		},
	}
	workOrders := manufacturing.ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ShopTask]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ShopTask, error) {
				return &manufacturing.ShopTask{Base: model.Base{ID: 9}, ProductionOrderID: 1, WorkCenterID: 7, State: manufacturing.ShopTaskStateInProgress}, nil
			},
			UpdateFunc: func(_ context.Context, wo *manufacturing.ShopTask) (*manufacturing.ShopTask, error) {
				return wo, nil
			},
		},
	}
	workCenters := dao.CRUDMock[reference.WorkCenter]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.WorkCenter, error) {
			return &reference.WorkCenter{Base: model.Base{ID: 7}, CostPerHour: helper.Ptr(60.0)}, nil
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, workOrders, manufacturing.ProductionStepDAOMock{}, workCenters, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, productionPosterMock(), accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"minutes":30,"journal_id":500,"wip_account_id":600,"applied_labor_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/shop-tasks/9/labor", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_RecordLabor_RejectsValidation(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"minutes":30}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/shop-tasks/9/labor", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_RecordLabor_RejectsInvalidIDs(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"minutes":30,"journal_id":500,"wip_account_id":600,"applied_labor_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/shop-tasks/9/labor", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_RecordLabor_MapsShopTaskMismatch(t *testing.T) {
	organizationID := uint64(10)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
		},
	}
	workOrders := manufacturing.ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ShopTask]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ShopTask, error) {
				return &manufacturing.ShopTask{Base: model.Base{ID: 9}, ProductionOrderID: 99, WorkCenterID: 7}, nil
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, workOrders, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"minutes":30,"journal_id":500,"wip_account_id":600,"applied_labor_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/shop-tasks/9/labor", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_RecordLabor_ReturnsNotFound(t *testing.T) {
	organizationID := uint64(10)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
		},
	}
	workOrders := manufacturing.ShopTaskDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ShopTask]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ShopTask, error) {
				return nil, nil
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, workOrders, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"minutes":30,"journal_id":500,"wip_account_id":600,"applied_labor_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/shop-tasks/9/labor", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductionHandler_SettleVariance_Settles(t *testing.T) {
	organizationID := uint64(10)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: manufacturing.ProductionOrderStateInProgress}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	lines := accounting.JournalLineDAOMock{
		BalanceByOriginAndAccountFunc: func(_ context.Context, _ string, _ uint64, _ uint64) (float64, error) {
			return 0, nil
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, lines, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"journal_id":500,"wip_account_id":600,"variance_account_id":800}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/settle", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionHandler_SettleVariance_RejectsValidation(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"journal_id":500}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/settle", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_SettleVariance_RejectsInvalidID(t *testing.T) {
	svc := productionServiceFor(manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"journal_id":500,"wip_account_id":600,"variance_account_id":800}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/settle", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionHandler_SettleVariance_MapsStateConflict(t *testing.T) {
	organizationID := uint64(10)
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: &organizationID, State: manufacturing.ProductionOrderStateDraft}, nil
			},
		},
	}
	svc := productionServiceFor(orders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.ShopTaskDAOMock{}, manufacturing.ProductionStepDAOMock{}, dao.CRUDMock[reference.WorkCenter]{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, accounting.JournalLineDAOMock{}, manufacturing.DefaultShopTaskSequenceMock())
	app := productionTestApp(t, svc)

	body := `{"journal_id":500,"wip_account_id":600,"variance_account_id":800}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/settle", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}
