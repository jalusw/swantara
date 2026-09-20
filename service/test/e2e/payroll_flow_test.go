//go:build e2e

package e2e

import (
	"net/http"
	"testing"
)

func TestPayrollHireToPayLifecycleE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	departmentID := createDepartment(t, e, f)
	employeeID := createEmployee(t, e, f, departmentID)

	contract := f.authed(t, http.MethodPost, "/contracts").
		WithJSON(map[string]any{
			"employee_id": employeeID,
			"wage":        8000000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("contract").
		Object()

	_ = uint64(contract.Value("id").Number().Raw())

	leaveTypeID := createLeaveType(t, e, f)

	leave := f.authed(t, http.MethodPost, "/leave-requests").
		WithJSON(map[string]any{
			"employee_id":   employeeID,
			"leave_type_id": leaveTypeID,
			"date_from":     "2026-06-01",
			"date_to":       "2026-06-02",
			"days":          2,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("leave_request").
		Object()

	leaveID := uint64(leave.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/leave-requests/"+itoa(leaveID)+"/submit").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/leave-requests/"+itoa(leaveID)+"/approve").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodPost, "/timesheets").
		WithJSON(map[string]any{
			"employee_id": employeeID,
			"date":        "2026-06-01",
			"hours":       8,
		}).
		Expect().
		Status(http.StatusCreated)

	run := f.authed(t, http.MethodPost, "/payroll-runs").
		WithJSON(map[string]any{
			"organization_id": f.organizationID(),
			"period_start":    "2026-06-01",
			"period_end":      "2026-06-30",
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("payroll_run").
		Object()

	runID := uint64(run.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/payroll-runs/"+itoa(runID)+"/confirm").
		Expect().
		Status(http.StatusOK)

	f.authed(t, http.MethodGet, "/payslips").
		WithQuery("payroll_run_id", itoa(runID)).
		Expect().
		Status(http.StatusOK)
}

func TestAttendanceCheckInOutE2E(t *testing.T) {
	f := newFlowFixture(t)
	e := newExpect(t)

	employeeID := createEmployee(t, e, f, 0)

	checkIn := f.authed(t, http.MethodPost, "/attendances/check-in").
		WithJSON(map[string]any{"employee_id": employeeID}).
		Expect().
		Status(http.StatusCreated).
		JSON().
		Object().
		Value("data").
		Object().
		Value("attendance").
		Object()

	attendanceID := uint64(checkIn.Value("id").Number().Raw())

	f.authed(t, http.MethodPost, "/attendances/"+itoa(attendanceID)+"/check-out").
		Expect().
		Status(http.StatusOK)
}
