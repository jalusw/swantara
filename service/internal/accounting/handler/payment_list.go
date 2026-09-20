package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

var paymentQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"contact_id":      {},
	"type":            {},
	"state":           {},
	"journal_id":      {},
	"date":            {},
	"name":            {},
	"reference":       {},
	"created_at":      {},
	"updated_at":      {},
}

// @Summary List payments
// @Description Lists inbound and outbound payments for the caller's tenant, applying pagination, sorting, and filtering over fields such as contact, type, state, journal, and date. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Payments
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. date:asc)"
// @Param filter query string false "Filters (repeatable, e.g. type:eq:inbound)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListPaymentsResponseEnvelope "Payments retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/payments [get]
func (h PaymentHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, paymentQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("payment list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve payments.", err)
	}

	items := make([]PaymentResponse, len(page.Items))
	for i, payment := range page.Items {
		items[i] = newPaymentResponse(payment)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "payments.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Payments retrieved successfully.", ListPaymentsResponse{
		Payments: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
