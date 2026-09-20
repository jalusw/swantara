//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestBatchLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	lot := f.authed(t, http.MethodPost, "/stock-lots").
		WithJSON(map[string]any{
			"item_id": f.variantID,
			"name":    "LOT-" + gofakeit.UUID(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("lot").
		Object()

	lotID := uint64(lot.Value("id").Number().Raw())

	f.authed(t, http.MethodPut, "/stock-lots/"+itoa(lotID)).
		WithJSON(map[string]any{"name": "LOT-" + gofakeit.UUID()}).
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/stock-lots/"+itoa(lotID)).
		Expect().
		Status(http.StatusOK)
}

func TestShipmentsReadE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.buildStock(t, e, 10, 600000)

	shipmentID := firstShipmentID(t, e, f)

	f.authed(t, http.MethodGet, "/stock-shipments/"+itoa(shipmentID)).
		Expect().
		Status(http.StatusOK)
}

func TestStockCountPostE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.buildStock(t, e, 10, 600000)

	accountID := firstAccountID(t, e, f)

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
		Object().
		Value("count").
		Object()

	countID := uint64(count.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/inventory-counts/"+itoa(countID)+"/lines").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/inventory-counts/"+itoa(countID)+"/post").
		WithJSON(map[string]any{
			"journal_id":           f.journalGeneralID,
			"gain_loss_account_id": accountID,
		}).
		Expect().
		Status(http.StatusOK)
}

func TestInboundCostLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.buildStock(t, e, 10, 600000)

	shipmentID := firstShipmentID(t, e, f)
	accountID := firstAccountID(t, e, f)

	landed := f.authed(t, http.MethodPost, "/landed-costs").
		WithJSON(map[string]any{
			"name":                "E2E " + gofakeit.Word(),
			"target_shipment_ids": []uint64{shipmentID},
			"lines": []map[string]any{
				{
					"item_id":      f.variantID,
					"description":  "Freight",
					"amount":       500000,
					"split_method": "by_quantity",
					"account_id":   accountID,
				},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	landedID := uint64(landed.Value("id").Number().Raw())

	f.authed(t, http.MethodGet, "/landed-costs/"+itoa(landedID)+"/lines").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/landed-costs/"+itoa(landedID)+"/post").
		Expect().
		Status(http.StatusOK)
}

func TestReorderCandidatesE2E(t *testing.T) {
	f := newFlowFixture(t)

	f.authed(t, http.MethodPost, "/reorder-rules").
		WithJSON(map[string]any{
			"item_id":      f.variantID,
			"warehouse_id": f.warehouseID,
			"min_qty":      100,
			"max_qty":      500,
		}).
		Expect().
		Status(http.StatusCreated)

	f.authed(t, http.MethodGet, "/reorder-rules/candidates").
		Expect().
		Status(http.StatusOK)
}

func TestStockMovementShipE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	f.buildStock(t, e, 10, 600000)

	customerLocation := f.authed(t, http.MethodPost, "/stock-locations").
		WithJSON(map[string]any{
			"name":         "CUST-" + gofakeit.UUID(),
			"code":         "CUST-" + gofakeit.UUID(),
			"usage":        "customer",
			"warehouse_id": f.warehouseID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("location").
		Object()

	customerLocationID := uint64(customerLocation.Value("id").Number().Raw())

	movement := f.authed(t, http.MethodPost, "/stock-movements").
		WithJSON(map[string]any{
			"item_id":         f.variantID,
			"qty":             "2",
			"src_location_id": f.stockLocationID,
			"dst_location_id": customerLocationID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("movement").
		Object()

	moveID := uint64(movement.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/stock-movements/"+itoa(moveID)+"/ship").
		WithJSON(map[string]any{"journal_id": f.journalSaleID}).
		Expect().
		Status(http.StatusOK)
}
