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

func countTestApp(t *testing.T, withTenant bool, counts inventory.StockCountDAOMock, lines inventory.StockCountLineDAOMock, quants inventory.StockBalanceDAOMock, layers inventory.CostLayerDAOMock, locations inventory.StockLocationDAOMock, resolver inventory.ItemResolverMock, poster inventory.PosterMock, tx inventory.TransactionerMock) *fiber.App {
	t.Helper()
	ledger := inventory.NewLedgerService(inventory.StockMovementDAOMock{}, quants, locations, inventory.TransactionerMock{})
	svc := inventory.NewStockCountService(counts, lines, quants, layers, ledger, resolver, poster, tx)
	h := NewStockCountHandler(svc)
	return inventoryTestApp(t, withTenant, h.Register)
}

func sampleStockCount() *inventory.StockCount {
	return &inventory.StockCount{
		Base:           model.Base{ID: 1},
		OrganizationID: inventoryUint64Ptr(10),
		Name:           inventoryStringPtr("Cycle Count"),
		LocationID:     inventoryUint64Ptr(100),
		State:          inventory.CountStateDraft,
	}
}

func sampleStockLocationForCount() *reference.StockLocation {
	return &reference.StockLocation{
		Base:           model.Base{ID: 100},
		OrganizationID: inventoryUint64Ptr(10),
		Name:           "Storage",
		Usage:          "internal",
	}
}

func sampleStockCountLine() *inventory.StockCountLine {
	return &inventory.StockCountLine{
		Base:           model.Base{ID: 1},
		StockCountID:   1,
		ItemID:         1,
		TheoreticalQty: 10,
		CountedQty:     12,
		DiffQty:        2,
	}
}

func TestStockCountHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockCount], error) {
			return &query.Page[inventory.StockCount]{Items: []*inventory.StockCount{sampleStockCount()}, Count: 1}, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/?page=1&size=10", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockCount], error) {
			return &query.Page[inventory.StockCount]{Items: []*inventory.StockCount{sampleStockCount()}, Count: 1}, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := countTestApp(t, false, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.StockCount], error) {
			return nil, errors.New("boom")
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockCountHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return &inventory.StockCount{Base: model.Base{ID: 1}, OrganizationID: inventoryUint64Ptr(99)}, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, errors.New("boom")
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockCountHandlerCreate(t *testing.T) {
	t.Run("success with lines", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocationForCount(), nil
		}
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, locations, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		body := `{"name":"Count","location_id":100,"lines":[{"item_id":1,"counted_qty":12}]}`
		resp, err := doRequest(app, http.MethodPost, "/stock-counts", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts", `{"name":"Count","location_id":100,"count_date":"bogus"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts", `{"name":"Count"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("service error", func(t *testing.T) {
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return nil, errors.New("boom")
		}
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, locations, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		body := `{"name":"Count","location_id":100,"lines":[{"item_id":1,"counted_qty":12}]}`
		resp, err := doRequest(app, http.MethodPost, "/stock-counts", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockCountHandlerListLines(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		lines := inventory.StockCountLineDAOMock{}
		lines.ListByCountFunc = func(_ context.Context, _ uint64) ([]*inventory.StockCountLine, error) {
			return []*inventory.StockCountLine{sampleStockCountLine()}, nil
		}
		app := countTestApp(t, true, counts, lines, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		lines := inventory.StockCountLineDAOMock{}
		lines.ListByCountFunc = func(_ context.Context, _ uint64) ([]*inventory.StockCountLine, error) {
			return []*inventory.StockCountLine{sampleStockCountLine()}, nil
		}
		app := countTestApp(t, true, counts, lines, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1/lines?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/abc/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("count lookup error", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, errors.New("boom")
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("line list error", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		lines := inventory.StockCountLineDAOMock{}
		lines.ListByCountFunc = func(_ context.Context, _ uint64) ([]*inventory.StockCountLine, error) {
			return nil, errors.New("boom")
		}
		app := countTestApp(t, true, counts, lines, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/stock-counts/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestStockCountHandlerPost(t *testing.T) {
	body := `{"journal_id":1,"gain_loss_account_id":2}`

	t.Run("success", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		lines := inventory.StockCountLineDAOMock{}
		lines.ListByCountFunc = func(_ context.Context, _ uint64) ([]*inventory.StockCountLine, error) {
			return []*inventory.StockCountLine{sampleStockCountLine()}, nil
		}
		layers := inventory.CostLayerDAOMock{}
		layers.ListOpenByItemFunc = func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{RemainingQty: 1, RemainingValue: 100}}, nil
		}
		resolver := inventory.ItemResolverMock{}
		resolver.ResolveFunc = func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 500}}, nil
		}
		app := countTestApp(t, true, counts, lines, inventory.StockBalanceDAOMock{}, layers, inventory.StockLocationDAOMock{}, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/1/post", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/abc/post", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/1/post", `{"journal_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/1/post", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("count lookup error", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, errors.New("boom")
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/1/post", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("already posted", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return &inventory.StockCount{Base: model.Base{ID: 1}, OrganizationID: inventoryUint64Ptr(10), State: inventory.CountStatePosted}, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/1/post", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("no lines", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		lines := inventory.StockCountLineDAOMock{}
		lines.ListByCountFunc = func(_ context.Context, _ uint64) ([]*inventory.StockCountLine, error) {
			return []*inventory.StockCountLine{}, nil
		}
		app := countTestApp(t, true, counts, lines, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/1/post", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no valuation account", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		lines := inventory.StockCountLineDAOMock{}
		lines.ListByCountFunc = func(_ context.Context, _ uint64) ([]*inventory.StockCountLine, error) {
			return []*inventory.StockCountLine{sampleStockCountLine()}, nil
		}
		layers := inventory.CostLayerDAOMock{}
		layers.ListOpenByItemFunc = func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{RemainingQty: 1, RemainingValue: 100}}, nil
		}
		resolver := inventory.ItemResolverMock{}
		resolver.ResolveFunc = func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{Tracking: "none"}, nil
		}
		app := countTestApp(t, true, counts, lines, inventory.StockBalanceDAOMock{}, layers, inventory.StockLocationDAOMock{}, resolver, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/stock-counts/1/post", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})
}

func TestStockCountHandlerDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := countTestApp(t, true, inventory.StockCountDAOMock{}, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-counts/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, nil
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("lookup error", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return nil, errors.New("boom")
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		counts := inventory.StockCountDAOMock{}
		counts.FindFunc = func(_ context.Context, _ uint64) (*inventory.StockCount, error) {
			return sampleStockCount(), nil
		}
		counts.DeleteFunc = func(_ context.Context, _ uint64) error {
			return errors.New("boom")
		}
		app := countTestApp(t, true, counts, inventory.StockCountLineDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.CostLayerDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodDelete, "/stock-counts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWriteCountError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: inventory.ErrStockCountNotFound, status: http.StatusNotFound},
		{name: "state", err: inventory.ErrStockCountState, status: http.StatusConflict},
		{name: "no lines", err: inventory.ErrStockCountNoLines, status: http.StatusUnprocessableEntity},
		{name: "location required", err: inventory.ErrLocationRequired, status: http.StatusUnprocessableEntity},
		{name: "gain loss account", err: inventory.ErrGainLossAccount, status: http.StatusUnprocessableEntity},
		{name: "organization missing", err: inventory.ErrOrganizationMissing, status: http.StatusUnprocessableEntity},
		{name: "valuation account", err: inventory.ErrValuationAccount, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeCountError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
