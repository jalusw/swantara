package handler

import (
	"errors"
	"net/http"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

var taxColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "name", "amount", "type", "scope", "price_include", "tax_account_id", "refund_tax_account_id", "active"}

func taxHandlerTest(t *testing.T, taxes dao.Base[reference.Tax], svc reference.TaxService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewTaxHandler(svc)
		h.Register(api, guards)
	})
}

func taxHandlerTestNoTenant(t *testing.T, taxes dao.Base[reference.Tax], svc reference.TaxService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, false, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewTaxHandler(svc)
		h.Register(api, guards)
	})
}

func taxTestSvc(taxes dao.Base[reference.Tax], accounts dao.Base[reference.Account]) reference.TaxService {
	return reference.NewTaxService(taxes, accounts)
}

func TestTaxHandler_List_ReturnsTaxes(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "taxes" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTestNoTenant(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "taxes" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "taxes" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Get_ReturnsTax(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Get_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_CreatesTax(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "taxes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_CreatesGroupTax(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "taxes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Composite","type":"group","scope":"none"}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"","type":"","scope":""}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTestNoTenant(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_MapsInvalidType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"bogus","scope":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_MapsInvalidScope(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_MapsMissingAmount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","type":"percent","scope":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_MapsTaxAccountNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"sale","tax_account_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_MapsRefundTaxAccountNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(6, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(6, timeNow(), timeNow(), nil, uint64(10), "A600", "Tax", "tax", false, nil, nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"sale","tax_account_id":6,"refund_tax_account_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_CreatesWithAccounts(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(6, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(6, timeNow(), timeNow(), nil, uint64(10), "A600", "Tax", "tax", false, nil, nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(7, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(7, timeNow(), timeNow(), nil, uint64(10), "A700", "Refund", "tax", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "taxes"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"sale","tax_account_id":6,"refund_tax_account_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_ReturnsServerErrorOnAccountFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnError(errors.New("db down"))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"sale","tax_account_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Create_ReturnsServerErrorOnInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "taxes"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN","amount":10,"type":"percent","scope":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/taxes/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_UpdatesTax(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "taxes" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"","type":"","scope":""}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_MapsInvalidType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"bogus","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_MapsInvalidScope(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"bogus","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_MapsMissingAmount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","type":"percent","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_MapsTaxAccountNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"sale","active":true,"tax_account_id":5}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "taxes" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"PPN 12","amount":12,"type":"percent","scope":"sale","active":true}`
	resp, err := doRequest(app, http.MethodPut, "/taxes/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Delete_DeletesTax(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "taxes" WHERE "taxes"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/taxes/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "taxes" WHERE id = $1 ORDER BY "taxes"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "PPN", 10.0, "percent", "sale", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "taxes" WHERE "taxes"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := taxHandlerTest(t, dao.NewBase[reference.Tax](db), taxTestSvc(dao.NewBase[reference.Tax](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/taxes/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxHandler_WriteTaxError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid type", err: reference.ErrInvalidTaxType, want: http.StatusUnprocessableEntity},
		{name: "invalid scope", err: reference.ErrInvalidTaxScope, want: http.StatusUnprocessableEntity},
		{name: "missing amount", err: reference.ErrTaxAmountMissing, want: http.StatusUnprocessableEntity},
		{name: "invalid account", err: reference.ErrInvalidTaxAccount, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeTaxError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
