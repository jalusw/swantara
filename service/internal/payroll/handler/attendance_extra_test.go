package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func fullHRService(attendances payroll.AttendanceDAOMock) payroll.HRService {
	return payroll.NewHRService(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, attendances, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
}

func orgAttendance() *payroll.Attendance {
	checkIn := time.Date(2026, 8, 1, 9, 0, 0, 0, time.UTC)
	return &payroll.Attendance{Base: model.Base{ID: 1}, OrganizationID: 1, EmployeeID: 7, CheckIn: &checkIn, Status: payroll.AttendanceStatusDraft}
}

func TestAttendanceHandler_Update_Delete(t *testing.T) {
	newApp := func(attendances payroll.AttendanceDAOMock, svc payroll.HRService) *fiber.App {
		app := fiber.New()
		NewAttendanceHandler(svc).Register(app, passthroughGuards())
		return app
	}

	t.Run("rejects invalid id", func(t *testing.T) {
		app := newApp(payroll.AttendanceDAOMock{}, fullHRService(payroll.AttendanceDAOMock{}))
		for _, tc := range []struct {
			method string
			path   string
		}{
			{method: http.MethodPut, path: "/attendances/abc"},
			{method: http.MethodDelete, path: "/attendances/abc"},
		} {
			resp, err := doRequest(app, tc.method, tc.path, "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("updates attendance", func(t *testing.T) {
		attendances := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
					return orgAttendance(), nil
				},
			},
		}
		app := newApp(attendances, fullHRService(attendances))
		resp, err := doRequest(app, http.MethodPut, "/attendances/1", `{"break_minutes":30}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("update maps errors", func(t *testing.T) {
		missing := payroll.AttendanceDAOMock{}
		app := newApp(missing, fullHRService(missing))
		resp, err := doRequest(app, http.MethodPut, "/attendances/1", `{"break_minutes":30}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		foreign := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
					att := orgAttendance()
					att.OrganizationID = 99
					return att, nil
				},
			},
		}
		app = newApp(foreign, fullHRService(foreign))
		resp, err = doRequest(app, http.MethodPut, "/attendances/1", `{"break_minutes":30}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		failing := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
					return nil, errors.New("db down")
				},
			},
		}
		app = newApp(failing, fullHRService(failing))
		resp, err = doRequest(app, http.MethodPut, "/attendances/1", `{"break_minutes":30}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("deletes attendance", func(t *testing.T) {
		attendances := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
					return orgAttendance(), nil
				},
			},
		}
		app := newApp(attendances, fullHRService(attendances))
		resp, err := doRequest(app, http.MethodDelete, "/attendances/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNoContent)
	})

	t.Run("delete maps errors", func(t *testing.T) {
		missing := payroll.AttendanceDAOMock{}
		app := newApp(missing, fullHRService(missing))
		resp, err := doRequest(app, http.MethodDelete, "/attendances/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)

		failing := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
					return nil, errors.New("db down")
				},
			},
		}
		app = newApp(failing, fullHRService(failing))
		resp, err = doRequest(app, http.MethodDelete, "/attendances/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestAttendanceHandler_Bulk_Missing(t *testing.T) {
	newApp := func(svc payroll.HRService) *fiber.App {
		app := fiber.New()
		NewAttendanceHandler(svc).Register(app, passthroughGuards())
		return app
	}

	t.Run("rejects invalid bulk body", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodPost, "/attendances/bulk-check-out", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/attendances/bulk-check-out", `{"attendance_ids":[1],"check_out":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("bulk checks out", func(t *testing.T) {
		attendances := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc:   func(_ context.Context, _ uint64) (*payroll.Attendance, error) { return orgAttendance(), nil },
				UpdateFunc: func(_ context.Context, a *payroll.Attendance) (*payroll.Attendance, error) { return a, nil },
			},
		}
		app := newApp(fullHRService(attendances))
		resp, err := doRequest(app, http.MethodPost, "/attendances/bulk-check-out", `{"attendance_ids":[1],"check_out":"2026-08-01T17:30:00Z"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("missing requires date and org", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodGet, "/attendances/missing", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodGet, "/attendances/missing?date=2026-08-01", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})
}

func TestAttendanceHandler_Approve_Reject(t *testing.T) {
	newApp := func(svc payroll.HRService) *fiber.App {
		app := fiber.New()
		NewAttendanceHandler(svc).Register(app, passthroughGuards())
		return app
	}

	t.Run("rejects invalid id", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		for _, path := range []string{"/attendances/abc/approve", "/attendances/abc/reject"} {
			resp, err := doRequest(app, http.MethodPost, path, "")
			if err != nil {
				t.Fatal(err)
			}
			helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
		}
	})

	t.Run("approves and rejects", func(t *testing.T) {
		attendances := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
					pending := orgAttendance()
					pending.Status = payroll.AttendanceStatusPendingApproval
					return pending, nil
				},
				UpdateFunc: func(_ context.Context, a *payroll.Attendance) (*payroll.Attendance, error) { return a, nil },
			},
		}
		app := newApp(fullHRService(attendances))
		resp, err := doRequest(app, http.MethodPost, "/attendances/1/approve", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)

		resp, err = doRequest(app, http.MethodPost, "/attendances/1/reject", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("maps approve error", func(t *testing.T) {
		attendances := payroll.AttendanceDAOMock{
			CRUDMock: dao.CRUDMock[payroll.Attendance]{
				FindFunc: func(_ context.Context, _ uint64) (*payroll.Attendance, error) {
					return nil, errors.New("db down")
				},
			},
		}
		app := newApp(fullHRService(attendances))
		resp, err := doRequest(app, http.MethodPost, "/attendances/1/approve", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)

		resp, err = doRequest(app, http.MethodPost, "/attendances/1/reject", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusInternalServerError)
	})
}

func TestAttendanceHandler_Auto_MarkAbsent(t *testing.T) {
	newApp := func(svc payroll.HRService) *fiber.App {
		app := fiber.New()
		NewAttendanceHandler(svc).Register(app, passthroughGuards())
		return app
	}

	t.Run("rejects invalid bodies", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodPost, "/attendances/auto-checkout", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/attendances/mark-absent", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("runs auto checkout and mark absent", func(t *testing.T) {
		autoSvc := payroll.NewHRService(payroll.EmployeeDAOMock{}, payroll.EmploymentContractDAOMock{}, payroll.LeaveRequestDAOMock{}, dao.CRUDMock[reference.LeaveType]{}, dao.CRUDMock[reference.Department]{}, dao.CRUDMock[reference.JobPosition]{}, dao.CRUDMock[reference.Organization]{}, contacts.ContactDAOMock{}, dao.CRUDMock[reference.Dimension]{}, iam.UserDAOMock{}, payroll.AttendanceDAOWithOrgMock{}, payroll.TimesheetDAOMock{}, payroll.ShiftDAOMock{}, payroll.ShiftAssignmentDAOMock{}, inventory.TransactionerMock{})
		app := newApp(autoSvc)
		resp, err := doRequest(app, http.MethodPost, "/attendances/auto-checkout", `{"organization_id":1}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)

		resp, err = doRequest(app, http.MethodPost, "/attendances/mark-absent", `{"date":"2026-08-01"}`)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d", resp.StatusCode)
		}
	})
}

