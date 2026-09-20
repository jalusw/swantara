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

func warehouseTestApp(t *testing.T, withTenant bool, warehouses inventory.WarehouseDAOMock, locations inventory.StockLocationDAOMock) *fiber.App {
	t.Helper()
	svc := inventory.NewWarehouseService(warehouses, locations)
	handler := NewWarehouseHandler(svc)
	return inventoryTestApp(t, withTenant, handler.Register)
}

func sampleWarehouse() *reference.Warehouse {
	return &reference.Warehouse{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		Name:           "Main",
		Code:           inventoryStringPtr("MAIN"),
	}
}

func sampleStockLocation() *reference.StockLocation {
	return &reference.StockLocation{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		WarehouseID:    inventoryUint64Ptr(1),
		Name:           "Storage",
		Code:           inventoryStringPtr("STORAGE"),
		Usage:          "internal",
	}
}

func TestWarehouseHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Warehouse], error) {
			return &query.Page[reference.Warehouse]{Items: []*reference.Warehouse{sampleWarehouse()}, Count: 1}, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := warehouseTestApp(t, false, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Warehouse], error) {
			return &query.Page[reference.Warehouse]{Items: []*reference.Warehouse{sampleWarehouse()}, Count: 1}, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.Warehouse], error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouse(), nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("foreign organization", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			warehouse := sampleWarehouse()
			warehouse.OrganizationID = inventoryUint64Ptr(99)
			return warehouse, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.CreateFunc = func(_ context.Context, warehouse *reference.Warehouse) (*reference.Warehouse, error) {
			warehouse.ID = 1
			return warehouse, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouses/", `{"name":"Main","code":"MAIN"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouses/", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("code taken", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.SearchFunc = func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
			existing := sampleWarehouse()
			existing.ID = 7
			return existing, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouses/", `{"name":"Main","code":"MAIN"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("search error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.SearchFunc = func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouses/", `{"name":"Main","code":"MAIN"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("create error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.CreateFunc = func(_ context.Context, _ *reference.Warehouse) (*reference.Warehouse, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouses/", `{"name":"Main"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouse(), nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPut, "/warehouses/1", `{"name":"Main","code":"MAIN"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPut, "/warehouses/1", `{"name":"Main"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPut, "/warehouses/abc", `{"name":"Main"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouse(), nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPut, "/warehouses/1", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("code taken", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouse(), nil
		}
		warehouses.SearchFunc = func(_ context.Context, _ string, _ any) (*reference.Warehouse, error) {
			existing := sampleWarehouse()
			existing.ID = 7
			return existing, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPut, "/warehouses/1", `{"name":"Main","code":"MAIN"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPut, "/warehouses/1", `{"name":"Main"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouse(), nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/warehouses/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouse(), nil
		}
		warehouses.DeleteFunc = func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/warehouses/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerListLocations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{sampleStockLocation()}, Count: 1}, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := warehouseTestApp(t, false, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{sampleStockLocation()}, Count: 1}, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerGetLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("foreign organization", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			location := sampleStockLocation()
			location.OrganizationID = inventoryUint64Ptr(99)
			return location, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodGet, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerCreateLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.CreateFunc = func(_ context.Context, location *reference.StockLocation) (*reference.StockLocation, error) {
			location.ID = 1
			return location, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPost, "/stock-locations/", `{"name":"Storage","usage":"internal"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-locations/", `{"name":"Storage"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid usage", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-locations/", `{"name":"Storage","usage":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("warehouse not found", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-locations/", `{"name":"Storage","usage":"internal","warehouse_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("parent not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPost, "/stock-locations/", `{"name":"Storage","usage":"internal","parent_id":9}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("warehouse mismatch", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			warehouse := sampleWarehouse()
			warehouse.OrganizationID = inventoryUint64Ptr(99)
			return warehouse, nil
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-locations/", `{"name":"Storage","usage":"internal","warehouse_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("warehouse find error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, warehouses, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-locations/", `{"name":"Storage","usage":"internal","warehouse_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerUpdateLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPut, "/stock-locations/1", `{"name":"Storage","usage":"internal"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPut, "/stock-locations/1", `{"name":"Storage","usage":"internal"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodPut, "/stock-locations/abc", `{"name":"Storage","usage":"internal"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid usage", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPut, "/stock-locations/1", `{"name":"Storage","usage":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPut, "/stock-locations/1", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPut, "/stock-locations/1", `{"name":"Storage","usage":"internal"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("update error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		locations.UpdateFunc = func(_ context.Context, _ *reference.StockLocation) (*reference.StockLocation, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodPut, "/stock-locations/1", `{"name":"Storage","usage":"internal"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWarehouseHandlerDeleteLocation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodDelete, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, nil
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodDelete, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, inventory.StockLocationDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-locations/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, errors.New("db down")
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodDelete, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		locations.DeleteFunc = func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		}
		app := warehouseTestApp(t, true, inventory.WarehouseDAOMock{}, locations)

		resp, err := doRequest(app, http.MethodDelete, "/stock-locations/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWriteWarehouseError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "name required", err: inventory.ErrWarehouseNameRequired, status: http.StatusUnprocessableEntity},
		{name: "code taken", err: inventory.ErrWarehouseCodeTaken, status: http.StatusConflict},
		{name: "warehouse not found", err: inventory.ErrWarehouseNotFound, status: http.StatusUnprocessableEntity},
		{name: "location name required", err: inventory.ErrLocationNameRequired, status: http.StatusUnprocessableEntity},
		{name: "invalid location type", err: inventory.ErrInvalidLocationType, status: http.StatusUnprocessableEntity},
		{name: "location parent", err: inventory.ErrLocationParent, status: http.StatusUnprocessableEntity},
		{name: "warehouse mismatch", err: inventory.ErrWarehouseMismatch, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeWarehouseError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
