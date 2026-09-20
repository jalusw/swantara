package manufacturing

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type Recipe struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	ItemID         uint64  `gorm:"not null" json:"item_id"`
	Code           *string `json:"code"`
	Qty            float64 `gorm:"type:numeric(18,4);default:1" json:"qty"`
	UnitID         *uint64 `json:"unit_id"`
	Type           string  `gorm:"type:text" json:"type"`
	Version        int     `gorm:"default:1" json:"version"`
	Active         bool    `gorm:"default:true" json:"active"`
}

type RecipeLine struct {
	model.Base
	RecipeID         uint64  `gorm:"not null" json:"recipe_id"`
	ComponentID      uint64  `gorm:"not null" json:"component_id"`
	Qty              float64 `gorm:"type:numeric(18,4);not null" json:"qty"`
	UnitID           *uint64 `json:"unit_id"`
	ScrapPct         float64 `gorm:"type:numeric(8,4);default:0" json:"scrap_pct"`
	ProductionStepID *uint64 `json:"production_step_id"`
}

const (
	RecipeTypeManufacture = "manufacture"
	RecipeTypeKit         = "kit"
	RecipeTypeSubcontract = "subcontract"
)

var validRecipeType = map[string]bool{
	RecipeTypeManufacture: true,
	RecipeTypeKit:         true,
	RecipeTypeSubcontract: true,
}

const (
	ProductionOrderStateDraft      = "draft"
	ProductionOrderStateConfirmed  = "confirmed"
	ProductionOrderStatePlanned    = "planned"
	ProductionOrderStateInProgress = "in_progress"
	ProductionOrderStateDone       = "done"
	ProductionOrderStateCancelled  = "cancelled"

	ProductionOrderOriginManual    = "manual"
	ProductionOrderOriginSaleOrder = "sale_order"
	ProductionOrderOriginReorder   = "reorder"

	SequenceProductionOrderCode = "production_order"
)

const (
	OutsideProcessingStateDraft     = "draft"
	OutsideProcessingStateSent      = "sent"
	OutsideProcessingStateReceived  = "received"
	OutsideProcessingStateDone      = "done"
	OutsideProcessingStateCancelled = "cancelled"
)

type OutsideProcessingOrder struct {
	model.Base
	ProductionOrderID uint64  `gorm:"not null" json:"production_order_id"`
	SupplierID        uint64  `gorm:"not null" json:"supplier_id"`
	PurchaseOrderID   *uint64 `json:"purchase_order_id"`
	State             string  `gorm:"type:text;default:draft" json:"state"`
}

type ProductionOrder struct {
	model.Base
	OrganizationID    *uint64    `json:"organization_id"`
	Name              *string    `json:"name"`
	ItemID            uint64     `gorm:"not null" json:"item_id"`
	RecipeID          *uint64    `json:"recipe_id"`
	QtyToProduce      float64    `gorm:"type:numeric(18,4);not null" json:"qty_to_produce"`
	QtyProduced       float64    `gorm:"type:numeric(18,4);default:0" json:"qty_produced"`
	UnitID            *uint64    `json:"unit_id"`
	SrcLocationID     *uint64    `json:"src_location_id"`
	DstLocationID     *uint64    `json:"dst_location_id"`
	State             string     `gorm:"type:text" json:"state"`
	DatePlannedStart  *time.Time `json:"date_planned_start"`
	DatePlannedFinish *time.Time `json:"date_planned_finish"`
	DateStart         *time.Time `json:"date_start"`
	DateFinished      *time.Time `json:"date_finished"`
	Origin            *string    `json:"origin"`
	Priority          int        `gorm:"default:0" json:"priority"`
}

type ConsumedMaterial struct {
	model.Base
	ProductionOrderID uint64  `gorm:"not null" json:"production_order_id"`
	ItemID            uint64  `gorm:"not null" json:"item_id"`
	QtyPlanned        float64 `gorm:"type:numeric(18,4)" json:"qty_planned"`
	QtyConsumed       float64 `gorm:"type:numeric(18,4);default:0" json:"qty_consumed"`
	UnitID            *uint64 `json:"unit_id"`
	StockMovementID   *uint64 `json:"stock_movement_id"`
}

func (ConsumedMaterial) TableName() string {
	return "consumed_materials"
}

