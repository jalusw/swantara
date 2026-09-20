package handler

import (
	"github.com/jalusw/swantara/apps/service/internal/interorganization"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type DropShipHandler struct {
	svc interorganization.DropShipService
}

func NewDropShipHandler(svc interorganization.DropShipService) DropShipHandler {
	return DropShipHandler{svc: svc}
}

type DropshipLinkResponse struct {
	ID                  uint64  `json:"id"`
	SaleOrderLineID     uint64  `json:"sale_order_line_id"`
	PurchaseOrderLineID uint64  `json:"purchase_order_line_id"`
	StockMovementID     *uint64 `json:"stock_movement_id"`
}

func newDropshipLinkResponse(link *interorganization.DropshipLink) DropshipLinkResponse {
	return DropshipLinkResponse{
		ID: link.ID, SaleOrderLineID: link.SaleOrderLineID,
		PurchaseOrderLineID: link.PurchaseOrderLineID, StockMovementID: link.StockMovementID,
	}
}

type DropshipOrderResponse struct {
	ID             uint64  `json:"id"`
	Name           *string `json:"name"`
	OrganizationID *uint64 `json:"organization_id"`
	SupplierID     uint64  `json:"supplier_id"`
	DestLocationID *uint64 `json:"dest_location_id"`
	State          string  `json:"state"`
	ReceiptStatus  string  `json:"receipt_status"`
	AmountTotal    float64 `json:"amount_total"`
}

func newDropshipOrderResponse(order *procurement.PurchaseOrder) DropshipOrderResponse {
	return DropshipOrderResponse{
		ID: order.ID, Name: order.Name, OrganizationID: order.OrganizationID,
		SupplierID: order.SupplierID, DestLocationID: order.DestLocationID,
		State: order.State, ReceiptStatus: order.ReceiptStatus, AmountTotal: order.AmountTotal,
	}
}
