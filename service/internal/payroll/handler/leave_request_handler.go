package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
)

type LeaveRequestHandler struct {
	svc payroll.HRService
}

func NewLeaveRequestHandler(svc payroll.HRService) LeaveRequestHandler {
	return LeaveRequestHandler{svc: svc}
}

type LeaveRequestResponse struct {
	ID          uint64  `json:"id"`
	EmployeeID  uint64  `json:"employee_id"`
	LeaveTypeID uint64  `json:"leave_type_id"`
	DateFrom    string  `json:"date_from"`
	DateTo      string  `json:"date_to"`
	Days        float64 `json:"days"`
	State       string  `json:"state"`
}

func newLeaveRequestResponse(leave *payroll.LeaveRequest) LeaveRequestResponse {
	return LeaveRequestResponse{
		ID:          leave.ID,
		EmployeeID:  leave.EmployeeID,
		LeaveTypeID: leave.LeaveTypeID,
		DateFrom:    leave.DateFrom.Format("2006-01-02"),
		DateTo:      leave.DateTo.Format("2006-01-02"),
		Days:        leave.Days,
		State:       leave.State,
	}
}

type CreateLeaveRequestRequest struct {
	EmployeeID  uint64  `json:"employee_id" validate:"required,gt=0"`
	LeaveTypeID uint64  `json:"leave_type_id" validate:"required,gt=0"`
	DateFrom    string  `json:"date_from" validate:"required"`
	DateTo      string  `json:"date_to" validate:"required"`
	Days        float64 `json:"days" validate:"required,gt=0"`
}

type ListLeaveRequestsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []LeaveRequestResponse `json:"data"`
}

type LeaveRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data LeaveRequestResponse `json:"data"`
}

type LeaveBalanceResponseEnvelope struct {
	httpx.EnvelopeBase
	Data LeaveBalanceResponse `json:"data"`
}

var leaveRequestQueryAllowlist = map[string]struct{}{
	"employee_id":   {},
	"leave_type_id": {},
	"state":         {},
}

// @Summary List leave requests
// @Description Lists leave requests with pagination, sorting, and filtering. Queries are restricted to allowlisted fields (employee, leave type, and state), and invalid query parameters return a 422 response.
// @Tags Leave Requests
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListLeaveRequestsResponseEnvelope "Leave requests retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-requests [get]
func (h LeaveRequestHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, leaveRequestQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	page, err := h.svc.ListLeaves(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("leave request list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve leave requests.", err)
	}
	items := make([]LeaveRequestResponse, len(page.Items))
	for i, leave := range page.Items {
		items[i] = newLeaveRequestResponse(leave)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Leave requests retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get leave request
// @Description Gets a single leave request by its id. The id must be a valid positive integer, and a 404 response is returned when no matching leave request exists.
// @Tags Leave Requests
// @Accept json
// @Produce json
// @Param id path integer true "Leave request ID"
// @Success 200 {object} LeaveRequestResponseEnvelope "Leave request retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Leave request not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-requests/{id} [get]
func (h LeaveRequestHandler) Get(c fiber.Ctx) error {
	leaveID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave request id provided.", nil)
	}
	leave, err := h.svc.FindLeave(c, leaveID)
	if err != nil {
		httpx.RequestLog(c).Error("leave request lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get leave request.", err)
	}
	if leave == nil {
		return httpx.CreateNotFoundResponse(c, "Leave request not found.")
	}
	return httpx.CreateSuccessResponse(c, "Leave request retrieved successfully.", newLeaveRequestResponse(leave))
}

// @Summary Create leave request
// @Description Creates a new leave request in draft state for an existing employee and leave type. The day count must be positive, the date range must be valid (date_to cannot precede date_from), dates must be in YYYY-MM-DD format, and both the employee and leave type must exist.
// @Tags Leave Requests
// @Accept json
// @Produce json
// @Param request body CreateLeaveRequestRequest true "Leave request details"
// @Success 201 {object} LeaveRequestResponseEnvelope "Leave request created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-requests [post]
func (h LeaveRequestHandler) Create(c fiber.Ctx) error {
	var request CreateLeaveRequestRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	dateFrom, err := helper.ParseDate(&request.DateFrom)
	if err != nil || dateFrom == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date from must be in YYYY-MM-DD format.", nil)
	}
	dateTo, err := helper.ParseDate(&request.DateTo)
	if err != nil || dateTo == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date to must be in YYYY-MM-DD format.", nil)
	}
	leave, err := h.svc.CreateLeaveRequest(c, &payroll.LeaveRequest{
		EmployeeID:  request.EmployeeID,
		LeaveTypeID: request.LeaveTypeID,
		DateFrom:    *dateFrom,
		DateTo:      *dateTo,
		Days:        request.Days,
	})
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Leave request created successfully.", newLeaveRequestResponse(leave))
}

