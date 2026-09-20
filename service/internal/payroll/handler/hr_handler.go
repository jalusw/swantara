package handler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
)

func writeHRError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, payroll.ErrEmployeeNotFound), errors.Is(err, payroll.ErrContractNotFound),
		errors.Is(err, payroll.ErrLeaveRequestNotFound), errors.Is(err, payroll.ErrAttendanceNotFound),
		errors.Is(err, payroll.ErrTimesheetNotFound), errors.Is(err, payroll.ErrLeaveTypeNotFound),
		errors.Is(err, payroll.ErrLeaveEmployee), errors.Is(err, payroll.ErrEmployeeUserNotFound),
		errors.Is(err, payroll.ErrShiftNotFound), errors.Is(err, payroll.ErrShiftAssignmentNotFound):
		return httpx.CreateNotFoundResponse(c, "Resource not found.")
	case errors.Is(err, payroll.ErrEmployeeNumberTaken), errors.Is(err, payroll.ErrEmployeeUserTaken),
		errors.Is(err, payroll.ErrEmployeeDepartment), errors.Is(err, payroll.ErrEmployeePosition),
		errors.Is(err, payroll.ErrEmployeeManager), errors.Is(err, payroll.ErrEmployeeContact),
		errors.Is(err, payroll.ErrEmployeeEmploymentType), errors.Is(err, payroll.ErrEmployeeOrganization),
		errors.Is(err, payroll.ErrEmployeeNumber), errors.Is(err, payroll.ErrContractEmployee),
		errors.Is(err, payroll.ErrContractWage), errors.Is(err, payroll.ErrContractState),
		errors.Is(err, payroll.ErrLeaveState), errors.Is(err, payroll.ErrLeaveDateRange),
		errors.Is(err, payroll.ErrLeaveDays), errors.Is(err, payroll.ErrLeaveBalance),
		errors.Is(err, payroll.ErrAttendanceCheckIn), errors.Is(err, payroll.ErrAttendanceCheckOut),
		errors.Is(err, payroll.ErrAttendanceEmployee), errors.Is(err, payroll.ErrTimesheetEmployee),
		errors.Is(err, payroll.ErrTimesheetHours), errors.Is(err, payroll.ErrTimesheetDimension),
		errors.Is(err, payroll.ErrAttendanceInactive), errors.Is(err, payroll.ErrAttendanceOpenExists),
		errors.Is(err, payroll.ErrAttendanceAlreadyClosed), errors.Is(err, payroll.ErrAttendanceState),
		errors.Is(err, payroll.ErrAttendanceOnLeave), errors.Is(err, payroll.ErrAttendanceNotPending),
		errors.Is(err, payroll.ErrShiftName), errors.Is(err, payroll.ErrShiftTime),
		errors.Is(err, payroll.ErrShiftOrganization), errors.Is(err, payroll.ErrShiftAssignmentExists),
		errors.Is(err, payroll.ErrShiftAssignmentEmployee), errors.Is(err, payroll.ErrShiftAssignmentShift),
		errors.Is(err, payroll.ErrAutoCheckoutHour), errors.Is(err, payroll.ErrAutoCheckoutNotConfigured):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Request cannot be processed.", nil)
	default:
		httpx.RequestLog(c).Error("hr write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process request.", err)
	}
}