func TestShiftHandler_CRUD(t *testing.T) {
	newApp := func(svc payroll.HRService) *fiber.App {
		app := fiber.New()
		NewShiftHandler(svc).Register(app, passthroughGuards())
		return app
	}

	t.Run("lists shifts", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodGet, "/shifts/", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusOK)
	})

	t.Run("creates shift", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodPost, "/shifts/", `{"organization_id":1,"name":"Morning","start_time":"08:00","end_time":"17:00"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusCreated)

		resp, err = doRequest(app, http.MethodPost, "/shifts/", `{"name":"Morning"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("updates and deletes shift", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodPut, "/shifts/abc", `{"organization_id":1,"name":"Morning","start_time":"08:00","end_time":"17:00"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodDelete, "/shifts/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodDelete, "/shifts/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})
}

func TestShiftAssignmentHandler_CRUD(t *testing.T) {
	newApp := func(svc payroll.HRService) *fiber.App {
		app := fiber.New()
		NewShiftAssignmentHandler(svc).Register(app, passthroughGuards())
		return app
	}

	t.Run("rejects invalid body", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodPost, "/shift-assignments/", `{}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)

		resp, err = doRequest(app, http.MethodPost, "/shift-assignments/", `{"employee_id":7,"shift_id":3,"date":"yesterday"}`)
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("rejects invalid delete id", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodDelete, "/shift-assignments/abc", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusUnprocessableEntity)
	})

	t.Run("deletes missing assignment", func(t *testing.T) {
		app := newApp(fullHRService(payroll.AttendanceDAOMock{}))
		resp, err := doRequest(app, http.MethodDelete, "/shift-assignments/1", "")
		if err != nil {
			t.Fatal(err)
		}
		helper.AssertStatus(t, resp.StatusCode, http.StatusNotFound)
	})
}

var _ = httpx.LocalOrganizationID
