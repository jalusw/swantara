package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListPOSOrdersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPOSOrdersResponse `json:"data"`
}
type ListPOSOrdersResponse struct {
	Orders []POSOrderResponse `json:"orders"`
}

// @Summary List POS orders
// @Description Lists POS orders with pagination, sorting, and filtering, scoped to the caller's organization.
// @Tags POS Orders
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated)"
// @Param filter query string false "Filters (repeatable)"
// @Success 200 {object} ListPOSOrdersResponseEnvelope "Orders retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/pos/orders [get]
func (h POSOrderHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, posOrderQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	page, err := h.svc.ListOrders(c, organizationID, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("pos order list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve POS orders.", err)
	}
	items := make([]POSOrderResponse, len(page.Items))
	for i, order := range page.Items {
		items[i] = newPOSOrderResponse(order)
	}
	return httpx.CreateSuccessResponseWithMeta(c, "POS orders retrieved successfully.", ListPOSOrdersResponse{
		Orders: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
