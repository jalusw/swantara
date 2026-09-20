package handler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

// @Summary Confirm purchase requisition
// @Description Confirms a purchase requisition, transitioning it from draft to confirmed. The state machine rejects the transition from any other state, so only draft requisitions can be confirmed.
// @Tags Purchase Requisitions
// @Accept json
// @Produce json
// @Param id path integer true "Purchase requisition ID"
// @Success 200 {object} GetPurchaseRequestResponseEnvelope "Purchase requisition confirmed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase requisition not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-requisitions/{id}/confirm [post]
func (h PurchaseRequestHandler) Confirm(c fiber.Ctx) error {
	requestID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase requisition id provided.", nil)
	}
	requisition, err := h.svc.Confirm(c, requestID)
	if err != nil {
		return writeRequisitionError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase requisition confirmed successfully.", GetPurchaseRequestResponse{
		Request: newPurchaseRequestResponse(requisition),
	})
}

// @Summary Approve purchase requisition
// @Description Approves a confirmed purchase requisition, moving it to the approved state from which it can be converted into a purchase order or purchase QuoteRequest. The state machine rejects the transition from any state other than confirmed.
// @Tags Purchase Requisitions
// @Accept json
// @Produce json
// @Param id path integer true "Purchase requisition ID"
// @Success 200 {object} GetPurchaseRequestResponseEnvelope "Purchase requisition approved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase requisition not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-requisitions/{id}/approve [post]
func (h PurchaseRequestHandler) Approve(c fiber.Ctx) error {
	requestID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase requisition id provided.", nil)
	}
	requisition, err := h.svc.Approve(c, requestID)
	if err != nil {
		return writeRequisitionError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase requisition approved successfully.", GetPurchaseRequestResponse{
		Request: newPurchaseRequestResponse(requisition),
	})
}

// @Summary Cancel purchase requisition
// @Description Cancels a purchase requisition that is in draft or confirmed state, moving it to cancelled before it is approved. The state machine rejects cancellation from any other state.
// @Tags Purchase Requisitions
// @Accept json
// @Produce json
// @Param id path integer true "Purchase requisition ID"
// @Success 200 {object} GetPurchaseRequestResponseEnvelope "Purchase requisition cancelled successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase requisition not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-requisitions/{id}/cancel [post]
func (h PurchaseRequestHandler) Cancel(c fiber.Ctx) error {
	requestID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase requisition id provided.", nil)
	}
	requisition, err := h.svc.Cancel(c, requestID)
	if err != nil {
		return writeRequisitionError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Purchase requisition cancelled successfully.", GetPurchaseRequestResponse{
		Request: newPurchaseRequestResponse(requisition),
	})
}

func writeRequisitionError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, procurement.ErrRequisitionNotFound):
		return httpx.CreateNotFoundResponse(c, "Purchase requisition not found.")
	case errors.Is(err, procurement.ErrRequestState):
		return httpx.CreateConflictResponse(c, "Purchase requisition cannot be processed in its current state.", err)
	case errors.Is(err, procurement.ErrRequisitionNoLines):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase requisition must have at least one line.", nil)
	case errors.Is(err, procurement.ErrRequisitionLineQty):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase requisition line quantity must be positive.", nil)
	case errors.Is(err, procurement.ErrRequisitionRequester):
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Purchase requisition requester does not exist.", nil)
	default:
		httpx.RequestLog(c).Error("purchase requisition write failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to process purchase requisition.", err)
	}
}
