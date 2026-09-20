package handler

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/httpx"
	"github.com/jalusw/swantara/apps/service/internal/manufacturing"
)

type PlanningRunResponse struct {
	ID             uint64                  `json:"id"`
	OrganizationID *uint64                 `json:"organization_id"`
	RunDate        *time.Time              `json:"run_date"`
	HorizonDays    int                     `json:"horizon_days"`
	State          string                  `json:"state"`
	Demands        []PlanningNeedResponse  `json:"demands,omitempty"`
	PlannedOrders  []PlannedSupplyResponse `json:"planned_orders,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
}

type PlanningNeedResponse struct {
	ID            uint64     `json:"id"`
	PlanningRunID uint64     `json:"planning_run_id"`
	ItemID        uint64     `json:"item_id"`
	WarehouseID   *uint64    `json:"warehouse_id"`
	SourceType    string     `json:"source_type"`
	SourceID      uint64     `json:"source_id"`
	Qty           float64    `json:"qty"`
	RequiredDate  *time.Time `json:"required_date"`
}

type PlannedSupplyResponse struct {
	ID               uint64     `json:"id"`
	PlanningRunID    uint64     `json:"planning_run_id"`
	ItemID           uint64     `json:"item_id"`
	WarehouseID      *uint64    `json:"warehouse_id"`
	Type             string     `json:"type"`
	Qty              float64    `json:"qty"`
	OrderDate        *time.Time `json:"order_date"`
	DueDate          *time.Time `json:"due_date"`
	PeggedDemandID   *uint64    `json:"pegged_demand_id"`
	Confirmed        bool       `json:"confirmed"`
	GeneratedDocType *string    `json:"generated_doc_type"`
	GeneratedDocID   *uint64    `json:"generated_doc_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type DemandPlanResponse struct {
	ID             uint64     `json:"id"`
	OrganizationID *uint64    `json:"organization_id"`
	ItemID         uint64     `json:"item_id"`
	WarehouseID    *uint64    `json:"warehouse_id"`
	PeriodStart    *time.Time `json:"period_start"`
	PeriodEnd      *time.Time `json:"period_end"`
	ForecastQty    float64    `json:"forecast_qty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ListPlanningRunsResponse struct {
	Runs []PlanningRunResponse `json:"runs"`
}

type ListDemandPlansResponse struct {
	Forecasts []DemandPlanResponse `json:"forecasts"`
}

type PlanningRunResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PlanningRunResponse `json:"data"`
}

type PlannedSupplyResponseEnvelope struct {
	httpx.EnvelopeBase
	Data PlannedSupplyResponse `json:"data"`
}

type DemandPlanResponseEnvelope struct {
	httpx.EnvelopeBase
	Data DemandPlanResponse `json:"data"`
}

type ListPlanningRunsResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListPlanningRunsResponse `json:"data"`
}

type ListDemandPlansResponseEnvelope struct {
	httpx.EnvelopeBase
	Data ListDemandPlansResponse `json:"data"`
}

type RunMrpRequest struct {
	HorizonDays int    `json:"horizon_days" validate:"required,gt=0"`
	RunDate     string `json:"run_date"`
}

type ConfirmPlannedRequest struct {
	SupplierID uint64 `json:"supplier_id"`
}

type CreateForecastRequest struct {
	ItemID      uint64  `json:"item_id" validate:"required,gt=0"`
	WarehouseID *uint64 `json:"warehouse_id"`
	PeriodStart string  `json:"period_start" validate:"required"`
	PeriodEnd   string  `json:"period_end" validate:"required"`
	ForecastQty float64 `json:"forecast_qty" validate:"required,gt=0"`
}

func newPlanningRunResponse(run *manufacturing.PlanningRun, demands []*manufacturing.PlanningNeed, planned []*manufacturing.PlannedSupply) PlanningRunResponse {
	response := PlanningRunResponse{
		ID:             run.ID,
		OrganizationID: run.OrganizationID,
		RunDate:        run.RunDate,
		HorizonDays:    run.HorizonDays,
		State:          run.State,
		CreatedAt:      run.CreatedAt,
		UpdatedAt:      run.UpdatedAt,
	}
	for _, demand := range demands {
		response.Demands = append(response.Demands, PlanningNeedResponse{
			ID:            demand.ID,
			PlanningRunID: demand.PlanningRunID,
			ItemID:        demand.ItemID,
			WarehouseID:   demand.WarehouseID,
			SourceType:    demand.SourceType,
			SourceID:      demand.SourceID,
			Qty:           demand.Qty,
			RequiredDate:  demand.RequiredDate,
		})
	}
	for _, order := range planned {
		response.PlannedOrders = append(response.PlannedOrders, PlannedSupplyResponse{
			ID:               order.ID,
			PlanningRunID:    order.PlanningRunID,
			ItemID:           order.ItemID,
			WarehouseID:      order.WarehouseID,
			Type:             order.Type,
			Qty:              order.Qty,
			OrderDate:        order.OrderDate,
			DueDate:          order.DueDate,
			PeggedDemandID:   order.PeggedDemandID,
			Confirmed:        order.Confirmed,
			GeneratedDocType: order.GeneratedDocType,
			GeneratedDocID:   order.GeneratedDocID,
			CreatedAt:        order.CreatedAt,
			UpdatedAt:        order.UpdatedAt,
		})
	}
	return response
}

func newDemandPlanResponse(forecast *manufacturing.DemandPlan) DemandPlanResponse {
	return DemandPlanResponse{
		ID:             forecast.ID,
		OrganizationID: forecast.OrganizationID,
		ItemID:         forecast.ItemID,
		WarehouseID:    forecast.WarehouseID,
		PeriodStart:    forecast.PeriodStart,
		PeriodEnd:      forecast.PeriodEnd,
		ForecastQty:    forecast.ForecastQty,
		CreatedAt:      forecast.CreatedAt,
		UpdatedAt:      forecast.UpdatedAt,
	}
}
