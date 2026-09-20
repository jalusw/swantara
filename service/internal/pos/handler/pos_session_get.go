package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetPOSSessionResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPOSSessionResponse `json:"data"`
}
type GetPOSSessionResponse struct {
	Session          POSSessionResponse         `json:"session"`
	PaymentBreakdown []POSPaymentMethodResponse `json:"payment_breakdown"`
}

// @Summary Get POS session
// @Description Returns a single POS session with its per-method payment breakdown. The session must belong to the caller's organization, otherwise a 404 is returned.
// @Tags POS Sessions
// @Accept json
// @Produce json
// @Param id path integer true "POS session ID"
// @Success 200 {object} GetPOSSessionResponseEnvelope "Session retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "POS session not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/sessions/{id} [get]
func (h POSSessionHandler) Get(c fiber.Ctx) error {
	sessionID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid POS session id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	session, err := h.svc.GetSession(c, organizationID, sessionID)
	if err != nil {
		return writePOSError(c, err)
	}
	breakdown, err := h.svc.SessionPaymentBreakdown(c, session.ID)
	if err != nil {
		httpx.RequestLog(c).Error("pos session breakdown failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve POS session payment breakdown.", err)
	}
	items := make([]POSPaymentMethodResponse, 0, len(breakdown))
	for method, total := range breakdown {
		items = append(items, POSPaymentMethodResponse{Method: method, Amount: total})
	}
	return httpx.CreateSuccessResponse(c, "POS session retrieved successfully.", GetPOSSessionResponse{
		Session:          newPOSSessionResponse(session),
		PaymentBreakdown: items,
	})
}
