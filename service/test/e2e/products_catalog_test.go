//go:build e2e

package e2e

import (
	"net/http"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
)

func TestProductCatalogLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	category := f.authed(t, http.MethodPost, "/item-categories").
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("category").
		Object()

	categoryID := uint64(category.Value("id").Number().Raw())

	item := f.authed(t, http.MethodPost, "/products").
		WithJSON(map[string]any{
			"name":           gofakeit.ProductName(),
			"type":           "stockable",
			"tracking":       "none",
			"is_sellable":    true,
			"is_purchasable": true,
			"list_price":     250000,
			"standard_cost":  150000,
			"category_id":    categoryID,
			"variants":       []map[string]any{{"sku": "SKU-" + gofakeit.UUID()}},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	item.NotEmpty()

	f.authed(t, http.MethodGet, "/products").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/item-categories").
		Expect().
		Status(http.StatusOK)
}

func TestPriceBookResolveLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)

	resolved := f.authed(t, http.MethodPost, "/price_books/"+itoa(f.price_bookID)+"/resolve").
		WithJSON(map[string]any{
			"price_book_id": f.price_bookID,
			"item_id":       f.variantID,
			"qty":           3,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object()

	resolved.NotEmpty()
}

func TestSupplierProductBestOfferE2E(t *testing.T) {
	f := newFlowFixture(t)

	supplierProduct := f.authed(t, http.MethodPost, "/supplier-products").
		WithJSON(map[string]any{
			"item_id":     f.variantID,
			"supplier_id": f.contactID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	supplierProduct.NotEmpty()

	f.authed(t, http.MethodPost, "/supplier-products/best-offer").
		WithJSON(map[string]any{
			"item_id": f.variantID,
			"qty":     5,
		}).
		Expect().
		Status(http.StatusOK)
}
