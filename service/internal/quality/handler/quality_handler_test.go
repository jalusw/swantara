package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/quality"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func qualityHandlerTest(
	t *testing.T,
	points quality.QualityPointDAOMock,
	checks quality.QualityCheckDAOMock,
	alerts quality.QualityAlertDAOMock,
	svc quality.QualityCheckService,
) *fiber.App {
	t.Helper()
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals(httpx.LocalOrganizationID, uint64(1))
		c.Locals(model.ActorKey, uint64(5))
		return c.Next()
	})
	pointSvc := quality.NewQualityPointService(points)
	pointHandler := NewQualityPointHandler(pointSvc)
	pointHandler.Register(app, passthroughGuards())
	checkHandler := NewQualityCheckHandler(svc)
	checkHandler.Register(app, passthroughGuards())
	alertHandler := NewQualityAlertHandler(svc)
	alertHandler.Register(app, passthroughGuards())
	return app
}

func sampleQualityPoint() *reference.QualityPoint {
	orgID := uint64(1)
	itemID := uint64(5)
	operation := "weight check"
	return &reference.QualityPoint{
		Base:           model.Base{ID: 1},
		OrganizationID: &orgID,
		ItemID:         &itemID,
		Operation:      &operation,
		TestType:       quality.TestTypePassFail,
	}
}

func sampleQualityCheck(result string) *quality.QualityCheck {
	now := time.Now()
	pointID := uint64(1)
	itemID := uint64(5)
	shipmentID := uint64(9)
	checkedBy := uint64(5)
	return &quality.QualityCheck{
		Base:       model.Base{ID: 1},
		PointID:    &pointID,
		ItemID:     &itemID,
		ShipmentID: &shipmentID,
		Result:     result,
		CheckedBy:  &checkedBy,
		CheckedAt:  &now,
	}
}

func sampleQualityAlert(state string) *quality.QualityAlert {
	itemID := uint64(5)
	title := "Quality check failed"
	return &quality.QualityAlert{
		Base:     model.Base{ID: 1},
		ItemID:   &itemID,
		Title:    &title,
		Severity: helper.Ptr("warning"),
		State:    state,
	}
}

