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
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

func outsideProcessingTestApp(t *testing.T, orders manufacturing.OutsideProcessingOrderDAO, svc manufacturing.OutsideProcessingService) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(10))
		return c.Next()
	})
	h := NewOutsideProcessingHandler(svc)
	h.Register(app, passthroughGuards())
	return app
}

func outsideProcessingServiceFor(
	outsideOrders manufacturing.OutsideProcessingOrderDAOMock,
	productionOrders manufacturing.ProductionOrderDAOMock,
	components manufacturing.ConsumedMaterialDAOMock,
	recipes manufacturing.RecipeDAOMock,
	purchases manufacturing.PurchaseOrderCreatorMock,
	poDAO procurement.PurchaseOrderDAOMock,
	poLines procurement.PurchaseOrderLineDAOMock,
	locations inventory.StockLocationDAOMock,
	movements inventory.StockMovementDAOMock,
	layers inventory.CostLayerDAOMock,
	resolver inventory.ItemResolverMock,
	poster inventory.PosterMock,
) manufacturing.OutsideProcessingService {
	return manufacturing.NewTestOutsideProcessingService(outsideOrders, productionOrders, components, recipes, purchases, poDAO, poLines, locations, movements, layers, resolver, poster)
}

func subcontractSupplierLocations() inventory.StockLocationDAOMock {
	return inventory.StockLocationDAOMock{
		CRUDMock: dao.CRUDMock[reference.StockLocation]{
			FindFunc: func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
				return &reference.StockLocation{Base: model.Base{ID: 30}, OrganizationID: uint64Ptr(10), Usage: "supplier"}, nil
			},
			ListFunc: func(_ context.Context, q *query.Query) (*query.Page[reference.StockLocation], error) {
				if q.Filters[0].Field != "usage" {
					return &query.Page[reference.StockLocation]{}, nil
				}
				return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{{Base: model.Base{ID: 30}, OrganizationID: uint64Ptr(10), Usage: "supplier"}}}, nil
			},
		},
	}
}

func outsideProcessingRecipes() manufacturing.RecipeDAOMock {
	return manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return &manufacturing.Recipe{Base: model.Base{ID: 1}, Type: manufacturing.RecipeTypeSubcontract}, nil
			},
		},
	}
}

func confirmedProductionOrder() *manufacturing.ProductionOrder {
	return &manufacturing.ProductionOrder{
		Base:           model.Base{ID: 1},
		OrganizationID: uint64Ptr(10),
		ItemID:         200,
		RecipeID:       uint64Ptr(1),
		QtyToProduce:   10,
		SrcLocationID:  uint64Ptr(20),
		DstLocationID:  uint64Ptr(21),
		State:          manufacturing.ProductionOrderStateConfirmed,
	}
}

func subcontractMoveMocks() (inventory.StockMovementDAOMock, inventory.CostLayerDAOMock, inventory.ItemResolverMock, inventory.PosterMock) {
	movements := inventory.StockMovementDAOMock{
		CRUDMock: dao.CRUDMock[inventory.StockMovement]{
			CreateFunc: func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
				movement.ID = 100
				movement.State = inventory.MovementStateDraft
				return movement, nil
			},
			FindFunc: func(_ context.Context, moveID uint64) (*inventory.StockMovement, error) {
				return &inventory.StockMovement{Base: model.Base{ID: moveID}, ItemID: 300, Qty: 10, SrcLocationID: 20, DstLocationID: 30, State: inventory.MovementStateDraft, OrganizationID: uint64Ptr(10)}, nil
			},
		},
		ApplyTxFunc: func(_ context.Context, _ *gorm.DB, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement.State = inventory.MovementStateDone
			return movement, nil
		},
	}
	layers := inventory.CostLayerDAOMock{
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
	resolver := inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 400}, Tracking: "none", CostMethod: "standard", StandardCost: 5}, nil
		},
	}
	poster := inventory.PosterMock{
		PostFunc: func(_ context.Context, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
		PostTxFunc: func(_ context.Context, _ *gorm.DB, request accounting.PostRequest) (*accounting.JournalEntry, error) {
			return &accounting.JournalEntry{Base: model.Base{ID: 1}}, nil
		},
	}
	return movements, layers, resolver, poster
}

