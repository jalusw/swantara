package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetJournalEntryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetJournalEntryResponse `json:"data"`
}
type GetJournalEntryResponse struct {
	Entry JournalEntryResponse  `json:"entry"`
	Lines []JournalLineResponse `json:"lines"`
}

// @Summary Get account entry
// @Description Gets a single account entry by its id together with its journal entry lines, returning 404 when the entry does not exist or belongs to another tenant. The response pairs the entry header with the full set of debit and credit lines that compose the entry.
// @Tags Journal Entries
// @Accept json
// @Produce json
// @Param id path integer true "Account entry ID"
// @Success 200 {object} GetJournalEntryResponseEnvelope "Account entry retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Account entry not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/account-movements/{id} [get]
func (h JournalEntryHandler) Get(c fiber.Ctx) error {
	entryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid account entry id provided.", nil)
	}

	entry, err := h.poster.FindEntry(c, entryID)
	if err != nil {
		httpx.RequestLog(c).Error("account entry lookup failed", "entry_id", entryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get account entry.", err)
	}
	if entry == nil || !httpx.OwnsTenant(c, &entry.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Account entry not found.")
	}

	lines, err := h.poster.ListLinesByMove(c, entry.ID)
	if err != nil {
		httpx.RequestLog(c).Error("account entry lines lookup failed", "entry_id", entry.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get account entry lines.", err)
	}
	lineItems := make([]JournalLineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newJournalLineResponse(line)
	}

	return httpx.CreateSuccessResponse(c, "Account entry retrieved successfully.", GetJournalEntryResponse{
		Entry: newJournalEntryResponse(entry),
		Lines: lineItems,
	})
}
