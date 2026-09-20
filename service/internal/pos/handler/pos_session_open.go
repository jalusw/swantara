package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type OpenPOSSessionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data OpenPOSSessionResponse `json:"data"`
}
type OpenPOSSessionResponse struct {
	Session POSSessionResponse `json:"session"`
}

// @Summary Open POS session
// @Description Opens a POS session against a config with the cashier's opening balance. Only one session per config may be open at a time.
// @Tags POS Sessions
// @Accept json
// @Produce json
// @Param body body OpenPOSSessionRequest true "Session opening details"
// @Success 201 {object} OpenPOSSessionResponseEnvelope "POS session opened successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or config not found"
// @Failure 409 {object} httpx.ErrorResponse "A session is already open"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/sessions [post]
func (h POSSessionHandler) Open(c fiber.Ctx) error {
	var request OpenPOSSessionRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}
	session, err := h.svc.OpenSession(c, request.ConfigID, request.CashierID, request.OpeningBalance)
	if err != nil {
		return writePOSError(c, err)
	}
	return httpx.CreateCreatedResponse(c, "POS session opened successfully.", OpenPOSSessionResponse{
		Session: newPOSSessionResponse(session),
	})
}
