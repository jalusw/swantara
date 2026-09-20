package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

// @Summary Delete journal
// @Description Deletes an accounting journal by id. The journal must belong to the caller's organization; a 404 is returned otherwise.
// @Tags Journals
// @Accept json
// @Produce json
// @Param id path integer true "Journal ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Journal not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/journals/{id} [delete]
func (h JournalHandler) Delete(c fiber.Ctx) error {
	journalID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid journal id provided.", nil)
	}

	journal, err := h.svc.Find(c, journalID)
	if err != nil {
		httpx.RequestLog(c).Error("journal lookup failed", "journal_id", journalID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete journal.", err)
	}
	if journal == nil || !httpx.OwnsTenant(c, &journal.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "Journal not found.")
	}

	if err := h.svc.Delete(c, journalID); err != nil {
		httpx.RequestLog(c).Error("journal deletion failed", "journal_id", journalID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to delete journal.", err)
	}

	return httpx.CreateNoContentResponse(c)
}
