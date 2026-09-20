//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestProductionOrderLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	component := e.POST(f.orgPath+"/products").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"name":           gofakeit.ProductName(),
			"type":           "stockable",
			"tracking":       "none",
			"is_sellable":    false,
			"is_purchasable": true,
			"list_price":     100000,
			"standard_cost":  50000,
			"variants":       []map[string]any{{"sku": "SKU-" + gofakeit.UUID()}},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("variants").
		Array().
		Element(0).
		Object()

	componentID := uint64(component.Value("id").Number().Raw())

	recipe := f.authed(t, http.MethodPost, "/recipes").
		WithJSON(map[string]any{
			"item_id": f.variantID,
			"type":    "standard",
			"lines": []map[string]any{
				{"item_id": componentID, "qty": 2},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("recipe").
		Object()

	_ = recipe

	mo := f.authed(t, http.MethodPost, "/manufacturing-orders").
		WithJSON(map[string]any{
			"item_id": f.variantID,
			"qty":     5,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("production_order").
		Object()

	moID := uint64(mo.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/manufacturing-orders/"+itoa(moID)+"/confirm").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/manufacturing-orders/"+itoa(moID)+"/start").
		Expect().
		Status(http.StatusOK)
}

func TestMRPRunLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	run := f.authed(t, http.MethodPost, "/planning/runs").
		WithJSON(map[string]any{"warehouse_id": f.warehouseID}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	run.NotEmpty()

	f.authed(t, http.MethodGet, "/planning/runs").
		Expect().
		Status(http.StatusOK)
}
