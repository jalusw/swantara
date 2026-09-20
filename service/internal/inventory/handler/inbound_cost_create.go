package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type CreateInboundCostLineRequest struct {
	ItemID             uint64  `json:"item_id" validate:"required,gt=0"`
	Description        string  `json:"description"`
	Amount             float64 `json:"amount" validate:"required,gt=0"`
	SupplierBillLineID *uint64 `json:"supplier_bill_line_id"`
	SplitMethod        string  `json:"split_method" validate:"required"`
	AccountID          *uint64 `json:"account_id" validate:"required"`
}

type CreateInboundCostRequest struct {
	Name              string                         `json:"name" validate:"required"`
	Date              *string                        `json:"date"`
	TargetShipmentIDs helper.Int64Array              `json:"target_shipment_ids" validate:"required,min=1"`
	Lines             []CreateInboundCostLineRequest `json:"lines" validate:"required,min=1"`
}

type CreateInboundCostResponseEnvelope struct {
	httpx.EnvelopeBase
	Data InboundCostResponse `json:"data"`
}

// @Summary Create inbound cost
// @Description Creates a draft inbound cost targeting receipt shipments, with lines to be split across the received products.
// @Tags Landed Costs
// @Accept json
// @Produce json
// @Param body body CreateInboundCostRequest true "Inbound cost details"
// @Success 201 {object} CreateInboundCostResponseEnvelope "Inbound cost created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Param organization_id path integer true "Organization ID"
// @Router /organizations/{organization_id}/landed-costs [post]
func (h InboundCostHandler) Create(c fiber.Ctx) error {
	var request CreateInboundCostRequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, nil)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Organization is required.", nil)
	}

	lines := make([]inventory.CreateInboundCostLineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = inventory.CreateInboundCostLineRequest{
			ItemID:             line.ItemID,
			Description:        line.Description,
			Amount:             line.Amount,
			SupplierBillLineID: line.SupplierBillLineID,
			SplitMethod:        line.SplitMethod,
			AccountID:          line.AccountID,
		}
	}

	date, err := helper.ParseDate(request.Date)
	if err != nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Invalid date provided.", nil)
	}

	cost, err := h.svc.Create(c, inventory.CreateInboundCostRequest{
		OrganizationID:    *organizationID,
		Name:              request.Name,
		Date:              date,
		TargetShipmentIDs: request.TargetShipmentIDs,
		Lines:             lines,
	})
	if err != nil {
		return writeInboundCostError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "Inbound cost created successfully.", CreateInboundCostResponseEnvelope{
		Data: newInboundCostResponse(cost),
	})
}
