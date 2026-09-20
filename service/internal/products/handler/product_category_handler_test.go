package handler

import (
	"database/sql/driver"
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/products"
	"github.com/jalusw/swantara/apps/service/internal/reference"
	"gorm.io/gorm"
)

var itemCategoryColumns = []string{
	"id", "created_at", "updated_at", "deleted_at", "organization_id", "name", "parent_id",
	"income_account_id", "expense_account_id", "stock_cost_account_id", "stock_input_account_id",
	"stock_output_account_id", "cogs_account_id", "cost_method", "valuation",
}

func itemCategoryRow(category *reference.ItemCategory) []driver.Value {
	return []driver.Value{
		category.ID, category.CreatedAt, category.UpdatedAt, nil, category.OrganizationID, category.Name,
		category.ParentID, category.IncomeAccountID, category.ExpenseAccountID, category.StockValuationAccountID,
		category.StockInputAccountID, category.StockOutputAccountID, category.CogsAccountID, category.CostMethod,
		category.Valuation,
	}
}

func TestItemCategoryHandler_List_ReturnsCategories(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "item_categories" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_List_RequiresTenant(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := itemCategoryHandlerTestNoTenant(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "item_categories" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "item_categories" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Get_ReturnsCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Get_ReturnsNotFoundForForeignTenant(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(foreignItemCategory())...))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodGet, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Create_CreatesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_categories"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme","cost_method":"average","valuation":"real_time"}`
	resp, err := doRequest(app, http.MethodPost, "/item-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/item-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Create_RequiresOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := itemCategoryHandlerTestNoTenant(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme"}`
	resp, err := doRequest(app, http.MethodPost, "/item-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Create_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "item_categories"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme"}`
	resp, err := doRequest(app, http.MethodPost, "/item-categories/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Update_UpdatesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "item_categories" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/item-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/item-categories/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/item-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/item-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Update_ReturnsNotFoundForForeignTenant(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(foreignItemCategory())...))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/item-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/item-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "item_categories" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	body := `{"name":"Acme Corp"}`
	resp, err := doRequest(app, http.MethodPut, "/item-categories/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Delete_DeletesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "item_categories" WHERE "item_categories"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodDelete, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodDelete, "/item-categories/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(gorm.ErrRecordNotFound)
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodDelete, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodDelete, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestItemCategoryHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "item_categories" WHERE id = $1 ORDER BY "item_categories"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(itemCategoryColumns).AddRow(itemCategoryRow(sampleItemCategory())...))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "item_categories" WHERE "item_categories"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := itemCategoryHandlerTest(t, products.NewItemCategoryService(dao.NewBase[reference.ItemCategory](db)))

	resp, err := doRequest(app, http.MethodDelete, "/item-categories/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}
