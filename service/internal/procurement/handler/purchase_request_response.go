package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/procurement"
)

type PurchaseRequestResponse struct {
	ID             uint64                        `json:"id"`
	OrganizationID *uint64                       `json:"organization_id"`
	Name           *string                       `json:"name"`
	RequesterID    uint64                        `json:"requester_id"`
	DepartmentID   *uint64                       `json:"department_id"`
	State          string                        `json:"state"`
	NeededBy       *time.Time                    `json:"needed_by"`
	Lines          []PurchaseRequestLineResponse `json:"lines,omitempty"`
}

type PurchaseRequestLineResponse struct {
	ID          uint64     `json:"id"`
	RequestID   uint64     `json:"request_id"`
	ItemID      *uint64    `json:"item_id"`
	Description *string    `json:"description"`
	Qty         float64    `json:"qty"`
	UnitID      *uint64    `json:"unit_id"`
	NeededBy    *time.Time `json:"needed_by"`
}

func newPurchaseRequestResponse(req *procurement.PurchaseRequest) PurchaseRequestResponse {
	return PurchaseRequestResponse{
		ID:             req.ID,
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		RequesterID:    req.RequesterID,
		DepartmentID:   req.DepartmentID,
		State:          req.State,
		NeededBy:       req.NeededBy,
	}
}

func newPurchaseRequestLineResponse(line *procurement.PurchaseRequestLine) PurchaseRequestLineResponse {
	return PurchaseRequestLineResponse{
		ID:          line.ID,
		RequestID:   line.RequestID,
		ItemID:      line.ItemID,
		Description: line.Description,
		Qty:         line.Qty,
		UnitID:      line.UnitID,
		NeededBy:    line.NeededBy,
	}
}

var requisitionQueryAllowlist = map[string]struct{}{
	"organization_id": {},
	"name":            {},
	"requester_id":    {},
	"department_id":   {},
	"state":           {},
	"needed_by":       {},
	"created_at":      {},
	"updated_at":      {},
}

type ListPurchaseRequestsResponse struct {
	Requisitions []PurchaseRequestResponse `json:"requisitions"`
}

type ListPurchaseRequestsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPurchaseRequestsResponse `json:"data"`
}

type GetPurchaseRequestResponse struct {
	Request PurchaseRequestResponse `json:"request"`
}

type GetPurchaseRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetPurchaseRequestResponse `json:"data"`
}

type PurchaseRequestLineRequest struct {
	ItemID      *uint64 `json:"item_id"`
	Description *string `json:"description"`
	Qty         float64 `json:"qty" validate:"required,gt=0"`
	UnitID      *uint64 `json:"unit_id"`
	NeededBy    *string `json:"needed_by"`
}

type CreatePurchaseRequestRequest struct {
	OrganizationID *uint64                      `json:"organization_id"`
	RequesterID    uint64                       `json:"requester_id" validate:"required,gt=0"`
	DepartmentID   *uint64                      `json:"department_id"`
	NeededBy       *string                      `json:"needed_by"`
	Lines          []PurchaseRequestLineRequest `json:"lines" validate:"required,min=1"`
}

type CreatePurchaseRequestResponse struct {
	Request PurchaseRequestResponse `json:"request"`
}

type CreatePurchaseRequestResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreatePurchaseRequestResponse `json:"data"`
}
