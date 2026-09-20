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

var uomColumns = []string{"id", "created_at", "updated_at", "deleted_at", "category_id", "name", "factor", "unit_type", "rounding"}

func uomHandlerTest(t *testing.T, units dao.Base[reference.Unit], categories dao.Base[reference.UnitGroup], svc reference.UnitService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewUnitHandler(svc)
		h.Register(api, guards)
	})
}

func uomTestSvc(units dao.Base[reference.Unit], categories dao.Base[reference.UnitGroup]) reference.UnitService {
	return reference.NewUnitService(categories, units)
}

func TestUnitHandler_List_ReturnsUnits(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "units"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "units"`)).
		WillReturnError(errors.New("db down"))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "units"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Get_ReturnsUnit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodGet, "/units/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Create_CreatesUnit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("kg", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "units"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"category_id":1,"name":"kg","factor":1}`
	resp, err := doRequest(app, http.MethodPost, "/units/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"category_id":0,"name":"","factor":0}`
	resp, err := doRequest(app, http.MethodPost, "/units/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Create_MapsMissingCategory(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"category_id":1,"name":"kg","factor":1}`
	resp, err := doRequest(app, http.MethodPost, "/units/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Create_MapsDuplicateName(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("kg", 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(9, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"category_id":1,"name":"kg","factor":1}`
	resp, err := doRequest(app, http.MethodPost, "/units/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Create_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"category_id":1,"name":"kg","factor":1}`
	resp, err := doRequest(app, http.MethodPost, "/units/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Create_ReturnsServerErrorOnInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("kg", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "units"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"category_id":1,"name":"kg","factor":1}`
	resp, err := doRequest(app, http.MethodPost, "/units/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Convert_Converts(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(2, timeNow(), timeNow(), nil, 1, "g", 0.001, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"from_id":1,"to_id":2,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/units/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Convert_RejectsInvalidQty(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"from_id":1,"to_id":2,"qty":"abc"}`
	resp, err := doRequest(app, http.MethodPost, "/units/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Convert_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"from_id":0,"to_id":2,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/units/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Convert_MapsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"from_id":1,"to_id":2,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/units/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Convert_MapsCategoryMismatch(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(2, timeNow(), timeNow(), nil, 2, "h", 1.0, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"from_id":1,"to_id":2,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/units/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Convert_MapsInvalidFactor(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 0, "", 0.001))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(2, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(2, timeNow(), timeNow(), nil, 1, "g", 0.001, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"from_id":1,"to_id":2,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/units/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Convert_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"from_id":1,"to_id":2,"qty":"2"}`
	resp, err := doRequest(app, http.MethodPost, "/units/convert", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Update_UpdatesUnit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("kilogram", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "units" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"name":"kilogram","factor":1}`
	resp, err := doRequest(app, http.MethodPut, "/units/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"name":"kilogram","factor":1}`
	resp, err := doRequest(app, http.MethodPut, "/units/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"name":"","factor":0}`
	resp, err := doRequest(app, http.MethodPut, "/units/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"name":"kilogram","factor":1}`
	resp, err := doRequest(app, http.MethodPut, "/units/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Update_MapsDuplicateName(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("kilogram", 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(9, timeNow(), timeNow(), nil, 1, "kilogram", 1.0, "", 0.001))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"name":"kilogram","factor":1}`
	resp, err := doRequest(app, http.MethodPut, "/units/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"name":"kilogram","factor":1}`
	resp, err := doRequest(app, http.MethodPut, "/units/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "unit_groups" WHERE id = $1 ORDER BY "unit_groups"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE name = $1 LIMIT $2`)).
		WithArgs("kilogram", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "units" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	body := `{"name":"kilogram","factor":1}`
	resp, err := doRequest(app, http.MethodPut, "/units/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Delete_DeletesUnit(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "units" WHERE "units"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodDelete, "/units/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodDelete, "/units/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodDelete, "/units/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodDelete, "/units/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestUnitHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "units" WHERE id = $1 ORDER BY "units"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(uomColumns).
			AddRow(1, timeNow(), timeNow(), nil, 1, "kg", 1.0, "", 0.001))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "units" WHERE "units"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := uomHandlerTest(t, dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db), uomTestSvc(dao.NewBase[reference.Unit](db), dao.NewBase[reference.UnitGroup](db)))

	resp, err := doRequest(app, http.MethodDelete, "/units/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}
