package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type WarehouseResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name"`
	Code           *string    `json:"code"`
	Line1          *string    `json:"line1"`
	Line2          *string    `json:"line2"`
	City           *string    `json:"city"`
	State          *string    `json:"state"`
	PostalCode     *string    `json:"postal_code"`
	CountryCode    *string    `json:"country_code"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

func newWarehouseResponse(warehouse *reference.Warehouse) WarehouseResponse {
	return WarehouseResponse{
		ID:             warehouse.ID,
		OrganizationID: warehouse.OrganizationID,
		Name:           warehouse.Name,
		Code:           warehouse.Code,
		Line1:          warehouse.Line1,
		Line2:          warehouse.Line2,
		City:           warehouse.City,
		State:          warehouse.State,
		PostalCode:     warehouse.PostalCode,
		CountryCode:    warehouse.CountryCode,
		CreatedAt:      warehouse.CreatedAt,
		UpdatedAt:      warehouse.UpdatedAt,
		DeletedAt:      warehouse.DeletedAt,
	}
}

var warehouseQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"code":            {},
	"created_at":      {},
	"updated_at":      {},
}

type ListWarehousesResponse struct {
	Warehouses []WarehouseResponse `json:"warehouses"`
}

type ListWarehousesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListWarehousesResponse `json:"data"`
}

// @Summary List warehouses
// @Description Lists the warehouses belonging to the caller's organization, honoring pagination, sorting, and filtering. Only allowlisted query fields are accepted and the tenant filter is always enforced, so callers can never see another organization's warehouses. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListWarehousesResponseEnvelope "Warehouses retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/warehouses [get]
func (h WarehouseHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, warehouseQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.ListWarehouses(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("warehouse list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve warehouses.", err)
	}

	items := make([]WarehouseResponse, len(page.Items))
	for i, warehouse := range page.Items {
		items[i] = newWarehouseResponse(warehouse)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "warehouses.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Warehouses retrieved successfully.", ListWarehousesResponse{
		Warehouses: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetWarehouseResponse struct {
	Warehouse WarehouseResponse `json:"warehouse"`
}

type GetWarehouseResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetWarehouseResponse `json:"data"`
}

// @Summary Get warehouse
// @Description Gets a single warehouse by its id, including its name, code, and address. The warehouse is looked up and returned only when it belongs to the caller's organization; otherwise the request is answered with 404 Not Found to avoid leaking the record's existence. A non-numeric or malformed id is rejected with 422 Unprocessable Entity.
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path integer true "Warehouse ID"
// @Success 200 {object} GetWarehouseResponseEnvelope "Warehouse retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Warehouse not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/warehouses/{id} [get]
func (h WarehouseHandler) Get(c fiber.Ctx) error {
	warehouseID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid warehouse id provided.", nil)
	}

	warehouse, err := h.svc.FindWarehouse(c, warehouseID)
	if err != nil {
		httpx.RequestLog(c).Error("warehouse lookup failed", "warehouse_id", warehouseID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get warehouse.", err)
	}
	if warehouse == nil || !httpx.OwnsTenant(c, warehouse.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Warehouse not found.")
	}

	return httpx.CreateSuccessResponse(c, "Warehouse retrieved successfully.", GetWarehouseResponse{
		Warehouse: newWarehouseResponse(warehouse),
	})
}

type CreateWarehouseRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	Code           *string `json:"code"`
	Line1          *string `json:"line1"`
	Line2          *string `json:"line2"`
	City           *string `json:"city"`
	State          *string `json:"state"`
	PostalCode     *string `json:"postal_code"`
	CountryCode    *string `json:"country_code"`
}

type CreateWarehouseResponse struct {
	Warehouse WarehouseResponse `json:"warehouse"`
}

type CreateWarehouseResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateWarehouseResponse `json:"data"`
}

// @Summary Create warehouse
// @Description Creates a warehouse for the caller's organization. The name is required, and the code, when provided, must be unique across the organization or a 409 Conflict is returned. Returns the created warehouse with a 201 status.
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param body body CreateWarehouseRequest true "Warehouse details"
// @Success 201 {object} CreateWarehouseResponseEnvelope "Warehouse created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/warehouses [post]
func (h WarehouseHandler) Create(c fiber.Ctx) error {
	var request CreateWarehouseRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	warehouse, err := h.svc.CreateWarehouse(c, &reference.Warehouse{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		Name:           request.Name,
		Code:           request.Code,
		Line1:          request.Line1,
		Line2:          request.Line2,
		City:           request.City,
		State:          request.State,
		PostalCode:     request.PostalCode,
		CountryCode:    request.CountryCode,
	})
	if err != nil {
		return writeWarehouseError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Warehouse created successfully.", CreateWarehouseResponse{
		Warehouse: newWarehouseResponse(warehouse),
	})
}

type UpdateWarehouseRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	Name           string  `json:"name" validate:"required"`
	Code           *string `json:"code"`
	Line1          *string `json:"line1"`
	Line2          *string `json:"line2"`
	City           *string `json:"city"`
	State          *string `json:"state"`
	PostalCode     *string `json:"postal_code"`
	CountryCode    *string `json:"country_code"`
}

type UpdateWarehouseResponse struct {
	Warehouse WarehouseResponse `json:"warehouse"`
}

type UpdateWarehouseResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateWarehouseResponse `json:"data"`
}

// @Summary Update warehouse
// @Description Updates an existing warehouse by its id after confirming it belongs to the caller's organization, otherwise a 404 Not Found is returned. The name is re-validated as required, and the code, when provided, must still be unique across the organization or a 409 Conflict is returned. Returns the updated warehouse.
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path integer true "Warehouse ID"
// @Param body body UpdateWarehouseRequest true "Warehouse details"
// @Success 200 {object} UpdateWarehouseResponseEnvelope "Warehouse updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Warehouse not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/warehouses/{id} [put]
func (h WarehouseHandler) Update(c fiber.Ctx) error {
	warehouseID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid warehouse id provided.", nil)
	}

	var request UpdateWarehouseRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	warehouse, err := h.svc.FindWarehouse(c, warehouseID)
	if err != nil {
		httpx.RequestLog(c).Error("warehouse lookup failed", "warehouse_id", warehouseID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update warehouse.", err)
	}
	if warehouse == nil || !httpx.OwnsTenant(c, warehouse.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Warehouse not found.")
	}

	warehouse.OrganizationID = httpx.TenantOrganizationID(c, warehouse.OrganizationID)
	warehouse.Name = request.Name
	warehouse.Code = request.Code
	warehouse.Line1 = request.Line1
	warehouse.Line2 = request.Line2
	warehouse.City = request.City
	warehouse.State = request.State
	warehouse.PostalCode = request.PostalCode
	warehouse.CountryCode = request.CountryCode

	updated, err := h.svc.UpdateWarehouse(c, warehouse)
	if err != nil {
		return writeWarehouseError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Warehouse updated successfully.", UpdateWarehouseResponse{
		Warehouse: newWarehouseResponse(updated),
	})
}

// @Summary Delete warehouse
// @Description Deletes a warehouse by its id. The warehouse is removed only when it exists and belongs to the caller's organization; otherwise a 404 Not Found is returned so the record's existence is not revealed. Responds with 204 No Content on success.
// @Tags Warehouses
// @Accept json
// @Produce json
// @Param id path integer true "Warehouse ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Warehouse not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/warehouses/{id} [delete]
func (h WarehouseHandler) Delete(c fiber.Ctx) error {
	warehouseID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid warehouse id provided.", nil)
	}

	warehouse, err := h.svc.FindWarehouse(c, warehouseID)
	if err != nil {
		httpx.RequestLog(c).Error("warehouse lookup failed", "warehouse_id", warehouseID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete warehouse.", err)
	}
	if warehouse == nil || !httpx.OwnsTenant(c, warehouse.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Warehouse not found.")
	}

	if err := h.svc.DeleteWarehouse(c, warehouseID); err != nil {
		httpx.RequestLog(c).Error("warehouse deletion failed", "warehouse_id", warehouseID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete warehouse.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
