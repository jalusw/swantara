package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func inboundCostTestApp(t *testing.T, withTenant bool, costs inventory.InboundCostDAOMock, lines inventory.InboundCostLineDAOMock, adjustments inventory.InboundCostAdjustmentDAOMock, movements inventory.StockMovementDAOMock, layers inventory.CostLayerDAOMock, resolver inventory.ItemResolverMock, poster inventory.PosterMock, configs inventory.InboundCostConfigSourceMock, tx inventory.TransactionerMock) *fiber.App {
	t.Helper()
	svc := inventory.NewInboundCostService(costs, lines, adjustments, movements, layers, resolver, poster, configs, nil, tx)
	h := NewInboundCostHandler(svc)
	return inventoryTestApp(t, withTenant, h.Register)
}

func sampleInboundCost() *inventory.InboundCost {
	return &inventory.InboundCost{
		Base:              model.Base{ID: 1},
		OrganizationID:    inventoryUint64Ptr(10),
		Name:              "Freight",
		State:             inventory.InboundCostStateDraft,
		TargetShipmentIDs: helper.Int64Array{1},
	}
}

func sampleInboundCostLine() *inventory.InboundCostLine {
	return &inventory.InboundCostLine{
		Base:          model.Base{ID: 1},
		InboundCostID: 1,
		ItemID:        1,
		Amount:        100,
		SplitMethod:   inventory.SplitMethodQuantity,
		AccountID:     inventoryUint64Ptr(2),
	}
}

func sampleInboundCostAdjustment() *inventory.InboundCostAdjustment {
	return &inventory.InboundCostAdjustment{
		Base:            model.Base{ID: 1},
		InboundCostID:   1,
		StockMovementID: 5,
		ItemID:          1,
		AdditionalCost:  100,
		CostLayerID:     9,
	}
}

