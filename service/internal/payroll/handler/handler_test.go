package handler

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/iam"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/dao"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

func noTenantGuards() httpx.RouteGuards {
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

func writeErrorStatus(path string, fn func(c fiber.Ctx) error) int {
	app := fiber.New()
	app.Post(path, func(c fiber.Ctx) error {
		return fn(c)
	})
	resp, err := doRequest(app, http.MethodPost, path, "")
	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func timeNow() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}

func hrServiceFor(depts *dao.CRUDMock[reference.Department], positions *dao.CRUDMock[reference.JobPosition], leaveTypes *dao.CRUDMock[reference.LeaveType]) payroll.HRService {
	var d dao.CRUDMock[reference.Department]
	if depts != nil {
		d = *depts
	}
	var p dao.CRUDMock[reference.JobPosition]
	if positions != nil {
		p = *positions
	}
	var lt dao.CRUDMock[reference.LeaveType]
	if leaveTypes != nil {
		lt = *leaveTypes
	}
	return payroll.NewHRService(
		payroll.EmployeeDAOMock{},
		payroll.EmploymentContractDAOMock{},
		payroll.LeaveRequestDAOMock{},
		lt,
		d,
		p,
		dao.CRUDMock[reference.Organization]{},
		contacts.ContactDAOMock{},
		dao.CRUDMock[reference.Dimension]{},
		iam.UserDAOMock{},
		payroll.AttendanceDAOMock{},
		payroll.TimesheetDAOMock{},
		payroll.ShiftDAOMock{},
		payroll.ShiftAssignmentDAOMock{},
		inventory.TransactionerMock{},
	)
}
