package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type InboundCostHandler struct {
	svc inventory.InboundCostService
}

func NewInboundCostHandler(svc inventory.InboundCostService) InboundCostHandler {
	return InboundCostHandler{svc: svc}
}

type InboundCostResponse struct {
	ID                uint64            `json:"id"`
	Name              string            `json:"name"`
	Date              *time.Time        `json:"date"`
	State             string            `json:"state"`
	TargetShipmentIDs helper.Int64Array `json:"target_shipment_ids"`
	MovementID        *uint64           `json:"movement_id"`
}

func newInboundCostResponse(cost *inventory.InboundCost) InboundCostResponse {
	return InboundCostResponse{
		ID:                cost.ID,
		Name:              cost.Name,
		Date:              cost.Date,
		State:             cost.State,
		TargetShipmentIDs: cost.TargetShipmentIDs,
		MovementID:        cost.MovementID,
	}
}
