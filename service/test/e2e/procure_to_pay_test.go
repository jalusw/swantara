//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestProcureToPayE2E(t *testing.T) {
	f := newFlowFixture(t)

	order := f.authed(t, http.MethodPost, "/purchase-orders").
		WithJSON(map[string]any{
			"supplier_id":  f.contactID,
			"warehouse_id": f.warehouseID,
			"lines": []map[string]any{
				{
					"item_id":     f.variantID,
					"qty_ordered": 10,
					"unit_price":  600000,
					"tax_ids":     []uint64{f.taxInputID},
				},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object()

	orderID := uint64(order.Value("id").Number().Raw())
	order.Value("state").String().Equal("draft")
	order.Value("amount_untaxed").Number().Equal(6000000)
	order.Value("amount_tax").Number().Equal(660000)
	order.Value("amount_total").Number().Equal(6660000)

	key := "p2p-" + gofakeit.UUID()

	f.authed(t, http.MethodPost, "/purchase-orders/"+itoa(orderID)+"/confirm").
		WithHeader("Idempotency-Key", key).
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object().
		Value("state").
		String().
		Equal("sent")

	received := f.authed(t, http.MethodPost, "/purchase-orders/"+itoa(orderID)+"/receive").
		WithHeader("Idempotency-Key", "recv-"+key).
		WithJSON(map[string]any{"journal_id": f.journalGeneralID}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object()

	received.Value("state").String().Equal("confirmed")
	received.Value("receipt_status").String().Equal("done")

	f.authed(t, http.MethodPost, "/purchase-orders/"+itoa(orderID)+"/supplier-bill").
		WithJSON(map[string]any{"journal_id": f.journalPurchaseID}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("invoice").
		Object().
		Value("amount_total").
		Number().
		Equal(6660000)

	billed := f.authed(t, http.MethodGet, "/purchase-orders/"+itoa(orderID)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object()

	billed.Value("invoice_status").String().Equal("invoiced")
	billed.Value("lines").Array().Element(0).Object().Value("qty_billed").Number().Equal(10)

	payment := f.authed(t, http.MethodPost, "/purchase-orders/"+itoa(orderID)+"/pay").
		WithJSON(map[string]any{"journal_id": f.journalBankID}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("payment").
		Object()

	payment.Value("amount").Number().Equal(6660000)

	paid := f.authed(t, http.MethodGet, "/purchase-orders/"+itoa(orderID)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object()

	paid.Value("receipt_status").String().Equal("done")
	paid.Value("invoice_status").String().Equal("invoiced")
}
