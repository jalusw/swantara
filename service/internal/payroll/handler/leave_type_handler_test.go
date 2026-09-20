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

func leaveTypeHandlerApp(t *testing.T, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewLeaveTypeHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestLeaveTypeHandler_List_ReturnsLeaveTypes(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.LeaveType], error) {
			orgID := uint64(1)
			allocDays := 12.0
			return &query.Page[reference.LeaveType]{
				Items: []*reference.LeaveType{
					{Base: model.Base{ID: 1}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays},
				},
				Count: 1,
			}, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := leaveTypeHandlerApp(t, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_List_ReturnsServerError(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[reference.LeaveType], error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Get_ReturnsLeaveType(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(1)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Get_ReturnsNotFound(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Get_RejectsForeignOrganization(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(99)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Get_ReturnsServerError(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Create_CreatesLeaveType(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		CreateFunc: func(_ context.Context, lt *reference.LeaveType) (*reference.LeaveType, error) {
			lt.Base = model.Base{ID: 1}
			return lt, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual","paid":true,"allocation_days":12}`
	resp, err := doRequest(app, http.MethodPost, "/leave-types/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Create_RejectsValidation(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPost, "/leave-types/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := leaveTypeHandlerApp(t, svc, noTenantGuards())

	body := `{"name":"Annual"}`
	resp, err := doRequest(app, http.MethodPost, "/leave-types/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Create_ReturnsServerError(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		CreateFunc: func(_ context.Context, _ *reference.LeaveType) (*reference.LeaveType, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual"}`
	resp, err := doRequest(app, http.MethodPost, "/leave-types/", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Update_UpdatesLeaveType(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(1)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
		UpdateFunc: func(_ context.Context, lt *reference.LeaveType) (*reference.LeaveType, error) {
			return lt, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual Leave","paid":true,"allocation_days":10}`
	resp, err := doRequest(app, http.MethodPut, "/leave-types/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual Leave","paid":true,"allocation_days":10}`
	resp, err := doRequest(app, http.MethodPut, "/leave-types/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Update_RejectsValidation(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(1)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":""}`
	resp, err := doRequest(app, http.MethodPut, "/leave-types/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Update_ReturnsNotFound(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual Leave","paid":true,"allocation_days":10}`
	resp, err := doRequest(app, http.MethodPut, "/leave-types/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Update_RejectsForeignOrganization(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(99)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual Leave","paid":true,"allocation_days":10}`
	resp, err := doRequest(app, http.MethodPut, "/leave-types/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual Leave","paid":true,"allocation_days":10}`
	resp, err := doRequest(app, http.MethodPut, "/leave-types/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(1)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
		UpdateFunc: func(_ context.Context, _ *reference.LeaveType) (*reference.LeaveType, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	body := `{"name":"Annual Leave","paid":true,"allocation_days":10}`
	resp, err := doRequest(app, http.MethodPut, "/leave-types/1", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Delete_DeletesLeaveType(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(1)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := hrServiceFor(nil, nil, nil)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/leave-types/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Delete_ReturnsNotFound(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Delete_RejectsForeignOrganization(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(99)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Delete_ReturnsServerErrorOnFind(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveTypeHandler_Delete_ReturnsServerErrorOnDelete(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, id uint64) (*reference.LeaveType, error) {
			orgID := uint64(1)
			allocDays := 12.0
			return &reference.LeaveType{Base: model.Base{ID: id}, OrganizationID: &orgID, Name: "Annual", Paid: true, AllocationDays: &allocDays}, nil
		},
		DeleteFunc: func(_ context.Context, _ uint64) error {
			return errors.New("db down")
		},
	}
	svc := hrServiceFor(nil, nil, &leaveTypes)
	app := leaveTypeHandlerApp(t, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/leave-types/1", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
