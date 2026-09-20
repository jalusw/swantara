package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type SendOutsideProcessingOrderRequest struct {
	JournalID    uint64  `json:"journal_id" validate:"required,gt=0"`
	WIPAccountID uint64  `json:"wip_account_id" validate:"required,gt=0"`
	Date         *string `json:"date"`
}

// @Summary Send outside processing order
// @Description Sends the outside processing order to the supplier: creates a draft supplier purchase order priced from the supplier catalog, dispatches the production order components as outbound movements to the supplier virtual location, and posts the material issue to subcontract work-in-progress. The outside processing order movements to the sent state.
// @Tags Subcontract Orders
// @Accept json
// @Produce json
// @Param id path integer true "Outside processing order ID"
// @Param body body SendOutsideProcessingOrderRequest true "Send details"
// @Success 200 {object} OutsideProcessingOrderEnvelope "Outside processing order sent successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 409 {object} httpx.ErrorResponse "State transition not allowed"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or supplier catalog price missing"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/subcontract-orders/{id}/send [post]
func (h OutsideProcessingHandler) Send(c fiber.Ctx) error {
	orderID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid outside processing order id provided.", nil)
	}

	var request SendOutsideProcessingOrderRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	date, err := helper.ParseDateOrToday(helper.Deref(request.Date, ""))
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}
	order, err := h.svc.Send(c, orderID, request.JournalID, request.WIPAccountID, date)
	if err != nil {
		return writeSubcontractError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "Outside processing order sent successfully.", newOutsideProcessingOrderResponse(order))
}
