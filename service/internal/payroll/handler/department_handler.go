package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type DepartmentHandler struct {
	svc payroll.HRService
}

func NewDepartmentHandler(svc payroll.HRService) DepartmentHandler {
	return DepartmentHandler{svc: svc}
}

type DepartmentResponse struct {
	ID             uint64  `json:"id"`
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	Description    *string `json:"description"`
	ParentID       *uint64 `json:"parent_id"`
	ManagerID      *uint64 `json:"manager_id"`
	DimensionID    *uint64 `json:"dimension_id"`
}

func newDepartmentResponse(department *reference.Department) DepartmentResponse {
	return DepartmentResponse{
		ID:             department.ID,
		OrganizationID: department.OrganizationID,
		Name:           department.Name,
		Description:    department.Description,
		ParentID:       department.ParentID,
		ManagerID:      department.ManagerID,
		DimensionID:    department.DimensionID,
	}
}

type CreateDepartmentRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	Description    *string `json:"description"`
	ParentID       *uint64 `json:"parent_id"`
	ManagerID      *uint64 `json:"manager_id"`
	DimensionID    *uint64 `json:"dimension_id"`
}

type ListDepartmentsResponse struct {
	Departments []DepartmentResponse `json:"departments"`
}

type ListDepartmentsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListDepartmentsResponse `json:"data"`
}

type GetDepartmentResponse struct {
	Department DepartmentResponse `json:"department"`
}

type DepartmentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetDepartmentResponse `json:"data"`
}

var departmentQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"description":     {},
	"parent_id":       {},
	"manager_id":      {},
}

// @Summary List departments
// @Description Lists departments with pagination, sorting, and filtering, scoped to the caller's organization. Queries are restricted to an allowlisted set of fields (organization, name, parent, and manager), and the organization filter is forced to the caller's tenant so cross-tenant departments are never returned.
// @Tags Departments
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListDepartmentsResponseEnvelope "Departments retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/departments [get]
func (h DepartmentHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, departmentQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListDepartments(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("department list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve departments.", err)
	}
	items := make([]DepartmentResponse, len(page.Items))
	for i, department := range page.Items {
		items[i] = newDepartmentResponse(department)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Departments retrieved successfully.", ListDepartmentsResponse{Departments: items}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get department
// @Description Gets a single department by its id. The id must be a valid positive integer, and the department must belong to the caller's organization; a 404 not found is returned when the id is invalid or the department exists in another tenant.
// @Tags Departments
// @Accept json
// @Produce json
// @Param id path integer true "Department ID"
// @Success 200 {object} DepartmentResponseEnvelope "Department retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Department not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/departments/{id} [get]
func (h DepartmentHandler) Get(c fiber.Ctx) error {
	departmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid department id provided.", nil)
	}
	department, err := h.svc.FindDepartment(c, departmentID)
	if err != nil {
		httpx.RequestLog(c).Error("department lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get department.", err)
	}
	if department == nil || !httpx.OwnsTenant(c, department.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Department not found.")
	}
	return httpx.CreateSuccessResponse(c, "Department retrieved successfully.", GetDepartmentResponse{Department: newDepartmentResponse(department)})
}

// @Summary Create department
// @Description Creates a new department under the caller's organization. The department name is required, and when no organization is supplied the caller's tenant is used; a 422 response is returned if no organization can be resolved.
// @Tags Departments
// @Accept json
// @Produce json
// @Param request body CreateDepartmentRequest true "Department details"
// @Success 201 {object} DepartmentResponseEnvelope "Department created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/departments [post]
func (h DepartmentHandler) Create(c fiber.Ctx) error {
	var request CreateDepartmentRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	department := &reference.Department{
		OrganizationID: organizationID,
		Name:           request.Name,
		Description:    request.Description,
		ParentID:       request.ParentID,
		ManagerID:      request.ManagerID,
		DimensionID:    request.DimensionID,
	}
	created, err := h.svc.CreateDepartment(c, department)
	if err != nil {
		httpx.RequestLog(c).Error("department create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create department.", err)
	}
	return httpx.CreateCreatedResponse(c, "Department created successfully.", GetDepartmentResponse{Department: newDepartmentResponse(created)})
}

// @Summary Update department
// @Description Updates an existing department's name, parent, manager, and dimension account by id. The department must belong to the caller's organization (404 otherwise), and the request is validated before any change is persisted.
// @Tags Departments
// @Accept json
// @Produce json
// @Param id path integer true "Department ID"
// @Param request body CreateDepartmentRequest true "Department details"
// @Success 200 {object} DepartmentResponseEnvelope "Department updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Department not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/departments/{id} [put]
func (h DepartmentHandler) Update(c fiber.Ctx) error {
	departmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid department id provided.", nil)
	}
	department, err := h.svc.FindDepartment(c, departmentID)
	if err != nil {
		httpx.RequestLog(c).Error("department lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get department.", err)
	}
	if department == nil || !httpx.OwnsTenant(c, department.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Department not found.")
	}
	var request CreateDepartmentRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	department.Name = request.Name
	department.Description = request.Description
	department.ParentID = request.ParentID
	department.ManagerID = request.ManagerID
	department.DimensionID = request.DimensionID
	updated, err := h.svc.UpdateDepartment(c, department)
	if err != nil {
		httpx.RequestLog(c).Error("department update failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update department.", err)
	}
	return httpx.CreateSuccessResponse(c, "Department updated successfully.", GetDepartmentResponse{Department: newDepartmentResponse(updated)})
}

// @Summary Delete department
// @Description Deletes a department by id. The department must exist within the caller's organization or a 404 response is returned, and a successful deletion returns no content.
// @Tags Departments
// @Accept json
// @Produce json
// @Param id path integer true "Department ID"
// @Success 204 {object} httpx.EmptyEnvelope "Department deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Department not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/departments/{id} [delete]
func (h DepartmentHandler) Delete(c fiber.Ctx) error {
	departmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid department id provided.", nil)
	}
	department, err := h.svc.FindDepartment(c, departmentID)
	if err != nil {
		httpx.RequestLog(c).Error("department lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get department.", err)
	}
	if department == nil || !httpx.OwnsTenant(c, department.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Department not found.")
	}
	if err := h.svc.DeleteDepartment(c, departmentID); err != nil {
		httpx.RequestLog(c).Error("department delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete department.", err)
	}
	return httpx.CreateNoContentResponse(c)
}
