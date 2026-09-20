package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type ProductionOrderResponse struct {
	ID                uint64     `json:"id"`
	OrganizationID    *uint64    `json:"organization_id"`
	Name              *string    `json:"name"`
	ItemID            uint64     `json:"item_id"`
	RecipeID          *uint64    `json:"recipe_id"`
	QtyToProduce      float64    `json:"qty_to_produce"`
	QtyProduced       float64    `json:"qty_produced"`
	UnitID            *uint64    `json:"unit_id"`
	SrcLocationID     *uint64    `json:"src_location_id"`
	DstLocationID     *uint64    `json:"dst_location_id"`
	State             string     `json:"state"`
	DatePlannedStart  *time.Time `json:"date_planned_start"`
	DatePlannedFinish *time.Time `json:"date_planned_finish"`
	DateStart         *time.Time `json:"date_start"`
	DateFinished      *time.Time `json:"date_finished"`
	Origin            *string    `json:"origin"`
	Priority          int        `json:"priority"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func newProductionOrderResponse(productionOrder *manufacturing.ProductionOrder) ProductionOrderResponse {
	return ProductionOrderResponse{
		ID:                productionOrder.ID,
		OrganizationID:    productionOrder.OrganizationID,
		Name:              productionOrder.Name,
		ItemID:            productionOrder.ItemID,
		RecipeID:          productionOrder.RecipeID,
		QtyToProduce:      productionOrder.QtyToProduce,
		QtyProduced:       productionOrder.QtyProduced,
		UnitID:            productionOrder.UnitID,
		SrcLocationID:     productionOrder.SrcLocationID,
		DstLocationID:     productionOrder.DstLocationID,
		State:             productionOrder.State,
		DatePlannedStart:  productionOrder.DatePlannedStart,
		DatePlannedFinish: productionOrder.DatePlannedFinish,
		DateStart:         productionOrder.DateStart,
		DateFinished:      productionOrder.DateFinished,
		Origin:            productionOrder.Origin,
		Priority:          productionOrder.Priority,
		CreatedAt:         productionOrder.CreatedAt,
		UpdatedAt:         productionOrder.UpdatedAt,
	}
}

type ConsumedMaterialResponse struct {
	ID                uint64    `json:"id"`
	ProductionOrderID uint64    `json:"production_order_id"`
	ItemID            uint64    `json:"item_id"`
	QtyPlanned        float64   `json:"qty_planned"`
	QtyConsumed       float64   `json:"qty_consumed"`
	UnitID            *uint64   `json:"unit_id"`
	StockMovementID   *uint64   `json:"stock_movement_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func newConsumedMaterialResponse(component *manufacturing.ConsumedMaterial) ConsumedMaterialResponse {
	return ConsumedMaterialResponse{
		ID:                component.ID,
		ProductionOrderID: component.ProductionOrderID,
		ItemID:            component.ItemID,
		QtyPlanned:        component.QtyPlanned,
		QtyConsumed:       component.QtyConsumed,
		UnitID:            component.UnitID,
		StockMovementID:   component.StockMovementID,
		CreatedAt:         component.CreatedAt,
		UpdatedAt:         component.UpdatedAt,
	}
}

var manufacturingOrderQueryAllowlist = map[string]struct{}{
	"organization_id":     {},
	"item_id":             {},
	"recipe_id":           {},
	"state":               {},
	"priority":            {},
	"date_planned_start":  {},
	"date_planned_finish": {},
	"created_at":          {},
	"updated_at":          {},
}

type ListProductionOrdersResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListProductionOrdersResponse `json:"data"`
}
type ListProductionOrdersResponse struct {
	ProductionOrders []ProductionOrderResponse `json:"production_orders"`
}

type GetProductionOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data GetProductionOrderResponse `json:"data"`
}
type GetProductionOrderResponse struct {
	ProductionOrder ProductionOrderResponse `json:"production_order"`
}

type CreateProductionOrderRequest struct {
	OrganizationID    *uint64    `json:"organization_id"`
	ItemID            uint64     `json:"item_id" validate:"required,gt=0"`
	RecipeID          *uint64    `json:"recipe_id" validate:"required,gt=0"`
	QtyToProduce      float64    `json:"qty_to_produce" validate:"required,gt=0"`
	UnitID            *uint64    `json:"unit_id"`
	SrcLocationID     *uint64    `json:"src_location_id" validate:"required,gt=0"`
	DstLocationID     *uint64    `json:"dst_location_id" validate:"required,gt=0"`
	DatePlannedStart  *time.Time `json:"date_planned_start"`
	DatePlannedFinish *time.Time `json:"date_planned_finish"`
	Origin            *string    `json:"origin"`
	Priority          int        `json:"priority"`
}

type CreateProductionOrderResponseEnvelope struct {
	httpx.EnvelopeBase
	Data CreateProductionOrderResponse `json:"data"`
}
type CreateProductionOrderResponse struct {
	ProductionOrder ProductionOrderResponse    `json:"production_order"`
	Components      []ConsumedMaterialResponse `json:"components"`
}

type ListConsumedMaterialsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListConsumedMaterialsResponse `json:"data"`
}
type ListConsumedMaterialsResponse struct {
	Components []ConsumedMaterialResponse `json:"components"`
}
