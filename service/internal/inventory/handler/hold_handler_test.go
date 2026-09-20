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
)

func reservationTestApp(t *testing.T, withTenant bool, reservations inventory.StockHoldDAOMock, quants inventory.StockBalanceDAOMock) *fiber.App {
	t.Helper()
	svc := inventory.NewHoldService(reservations, quants)
	handler := NewHoldHandler(svc)
	return inventoryTestApp(t, withTenant, handler.Register)
}

func sampleStockHold() *inventory.StockHold {
	return &inventory.StockHold{
		Base:       model.Base{ID: 1},
		MovementID: inventoryUint64Ptr(3),
		BalanceID:  2,
		Qty:        5,
	}
}

func TestHoldHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockHold], error) {
			return &query.Page[inventory.StockHold]{Items: []*inventory.StockHold{sampleStockHold()}, Count: 1}, nil
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-holds/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-holds/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reservationTestApp(t, false, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-holds/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockHold], error) {
			return &query.Page[inventory.StockHold]{Items: []*inventory.StockHold{sampleStockHold()}, Count: 1}, nil
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-holds/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockHold], error) {
			return nil, errors.New("db down")
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-holds/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestHoldHandlerReserve(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{
				Base:           model.Base{ID: 1},
				OrganizationID: inventoryUint64Ptr(10),
				ItemID:         1,
				LocationID:     2,
				Quantity:       10,
			}, nil
		}
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, quants)

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"5"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reservationTestApp(t, false, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"5"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"abc"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("quant not found", func(t *testing.T) {
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"5"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("quant belongs to another organization", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{
				Base:           model.Base{ID: 1},
				OrganizationID: inventoryUint64Ptr(99),
				Quantity:       10,
			}, nil
		}
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, quants)

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"5"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("zero quantity", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{
				OrganizationID: inventoryUint64Ptr(10),
				Quantity:       10,
			}, nil
		}
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, quants)

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"0"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("reservation overflow", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return &inventory.StockBalance{
				OrganizationID: inventoryUint64Ptr(10),
				Quantity:       5,
			}, nil
		}
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, quants)

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"10"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		quants := inventory.StockBalanceDAOMock{}
		quants.FindByKeyFunc = func(_ context.Context, _, _ uint64, _ *uint64) (*inventory.StockBalance, error) {
			return nil, errors.New("db down")
		}
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, quants)

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/", `{"item_id":1,"location_id":2,"qty":"5"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestHoldHandlerReleaseByMovement(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/release-by-movement", `{"movement_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/release-by-movement", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reservationTestApp(t, false, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/release-by-movement", `{"movement_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("release error", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.ReleaseByMovementFunc = func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-holds/release-by-movement", `{"movement_id":3}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestHoldHandlerRelease(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockHold, error) {
			return sampleStockHold(), nil
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-holds/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockHold, error) {
			return nil, nil
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-holds/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := reservationTestApp(t, true, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-holds/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reservationTestApp(t, false, inventory.StockHoldDAOMock{}, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-holds/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockHold, error) {
			return nil, errors.New("db down")
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-holds/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("release error", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockHold, error) {
			return sampleStockHold(), nil
		}
		reservations.ReleaseFunc = func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-holds/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("reservation not found on release", func(t *testing.T) {
		reservations := inventory.StockHoldDAOMock{}
		reservations.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockHold, error) {
			return sampleStockHold(), nil
		}
		reservations.ReleaseFunc = func(_ context.Context, _ uint64) error {
			return inventory.ErrHoldNotFound
		}
		app := reservationTestApp(t, true, reservations, inventory.StockBalanceDAOMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-holds/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})
}

func TestWriteReservationError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "quant not found", err: inventory.ErrBalanceNotFound, status: http.StatusUnprocessableEntity},
		{name: "movement qty", err: inventory.ErrMovementQty, status: http.StatusUnprocessableEntity},
		{name: "reservation overflow", err: inventory.ErrHoldOverflow, status: http.StatusConflict},
		{name: "reservation not found", err: inventory.ErrHoldNotFound, status: http.StatusNotFound},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeReservationError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