func sampleOutsideProcessingOrder() *manufacturing.OutsideProcessingOrder {
	return &manufacturing.OutsideProcessingOrder{
		Base:              model.Base{ID: 1},
		ProductionOrderID: 1,
		SupplierID:        7,
		State:             manufacturing.OutsideProcessingStateDraft,
	}
}

func TestOutsideProcessingHandler_Create_CreatesOrder(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			CreateFunc: func(_ context.Context, sub *manufacturing.OutsideProcessingOrder) (*manufacturing.OutsideProcessingOrder, error) {
				sub.ID = 1
				return sub, nil
			},
		},
	}
	productionOrders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return confirmedProductionOrder(), nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, manufacturing.ConsumedMaterialDAOMock{}, outsideProcessingRecipes(), manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"production_order_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Create_RejectsValidation(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"production_order_id":0,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Create_MapsNotFound(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	productionOrders := manufacturing.ProductionOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"production_order_id":99,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Create_MapsDuplicate(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[manufacturing.OutsideProcessingOrder], error) {
				return &query.Page[manufacturing.OutsideProcessingOrder]{Items: []*manufacturing.OutsideProcessingOrder{sampleOutsideProcessingOrder()}}, nil
			},
		},
	}
	productionOrders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return confirmedProductionOrder(), nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, manufacturing.ConsumedMaterialDAOMock{}, outsideProcessingRecipes(), manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"production_order_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Create_MapsNotSubcontracted(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	productionOrders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return confirmedProductionOrder(), nil
			},
		},
	}
	recipes := manufacturing.RecipeDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.Recipe]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.Recipe, error) {
				return &manufacturing.Recipe{Base: model.Base{ID: 1}, Type: manufacturing.RecipeTypeManufacture}, nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, manufacturing.ConsumedMaterialDAOMock{}, recipes, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"production_order_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Create_ReturnsServerError(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	productionOrders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"production_order_id":1,"supplier_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_SendsOrder(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return sampleOutsideProcessingOrder(), nil
			},
			UpdateFunc: func(_ context.Context, sub *manufacturing.OutsideProcessingOrder) (*manufacturing.OutsideProcessingOrder, error) {
				return sub, nil
			},
		},
	}
	productionOrders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return confirmedProductionOrder(), nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ConsumedMaterial, error) {
			return []*manufacturing.ConsumedMaterial{{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 10}}, nil
		},
	}
	purchases := manufacturing.PurchaseOrderCreatorMock{
		CreateFunc: func(_ context.Context, po *procurement.PurchaseOrder, _ []*procurement.PurchaseOrderLine) (*procurement.PurchaseOrder, error) {
			po.ID = 55
			po.AmountUntaxed = 120
			return po, nil
		},
	}
	movements, layers, resolver, poster := subcontractMoveMocks()
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, components, outsideProcessingRecipes(), purchases, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, subcontractSupplierLocations(), movements, layers, resolver, poster)
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_RejectsValidation(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":0,"wip_account_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_RejectsInvalidID(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/abc/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_MapsNotFound(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/99/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_MapsStateConflict(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return &manufacturing.OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: manufacturing.OutsideProcessingStateSent}, nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_MapsNoComponents(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return sampleOutsideProcessingOrder(), nil
			},
		},
	}
	productionOrders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return confirmedProductionOrder(), nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, manufacturing.ConsumedMaterialDAOMock{}, outsideProcessingRecipes(), manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_MapsWIPAccount(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return sampleOutsideProcessingOrder(), nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Send_ReturnsServerError(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/send", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Receive_ReceivesOrder(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return &manufacturing.OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, SupplierID: 7, PurchaseOrderID: uint64Ptr(55), State: manufacturing.OutsideProcessingStateSent}, nil
			},
			UpdateFunc: func(_ context.Context, sub *manufacturing.OutsideProcessingOrder) (*manufacturing.OutsideProcessingOrder, error) {
				return sub, nil
			},
		},
	}
	productionOrder := confirmedProductionOrder()
	productionOrders := manufacturing.ProductionOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.ProductionOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.ProductionOrder, error) {
				return productionOrder, nil
			},
			UpdateFunc: func(_ context.Context, updated *manufacturing.ProductionOrder) (*manufacturing.ProductionOrder, error) {
				return updated, nil
			},
		},
	}
	components := manufacturing.ConsumedMaterialDAOMock{
		ListByProductionOrderFunc: func(_ context.Context, _ uint64) ([]*manufacturing.ConsumedMaterial, error) {
			return []*manufacturing.ConsumedMaterial{{Base: model.Base{ID: 5}, ProductionOrderID: 1, ItemID: 300, QtyPlanned: 10}}, nil
		},
	}
	poDAO := procurement.PurchaseOrderDAOMock{
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*procurement.PurchaseOrder, error) {
				return &procurement.PurchaseOrder{Base: model.Base{ID: 55}, AmountUntaxed: 120}, nil
			},
		},
	}
	poLines := procurement.PurchaseOrderLineDAOMock{
		ListByOrderFunc: func(_ context.Context, _ uint64) ([]*procurement.PurchaseOrderLine, error) {
			return []*procurement.PurchaseOrderLine{{Base: model.Base{ID: 1}, QtyOrdered: 10}}, nil
		},
		CRUDMock: dao.CRUDMock[procurement.PurchaseOrderLine]{
			UpdateFunc: func(_ context.Context, line *procurement.PurchaseOrderLine) (*procurement.PurchaseOrderLine, error) {
				return line, nil
			},
		},
	}
	movements, layers, resolver, poster := subcontractMoveMocks()
	svc := outsideProcessingServiceFor(outsideOrders, productionOrders, components, outsideProcessingRecipes(), manufacturing.PurchaseOrderCreatorMock{}, poDAO, poLines, subcontractSupplierLocations(), movements, layers, resolver, poster)
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600,"ap_payable_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Receive_RejectsValidation(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":0,"wip_account_id":0,"ap_payable_account_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Receive_RejectsInvalidID(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600,"ap_payable_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/abc/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Receive_MapsNotFound(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600,"ap_payable_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/99/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Receive_MapsOperationError(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return sampleOutsideProcessingOrder(), nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600,"ap_payable_account_id":0}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Receive_MapsMissingPO(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return &manufacturing.OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: manufacturing.OutsideProcessingStateSent}, nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600,"ap_payable_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Receive_ReturnsServerError(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	body := `{"journal_id":500,"wip_account_id":600,"ap_payable_account_id":700}`
	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/receive", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Done_CompletesOrder(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return &manufacturing.OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: manufacturing.OutsideProcessingStateReceived}, nil
			},
			UpdateFunc: func(_ context.Context, sub *manufacturing.OutsideProcessingOrder) (*manufacturing.OutsideProcessingOrder, error) {
				return sub, nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Done_RejectsInvalidID(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/abc/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Done_MapsNotFound(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/99/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Done_MapsStateConflict(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return sampleOutsideProcessingOrder(), nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Done_ReturnsServerError(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/done", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Cancel_CancelsOrder(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return sampleOutsideProcessingOrder(), nil
			},
			UpdateFunc: func(_ context.Context, sub *manufacturing.OutsideProcessingOrder) (*manufacturing.OutsideProcessingOrder, error) {
				return sub, nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Cancel_RejectsInvalidID(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/abc/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Cancel_MapsNotFound(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/99/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Cancel_MapsStateConflict(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return &manufacturing.OutsideProcessingOrder{Base: model.Base{ID: 1}, ProductionOrderID: 1, State: manufacturing.OutsideProcessingStateDone}, nil
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestOutsideProcessingHandler_Cancel_ReturnsServerError(t *testing.T) {
	outsideOrders := manufacturing.OutsideProcessingOrderDAOMock{
		CRUDMock: dao.CRUDMock[manufacturing.OutsideProcessingOrder]{
			FindFunc: func(_ context.Context, _ uint64) (*manufacturing.OutsideProcessingOrder, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := outsideProcessingServiceFor(outsideOrders, manufacturing.ProductionOrderDAOMock{}, manufacturing.ConsumedMaterialDAOMock{}, manufacturing.RecipeDAOMock{}, manufacturing.PurchaseOrderCreatorMock{}, procurement.PurchaseOrderDAOMock{}, procurement.PurchaseOrderLineDAOMock{}, inventory.StockLocationDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{})
	app := outsideProcessingTestApp(t, outsideOrders, svc)

	resp, err := doRequest(app, http.MethodPost, "/outside-processing-orders/1/cancel", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
