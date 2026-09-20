package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

// @Summary Confirm Planning planned order
// @Description Confirms a Planning planned order, generating the underlying purchase order or production order based on its type. Purchase planned orders require a supplier and price their line at the item's standard cost, while manufacture planned orders require an active manufacture recipe and create and confirm the production order. The planned order is marked confirmed with a reference to the generated document.
// @Tags Planning
// @Accept json
// @Produce json
// @Param id path integer true "Planned order ID"
// @Param body body ConfirmPlannedRequest true "Confirmation details"
// @Success 200 {object} PlannedSupplyResponseEnvelope "Planned order confirmed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 409 {object} httpx.ErrorResponse "Planned order already confirmed"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/planning/planned-orders/{id}/confirm [post]
func (h PlanningHandler) Confirm(c fiber.Ctx) error {
	plannedOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid planned order id.", nil)
	}
	var request ConfirmPlannedRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	order, err := h.svc.Confirm(c, plannedOrderID, request.SupplierID)
	if err != nil {
		return writeMrpError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Planned order confirmed successfully.", PlannedSupplyResponse{
		ID:               order.ID,
		PlanningRunID:    order.PlanningRunID,
		ItemID:           order.ItemID,
		WarehouseID:      order.WarehouseID,
		Type:             order.Type,
		Qty:              order.Qty,
		OrderDate:        order.OrderDate,
		DueDate:          order.DueDate,
		PeggedDemandID:   order.PeggedDemandID,
		Confirmed:        order.Confirmed,
		GeneratedDocType: order.GeneratedDocType,
		GeneratedDocID:   order.GeneratedDocID,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
	})
}

func writeMrpError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, manufacturing.ErrPlanningRunNotRequired):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Planning run horizon must be greater than zero.", nil)
	case errors.Is(err, manufacturing.ErrPlanningItem):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Planning item variant does not exist.", nil)
	case errors.Is(err, manufacturing.ErrPlanningNoRecipe):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Planning manufacture planned order requires a recipe.", nil)
	case errors.Is(err, manufacturing.ErrPlanningPlannedType):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Planning planned order type is not supported.", nil)
	case errors.Is(err, manufacturing.ErrPlanningPlannedState):
		return httpx.CreateConflictResponse(c, "Planning planned order is already confirmed.", err)
	case errors.Is(err, manufacturing.ErrPlanningSupplier):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Planning purchase planned order requires a supplier.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderOrganization):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order requires an organization.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production order location does not exist.", nil)
	case errors.Is(err, manufacturing.ErrProductionLocation):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Production location does not exist.", nil)
	case errors.Is(err, manufacturing.ErrProductionOrderNotFound):
		return httpx.CreateNotFoundResponse(c, "Planning run or planned order not found.")
	default:
		httpx.RequestLog(c).Error("planning write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process Planning operation.", err)
	}
}
