package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
)

type InboundCostLineResponse struct {
	ID                 uint64  `json:"id"`
	ItemID             uint64  `json:"item_id"`
	Description        *string `json:"description"`
	Amount             float64 `json:"amount"`
	SupplierBillLineID *uint64 `json:"supplier_bill_line_id"`
	SplitMethod        string  `json:"split_method"`
	AccountID          *uint64 `json:"account_id"`
}

type ListInboundCostLinesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListInboundCostLinesResponse `json:"data"`
}

type ListInboundCostLinesResponse struct {
	Lines []InboundCostLineResponse `json:"lines"`
}

// @Summary List inbound cost lines
// @Description Lists the lines of a inbound cost for the caller's organization.
// @Tags Landed Costs
// @Accept json
// @Produce json
// @Param id path integer true "Inbound cost ID"
// @Success 200 {object} ListInboundCostLinesResponseEnvelope "Inbound cost lines retrieved successfully."
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Inbound cost not found"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/landed-costs/{id}/lines [get]
func (h InboundCostHandler) ListLines(c fiber.Ctx) error {
	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	costID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid inbound cost id provided.", nil)
	}

	lines, err := h.svc.ListLines(c, *organizationID, costID)
	if err != nil {
		return writeInboundCostError(c, err)
	}

	items := make([]InboundCostLineResponse, len(lines))
	for i, line := range lines {
		items[i] = InboundCostLineResponse{
			ID:                 line.ID,
			ItemID:             line.ItemID,
			Description:        line.Description,
			Amount:             line.Amount,
			SupplierBillLineID: line.SupplierBillLineID,
			SplitMethod:        line.SplitMethod,
			AccountID:          line.AccountID,
		}
	}

	return httpx.CreateSuccessResponse(c, "Inbound cost lines retrieved successfully.", ListInboundCostLinesResponse{
		Lines: items,
	})
}
