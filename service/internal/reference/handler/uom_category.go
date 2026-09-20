package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/reference"
)

type UnitGroupHandler struct {
	svc reference.UnitService
}

func NewUnitGroupHandler(svc reference.UnitService) UnitGroupHandler {
	return UnitGroupHandler{svc: svc}
}

type UnitGroupResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func newUnitGroupResponse(category *reference.UnitGroup) UnitGroupResponse {
	return UnitGroupResponse{
		ID:        category.ID,
		Name:      category.Name,
		CreatedAt: category.CreatedAt,
		UpdatedAt: category.UpdatedAt,
	}
}

var unitGroupQueryAllowlist = map[string]struct{}{
	"name":       {},
	"created_at": {},
	"updated_at": {},
}

type ListUnitCategoriesResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListUnitCategoriesResponse `json:"data"`
}
type ListUnitCategoriesResponse struct {
	Categories []UnitGroupResponse `json:"categories"`
}

type GetUnitGroupResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetUnitGroupResponse `json:"data"`
}
type GetUnitGroupResponse struct {
	Category UnitGroupResponse `json:"category"`
}

type CreateUnitGroupRequest struct {
	Name string `json:"name" validate:"required"`
}

type CreateUnitGroupResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateUnitGroupResponse `json:"data"`
}
type CreateUnitGroupResponse struct {
	Category UnitGroupResponse `json:"category"`
}

type UpdateUnitGroupRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateUnitGroupResponseEnvelope struct {
	httpx.EnvelopeBase
	Data UpdateUnitGroupResponse `json:"data"`
}
type UpdateUnitGroupResponse struct {
	Category UnitGroupResponse `json:"category"`
}
