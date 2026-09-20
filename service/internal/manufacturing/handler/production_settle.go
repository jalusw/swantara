package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type SettleVarianceRequest struct {
	JournalID         uint64 `json:"journal_id" validate:"required,gt=0"`
	WIPAccountID      uint64 `json:"wip_account_id" validate:"required,gt=0"`
	VarianceAccountID uint64 `json:"variance_account_id" validate:"required,gt=0"`
	Date              string `json:"date"`
}

// @Summary Settle manufacturing variance
// @Description Settles the manufacturing variance for an in-progress production order by comparing the actual WIP balance booked against the standard cost applied during production. Any residual balance is posted as a manufacturing variance entry, and the order is then moved to the done state with its finished date stamped.
// @Tags Work Orders
// @Accept json
// @Produce json
// @Param id path integer true "Production order ID"
// @Param body body SettleVarianceRequest true "Settlement details"
// @Success 200 {object} GetProductionOrderResponseEnvelope "Manufacturing variance settled successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Production order not found"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or unknown references"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/manufacturing-orders/{id}/settle [post]
func (h ProductionHandler) SettleVariance(c fiber.Ctx) error {
	productionOrderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid production order id provided.", nil)
	}

	var request SettleVarianceRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := helper.ParseDateOrToday(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	productionOrder, err := h.svc.SettleVariance(c, productionOrderID, request.JournalID, request.WIPAccountID, request.VarianceAccountID, date)
	if err != nil {
		return writeProductionError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Manufacturing variance settled successfully.", GetProductionOrderResponse{
		ProductionOrder: newProductionOrderResponse(productionOrder),
	})
}
