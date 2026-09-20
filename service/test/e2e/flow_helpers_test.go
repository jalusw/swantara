//go:build e2e

package e2e

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/gavv/httpexpect/v2"
)

type flowContext struct {
	tokens             authTokens
	orgPath            string
	contactID          uint64
	variantID          uint64
	warehouseID        uint64
	stockLocationID    uint64
	supplierLocationID uint64
	journalSaleID      uint64
	journalPurchaseID  uint64
	journalBankID      uint64
	journalGeneralID   uint64
	taxOutputID        uint64
	taxInputID         uint64
	price_bookID       uint64
}

func newFlowFixture(t *testing.T) flowContext {
	t.Helper()

	e := newExpect(t)
	admin := login(t, adminEmail, adminPassword)

	orgID := defaultOrganizationID(t, admin.AccessToken)
	orgPath := "/api/v1/organizations/" + itoa(orgID)

	f := flowContext{
		tokens:  admin,
		orgPath: orgPath,
	}

	f.contactID = f.createContact(t, e)
	f.variantID = f.createSellableProduct(t, e)
	f.price_bookID = f.createPriceBook(t, e)
	f.createPriceRule(t, e)
	f.warehouseID = f.createWarehouse(t, e)
	f.stockLocationID = f.ensureStockLocation(t, e, "STOCK", "internal")
	f.supplierLocationID = f.ensureStockLocation(t, e, "SUP", "supplier")
	f.ensureStockLocation(t, e, "CUST", "customer")
	f.journalSaleID = f.journalByCode(t, e, "SALE")
	f.journalPurchaseID = f.journalByCode(t, e, "PURCHASE")
	f.journalBankID = f.journalByCode(t, e, "BANK")
	f.journalGeneralID = f.journalByCode(t, e, "GENERAL")
	f.taxOutputID = f.taxByName(t, e, "PPN Output")
	f.taxInputID = f.taxByName(t, e, "PPN Input")

	return f
}

func defaultOrganizationID(t *testing.T, accessToken string) uint64 {
	t.Helper()

	orgs := newExpect(t).GET("/api/v1/me/organizations").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("organizations").
		Array()

	return uint64(orgs.Element(0).Object().Value("id").Number().Raw())
}

func (f flowContext) authed(t *testing.T, method, path string) *httpexpect.Request {
	t.Helper()

	return newExpect(t).Request(method, f.orgPath+path).
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken)
}

func (f flowContext) createContact(t *testing.T, e *httpexpect.Expect) uint64 {
	t.Helper()

	contact := e.POST(f.orgPath+"/contacts").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"name":            gofakeit.Company() + " " + gofakeit.Word(),
			"email":           gofakeit.Email(),
			"is_organization": true,
			"customer":        map[string]any{"active": true},
			"supplier":        map[string]any{"active": true},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("contact").
		Object()

	return uint64(contact.Value("id").Number().Raw())
}

func (f flowContext) createSellableProduct(t *testing.T, e *httpexpect.Expect) uint64 {
	t.Helper()

	categoryID := uint64(0)
	categories := e.GET(f.orgPath+"/item-categories").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("categories").
		Array().
		Iter()

	for _, category := range categories {
		if category.Object().Value("name").String().Raw() == "Consumer Goods" {
			categoryID = uint64(category.Object().Value("id").Number().Raw())
		}
	}

	sku := "SKU-" + gofakeit.UUID()

	item := e.POST(f.orgPath+"/products").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"name":           gofakeit.ProductName(),
			"type":           "stockable",
			"tracking":       "none",
			"is_sellable":    true,
			"is_purchasable": true,
			"list_price":     1000000,
			"standard_cost":  600000,
			"category_id":    categoryID,
			"variants":       []map[string]any{{"sku": sku}},
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object()

	variants := item.Value("variants").Array()
	return uint64(variants.Element(0).Object().Value("id").Number().Raw())
}

func (f flowContext) createPriceBook(t *testing.T, e *httpexpect.Expect) uint64 {
	t.Helper()

	price_book := e.POST(f.orgPath+"/price_books").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{"name": "E2E " + gofakeit.Word()}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("price_book").
		Object()

	return uint64(price_book.Value("id").Number().Raw())
}

func (f flowContext) createPriceRule(t *testing.T, e *httpexpect.Expect) {
	t.Helper()

	e.POST(f.orgPath+"/price_books/"+itoa(f.price_bookID)+"/rules").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"applies_to":   "variant",
			"item_id":      f.variantID,
			"compute_type": "fixed",
			"fixed_price":  1000000,
		}).
		Expect().
		Status(http.StatusCreated)
}

