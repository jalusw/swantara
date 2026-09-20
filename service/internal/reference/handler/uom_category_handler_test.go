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

var unitGroupColumns = []string{"id", "created_at", "updated_at", "deleted_at", "name"}

func unitGroupHandlerTest(t *testing.T, categories dao.Base[reference.UnitGroup]) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewUnitGroupHandler(reference.NewUnitService(categories, dao.NewBase[reference.Unit](nil)))
		h.Register(api, guards)
	})
}

func TestUnitGroupHandler_List_ReturnsCategories(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "unit_groups"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns).
			AddRow(1, timeNow(), timeNow(), nil, "Weight"))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "unit_groups"`)).
		WillReturnError(errors.New("db down"))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "unit_groups"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns).
			AddRow(1, timeNow(), timeNow(), nil, "Weight"))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Get_ReturnsCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns).
			AddRow(1, timeNow(), timeNow(), nil, "Weight"))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodGet, "/unit-groups/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Create_CreatesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "unit_groups"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":"Weight"}`
	resp, err := doRequest(app, http.MethodPost, "/unit-groups/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/unit-groups/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Create_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "unit_groups"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":"Weight"}`
	resp, err := doRequest(app, http.MethodPost, "/unit-groups/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Update_UpdatesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns).
			AddRow(1, timeNow(), timeNow(), nil, "Weight"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "unit_groups" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":"Volume"}`
	resp, err := doRequest(app, http.MethodPut, "/unit-groups/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":"Volume"}`
	resp, err := doRequest(app, http.MethodPut, "/unit-groups/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{}`
	resp, err := doRequest(app, http.MethodPut, "/unit-groups/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":"Volume"}`
	resp, err := doRequest(app, http.MethodPut, "/unit-groups/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":"Volume"}`
	resp, err := doRequest(app, http.MethodPut, "/unit-groups/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns).
			AddRow(1, timeNow(), timeNow(), nil, "Weight"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "unit_groups" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	body := `{"name":"Volume"}`
	resp, err := doRequest(app, http.MethodPut, "/unit-groups/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Delete_DeletesCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns).
			AddRow(1, timeNow(), timeNow(), nil, "Weight"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "unit_groups" WHERE "unit_groups"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodDelete, "/unit-groups/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodDelete, "/unit-groups/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodDelete, "/unit-groups/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodDelete, "/unit-groups/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitGroupHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(unitGroupColumns).
			AddRow(1, timeNow(), timeNow(), nil, "Weight"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "unit_groups" WHERE "unit_groups"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := unitGroupHandlerTest(t, dao.NewBase[reference.UnitGroup](db))

	resp, err := doRequest(app, http.MethodDelete, "/unit-groups/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}
