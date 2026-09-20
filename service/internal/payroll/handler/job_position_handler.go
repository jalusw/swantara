package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type JobPositionHandler struct {
	svc payroll.HRService
}

func NewJobPositionHandler(svc payroll.HRService) JobPositionHandler {
	return JobPositionHandler{svc: svc}
}

type JobPositionResponse struct {
	ID             uint64  `json:"id"`
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name"`
	DepartmentID   *uint64 `json:"department_id"`
}

func newJobPositionResponse(position *reference.JobPosition) JobPositionResponse {
	return JobPositionResponse{ID: position.ID, OrganizationID: position.OrganizationID, Name: position.Name, DepartmentID: position.DepartmentID}
}

type CreateJobPositionRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	DepartmentID   *uint64 `json:"department_id"`
}

type ListJobPositionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []JobPositionResponse `json:"data"`
}

type JobPositionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data JobPositionResponse `json:"data"`
}

var jobPositionQueryAllowlist = map[string]struct{}{
	"name":            {},
	"department_id":   {},
	"organization_id": {},
}

// @Summary List job positions
// @Description Lists job positions with pagination, sorting, and filtering, scoped to the caller's organization. Queries are restricted to an allowlisted set of fields (name and department), and the organization filter is forced to the caller's tenant so cross-tenant job positions are never returned.
// @Tags Job Positions
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListJobPositionsResponseEnvelope "Job positions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/job-positions [get]
func (h JobPositionHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, jobPositionQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListJobPositions(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("job position list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve job positions.", err)
	}
	items := make([]JobPositionResponse, len(page.Items))
	for i, position := range page.Items {
		items[i] = newJobPositionResponse(position)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Job positions retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get job position
// @Description Gets a single job position by its id. The job position must belong to the caller's organization, and a 404 not found is returned when it does not, even if a job position with that id exists in another tenant.
// @Tags Job Positions
// @Accept json
// @Produce json
// @Param id path integer true "Job position ID"
// @Success 200 {object} JobPositionResponseEnvelope "Job position retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Job position not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/job-positions/{id} [get]
func (h JobPositionHandler) Get(c fiber.Ctx) error {
	positionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid job position id provided.", nil)
	}
	position, err := h.svc.FindJobPosition(c, positionID)
	if err != nil {
		httpx.RequestLog(c).Error("job position lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get job position.", err)
	}
	if position == nil || !httpx.OwnsTenant(c, position.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Job position not found.")
	}
	return httpx.CreateSuccessResponse(c, "Job position retrieved successfully.", newJobPositionResponse(position))
}

// @Summary Create job position
// @Description Creates a new job position under the caller's organization, optionally linked to a department. The position name is required, and when no organization is supplied the caller's tenant is used.
// @Tags Job Positions
// @Accept json
// @Produce json
// @Param request body CreateJobPositionRequest true "Job position details"
// @Success 201 {object} JobPositionResponseEnvelope "Job position created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/job-positions [post]
func (h JobPositionHandler) Create(c fiber.Ctx) error {
	var request CreateJobPositionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	created, err := h.svc.CreateJobPosition(c, &reference.JobPosition{
		OrganizationID: organizationID,
		Name:           request.Name,
		DepartmentID:   request.DepartmentID,
	})
	if err != nil {
		httpx.RequestLog(c).Error("job position create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create job position.", err)
	}
	return httpx.CreateCreatedResponse(c, "Job position created successfully.", newJobPositionResponse(created))
}

// @Summary Update job position
// @Description Updates an existing job position's name and department link by id. The job position must belong to the caller's organization (404 otherwise), and the updated payload must pass validation.
// @Tags Job Positions
// @Accept json
// @Produce json
// @Param id path integer true "Job position ID"
// @Param request body CreateJobPositionRequest true "Job position details"
// @Success 200 {object} JobPositionResponseEnvelope "Job position updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Job position not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/job-positions/{id} [put]
func (h JobPositionHandler) Update(c fiber.Ctx) error {
	positionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid job position id provided.", nil)
	}
	position, err := h.svc.FindJobPosition(c, positionID)
	if err != nil {
		httpx.RequestLog(c).Error("job position lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get job position.", err)
	}
	if position == nil || !httpx.OwnsTenant(c, position.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Job position not found.")
	}
	var request CreateJobPositionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	position.Name = request.Name
	position.DepartmentID = request.DepartmentID
	updated, err := h.svc.UpdateJobPosition(c, position)
	if err != nil {
		httpx.RequestLog(c).Error("job position update failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update job position.", err)
	}
	return httpx.CreateSuccessResponse(c, "Job position updated successfully.", newJobPositionResponse(updated))
}

// @Summary Delete job position
// @Description Deletes a job position by id. The job position must exist within the caller's organization or a 404 response is returned, and a successful deletion returns no content.
// @Tags Job Positions
// @Accept json
// @Produce json
// @Param id path integer true "Job position ID"
// @Success 204 {object} httpx.EmptyEnvelope "Job position deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Job position not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/job-positions/{id} [delete]
func (h JobPositionHandler) Delete(c fiber.Ctx) error {
	positionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid job position id provided.", nil)
	}
	position, err := h.svc.FindJobPosition(c, positionID)
	if err != nil {
		httpx.RequestLog(c).Error("job position lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get job position.", err)
	}
	if position == nil || !httpx.OwnsTenant(c, position.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Job position not found.")
	}
	if err := h.svc.DeleteJobPosition(c, positionID); err != nil {
		httpx.RequestLog(c).Error("job position delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete job position.", err)
	}
	return httpx.CreateNoContentResponse(c)
}
