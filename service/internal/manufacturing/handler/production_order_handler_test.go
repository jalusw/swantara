package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/kernel/sequence"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func orderTestApp(t *testing.T, svc manufacturing.ProductionOrderService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewProductionOrderHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func sampleMO() *manufacturing.ProductionOrder {
	return &manufacturing.ProductionOrder{
		Base:           model.Base{ID: 1},
		OrganizationID: uint64Ptr(10),
		ItemID:         200,
		RecipeID:       uint64Ptr(1),
		QtyToProduce:   10,
		SrcLocationID:  uint64Ptr(30),
		DstLocationID:  uint64Ptr(40),
		State:          manufacturing.ProductionOrderStateDraft,
	}
}

func sampleConsumedMaterial() *manufacturing.ConsumedMaterial {
	return &manufacturing.ConsumedMaterial{
		Base:              model.Base{ID: 1},
		ProductionOrderID: 1,
		ItemID:            300,
		QtyPlanned:        20,
	}
}

func orderServiceFor(orders manufacturing.ProductionOrderDAOMock, components manufacturing.ConsumedMaterialDAOMock, recipes manufacturing.RecipeDAOMock, lines manufacturing.RecipeLineDAOMock, variants products.ItemVariantDAOMock, locations inventory.StockLocationDAOMock, reservations inventory.StockHoldDAOMock, quants inventory.StockBalanceDAOMock, sequences sequence.DAOMock) manufacturing.ProductionOrderService {
	return manufacturing.NewTestProductionOrderService(orders, components, recipes, lines, variants, locations, reservations, quants, sequences)
}

func orderServiceWith(orders manufacturing.ProductionOrderDAOMock, components manufacturing.ConsumedMaterialDAOMock) manufacturing.ProductionOrderService {
	return orderServiceFor(orders, components, defaultRecipes(), defaultLines(), defaultVariants(), defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock())
}

func defaultRecipes() manufacturing.RecipeDAOMock {
	return manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return &manufacturing.Recipe{Base: model.Base{ID: 1}, ItemID: 200, Qty: 1, Type: manufacturing.RecipeTypeManufacture}, nil
			},
		},
	}
}

func defaultLines() manufacturing.RecipeLineDAOMock {
	return manufacturing.RecipeLineDAOMock{
		ListByRecipeFunc: func(_ context.Context, _ uint64) ([]*manufacturing.RecipeLine, error) {
			return []*manufacturing.RecipeLine{{Base: model.Base{ID: 1}, RecipeID: 1, ComponentID: 300, Qty: 2}}, nil
		},
	}
}

func defaultVariants() products.ItemVariantDAOMock {
	return products.ItemVariantDAOMock{
		CRUDMock: dao.CRUDMock[products.ItemVariant]{
			FindFunc: func(_ context.Context, id uint64) (*products.ItemVariant, error) {
				return &products.ItemVariant{Base: model.Base{ID: id}, ItemID: 1}, nil
			},
		},
	}
}

func defaultLocations() inventory.StockLocationDAOMock {
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return referenceLocation(30, "internal", 10), nil
			},
		},
	}
}

func moDefaultService() manufacturing.ProductionOrderService {
	return orderServiceFor(
		manufacturing.ProductionOrderDAOMock{},
		manufacturing.ConsumedMaterialDAOMock{},
		defaultRecipes(),
		defaultLines(),
		defaultVariants(),
		defaultLocations(),
		inventory.StockHoldDAOMock{},
		inventory.StockBalanceDAOMock{},
		manufacturing.DefaultOrderSequenceMock(),
	)
}

