package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/accounting"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type ReverseJournalEntryRequest struct {
	JournalID   uint64 `json:"journal_id" validate:"required,gt=0"`
	Date        string `json:"date"`
	Ref         string `json:"ref"`
	Description string `json:"description" validate:"required"`
}

type ReverseJournalEntryResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ReverseJournalEntryResponse `json:"data"`
}
type ReverseJournalEntryResponse struct {
	Entry JournalEntryResponse `json:"entry"`
}

// @Summary Reverse account entry
// @Description Reverses an already posted account entry by creating and posting a mirroring journal entry with debits and credits swapped, using the supplied reversal date or today when omitted. Only posted movements that have not already been reversed can be reversed, and the resulting reversal entry links back to the original entry.
// @Tags Journal Entries
// @Accept json
// @Produce json
// @Param id path integer true "Account entry ID"
// @Param body body ReverseJournalEntryRequest true "Reversal details"
// @Success 201 {object} ReverseJournalEntryResponseEnvelope "Account entry reversed successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Account entry not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error or entry is not reversible"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/account-movements/{id}/reverse [post]
func (h JournalEntryHandler) Reverse(c fiber.Ctx) error {
	entryID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid account entry id provided.", nil)
	}

	var request ReverseJournalEntryRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	entry, err := h.poster.FindEntry(c, entryID)
	if err != nil {
		httpx.RequestLog(c).Error("account entry lookup failed", "entry_id", entryID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to reverse account entry.", err)
	}
	if entry == nil || !httpx.OwnsTenant(c, &entry.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Account entry not found.")
	}

	date, err := helper.ParseDate(&request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Date must be in YYYY-MM-DD format.", nil)
	}
	reversalDate := time.Now().UTC()
	if date != nil && !date.IsZero() {
		reversalDate = *date
	}

	reversal, err := h.poster.Reverse(c, accounting.ReverseRequest{
		OrganizationID: entry.OrganizationID,
		JournalID:      request.JournalID,
		Date:           reversalDate,
		Ref:            request.Ref,
		Description:    request.Description,
		EntryID:        entry.ID,
	})
	if err != nil {
		return writeJournalEntryError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Account entry reversed successfully.", ReverseJournalEntryResponse{
		Entry: newJournalEntryResponse(reversal),
	})
}
