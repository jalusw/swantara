package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

var budgetQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"date_start":      {},
	"date_end":        {},
	"state":           {},
	"created_at":      {},
	"updated_at":      {},
}

// @Summary List budgets
// @Description Lists budgets for the caller's tenant, applying pagination, sorting, and filtering over fields such as name, date range, and state. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Budgets
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. name:asc)"
// @Param filter query string false "Filters (repeatable, e.g. state:eq:confirmed)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListBudgetsResponseEnvelope "Budgets retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/budgets [get]
func (h BudgetHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, budgetQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("budget list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve budgets.", err)
	}

	items := make([]BudgetResponse, len(page.Items))
	for i, budget := range page.Items {
		items[i] = newBudgetResponse(budget)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "budgets.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Budgets retrieved successfully.", ListBudgetsResponse{
		Budgets: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
