//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestPurchaseRequestLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	_ = e

	requisition := f.authed(t, http.MethodPost, "/purchase-requisitions").
		WithJSON(map[string]any{
			"requester_id": 1,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 5, "price": 600000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("requisition").
		Object()

	requisitionID := uint64(requisition.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/purchase-requisitions/"+itoa(requisitionID)+"/confirm").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/purchase-requisitions/"+itoa(requisitionID)+"/approve").
		Expect().
		Status(http.StatusOK)
}

func TestSupplyAgreementLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	agreement := f.authed(t, http.MethodPost, "/purchase-agreements").
		WithJSON(map[string]any{
			"supplier_id": f.contactID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 100, "price": 600000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("agreement").
		Object()

	agreementID := uint64(agreement.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/purchase-agreements/"+itoa(agreementID)+"/activate").
		Expect().
		Status(http.StatusOK)
}

func TestSupplierQuoteRequestLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	quote_request := f.authed(t, http.MethodPost, "/purchase-rfqs").
		WithJSON(map[string]any{
			"supplier_ids": []uint64{f.contactID},
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 10},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	quote_request.NotEmpty()
}

func TestSupplierScorecardReadE2E(t *testing.T) {
	f := newFlowFixture(t)

	f.authed(t, http.MethodGet, "/supplier-scorecards/supplier/"+itoa(f.contactID)).
		Expect().
		Status(http.StatusOK)
}

func TestPurchaseOrderCreditMemoE2E(t *testing.T) {
	f := newFlowFixture(t)

	order := f.authed(t, http.MethodPost, "/purchase-orders").
		WithJSON(map[string]any{
			"supplier_id":  f.contactID,
			"warehouse_id": f.warehouseID,
			"lines": []map[string]any{
				{
					"item_id":     f.variantID,
					"qty_ordered": 4,
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
	_ = gofakeit.Word()

	f.authed(t, http.MethodPost, "/purchase-orders/"+itoa(orderID)+"/confirm").
		WithJSON(map[string]any{}).
		Expect().
		Status(http.StatusOK)
}
