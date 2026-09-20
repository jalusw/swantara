package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type ProductionHandler struct {
	svc manufacturing.ProductionService
}

func NewProductionHandler(svc manufacturing.ProductionService) ProductionHandler {
	return ProductionHandler{svc: svc}
}

type ShopTaskResponse struct {
	ID                uint64     `json:"id"`
	OrganizationID    *uint64    `json:"organization_id"`
	ProductionOrderID uint64     `json:"production_order_id"`
	ProductionStepID  *uint64    `json:"production_step_id"`
	WorkCenterID      uint64     `json:"work_center_id"`
	Name              *string    `json:"name"`
	State             string     `json:"state"`
	Sequence          int        `json:"sequence"`
	PlannedStart      *time.Time `json:"planned_start"`
	PlannedFinish     *time.Time `json:"planned_finish"`
	DateStart         *time.Time `json:"date_start"`
	DateFinished      *time.Time `json:"date_finished"`
	PlannedMinutes    float64    `json:"planned_minutes"`
	ActualMinutes     float64    `json:"actual_minutes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func newShopTaskResponse(workOrder *manufacturing.ShopTask) ShopTaskResponse {
	return ShopTaskResponse{
		ID:                workOrder.ID,
		OrganizationID:    workOrder.OrganizationID,
		ProductionOrderID: workOrder.ProductionOrderID,
		ProductionStepID:  workOrder.ProductionStepID,
		WorkCenterID:      workOrder.WorkCenterID,
		Name:              workOrder.Name,
		State:             workOrder.State,
		Sequence:          workOrder.Sequence,
		PlannedStart:      workOrder.PlannedStart,
		PlannedFinish:     workOrder.PlannedFinish,
		DateStart:         workOrder.DateStart,
		DateFinished:      workOrder.DateFinished,
		PlannedMinutes:    workOrder.PlannedMinutes,
		ActualMinutes:     workOrder.ActualMinutes,
		CreatedAt:         workOrder.CreatedAt,
		UpdatedAt:         workOrder.UpdatedAt,
	}
}

// @Summary Start production order
// @Description Starts a planned production order, transitioning it to the in_progress state and stamping the actual start date. Only planned orders can be started; any other state is rejected with a conflict.
// @Tags Work Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Success 200 {object} GetProductionOrderResponseEnvelope "Production order started successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/start [post]
func (h ProductionHandler) Start(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	productionOrder, err := h.svc.Start(c, productionOrderID)
	if err != nil {
		return writeProductionOrderError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Production order started successfully.", GetProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
	})
}