func TestProductionOrderHandler_List_ReturnsOrders(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.ProductionOrder], error) {
				return &query.Page[manufacturing.ProductionOrder]{Items: []*manufacturing.ProductionOrder{sampleMO()}, Count: 1}, nil
			},
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionOrderHandler_List_RejectsInvalidQuery(t *testing.T) {
	app := orderTestApp(t, moDefaultService())

	resp, err := doRequest(app, http.MethodGet, "/production-orders/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{}
	app := fiber.New()
	h := NewProductionOrderHandler(orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/production-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestProductionOrderHandler_List_ReturnsServerError(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.ProductionOrder], error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Get_ReturnsOrder(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return sampleMO(), nil
			},
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Get_ReturnsNotFound(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Get_RejectsForeignOrganization(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(99), ItemID: 200, State: manufacturing.ProductionOrderStateDraft}, nil
			},
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Get_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, moDefaultService())

	resp, err := doRequest(app, http.MethodGet, "/production-orders/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Get_ReturnsServerError(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Create_CreatesOrder(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{},
		CreateWithComponentsFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder, components []*manufacturing.ConsumedMaterial) (*manufacturing.ProductionOrder, error) {
			productionOrder.ID = 1
			return productionOrder, nil
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ConsumedMaterial, error) {
			return []*manufacturing.ConsumedMaterial{sampleConsumedMaterial()}, nil
		},
	}
	svc := orderServiceFor(
		orders, components, defaultRecipes(), defaultLines(), defaultVariants(), defaultLocations(),
		inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	body := `{"item_id":200,"recipe_id":1,"qty_to_produce":10,"src_location_id":30,"dst_location_id":40}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Create_RejectsValidation(t *testing.T) {
	svc := moDefaultService()
	app := orderTestApp(t, svc)

	body := `{"item_id":200}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Create_MapsMissingRecipe(t *testing.T) {
	recipes := manufacturing.RecipeDAOMock{}
	svc := orderServiceFor(
		manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, recipes, defaultLines(),
		defaultVariants(), defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	body := `{"item_id":200,"recipe_id":1,"qty_to_produce":10,"src_location_id":30,"dst_location_id":40}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Create_ReturnsServerError(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{},
		CreateWithComponentsFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder, components []*manufacturing.ConsumedMaterial) (*manufacturing.ProductionOrder, error) {
			return nil, errors.New("db down")
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	body := `{"item_id":200,"recipe_id":1,"qty_to_produce":10,"src_location_id":30,"dst_location_id":40}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Create_MapsUnknownProduct(t *testing.T) {
	variants := products.ItemVariantDAOMock{}
	svc := orderServiceFor(
		manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(),
		variants, defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	body := `{"item_id":200,"recipe_id":1,"qty_to_produce":10,"src_location_id":30,"dst_location_id":40}`
	resp, err := doRequest(app, http.MethodPost, "/production-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Confirm_ConfirmsOrder(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return sampleMO(), nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Confirm_ReturnsNotFound(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return nil, nil
			},
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Confirm_MapsStateConflict(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), State: manufacturing.ProductionOrderStateConfirmed}, nil
			},
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Confirm_RejectsInvalidID(t *testing.T) {
	svc := moDefaultService()
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Confirm_ReturnsServerError(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return sampleMO(), nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/confirm", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Plan_PlansOrder(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), ItemID: 200, RecipeID: uint64Ptr(1), QtyToProduce: 10, SrcLocationID: uint64Ptr(30), DstLocationID: uint64Ptr(40), State: manufacturing.ProductionOrderStateConfirmed}, nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ConsumedMaterial, error) {
			return []*manufacturing.ConsumedMaterial{sampleConsumedMaterial()}, nil
		},
	}
	quants := inventory.StockBalanceDAOMock{
		FindByKeyFunc: func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), Quantity: 100, ReservedQty: 0}, nil
		},
	}
	svc := orderServiceFor(
		orders, components, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, quants, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/plan", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Plan_MapsReservationConflict(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), ItemID: 200, RecipeID: uint64Ptr(1), QtyToProduce: 10, SrcLocationID: uint64Ptr(30), DstLocationID: uint64Ptr(40), State: manufacturing.ProductionOrderStateConfirmed}, nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ConsumedMaterial, error) {
			return []*manufacturing.ConsumedMaterial{sampleConsumedMaterial()}, nil
		},
	}
	svc := orderServiceFor(
		orders, components, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/plan", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Plan_RejectsInvalidID(t *testing.T) {
	svc := moDefaultService()
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/plan", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Cancel_CancelsOrder(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return sampleMO(), nil
			},
			UpdateFunc: func(_ context.Context, productionOrder *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return productionOrder, nil
			},
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Cancel_MapsStateConflict(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return &manufacturing.ProductionOrder{Base: model.Base{ID: 1}, OrganizationID: uint64Ptr(10), State: manufacturing.ProductionOrderStateDone}, nil
			},
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Cancel_RejectsInvalidID(t *testing.T) {
	svc := moDefaultService()
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/abc/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_Cancel_ReturnsNotFound(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return nil, nil
			},
		},
	}
	svc := orderServiceFor(
		orders, manufacturing.ConsumedMaterialDAOMock{}, defaultRecipes(), defaultLines(), defaultVariants(),
		defaultLocations(), inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{}, manufacturing.DefaultOrderSequenceMock(),
	)
	app := orderTestApp(t, svc)

	resp, err := doRequest(app, http.MethodPost, "/production-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductionOrderHandler_ListComponents_ReturnsComponents(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return sampleMO(), nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ConsumedMaterial, error) {
			return []*manufacturing.ConsumedMaterial{sampleConsumedMaterial()}, nil
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, components))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1/components", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestProductionOrderHandler_ListComponents_ReturnsNotFound(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return nil, nil
			},
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, manufacturing.ConsumedMaterialDAOMock{}))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1/components", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestProductionOrderHandler_ListComponents_RejectsInvalidID(t *testing.T) {
	app := orderTestApp(t, moDefaultService())

	resp, err := doRequest(app, http.MethodGet, "/production-orders/abc/components", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestProductionOrderHandler_ListComponents_ReturnsServerError(t *testing.T) {
	orders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return sampleMO(), nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ConsumedMaterial, error) {
			return nil, errors.New("db down")
		},
	}
	app := orderTestApp(t, orderServiceWith(orders, components))

	resp, err := doRequest(app, http.MethodGet, "/production-orders/1/components", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
