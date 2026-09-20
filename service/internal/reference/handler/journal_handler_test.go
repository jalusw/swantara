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

var journalColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "name", "code", "type", "default_account_id", "currency_code", "bank_account_id", "sequence_id"}

func journalHandlerTest(t *testing.T, journals dao.Base[reference.Journal], svc reference.JournalService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewJournalHandler(svc)
		h.Register(api, guards)
	})
}

func journalHandlerTestNoTenant(t *testing.T, journals dao.Base[reference.Journal], svc reference.JournalService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, false, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewJournalHandler(svc)
		h.Register(api, guards)
	})
}

func journalTestSvc(journals dao.Base[reference.Journal], accounts dao.Base[reference.Account]) reference.JournalService {
	return reference.NewJournalService(journals, accounts)
}

func TestJournalHandler_List_ReturnsJournals(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journals" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTestNoTenant(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journals" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "journals" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Get_ReturnsJournal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Get_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "Sales", "S", "sale", nil, nil, nil, nil))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodGet, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_CreatesJournal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journals"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"","type":""}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTestNoTenant(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_MapsInvalidType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"bogus"}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_MapsDefaultAccountNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"sale","default_account_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_MapsForeignDefaultAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(5, timeNow(), timeNow(), nil, uint64(99), "A500", "Foreign", "asset", false, nil, nil, true))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"sale","default_account_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_CreatesWithDefaultAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns).AddRow(5, timeNow(), timeNow(), nil, uint64(10), "A500", "Revenue", "income", false, nil, nil, true))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journals"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"sale","default_account_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_ReturnsServerErrorOnAccountFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnError(errors.New("db down"))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"sale","default_account_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Create_ReturnsServerErrorOnInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "journals"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales","type":"sale"}`
	resp, err := doRequest(app, http.MethodPost, "/journals/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_UpdatesJournal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "journals" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"sale"}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"sale"}`
	resp, err := doRequest(app, http.MethodPut, "/journals/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"","type":""}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"sale"}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "Sales", "S", "sale", nil, nil, nil, nil))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"sale"}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_MapsInvalidType(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"bogus"}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_MapsDefaultAccountNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "accounts" WHERE id = $1 ORDER BY "accounts"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(accountColumns))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"sale","default_account_id":5}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"sale"}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "journals" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	body := `{"name":"Sales Book","type":"sale"}`
	resp, err := doRequest(app, http.MethodPut, "/journals/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Delete_DeletesJournal(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "journals" WHERE "journals"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/journals/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "journals" WHERE id = $1 ORDER BY "journals"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(journalColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Sales", "S", "sale", nil, nil, nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "journals" WHERE "journals"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := journalHandlerTest(t, dao.NewBase[reference.Journal](db), journalTestSvc(dao.NewBase[reference.Journal](db), dao.NewBase[reference.Account](db)))

	resp, err := doRequest(app, http.MethodDelete, "/journals/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestJournalHandler_WriteJournalError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid type", err: reference.ErrInvalidJournalType, want: http.StatusUnprocessableEntity},
		{name: "default account", err: reference.ErrJournalDefaultAccount, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeJournalError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
