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

func departmentHandlerApp(t *testing.T, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewDepartmentHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestDepartmentHandler_List_ReturnsDepartments(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Department], error) {
			orgID := uint64(1)
			return &query.Page[reference.Department]{
				Items: []*reference.Department{
					{Base: model.Base{ID: 1}, OrganizationID: &orgID, Name: "Engineering"},
				},
				Count: 1,
			}, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDepartmentHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := hrServiceFor(&dao.CRUDMock[reference.Department]{}, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDepartmentHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	svc := hrServiceFor(&dao.CRUDMock[reference.Department]{}, nil, nil)
	app := departmentHandlerApp(t, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestDepartmentHandler_List_ReturnsServerError(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.Department], error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestDepartmentHandler_Get_ReturnsDepartment(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(1)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDepartmentHandler_Get_ReturnsNotFound(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Department, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDepartmentHandler_Get_RejectsForeignOrganization(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(99)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDepartmentHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(&dao.CRUDMock[reference.Department]{}, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDepartmentHandler_Get_ReturnsServerError(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Department, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestDepartmentHandler_Create_CreatesDepartment(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		CreateFunc: func(_ context.Context, dept *reference.Department) (*reference.Department, error) {
			dept.Base = model.Base{ID: 1}
			return dept, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering"}`
	resp, err := doRequest(app, http.MethodPost, "/departments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestDepartmentHandler_Create_RejectsValidation(t *testing.T) {
	svc := hrServiceFor(&dao.CRUDMock[reference.Department]{}, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/departments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDepartmentHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := hrServiceFor(&dao.CRUDMock[reference.Department]{}, nil, nil)
	app := departmentHandlerApp(t, svc, noTenantGuards())

	body := `{"name":"Engineering"}`
	resp, err := doRequest(app, http.MethodPost, "/departments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDepartmentHandler_Create_ReturnsServerError(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		CreateFunc: func(_ context.Context, _ *reference.Department) (*reference.Department, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering"}`
	resp, err := doRequest(app, http.MethodPost, "/departments/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestDepartmentHandler_Update_UpdatesDepartment(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(1)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
		UpdateFunc: func(_ context.Context, dept *reference.Department) (*reference.Department, error) {
			return dept, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering Services"}`
	resp, err := doRequest(app, http.MethodPut, "/departments/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestDepartmentHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(&dao.CRUDMock[reference.Department]{}, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering Services"}`
	resp, err := doRequest(app, http.MethodPut, "/departments/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDepartmentHandler_Update_RejectsValidation(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(1)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPut, "/departments/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDepartmentHandler_Update_ReturnsNotFound(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Department, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering Services"}`
	resp, err := doRequest(app, http.MethodPut, "/departments/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDepartmentHandler_Update_RejectsForeignOrganization(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(99)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering Services"}`
	resp, err := doRequest(app, http.MethodPut, "/departments/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDepartmentHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Department, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering Services"}`
	resp, err := doRequest(app, http.MethodPut, "/departments/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestDepartmentHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(1)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
		UpdateFunc: func(_ context.Context, _ *reference.Department) (*reference.Department, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Engineering Services"}`
	resp, err := doRequest(app, http.MethodPut, "/departments/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestDepartmentHandler_Delete_DeletesDepartment(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(1)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestDepartmentHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(&dao.CRUDMock[reference.Department]{}, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/departments/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestDepartmentHandler_Delete_ReturnsNotFound(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Department, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDepartmentHandler_Delete_RejectsForeignOrganization(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(99)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestDepartmentHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Department, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestDepartmentHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	departments := dao.CRUDMock[reference.Department]{
		FindFunc: func(_ context.Context, id uint64) (*reference.Department, error) {
			orgID := uint64(1)
			return &reference.Department{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Engineering"}, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	}
	svc := hrServiceFor(&departments, nil, nil)
	app := departmentHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/departments/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
