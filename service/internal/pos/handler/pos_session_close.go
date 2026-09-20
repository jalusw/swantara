package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ClosePOSSessionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ClosePOSSessionResponse `json:"data"`
}
type ClosePOSSessionResponse struct {
	Session POSSessionResponse `json:"session"`
}

// @Summary Close POS session
// @Description Transitions a session from opened to closing, awaiting cash reconciliation.
// @Tags POS Sessions
// @Accept json
// @Produce json
// @Param id path integer true "POS session ID"
// @Success 200 {object} ClosePOSSessionResponseEnvelope "POS session closing started."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "POS session not found"
// @Failure 409 {object} httpx.ErrorResponse "Session cannot be closed in its current state"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/sessions/{id}/closing [post]
func (h POSSessionHandler) StartClosing(c fiber.Ctx) error {
	sessionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid POS session id provided.", nil)
	}
	session, err := h.svc.StartClosing(c, sessionID)
	if err != nil {
		return writePOSError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "POS session closing started.", ClosePOSSessionResponse{
		Session: newPOSSessionResponse(session),
	})
}

// @Summary Close POS session
// @Description Reconciles the session cash drawer against the opening balance and recorded cash payments, then closes it. The closing balance must equal the opening balance plus all cash payments recorded in the session; non-cash payments (card and similar) are not part of the drawer.
// @Tags POS Sessions
// @Accept json
// @Produce json
// @Param id path integer true "POS session ID"
// @Param body body ClosePOSSessionRequest true "Closing balance"
// @Success 200 {object} ClosePOSSessionResponseEnvelope "POS session closed successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "POS session not found"
// @Failure 409 {object} httpx.ErrorResponse "Session cannot be closed in its current state or cash does not reconcile"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/sessions/{id}/close [post]
func (h POSSessionHandler) Close(c fiber.Ctx) error {
	sessionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid POS session id provided.", nil)
	}
	var request ClosePOSSessionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	session, err := h.svc.CloseSession(c, sessionID, request.ClosingBalance)
	if err != nil {
		return writePOSError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "POS session closed successfully.", ClosePOSSessionResponse{
		Session: newPOSSessionResponse(session),
	})
}
