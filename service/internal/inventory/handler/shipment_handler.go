package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type ShipmentResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	Type           string     `json:"type"`
	ContactID      *uint64    `json:"contact_id"`
	SrcLocationID  *uint64    `json:"src_location_id"`
	DstLocationID  *uint64    `json:"dst_location_id"`
	State          string     `json:"state"`
	ScheduledDate  *time.Time `json:"scheduled_date"`
	DateDone       *time.Time `json:"date_done"`
	Origin         *string    `json:"origin"`
	CarrierID      *uint64    `json:"carrier_id"`
	TrackingRef    *string    `json:"tracking_ref"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newShipmentResponse(shipment *inventory.Shipment) ShipmentResponse {
	return ShipmentResponse{
		ID:             shipment.ID,
		OrganizationID: shipment.OrganizationID,
		Name:           shipment.Name,
		Type:           shipment.Type,
		ContactID:      shipment.ContactID,
		SrcLocationID:  shipment.SrcLocationID,
		DstLocationID:  shipment.DstLocationID,
		State:          shipment.State,
		ScheduledDate:  shipment.ScheduledDate,
		DateDone:       shipment.DateDone,
		Origin:         shipment.Origin,
		CarrierID:      shipment.CarrierID,
		TrackingRef:    shipment.TrackingRef,
		CreatedAt:      shipment.CreatedAt,
		UpdatedAt:      shipment.UpdatedAt,
	}
}

var stockShipmentQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"type":            {},
	"state":           {},
	"src_location_id": {},
	"dst_location_id": {},
	"created_at":      {},
	"updated_at":      {},
}

type ListShipmentsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListShipmentsResponse `json:"data"`
}
type ListShipmentsResponse struct {
	Shipments []ShipmentResponse `json:"shipments"`
}

// @Summary List stock shipments
// @Description Lists the stock shipments belonging to the caller's organization, honoring pagination, sorting, and filtering. Only allowlisted query fields are accepted and the tenant filter is always enforced, so callers can never see another organization's shipments. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Stock Shipments
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListShipmentsResponseEnvelope "Stock shipments retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-shipments [get]
func (h StockHandler) ListShipments(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, stockShipmentQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.stock.ListShipments(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("stock shipment list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve stock shipments.", err)
	}

	items := make([]ShipmentResponse, len(page.Items))
	for i, shipment := range page.Items {
		items[i] = newShipmentResponse(shipment)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "stock-shipments.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Stock shipments retrieved successfully.", ListShipmentsResponse{
		Shipments: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetShipmentResponse struct {
	Shipment ShipmentResponse `json:"shipment"`
}

type GetShipmentResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetShipmentResponse `json:"data"`
}

// @Summary Get stock shipment
// @Description Gets a single stock shipment by its id, including its type, state, locations, and scheduling details. The shipment is looked up and returned only when it belongs to the caller's organization; otherwise the request is answered with 404 Not Found to avoid leaking the record's existence. A non-numeric or malformed id is rejected with 422 Unprocessable Entity.
// @Tags Stock Shipments
// @Accept json
// @Produce json
// @Param id path integer true "Stock shipment ID"
// @Success 200 {object} GetShipmentResponseEnvelope "Stock shipment retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock shipment not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-shipments/{id} [get]
func (h StockHandler) GetShipment(c fiber.Ctx) error {
	shipmentID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock shipment id provided.", nil)
	}

	shipment, err := h.stock.FindShipment(c, shipmentID)
	if err != nil {
		httpx.RequestLog(c).Error("stock shipment lookup failed", "shipment_id", shipmentID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get stock shipment.", err)
	}
	if shipment == nil || !httpx.OwnsTenant(c, shipment.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock shipment not found.")
	}

	return httpx.CreateSuccessResponse(c, "Stock shipment retrieved successfully.", GetShipmentResponse{
		Shipment: newShipmentResponse(shipment),
	})
}
