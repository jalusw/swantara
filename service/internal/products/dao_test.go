package products

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"gorm.io/gorm"
)

func TestItemDAO_CreateWithVariants_CreatesTemplateAndVariants(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_variants"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_variants"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))
	mock.ExpectCommit()

	templates := NewItemDAO(db)
	template := &Item{Name: "T-Shirt"}
	variants := []*ItemVariant{{Sku: ptr("TS-RED")}, {Sku: ptr("TS-BLUE")}}

	created, err := templates.CreateWithVariants(ctx, template, variants)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != 1 {
		t.Errorf("template id = %d, want 1", created.ID)
	}
	if variants[0].ItemID != 1 || variants[1].ItemID != 1 {
		t.Errorf("variants = %+v, want item_id 1 on all", variants)
	}

	query.AssertDBMockDone(t, mock)
}

func TestItemDAO_CreateWithVariants_RollsBackOnTemplateError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	templates := NewItemDAO(db)
	_, err := templates.CreateWithVariants(ctx, &Item{Name: "T-Shirt"}, []*ItemVariant{{}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestItemDAO_CreateWithVariants_RollsBackOnVariantError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "items"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_variants"`)).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	templates := NewItemDAO(db)
	_, err := templates.CreateWithVariants(ctx, &Item{Name: "T-Shirt"}, []*ItemVariant{{}})

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestItemVariantDAO_ListByTemplate_ReturnsVariants(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "item_variants" WHERE item_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_variants" WHERE item_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "sku"}).
			AddRow(1, 7, "TS-RED").
			AddRow(2, 7, "TS-BLUE"))

	variants := NewItemVariantDAO(db)

	items, err := variants.ListByTemplate(ctx, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 || items[0].Sku == nil || *items[0].Sku != "TS-RED" {
		t.Errorf("items = %+v, want two variants", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestItemVariantDAO_ListByTemplate_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "item_variants"`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	variants := NewItemVariantDAO(db)
	_, err := variants.ListByTemplate(ctx, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestItemVariantDAO_VariantExists_TrueWhenFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_variants" WHERE id = $1`)).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id"}).AddRow(7, 1))

	variants := NewItemVariantDAO(db)

	found, err := variants.VariantExists(ctx, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Error("found = false, want true")
	}

	query.AssertDBMockDone(t, mock)
}

func TestItemVariantDAO_VariantExists_FalseWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_variants" WHERE id = $1`)).
		WithArgs(99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	variants := NewItemVariantDAO(db)

	found, err := variants.VariantExists(ctx, 99)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("found = true, want false")
	}

	query.AssertDBMockDone(t, mock)
}

func TestItemVariantDAO_VariantExists_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_variants" WHERE id = $1`)).
		WithArgs(7, 1).
		WillReturnError(errors.New("db down"))

	variants := NewItemVariantDAO(db)
	_, err := variants.VariantExists(ctx, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestPriceRuleDAO_ListByPriceBook_ReturnsRules(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "price_rules" WHERE price_book_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "price_rules" WHERE price_book_id = $1`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "price_book_id", "applies_to", "compute_type"}).
			AddRow(1, 3, AppliesToAll, ComputePercent))

	rules := NewPriceRuleDAO(db)

	items, err := rules.ListByPriceBook(ctx, 3)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].AppliesTo != AppliesToAll {
		t.Errorf("items = %+v, want one rule", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestPriceRuleDAO_ListByPriceBook_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "price_rules"`)).
		WithArgs(3).
		WillReturnError(errors.New("db down"))

	rules := NewPriceRuleDAO(db)
	_, err := rules.ListByPriceBook(ctx, 3)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListByItem_ReturnsOffers(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products" WHERE item_id = $1`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_products" WHERE item_id = $1`)).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "supplier_id", "price"}).
			AddRow(1, 5, 10, 80.0))

	offers := NewSupplierProductDAO(db)

	items, err := offers.ListByItem(ctx, 5)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].SupplierID != 10 {
		t.Errorf("items = %+v, want one offer for supplier 10", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListByItem_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products"`)).
		WithArgs(5).
		WillReturnError(errors.New("db down"))

	offers := NewSupplierProductDAO(db)
	_, err := offers.ListByItem(ctx, 5)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListBySupplier_ReturnsOffers(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products" WHERE supplier_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "supplier_products" WHERE supplier_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "supplier_id", "price"}).
			AddRow(1, 5, 10, 80.0))

	offers := NewSupplierProductDAO(db)

	items, err := offers.ListBySupplier(ctx, 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ItemID != 5 {
		t.Errorf("items = %+v, want one offer for item 5", items)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListBySupplier_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products"`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))

	offers := NewSupplierProductDAO(db)
	_, err := offers.ListBySupplier(ctx, 10)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListInOrg_ReturnsPage(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products" JOIN item_variants ON item_variants.id = supplier_products.item_id JOIN items ON items.id = item_variants.item_id AND items.organization_id = $1 WHERE supplier_id = $2`)).
		WithArgs(7, 10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "supplier_products"."id"`)).
		WithArgs(7, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "supplier_id", "price"}).
			AddRow(1, 5, 10, 80.0))

	offers := NewSupplierProductDAO(db)

	page, err := offers.ListInOrg(ctx, &query.Query{Filters: []query.Filter{{Field: "supplier_id", Operator: query.Equal, Value: 10}}}, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Count != 1 || len(page.Items) != 1 || page.Items[0].SupplierID != 10 {
		t.Errorf("page = %+v, want one offer for supplier 10", page)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListInOrg_WithoutQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products" JOIN item_variants ON item_variants.id = supplier_products.item_id JOIN items ON items.id = item_variants.item_id AND items.organization_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "supplier_products"."id"`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "supplier_id"}))

	offers := NewSupplierProductDAO(db)

	page, err := offers.ListInOrg(ctx, nil, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page.Items) != 0 {
		t.Errorf("page = %+v, want no offers", page)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListInOrg_PropagatesCountError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products" JOIN item_variants ON item_variants.id = supplier_products.item_id JOIN items ON items.id = item_variants.item_id AND items.organization_id = $1`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	offers := NewSupplierProductDAO(db)
	_, err := offers.ListInOrg(ctx, nil, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_ListInOrg_PropagatesFindError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "supplier_products" JOIN item_variants ON item_variants.id = supplier_products.item_id JOIN items ON items.id = item_variants.item_id AND items.organization_id = $1`)).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "supplier_products"."id"`)).
		WithArgs(7).
		WillReturnError(errors.New("db down"))

	offers := NewSupplierProductDAO(db)
	_, err := offers.ListInOrg(ctx, nil, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_FindInOrg_FindsOffer(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "supplier_products"."id"`)).
		WithArgs(7, 5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "item_id", "supplier_id"}).AddRow(5, 100, 10))

	offers := NewSupplierProductDAO(db)

	offer, err := offers.FindInOrg(ctx, 5, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offer == nil || offer.SupplierID != 10 {
		t.Errorf("offer = %+v, want supplier 10", offer)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_FindInOrg_ReturnsNilWhenMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "supplier_products"."id"`)).
		WithArgs(7, 99, 1).
		WillReturnError(gorm.ErrRecordNotFound)

	offers := NewSupplierProductDAO(db)

	offer, err := offers.FindInOrg(ctx, 99, 7)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if offer != nil {
		t.Errorf("offer = %+v, want nil", offer)
	}

	query.AssertDBMockDone(t, mock)
}

func TestSupplierProductDAO_FindInOrg_PropagatesError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "supplier_products"."id"`)).
		WithArgs(7, 5, 1).
		WillReturnError(errors.New("db down"))

	offers := NewSupplierProductDAO(db)
	_, err := offers.FindInOrg(ctx, 5, 7)

	if helper.AssertError(t, err, true, nil) {
		return
	}

	query.AssertDBMockDone(t, mock)
}

func TestNewPriceBookDAO(t *testing.T) {
	db, mock := query.NewMockDB(t)
	ctx := context.Background()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "price_books" WHERE id = $1`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Retail"))

	price_books := NewPriceBookDAO(db)

	found, err := price_books.Find(ctx, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil || found.Name != "Retail" {
		t.Errorf("price_book = %+v, want Retail", found)
	}

	query.AssertDBMockDone(t, mock)
}
