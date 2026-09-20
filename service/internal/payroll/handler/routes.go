package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h DepartmentHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	departments := api.Group("/departments", guards.AuthN)

	departments.Get("/", guards.Guard("department", "view"), h.List)
	departments.Post("/", guards.Guard("department", "create"), httpx.IdempotencyGuard(guards.Idempotency, "department-create"), h.Create)
	departments.Get("/:id", guards.Guard("department", "view"), h.Get)
	departments.Put("/:id", guards.Guard("department", "update"), httpx.IdempotencyGuard(guards.Idempotency, "department-update"), h.Update)
	departments.Delete("/:id", guards.Guard("department", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "department-delete"), h.Delete)
}

func (h JobPositionHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	positions := api.Group("/job-positions", guards.AuthN)

	positions.Get("/", guards.Guard("job_position", "view"), h.List)
	positions.Post("/", guards.Guard("job_position", "create"), httpx.IdempotencyGuard(guards.Idempotency, "job-position-create"), h.Create)
	positions.Get("/:id", guards.Guard("job_position", "view"), h.Get)
	positions.Put("/:id", guards.Guard("job_position", "update"), httpx.IdempotencyGuard(guards.Idempotency, "job-position-update"), h.Update)
	positions.Delete("/:id", guards.Guard("job_position", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "job-position-delete"), h.Delete)
}

