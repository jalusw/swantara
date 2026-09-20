package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type OutsideProcessingHandler struct {
	svc manufacturing.OutsideProcessingService
}

func NewOutsideProcessingHandler(svc manufacturing.OutsideProcessingService) OutsideProcessingHandler {
	return OutsideProcessingHandler{svc: svc}
}

type OutsideProcessingOrderResponse struct {
	ID                uint64    `json:"id"`
	ProductionOrderID uint64    `json:"production_order_id"`
	SupplierID        uint64    `json:"supplier_id"`
	PurchaseOrderID   *uint64   `json:"purchase_order_id"`
	State             string    `json:"state"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func newOutsideProcessingOrderResponse(order *manufacturing.OutsideProcessingOrder) OutsideProcessingOrderResponse {
	return OutsideProcessingOrderResponse{
		ID:                order.ID,
		ProductionOrderID: order.ProductionOrderID,
		SupplierID:        order.SupplierID,
		PurchaseOrderID:   order.PurchaseOrderID,
		State:             order.State,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}
}

type OutsideProcessingOrderEnvelope struct {
	httpx.EnvelopeBase
	Data OutsideProcessingOrderResponse `json:"data"`
}
