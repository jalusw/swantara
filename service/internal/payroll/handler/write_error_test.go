package handler

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
)

func TestWriteHRError_MapsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "employee not found", err: payroll.ErrEmployeeNotFound},
		{name: "contract not found", err: payroll.ErrContractNotFound},
		{name: "leave request not found", err: payroll.ErrLeaveRequestNotFound},
		{name: "attendance not found", err: payroll.ErrAttendanceNotFound},
		{name: "timesheet not found", err: payroll.ErrTimesheetNotFound},
		{name: "leave type not found", err: payroll.ErrLeaveTypeNotFound},
		{name: "leave employee", err: payroll.ErrLeaveEmployee},
		{name: "employee user not found", err: payroll.ErrEmployeeUserNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeHRError(c, tt.err)
			})
			if got != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", got)
			}
		})
	}
}

func TestWriteHRError_MapsUnprocessable(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "employee number taken", err: payroll.ErrEmployeeNumberTaken},
		{name: "employee user taken", err: payroll.ErrEmployeeUserTaken},
		{name: "employee department", err: payroll.ErrEmployeeDepartment},
		{name: "employee position", err: payroll.ErrEmployeePosition},
		{name: "employee manager", err: payroll.ErrEmployeeManager},
		{name: "employee contact", err: payroll.ErrEmployeeContact},
		{name: "employee employment type", err: payroll.ErrEmployeeEmploymentType},
		{name: "employee organization", err: payroll.ErrEmployeeOrganization},
		{name: "employee number", err: payroll.ErrEmployeeNumber},
		{name: "contract employee", err: payroll.ErrContractEmployee},
		{name: "contract wage", err: payroll.ErrContractWage},
		{name: "contract state", err: payroll.ErrContractState},
		{name: "leave state", err: payroll.ErrLeaveState},
		{name: "leave date range", err: payroll.ErrLeaveDateRange},
		{name: "leave days", err: payroll.ErrLeaveDays},
		{name: "leave balance", err: payroll.ErrLeaveBalance},
		{name: "attendance check in", err: payroll.ErrAttendanceCheckIn},
		{name: "attendance check out", err: payroll.ErrAttendanceCheckOut},
		{name: "attendance employee", err: payroll.ErrAttendanceEmployee},
		{name: "attendance inactive", err: payroll.ErrAttendanceInactive},
		{name: "attendance open exists", err: payroll.ErrAttendanceOpenExists},
		{name: "attendance already closed", err: payroll.ErrAttendanceAlreadyClosed},
		{name: "attendance state", err: payroll.ErrAttendanceState},
		{name: "attendance on leave", err: payroll.ErrAttendanceOnLeave},
		{name: "timesheet employee", err: payroll.ErrTimesheetEmployee},
		{name: "timesheet hours", err: payroll.ErrTimesheetHours},
		{name: "timesheet dimension", err: payroll.ErrTimesheetDimension},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writeHRError(c, tt.err)
			})
			if got != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", got)
			}
		})
	}
}

func TestWriteHRError_DefaultsToServerError(t *testing.T) {
	got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
		return writeHRError(c, errors.New("boom"))
	})
	if got != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", got)
	}
}

func TestWritePayrollError_MapsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "run not found", err: payroll.ErrRunNotFound},
		{name: "payslip not found", err: payroll.ErrPayslipNotFound},
		{name: "rule not found", err: payroll.ErrRuleNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writePayrollError(c, tt.err)
			})
			if got != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", got)
			}
		})
	}
}

func TestWritePayrollError_MapsUnprocessable(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "run state", err: payroll.ErrRunState},
		{name: "run period", err: payroll.ErrRunPeriod},
		{name: "run organization", err: payroll.ErrRunOrganization},
		{name: "run no payslips", err: payroll.ErrRunNoPayslips},
		{name: "rule code", err: payroll.ErrRuleCode},
		{name: "rule category", err: payroll.ErrRuleCategory},
		{name: "rule compute type", err: payroll.ErrRuleComputeType},
		{name: "rule accounts", err: payroll.ErrRuleAccounts},
		{name: "rule account", err: payroll.ErrRuleAccount},
		{name: "no bank account", err: payroll.ErrNoBankAccount},
		{name: "no net payable", err: payroll.ErrNoNetPayable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
				return writePayrollError(c, tt.err)
			})
			if got != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", got)
			}
		})
	}
}

func TestWritePayrollError_MapsRunSequence(t *testing.T) {
	got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
		return writePayrollError(c, payroll.ErrRunSequence)
	})
	if got != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", got)
	}
}

func TestWritePayrollError_DefaultsToServerError(t *testing.T) {
	got := writeErrorStatus("/write-error", func(c fiber.Ctx) error {
		return writePayrollError(c, errors.New("boom"))
	})
	if got != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", got)
	}
}
