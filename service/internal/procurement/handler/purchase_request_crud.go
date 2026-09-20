package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

// @Summary List purchase requisitions
// @Description Lists purchase requisitions across the caller's organization with pagination, sorting, and filtering on attributes such as requester, department, state, and needed-by date. The result set is always scoped to the caller's organization even when no organization_id is supplied. Returns the matching requisitions together with pagination metadata.
// @Tags Purchase Requisitions
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListPurchaseRequestsResponseEnvelope "Purchase requisitions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-requisitions [get]
func (h PurchaseRequestHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, requisitionQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}
	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("purchase requisition list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve purchase requisitions.", err)
	}
	items := make([]PurchaseRequestResponse, len(page.Items))
	for i, req := range page.Items {
		items[i] = newPurchaseRequestResponse(req)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "Purchase requisitions retrieved successfully.", ListPurchaseRequestsResponse{
		Requisitions: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

// @Summary Get purchase requisition
// @Description Gets a single purchase requisition by id together with all of its lines, including item, quantity, and needed-by date. The lookup is scoped to the caller's organization, and a 404 is returned when the requisition does not exist or belongs to another tenant.
// @Tags Purchase Requisitions
// @Accept json
// @Produce json
// @Param id path integer true "Purchase requisition ID"
// @Success 200 {object} GetPurchaseRequestResponseEnvelope "Purchase requisition retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase requisition not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-requisitions/{id} [get]
func (h PurchaseRequestHandler) Get(c fiber.Ctx) error {
	requestID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase requisition id provided.", nil)
	}
	requisition, err := h.svc.Find(c, requestID)
	if err != nil {
		httpx.RequestLog(c).Error("purchase requisition lookup failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get purchase requisition.", err)
	}
	if requisition == nil || !httpx.OwnsTenant(c, requisition.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Purchase requisition not found.")
	}
	lines, err := h.svc.ListLines(c, requestID)
	if err != nil {
		httpx.RequestLog(c).Error("purchase requisition line list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get purchase requisition.", err)
	}
	response := newPurchaseRequestResponse(requisition)
	response.Lines = make([]PurchaseRequestLineResponse, len(lines))
	for i, line := range lines {
		response.Lines[i] = newPurchaseRequestLineResponse(line)
	}
	return httpx.CreateSuccessResponse(c, "Purchase requisition retrieved successfully.", GetPurchaseRequestResponse{
		Request: response,
	})
}

// @Summary Create purchase requisition
// @Description Creates a new draft purchase requisition with its request lines, validating that the requester exists and that every line quantity is positive. A sequence number is generated for the requisition and it is saved in draft state, ready for confirmation and later approval.
// @Tags Purchase Requisitions
// @Accept json
// @Produce json
// @Param body body CreatePurchaseRequestRequest true "Purchase requisition details"
// @Success 201 {object} CreatePurchaseRequestResponseEnvelope "Purchase requisition created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-requisitions [post]
func (h PurchaseRequestHandler) Create(c fiber.Ctx) error {
	var request CreatePurchaseRequestRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	neededBy, err := helper.ParseDate(request.NeededBy)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid needed_by date provided.", nil)
	}
	lines := make([]*procurement.PurchaseRequestLine, len(request.Lines))
	for i, line := range request.Lines {
		lineNeededBy, err := helper.ParseDate(line.NeededBy)
		if err != nil {
			return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid line needed_by date provided.", nil)
		}
		lines[i] = &procurement.PurchaseRequestLine{
			ItemID:      line.ItemID,
			Description: line.Description,
			Qty:         line.Qty,
			UnitID:      line.UnitID,
			NeededBy:    lineNeededBy,
		}
	}
	requisition, err := h.svc.Create(c, &procurement.PurchaseRequest{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		RequesterID:    request.RequesterID,
		DepartmentID:   request.DepartmentID,
		NeededBy:       neededBy,
	}, lines)
	if err != nil {
		return writeRequisitionError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "Purchase requisition created successfully.", CreatePurchaseRequestResponse{
		Request: newPurchaseRequestResponse(requisition),
	})
}
