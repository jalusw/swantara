package handler

import (
	"context"
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

var carrierColumns = []string{"id", "created_at", "updated_at", "deleted_at", "name", "tracking_url_tpl", "delivery_item_id"}

func carrierHandlerTest(t *testing.T, carriers dao.Base[reference.Carrier], svc reference.CarrierService) *fiber.App {
	t.Helper()
	return referenceTestApp(t, true, func(api fiber.Router, guards httpx.RouteGuards) {
		h := NewCarrierHandler(svc)
		h.Register(api, guards)
	})
}

func carrierTestSvc(carriers dao.Base[reference.Carrier], variants reference.VariantExistenceChecker) reference.CarrierService {
	return reference.NewCarrierService(carriers, variants)
}

func TestCarrierHandler_List_ReturnsCarriers(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "carriers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(carrierColumns).
			AddRow(1, timeNow(), timeNow(), nil, "DHL", nil, nil))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_List_RejectsInvalidQuery(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_List_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "carriers"`)).
		WillReturnError(errors.New("db down"))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_List_ExportsCSV(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "carriers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" LIMIT $1`)).
		WithArgs(20).
		WillReturnRows(sqlmock.NewRows(carrierColumns).
			AddRow(1, timeNow(), timeNow(), nil, "DHL", nil, nil))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/?format=csv", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Get_ReturnsCarrier(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns).
			AddRow(1, timeNow(), timeNow(), nil, "DHL", nil, nil))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Get_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Get_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Get_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodGet, "/carriers/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Create_CreatesCarrier(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "carriers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"DHL"}`
	resp, err := doRequest(app, http.MethodPost, "/carriers/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Create_CreatesWithDeliveryProduct(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "carriers"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()
	variants := reference.VariantExistenceMock{
		VariantExistsFunc: func(_ context.Context, _ uint64) (bool, error) {
			return true, nil
		},
	}
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), variants))

	body := `{"name":"DHL","delivery_item_id":7}`
	resp, err := doRequest(app, http.MethodPost, "/carriers/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Create_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/carriers/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Create_MapsBlankName(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"   "}`
	resp, err := doRequest(app, http.MethodPost, "/carriers/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Create_MapsUnknownDeliveryProduct(t *testing.T) {
	db, mock := query.NewMockDB(t)
	variants := reference.VariantExistenceMock{
		VariantExistsFunc: func(_ context.Context, _ uint64) (bool, error) {
			return false, nil
		},
	}
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), variants))

	body := `{"name":"DHL","delivery_item_id":99}`
	resp, err := doRequest(app, http.MethodPost, "/carriers/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Create_ReturnsServerError(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "carriers"`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"DHL"}`
	resp, err := doRequest(app, http.MethodPost, "/carriers/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Update_UpdatesCarrier(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns).
			AddRow(1, timeNow(), timeNow(), nil, "DHL", nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "carriers" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"DHL Express"}`
	resp, err := doRequest(app, http.MethodPut, "/carriers/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Update_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"DHL Express"}`
	resp, err := doRequest(app, http.MethodPut, "/carriers/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Update_RejectsValidation(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPut, "/carriers/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Update_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"DHL Express"}`
	resp, err := doRequest(app, http.MethodPut, "/carriers/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"DHL Express"}`
	resp, err := doRequest(app, http.MethodPut, "/carriers/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns).
			AddRow(1, timeNow(), timeNow(), nil, "DHL", nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "carriers" SET`)).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	body := `{"name":"DHL Express"}`
	resp, err := doRequest(app, http.MethodPut, "/carriers/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Delete_DeletesCarrier(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns).
			AddRow(1, timeNow(), timeNow(), nil, "DHL", nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "carriers" WHERE "carriers"."id" = $1`)).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodDelete, "/carriers/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Delete_RejectsInvalidID(t *testing.T) {
	db, mock := query.NewMockDB(t)
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodDelete, "/carriers/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Delete_ReturnsNotFound(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodDelete, "/carriers/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnError(errors.New("db down"))
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodDelete, "/carriers/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	db, mock := query.NewMockDB(t)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "carriers" WHERE id = $1 ORDER BY "carriers"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows(carrierColumns).
			AddRow(1, timeNow(), timeNow(), nil, "DHL", nil, nil))
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "carriers" WHERE "carriers"."id" = $1`)).
		WithArgs(1).
		WillReturnError(errors.New("db down"))
	mock.ExpectRollback()
	app := carrierHandlerTest(t, dao.NewBase[reference.Carrier](db), carrierTestSvc(dao.NewBase[reference.Carrier](db), nil))

	resp, err := doRequest(app, http.MethodDelete, "/carriers/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
	query.AssertDBMockDone(t, mock)
}

func TestCarrierHandler_WriteCarrierError_MapErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "blank name", err: reference.ErrCarrierName, want: http.StatusUnprocessableEntity},
		{name: "unknown delivery item", err: reference.ErrCarrierDeliveryProduct, want: http.StatusUnprocessableEntity},
		{name: "unexpected", err: errors.New("boom"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeCarrierError(c, tt.err)
			})
			if got != tt.want {
				t.Errorf("status = %d, want %d", got, tt.want)
			}
		})
	}
}