func TestQualityPointHandler_List_ReturnsPoints(t *testing.T) {
	points := quality.QualityPointDAOMock{
		CRUDMock: daoListFunc(sampleQualityPoint()),
	}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-points/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityPointHandler_List_RejectsInvalidQuery(t *testing.T) {
	points := quality.QualityPointDAOMock{}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-points/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityPointHandler_List_ReturnsServerError(t *testing.T) {
	points := quality.QualityPointDAOMock{
		CRUDMock: daoListErrFunc[reference.QualityPoint](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-points/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityPointHandler_Get_ReturnsPoint(t *testing.T) {
	points := quality.QualityPointDAOMock{
		CRUDMock: daoFindFunc(sampleQualityPoint()),
	}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-points/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityPointHandler_Get_ReturnsNotFound(t *testing.T) {
	points := quality.QualityPointDAOMock{
		CRUDMock: daoFindFunc[reference.QualityPoint](nil),
	}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-points/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestQualityPointHandler_Get_RejectsInvalidID(t *testing.T) {
	points := quality.QualityPointDAOMock{}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-points/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityPointHandler_Get_ReturnsServerError(t *testing.T) {
	points := quality.QualityPointDAOMock{
		CRUDMock: daoFindErrFunc[reference.QualityPoint](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-points/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityPointHandler_Create_CreatesPoint(t *testing.T) {
	created := sampleQualityPoint()
	points := quality.QualityPointDAOMock{
		CRUDMock: daoCreateFunc(created),
	}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	body := `{"item_id":5,"operation":"weight check","test_type":"pass_fail"}`
	resp, err := doRequest(app, http.MethodPost, "/quality-points/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestQualityPointHandler_Create_RejectsValidation(t *testing.T) {
	points := quality.QualityPointDAOMock{}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/quality-points/", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityPointHandler_Create_ReturnsServerError(t *testing.T) {
	points := quality.QualityPointDAOMock{
		CRUDMock: daoCreateErrFunc[reference.QualityPoint](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, quality.QualityCheckDAOMock{}, quality.QualityAlertDAOMock{}, svc)

	body := `{"item_id":5,"operation":"weight check","test_type":"pass_fail"}`
	resp, err := doRequest(app, http.MethodPost, "/quality-points/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityCheckHandler_List_ReturnsChecks(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoListFunc(sampleQualityCheck(quality.CheckResultPending)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityCheckHandler_List_RejectsMissingTenant(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := fiber.New()
	h := NewQualityCheckHandler(svc)
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_List_ReturnsServerError(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoListErrFunc[quality.QualityCheck](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityCheckHandler_Get_ReturnsCheck(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(sampleQualityCheck(quality.CheckResultPending)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityCheckHandler_Get_ReturnsNotFound(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc[quality.QualityCheck](nil),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestQualityCheckHandler_Get_RejectsInvalidID(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_Get_ReturnsServerError(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindErrFunc[quality.QualityCheck](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_RecordsPass(t *testing.T) {
	pending := sampleQualityCheck(quality.CheckResultPending)
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(pending),
	}
	points := quality.QualityPointDAOMock{
		CRUDMock: daoFindFunc(sampleQualityPoint()),
	}
	svc := quality.NewQualityCheckService(points, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_ReturnsNotFound(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc[quality.QualityCheck](nil),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_RejectsInvalidID(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/abc/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_RejectsMissingTenant(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := fiber.New()
	h := NewQualityCheckHandler(svc)
	h.Register(app, passthroughGuards())

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_ReturnsConflictForResolved(t *testing.T) {
	resolved := sampleQualityCheck(quality.CheckResultPass)
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(resolved),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_RequiresMeasuredValue(t *testing.T) {
	pending := sampleQualityCheck(quality.CheckResultPending)
	measurePoint := sampleQualityPoint()
	measurePoint.TestType = quality.TestTypeMeasure
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(pending),
	}
	points := quality.QualityPointDAOMock{
		CRUDMock: daoFindFunc(measurePoint),
	}
	svc := quality.NewQualityCheckService(points, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_ReturnsServerError(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindErrFunc[quality.QualityCheck](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_RejectsInvalidBody(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(sampleQualityCheck(quality.CheckResultPending)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", `{"pass":"yes"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_WithCheckedBy(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(sampleQualityCheck(quality.CheckResultPending)),
	}
	points := quality.QualityPointDAOMock{
		CRUDMock: daoFindFunc(sampleQualityPoint()),
	}
	svc := quality.NewQualityCheckService(points, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true,"checked_by":7}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_ReturnsPointNotFound(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(sampleQualityCheck(quality.CheckResultPending)),
	}
	points := quality.QualityPointDAOMock{
		CRUDMock: daoFindFunc[reference.QualityPoint](nil),
	}
	svc := quality.NewQualityCheckService(points, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, points, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RecordResult_ReturnsNoProduct(t *testing.T) {
	noProduct := sampleQualityCheck(quality.CheckResultPending)
	noProduct.PointID = nil
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoFindFunc(noProduct),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"pass":true}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/1/result", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_List_RejectsInvalidQuery(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_Get_RejectsMissingTenant(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := fiber.New()
	h := NewQualityCheckHandler(svc)
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/quality-checks/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RouteToScrap_MapsHandlerErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  int
		wantError bool
	}{
		{"check not found", quality.ErrQualityCheckNotFound, http.StatusNotFound, false},
		{"check no item", quality.ErrQualityCheckNoProduct, http.StatusUnprocessableEntity, false},
		{"point not found", quality.ErrQualityPointNotFound, http.StatusNotFound, false},
		{"alert not found", quality.ErrQualityAlertNotFound, http.StatusNotFound, false},
		{"scrap location", inventory.ErrScrapLocation, http.StatusNotFound, false},
		{"scrap account", inventory.ErrScrapAccount, http.StatusUnprocessableEntity, false},
		{"movement not found", inventory.ErrMovementNotFound, http.StatusNotFound, false},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checks := quality.QualityCheckDAOMock{
				ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*quality.QualityCheck, error) {
					return []*quality.QualityCheck{sampleQualityCheck(quality.CheckResultFail)}, nil
				},
			}
			svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
			svc.SetScrapRouter(qualityScrapRouterMock{
				route: func(_ context.Context, _, _, _, _, _ uint64, _ time.Time) error {
					return tt.err
				},
			})
			app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

			body := `{"journal_id":1,"expense_account_id":2}`
			resp, err := doRequest(app, http.MethodPost, "/quality-checks/shipments/9/scrap", body)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tt.wantCode {
				t.Fatalf("status = %d, want %d", resp.StatusCode, tt.wantCode)
			}
		})
	}
}

func TestQualityCheckHandler_RouteToScrap_RoutesFailed(t *testing.T) {
	failed := sampleQualityCheck(quality.CheckResultFail)
	checks := quality.QualityCheckDAOMock{
		CRUDMock: daoListFunc(failed),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	svc.SetScrapRouter(qualityScrapRouterMock{
		route: func(_ context.Context, _, _, _, _, _ uint64, _ time.Time) error {
			return nil
		},
	})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"journal_id":1,"expense_account_id":2}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/shipments/9/scrap", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RouteToScrap_RejectsInvalidShipmentID(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"journal_id":1,"expense_account_id":2}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/shipments/abc/scrap", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RouteToScrap_RejectsValidation(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	resp, err := doRequest(app, http.MethodPost, "/quality-checks/shipments/9/scrap", `{}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RouteToScrap_RejectsMissingTenant(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := fiber.New()
	h := NewQualityCheckHandler(svc)
	h.Register(app, passthroughGuards())

	body := `{"journal_id":1,"expense_account_id":2}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/shipments/9/scrap", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RouteToScrap_ReturnsUnconfigured(t *testing.T) {
	checks := quality.QualityCheckDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"journal_id":1,"expense_account_id":2}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/shipments/9/scrap", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityCheckHandler_RouteToScrap_MapsScrapErrors(t *testing.T) {
	checks := quality.QualityCheckDAOMock{
		ListByShipmentFunc: func(_ context.Context, _ uint64) ([]*quality.QualityCheck, error) {
			return []*quality.QualityCheck{sampleQualityCheck(quality.CheckResultFail)}, nil
		},
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{})
	svc.SetScrapRouter(qualityScrapRouterMock{
		route: func(_ context.Context, _, _, _, _, _ uint64, _ time.Time) error {
			return inventory.ErrInsufficientStock
		},
	})
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, checks, quality.QualityAlertDAOMock{}, svc)

	body := `{"journal_id":1,"expense_account_id":2}`
	resp, err := doRequest(app, http.MethodPost, "/quality-checks/shipments/9/scrap", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestQualityAlertHandler_List_ReturnsAlerts(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoListFunc(sampleQualityAlert(quality.AlertStateOpen)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityAlertHandler_List_RejectsMissingTenant(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := fiber.New()
	h := NewQualityAlertHandler(svc)
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityAlertHandler_List_ReturnsServerError(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoListErrFunc[quality.QualityAlert](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityAlertHandler_Get_ReturnsAlert(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindFunc(sampleQualityAlert(quality.AlertStateOpen)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityAlertHandler_Get_ReturnsNotFound(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindFunc[quality.QualityAlert](nil),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestQualityAlertHandler_Get_RejectsInvalidID(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityAlertHandler_Get_ReturnsServerError(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindErrFunc[quality.QualityAlert](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestQualityAlertHandler_UpdateState_UpdatesAlert(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindFunc(sampleQualityAlert(quality.AlertStateOpen)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	body := `{"state":"in_progress"}`
	resp, err := doRequest(app, http.MethodPut, "/quality-alerts/1/state", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestQualityAlertHandler_UpdateState_ReturnsNotFound(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindFunc[quality.QualityAlert](nil),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	body := `{"state":"in_progress"}`
	resp, err := doRequest(app, http.MethodPut, "/quality-alerts/1/state", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestQualityAlertHandler_UpdateState_ReturnsStateConflict(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindFunc(sampleQualityAlert(quality.AlertStateSolved)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	body := `{"state":"in_progress"}`
	resp, err := doRequest(app, http.MethodPut, "/quality-alerts/1/state", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
}

func TestQualityAlertHandler_UpdateState_RejectsValidation(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindFunc(sampleQualityAlert(quality.AlertStateOpen)),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodPut, "/quality-alerts/1/state", `{"state":"bogus"}`)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityAlertHandler_UpdateState_RejectsInvalidID(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	body := `{"state":"in_progress"}`
	resp, err := doRequest(app, http.MethodPut, "/quality-alerts/abc/state", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityAlertHandler_List_RejectsInvalidQuery(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityAlertHandler_Get_RejectsMissingTenant(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := fiber.New()
	h := NewQualityAlertHandler(svc)
	h.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/quality-alerts/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityAlertHandler_UpdateState_RejectsMissingTenant(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := fiber.New()
	h := NewQualityAlertHandler(svc)
	h.Register(app, passthroughGuards())

	body := `{"state":"in_progress"}`
	resp, err := doRequest(app, http.MethodPut, "/quality-alerts/1/state", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestQualityAlertHandler_UpdateState_ReturnsServerError(t *testing.T) {
	alerts := quality.QualityAlertDAOMock{
		CRUDMock: daoFindErrFunc[quality.QualityAlert](errors.New("db down")),
	}
	svc := quality.NewQualityCheckService(quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts)
	app := qualityHandlerTest(t, quality.QualityPointDAOMock{}, quality.QualityCheckDAOMock{}, alerts, svc)

	body := `{"state":"in_progress"}`
	resp, err := doRequest(app, http.MethodPut, "/quality-alerts/1/state", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			return c.Next()
		},
		Guard: func(_, _ string) fiber.Handler {
			return func(c fiber.Ctx) error {
				return c.Next()
			}
		},
	}
}

func doRequest(app *fiber.App, method, path, body string) (*http.Response, error) {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	return app.Test(req)
}
