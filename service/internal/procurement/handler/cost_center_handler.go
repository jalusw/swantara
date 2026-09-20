package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type CostCenterHandler struct {
	svc procurement.CostCenterService
}

func NewCostCenterHandler(svc procurement.CostCenterService) CostCenterHandler {
	return CostCenterHandler{svc: svc}
}

type CostCenterResponse struct {
	ID     uint64  `json:"id"`
	Name   *string `json:"name"`
	Code   *string `json:"code"`
	Active bool    `json:"active"`
}

func newCostCenterResponse(cc *procurement.CostCenter) CostCenterResponse {
	return CostCenterResponse{
		ID:     cc.ID,
		Name:   cc.Name,
		Code:   cc.Code,
		Active: cc.Active,
	}
}

var costCenterQueryAllowlist = map[string]struct{}{
	"name":       {},
	"code":       {},
	"active":     {},
	"created_at": {},
	"updated_at": {},
}

type ListCostCentersResponse struct {
	CostCenters []CostCenterResponse `json:"cost_centers"`
}

type ListCostCentersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListCostCentersResponse `json:"data"`
}

// @Summary List cost centers
// @Description Lists cost centers across the caller's organization with pagination, sorting, and filtering.
// @Tags Cost Centers
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListCostCentersResponseEnvelope "Cost centers retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/cost-centers [get]
func (h CostCenterHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, costCenterQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("cost center list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve cost centers.", err)
	}
	items := make([]CostCenterResponse, len(page.Items))
	for i, cc := range page.Items {
		items[i] = newCostCenterResponse(cc)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Cost centers retrieved successfully.", ListCostCentersResponse{
		CostCenters: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetCostCenterResponse struct {
	CostCenter CostCenterResponse `json:"cost_center"`
}

type GetCostCenterResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetCostCenterResponse `json:"data"`
}

// @Summary Get cost center
// @Description Gets a single cost center by id.
// @Tags Cost Centers
// @Accept json
// @Produce json
// @Param id path integer true "Cost center ID"
// @Success 200 {object} GetCostCenterResponseEnvelope "Cost center retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Cost center not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/cost-centers/{id} [get]
func (h CostCenterHandler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid cost center id provided.", nil)
	}
	cc, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("cost center lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get cost center.", err)
	}
	if cc == nil || !httpx.OwnsTenant(c, cc.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Cost center not found.")
	}
	return httpx.CreateSuccessResponse(c, "Cost center retrieved successfully.", GetCostCenterResponse{
		CostCenter: newCostCenterResponse(cc),
	})
}

type CreateCostCenterRequest struct {
	Name   *string `json:"name" validate:"required"`
	Code   *string `json:"code" validate:"required"`
	Active *bool   `json:"active"`
}

type CreateCostCenterResponse struct {
	CostCenter CostCenterResponse `json:"cost_center"`
}

type CreateCostCenterResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateCostCenterResponse `json:"data"`
}

// @Summary Create cost center
// @Description Creates a new cost center for the caller's organization.
// @Tags Cost Centers
// @Accept json
// @Produce json
// @Param body body CreateCostCenterRequest true "Cost center details"
// @Success 201 {object} CreateCostCenterResponseEnvelope "Cost center created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/cost-centers [post]
func (h CostCenterHandler) Create(c fiber.Ctx) error {
	var request CreateCostCenterRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	active := true
	if request.Active != nil {
		active = *request.Active
	}
	cc := &procurement.CostCenter{
		OrganizationID: httpx.TenantOrganizationID(c, nil),
		Name:           request.Name,
		Code:           request.Code,
		Active:         active,
	}
	created, err := h.svc.Create(c, cc)
	if err != nil {
		httpx.RequestLog(c).Error("cost center create failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to create cost center.", err)
	}
	return httpx.CreateCreatedResponse(c, "Cost center created successfully.", CreateCostCenterResponse{
		CostCenter: newCostCenterResponse(created),
	})
}

type UpdateCostCenterRequest struct {
	Name   *string `json:"name"`
	Code   *string `json:"code"`
	Active *bool   `json:"active"`
}

// @Summary Update cost center
// @Description Updates an existing cost center.
// @Tags Cost Centers
// @Accept json
// @Produce json
// @Param id path integer true "Cost center ID"
// @Param body body UpdateCostCenterRequest true "Cost center update"
// @Success 200 {object} GetCostCenterResponseEnvelope "Cost center updated successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Cost center not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/cost-centers/{id} [put]
func (h CostCenterHandler) Update(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid cost center id provided.", nil)
	}
	existing, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("cost center lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get cost center.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Cost center not found.")
	}
	var request UpdateCostCenterRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	if request.Name != nil {
		existing.Name = request.Name
	}
	if request.Code != nil {
		existing.Code = request.Code
	}
	if request.Active != nil {
		existing.Active = *request.Active
	}
	updated, err := h.svc.Update(c, existing)
	if err != nil {
		httpx.RequestLog(c).Error("cost center update failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update cost center.", err)
	}
	return httpx.CreateSuccessResponse(c, "Cost center updated successfully.", GetCostCenterResponse{
		CostCenter: newCostCenterResponse(updated),
	})
}

type DeleteCostCenterResponse struct {
	Message string `json:"message"`
}

type DeleteCostCenterResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DeleteCostCenterResponse `json:"data"`
}

// @Summary Delete cost center
// @Description Deletes a cost center by id.
// @Tags Cost Centers
// @Accept json
// @Produce json
// @Param id path integer true "Cost center ID"
// @Success 200 {object} DeleteCostCenterResponseEnvelope "Cost center deleted successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Cost center not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/cost-centers/{id} [delete]
func (h CostCenterHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid cost center id provided.", nil)
	}
	existing, err := h.svc.Find(c, id)
	if err != nil {
		httpx.RequestLog(c).Error("cost center lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get cost center.", err)
	}
	if existing == nil || !httpx.OwnsTenant(c, existing.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Cost center not found.")
	}
	if err := h.svc.Delete(c, id); err != nil {
		httpx.RequestLog(c).Error("cost center delete failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete cost center.", err)
	}
	return httpx.CreateSuccessResponse(c, "Cost center deleted successfully.", DeleteCostCenterResponse{
		Message: "Cost center deleted successfully.",
	})
}
