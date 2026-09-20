package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/kernel/query"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func leaveRequestTestSvc(employees payroll.EmployeeDAOMock, contracts payroll.EmploymentContractDAOMock, leaves payroll.LeaveRequestDAOMock, leaveTypes dao.CRUD[reference.LeaveType]) payroll.HRService {
	return payroll.NewHRService(employees, contracts, leaves, leaveTypes, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
}

func leaveRequestHandlerApp(t *testing.T, leaves payroll.LeaveRequestDAOMock, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewLeaveRequestHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestLeaveRequestHandler_List_ReturnsRequests(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.LeaveRequest], error) {
				return &query.Page[payroll.LeaveRequest]{Items: []*payroll.LeaveRequest{
					{Base: model.Base{ID: 9}, EmployeeID: 7, LeaveTypeID: 3, DateFrom: timeNow(), DateTo: timeNow().AddDate(0, 0, 1), Days: 2, State: "draft"},
				}, Count: 1}, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_List_ReturnsServerError(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.LeaveRequest], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Get_ReturnsRequest(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, EmployeeID: 7, LeaveTypeID: 3, DateFrom: timeNow(), DateTo: timeNow().AddDate(0, 0, 1), Days: 2, State: "draft"}, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/9", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Get_ReturnsNotFound(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return nil, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/9", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Get_ReturnsServerError(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/9", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Create_RejectsValidation(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":0,"leave_type_id":0,"date_from":"","date_to":"","days":0}`
	resp, err := doRequest(app, http.MethodPost, "/leave-requests", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Create_RejectsInvalidDateFrom(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"leave_type_id":3,"date_from":"bad","date_to":"2026-08-10","days":2}`
	resp, err := doRequest(app, http.MethodPost, "/leave-requests", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Create_RejectsInvalidDateTo(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"leave_type_id":3,"date_from":"2026-08-05","date_to":"bad","days":2}`
	resp, err := doRequest(app, http.MethodPost, "/leave-requests", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Submit_SubmitsDraft(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, EmployeeID: 7, LeaveTypeID: 3, DateFrom: timeNow(), DateTo: timeNow().AddDate(0, 0, 1), Days: 2, State: "draft"}, nil
			},
			UpdateFunc: func(_ context.Context, leave *payroll.LeaveRequest) (*payroll.LeaveRequest, error) {
				return leave, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Submit_RejectsInvalidID(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/abc/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Submit_ReturnsNotFound(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return nil, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Submit_MapsInvalidState(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, State: "submitted"}, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/submit", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Approve_ApprovesSubmitted(t *testing.T) {
	allocation := 12.0
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return &reference.LeaveType{Base: model.Base{ID: 3}, AllocationDays: &allocation}, nil
		},
	}
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, EmployeeID: 7, LeaveTypeID: 3, DateFrom: timeNow(), DateTo: timeNow().AddDate(0, 0, 1), Days: 2, State: "submitted"}, nil
			},
			UpdateFunc: func(_ context.Context, leave *payroll.LeaveRequest) (*payroll.LeaveRequest, error) {
				return leave, nil
			},
		},
		ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*payroll.LeaveRequest, error) {
			return []*payroll.LeaveRequest{}, nil
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, leaveTypes)
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Approve_RejectsInvalidID(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/abc/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Approve_MapsInsufficientBalance(t *testing.T) {
	allocation := 12.0
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return &reference.LeaveType{Base: model.Base{ID: 3}, AllocationDays: &allocation}, nil
		},
	}
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, EmployeeID: 7, LeaveTypeID: 3, DateFrom: timeNow(), DateTo: timeNow().AddDate(0, 0, 1), Days: 13, State: "submitted"}, nil
			},
		},
		ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*payroll.LeaveRequest, error) {
			return []*payroll.LeaveRequest{}, nil
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, leaveTypes)
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Approve_MapsInvalidState(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, State: "draft"}, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/approve", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Refuse_RefusesSubmitted(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, EmployeeID: 7, LeaveTypeID: 3, DateFrom: timeNow(), DateTo: timeNow().AddDate(0, 0, 1), Days: 2, State: "submitted"}, nil
			},
			UpdateFunc: func(_ context.Context, leave *payroll.LeaveRequest) (*payroll.LeaveRequest, error) {
				return leave, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/refuse", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Refuse_RejectsInvalidID(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/abc/refuse", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Refuse_MapsInvalidState(t *testing.T) {
	leaves := payroll.LeaveRequestDAOMock{
		CRUDMock: dao.CRUDMock[payroll.LeaveRequest]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.LeaveRequest, error) {
				return &payroll.LeaveRequest{Base: model.Base{ID: 9}, State: "draft"}, nil
			},
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodPost, "/leave-requests/9/refuse", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Balance_ReturnsBalance(t *testing.T) {
	allocation := 12.0
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return &reference.LeaveType{Base: model.Base{ID: 3}, AllocationDays: &allocation}, nil
		},
	}
	leaves := payroll.LeaveRequestDAOMock{
		ListApprovedByEmployeeAndTypeFunc: func(_ context.Context, _, _ uint64) ([]*payroll.LeaveRequest, error) {
			return []*payroll.LeaveRequest{
				{Base: model.Base{ID: 9}, Days: 2},
			}, nil
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, leaves, leaveTypes)
	app := leaveRequestHandlerApp(t, leaves, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/balance/7/3", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Balance_RejectsInvalidEmployeeID(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/balance/abc/3", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Balance_RejectsInvalidLeaveTypeID(t *testing.T) {
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{})
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/balance/7/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestLeaveRequestHandler_Balance_ReturnsNotFound(t *testing.T) {
	leaveTypes := dao.CRUDMock[reference.LeaveType]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.LeaveType, error) {
			return nil, nil
		},
	}
	svc := leaveRequestTestSvc(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, leaveTypes)
	app := leaveRequestHandlerApp(t, payroll.LeaveRequestDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/leave-requests/balance/7/3", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}
