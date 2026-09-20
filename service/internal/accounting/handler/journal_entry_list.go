package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ListJournalEntrysResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListJournalEntrysResponse `json:"data"`
}
type ListJournalEntrysResponse struct {
	Moves []JournalEntryResponse `json:"movements"`
}

// @Summary List account movements
// @Description Lists account movements (journal entries) for the caller's tenant, applying pagination, sorting, and filtering over fields such as journal, state, date, origin, and reference. Results are scoped to the organization of the authenticated tenant and may be exported as CSV, JSON, or XML.
// @Tags Journal Entries
// @Accept json
// @Produce json
// @Param page query integer false "Page number" default(1)
// @Param size query integer false "Items per page (max 100)" default(20)
// @Param sort query string false "Sort fields (comma separated, e.g. date:asc)"
// @Param filter query string false "Filters (repeatable, e.g. journal_id:eq:1)"
// @Param format query string false "Response format" Enums(json, xml, csv)
// @Success 200 {object} ListJournalEntrysResponseEnvelope "Account movements retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/account-movements [get]
func (h JournalEntryHandler) List(c fiber.Ctx) error {
	parsedQuery, err := httpx.ParseQueryParams(c, accountMoveQueryAllowlist)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid query parameters.", nil)
	}
	if err := httpx.ForceTenantFilter(c, parsedQuery); err != nil {
		return httpx.CreateUnauthorizedErrorResponse(c, "Unauthorized.", nil)
	}

	page, err := h.poster.ListEntries(c, parsedQuery)
	if err != nil {
		httpx.RequestLog(c).Error("account entry list failed", "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to retrieve account movements.", err)
	}

	items := make([]JournalEntryResponse, len(page.Items))
	for i, entry := range page.Items {
		items[i] = newJournalEntryResponse(entry)
	}

	if httpx.RequestFormat(c) == httpx.FormatCSV {
		return httpx.ExportCSV(c, fiber.StatusOK, "account-movements.csv", items)
	}

	return httpx.CreateSuccessResponseWithMeta(c, "Account movements retrieved successfully.", ListJournalEntrysResponse{
		Moves: items,
	}, httpx.BuildListMeta(parsedQuery, page.Count))
}
