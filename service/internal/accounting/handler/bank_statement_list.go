package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

var bankStatementQueryAllowlist = map[string]struct{}{
	"journal_id":    {},
	"name":          {},
	"date":          {},
	"state":         {},
	"balance_start": {},
	"balance_end":   {},
	"created_at":    {},
	"updated_at":    {},
}

type ListBankStatementsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListBankStatementsResponse `json:"data"`
}
type ListBankStatementsResponse struct {
	BankStatements []BankStatementResponse `json:"bank_statements"`
}

// @Summary List bank statements
// @Description Lists bank statements for the caller's tenant, applying pagination, sorting, and filtering over fields such as journal, date, state, and balances. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Bank Statements
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. date:desc)"
// @Param filter query string false "Filters (repeatable, e.g. state:eq:open)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListBankStatementsResponseEnvelope "Bank statements retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/bank-statements [get]
func (h BankStatementHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, bankStatementQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.svc.List(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("bank statement list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve bank statements.", err)
	}

	items := make([]BankStatementResponse, len(page.Items))
	for i, statement := range page.Items {
		items[i] = newBankStatementResponse(statement)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "bank-statements.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Bank statements retrieved successfully.", ListBankStatementsResponse{
		BankStatements: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
