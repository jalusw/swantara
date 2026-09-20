package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type GetRMAResponse struct {
	RMA   RMAResponse       `json:"rma"`
	Lines []RMALineResponse `json:"lines"`
}

type GetRMAResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetRMAResponse `json:"data"`
}

// @Summary Get RMA
// @Description Gets a single return merchandise authorization together with all of its return lines, including each line's disposition, stock movement, and credit note references. RMAs not belonging to the caller's organization are treated as not found.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param id path integer true "RMA ID"
// @Success 200 {object} GetRMAResponseEnvelope "RMA retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "RMA not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas/{id} [get]
func (h RMAHandler) Get(c fiber.Ctx) error {
	rmaID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid rma id provided.", nil)
	}

	rma, err := h.svc.Find(c, rmaID)
	if err != nil {
		httpx.RequestLog(c).Error("rma lookup failed", "rma_id", rmaID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get rma.", err)
	}
	if rma == nil || !httpx.OwnsTenant(c, rma.OrganizationID) {
		return httpx.CreateNotFoundResponse(c, "RMA not found.")
	}

	lines, err := h.svc.ListLines(c, rma.ID)
	if err != nil {
		httpx.RequestLog(c).Error("rma lines lookup failed", "rma_id", rma.ID, "error", err)
		return httpx.CreateInternalServerErrorResponse(c, "Failed to get rma lines.", err)
	}
	lineItems := make([]RMALineResponse, len(lines))
	for i, line := range lines {
		lineItems[i] = newRMALineResponse(line)
	}

	return httpx.CreateSuccessResponse(c, "RMA retrieved successfully.", GetRMAResponse{
		RMA:   newRMAResponse(rma),
		Lines: lineItems,
	})
}
