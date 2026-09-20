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

func transferTestApp(t *testing.T, withTenant bool, transfers inventory.WarehouseTransferDAOMock, movements inventory.StockMovementDAOMock, locations inventory.StockLocationDAOMock, warehouses inventory.WarehouseDAOMock, layers inventory.CostLayerDAOMock, quants inventory.StockBalanceDAOMock, resolver inventory.ItemResolverMock, poster inventory.PosterMock, tx inventory.TransactionerMock) *fiber.App {
	t.Helper()
	ledger := inventory.NewLedgerService(movements, quants, locations, inventory.TransactionerMock{})
	svc := inventory.NewTransferService(transfers, movements, locations, warehouses, layers, ledger, resolver, poster, tx)
	h := NewTransferHandler(svc)
	return inventoryTestApp(t, withTenant, h.Register)
}

func sampleWarehouseTransfer() *inventory.WarehouseTransfer {
	return &inventory.WarehouseTransfer{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		Name:           inventoryStringPtr("Transfer"),
		SrcWarehouseID: 1,
		DstWarehouseID: 2,
		State:          inventory.TransferStateDraft,
		OutShipmentID:  inventoryUint64Ptr(7),
		InShipmentID:   inventoryUint64Ptr(8),
	}
}

func sampleTransitLocation() *reference.StockLocation {
	return &reference.StockLocation{
		Base:           model.Base{ID: 90},
		OrganizationID: inventoryUint64Ptr(10),
		Name:           "Transit",
		Usage:          "transit",
	}
}

func sampleWarehouseForTransfer() *reference.Warehouse {
	return &reference.Warehouse{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		Name:           "Main",
	}
}

func sampleTransferMove() *inventory.StockMovement {
	return &inventory.StockMovement{
		Base:          model.Base{ID: 1},
		ItemID:        1,
		Qty:           5,
		SrcLocationID: 100,
		DstLocationID: 90,
		State:         inventory.MovementStateDraft,
	}
}

func TestTransferHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.WarehouseTransfer], error) {
			return &query.Page[inventory.WarehouseTransfer]{Items: []*inventory.WarehouseTransfer{sampleWarehouseTransfer()}, Count: 1}, nil
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.WarehouseTransfer], error) {
			return &query.Page[inventory.WarehouseTransfer]{Items: []*inventory.WarehouseTransfer{sampleWarehouseTransfer()}, Count: 1}, nil
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := transferTestApp(t, false, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.WarehouseTransfer], error) {
			return nil, errors.New("boom")
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTransferHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return sampleWarehouseTransfer(), nil
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return nil, nil
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return &inventory.WarehouseTransfer{Base: model.Base{ID: 1}, OrganizationID: inventoryUint64Ptr(99)}, nil
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return nil, errors.New("boom")
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/warehouse-transfers/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTransferHandlerCreate(t *testing.T) {
	body := `{"name":"Transfer","src_warehouse_id":1,"dst_warehouse_id":2,"lines":[{"item_id":1,"qty":5,"src_location_id":100,"dst_location_id":101}]}`

	t.Run("success", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{sampleTransitLocation()}, Count: 1}, nil
		}
		locations.FindFunc = func(_ context.Context, id uint64) (*reference.StockLocation, error) {
			return &reference.StockLocation{Base: model.Base{ID: id}, OrganizationID: inventoryUint64Ptr(10), Usage: "internal"}, nil
		}
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouseForTransfer(), nil
		}
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, locations, warehouses, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers", `{"name":"Transfer","src_warehouse_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("same warehouse", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		body := `{"name":"Transfer","src_warehouse_id":1,"dst_warehouse_id":1}`
		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers", body)
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
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, warehouses, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("transit not found", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[reference.StockLocation], error) {
			return &query.Page[reference.StockLocation]{Items: []*reference.StockLocation{}, Count: 0}, nil
		}
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return sampleWarehouseForTransfer(), nil
		}
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, locations, warehouses, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("service error", func(t *testing.T) {
		warehouses := inventory.WarehouseDAOMock{}
		warehouses.FindFunc = func(_ context.Context, _ uint64) (*reference.Warehouse, error) {
			return nil, errors.New("boom")
		}
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, warehouses, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestTransferHandlerSend(t *testing.T) {
	body := `{"journal_id":1,"transit_account_id":2}`

	sendMocks := func(state string, outShipmentID *uint64) (inventory.WarehouseTransferDAOMock, inventory.StockMovementDAOMock, inventory.StockLocationDAOMock, inventory.StockBalanceDAOMock, inventory.CostLayerDAOMock, inventory.ItemResolverMock) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return &inventory.WarehouseTransfer{
				Base:           model.Base{ID: 1},
				OrganizationID: inventoryUint64Ptr(10),
				State:          state,
				OutShipmentID:  outShipmentID,
			}, nil
		}
		movements := inventory.StockMovementDAOMock{}
		movements.ListByShipmentFunc = func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{sampleTransferMove()}, nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocationForCount(), nil
		}
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Quantity: 100}, nil
		}
		layers := inventory.CostLayerDAOMock{}
		layers.ListOpenByItemFunc = func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{RemainingQty: 1, RemainingValue: 100}}, nil
		}
		resolver := inventory.ItemResolverMock{}
		resolver.ResolveFunc = func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 500}}, nil
		}
		return transfers, movements, locations, quants, layers, resolver
	}

	t.Run("success", func(t *testing.T) {
		transfers, movements, locations, quants, layers, resolver := sendMocks(inventory.TransferStateDraft, inventoryUint64Ptr(7))
		app := transferTestApp(t, true, transfers, movements, locations, inventory.WarehouseDAOMock{}, layers, quants, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/send", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/abc/send", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/send", `{"journal_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return nil, nil
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/send", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("lookup error", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return nil, errors.New("boom")
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/send", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		transfers, movements, locations, quants, layers, resolver := sendMocks(inventory.TransferStateReceived, inventoryUint64Ptr(7))
		app := transferTestApp(t, true, transfers, movements, locations, inventory.WarehouseDAOMock{}, layers, quants, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/send", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("insufficient stock", func(t *testing.T) {
		transfers, movements, locations, _, layers, resolver := sendMocks(inventory.TransferStateDraft, inventoryUint64Ptr(7))
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{Quantity: 1}, nil
		}
		app := transferTestApp(t, true, transfers, movements, locations, inventory.WarehouseDAOMock{}, layers, quants, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/send", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("no valuation account", func(t *testing.T) {
		transfers, movements, locations, quants, layers, _ := sendMocks(inventory.TransferStateDraft, inventoryUint64Ptr(7))
		resolver := inventory.ItemResolverMock{}
		resolver.ResolveFunc = func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{Tracking: "none"}, nil
		}
		app := transferTestApp(t, true, transfers, movements, locations, inventory.WarehouseDAOMock{}, layers, quants, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/send", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestTransferHandlerReceive(t *testing.T) {
	body := `{"journal_id":1,"transit_account_id":2}`

	receiveMocks := func(state string, inShipmentID *uint64) (inventory.WarehouseTransferDAOMock, inventory.StockMovementDAOMock, inventory.CostLayerDAOMock, inventory.ItemResolverMock) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return &inventory.WarehouseTransfer{
				Base:           model.Base{ID: 1},
				OrganizationID: inventoryUint64Ptr(10),
				State:          state,
				InShipmentID:   inShipmentID,
			}, nil
		}
		movements := inventory.StockMovementDAOMock{}
		movements.ListByShipmentFunc = func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{sampleTransferMove()}, nil
		}
		layers := inventory.CostLayerDAOMock{}
		layers.ListOpenByItemFunc = func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{RemainingQty: 1, RemainingValue: 100}}, nil
		}
		resolver := inventory.ItemResolverMock{}
		resolver.ResolveFunc = func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 500}}, nil
		}
		return transfers, movements, layers, resolver
	}

	t.Run("success", func(t *testing.T) {
		transfers, movements, layers, resolver := receiveMocks(inventory.TransferStateInTransit, inventoryUint64Ptr(8))
		app := transferTestApp(t, true, transfers, movements, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, layers, inventory.StockBalanceDAOMock{}, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/receive", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/abc/receive", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := transferTestApp(t, true, inventory.WarehouseTransferDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/receive", `{"journal_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return nil, nil
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/receive", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		transfers, movements, layers, resolver := receiveMocks(inventory.TransferStateDraft, inventoryUint64Ptr(8))
		app := transferTestApp(t, true, transfers, movements, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, layers, inventory.StockBalanceDAOMock{}, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/receive", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("lookup error", func(t *testing.T) {
		transfers := inventory.WarehouseTransferDAOMock{}
		transfers.FindFunc = func(_ context.Context, _ uint64) (*inventory.WarehouseTransfer, error) {
			return nil, errors.New("boom")
		}
		app := transferTestApp(t, true, transfers, inventory.StockMovementDAOMock{}, inventory.StockLocationDAOMock{}, inventory.WarehouseDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/warehouse-transfers/1/receive", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWriteTransferError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: inventory.ErrWarehouseTransferNotFound, status: http.StatusNotFound},
		{name: "state", err: inventory.ErrWarehouseTransferState, status: http.StatusConflict},
		{name: "same warehouse", err: inventory.ErrSameWarehouse, status: http.StatusUnprocessableEntity},
		{name: "warehouse not found", err: inventory.ErrWarehouseNotFound, status: http.StatusUnprocessableEntity},
		{name: "transit not found", err: inventory.ErrTransitNotFound, status: http.StatusUnprocessableEntity},
		{name: "transit account", err: inventory.ErrTransitAccount, status: http.StatusUnprocessableEntity},
		{name: "organization missing", err: inventory.ErrOrganizationMissing, status: http.StatusUnprocessableEntity},
		{name: "movement qty", err: inventory.ErrMovementQty, status: http.StatusUnprocessableEntity},
		{name: "insufficient stock", err: inventory.ErrInsufficientStock, status: http.StatusConflict},
		{name: "location required", err: inventory.ErrLocationRequired, status: http.StatusUnprocessableEntity},
		{name: "valuation account", err: inventory.ErrValuationAccount, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeTransferError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
