package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
	"github.com/jalusw/swantara/apps/service/internal/project"
)

type CreateProjectRequest struct {
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name" validate:"required"`
	ContactID      uint64     `json:"contact_id" validate:"required,gt=0"`
	ManagerID      *uint64    `json:"manager_id"`
	DimensionID    *uint64    `json:"dimension_id"`
	SaleOrderID    *uint64    `json:"sale_order_id"`
	BillingType    string     `json:"billing_type" validate:"required,oneof=fixed time_material milestone"`
	BillableRate   float64    `json:"billable_rate"`
	DateStart      *time.Time `json:"date_start"`
	DateEnd        *time.Time `json:"date_end"`
}

type ListProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

type ListProjectsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListProjectsResponse `json:"data"`
}

// @Summary List projects
// @Description Lists projects with pagination, sorting, and filtering. Results are scoped to the caller's organization and can be narrowed by contact, manager, billing type, and state.
// @Tags Projects
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListProjectsResponseEnvelope "Projects retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects [get]
func (h ProjectHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, projectQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("project list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve projects.", err)
	}
	items := make([]ProjectResponse, len(page.Items))
	for i, p := range page.Items {
		items[i] = newProjectResponse(p)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Projects retrieved successfully.", ListProjectsResponse{
		Projects: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetProjectResponse struct {
	Project ProjectResponse `json:"project"`
}

type GetProjectResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetProjectResponse `json:"data"`
}

// @Summary Get project
// @Description Gets a single project by id, including its contact, manager, billing type, billable rate, and lifecycle state. Projects not belonging to the caller's organization are treated as not found.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path integer true "Project ID"
// @Success 200 {object} GetProjectResponseEnvelope "Project retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Project not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{id} [get]
func (h ProjectHandler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	project, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("project get failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve project.", err)
	}
	if project == nil || !httpx.OwnsTenant(c, helper.Ptr(project.OrganizationID)) {
		return httpx.CreateNotFoundResponse(c, "Project not found.")
	}
	return httpx.CreateSuccessResponse(c, "Project retrieved successfully.", GetProjectResponse{
		Project: newProjectResponse(project),
	})
}

type CreateProjectResponse struct {
	Project ProjectResponse `json:"project"`
}

type CreateProjectResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateProjectResponse `json:"data"`
}

// @Summary Create project
// @Description Creates a new project within the caller's organization in draft state. The billing type must be fixed, time and material, or milestone, the contact and dimension account (when given) must exist, and an optional date range must not be reversed.
// @Tags Projects
// @Accept json
// @Produce json
// @Param body body CreateProjectRequest true "Project details"
// @Success 201 {object} CreateProjectResponseEnvelope "Project created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects [post]
func (h ProjectHandler) Create(c fiber.Ctx) error {
	var request CreateProjectRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization ID is required.", nil)
	}
	created, err := h.svc.CreateProject(c, &project.Project{
		OrganizationID: *organizationID,
		Name:           request.Name,
		ContactID:      request.ContactID,
		ManagerID:      request.ManagerID,
		DimensionID:    request.DimensionID,
		SaleOrderID:    request.SaleOrderID,
		BillingType:    request.BillingType,
		BillableRate:   request.BillableRate,
		DateStart:      request.DateStart,
		DateEnd:        request.DateEnd,
	})
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Project created successfully.", CreateProjectResponse{
		Project: newProjectResponse(created),
	})
}

type UpdateProjectRequest struct {
	Name         string     `json:"name" validate:"required"`
	ContactID    uint64     `json:"contact_id" validate:"required,gt=0"`
	ManagerID    *uint64    `json:"manager_id"`
	DimensionID  *uint64    `json:"dimension_id"`
	SaleOrderID  *uint64    `json:"sale_order_id"`
	BillingType  string     `json:"billing_type" validate:"required,oneof=fixed time_material milestone"`
	BillableRate float64    `json:"billable_rate"`
	DateStart    *time.Time `json:"date_start"`
	DateEnd      *time.Time `json:"date_end"`
}

// @Summary Update project
// @Description Updates an existing project's details such as name, contact, billing configuration, and dates. The project's lifecycle state cannot be changed through this endpoint, and the contact and dimension account references are validated.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path integer true "Project ID"
// @Param body body UpdateProjectRequest true "Project details"
// @Success 200 {object} GetProjectResponseEnvelope "Project updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Project not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{id} [put]
func (h ProjectHandler) Update(c fiber.Ctx) error {
	var request UpdateProjectRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	updated, err := h.svc.UpdateProject(c, &project.Project{
		Base:         model.Base{ID: id},
		Name:         request.Name,
		ContactID:    request.ContactID,
		ManagerID:    request.ManagerID,
		DimensionID:  request.DimensionID,
		SaleOrderID:  request.SaleOrderID,
		BillingType:  request.BillingType,
		BillableRate: request.BillableRate,
		DateStart:    request.DateStart,
		DateEnd:      request.DateEnd,
	})
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Project updated successfully.", GetProjectResponse{
		Project: newProjectResponse(updated),
	})
}

type SetProjectStateRequest struct {
	State string `json:"state" validate:"required,oneof=draft open closed cancelled"`
}

// @Summary Set project state
// @Description Moves a project through its lifecycle by transitioning state: draft to open, draft or open to cancelled, and open to closed. Disallowed transitions are rejected so the project workflow stays consistent.
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path integer true "Project ID"
// @Param body body SetProjectStateRequest true "Target state"
// @Success 200 {object} GetProjectResponseEnvelope "Project state updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Project not found"
// @Failure 422 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/projects/{id}/state [put]
func (h ProjectHandler) SetState(c fiber.Ctx) error {
	var request SetProjectStateRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid id.", nil)
	}
	updated, err := h.svc.SetProjectState(c, id, request.State)
	if err != nil {
		return writeProjectError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Project state updated successfully.", GetProjectResponse{
		Project: newProjectResponse(updated),
	})
}
