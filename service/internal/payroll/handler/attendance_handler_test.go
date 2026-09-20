package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

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

func attendanceTestSvc(employees payroll.EmployeeDAOMock, attendances payroll.AttendanceDAOMock) payroll.HRService {
	return payroll.NewHRService(employees, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
}

func attendanceHandlerApp(t *testing.T, attendances payroll.AttendanceDAOMock, svc payroll.HRService, guards httpx.RouteGuards) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := NewAttendanceHandler(svc)
	handler.Register(app, guards)
	return app
}

func TestAttendanceHandler_List_ReturnsAttendances(t *testing.T) {
	checkIn := timeNow()
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Attendance], error) {
				return &query.Page[payroll.Attendance]{Items: []*payroll.Attendance{
					{Base: model.Base{ID: 5}, OrganizationID: 1, EmployeeID: 7, CheckIn: &checkIn, WorkedHours: 8, Status: payroll.AttendanceStatusDraft},
				}, Count: 1}, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAttendanceHandler_List_RejectsInvalidQuery(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/?filter=bogus:eq:x", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_List_ReturnsServerError(t *testing.T) {
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			ListFunc: func(_ context.Context, _ *query.Query) (*query.Page[payroll.Attendance], error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttendanceHandler_Get_ReturnsAttendance(t *testing.T) {
	checkIn := timeNow()
	checkOut := timeNow().Add(8 * time.Hour)
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return &payroll.Attendance{Base: model.Base{ID: 5}, OrganizationID: 1, EmployeeID: 7, CheckIn: &checkIn, CheckOut: &checkOut, WorkedHours: 8, Status: payroll.AttendanceStatusConfirmed}, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAttendanceHandler_Get_RejectsInvalidID(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_Get_ReturnsNotFound(t *testing.T) {
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return nil, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttendanceHandler_Get_ReturnsServerError(t *testing.T) {
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttendanceHandler_Get_ReturnsNotFoundForOtherTenant(t *testing.T) {
	checkIn := timeNow()
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return &payroll.Attendance{Base: model.Base{ID: 5}, OrganizationID: 99, EmployeeID: 7, CheckIn: &checkIn}, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodGet, "/attendances/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckIn_RecordsCheckIn(t *testing.T) {
	organizationID := uint64(1)
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: &organizationID, Active: true}, nil
			},
		},
	}
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			CreateFunc: func(_ context.Context, attendance *payroll.Attendance) (*payroll.Attendance, error) {
				attendance.ID = 5
				return attendance, nil
			},
		},
	}
	svc := attendanceTestSvc(employees, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	body := `{"employee_id":7,"check_in":"2026-08-05T08:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/check-in", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want 201", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckIn_RejectsValidation(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":0,"check_in":""}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/check-in", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckIn_RejectsInvalidTimestamp(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"check_in":"bad"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/check-in", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckIn_MapsMissingEmployee(t *testing.T) {
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return nil, nil
			},
		},
	}
	svc := attendanceTestSvc(employees, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	body := `{"employee_id":7,"check_in":"2026-08-05T08:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/check-in", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckIn_ReturnsServerError(t *testing.T) {
	organizationID := uint64(1)
	employees := payroll.EmployeeDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Employee]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Employee, error) {
				return &payroll.Employee{Base: model.Base{ID: 7}, OrganizationID: &organizationID, Active: true}, nil
			},
		},
	}
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			CreateFunc: func(_ context.Context, _ *payroll.Attendance) (*payroll.Attendance, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := attendanceTestSvc(employees, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	body := `{"employee_id":7,"check_in":"2026-08-05T08:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/check-in", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckOut_RecordsCheckOut(t *testing.T) {
	checkIn := timeNow()
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return &payroll.Attendance{Base: model.Base{ID: 5}, OrganizationID: 1, EmployeeID: 7, CheckIn: &checkIn}, nil
			},
			UpdateFunc: func(_ context.Context, attendance *payroll.Attendance) (*payroll.Attendance, error) {
				return attendance, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	body := `{"check_out":"2026-08-05T17:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/5/check-out", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckOut_RejectsInvalidID(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	body := `{"check_out":"2026-08-05T17:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/abc/check-out", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckOut_RejectsValidation(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	body := `{"check_out":""}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/5/check-out", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckOut_RejectsInvalidTimestamp(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	body := `{"check_out":"bad"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/5/check-out", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckOut_ReturnsNotFound(t *testing.T) {
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return nil, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	body := `{"check_out":"2026-08-05T17:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/5/check-out", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckOut_MapsCheckOutBeforeCheckIn(t *testing.T) {
	checkIn := timeNow().AddDate(1, 0, 0)
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return &payroll.Attendance{Base: model.Base{ID: 5}, OrganizationID: 1, EmployeeID: 7, CheckIn: &checkIn}, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	body := `{"check_out":"2026-08-05T08:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/5/check-out", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

func TestAttendanceHandler_CheckOut_ReturnsServerError(t *testing.T) {
	checkIn := timeNow()
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return &payroll.Attendance{Base: model.Base{ID: 5}, OrganizationID: 1, EmployeeID: 7, CheckIn: &checkIn}, nil
			},
			UpdateFunc: func(_ context.Context, _ *payroll.Attendance) (*payroll.Attendance, error) {
				return nil, errors.New("db down")
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	body := `{"check_out":"2026-08-05T17:00:00Z"}`
	resp, err := doRequest(app, http.MethodPost, "/attendances/5/check-out", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", resp.StatusCode)
	}
}

func TestAttendanceHandler_Delete_DeletesAttendance(t *testing.T) {
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return &payroll.Attendance{Base: model.Base{ID: 5}, OrganizationID: 1, EmployeeID: 7}, nil
			},
			DeleteFunc: func(_ context.Context, _ uint64) error {
				return nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/attendances/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", resp.StatusCode)
	}
}

func TestAttendanceHandler_Delete_ReturnsNotFound(t *testing.T) {
	attendances := payroll.AttendanceDAOMock{
		CRUDMock: dao.CRUDMock[payroll.Attendance]{
			FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
				return nil, nil
			},
		},
	}
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, attendances)
	app := attendanceHandlerApp(t, attendances, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/attendances/5", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAttendanceHandler_Delete_RejectsInvalidID(t *testing.T) {
	svc := attendanceTestSvc(payroll.EmployeeDAOMock{}, payroll.AttendanceDAOMock{})
	app := attendanceHandlerApp(t, payroll.AttendanceDAOMock{}, svc, passthroughGuards())

	resp, err := doRequest(app, http.MethodDelete, "/attendances/abc", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", resp.StatusCode)
	}
}

var _ = time.RFC3339