func (f flowContext) firstWarehouseID(t *testing.T, e *httpexpect.Expect) uint64 {
	t.Helper()

	warehouses := e.GET(f.orgPath+"/warehouses").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("warehouses").
		Array()

	return uint64(warehouses.Element(0).Object().Value("id").Number().Raw())
}

func (f flowContext) createWarehouse(t *testing.T, e *httpexpect.Expect) uint64 {
	t.Helper()

	warehouse := e.POST(f.orgPath+"/warehouses").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"name": "E2E " + gofakeit.Word(),
			"code": "WH-" + strings.ToUpper(gofakeit.Word()),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("warehouse").
		Object()

	return uint64(warehouse.Value("id").Number().Raw())
}

func (f flowContext) ensureStockLocation(t *testing.T, e *httpexpect.Expect, code, usage string) uint64 {
	t.Helper()

	location := e.POST(f.orgPath+"/stock-locations").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"name":         code,
			"code":         code,
			"usage":        usage,
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

	return uint64(location.Value("id").Number().Raw())
}

func (f flowContext) journalByCode(t *testing.T, e *httpexpect.Expect, code string) uint64 {
	t.Helper()

	journalID := uint64(0)
	journals := e.GET(f.orgPath+"/journals").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("journals").
		Array().
		Iter()

	for _, journal := range journals {
		if journal.Object().Value("code").String().Raw() == code {
			journalID = uint64(journal.Object().Value("id").Number().Raw())
		}
	}

	return journalID
}

func (f flowContext) taxByName(t *testing.T, e *httpexpect.Expect, name string) uint64 {
	t.Helper()

	taxID := uint64(0)
	taxes := e.GET(f.orgPath+"/taxes").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("taxes").
		Array().
		Iter()

	for _, tax := range taxes {
		if strings.Contains(tax.Object().Value("name").String().Raw(), name) {
			taxID = uint64(tax.Object().Value("id").Number().Raw())
		}
	}

	return taxID
}

func (f flowContext) buildStock(t *testing.T, e *httpexpect.Expect, qty, unitCost int) {
	t.Helper()

	movement := e.POST(f.orgPath+"/stock-movements").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"item_id":         f.variantID,
			"qty":             itoa(uint64(qty)),
			"src_location_id": f.supplierLocationID,
			"dst_location_id": f.stockLocationID,
			"unit_cost":       unitCost,
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

	e.POST(f.orgPath+"/stock-movements/"+itoa(moveID)+"/receive").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"unit_cost":  itoa(uint64(unitCost)),
			"journal_id": f.journalGeneralID,
		}).
		Expect().
		Status(http.StatusOK)

	e.POST(f.orgPath+"/stock/rebuild").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK)
}

func (f flowContext) createWonOpportunity(t *testing.T, e *httpexpect.Expect) uint64 {
	t.Helper()

	lead := e.POST(f.orgPath+"/crm/leads").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"name":             gofakeit.Company() + " " + gofakeit.Word(),
			"email":            gofakeit.Email(),
			"source":           "website",
			"expected_revenue": 11100000,
			"probability":      60,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("lead").
		Object()

	prospectID := uint64(lead.Value("id").Number().Raw())

	stageID := f.firstOpportunityStageID(t, e)

	opportunity := e.POST(f.orgPath+"/crm/leads/"+itoa(prospectID)+"/promote").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		WithJSON(map[string]any{
			"stage_id":         stageID,
			"expected_revenue": 11100000,
			"probability":      60,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("lead").
		Object()

	opportunityID := uint64(opportunity.Value("id").Number().Raw())

	won := e.POST(f.orgPath+"/crm/opportunities/"+itoa(opportunityID)+"/win").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("opportunity").
		Object()

	if !won.Value("is_won").Boolean().Raw() {
		t.Fatalf("expected opportunity %d to be won", opportunityID)
	}

	return opportunityID
}

func (f flowContext) firstOpportunityStageID(t *testing.T, e *httpexpect.Expect) uint64 {
	t.Helper()

	stages := e.GET(f.orgPath+"/crm/stages").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("stages").
		Array()

	return uint64(stages.Element(0).Object().Value("id").Number().Raw())
}

func (f flowContext) otherStageID(t *testing.T, e *httpexpect.Expect, excludeID uint64) uint64 {
	t.Helper()

	stages := e.GET(f.orgPath+"/crm/stages").
		WithHeader("Authorization", "Bearer "+f.tokens.AccessToken).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		Value("data").
		Object().
		Value("stages").
		Array().
		Iter()

	for _, stage := range stages {
		stageID := uint64(stage.Object().Value("id").Number().Raw())
		if stageID != excludeID && !stage.Object().Value("is_won").Boolean().Raw() {
			return stageID
		}
	}

	return 0
}

func itoa(v uint64) string {
	return strconv.FormatUint(v, 10)
}
