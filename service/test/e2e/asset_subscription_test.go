//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestFixedAssetLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	categoryID := createAssetCategory(t, e, f)

	asset := f.authed(t, http.MethodPost, "/fixed-assets").
		WithJSON(map[string]any{
			"name":           "E2E " + gofakeit.Word(),
			"category_id":    categoryID,
			"purchase_value": 12000000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("fixed_asset").
		Object()

	assetID := uint64(asset.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/fixed-assets/"+itoa(assetID)+"/schedule").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/fixed-assets/"+itoa(assetID)+"/post-depreciation").
		WithJSON(map[string]any{
			"journal_id": f.journalGeneralID,
			"date":       "2026-06-30",
		}).
		Expect().
		Status(http.StatusCreated)
}

func TestSubscriptionLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	planID := createSubscriptionPlan(t, e, f)

	subscription := f.authed(t, http.MethodPost, "/subscriptions").
		WithJSON(map[string]any{
			"name":          "E2E " + gofakeit.Word(),
			"contact_id":    f.contactID,
			"plan_id":       planID,
			"price_book_id": f.price_bookID,
			"currency_code": "IDR",
			"lines": []map[string]any{
				{"item_id": f.variantID, "qty": 1, "unit_price": 1000000},
			},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("subscription").
		Object()

	subscriptionID := uint64(subscription.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/subscriptions/"+itoa(subscriptionID)+"/activate").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/subscriptions/metrics").
		Expect().
		Status(http.StatusOK)
}

func TestCommissionLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	_ = e

	plan := f.authed(t, http.MethodPost, "/commission-plans").
		WithJSON(map[string]any{
			"name":  "E2E " + gofakeit.Word(),
			"basis": "revenue",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("commission_plan").
		Object()

	planID := uint64(plan.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/commission-plans/"+itoa(planID)+"/rules").
		WithJSON(map[string]any{"rate": 5}).
		Expect().
		Status(http.StatusCreated)

	f.authed(t, http.MethodGet, "/commission-entries").
		Expect().
		Status(http.StatusOK)
}

func TestGiftCardAndCouponLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	card := f.authed(t, http.MethodPost, "/gift-cards").
		WithJSON(map[string]any{
			"amount":   500000,
			"currency": "IDR",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("gift_card").
		Object()

	cardID := uint64(card.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/gift-cards/"+itoa(cardID)+"/redeem").
		WithJSON(map[string]any{"amount": 100000}).
		Expect().
		Status(http.StatusOK)

	coupon := f.authed(t, http.MethodPost, "/coupons").
		WithJSON(map[string]any{
			"code":         "E2E-" + gofakeit.UUID(),
			"discount_pct": 10,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("coupon").
		Object()

	_ = coupon
}
