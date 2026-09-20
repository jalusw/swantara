package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type StockLocationResponse struct {
	ID          uint64     `json:"id"`
	WarehouseID *uint64    `json:"warehouse_id"`
	Name        string     `json:"name"`
	Code        *string    `json:"code"`
	ParentID    *uint64    `json:"parent_id"`
	Usage       string     `json:"usage"`
	Barcode     *string    `json:"barcode"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

func newStockLocationResponse(location *reference.StockLocation) StockLocationResponse {
	return StockLocationResponse{
		ID:          location.ID,
		WarehouseID: location.WarehouseID,
		Name:        location.Name,
		Code:        location.Code,
		ParentID:    location.ParentID,
		Usage:       location.Usage,
		Barcode:     location.Barcode,
		CreatedAt:   location.CreatedAt,
		UpdatedAt:   location.UpdatedAt,
		DeletedAt:   location.DeletedAt,
	}
}

var stockLocationQueryAllowlist = map[string]struct{}{
	"warehouse_id": {},
	"name":         {},
	"code":         {},
	"parent_id":    {},
	"usage":        {},
	"created_at":   {},
	"updated_at":   {},
}

type ListStockLocationsResponse struct {
	Locations []StockLocationResponse `json:"locations"`
}

type ListStockLocationsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListStockLocationsResponse `json:"data"`
}

// @Summary List stock locations
// @Description Lists the stock locations belonging to the caller's organization, honoring pagination, sorting, and filtering. Only allowlisted query fields are accepted and the tenant filter is always enforced, so callers can never see another organization's locations. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Stock Locations
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListStockLocationsResponseEnvelope "Stock locations retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-locations [get]
func (h WarehouseHandler) ListLocations(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, stockLocationQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.ListLocations(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("stock location list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve stock locations.", err)
	}

	items := make([]StockLocationResponse, len(page.Items))
	for i, location := range page.Items {
		items[i] = newStockLocationResponse(location)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "stock-locations.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Stock locations retrieved successfully.", ListStockLocationsResponse{
		Locations: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetStockLocationResponse struct {
	Location StockLocationResponse `json:"location"`
}

type GetStockLocationResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetStockLocationResponse `json:"data"`
}

// @Summary Get stock location
// @Description Gets a single stock location by its id, including its warehouse, parent, usage, and barcode. The location is looked up and returned only when it belongs to the caller's organization; otherwise the request is answered with 404 Not Found to avoid leaking the record's existence. A non-numeric or malformed id is rejected with 422 Unprocessable Entity.
// @Tags Stock Locations
// @Accept json
// @Produce json
// @Param id path integer true "Stock location ID"
// @Success 200 {object} GetStockLocationResponseEnvelope "Stock location retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock location not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-locations/{id} [get]
func (h WarehouseHandler) GetLocation(c fiber.Ctx) error {
	locationID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock location id provided.", nil)
	}

	location, err := h.svc.FindLocation(c, locationID)
	if err != nil {
		httpx.RequestLog(c).Error("stock location lookup failed", "location_id", locationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get stock location.", err)
	}
	if location == nil || !httpx.OwnsTenant(c, location.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock location not found.")
	}

	return httpx.CreateSuccessResponse(c, "Stock location retrieved successfully.", GetStockLocationResponse{
		Location: newStockLocationResponse(location),
	})
}

type CreateStockLocationRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	WarehouseID    *uint64 `json:"warehouse_id"`
	Name           string  `json:"name" validate:"required"`
	Code           *string `json:"code"`
	ParentID       *uint64 `json:"parent_id"`
	Usage          string  `json:"usage" validate:"required"`
	Barcode        *string `json:"barcode"`
}

type CreateStockLocationResponse struct {
	Location StockLocationResponse `json:"location"`
}

type CreateStockLocationResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateStockLocationResponse `json:"data"`
}

// @Summary Create stock location
// @Description Creates a stock location for the caller's organization. The name and a valid usage type are required, and when a parent location or warehouse is provided it must belong to the same organization or the request is rejected. Returns the created location with a 201 status.
// @Tags Stock Locations
// @Accept json
// @Produce json
// @Param body body CreateStockLocationRequest true "Stock location details"
// @Success 201 {object} CreateStockLocationResponseEnvelope "Stock location created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-locations [post]
func (h WarehouseHandler) CreateLocation(c fiber.Ctx) error {
	var request CreateStockLocationRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	location, err := h.svc.CreateLocation(c, &reference.StockLocation{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		WarehouseID:    request.WarehouseID,
		Name:           request.Name,
		Code:           request.Code,
		ParentID:       request.ParentID,
		Usage:          request.Usage,
		Barcode:        request.Barcode,
	})
	if err != nil {
		return writeWarehouseError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Stock location created successfully.", CreateStockLocationResponse{
		Location: newStockLocationResponse(location),
	})
}

type UpdateStockLocationRequest struct {
	OrganizationID *uint64 `json:"organization_id"`
	WarehouseID    *uint64 `json:"warehouse_id"`
	Name           string  `json:"name" validate:"required"`
	Code           *string `json:"code"`
	ParentID       *uint64 `json:"parent_id"`
	Usage          string  `json:"usage" validate:"required"`
	Barcode        *string `json:"barcode"`
}

type UpdateStockLocationResponse struct {
	Location StockLocationResponse `json:"location"`
}

type UpdateStockLocationResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateStockLocationResponse `json:"data"`
}

// @Summary Update stock location
// @Description Updates an existing stock location by its id after confirming it belongs to the caller's organization, otherwise a 404 Not Found is returned. The usage type is re-validated, and any parent location or warehouse must still belong to the same organization. Returns the updated location.
// @Tags Stock Locations
// @Accept json
// @Produce json
// @Param id path integer true "Stock location ID"
// @Param body body UpdateStockLocationRequest true "Stock location details"
// @Success 200 {object} UpdateStockLocationResponseEnvelope "Stock location updated successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock location not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-locations/{id} [put]
func (h WarehouseHandler) UpdateLocation(c fiber.Ctx) error {
	locationID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock location id provided.", nil)
	}

	var request UpdateStockLocationRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	location, err := h.svc.FindLocation(c, locationID)
	if err != nil {
		httpx.RequestLog(c).Error("stock location lookup failed", "location_id", locationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to update stock location.", err)
	}
	if location == nil || !httpx.OwnsTenant(c, location.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock location not found.")
	}

	location.OrganizationID = httpx.TenantOrganizationID(c, location.OrganizationID)
	location.WarehouseID = request.WarehouseID
	location.Name = request.Name
	location.Code = request.Code
	location.ParentID = request.ParentID
	location.Usage = request.Usage
	location.Barcode = request.Barcode

	updated, err := h.svc.UpdateLocation(c, location)
	if err != nil {
		return writeWarehouseError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Stock location updated successfully.", UpdateStockLocationResponse{
		Location: newStockLocationResponse(updated),
	})
}

// @Summary Delete stock location
// @Description Deletes a stock location by its id. The location is removed only when it exists and belongs to the caller's organization; otherwise a 404 Not Found is returned so the record's existence is not revealed. Responds with 204 No Content on success.
// @Tags Stock Locations
// @Accept json
// @Produce json
// @Param id path integer true "Stock location ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock location not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-locations/{id} [delete]
func (h WarehouseHandler) DeleteLocation(c fiber.Ctx) error {
	locationID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock location id provided.", nil)
	}

	location, err := h.svc.FindLocation(c, locationID)
	if err != nil {
		httpx.RequestLog(c).Error("stock location lookup failed", "location_id", locationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete stock location.", err)
	}
	if location == nil || !httpx.OwnsTenant(c, location.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock location not found.")
	}

	if err := h.svc.DeleteLocation(c, locationID); err != nil {
		httpx.RequestLog(c).Error("stock location deletion failed", "location_id", locationID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete stock location.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
