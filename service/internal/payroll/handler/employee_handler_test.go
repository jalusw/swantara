package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func employeeHandlerTestSvc(employees payroll.EmployeeDAOMock, contracts payroll.EmploymentContractDAOMock, contactsDAO contacts.ContactDAOMock, users iam.UserDAOMock) payroll.HRService {
	return payroll.NewHRService(employees, contracts, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contactsDAO, dao.CRUDMock[reference.Dimension]{}, users, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
}

func employeeHandlerApp(t *testing.T, employees payroll.EmployeeDAOMock, contracts payroll.EmploymentContractDAOMock, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewEmployeeHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestEmployeeHandler_List_ReturnsEmployees(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Employee], error) {
				return &query.Page[payroll.Employee]{Items: []*payroll.Employee{
					{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1)), EmployeeNumber: "EMP-001"},
				}, Count: 1}, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestEmployeeHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_List_ReturnsUnauthorizedWithoutTenant(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, noTenantGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestEmployeeHandler_List_ReturnsServerError(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Employee], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestEmployeeHandler_Get_ReturnsEmployee(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1)), EmployeeNumber: "EMP-001"}, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/7", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestEmployeeHandler_Get_ReturnsNotFound(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/7", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEmployeeHandler_Get_RejectsForeignOrganization(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/7", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEmployeeHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Get_ReturnsServerError(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/employees/7", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestEmployeeHandler_Create_RejectsValidation(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"name":"","employee_number":"","wage":0}`
	resp, err := doRequest(app, http.MethodPost, "/employees", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Create_ReturnsUnresolvedOrganization(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, noTenantGuards())

	body := `{"name":"John Doe","employee_number":"EMP-014","wage":5000000}`
	resp, err := doRequest(app, http.MethodPost, "/employees", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Create_RejectsInvalidHireDate(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"name":"John Doe","employee_number":"EMP-014","wage":5000000,"hire_date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPost, "/employees", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_UpdatesEmployee(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1)), EmployeeNumber: "EMP-001"}, nil
			},
			SearchFunc: func(_ context.Context, _ string, _ any) (*payroll.Employee, error) {
				return nil, nil
			},
			UpdateFunc: func(_ context.Context, employee *payroll.Employee) (*payroll.Employee, error) {
				return employee, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002","employment_type":"full_time"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_RejectsInvalidID(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/abc", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_ReturnsNotFound(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_RejectsForeignOrganization(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(99))}, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_ReturnsServerErrorOnFind(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_RejectsValidation(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"name":"John Doe"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_RejectsInvalidHireDate(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002","hire_date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_RejectsInvalidTerminationDate(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1))}, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002","termination_date":"not-a-date"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_MapsNumberTaken(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1)), EmployeeNumber: "EMP-001"}, nil
			},
			SearchFunc: func(_ context.Context, field string, _ any) (*payroll.Employee, error) {
				if field == "employee_number" {
					return &payroll.Employee{Base: model.Base{ID: 99}, EmployeeNumber: "EMP-002"}, nil
				}
				return nil, nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Update_ReturnsServerErrorOnSave(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: helper.Ptr(uint64(1)), EmployeeNumber: "EMP-001"}, nil
			},
			SearchFunc: func(_ context.Context, _ string, _ any) (*payroll.Employee, error) {
				return nil, nil
			},
			UpdateFunc: func(_ context.Context, _ *payroll.Employee) (*payroll.Employee, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	body := `{"employee_number":"EMP-002"}`
	resp, err := doRequest(app, http.MethodPut, "/employees/7", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestEmployeeHandler_Delete_DeletesEmployee(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/employees/7", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestEmployeeHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := employeeHandlerTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/employees/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Delete_ReturnsServerError(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	svc := employeeHandlerTestSvc(employees, payroll.EmploymentContractDAOMock{}, contacts.ContactDAOMock{}, iam.UserDAOMock{})
	app := employeeHandlerApp(t, employees, payroll.EmploymentContractDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/employees/7", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
