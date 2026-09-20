//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestServiceOrderLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	equipmentID := createEquipment(t, e, f)

	contract := f.authed(t, http.MethodPost, "/service-contracts").
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	contract.NotEmpty()

	order := f.authed(t, http.MethodPost, "/service-orders").
		WithJSON(map[string]any{
			"name":         "E2E " + gofakeit.Word(),
			"type":         "repair",
			"equipment_id": equipmentID,
			"contact_id":   f.contactID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("service_order").
		Object()

	orderID := uint64(order.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/service-orders/"+itoa(orderID)+"/schedule").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/service-orders/"+itoa(orderID)+"/start").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/service-orders/"+itoa(orderID)+"/complete").
		Expect().
		Status(http.StatusOK)
}

func TestQualityCheckLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	pointID := createQualityPoint(t, e, f)

	checks := f.authed(t, http.MethodGet, "/quality-checks").
		WithQuery("point_id", itoa(pointID)).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("checks").
		Array()

	_ = checks

	f.authed(t, http.MethodGet, "/quality-alerts").
		Expect().
		Status(http.StatusOK)
}

func TestRMALifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	_ = e

	f.buildStock(t, newExpect(t), 10, 600000)

	opportunityID := f.createWonOpportunity(t, newExpect(t))

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
					"qty_ordered": 2,
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

	rma := f.authed(t, http.MethodPost, "/rmas").
		WithJSON(map[string]any{
			"type":              "customer",
			"contact_id":        f.contactID,
			"origin_order_type": "sale",
			"origin_order_id":   orderID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 1, "disposition": "restock"},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("rma").
		Object()

	rmaID := uint64(rma.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/rmas/"+itoa(rmaID)+"/confirm").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/rmas/"+itoa(rmaID)+"/receive").
		WithJSON(map[string]any{"journal_id": f.journalGeneralID}).
		Expect().
		Status(http.StatusOK)
}
