package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/payroll"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type LeaveTypeHandler struct {
	svc payroll.HRService
}

func NewLeaveTypeHandler(svc payroll.HRService) LeaveTypeHandler {
	return LeaveTypeHandler{svc: svc}
}

type LeaveTypeResponse struct {
	ID             uint64   `json:"id"`
	OrganizationID *uint64  `json:"organization_id"`
	Name           string   `json:"name"`
	Paid           bool     `json:"paid"`
	AllocationDays *float64 `json:"allocation_days"`
}

func newLeaveTypeResponse(leaveType *reference.LeaveType) LeaveTypeResponse {
	return LeaveTypeResponse{ID: leaveType.ID, OrganizationID: leaveType.OrganizationID, Name: leaveType.Name, Paid: leaveType.Paid, AllocationDays: leaveType.AllocationDays}
}

type CreateLeaveTypeRequest struct {
	OrganizationID *uint64  `json:"organization_id"`
	Name           string   `json:"name" validate:"required"`
	Paid           bool     `json:"paid"`
	AllocationDays *float64 `json:"allocation_days"`
}

type ListLeaveTypesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data []LeaveTypeResponse `json:"data"`
}

type LeaveTypeResponseEnvelope struct {
	httpx.EnvelopeBase
	Data LeaveTypeResponse `json:"data"`
}

var leaveTypeQueryAllowlist = map[string]struct{}{
	"name":            {},
	"paid":            {},
	"allocation_days": {},
	"organization_id": {},
}

// @Summary List leave types
// @Description Lists leave types with pagination, sorting, and filtering, scoped to the caller's organization. Queries are restricted to allowlisted fields (name, paid status, and allocation days), and the organization filter is forced to the caller's tenant so cross-tenant leave types are never returned.
// @Tags Leave Types
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListLeaveTypesResponseEnvelope "Leave types retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-types [get]
func (h LeaveTypeHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, leaveTypeQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.ListLeaveTypes(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("leave type list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve leave types.", err)
	}
	items := make([]LeaveTypeResponse, len(page.Items))
	for i, leaveType := range page.Items {
		items[i] = newLeaveTypeResponse(leaveType)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Leave types retrieved successfully.", items, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get leave type
// @Description Gets a single leave type by its id. The leave type must belong to the caller's organization, and a 404 not found is returned when it does not, even if a leave type with that id exists in another tenant.
// @Tags Leave Types
// @Accept json
// @Produce json
// @Param id path integer true "Leave type ID"
// @Success 200 {object} LeaveTypeResponseEnvelope "Leave type retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Leave type not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-types/{id} [get]
func (h LeaveTypeHandler) Get(c fiber.Ctx) error {
	leaveTypeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave type id provided.", nil)
	}
	leaveType, err := h.svc.FindLeaveType(c, leaveTypeID)
	if err != nil {
		httpx.RequestLog(c).Error("leave type lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get leave type.", err)
	}
	if leaveType == nil || !httpx.OwnsTenant(c, leaveType.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Leave type not found.")
	}
	return httpx.CreateSuccessResponse(c, "Leave type retrieved successfully.", newLeaveTypeResponse(leaveType))
}

// @Summary Create leave type
// @Description Creates a new leave type under the caller's organization with a paid status flag and an optional annual day allocation. The leave type name is required, and when no organization is supplied the caller's tenant is used.
// @Tags Leave Types
// @Accept json
// @Produce json
// @Param request body CreateLeaveTypeRequest true "Leave type details"
// @Success 201 {object} LeaveTypeResponseEnvelope "Leave type created successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-types [post]
func (h LeaveTypeHandler) Create(c fiber.Ctx) error {
	var request CreateLeaveTypeRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	created, err := h.svc.CreateLeaveType(c, &reference.LeaveType{
		OrganizationID: organizationID,
		Name:           request.Name,
		Paid:           request.Paid,
		AllocationDays: request.AllocationDays,
	})
	if err != nil {
		httpx.RequestLog(c).Error("leave type create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create leave type.", err)
	}
	return httpx.CreateCreatedResponse(c, "Leave type created successfully.", newLeaveTypeResponse(created))
}

// @Summary Update leave type
// @Description Updates an existing leave type's name, paid status, and annual day allocation by id. The leave type must belong to the caller's organization (404 otherwise), and the updated payload must pass validation.
// @Tags Leave Types
// @Accept json
// @Produce json
// @Param id path integer true "Leave type ID"
// @Param request body CreateLeaveTypeRequest true "Leave type details"
// @Success 200 {object} LeaveTypeResponseEnvelope "Leave type updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Leave type not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-types/{id} [put]
func (h LeaveTypeHandler) Update(c fiber.Ctx) error {
	leaveTypeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave type id provided.", nil)
	}
	leaveType, err := h.svc.FindLeaveType(c, leaveTypeID)
	if err != nil {
		httpx.RequestLog(c).Error("leave type lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get leave type.", err)
	}
	if leaveType == nil || !httpx.OwnsTenant(c, leaveType.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Leave type not found.")
	}
	var request CreateLeaveTypeRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	leaveType.Name = request.Name
	leaveType.Paid = request.Paid
	leaveType.AllocationDays = request.AllocationDays
	updated, err := h.svc.UpdateLeaveType(c, leaveType)
	if err != nil {
		httpx.RequestLog(c).Error("leave type update failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update leave type.", err)
	}
	return httpx.CreateSuccessResponse(c, "Leave type updated successfully.", newLeaveTypeResponse(updated))
}

// @Summary Delete leave type
// @Description Deletes a leave type by id. The leave type must exist within the caller's organization or a 404 response is returned, and a successful deletion returns no content.
// @Tags Leave Types
// @Accept json
// @Produce json
// @Param id path integer true "Leave type ID"
// @Success 204 {object} httpx.EmptyEnvelope "Leave type deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Leave type not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/leave-types/{id} [delete]
func (h LeaveTypeHandler) Delete(c fiber.Ctx) error {
	leaveTypeID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid leave type id provided.", nil)
	}
	leaveType, err := h.svc.FindLeaveType(c, leaveTypeID)
	if err != nil {
		httpx.RequestLog(c).Error("leave type lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get leave type.", err)
	}
	if leaveType == nil || !httpx.OwnsTenant(c, leaveType.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Leave type not found.")
	}
	if err := h.svc.DeleteLeaveType(c, leaveTypeID); err != nil {
		httpx.RequestLog(c).Error("leave type delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete leave type.", err)
	}
	return httpx.CreateNoContentResponse(c)
}
