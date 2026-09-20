package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/returns"
)

type CreateRMALineRequest struct {
	ItemID      uint64  `json:"item_id" validate:"required,gt=0"`
	Qty         float64 `json:"qty" validate:"required,gt=0"`
	BatchID     *uint64 `json:"batch_id"`
	Disposition string  `json:"disposition" validate:"required"`
}

type CreateRMARequest struct {
	OrganizationID  *uint64                `json:"organization_id"`
	Type            string                 `json:"type" validate:"required"`
	ContactID       uint64                 `json:"contact_id" validate:"required,gt=0"`
	OriginOrderType string                 `json:"origin_order_type" validate:"required"`
	OriginOrderID   uint64                 `json:"origin_order_id" validate:"required,gt=0"`
	Reason          string                 `json:"reason"`
	Lines           []CreateRMALineRequest `json:"lines" validate:"required,min=1"`
}

type CreateRMAResponse struct {
	RMA RMAResponse `json:"rma"`
}

type CreateRMAResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateRMAResponse `json:"data"`
}

// @Summary Create RMA
// @Description Creates a draft return merchandise authorization for either a customer return or a supplier return, referencing a confirmed or done customer or supplier order. The origin order's contact must match the request, each line needs a valid item, positive quantity, and a recognized disposition (restock, scrap, repair, or replace), and a sequential RMA number is assigned.
// @Tags RMAs
// @Accept json
// @Produce json
// @Param body body CreateRMARequest true "RMA details"
// @Success 201 {object} CreateRMAResponseEnvelope "RMA created successfully."
// @Failure 400 {object} httpx.ErrorResponse "Bad request"
// @Failure 401 {object} httpx.ErrorResponse "Unauthorized"
// @Failure 404 {object} httpx.ErrorResponse "Not found"
// @Failure 409 {object} httpx.ErrorResponse "Conflict"
// @Failure 422 {object} httpx.ErrorResponse "Validation error"
// @Failure 500 {object} httpx.ErrorResponse "Internal server error"
// @Security BearerAuth
// @Router /organizations/{organization_id}/rmas [post]
func (h RMAHandler) Create(c fiber.Ctx) error {
	var request CreateRMARequest
	if !httpx.BindAndValidate(c, &request) {
		return nil
	}

	organizationID := httpx.TenantOrganizationID(c, request.OrganizationID)
	if organizationID == nil {
		return httpx.CreateUnprocessableEntityErrorResponse(c, "Unable to resolve organization.", nil)
	}

	lines := make([]returns.RMALineRequest, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = returns.RMALineRequest{
			ItemID:      line.ItemID,
			Qty:         line.Qty,
			BatchID:     line.BatchID,
			Disposition: line.Disposition,
		}
	}

	rma, err := h.svc.Create(c, returns.CreateRMARequest{
		OrganizationID:  *organizationID,
		Type:            request.Type,
		ContactID:       request.ContactID,
		OriginOrderType: request.OriginOrderType,
		OriginOrderID:   request.OriginOrderID,
		Reason:          request.Reason,
		Lines:           lines,
	})
	if err != nil {
		return writeRMAError(c, err)
	}

	return httpx.CreateCreatedResponse(c, "RMA created successfully.", CreateRMAResponse{
		RMA: newRMAResponse(rma),
	})
}
