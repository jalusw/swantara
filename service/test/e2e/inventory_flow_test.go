//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestInventoryReceiveAndReserveE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.buildStock(t, e, 20, 600000)

	reservation := f.authed(t, http.MethodPost, "/stock-reservations").
		WithJSON(map[string]any{
			"item_id":     f.variantID,
			"location_id": f.stockLocationID,
			"qty":         "5",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	reservation.NotEmpty()

	available := f.authed(t, http.MethodGet, "/stock/available-to-promise").
		WithQuery("item_id", itoa(f.variantID)).
		WithQuery("location_id", itoa(f.stockLocationID)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object()

	available.NotEmpty()
}

func TestInventoryTransferLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.buildStock(t, e, 10, 600000)

	secondWarehouse := e.POST(f.orgPath+"/warehouses").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"name": "E2E " + gofakeit.Word(),
			"code": "WH-" + gofakeit.UUID(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("warehouse").
		Object()
	dstWarehouseID := uint64(secondWarehouse.Value("id").Number().Raw())

	transitAccount := firstAccountID(t, e, f)

	transfer := f.authed(t, http.MethodPost, "/transfer-orders").
		WithJSON(map[string]any{
			"src_warehouse_id": f.warehouseID,
			"dst_warehouse_id": dstWarehouseID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 4},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("transfer").
		Object()

	transferID := uint64(transfer.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/transfer-orders/"+itoa(transferID)+"/send").
		WithJSON(map[string]any{
			"journal_id":         f.journalGeneralID,
			"transit_account_id": transitAccount,
		}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/transfer-orders/"+itoa(transferID)+"/receive").
		WithJSON(map[string]any{
			"journal_id":         f.journalGeneralID,
			"transit_account_id": transitAccount,
		}).
		Expect().
		Status(http.StatusOK)
}

func TestStockCountLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	f.buildStock(t, newExpect(t), 10, 600000)

	count := f.authed(t, http.MethodPost, "/inventory-counts").
		WithJSON(map[string]any{
			"name":        "E2E " + gofakeit.Word(),
			"location_id": f.stockLocationID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "counted_qty": 10},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	count.NotEmpty()
}

func TestInventoryReorderRuleLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	rule := f.authed(t, http.MethodPost, "/reorder-rules").
		WithJSON(map[string]any{
			"item_id":      f.variantID,
			"warehouse_id": f.warehouseID,
			"min_qty":      5,
			"max_qty":      50,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	rule.NotEmpty()

	f.authed(t, http.MethodGet, "/reorder-rules").
		Expect().
		Status(http.StatusOK)
}
