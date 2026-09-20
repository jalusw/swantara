package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func jobPositionHandlerApp(t *testing.T, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewJobPositionHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestJobPositionHandler_List_ReturnsPositions(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.JobPosition], error) {
			orgID := uint64(1)
			return &query.Page[reference.JobPosition]{
				Items: []*reference.JobPosition{
					{Base: model.Base{ID: 1}, OrganizationID: &orgID, Name: "Engineer"},
				},
				Count: 1,
			}, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestJobPositionHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestJobPositionHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := jobPositionHandlerApp(t, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestJobPositionHandler_List_ReturnsServerError(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.JobPosition], error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestJobPositionHandler_Get_ReturnsPosition(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(1)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestJobPositionHandler_Get_ReturnsNotFound(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.JobPosition, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestJobPositionHandler_Get_RejectsForeignOrganization(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(99)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestJobPositionHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestJobPositionHandler_Get_ReturnsServerError(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.JobPosition, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestJobPositionHandler_Create_CreatesPosition(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		CreateFunc: func(_ context.Context, pos *reference.JobPosition) (*reference.JobPosition, error) {
			pos.Base = model.Base{ID: 1}
			return pos, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineer"}`
	resp, err := doRequest(app, http.MethodPost, "/job-positions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestJobPositionHandler_Create_RejectsValidation(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/job-positions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestJobPositionHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := jobPositionHandlerApp(t, svc, noTenantGuards())

	body := `{"name":"Engineer"}`
	resp, err := doRequest(app, http.MethodPost, "/job-positions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestJobPositionHandler_Create_ReturnsServerError(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		CreateFunc: func(_ context.Context, _ *reference.JobPosition) (*reference.JobPosition, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineer"}`
	resp, err := doRequest(app, http.MethodPost, "/job-positions/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestJobPositionHandler_Update_UpdatesPosition(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(1)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
		UpdateFunc: func(_ context.Context, pos *reference.JobPosition) (*reference.JobPosition, error) {
			return pos, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Senior Engineer"}`
	resp, err := doRequest(app, http.MethodPut, "/job-positions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestJobPositionHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Senior Engineer"}`
	resp, err := doRequest(app, http.MethodPut, "/job-positions/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestJobPositionHandler_Update_RejectsValidation(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(1)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPut, "/job-positions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestJobPositionHandler_Update_ReturnsNotFound(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.JobPosition, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Senior Engineer"}`
	resp, err := doRequest(app, http.MethodPut, "/job-positions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestJobPositionHandler_Update_RejectsForeignOrganization(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(99)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Senior Engineer"}`
	resp, err := doRequest(app, http.MethodPut, "/job-positions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestJobPositionHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.JobPosition, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Senior Engineer"}`
	resp, err := doRequest(app, http.MethodPut, "/job-positions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestJobPositionHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(1)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
		UpdateFunc: func(_ context.Context, _ *reference.JobPosition) (*reference.JobPosition, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Senior Engineer"}`
	resp, err := doRequest(app, http.MethodPut, "/job-positions/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestJobPositionHandler_Delete_DeletesPosition(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(1)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestJobPositionHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/job-positions/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestJobPositionHandler_Delete_ReturnsNotFound(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.JobPosition, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestJobPositionHandler_Delete_RejectsForeignOrganization(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(99)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestJobPositionHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.JobPosition, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestJobPositionHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	positions := dao.CRUDMock[reference.JobPosition]{
		FindFunc: func(_ context.Context, id uint64) (*reference.JobPosition, error) {
			orgID := uint64(1)
			return &reference.JobPosition{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineer"}, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, &positions, nil)
	app := jobPositionHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/job-positions/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