type ProductionStep struct {
	model.Base
	RecipeID     *uint64 `json:"recipe_id"`
	WorkCenterID *uint64 `json:"work_center_id"`
	Name         *string `json:"name"`
	Sequence     int     `gorm:"default:0" json:"sequence"`
	SetupMinutes float64 `gorm:"type:numeric(12,2);default:0" json:"setup_minutes"`
	TimeMinutes  float64 `gorm:"type:numeric(12,2)" json:"time_minutes"`
}

const (
	ShopTaskStateDraft      = "draft"
	ShopTaskStateConfirmed  = "confirmed"
	ShopTaskStatePlanned    = "planned"
	ShopTaskStateInProgress = "in_progress"
	ShopTaskStateDone       = "done"
	ShopTaskStateCancelled  = "cancelled"

	SequenceShopTaskCode = "work_order"
)

type ShopTask struct {
	model.Base
	OrganizationID    *uint64    `json:"organization_id"`
	ProductionOrderID uint64     `gorm:"not null" json:"production_order_id"`
	ProductionStepID  *uint64    `json:"production_step_id"`
	WorkCenterID      uint64     `gorm:"not null" json:"work_center_id"`
	Name              *string    `json:"name"`
	State             string     `gorm:"type:text" json:"state"`
	Sequence          int        `gorm:"default:0" json:"sequence"`
	PlannedStart      *time.Time `json:"planned_start"`
	PlannedFinish     *time.Time `json:"planned_finish"`
	DateStart         *time.Time `json:"date_start"`
	DateFinished      *time.Time `json:"date_finished"`
	PlannedMinutes    float64    `gorm:"type:numeric(12,2);default:0" json:"planned_minutes"`
	ActualMinutes     float64    `gorm:"type:numeric(12,2);default:0" json:"actual_minutes"`
}

const (
	PlanningRunStateRunning = "running"
	PlanningRunStateDone    = "done"
	PlanningRunStateFailed  = "failed"

	PlanningNeedSourceSaleOrder = "sale_order"
	PlanningNeedSourceForecast  = "forecast"
	PlanningNeedSourceReorder   = "reorder"
	PlanningNeedSourceRecipe    = "recipe"

	PlannedSupplyTypePurchase    = "purchase"
	PlannedSupplyTypeManufacture = "manufacture"
	PlannedSupplyTypeTransfer    = "transfer"
)

type PlanningRun struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	RunDate        *time.Time `gorm:"type:date" json:"run_date"`
	HorizonDays    int        `json:"horizon_days"`
	State          string     `gorm:"type:text" json:"state"`
}

type PlanningNeed struct {
	model.Base
	PlanningRunID uint64     `gorm:"not null" json:"planning_run_id"`
	ItemID        uint64     `gorm:"not null" json:"item_id"`
	WarehouseID   *uint64    `json:"warehouse_id"`
	SourceType    string     `gorm:"type:text" json:"source_type"`
	SourceID      uint64     `json:"source_id"`
	Qty           float64    `gorm:"type:numeric(18,4)" json:"qty"`
	RequiredDate  *time.Time `gorm:"type:date" json:"required_date"`
}

type PlannedSupply struct {
	model.Base
	PlanningRunID    uint64     `gorm:"not null" json:"planning_run_id"`
	ItemID           uint64     `gorm:"not null" json:"item_id"`
	WarehouseID      *uint64    `json:"warehouse_id"`
	SrcWarehouseID   *uint64    `json:"src_warehouse_id"`
	Type             string     `gorm:"type:text" json:"type"`
	Qty              float64    `gorm:"type:numeric(18,4)" json:"qty"`
	OrderDate        *time.Time `gorm:"type:date" json:"order_date"`
	DueDate          *time.Time `gorm:"type:date" json:"due_date"`
	PeggedDemandID   *uint64    `json:"pegged_demand_id"`
	Confirmed        bool       `gorm:"default:false" json:"confirmed"`
	GeneratedDocType *string    `json:"generated_doc_type"`
	GeneratedDocID   *uint64    `json:"generated_doc_id"`
}

type DemandPlan struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	ItemID         uint64     `gorm:"not null" json:"item_id"`
	WarehouseID    *uint64    `json:"warehouse_id"`
	PeriodStart    *time.Time `gorm:"type:date" json:"period_start"`
	PeriodEnd      *time.Time `gorm:"type:date" json:"period_end"`
	ForecastQty    float64    `gorm:"type:numeric(18,4)" json:"forecast_qty"`
}
