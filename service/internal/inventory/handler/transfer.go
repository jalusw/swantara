package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/inventory"
)

type TransferHandler struct {
	svc inventory.TransferService
}

func NewTransferHandler(svc inventory.TransferService) TransferHandler {
	return TransferHandler{svc: svc}
}

type WarehouseTransferResponse struct {
	ID                  uint64     `json:"id"`
	OrganizationID      *uint64    `json:"organization_id"`
	Name                *string    `json:"name"`
	SrcWarehouseID      uint64     `json:"src_warehouse_id"`
	DstWarehouseID      uint64     `json:"dst_warehouse_id"`
	State               string     `json:"state"`
	OutShipmentID       *uint64    `json:"out_shipment_id"`
	InShipmentID        *uint64    `json:"in_shipment_id"`
	IsInterorganization bool       `json:"is_interorganization"`
	ScheduledDate       *time.Time `json:"scheduled_date"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func newWarehouseTransferResponse(transfer *inventory.WarehouseTransfer) WarehouseTransferResponse {
	return WarehouseTransferResponse{
		ID:                  transfer.ID,
		OrganizationID:      transfer.OrganizationID,
		Name:                transfer.Name,
		SrcWarehouseID:      transfer.SrcWarehouseID,
		DstWarehouseID:      transfer.DstWarehouseID,
		State:               transfer.State,
		OutShipmentID:       transfer.OutShipmentID,
		InShipmentID:        transfer.InShipmentID,
		IsInterorganization: transfer.IsInterorganization,
		ScheduledDate:       transfer.ScheduledDate,
		CreatedAt:           transfer.CreatedAt,
		UpdatedAt:           transfer.UpdatedAt,
	}
}

var transferOrderQueryAllowlist = map[string]struct{}{
	"organization_id":  {},
	"name":             {},
	"src_warehouse_id": {},
	"dst_warehouse_id": {},
	"state":            {},
	"created_at":       {},
	"updated_at":       {},
}

type ListWarehouseTransfersResponse struct {
	WarehouseTransfers []WarehouseTransferResponse `json:"warehouse_transfers"`
}

type ListWarehouseTransfersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListWarehouseTransfersResponse `json:"data"`
}

type GetWarehouseTransferResponse struct {
	WarehouseTransfer WarehouseTransferResponse `json:"warehouse_transfer"`
}

type GetWarehouseTransferResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetWarehouseTransferResponse `json:"data"`
}

type TransferLineRequest struct {
	ItemID        uint64  `json:"item_id" validate:"required,gt=0"`
	Qty           float64 `json:"qty" validate:"required,gt=0"`
	BatchID       *uint64 `json:"batch_id"`
	SrcLocationID *uint64 `json:"src_location_id"`
	DstLocationID *uint64 `json:"dst_location_id"`
}

type CreateWarehouseTransferRequest struct {
	OrganizationID *uint64               `json:"organization_id"`
	Name           *string               `json:"name"`
	SrcWarehouseID uint64                `json:"src_warehouse_id" validate:"required,gt=0"`
	DstWarehouseID uint64                `json:"dst_warehouse_id" validate:"required,gt=0"`
	ScheduledDate  *string               `json:"scheduled_date"`
	Lines          []TransferLineRequest `json:"lines"`
}

type CreateWarehouseTransferResponse struct {
	WarehouseTransfer WarehouseTransferResponse `json:"warehouse_transfer"`
}

type CreateWarehouseTransferResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateWarehouseTransferResponse `json:"data"`
}

type TransferActionRequest struct {
	JournalID        uint64 `json:"journal_id" validate:"required,gt=0"`
	TransitAccountID uint64 `json:"transit_account_id" validate:"required,gt=0"`
	Date             string `json:"date"`
}
