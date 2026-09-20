package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListPOSSessionsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPOSSessionsResponse `json:"data"`
}
type ListPOSSessionsResponse struct {
	Sessions []POSSessionResponse `json:"sessions"`
}

// @Summary List POS sessions
// @Description Lists POS sessions with pagination, sorting, and filtering, scoped to the caller's organization.
// @Tags POS Sessions
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListPOSSessionsResponseEnvelope "Sessions retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/sessions [get]
func (h POSSessionHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, posSessionQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	page, err := h.svc.ListSessions(c, organizationID, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("pos session list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve POS sessions.", err)
	}
	items := make([]POSSessionResponse, len(page.Items))
	for i, session := range page.Items {
		items[i] = newPOSSessionResponse(session)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "POS sessions retrieved successfully.", ListPOSSessionsResponse{
		Sessions: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
