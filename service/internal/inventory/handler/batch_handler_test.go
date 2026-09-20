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

func lotTestApp(t *testing.T, withTenant bool, lots inventory.BatchDAOMock, resolver inventory.ItemResolverMock) *fiber.App {
	t.Helper()
	svc := inventory.NewBatchService(lots, resolver)
	handler := NewBatchHandler(svc)
	return inventoryTestApp(t, withTenant, handler.Register)
}

func sampleBatch() *inventory.Batch {
	return &inventory.Batch{
		Base:   model.Base{ID: 1},
		ItemID: 5,
		Name:   "LOT-001",
	}
}

func trackingResolver(tracking string) inventory.ItemResolverMock {
	return inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{Tracking: tracking, OrganizationID: inventoryUint64Ptr(10)}, nil
		},
	}
}

func TestBatchHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.Batch], error) {
			return &query.Page[inventory.Batch]{Items: []*inventory.Batch{sampleBatch()}, Count: 1}, nil
		}
		app := lotTestApp(t, true, lots, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := lotTestApp(t, false, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.Batch], error) {
			return &query.Page[inventory.Batch]{Items: []*inventory.Batch{sampleBatch()}, Count: 1}, nil
		}
		app := lotTestApp(t, true, lots, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.Batch], error) {
			return nil, errors.New("db down")
		}
		app := lotTestApp(t, true, lots, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestBatchHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return sampleBatch(), nil
		}
		app := lotTestApp(t, true, lots, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return nil, nil
		}
		app := lotTestApp(t, true, lots, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := lotTestApp(t, false, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return nil, errors.New("db down")
		}
		app := lotTestApp(t, true, lots, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/batches/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestBatchHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.CreateFunc = func(_ context.Context, batch *inventory.Batch) (*inventory.Batch, error) {
			batch.ID = 1
			return batch, nil
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := lotTestApp(t, false, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("variant not found", func(t *testing.T) {
		resolver := inventory.ItemResolverMock{
			ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
				return inventory.ResolvedItem{}, inventory.ErrVariantNotFound
			},
		}
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, resolver)

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("tracking disabled", func(t *testing.T) {
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, trackingResolver("none"))

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("template not found", func(t *testing.T) {
		resolver := inventory.ItemResolverMock{
			ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
				return inventory.ResolvedItem{Tracking: "batch"}, nil
			},
		}
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, resolver)

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("batch name taken", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.ListByItemFunc = func(_ context.Context, _ uint64) ([]*inventory.Batch, error) {
			return []*inventory.Batch{sampleBatch()}, nil
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("create error", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.CreateFunc = func(_ context.Context, _ *inventory.Batch) (*inventory.Batch, error) {
			return nil, errors.New("db down")
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("resolver error", func(t *testing.T) {
		resolver := inventory.ItemResolverMock{
			ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
				return inventory.ResolvedItem{}, errors.New("db down")
			},
		}
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, resolver)

		resp, err := doRequest(app, http.MethodPost, "/batches/", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestBatchHandlerUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return sampleBatch(), nil
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPut, "/batches/1", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return nil, nil
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPut, "/batches/1", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := lotTestApp(t, true, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPut, "/batches/abc", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return sampleBatch(), nil
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPut, "/batches/1", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := lotTestApp(t, false, inventory.BatchDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPut, "/batches/1", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("tracking disabled", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return sampleBatch(), nil
		}
		app := lotTestApp(t, true, lots, trackingResolver("none"))

		resp, err := doRequest(app, http.MethodPut, "/batches/1", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return nil, errors.New("db down")
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPut, "/batches/1", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("update error", func(t *testing.T) {
		lots := inventory.BatchDAOMock{}
		lots.FindFunc = func(_ context.Context, _ uint64) (*inventory.Batch, error) {
			return sampleBatch(), nil
		}
		lots.UpdateFunc = func(_ context.Context, _ *inventory.Batch) (*inventory.Batch, error) {
			return nil, errors.New("db down")
		}
		app := lotTestApp(t, true, lots, trackingResolver("batch"))

		resp, err := doRequest(app, http.MethodPut, "/batches/1", `{"item_id":5,"name":"LOT-001"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWriteLotError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "variant not found", err: inventory.ErrVariantNotFound, status: http.StatusUnprocessableEntity},
		{name: "template not found", err: inventory.ErrItemNotFound, status: http.StatusUnprocessableEntity},
		{name: "tracking disabled", err: inventory.ErrBatchTrackingDisabled, status: http.StatusUnprocessableEntity},
		{name: "batch name taken", err: inventory.ErrBatchNameTaken, status: http.StatusConflict},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeLotError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
