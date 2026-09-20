package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

func (h JournalEntryHandler) Confirm(c fiber.Ctx) error {
	moveID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid account entry id provided.", nil)
	}
	organizationID, ok := httpx.CallerOrganizationID(c)
	if !ok {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}
	entry, err := h.poster.Confirm(c, organizationID, moveID)
	if err != nil {
		return writeJournalEntryError(c, err)
	}
	return httpx.CreateSuccessResponse(c, "Journal entry posted successfully.", PostJournalEntryResponse{
		Entry: newJournalEntryResponse(entry),
	})
}
