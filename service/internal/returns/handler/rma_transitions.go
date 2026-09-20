package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Confirm RMA
// @Description Confirms a draft RMA, moving it into the confirmed state so the returned stock can subsequently be received. Only RMAs still in draft can be confirmed; otherwise a conflict is returned.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param id path integer true "RMA ID"
// @Success 200 {object} TransitionResponseEnvelope "RMA confirmed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "RMA not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas/{id}/confirm [post]
func (h RMAHandler) Confirm(c fiber.Ctx) error {
	rmaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rma id provided.", nil)
	}

	rma, err := h.svc.Confirm(c, rmaID)
	if err != nil {
		return writeRMAError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "RMA confirmed successfully.", TransitionResponse{
		RMA: newRMAResponse(rma),
	})
}

type ReceiveRMARequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date" validate:"required"`
}

// @Summary Receive RMA
// @Description Receives the returned stock for a confirmed RMA by creating inbound stock movements that mirror the original order's movements. Depending on disposition and return type, the receipt either restocks inventory at the original unit cost, returns goods to the supplier, or routes them to a scrap location, and the RMA advances to received.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param id path integer true "RMA ID"
// @Param body body ReceiveRMARequest true "Receipt details"
// @Success 200 {object} TransitionResponseEnvelope "RMA received successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "RMA not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas/{id}/receive [post]
func (h RMAHandler) Receive(c fiber.Ctx) error {
	rmaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rma id provided.", nil)
	}

	var request ReceiveRMARequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(&request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	rma, err := h.svc.Receive(c, rmaID, request.JournalID, *date)
	if err != nil {
		return writeRMAError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "RMA received successfully.", TransitionResponse{
		RMA: newRMAResponse(rma),
	})
}

type RefundRMARequest struct {
	JournalID uint64 `json:"journal_id" validate:"required,gt=0"`
	Date      string `json:"date" validate:"required"`
	Reference string `json:"reference"`
}

// @Summary Refund RMA
// @Description Refunds a received RMA by creating a credit note against the original invoice for the returned quantities, recording the credit note on each return line, and moving the RMA to refunded. Requires an existing invoice on the origin order and a receipt prior to refunding.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param id path integer true "RMA ID"
// @Param body body RefundRMARequest true "Refund details"
// @Success 200 {object} TransitionResponseEnvelope "RMA refunded successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "RMA not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas/{id}/refund [post]
func (h RMAHandler) Refund(c fiber.Ctx) error {
	rmaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rma id provided.", nil)
	}

	var request RefundRMARequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	date, err := helper.ParseDate(&request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}

	rma, err := h.svc.Refund(c, rmaID, request.JournalID, *date, request.Reference)
	if err != nil {
		return writeRMAError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "RMA refunded successfully.", TransitionResponse{
		RMA: newRMAResponse(rma),
	})
}

// @Summary Complete RMA
// @Description Completes the return cycle by marking a refunded RMA as done. Only RMAs that have already been refunded can be completed, and invalid transitions are rejected.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param id path integer true "RMA ID"
// @Success 200 {object} TransitionResponseEnvelope "RMA completed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "RMA not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas/{id}/done [post]
func (h RMAHandler) Done(c fiber.Ctx) error {
	rmaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rma id provided.", nil)
	}

	rma, err := h.svc.Done(c, rmaID)
	if err != nil {
		return writeRMAError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "RMA completed successfully.", TransitionResponse{
		RMA: newRMAResponse(rma),
	})
}

// @Summary Cancel RMA
// @Description Cancels an RMA that is still in draft or confirmed, before the returned stock has been received, and movements it to the cancelled state. RMAs that have already been received or refunded cannot be cancelled.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param id path integer true "RMA ID"
// @Success 200 {object} TransitionResponseEnvelope "RMA cancelled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "RMA not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas/{id}/cancel [post]
func (h RMAHandler) Cancel(c fiber.Ctx) error {
	rmaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rma id provided.", nil)
	}

	rma, err := h.svc.Cancel(c, rmaID)
	if err != nil {
		return writeRMAError(c, err)
	}

	return httpx.CreateSuccessResponse(c, "RMA cancelled successfully.", TransitionResponse{
		RMA: newRMAResponse(rma),
	})
}
