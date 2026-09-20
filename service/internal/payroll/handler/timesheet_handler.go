package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
)

type TimesheetHandler struct {
	svc payroll.HRService
}

func NewTimesheetHandler(svc payroll.HRService) TimesheetHandler {
	return TimesheetHandler{svc: svc}
}

type TimesheetResponse struct {
	ID          uint64  `json:"id"`
	EmployeeID  uint64  `json:"employee_id"`
	Date        string  `json:"date"`
	ProjectID   *uint64 `json:"project_id"`
	TaskID      *uint64 `json:"task_id"`
	DimensionID *uint64 `json:"dimension_id"`
	Hours       float64 `json:"hours"`
	Description *string `json:"description"`
}

func newTimesheetResponse(timesheet *payroll.Timesheet) TimesheetResponse {
	return TimesheetResponse{
		ID:          timesheet.ID,
		EmployeeID:  timesheet.EmployeeID,
		Date:        timesheet.Date.Format("2006-01-02"),
		ProjectID:   timesheet.ProjectID,
		TaskID:      timesheet.TaskID,
		DimensionID: timesheet.DimensionID,
		Hours:       timesheet.Hours,
		Description: timesheet.Description,
	}
}

type CreateTimesheetRequest struct {
	EmployeeID  uint64  `json:"employee_id" validate:"required,gt=0"`
	Date        string  `json:"date" validate:"required"`
	ProjectID   *uint64 `json:"project_id"`
	TaskID      *uint64 `json:"task_id"`
	DimensionID *uint64 `json:"dimension_id"`
	Hours       float64 `json:"hours" validate:"required,gt=0"`
	Description *string `json:"description"`
}

type ListTimesheetsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []TimesheetResponse `json:"data"`
}

type TimesheetResponseEnvelope struct {
	httpx.EnvelopeBase
	Data TimesheetResponse `json:"data"`
}

var timesheetQueryAllowlist = map[string]struct{}{
	"employee_id":  {},
	"date":         {},
	"project_id":   {},
	"task_id":      {},
	"dimension_id": {},
}

// @Summary List timesheets
// @Description Lists timesheets with pagination, sorting, and filtering. Queries are restricted to allowlisted fields (employee, date, project, task, and dimension account), and invalid query parameters return a 422 response.
// @Tags Timesheets
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListTimesheetsResponseEnvelope "Timesheets retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/timesheets [get]
func (h TimesheetHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, timesheetQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	page, err := h.svc.ListTimesheets(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("timesheet list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve timesheets.", err)
	}
	items := make([]TimesheetResponse, len(page.Items))
	for i, timesheet := range page.Items {
		items[i] = newTimesheetResponse(timesheet)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Timesheets retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get timesheet
// @Description Gets a single timesheet by its id. The id must be a valid positive integer, and a 404 response is returned when no matching timesheet exists.
// @Tags Timesheets
// @Accept json
// @Produce json
// @Param id path integer true "Timesheet ID"
// @Success 200 {object} TimesheetResponseEnvelope "Timesheet retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Timesheet not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/timesheets/{id} [get]
func (h TimesheetHandler) Get(c fiber.Ctx) error {
	timesheetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid timesheet id provided.", nil)
	}
	timesheet, err := h.svc.FindTimesheet(c, timesheetID)
	if err != nil {
		httpx.RequestLog(c).Error("timesheet lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get timesheet.", err)
	}
	if timesheet == nil {
		return httpx.CreateNotFoundResponse(c, "Timesheet not found.")
	}
	return httpx.CreateSuccessResponse(c, "Timesheet retrieved successfully.", newTimesheetResponse(timesheet))
}

// @Summary Create timesheet
// @Description Creates a new timesheet entry recording hours worked for an existing employee, optionally attributed to a project, task, or dimension account. The date must be in YYYY-MM-DD format, hours must be positive, and any referenced dimension account must exist.
// @Tags Timesheets
// @Accept json
// @Produce json
// @Param request body CreateTimesheetRequest true "Timesheet details"
// @Success 201 {object} TimesheetResponseEnvelope "Timesheet created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/timesheets [post]
func (h TimesheetHandler) Create(c fiber.Ctx) error {
	var request CreateTimesheetRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(&request.Date)
	if err != nil || date == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	timesheet, err := h.svc.CreateTimesheet(c, &payroll.Timesheet{
		EmployeeID:  request.EmployeeID,
		Date:        *date,
		ProjectID:   request.ProjectID,
		TaskID:      request.TaskID,
		DimensionID: request.DimensionID,
		Hours:       request.Hours,
		Description: request.Description,
	})
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Timesheet created successfully.", newTimesheetResponse(timesheet))
}

// @Summary Delete timesheet
// @Description Deletes a timesheet by id. The id must be a valid positive integer, and a successful deletion returns no content.
// @Tags Timesheets
// @Accept json
// @Produce json
// @Param id path integer true "Timesheet ID"
// @Success 204 {object} httpx.EmptyEnvelope "Timesheet deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/timesheets/{id} [delete]
func (h TimesheetHandler) Delete(c fiber.Ctx) error {
	timesheetID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid timesheet id provided.", nil)
	}
	if err := h.svc.DeleteTimesheet(c, timesheetID); err != nil {
		httpx.RequestLog(c).Error("timesheet delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete timesheet.", err)
	}
	return httpx.CreateNoContentResponse(c)
}
