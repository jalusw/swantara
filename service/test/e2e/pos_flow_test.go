//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestPOSSessionLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	_ = e

	config := f.authed(t, http.MethodPost, "/pos/configs").
		WithJSON(map[string]any{
			"name":          "E2E " + gofakeit.Word(),
			"warehouse_id":  f.warehouseID,
			"journal_id":    f.journalSaleID,
			"price_book_id": f.price_bookID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("config").
		Object()

	configID := uint64(config.Value("id").Number().Raw())

	admin := login(t, adminEmail, adminPassword)
	cashierID := defaultUserID(t, admin.AccessToken)

	session := f.authed(t, http.MethodPost, "/pos/sessions").
		WithJSON(map[string]any{
			"config_id":       configID,
			"cashier_id":      cashierID,
			"opening_balance": 100000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("session").
		Object()

	sessionID := uint64(session.Value("id").Number().Raw())

	order := f.authed(t, http.MethodPost, "/pos/orders").
		WithJSON(map[string]any{
			"config_id":  configID,
			"session_id": sessionID,
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 2},
			},
			"payments": []map[string]any{
				{"method": "cash", "amount": 2000000},
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
	_ = orderID

	f.authed(t, http.MethodPost, "/pos/sessions/"+itoa(sessionID)+"/close").
		Expect().
		Status(http.StatusOK)
}
