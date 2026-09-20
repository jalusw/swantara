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

func reorderTestApp(t *testing.T, withTenant bool, rules inventory.ReorderRuleDAOMock, movements inventory.StockMovementDAOMock, quants inventory.StockBalanceDAOMock, locations inventory.StockLocationDAOMock, resolver inventory.ItemResolverMock) *fiber.App {
	t.Helper()
	ledger := inventory.NewLedgerService(movements, quants, locations, inventory.TransactionerMock{})
	svc := inventory.NewReorderService(rules, ledger, resolver)
	handler := NewReorderHandler(svc)
	return inventoryTestApp(t, withTenant, handler.Register)
}

func sampleReorderRule() *inventory.ReorderRule {
	return &inventory.ReorderRule{
		Base:        model.Base{ID: 1},
		ItemID:      5,
		WarehouseID: inventoryUint64Ptr(1),
		LocationID:  inventoryUint64Ptr(2),
		MinQty:      10,
		MaxQty:      20,
		QtyMultiple: 5,
		Active:      true,
	}
}

func reorderResolver() inventory.ItemResolverMock {
	return inventory.ItemResolverMock{
		ResolveFunc: func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{Tracking: "none", OrganizationID: inventoryUint64Ptr(10)}, nil
		},
	}
}

func TestReorderHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.ReorderRule], error) {
			return &query.Page[inventory.ReorderRule]{Items: []*inventory.ReorderRule{sampleReorderRule()}, Count: 1}, nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid query", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/?filter=bogus:eq:1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reorderTestApp(t, false, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.ReorderRule], error) {
			return &query.Page[inventory.ReorderRule]{Items: []*inventory.ReorderRule{sampleReorderRule()}, Count: 1}, nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.ListFunc = func(_ context.Context, _ *query.Query) (*query.Page[inventory.ReorderRule], error) {
			return nil, errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestReorderHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return sampleReorderRule(), nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return nil, nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reorderTestApp(t, false, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return nil, errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestReorderHandlerCreate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.CreateFunc = func(_ context.Context, rule *inventory.ReorderRule) (*inventory.ReorderRule, error) {
			rule.ID = 1
			return rule, nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reorderTestApp(t, false, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("location required", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid quantities", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5,"location_id":2,"min_qty":5,"max_qty":1}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid multiple", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5,"location_id":2,"qty_multiple":0}`)
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
				return inventory.ResolvedItem{Tracking: "none"}, nil
			},
		}
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, resolver)

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("create error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.CreateFunc = func(_ context.Context, _ *inventory.ReorderRule) (*inventory.ReorderRule, error) {
			return nil, errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("inactive", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.CreateFunc = func(_ context.Context, rule *inventory.ReorderRule) (*inventory.ReorderRule, error) {
			rule.ID = 1
			return rule, nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPost, "/reorder-rules/", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5,"active":false}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})
}

func TestReorderHandlerUpdate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return sampleReorderRule(), nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/1", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return nil, nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/1", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/abc", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reorderTestApp(t, false, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/1", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return sampleReorderRule(), nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/1", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return nil, errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/1", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("update error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return sampleReorderRule(), nil
		}
		rules.UpdateFunc = func(_ context.Context, _ *inventory.ReorderRule) (*inventory.ReorderRule, error) {
			return nil, errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/1", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("active provided", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return sampleReorderRule(), nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, reorderResolver())

		resp, err := doRequest(app, http.MethodPut, "/reorder-rules/1", `{"item_id":5,"location_id":2,"min_qty":10,"max_qty":20,"qty_multiple":5,"active":false}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestReorderHandlerDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return sampleReorderRule(), nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return nil, nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := reorderTestApp(t, true, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/reorder-rules/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reorderTestApp(t, false, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("find error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return nil, errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.FindFunc = func(_ context.Context, _ uint64) (*inventory.ReorderRule, error) {
			return sampleReorderRule(), nil
		}
		rules.DeleteFunc = func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodDelete, "/reorder-rules/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestReorderHandlerCandidates(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.ListActiveFunc = func(_ context.Context) ([]*inventory.ReorderRule, error) {
			return []*inventory.ReorderRule{sampleReorderRule()}, nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/candidates", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("no tenant", func(t *testing.T) {
		app := reorderTestApp(t, false, inventory.ReorderRuleDAOMock{}, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/candidates", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("list error", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.ListActiveFunc = func(_ context.Context) ([]*inventory.ReorderRule, error) {
			return nil, errors.New("db down")
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, inventory.StockLocationDAOMock{}, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/candidates", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})

	t.Run("csv", func(t *testing.T) {
		rules := inventory.ReorderRuleDAOMock{}
		rules.ListActiveFunc = func(_ context.Context) ([]*inventory.ReorderRule, error) {
			return []*inventory.ReorderRule{sampleReorderRule()}, nil
		}
		locations := inventory.StockLocationDAOMock{}
		locations.FindFunc = func(_ context.Context, _ uint64) (*reference.StockLocation, error) {
			return sampleStockLocation(), nil
		}
		app := reorderTestApp(t, true, rules, inventory.StockMovementDAOMock{}, inventory.StockBalanceDAOMock{}, locations, inventory.ItemResolverMock{})

		resp, err := doRequest(app, http.MethodGet, "/reorder-rules/candidates?format=csv", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})
}

func TestWriteReorderError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "invalid rule qty", err: inventory.ErrInvalidRuleQty, status: http.StatusUnprocessableEntity},
		{name: "location required", err: inventory.ErrLocationRequired, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeReorderError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
