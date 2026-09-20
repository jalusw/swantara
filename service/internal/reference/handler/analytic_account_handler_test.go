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

var dimensionColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "name", "code", "kind", "parent_id", "active"}

func dimensionHandlerTest(t *testing.T, accounts dao.Base[reference.Dimension], svc reference.DimensionService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewDimensionHandler(svc)
		h.Register(api, guards)
	})
}

func dimensionTestSvc(accounts dao.Base[reference.Dimension]) reference.DimensionService {
	return reference.NewDimensionService(accounts)
}

func TestDimensionHandler_List_ReturnsAccounts(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dimensions" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := referenceTestApp(t, false, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewDimensionHandler(dimensionTestSvc(dao.NewBase[reference.Dimension](db)))
		h.Register(api, guards)
	})

	resp, err := doRequest(app, http.MethodGet, "/dimensions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dimensions" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "dimensions" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Get_ReturnsAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Get_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "Marketing", "MKT", "cost", nil, true))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodGet, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Create_CreatesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dimensions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing"}`
	resp, err := doRequest(app, http.MethodPost, "/dimensions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/dimensions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Create_MapsDuplicateCode(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("MKT", 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(2, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing","code":"MKT"}`
	resp, err := doRequest(app, http.MethodPost, "/dimensions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Create_AllowsSameCodeInOtherOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("MKT", 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(2, timeNow(), timeNow(), nil, uint64(99), "Other", "MKT", "cost", nil, true))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dimensions"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing","code":"MKT"}`
	resp, err := doRequest(app, http.MethodPost, "/dimensions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Create_MapsParentNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing","parent_id":5}`
	resp, err := doRequest(app, http.MethodPost, "/dimensions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Create_ReturnsServerErrorOnSearch(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("MKT", 1).
		WillReturnError(errors.New("db down"))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing","code":"MKT"}`
	resp, err := doRequest(app, http.MethodPost, "/dimensions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Create_ReturnsServerErrorOnInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "dimensions"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing"}`
	resp, err := doRequest(app, http.MethodPost, "/dimensions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_UpdatesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "dimensions" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops"}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops"}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops"}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "Marketing", "MKT", "cost", nil, true))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops"}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_MapsDuplicateCode(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE code = $1 LIMIT $2`)).
		WithArgs("OPS", 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(2, timeNow(), timeNow(), nil, uint64(10), "Operations", "OPS", "cost", nil, true))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops","code":"OPS"}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_MapsParentNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops","parent_id":5}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops"}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "dimensions" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	body := `{"name":"Marketing Ops"}`
	resp, err := doRequest(app, http.MethodPut, "/dimensions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Delete_DeletesAccount(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "dimensions" WHERE "dimensions"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodDelete, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodDelete, "/dimensions/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodDelete, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodDelete, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "dimensions" WHERE id = $1 ORDER BY "dimensions"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(dimensionColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "Marketing", "MKT", "cost", nil, true))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "dimensions" WHERE "dimensions"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := dimensionHandlerTest(t, dao.NewBase[reference.Dimension](db), dimensionTestSvc(dao.NewBase[reference.Dimension](db)))

	resp, err := doRequest(app, http.MethodDelete, "/dimensions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestDimensionHandler_WriteDimensionError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "duplicate code", err: reference.ErrDuplicateCode, want: http.StatusUnprocessableEntity},
		{name: "parent not found", err: reference.ErrParentNotFound, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeDimensionError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
