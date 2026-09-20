package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type RecordLaborRequest struct {
	Minutes               float64 `json:"minutes" validate:"required,gt=0"`
	JournalID             uint64  `json:"journal_id" validate:"required,gt=0"`
	WIPAccountID          uint64  `json:"wip_account_id" validate:"required,gt=0"`
	AppliedLaborAccountID uint64  `json:"applied_labor_account_id" validate:"required,gt=0"`
	Date                  string  `json:"date"`
}

type GetShopTaskResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetShopTaskResponse `json:"data"`
}
type GetShopTaskResponse struct {
	ShopTask ShopTaskResponse `json:"shop_task"`
}

// @Summary Record labor / overhead
// @Description Records actual labor and overhead minutes against a shop task of an in-progress production order, posting a Dr WIP / Cr Applied Labor journal entry computed from the work center's cost per hour. The shop task must belong to the given production order. The shop task's actual minutes and finished date are updated and it is marked done.
// @Tags Work Orders
// @Accept json
// @Produce json
// @Param production_order_id path integer true "Production order ID"
// @Param id path integer true "Shop task ID"
// @Param body body RecordLaborRequest true "Labor details"
// @Success 200 {object} GetShopTaskResponseEnvelope "Labor recorded successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Shop task not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown references"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{production_order_id}/work-orders/{id}/labor [post]
func (h ProductionHandler) RecordLabor(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("production_order_id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}
	workOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid shop task id provided.", nil)
	}

	var request RecordLaborRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	workOrder, err := h.svc.RecordLabor(c, productionOrderID, workOrderID, request.Minutes, request.JournalID, request.WIPAccountID, request.AppliedLaborAccountID, date)
	if err != nil {
		return writeProductionError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Labor recorded successfully.", GetShopTaskResponse{
		ShopTask: newShopTaskResponse(workOrder),
	})
}
