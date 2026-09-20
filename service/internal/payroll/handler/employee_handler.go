package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/contacts"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
)

type EmployeeHandler struct {
	svc payroll.HRService
}

func NewEmployeeHandler(svc payroll.HRService) EmployeeHandler {
	return EmployeeHandler{svc: svc}
}

type EmployeeResponse struct {
	ID               uint64  `json:"id"`
	OrganizationID   *uint64 `json:"organization_id"`
	ContactID        uint64  `json:"contact_id"`
	UserID           *uint64 `json:"user_id"`
	EmployeeNumber   string  `json:"employee_number"`
	DepartmentID     *uint64 `json:"department_id"`
	JobPositionID    *uint64 `json:"job_position_id"`
	ManagerID        *uint64 `json:"manager_id"`
	HireDate         *string `json:"hire_date"`
	TerminationDate  *string `json:"termination_date"`
	EmploymentType   string  `json:"employment_type"`
	WorkLocation     *string `json:"work_location"`
	Active           bool    `json:"active"`
	RequiresApproval bool    `json:"requires_approval"`
}

func newEmployeeResponse(employee *payroll.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:               employee.ID,
		OrganizationID:   employee.OrganizationID,
		ContactID:        employee.ContactID,
		UserID:           employee.UserID,
		EmployeeNumber:   employee.EmployeeNumber,
		DepartmentID:     employee.DepartmentID,
		JobPositionID:    employee.JobPositionID,
		ManagerID:        employee.ManagerID,
		HireDate:         helper.FormatDatePtr(employee.HireDate),
		TerminationDate:  helper.FormatDatePtr(employee.TerminationDate),
		EmploymentType:   employee.EmploymentType,
		WorkLocation:     employee.WorkLocation,
		Active:           employee.Active,
		RequiresApproval: employee.RequiresApproval,
	}
}

type CreateEmployeeRequest struct {
	OrganizationID   *uint64 `json:"organization_id"`
	Name             string  `json:"name" validate:"required"`
	Email            *string `json:"email"`
	Phone            *string `json:"phone"`
	UserID           *uint64 `json:"user_id"`
	EmployeeNumber   string  `json:"employee_number" validate:"required"`
	DepartmentID     *uint64 `json:"department_id"`
	JobPositionID    *uint64 `json:"job_position_id"`
	ManagerID        *uint64 `json:"manager_id"`
	HireDate         *string `json:"hire_date"`
	EmploymentType   string  `json:"employment_type"`
	WorkLocation     *string `json:"work_location"`
	Wage             float64 `json:"wage" validate:"required,gt=0"`
	WageType         string  `json:"wage_type"`
	CurrencyCode     string  `json:"currency_code"`
	RequiresApproval bool    `json:"requires_approval"`
}

type ListEmployeesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []EmployeeResponse `json:"data"`
}

type EmployeeResponseEnvelope struct {
	httpx.EnvelopeBase
	Data EmployeeResponse `json:"data"`
}

var employeeQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"employee_number": {},
	"user_id":         {},
	"department_id":   {},
	"job_position_id": {},
	"manager_id":      {},
	"employment_type": {},
	"active":          {},
}

