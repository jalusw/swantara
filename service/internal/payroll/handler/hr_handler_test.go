package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func passthroughGuards() httpx.RouteGuards {
	return httpx.RouteGuards{
		AuthN: func(c fiber.Ctx) error {
			c.Locals(httpx.LocalOrganizationID, uint64(1))
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

func TestLeaveRequestHandler_Create_RejectsInvalidDateRange(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}}, nil
			},
		},
	}
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return &reference.LeaveType{Base: model.Base{ID: 3}}, nil
		},
	}
	svc := payroll.NewHRService(employees, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	handler := NewLeaveRequestHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"employee_id":7,"leave_type_id":3,"date_from":"2026-08-10","date_to":"2026-08-05","days":2}`
	resp, err := doRequest(app, http.MethodPost, "/leave-requests", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Create_RejectsUnknownLeaveType(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}}, nil
			},
		},
	}
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, nil
		},
	}
	svc := payroll.NewHRService(employees, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	handler := NewLeaveRequestHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"employee_id":7,"leave_type_id":3,"date_from":"2026-08-05","date_to":"2026-08-10","days":2}`
	resp, err := doRequest(app, http.MethodPost, "/leave-requests", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Create_CreatesDraft(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}}, nil
			},
		},
	}
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return &reference.LeaveType{Base: model.Base{ID: 3}}, nil
		},
	}
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			CreateFunc: func(_ context.Context, request *payroll.LeaveRequest) (*payroll.LeaveRequest, error) {
				request.ID = 9
				return request, nil
			},
		},
	}
	svc := payroll.NewHRService(employees, payroll.EmploymentContractDAOMock{}, leaves, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	handler := NewLeaveRequestHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"employee_id":7,"leave_type_id":3,"date_from":"2026-08-05","date_to":"2026-08-10","days":2}`
	resp, err := doRequest(app, http.MethodPost, "/leave-requests", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestContractHandler_Terminate_ClosesActive(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, State: payroll.ContractStateActive}, nil
			},
			UpdateFunc: func(_ context.Context, contract *payroll.EmploymentContract) (*payroll.EmploymentContract, error) {
				return contract, nil
			},
		},
	}
	svc := payroll.NewHRService(payroll.EmployeeDAOMock{}, contracts, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	handler := NewContractHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/contracts/4/terminate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestContractHandler_Terminate_RejectsAlreadyClosed(t *testing.T) {
	contracts := payroll.EmploymentContractDAOMock{
		CRUDMock: dao.CRUDMock[payroll.EmploymentContract]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.EmploymentContract, error) {
				return &payroll.EmploymentContract{Base: model.Base{ID: 4}, EmployeeID: 7, State: payroll.ContractStateClosed}, nil
			},
		},
	}
	svc := payroll.NewHRService(payroll.EmployeeDAOMock{}, contracts, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	handler := NewContractHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/contracts/4/terminate", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestEmployeeHandler_Create_ProvisionsUserLink(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			CreateFunc: func(_ context.Context, employee *payroll.Employee) (*payroll.Employee, error) {
				employee.ID = 7
				return employee, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, nil
			},
			SearchFunc: func(_ context.Context, _ string, _ any) (*payroll.Employee, error) {
				return nil, nil
			},
		},
	}
	contactDAO := contacts.ContactDAOMock{
		CreateWithDetailsFunc: func(_ context.Context, contact *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
			contact.ID = 3
			return contact, nil
		},
	}
	svc := payroll.NewHRService(employees, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contactDAO, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{DAOMock: iam.DAOMock[iam.User]{FindFunc: func(_ context.Context, _ uint64) (*iam.User, error) { return &iam.User{}, nil }}}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	handler := NewEmployeeHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"organization_id":1,"name":"John Doe","employee_number":"EMP-011","user_id":9,"wage":5000000}`
	resp, err := doRequest(app, http.MethodPost, "/employees", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestEmployeeHandler_Create_RejectsMissingUser(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			CreateFunc: func(_ context.Context, employee *payroll.Employee) (*payroll.Employee, error) {
				employee.ID = 7
				return employee, nil
			},
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, nil
			},
			SearchFunc: func(_ context.Context, _ string, _ any) (*payroll.Employee, error) {
				return nil, nil
			},
		},
	}
	contactDAO := contacts.ContactDAOMock{
		CreateWithDetailsFunc: func(_ context.Context, contact *contacts.Contact, _ []*contacts.ContactAddress, _ []*contacts.ContactBankAccount, _ *contacts.CustomerProfile, _ *contacts.SupplierProfile) (*contacts.Contact, error) {
			contact.ID = 3
			return contact, nil
		},
	}
	svc := payroll.NewHRService(employees, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contactDAO, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
	handler := NewEmployeeHandler(svc)

	app := fiber.New()
	handler.Register(app, passthroughGuards())

	body := `{"organization_id":1,"name":"John Doe","employee_number":"EMP-013","user_id":99,"wage":5000000}`
	resp, err := doRequest(app, http.MethodPost, "/employees", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}
