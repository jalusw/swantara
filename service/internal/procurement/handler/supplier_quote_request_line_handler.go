package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type SupplierQuoteRequestLineResponse struct {
	ID             uint64     `json:"id"`
	QuoteRequestID uint64     `json:"quoteRequest_id"`
	ItemID         *uint64    `json:"item_id"`
	Description    *string    `json:"description"`
	Qty            float64    `json:"qty"`
	UnitID         *uint64    `json:"unit_id"`
	NeededBy       *time.Time `json:"needed_by"`
}

type ListSupplierQuoteRequestLinesResponse struct {
	Lines []SupplierQuoteRequestLineResponse `json:"lines"`
}

type ListSupplierQuoteRequestLinesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListSupplierQuoteRequestLinesResponse `json:"data"`
}

func newSupplierQuoteRequestLineResponse(line *procurement.SupplierQuoteRequestLine) SupplierQuoteRequestLineResponse {
	return SupplierQuoteRequestLineResponse{
		ID:             line.ID,
		QuoteRequestID: line.QuoteRequestID,
		ItemID:         line.ItemID,
		Description:    line.Description,
		Qty:            line.Qty,
		UnitID:         line.UnitID,
		NeededBy:       line.NeededBy,
	}
}

// @Summary List purchase QuoteRequest lines
// @Description Lists the request lines of a purchase QuoteRequest, including the requested products, quantities, units of measure, and needed-by dates. The lines are returned for the QuoteRequest identified by the path id.
// @Tags Purchase QuoteRequests
// @Accept json
// @Produce json
// @Param id path integer true "Purchase QuoteRequest ID"
// @Success 200 {object} ListSupplierQuoteRequestLinesResponseEnvelope "Purchase QuoteRequest lines retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Purchase QuoteRequest not found"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/purchase-quote_requests/{id}/lines [get]
func (h SupplierQuoteRequestHandler) ListLines(c fiber.Ctx) error {
	rfqID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid purchase QuoteRequest id provided.", nil)
	}
	lines, err := h.svc.ListLines(c, rfqID)
	if err != nil {
		return writePurchaseOrderError(c, err)
	}
	items := make([]SupplierQuoteRequestLineResponse, len(lines))
	for i, line := range lines {
		items[i] = newSupplierQuoteRequestLineResponse(line)
	}
	return httpx.CreateSuccessResponse(c, "Purchase QuoteRequest lines retrieved successfully.", ListSupplierQuoteRequestLinesResponse{
		Lines: items,
	})
}