// @Summary List employees
// @Description Lists employees with pagination, sorting, and filtering, scoped to the caller's organization. Queries are restricted to an allowlisted set of fields, and the organization filter is forced to the caller's tenant so cross-tenant employees are never returned.
// @Tags Employees
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListEmployeesResponseEnvelope "Employees retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/employees [get]
func (h EmployeeHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, employeeQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListEmployees(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("employee list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve employees.", err)
	}
	items := make([]EmployeeResponse, len(page.Items))
	for i, employee := range page.Items {
		items[i] = newEmployeeResponse(employee)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Employees retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get employee
// @Description Gets a single employee by its id. The employee must belong to the caller's organization, and a 404 not found is returned when it does not, even if an employee with that id exists in another tenant.
// @Tags Employees
// @Accept json
// @Produce json
// @Param id path integer true "Employee ID"
// @Success 200 {object} EmployeeResponseEnvelope "Employee retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Employee not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/employees/{id} [get]
func (h EmployeeHandler) Get(c fiber.Ctx) error {
	employeeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid employee id provided.", nil)
	}
	employee, err := h.svc.FindEmployee(c, employeeID)
	if err != nil {
		httpx.RequestLog(c).Error("employee lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get employee.", err)
	}
	if employee == nil || !httpx.OwnsTenant(c, employee.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Employee not found.")
	}
	return httpx.CreateSuccessResponse(c, "Employee retrieved successfully.", newEmployeeResponse(employee))
}

// @Summary Create employee
// @Description Creates a new employee together with a linked contact profile and an initial active employment contract within a single transaction. The employee name, employee number, and a positive wage are required, the hire date must be in YYYY-MM-DD format, and the employee number must be unique; default currency (IDR) and monthly wage type are applied when not supplied.
// @Tags Employees
// @Accept json
// @Produce json
// @Param request body CreateEmployeeRequest true "Employee details"
// @Success 201 {object} EmployeeResponseEnvelope "Employee created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Resource not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/employees [post]
func (h EmployeeHandler) Create(c fiber.Ctx) error {
	var request CreateEmployeeRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	hireDate, err := helper.ParseDate(request.HireDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Hire date must be in YYYY-MM-DD format.", nil)
	}
	employee, err := h.svc.CreateEmployee(c, payroll.CreateEmployeeRequest{
		Employee: &payroll.Employee{
			OrganizationID:   organizationID,
			UserID:           request.UserID,
			EmployeeNumber:   request.EmployeeNumber,
			DepartmentID:     request.DepartmentID,
			JobPositionID:    request.JobPositionID,
			ManagerID:        request.ManagerID,
			HireDate:         hireDate,
			EmploymentType:   request.EmploymentType,
			WorkLocation:     request.WorkLocation,
			Active:           true,
			RequiresApproval: request.RequiresApproval,
		},
		Contract: &payroll.EmploymentContract{
			Wage:         request.Wage,
			WageType:     request.WageType,
			CurrencyCode: request.CurrencyCode,
		},
		Contact: &contacts.Contact{
			OrganizationID: organizationID,
			Name:           request.Name,
			Email:          request.Email,
			Phone:          request.Phone,
			Active:         true,
		},
	})
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Employee created successfully.", newEmployeeResponse(employee))
}

type UpdateEmployeeRequest struct {
	Name             string  `json:"name"`
	EmployeeNumber   string  `json:"employee_number" validate:"required"`
	UserID           *uint64 `json:"user_id"`
	DepartmentID     *uint64 `json:"department_id"`
	JobPositionID    *uint64 `json:"job_position_id"`
	ManagerID        *uint64 `json:"manager_id"`
	HireDate         *string `json:"hire_date"`
	TerminationDate  *string `json:"termination_date"`
	EmploymentType   string  `json:"employment_type"`
	WorkLocation     *string `json:"work_location"`
	Active           bool    `json:"active"`
	RequiresApproval bool    `json:"requires_approval"`
}

// @Summary Update employee
// @Description Updates an existing employee's employment details, including employee number, manager, department, job position, dates, and active flag. The employee must belong to the caller's organization (404 otherwise), hire and termination dates must be in YYYY-MM-DD format, and the employee number must remain unique.
// @Tags Employees
// @Accept json
// @Produce json
// @Param id path integer true "Employee ID"
// @Param request body UpdateEmployeeRequest true "Employee details"
// @Success 200 {object} EmployeeResponseEnvelope "Employee updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Employee not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/employees/{id} [put]
func (h EmployeeHandler) Update(c fiber.Ctx) error {
	employeeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid employee id provided.", nil)
	}
	employee, err := h.svc.FindEmployee(c, employeeID)
	if err != nil {
		httpx.RequestLog(c).Error("employee lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get employee.", err)
	}
	if employee == nil || !httpx.OwnsTenant(c, employee.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Employee not found.")
	}
	var request UpdateEmployeeRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	hireDate, err := helper.ParseDate(request.HireDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Hire date must be in YYYY-MM-DD format.", nil)
	}
	terminationDate, err := helper.ParseDate(request.TerminationDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Termination date must be in YYYY-MM-DD format.", nil)
	}
	employee.EmployeeNumber = request.EmployeeNumber
	employee.UserID = request.UserID
	employee.DepartmentID = request.DepartmentID
	employee.JobPositionID = request.JobPositionID
	employee.ManagerID = request.ManagerID
	employee.HireDate = hireDate
	employee.TerminationDate = terminationDate
	employee.EmploymentType = request.EmploymentType
	employee.WorkLocation = request.WorkLocation
	employee.Active = request.Active
	employee.RequiresApproval = request.RequiresApproval
	updated, err := h.svc.UpdateEmployee(c, employee)
	if err != nil {
		return writeHRError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Employee updated successfully.", newEmployeeResponse(updated))
}

// @Summary Delete employee
// @Description Deletes an employee by id. The id must be a valid positive integer, and a successful deletion returns no content.
// @Tags Employees
// @Accept json
// @Produce json
// @Param id path integer true "Employee ID"
// @Success 204 {object} httpx.EmptyEnvelope "Employee deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/employees/{id} [delete]
func (h EmployeeHandler) Delete(c fiber.Ctx) error {
	employeeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid employee id provided.", nil)
	}
	if err := h.svc.DeleteEmployee(c, employeeID); err != nil {
		httpx.RequestLog(c).Error("employee delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete employee.", err)
	}
	return httpx.CreateNoContentResponse(c)
}