func (h LeaveTypeHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	leaveTypes := api.Group("/leave-types", guards.AuthN)

	leaveTypes.Get("/", guards.Guard("leave_request", "view"), h.List)
	leaveTypes.Post("/", guards.Guard("leave_request", "create"), httpx.IdempotencyGuard(guards.Idempotency, "leave-type-create"), h.Create)
	leaveTypes.Get("/:id", guards.Guard("leave_request", "view"), h.Get)
	leaveTypes.Put("/:id", guards.Guard("leave_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "leave-type-update"), h.Update)
	leaveTypes.Delete("/:id", guards.Guard("leave_request", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "leave-type-delete"), h.Delete)
}

func (h EmployeeHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	employees := api.Group("/employees", guards.AuthN)

	employees.Get("/", guards.Guard("employee", "view"), h.List)
	employees.Post("/", guards.Guard("employee", "create"), httpx.IdempotencyGuard(guards.Idempotency, "employee-create"), h.Create)
	employees.Get("/:id", guards.Guard("employee", "view"), h.Get)
	employees.Put("/:id", guards.Guard("employee", "update"), httpx.IdempotencyGuard(guards.Idempotency, "employee-update"), h.Update)
	employees.Delete("/:id", guards.Guard("employee", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "employee-delete"), h.Delete)
}

func (h ContractHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	contracts := api.Group("/contracts", guards.AuthN)

	contracts.Get("/", guards.Guard("employment_contract", "view"), h.List)
	contracts.Post("/", guards.Guard("employment_contract", "create"), httpx.IdempotencyGuard(guards.Idempotency, "contract-create"), h.Create)
	contracts.Get("/:id", guards.Guard("employment_contract", "view"), h.Get)
	contracts.Put("/:id", guards.Guard("employment_contract", "update"), httpx.IdempotencyGuard(guards.Idempotency, "contract-update"), h.Update)
	contracts.Post("/:id/terminate", guards.Guard("employment_contract", "update"), httpx.IdempotencyGuard(guards.Idempotency, "contract-terminate"), h.Terminate)
}

func (h LeaveRequestHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	leaves := api.Group("/leave-requests", guards.AuthN)

	leaves.Get("/", guards.Guard("leave_request", "view"), h.List)
	leaves.Post("/", guards.Guard("leave_request", "create"), httpx.IdempotencyGuard(guards.Idempotency, "leave-request-create"), h.Create)
	leaves.Get("/:id", guards.Guard("leave_request", "view"), h.Get)
	leaves.Post("/:id/submit", guards.Guard("leave_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "leave-request-submit"), h.Submit)
	leaves.Post("/:id/approve", guards.Guard("leave_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "leave-request-approve"), h.Approve)
	leaves.Post("/:id/refuse", guards.Guard("leave_request", "update"), httpx.IdempotencyGuard(guards.Idempotency, "leave-request-refuse"), h.Refuse)
	leaves.Get("/balance/:employee_id/:leave_type_id", guards.Guard("leave_request", "view"), h.Balance)
}

func (h AttendanceHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	attendances := api.Group("/attendances", guards.AuthN)

	attendances.Get("/", guards.Guard("attendance", "view"), h.List)
	attendances.Get("/missing", guards.Guard("attendance", "view"), h.ListMissing)
	attendances.Post("/check-in", guards.Guard("attendance", "create"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-check-in"), h.CheckIn)
	attendances.Post("/bulk-check-out", guards.Guard("attendance", "update"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-bulk-check-out"), h.BulkCheckOut)
	attendances.Post("/auto-checkout", guards.Guard("attendance", "update"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-auto-checkout"), h.AutoCheckout)
	attendances.Post("/mark-absent", guards.Guard("attendance", "create"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-mark-absent"), h.MarkAbsent)
	attendances.Get("/:id", guards.Guard("attendance", "view"), h.Get)
	attendances.Put("/:id", guards.Guard("attendance", "update"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-update"), h.Update)
	attendances.Post("/:id/check-out", guards.Guard("attendance", "update"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-check-out"), h.CheckOut)
	attendances.Post("/:id/approve", guards.Guard("attendance", "update"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-approve"), h.Approve)
	attendances.Post("/:id/reject", guards.Guard("attendance", "update"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-reject"), h.Reject)
	attendances.Delete("/:id", guards.Guard("attendance", "update"), httpx.IdempotencyGuard(guards.Idempotency, "attendance-delete"), h.Delete)
}

func (h TimesheetHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	timesheets := api.Group("/timesheets", guards.AuthN)

	timesheets.Get("/", guards.Guard("timesheet", "view"), h.List)
	timesheets.Post("/", guards.Guard("timesheet", "create"), httpx.IdempotencyGuard(guards.Idempotency, "timesheet-create"), h.Create)
	timesheets.Get("/:id", guards.Guard("timesheet", "view"), h.Get)
	timesheets.Delete("/:id", guards.Guard("timesheet", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "timesheet-delete"), h.Delete)
}

func (h SalaryRuleHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	rules := api.Group("/salary-rules", guards.AuthN)

	rules.Get("/", guards.Guard("salary_rule", "view"), h.List)
	rules.Post("/", guards.Guard("salary_rule", "create"), httpx.IdempotencyGuard(guards.Idempotency, "salary-rule-create"), h.Create)
	rules.Get("/:id", guards.Guard("salary_rule", "view"), h.Get)
	rules.Put("/:id", guards.Guard("salary_rule", "update"), httpx.IdempotencyGuard(guards.Idempotency, "salary-rule-update"), h.Update)
	rules.Delete("/:id", guards.Guard("salary_rule", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "salary-rule-delete"), h.Delete)
}

func (h PayrollRunHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	runs := api.Group("/payroll-runs", guards.AuthN)

	runs.Get("/", guards.Guard("payroll_run", "view"), h.List)
	runs.Post("/", guards.Guard("payroll_run", "create"), httpx.IdempotencyGuard(guards.Idempotency, "payroll-run-create"), h.Create)
	runs.Get("/:id", guards.Guard("payroll_run", "view"), h.Get)
	runs.Post("/:id/confirm", guards.Guard("payroll_run", "update"), httpx.IdempotencyGuard(guards.Idempotency, "payroll-run-confirm"), h.Confirm)
	runs.Post("/:id/pay", guards.Guard("payroll_run", "update"), httpx.IdempotencyGuard(guards.Idempotency, "payroll-run-pay"), h.Pay)
	runs.Post("/:id/close", guards.Guard("payroll_run", "update"), httpx.IdempotencyGuard(guards.Idempotency, "payroll-run-close"), h.Close)
}

func (h PayslipHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	payslips := api.Group("/payslips", guards.AuthN)

	payslips.Get("/", guards.Guard("payslip", "view"), h.List)
	payslips.Get("/:id", guards.Guard("payslip", "view"), h.Get)
}

func (h ShiftHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	shifts := api.Group("/shifts", guards.AuthN)

	shifts.Get("/", guards.Guard("shift", "view"), h.List)
	shifts.Post("/", guards.Guard("shift", "create"), httpx.IdempotencyGuard(guards.Idempotency, "shift-create"), h.Create)
	shifts.Put("/:id", guards.Guard("shift", "update"), httpx.IdempotencyGuard(guards.Idempotency, "shift-update"), h.Update)
	shifts.Delete("/:id", guards.Guard("shift", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "shift-delete"), h.Delete)
}

func (h ShiftAssignmentHandler) Register(api fiber.Router, guards httpx.RouteGuards) {
	assignments := api.Group("/shift-assignments", guards.AuthN)

	assignments.Post("/", guards.Guard("shift_assignment", "create"), httpx.IdempotencyGuard(guards.Idempotency, "shift-assignment-create"), h.Create)
	assignments.Delete("/:id", guards.Guard("shift_assignment", "delete"), httpx.IdempotencyGuard(guards.Idempotency, "shift-assignment-delete"), h.Delete)
}
