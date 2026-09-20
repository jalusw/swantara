package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/kernel/amount"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type UnitHandler struct {
	svc reference.UnitService
}

func NewUnitHandler(svc reference.UnitService) UnitHandler {
	return UnitHandler{svc: svc}
}

type UnitResponse struct {
	ID         uint64    `json:"id"`
	CategoryID uint64    `json:"category_id"`
	Name       string    `json:"name"`
	Factor     float64   `json:"factor"`
	UnitType   string    `json:"unit_type"`
	Rounding   float64   `json:"rounding"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func newUnitResponse(unit *reference.Unit) UnitResponse {
	return UnitResponse{
		ID:         unit.ID,
		CategoryID: unit.CategoryID,
		Name:       unit.Name,
		Factor:     unit.Factor,
		UnitType:   unit.UnitType,
		Rounding:   unit.Rounding,
		CreatedAt:  unit.CreatedAt,
		UpdatedAt:  unit.UpdatedAt,
	}
}

var uomQueryAllowlist = map[string]struct{}{
	"category_id": {},
	"name":        {},
	"unit_type":   {},
	"created_at":  {},
	"updated_at":  {},
}

type ListUnitsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListUnitsResponse `json:"data"`
}
type ListUnitsResponse struct {
	Units []UnitResponse `json:"units"`
}

type GetUnitResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetUnitResponse `json:"data"`
}
type GetUnitResponse struct {
	Unit UnitResponse `json:"unit"`
}

type CreateUnitRequest struct {
	CategoryID uint64  `json:"category_id" validate:"required,gt=0"`
	Name       string  `json:"name" validate:"required"`
	Factor     float64 `json:"factor" validate:"required,gt=0"`
	UnitType   string  `json:"unit_type"`
	Rounding   float64 `json:"rounding"`
}

type CreateUnitResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateUnitResponse `json:"data"`
}
type CreateUnitResponse struct {
	Unit UnitResponse `json:"unit"`
}

type UpdateUnitRequest struct {
	Name     string  `json:"name" validate:"required"`
	Factor   float64 `json:"factor" validate:"required,gt=0"`
	UnitType string  `json:"unit_type"`
	Rounding float64 `json:"rounding"`
}

type UpdateUnitResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateUnitResponse `json:"data"`
}
type UpdateUnitResponse struct {
	Unit UnitResponse `json:"unit"`
}

type ConvertUnitRequest struct {
	FromID uint64 `json:"from_id" validate:"required,gt=0"`
	ToID   uint64 `json:"to_id" validate:"required,gt=0"`
	Qty    string `json:"qty" validate:"required"`
}

type ConvertUnitResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ConvertUnitResponse `json:"data"`
}
type ConvertUnitResponse struct {
	Value amount.Amount `json:"value"`
}
