package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListAccountsResponse struct {
	Accounts []AccountResponse `json:"accounts"`
}

type ListAccountsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListAccountsResponse `json:"data"`
}

// @Summary List accounts
// @Description Lists chart of accounts entries with pagination, sorting, and filtering, automatically scoping results to the caller's organization; the response can be exported as JSON, XML, or CSV.
// @Tags Accounts
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. code:asc)"
// @Param filter query string false "Filters (repeatable, e.g. type:eq:asset)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListAccountsResponseEnvelope "Accounts retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/accounts [get]
func (h AccountHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, accountQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", err)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("account list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve accounts.", err)
	}

	items := make([]AccountResponse, len(page.Items))
	for i, account := range page.Items {
		items[i] = newAccountResponse(account)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "accounts.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Accounts retrieved successfully.", ListAccountsResponse{
		Accounts: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
