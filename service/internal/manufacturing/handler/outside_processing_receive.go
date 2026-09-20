package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ReceiveOutsideProcessingOrderRequest struct {
	JournalID          uint64  `json:"journal_id" validate:"required,gt=0"`
	WIPAccountID       uint64  `json:"wip_account_id" validate:"required,gt=0"`
	APPayableAccountID uint64  `json:"ap_payable_account_id" validate:"required,gt=0"`
	Date               *string `json:"date"`
}

// @Summary Receive outside processing order
// @Description Receives the subcontract finished goods: consumes the supplier purchase order by incrementing the received quantities, nets the accounts payable for the subcontract operation against work-in-progress, produces the finished goods into the destination location, and marks the production order done. The outside processing order movements to the received state.
// @Tags Subcontract Orders
// @Accept json
// @Produce json
// @Param id path integer true "Outside processing order ID"
// @Param body body ReceiveOutsideProcessingOrderRequest true "Receive details"
// @Success 200 {object} OutsideProcessingOrderEnvelope "Outside processing order received successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or subcontract operation amount missing"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subcontract-orders/{id}/receive [post]
func (h OutsideProcessingHandler) Receive(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid outside processing order id provided.", nil)
	}

	var request ReceiveOutsideProcessingOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := helper.ParseDateOrToday(helper.Deref(request.Date, ""))
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	order, err := h.svc.Receive(c, orderID, request.JournalID, request.WIPAccountID, request.APPayableAccountID, date)
	if err != nil {
		return writeSubcontractError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Outside processing order received successfully.", newOutsideProcessingOrderResponse(order))
}
