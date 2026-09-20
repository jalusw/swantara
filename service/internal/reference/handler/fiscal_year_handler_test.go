package handler

import (
	"errors"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

var taxYearColumns = []string{"id", "created_at", "updated_at", "deleted_at", "organization_id", "name", "date_start", "date_end", "state"}

var (
	taxYearStart = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	taxYearEnd   = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
)

func taxYearHandlerTest(t *testing.T, years dao.Base[reference.TaxYear], svc reference.TaxYearService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewTaxYearHandler(svc)
		h.Register(api, guards)
	})
}

func taxYearHandlerTestNoTenant(t *testing.T, years dao.Base[reference.TaxYear], svc reference.TaxYearService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, false, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewTaxYearHandler(svc)
		h.Register(api, guards)
	})
}

func taxYearTestSvc(years dao.Base[reference.TaxYear]) reference.TaxYearService {
	return reference.NewTaxYearService(years)
}

func TestTaxYearHandler_List_ReturnsYears(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_years" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_List_ReturnsUnauthorizedWhenTenantMissing(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTestNoTenant(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_years" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnError(errors.New("db down"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "tax_years" WHERE organization_id = $1`)).
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE organization_id = $1 LIMIT $2`)).
		WithArgs(10, 20).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Get_ReturnsYear(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Get_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodGet, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Create_CreatesYear(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_years"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPost, "/tax-years/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"","date_start":"","date_end":""}`
	resp, err := doRequest(app, http.MethodPost, "/tax-years/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTestNoTenant(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPost, "/tax-years/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Create_RejectsInvalidDate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026","date_start":"not-a-date","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPost, "/tax-years/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Create_MapsInvalidDateRange(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026","date_start":"2026-12-31","date_end":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPost, "/tax-years/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Create_ReturnsServerErrorOnInsert(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "tax_years"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPost, "/tax-years/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_UpdatesYear(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_years" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-01-01","date_end":"2026-12-31","state":"open"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"","date_start":"","date_end":""}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_RejectsInvalidDate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"not-a-date","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_RejectsForeignOrganization(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(99), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_MapsInvalidDateRange(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-12-31","date_end":"2026-01-01"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Create_RejectsInvalidEndDate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026","date_start":"2026-01-01","date_end":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/tax-years/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_RejectsInvalidEndDate(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-01-01","date_end":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "tax_years" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	body := `{"name":"FY2026v2","date_start":"2026-01-01","date_end":"2026-12-31"}`
	resp, err := doRequest(app, http.MethodPut, "/tax-years/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Delete_DeletesYear(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "tax_years" WHERE "tax_years"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodDelete, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodDelete, "/tax-years/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodDelete, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodDelete, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "tax_years" WHERE id = $1 ORDER BY "tax_years"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(taxYearColumns).AddRow(1, timeNow(), timeNow(), nil, uint64(10), "FY2026", taxYearStart, taxYearEnd, "open"))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "tax_years" WHERE "tax_years"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := taxYearHandlerTest(t, dao.NewBase[reference.TaxYear](db), taxYearTestSvc(dao.NewBase[reference.TaxYear](db)))

	resp, err := doRequest(app, http.MethodDelete, "/tax-years/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestTaxYearHandler_WriteTaxYearError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "invalid date range", err: reference.ErrInvalidTaxYear, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeTaxYearError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
