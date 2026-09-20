package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func stockTestApp(t *testing.T, withTenant bool, movements inventory.StockMovementDAOMock, shipments inventory.ShipmentDAOMock, quants inventory.StockBalanceDAOMock, locations inventory.StockLocationDAOMock, layers inventory.CostLayerDAOMock, resolver inventory.ItemResolverMock) *fiber.App {
	t.Helper()
	ledger := inventory.NewLedgerService(movements, quants, locations, inventory.TransactionerMock{})
	valuation := inventory.NewValuationService(movements, layers, locations, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})
	handler := NewStockHandler(inventory.NewStockService(movements, shipments, quants), ledger, valuation)
	return inventoryTestApp(t, withTenant, handler.Register)
}

func sampleStockMovement() *inventory.StockMovement {
	return &inventory.StockMovement{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		ItemID:         5,
		Qty:            2,
		SrcLocationID:  2,
		DstLocationID:  3,
		State:          inventory.MovementStateDraft,
	}
}

func sampleShipment() *inventory.Shipment {
	return &inventory.Shipment{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		Name:           inventoryStringPtr("PO-001"),
		Type:           inventory.ShipmentTypeIncoming,
		State:          inventory.ShipmentStateDraft,
	}
}

func sampleStockBalance() *inventory.StockBalance {
	return &inventory.StockBalance{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		ItemID:         5,
		LocationID:     2,
		Quantity:       10,
	}
}

func sampleCostLayer() *inventory.CostLayer {
	return &inventory.CostLayer{
		Base:           model.Base{ID: 9},
		ItemID:         5,
		Quantity:       10,
		UnitCost:       inventoryFloat64Ptr(5),
		Value:          50,
		RemainingQty:   10,
		RemainingValue: 50,
	}
}

func valuationResolver(accounts inventory.StockAccounts, tracking string) inventory.ItemResolverMock {
	return inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: accounts, Tracking: tracking, CostMethod: "fifo", OrganizationID: inventoryUint64Ptr(10)}, nil
		},
	}
}

func supplierLocation() *reference.StockLocation {
	location := sampleStockLocation()
	location.Usage = "supplier"
	return location
}

func customerLocation() *reference.StockLocation {
	location := sampleStockLocation()
	location.Usage = "customer"
	return location
}