func TestInboundCostHandlerList(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.ListFunc = func(_ context.Context, q *query.Query) (*query.Page[inventory.InboundCost], error) {
			return &query.Page[inventory.InboundCost]{Items: []*inventory.InboundCost{sampleInboundCost()}, Count: 1}, nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("missing organization", func(t *testing.T) {
		app := inboundCostTestApp(t, false, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("dao error", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.ListFunc = func(_ context.Context, q *query.Query) (*query.Page[inventory.InboundCost], error) {
			return nil, errors.New("boom")
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestInboundCostHandlerGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return sampleInboundCost(), nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("missing organization", func(t *testing.T) {
		app := inboundCostTestApp(t, false, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := inboundCostTestApp(t, true, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return nil, nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("tenant mismatch", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return &inventory.InboundCost{Base: model.Base{ID: 1}, OrganizationID: inventoryUint64Ptr(99)}, nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("search error", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return nil, errors.New("boom")
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestInboundCostHandlerListLines(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return sampleInboundCost(), nil
		}
		lines := inventory.InboundCostLineDAOMock{}
		lines.ListByInboundCostFunc = func(_ context.Context, _ uint64) ([]*inventory.InboundCostLine, error) {
			return []*inventory.InboundCostLine{sampleInboundCostLine()}, nil
		}
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("missing organization", func(t *testing.T) {
		app := inboundCostTestApp(t, false, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := inboundCostTestApp(t, true, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/abc/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return nil, nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("lines error", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return sampleInboundCost(), nil
		}
		lines := inventory.InboundCostLineDAOMock{}
		lines.ListByInboundCostFunc = func(_ context.Context, _ uint64) ([]*inventory.InboundCostLine, error) {
			return nil, errors.New("boom")
		}
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/lines", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestInboundCostHandlerListAdjustments(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return sampleInboundCost(), nil
		}
		adjustments := inventory.InboundCostAdjustmentDAOMock{}
		adjustments.ListByInboundCostFunc = func(_ context.Context, _ uint64) ([]*inventory.InboundCostAdjustment, error) {
			return []*inventory.InboundCostAdjustment{sampleInboundCostAdjustment()}, nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, adjustments, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/adjustments", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("missing organization", func(t *testing.T) {
		app := inboundCostTestApp(t, false, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/adjustments", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := inboundCostTestApp(t, true, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/abc/adjustments", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return nil, nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/adjustments", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("adjustments error", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return sampleInboundCost(), nil
		}
		adjustments := inventory.InboundCostAdjustmentDAOMock{}
		adjustments.ListByInboundCostFunc = func(_ context.Context, _ uint64) ([]*inventory.InboundCostAdjustment, error) {
			return nil, errors.New("boom")
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, adjustments, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodGet, "/inbound-costs/1/adjustments", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestInboundCostHandlerCreate(t *testing.T) {
	body := `{"name":"Freight","target_shipment_ids":[1],"lines":[{"item_id":1,"amount":100,"split_method":"by_quantity","account_id":2}]}`

	t.Run("success", func(t *testing.T) {
		app := inboundCostTestApp(t, true, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		app := inboundCostTestApp(t, true, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs", `{"name":"Freight"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("missing organization", func(t *testing.T) {
		app := inboundCostTestApp(t, false, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("service error", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.CreateTxFunc = func(_ context.Context, _ *gorm.DB, cost *inventory.InboundCost) (*inventory.InboundCost, error) {
			return nil, errors.New("boom")
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs", body)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestInboundCostHandlerPost(t *testing.T) {
	postMocks := func(state string, shipmentIDs helper.Int64Array, withLine bool, journalID uint64) (inventory.InboundCostDAOMock, inventory.InboundCostLineDAOMock, inventory.StockMovementDAOMock, inventory.CostLayerDAOMock, inventory.InboundCostConfigSourceMock) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return &inventory.InboundCost{
				Base:              model.Base{ID: 1},
				OrganizationID:    inventoryUint64Ptr(10),
				Name:              "Freight",
				State:             state,
				TargetShipmentIDs: shipmentIDs,
			}, nil
		}
		lines := inventory.InboundCostLineDAOMock{}
		if withLine {
			lines.ListByInboundCostFunc = func(_ context.Context, _ uint64) ([]*inventory.InboundCostLine, error) {
				return []*inventory.InboundCostLine{sampleInboundCostLine()}, nil
			}
		}
		movements := inventory.StockMovementDAOMock{}
		movements.ListByShipmentFunc = func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{{Base: model.Base{ID: 5}, ItemID: 1, Qty: 2}}, nil
		}
		layers := inventory.CostLayerDAOMock{}
		layers.ListByMovementFunc = func(_ context.Context, _ uint64) ([]*inventory.CostLayer, error) {
			return []*inventory.CostLayer{{Base: model.Base{ID: 9}, RemainingQty: 1}}, nil
		}
		configs := inventory.InboundCostConfigSourceMock{}
		configs.JournalIDFunc = func(_ context.Context, _ uint64) (uint64, error) {
			return journalID, nil
		}
		return costs, lines, movements, layers, configs
	}

	t.Run("success", func(t *testing.T) {
		costs, lines, movements, layers, configs := postMocks(inventory.InboundCostStateDraft, helper.Int64Array{1}, true, 9)
		resolver := inventory.ItemResolverMock{}
		resolver.ResolveFunc = func(_ context.Context, _ uint64) (inventory.ResolvedItem, error) {
			return inventory.ResolvedItem{StockAccounts: inventory.StockAccounts{StockValuationAccountID: 500}}, nil
		}
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, movements, layers, resolver, inventory.PosterMock{}, configs, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("missing organization", func(t *testing.T) {
		app := inboundCostTestApp(t, false, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := inboundCostTestApp(t, true, inventory.InboundCostDAOMock{}, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/abc/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("not found", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return nil, nil
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", resp.StatusCode)
		}
	})

	t.Run("invalid state", func(t *testing.T) {
		costs, lines, movements, layers, configs := postMocks(inventory.InboundCostStatePosted, helper.Int64Array{1}, true, 9)
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, movements, layers, inventory.ItemResolverMock{}, inventory.PosterMock{}, configs, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusConflict {
			t.Fatalf("expected 409, got %d", resp.StatusCode)
		}
	})

	t.Run("no lines", func(t *testing.T) {
		costs, lines, movements, layers, configs := postMocks(inventory.InboundCostStateDraft, helper.Int64Array{1}, false, 9)
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, movements, layers, inventory.ItemResolverMock{}, inventory.PosterMock{}, configs, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no shipments", func(t *testing.T) {
		costs, lines, movements, layers, configs := postMocks(inventory.InboundCostStateDraft, helper.Int64Array{}, true, 9)
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, movements, layers, inventory.ItemResolverMock{}, inventory.PosterMock{}, configs, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no journal", func(t *testing.T) {
		costs, lines, movements, layers, configs := postMocks(inventory.InboundCostStateDraft, helper.Int64Array{1}, true, 0)
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, movements, layers, inventory.ItemResolverMock{}, inventory.PosterMock{}, configs, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("no movements", func(t *testing.T) {
		costs, lines, _, layers, configs := postMocks(inventory.InboundCostStateDraft, helper.Int64Array{1}, true, 9)
		movements := inventory.StockMovementDAOMock{}
		movements.ListByShipmentFunc = func(_ context.Context, _ uint64) ([]*inventory.StockMovement, error) {
			return []*inventory.StockMovement{}, nil
		}
		app := inboundCostTestApp(t, true, costs, lines, inventory.InboundCostAdjustmentDAOMock{}, movements, layers, inventory.ItemResolverMock{}, inventory.PosterMock{}, configs, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("expected 422, got %d", resp.StatusCode)
		}
	})

	t.Run("search error", func(t *testing.T) {
		costs := inventory.InboundCostDAOMock{}
		costs.SearchFunc = func(_ context.Context, _ string, _ any) (*inventory.InboundCost, error) {
			return nil, errors.New("boom")
		}
		app := inboundCostTestApp(t, true, costs, inventory.InboundCostLineDAOMock{}, inventory.InboundCostAdjustmentDAOMock{}, inventory.StockMovementDAOMock{}, inventory.CostLayerDAOMock{}, inventory.ItemResolverMock{}, inventory.PosterMock{}, inventory.InboundCostConfigSourceMock{}, inventory.TransactionerMock{})

		resp, err := doRequest(app, http.MethodPost, "/inbound-costs/1/post", "")
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", resp.StatusCode)
		}
	})
}

func TestWriteInboundCostError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "not found", err: inventory.ErrInboundCostNotFound, status: http.StatusNotFound},
		{name: "state", err: inventory.ErrInboundCostState, status: http.StatusConflict},
		{name: "no lines", err: inventory.ErrInboundCostNoLines, status: http.StatusUnprocessableEntity},
		{name: "no shipments", err: inventory.ErrInboundCostNoShipments, status: http.StatusUnprocessableEntity},
		{name: "no movements", err: inventory.ErrInboundCostNoMovements, status: http.StatusUnprocessableEntity},
		{name: "no account", err: inventory.ErrInboundCostNoAccount, status: http.StatusUnprocessableEntity},
		{name: "no journal", err: inventory.ErrInboundCostNoJournal, status: http.StatusUnprocessableEntity},
		{name: "split", err: inventory.ErrInboundCostSplit, status: http.StatusUnprocessableEntity},
		{name: "value", err: inventory.ErrInboundCostValue, status: http.StatusUnprocessableEntity},
		{name: "valuation account", err: inventory.ErrValuationAccount, status: http.StatusUnprocessableEntity},
		{name: "organization missing", err: inventory.ErrOrganizationMissing, status: http.StatusUnprocessableEntity},
		{name: "default", err: errors.New("boom"), status: http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := writeErrorStatus("/test", func(c fiber.Ctx) error {
				return writeInboundCostError(c, test.err)
			})
			if status != test.status {
				t.Fatalf("expected %d, got %d", test.status, status)
			}
		})
	}
}
