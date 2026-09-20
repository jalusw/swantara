package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
)

type StockMovementResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	ShipmentID     *uint64    `json:"shipment_id"`
	ItemID         uint64     `json:"item_id"`
	Qty            float64    `json:"qty"`
	UnitID         *uint64    `json:"unit_id"`
	SrcLocationID  uint64     `json:"src_location_id"`
	DstLocationID  uint64     `json:"dst_location_id"`
	BatchID        *uint64    `json:"batch_id"`
	State          string     `json:"state"`
	UnitCost       *float64   `json:"unit_cost"`
	OriginType     *string    `json:"origin_type"`
	OriginID       *uint64    `json:"origin_id"`
	ScheduledDate  *time.Time `json:"scheduled_date"`
	DateDone       *time.Time `json:"date_done"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newStockMovementResponse(movement *inventory.StockMovement) StockMovementResponse {
	return StockMovementResponse{
		ID:             movement.ID,
		OrganizationID: movement.OrganizationID,
		ShipmentID:     movement.ShipmentID,
		ItemID:         movement.ItemID,
		Qty:            movement.Qty,
		UnitID:         movement.UnitID,
		SrcLocationID:  movement.SrcLocationID,
		DstLocationID:  movement.DstLocationID,
		BatchID:        movement.BatchID,
		State:          movement.State,
		UnitCost:       movement.UnitCost,
		OriginType:     movement.OriginType,
		OriginID:       movement.OriginID,
		ScheduledDate:  movement.ScheduledDate,
		DateDone:       movement.DateDone,
		CreatedAt:      movement.CreatedAt,
		UpdatedAt:      movement.UpdatedAt,
	}
}

var stockMoveQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"shipment_id":     {},
	"item_id":         {},
	"src_location_id": {},
	"dst_location_id": {},
	"batch_id":        {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}

type ListStockMovementsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListStockMovementsResponse `json:"data"`
}
type ListStockMovementsResponse struct {
	Movements []StockMovementResponse `json:"movements"`
}

// @Summary List stock movements
// @Description Lists the stock movements belonging to the caller's organization, honoring pagination, sorting, and filtering. Only allowlisted query fields are accepted and the tenant filter is always enforced, so callers can never see another organization's movements. Results can be returned as a paginated JSON envelope or exported as a CSV file when the Accept header requests CSV.
// @Tags Stock Movements
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListStockMovementsResponseEnvelope "Stock movements retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-movements [get]
func (h StockHandler) ListMovements(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, stockMoveQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.stock.ListMovements(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("stock movement list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve stock movements.", err)
	}

	items := make([]StockMovementResponse, len(page.Items))
	for i, movement := range page.Items {
		items[i] = newStockMovementResponse(movement)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "stock-movements.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Stock movements retrieved successfully.", ListStockMovementsResponse{
		Movements: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}

type GetStockMovementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetStockMovementResponse `json:"data"`
}
type GetStockMovementResponse struct {
	Movement StockMovementResponse `json:"movement"`
}

// @Summary Get stock movement
// @Description Gets a single stock movement by its id, including its quantity, locations, state, and scheduling details. The movement is looked up and returned only when it belongs to the caller's organization; otherwise the request is answered with 404 Not Found to avoid leaking the record's existence. A non-numeric or malformed id is rejected with 422 Unprocessable Entity.
// @Tags Stock Movements
// @Accept json
// @Produce json
// @Param id path integer true "Stock movement ID"
// @Success 200 {object} GetStockMovementResponseEnvelope "Stock movement retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock movement not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-movements/{id} [get]
func (h StockHandler) GetMovement(c fiber.Ctx) error {
	movementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock movement id provided.", nil)
	}

	movement, err := h.stock.FindMovement(c, movementID)
	if err != nil {
		httpx.RequestLog(c).Error("stock movement lookup failed", "movement_id", movementID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get stock movement.", err)
	}
	if movement == nil || !httpx.OwnsTenant(c, movement.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock movement not found.")
	}

	return httpx.CreateSuccessResponse(c, "Stock movement retrieved successfully.", GetStockMovementResponse{
		Movement: newStockMovementResponse(movement),
	})
}

type CreateStockMovementRequest struct {
	OrganizationID *uint64  `json:"organization_id"`
	ItemID         uint64   `json:"item_id" validate:"required,gt=0"`
	Qty            string   `json:"qty" validate:"required"`
	UnitID         *uint64  `json:"unit_id"`
	SrcLocationID  uint64   `json:"src_location_id" validate:"required,gt=0"`
	DstLocationID  uint64   `json:"dst_location_id" validate:"required,gt=0"`
	BatchID        *uint64  `json:"batch_id"`
	UnitCost       *float64 `json:"unit_cost"`
	OriginType     *string  `json:"origin_type"`
	OriginID       *uint64  `json:"origin_id"`
	ScheduledDate  *string  `json:"scheduled_date"`
}

type CreateStockMovementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateStockMovementResponse `json:"data"`
}
type CreateStockMovementResponse struct {
	Movement StockMovementResponse `json:"movement"`
}

// @Summary Create stock movement
// @Description Creates a manual stock movement between a source and a destination stock location, leaving it in draft state for later processing. The request must specify a item, a positive decimal quantity, and both locations; the ledger service validates that the item is present and that the source and destination locations actually exist. Optional batch, unit cost, origin, and scheduled date fields can be attached. Returns the newly created draft movement with a 201 status.
// @Tags Stock Movements
// @Accept json
// @Produce json
// @Param body body CreateStockMovementRequest true "Stock movement details"
// @Success 201 {object} CreateStockMovementResponseEnvelope "Stock movement created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-movements [post]
func (h StockHandler) CreateMovement(c fiber.Ctx) error {
	var request CreateStockMovementRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	qty, err := amount.FromString(request.Qty)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Quantity must be a valid decimal.", nil)
	}

	scheduledDate, err := helper.ParseDate(request.ScheduledDate)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid scheduled date provided.", nil)
	}

	movement, err := h.ledger.CreateMovement(c, &inventory.StockMovement{
		OrganizationID: httpx.TenantOrganizationID(c, request.OrganizationID),
		ItemID:         request.ItemID,
		Qty:            qty.Float64(),
		UnitID:         request.UnitID,
		SrcLocationID:  request.SrcLocationID,
		DstLocationID:  request.DstLocationID,
		BatchID:        request.BatchID,
		UnitCost:       request.UnitCost,
		OriginType:     request.OriginType,
		OriginID:       request.OriginID,
		ScheduledDate:  scheduledDate,
	})
	if err != nil {
		return writeStockError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Stock movement created successfully.", CreateStockMovementResponse{
		Movement: newStockMovementResponse(movement),
	})
}

// @Summary Delete stock movement
// @Description Deletes a stock movement by its id. The movement is removed only when it exists and belongs to the caller's organization; otherwise a 404 Not Found is returned so the record's existence is not revealed. Responds with 204 No Content on success.
// @Tags Stock Movements
// @Accept json
// @Produce json
// @Param id path integer true "Stock movement ID"
// @Success 204 "No Content"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock movement not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-movements/{id} [delete]
func (h StockHandler) DeleteMovement(c fiber.Ctx) error {
	movementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock movement id provided.", nil)
	}

	movement, err := h.stock.FindMovement(c, movementID)
	if err != nil {
		httpx.RequestLog(c).Error("stock movement lookup failed", "movement_id", movementID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete stock movement.", err)
	}
	if movement == nil || !httpx.OwnsTenant(c, movement.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock movement not found.")
	}

	if err := h.stock.DeleteMovement(c, movementID); err != nil {
		httpx.RequestLog(c).Error("stock movement deletion failed", "movement_id", movementID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete stock movement.", err)
	}

	return httpx.CreateNoContentResponse(c)
}

type ReceiveMoveRequest struct {
	UnitCost  string `json:"unit_cost" validate:"required"`
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date"`
}