func TestStockHandlerListMoves(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockMovement], error) {
			return &query.Page[inventory.StockMovement]{Items: []*inventory.StockMovement{sampleStockMovement()}, Count: 1}, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := stockTestApp(t, false, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockMovement], error) {
			return &query.Page[inventory.StockMovement]{Items: []*inventory.StockMovement{sampleStockMovement()}, Count: 1}, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockMovement], error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerGetMovement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("foreign organization", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			movement := sampleStockMovement()
			movement.OrganizationID = inventoryUint64Ptr(99)
			return movement, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerCreateMovement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		movements := inventory.StockMovementDAOMock{}
		movements.CreateFunc = func(_ context.Context, movement *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement.ID = 1
			return movement, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/", `{"item_id":5,"qty":"2","src_location_id":2,"dst_location_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("with scheduled date", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.CreateFunc = func(_ context.Context, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
			movement := sampleStockMovement()
			movement.ID = 1
			return movement, nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/", `{"item_id":5,"qty":"2","src_location_id":2,"dst_location_id":3,"scheduled_date":"2026-01-15"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/", `{"item_id":5,"qty":"abc","src_location_id":2,"dst_location_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("location not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/", `{"item_id":5,"qty":"2","src_location_id":2,"dst_location_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("location find error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/", `{"item_id":5,"qty":"2","src_location_id":2,"dst_location_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("create error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		movements := inventory.StockMovementDAOMock{}
		movements.CreateFunc = func(_ context.Context, _ *inventory.StockMovement) (*inventory.StockMovement, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/", `{"item_id":5,"qty":"2","src_location_id":2,"dst_location_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerDeleteMovement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-movements/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		movements.DeleteFunc = func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-movements/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerListShipments(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		shipments := inventory.ShipmentDAOMock{}
		shipments.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
			return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{sampleShipment()}, Count: 1}, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, shipments, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := stockTestApp(t, false, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		shipments := inventory.ShipmentDAOMock{}
		shipments.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, shipments, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		shipments := inventory.ShipmentDAOMock{}
		shipments.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.Shipment], error) {
			return &query.Page[inventory.Shipment]{Items: []*inventory.Shipment{sampleShipment()}, Count: 1}, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, shipments, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerGetShipment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		shipments := inventory.ShipmentDAOMock{}
		shipments.FindFunc = func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
			return sampleShipment(), nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, shipments, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		shipments := inventory.ShipmentDAOMock{}
		shipments.FindFunc = func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, shipments, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("foreign organization", func(t *testing.T) {
		shipments := inventory.ShipmentDAOMock{}
		shipments.FindFunc = func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
			shipment := sampleShipment()
			shipment.OrganizationID = inventoryUint64Ptr(99)
			return shipment, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, shipments, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		shipments := inventory.ShipmentDAOMock{}
		shipments.FindFunc = func(_ context.Context, _ uint64) (*inventory.Shipment, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, shipments, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/shipments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerListBalances(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockBalance], error) {
			return &query.Page[inventory.StockBalance]{Items: []*inventory.StockBalance{sampleStockBalance()}, Count: 1}, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, quants, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/balances", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/balances?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := stockTestApp(t, false, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/balances", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockBalance], error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, quants, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/balances", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockBalance], error) {
			return &query.Page[inventory.StockBalance]{Items: []*inventory.StockBalance{sampleStockBalance()}, Count: 1}, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, quants, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/balances?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerOnHand(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return sampleStockBalance(), nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, quants, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/on-hand?item_id=5&location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("missing item id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/on-hand?location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("missing location id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/on-hand?item_id=5", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("location not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/on-hand?item_id=5&location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("location find error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/on-hand?item_id=5&location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerAvailableToPromise(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			quant := sampleStockBalance()
			quant.ReservedQty = 2
			return quant, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, quants, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/available-to-promise?item_id=5&location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("missing item id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/available-to-promise?location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("missing location id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/available-to-promise?item_id=5", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("location not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/available-to-promise?item_id=5&location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock/available-to-promise?item_id=5&location_id=2", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerRebuildBalances(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock/rebuild", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("ledger error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.LedgerTotalsFunc = func(_ context.Context, _ *uint64) ([]inventory.LedgerTotal, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock/rebuild", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerReceive(t *testing.T) {
	accounts := inventory.StockAccounts{StockValuationAccountID: 100, StockInputAccountID: 200}

	t.Run("success", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("with valid date", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5,"date":"2026-01-15"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("with invalid date", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5,"date":"not-a-date"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/abc/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("movement already done", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			movement := sampleStockMovement()
			movement.State = inventory.MovementStateDone
			return movement, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("negative cost", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"-5","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not a receipt", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("batch required", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "batch"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("organization missing", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			movement := sampleStockMovement()
			movement.OrganizationID = nil
			return movement, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("no valuation account", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(inventory.StockAccounts{}, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no input account", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(inventory.StockAccounts{StockValuationAccountID: 100}, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"10","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("posting error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return supplierLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/receive", `{"unit_cost":"abc","journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestStockHandlerShip(t *testing.T) {
	accounts := inventory.StockAccounts{StockValuationAccountID: 100, CogsAccountID: 300}

	t.Run("success", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return customerLocation(), nil
		}
		layers := inventory.CostLayerDAOMock{}
		layers.ListOpenByItemFunc = func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{sampleCostLayer()}, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, layers, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/abc/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := stockTestApp(t, true, inventory.StockMovementDAOMock{}, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return nil, errors.New("db down")
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("movement already done", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			movement := sampleStockMovement()
			movement.State = inventory.MovementStateDone
			return movement, nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("not a shipment", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no cogs account", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return customerLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(inventory.StockAccounts{StockValuationAccountID: 100}, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return customerLocation(), nil
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, valuationResolver(accounts, "none"))

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("resolver error", func(t *testing.T) {
		movements := inventory.StockMovementDAOMock{}
		movements.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockMovement, error) {
			return sampleStockMovement(), nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return customerLocation(), nil
		}
		resolver := inventory.ItemResolverMock{
			ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
				return inventory.ResolvedItem{}, errors.New("db down")
			},
		}
		app := stockTestApp(t, true, movements, inventory.ShipmentDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.CostLayerDAOMock{}, resolver)

		resp, err := doRequest(app, http.MethodPost, "/stock-movements/1/ship", `{"journal_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWriteStockError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "movement not found", err: inventory.ErrMovementNotFound, status: http.StatusNotFound},
		{name: "movement item", err: inventory.ErrMovementItem, status: http.StatusUnprocessableEntity},
		{name: "movement qty", err: inventory.ErrMovementQty, status: http.StatusUnprocessableEntity},
		{name: "movement state", err: inventory.ErrMovementState, status: http.StatusConflict},
		{name: "location not found", err: inventory.ErrLocationNotFound, status: http.StatusUnprocessableEntity},
		{name: "insufficient stock", err: inventory.ErrInsufficientStock, status: http.StatusConflict},
		{name: "not receipt", err: inventory.ErrNotReceipt, status: http.StatusUnprocessableEntity},
		{name: "not shipment", err: inventory.ErrNotShipment, status: http.StatusUnprocessableEntity},
		{name: "batch required", err: inventory.ErrBatchRequired, status: http.StatusUnprocessableEntity},
		{name: "negative cost", err: inventory.ErrNegativeCost, status: http.StatusUnprocessableEntity},
		{name: "organization missing", err: inventory.ErrOrganizationMissing, status: http.StatusUnprocessableEntity},
		{name: "valuation account", err: inventory.ErrValuationAccount, status: http.StatusUnprocessableEntity},
		{name: "input account", err: inventory.ErrInputAccount, status: http.StatusUnprocessableEntity},
		{name: "cogs account", err: inventory.ErrCogsAccount, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeStockError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
