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

func timesheetTestSvc(employees payroll.EmployeeDAOMock, dimensions dao.CRUD[reference.Dimension], timesheets payroll.TimesheetDAOMock) payroll.HRService {
	return payroll.NewHRService(employees, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dimensions, iam.UserDAOMock{}, payroll.AttendanceDAOMock{}, timesheets, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
}

func timesheetHandlerApp(t *testing.T, timesheets payroll.TimesheetDAOMock, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewTimesheetHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestTimesheetHandler_List_ReturnsTimesheets(t *testing.T) {
	description := "code review"
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Timesheet], error) {
				return &query.Page[payroll.Timesheet]{Items: []*payroll.Timesheet{
					{Base: model.Base{ID: 4}, EmployeeID: 7, Date: timeNow(), Hours: 8, Description: &description},
				}, Count: 1}, nil
			},
		},
	}
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/timesheets/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestTimesheetHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/timesheets/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_List_ReturnsServerError(t *testing.T) {
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Timesheet], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/timesheets/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestTimesheetHandler_Get_ReturnsTimesheet(t *testing.T) {
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Timesheet, error) {
				return &payroll.Timesheet{Base: model.Base{ID: 4}, EmployeeID: 7, Date: timeNow(), Hours: 8}, nil
			},
		},
	}
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/timesheets/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestTimesheetHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/timesheets/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_Get_ReturnsNotFound(t *testing.T) {
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Timesheet, error) {
				return nil, nil
			},
		},
	}
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/timesheets/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestTimesheetHandler_Get_ReturnsServerError(t *testing.T) {
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Timesheet, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/timesheets/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestTimesheetHandler_Create_CreatesTimesheet(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}}, nil
			},
		},
	}
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			CreateFunc: func(_ context.Context, timesheet *payroll.Timesheet) (*payroll.Timesheet, error) {
				timesheet.ID = 4
				return timesheet, nil
			},
		},
	}
	svc := timesheetTestSvc(employees, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	body := `{"employee_id":7,"date":"2026-08-05","hours":8,"description":"code review"}`
	resp, err := doRequest(app, http.MethodPost, "/timesheets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestTimesheetHandler_Create_RejectsValidation(t *testing.T) {
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":0,"date":"","hours":0}`
	resp, err := doRequest(app, http.MethodPost, "/timesheets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_Create_RejectsInvalidDate(t *testing.T) {
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"date":"bad","hours":8}`
	resp, err := doRequest(app, http.MethodPost, "/timesheets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_Create_MapsMissingEmployee(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, nil
			},
		},
	}
	svc := timesheetTestSvc(employees, dao.CRUDMock[reference.Dimension]{}, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"date":"2026-08-05","hours":8}`
	resp, err := doRequest(app, http.MethodPost, "/timesheets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_Create_MapsInvalidHours(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}}, nil
			},
		},
	}
	svc := timesheetTestSvc(employees, dao.CRUDMock[reference.Dimension]{}, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"date":"2026-08-05","hours":-1}`
	resp, err := doRequest(app, http.MethodPost, "/timesheets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_Create_MapsMissingDimension(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}}, nil
			},
		},
	}
	dimensions := dao.CRUDMock[reference.Dimension]{
		FindFunc: func(_ context.Context, _ uint64) (*reference.Dimension, error) {
			return nil, nil
		},
	}
	svc := timesheetTestSvc(employees, dimensions, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"date":"2026-08-05","hours":8,"dimension_id":99}`
	resp, err := doRequest(app, http.MethodPost, "/timesheets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_Create_ReturnsServerError(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}}, nil
			},
		},
	}
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			CreateFunc: func(_ context.Context, _ *payroll.Timesheet) (*payroll.Timesheet, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := timesheetTestSvc(employees, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	body := `{"employee_id":7,"date":"2026-08-05","hours":8}`
	resp, err := doRequest(app, http.MethodPost, "/timesheets", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestTimesheetHandler_Delete_DeletesTimesheet(t *testing.T) {
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/timesheets/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestTimesheetHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, payroll.TimesheetDAOMock{})
	app := timesheetHandlerApp(t, payroll.TimesheetDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/timesheets/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestTimesheetHandler_Delete_ReturnsServerError(t *testing.T) {
	timesheets := payroll.TimesheetDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Timesheet]{
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return errors.New("db down")
			},
		},
	}
	svc := timesheetTestSvc(payroll.EmployeeDAOMock{}, dao.CRUDMock[reference.Dimension]{}, timesheets)
	app := timesheetHandlerApp(t, timesheets, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/timesheets/4", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}
