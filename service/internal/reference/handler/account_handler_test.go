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

var accountColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "code", "name", "type", "reconcilable", "currency_code", "parent_id", "active"}

func accountHandlerTest(t *testing.T, accounts dao.Base[reference.Account], svc reference.AccountService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewAccountHandler(svc)
		h.Register(api, guards)
	})
}

func accountHandlerTestNoTenant(t *testing.T, accounts dao.Base[reference.Account], svc reference.AccountService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, false, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewAccountHandler(svc)
		h.Register(api, guards)
	})
}

func accountTestSvc(accounts dao.Base[reference.Account]) reference.AccountService {
	return reference.NewAccountService(accounts)
}

func TestAccountHandler_List_ReturnsAccounts(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "accounts" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTestNoTenant(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "accounts" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "accounts" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Get_ReturnsAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Get_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "A100", "Cash", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_CreatesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A100", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A100","name":"Cash","type":"asset"}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"","name":"","type":""}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTestNoTenant(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A100","name":"Cash","type":"asset"}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_MapsInvalidType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A100","name":"Cash","type":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_MapsDuplicateCode(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A100", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(2, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A100","name":"Cash","type":"asset"}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_MapsInvalidParent(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A110", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A110","name":"Sub","type":"asset","parent_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_ReturnsServerErrorOnSearch(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A100", 1).
		WillReturnError(errors.New("db down"))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A100","name":"Cash","type":"asset"}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_ReturnsServerErrorOnInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A100", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A100","name":"Cash","type":"asset"}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_UpdatesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A101", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"","name":"","type":""}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "A100", "Cash", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_MapsInvalidType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"bogus"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_MapsDuplicateCode(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A101", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(2, timeNow(), timeNow(), nil, uint64(10), "A101", "Other", "asset", false, nil, nil, true))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_MapsInvalidParent(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A101", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset","parent_id":5}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A101", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "accounts" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A101","name":"Cash Box","type":"asset"}`
	resp, err := doRequest(app, http.MethodPut, "/accounts/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Delete_DeletesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "accounts" WHERE "accounts"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/accounts/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Create_CreatesInactiveAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE code = $1 LIMIT $2`)).
		WithArgs("A100", 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "accounts"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	body := `{"code":"A100","name":"Cash","type":"asset","active":false}`
	resp, err := doRequest(app, http.MethodPost, "/accounts/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "A100", "Cash", "asset", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "accounts" WHERE "accounts"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := accountHandlerTest(t, dao.NewBase[reference.Account](db), accountTestSvc(dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/accounts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestAccountHandler_WriteAccountError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid type", err: reference.ErrInvalidAccountType, want: http.StatusUnprocessableEntity},
		{name: "duplicate code", err: reference.ErrDuplicateAccount, want: http.StatusUnprocessableEntity},
		{name: "invalid parent", err: reference.ErrInvalidParent, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeAccountError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