// @Summary Submit leave request
// @Description Submits a leave request for approval, moving it from draft to submitted state. Only draft requests can be submitted, and submitting a request in any other state returns a 422 response.
// @Tags Leave Requests
// @Accept json
// @Produce json
// @Param id path integer true "Leave request ID"
// @Success 200 {object} LeaveRequestResponseEnvelope "Leave request submitted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Leave request not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-requests/{id}/submit [post]
func (h LeaveRequestHandler) Submit(c fiber.Ctx) error {
	leaveID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave request id provided.", nil)
	}
	leave, err := h.svc.SubmitLeave(c, leaveID)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Leave request submitted successfully.", newLeaveRequestResponse(leave))
}

// @Summary Approve leave request
// @Description Approves a submitted leave request, moving it to approved state. The request must be in submitted state and the employee's remaining leave balance must cover the requested days; an insufficient balance returns a 422 response.
// @Tags Leave Requests
// @Accept json
// @Produce json
// @Param id path integer true "Leave request ID"
// @Success 200 {object} LeaveRequestResponseEnvelope "Leave request approved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Leave request not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-requests/{id}/approve [post]
func (h LeaveRequestHandler) Approve(c fiber.Ctx) error {
	leaveID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave request id provided.", nil)
	}
	leave, err := h.svc.ApproveLeave(c, leaveID)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Leave request approved successfully.", newLeaveRequestResponse(leave))
}

// @Summary Refuse leave request
// @Description Refuses a submitted leave request, moving it to refused state. Only requests in submitted state can be refused, and refusing a request in any other state returns a 422 response.
// @Tags Leave Requests
// @Accept json
// @Produce json
// @Param id path integer true "Leave request ID"
// @Success 200 {object} LeaveRequestResponseEnvelope "Leave request refused successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Leave request not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-requests/{id}/refuse [post]
func (h LeaveRequestHandler) Refuse(c fiber.Ctx) error {
	leaveID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave request id provided.", nil)
	}
	leave, err := h.svc.RefuseLeave(c, leaveID)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Leave request refused successfully.", newLeaveRequestResponse(leave))
}

type LeaveBalanceResponse struct {
	EmployeeID  uint64  `json:"employee_id"`
	LeaveTypeID uint64  `json:"leave_type_id"`
	Balance     float64 `json:"balance"`
}

// @Summary Get leave balance
// @Description Gets the remaining leave balance for an employee and leave type. The balance is computed as the leave type's annual day allocation minus the days of all approved requests, floored at zero, and a 404 response is returned when the leave type does not exist.
// @Tags Leave Requests
// @Accept json
// @Produce json
// @Param employee_id path integer true "Employee ID"
// @Param leave_type_id path integer true "Leave type ID"
// @Success 200 {object} LeaveBalanceResponseEnvelope "Leave balance retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-requests/balance/{employee_id}/{leave_type_id} [get]
func (h LeaveRequestHandler) Balance(c fiber.Ctx) error {
	employeeID, err := strconv.ParseUint(c.Params("employee_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid employee id provided.", nil)
	}
	leaveTypeID, err := strconv.ParseUint(c.Params("leave_type_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave type id provided.", nil)
	}
	balance, err := h.svc.LeaveBalance(c, employeeID, leaveTypeID)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Leave balance retrieved successfully.", LeaveBalanceResponse{
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Balance:     balance,
	})
}