type PostStockMovementResponse struct {
	Layer inventory.CostLayer `json:"layer"`
}

type PostStockMovementResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PostStockMovementResponse `json:"data"`
}

// @Summary Receive stock movement
// @Description Receives a stock movement whose source is a supplier location, applying the movement and marking it done. The unit cost must be a non-negative decimal and a journal is required; the movement must not already be done or cancelled, batch tracking is enforced when the item uses it, and the item category must be configured with valuation and stock input accounts. A valuation layer is created at the given unit cost and a goods receipt journal entry is posted debiting Inventory and crediting Stock Input.
// @Tags Stock Movements
// @Accept json
// @Produce json
// @Param id path integer true "Stock movement ID"
// @Param request body ReceiveMoveRequest true "Receipt details"
// @Success 200 {object} PostStockMovementResponseEnvelope "Stock movement received and valued successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock movement not found"
// @Failure 409 {object} httpx.ErrorResponse "Stock movement cannot be processed in its current state"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-movements/{id}/receive [post]
func (h StockHandler) Receive(c fiber.Ctx) error {
	movementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock movement id provided.", nil)
	}

	var request ReceiveMoveRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	movement, err := h.stock.FindMovement(c, movementID)
	if err != nil {
		httpx.RequestLog(c).Error("stock movement lookup failed", "movement_id", movementID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to receive stock movement.", err)
	}
	if movement == nil || !httpx.OwnsTenant(c, movement.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock movement not found.")
	}

	unitCost, err := amount.FromString(request.UnitCost)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unit cost must be a valid decimal.", nil)
	}

	layerDate, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}

	layer, err := h.valuation.Receive(c, movementID, unitCost, request.JournalID, layerDate)
	if err != nil {
		return writeStockError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Stock movement received and valued successfully.", PostStockMovementResponse{
		Layer: *layer,
	})
}

type ShipMoveRequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date"`
}

// @Summary Ship stock movement
// @Description Ships a stock movement whose destination is a customer location, applying the movement and marking it done. The movement must not already be done or cancelled, batch tracking is enforced when the item uses it, and the item category must be configured with valuation and COGS accounts. Open valuation layers are consumed to compute the cost of goods sold, a negative valuation layer records the shipment, and a goods shipment journal entry is posted debiting COGS and crediting Inventory.
// @Tags Stock Movements
// @Accept json
// @Produce json
// @Param id path integer true "Stock movement ID"
// @Param request body ShipMoveRequest true "Shipment details"
// @Success 200 {object} PostStockMovementResponseEnvelope "Stock movement shipped and valued successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Stock movement not found"
// @Failure 409 {object} httpx.ErrorResponse "Stock movement cannot be processed in its current state"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/stock-movements/{id}/ship [post]
func (h StockHandler) Ship(c fiber.Ctx) error {
	movementID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid stock movement id provided.", nil)
	}

	var request ShipMoveRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	movement, err := h.stock.FindMovement(c, movementID)
	if err != nil {
		httpx.RequestLog(c).Error("stock movement lookup failed", "movement_id", movementID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to ship stock movement.", err)
	}
	if movement == nil || !httpx.OwnsTenant(c, movement.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Stock movement not found.")
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}

	layer, err := h.valuation.Ship(c, movementID, request.JournalID, date)
	if err != nil {
		return writeStockError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Stock movement shipped and valued successfully.", PostStockMovementResponse{
		Layer: *layer,
	})
}
