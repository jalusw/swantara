//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestOrderToCashE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	opportunityID := f.createWonOpportunity(t, e)
	f.buildStock(t, e, 10, 600000)

	order := f.authed(t, http.MethodPost, "/sale-orders").
		WithJSON(map[string]any{
			"contact_id":      f.contactID,
			"prospect_id":     opportunityID,
			"warehouse_id":    f.warehouseID,
			"price_book_id":   f.price_bookID,
			"payment_term_id": 1,
			"lines": []map[string]any{
				{
					"item_id":     f.variantID,
					"qty_ordered": 10,
					"tax_ids":     []uint64{f.taxOutputID},
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
	order.Value("amount_untaxed").Number().Equal(10000000)
	order.Value("amount_tax").Number().Equal(1100000)
	order.Value("amount_total").Number().Equal(11100000)

	f.authed(t, http.MethodPost, "/sale-orders/"+itoa(orderID)+"/send").
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

	key := "o2c-" + gofakeit.UUID()

	f.authed(t, http.MethodPost, "/sale-orders/"+itoa(orderID)+"/confirm").
		WithHeader("Idempotency-Key", key).
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
		Equal("confirmed")

	delivered := f.authed(t, http.MethodPost, "/sale-orders/"+itoa(orderID)+"/deliver").
		WithJSON(map[string]any{"journal_id": f.journalSaleID}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object()

	delivered.Value("delivery_status").String().Equal("done")

	invoice := f.authed(t, http.MethodPost, "/sale-orders/"+itoa(orderID)+"/invoice").
		WithJSON(map[string]any{"journal_id": f.journalSaleID}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("invoice").
		Object()

	invoice.Value("state").String().Equal("posted")
	invoice.Value("amount_total").Number().Equal(11100000)

	invoiced := f.authed(t, http.MethodGet, "/sale-orders/"+itoa(orderID)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object()

	invoiced.Value("invoice_status").String().Equal("invoiced")
	invoiced.Value("lines").Array().Element(0).Object().Value("qty_delivered").Number().Equal(10)

	payment := f.authed(t, http.MethodPost, "/sale-orders/"+itoa(orderID)+"/pay").
		WithJSON(map[string]any{
			"journal_id": f.journalBankID,
			"amount":     11100000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("payment").
		Object()

	payment.Value("amount").Number().Equal(11100000)

	done := f.authed(t, http.MethodPost, "/sale-orders/"+itoa(orderID)+"/done").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("order").
		Object()

	done.Value("state").String().Equal("done")
}
