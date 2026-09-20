package inventory

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/helper"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	MovementStateDraft     = "draft"
	MovementStateConfirmed = "confirmed"
	MovementStateAssigned  = "assigned"
	MovementStateDone      = "done"
	MovementStateCancelled = "cancelled"

	ShipmentStateDraft     = "draft"
	ShipmentStateWaiting   = "waiting"
	ShipmentStateConfirmed = "confirmed"
	ShipmentStateAssigned  = "assigned"
	ShipmentStateDone      = "done"
	ShipmentStateCancelled = "cancelled"

	ShipmentTypeIncoming = "incoming"
	ShipmentTypeOutgoing = "outgoing"
	ShipmentTypeInternal = "internal"
)

type Shipment struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	Type           string     `gorm:"type:text" json:"type"`
	ContactID      *uint64    `json:"contact_id"`
	SrcLocationID  *uint64    `json:"src_location_id"`
	DstLocationID  *uint64    `json:"dst_location_id"`
	State          string     `gorm:"type:text" json:"state"`
	ScheduledDate  *time.Time `json:"scheduled_date"`
	DateDone       *time.Time `json:"date_done"`
	Origin         *string    `json:"origin"`
	CarrierID      *uint64    `json:"carrier_id"`
	TrackingRef    *string    `json:"tracking_ref"`
}

type StockMovement struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	ShipmentID     *uint64    `json:"shipment_id"`
	ItemID         uint64     `gorm:"not null" json:"item_id"`
	Qty            float64    `gorm:"type:numeric(18,4);not null" json:"qty"`
	UnitID         *uint64    `json:"unit_id"`
	SrcLocationID  uint64     `gorm:"not null" json:"src_location_id"`
	DstLocationID  uint64     `gorm:"not null" json:"dst_location_id"`
	BatchID        *uint64    `json:"batch_id"`
	State          string     `gorm:"type:text" json:"state"`
	UnitCost       *float64   `gorm:"type:numeric(18,4)" json:"unit_cost"`
	OriginType     *string    `json:"origin_type"`
	OriginID       *uint64    `json:"origin_id"`
	ScheduledDate  *time.Time `json:"scheduled_date"`
	DateDone       *time.Time `json:"date_done"`
}

type StockBalance struct {
	model.Base
	OrganizationID *uint64 `json:"organization_id"`
	ItemID         uint64  `gorm:"not null" json:"item_id"`
	LocationID     uint64  `gorm:"not null" json:"location_id"`
	BatchID        *uint64 `json:"batch_id"`
	Quantity       float64 `gorm:"type:numeric(18,4);default:0" json:"quantity"`
	ReservedQty    float64 `gorm:"type:numeric(18,4);default:0" json:"reserved_qty"`
}

type Batch struct {
	model.Base
	ItemID         uint64     `gorm:"not null" json:"item_id"`
	Name           string     `gorm:"not null" json:"name"`
	Ref            *string    `json:"ref"`
	ExpiryDate     *time.Time `gorm:"type:date" json:"expiry_date"`
	BestBeforeDate *time.Time `gorm:"type:date" json:"best_before_date"`
}

type StockHold struct {
	model.Base
	MovementID *uint64 `json:"movement_id"`
	BalanceID  uint64  `gorm:"not null" json:"balance_id"`
	Qty        float64 `gorm:"type:numeric(18,4)" json:"qty"`
}

type CostLayer struct {
	model.Base
	MovementID     *uint64  `json:"movement_id"`
	ItemID         uint64   `gorm:"not null" json:"item_id"`
	Quantity       float64  `gorm:"type:numeric(18,4)" json:"quantity"`
	UnitCost       *float64 `gorm:"type:numeric(18,4)" json:"unit_cost"`
	Value          float64  `gorm:"type:numeric(18,4)" json:"value"`
	RemainingQty   float64  `gorm:"type:numeric(18,4)" json:"remaining_qty"`
	RemainingValue float64  `gorm:"type:numeric(18,4)" json:"remaining_value"`
	JournalEntryID *uint64  `json:"journal_entry_id"`
	Description    *string  `json:"description"`
}

type ReorderRule struct {
	model.Base
	ItemID       uint64  `gorm:"not null" json:"item_id"`
	WarehouseID  *uint64 `json:"warehouse_id"`
	LocationID   *uint64 `json:"location_id"`
	MinQty       float64 `gorm:"type:numeric(18,4)" json:"min_qty"`
	MaxQty       float64 `gorm:"type:numeric(18,4)" json:"max_qty"`
	QtyMultiple  float64 `gorm:"type:numeric(18,4);default:1" json:"qty_multiple"`
	LeadTimeDays *int    `json:"lead_time_days"`
	Active       bool    `gorm:"default:true" json:"active"`
}

type StockCount struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           *string    `json:"name"`
	LocationID     *uint64    `json:"location_id"`
	State          string     `gorm:"type:text" json:"state"`
	CountDate      *time.Time `gorm:"type:date" json:"count_date"`
}

type StockCountLine struct {
	model.Base
	StockCountID   uint64  `gorm:"not null" json:"stock_count_id"`
	ItemID         uint64  `gorm:"not null" json:"item_id"`
	BatchID        *uint64 `json:"batch_id"`
	TheoreticalQty float64 `gorm:"type:numeric(18,4)" json:"theoretical_qty"`
	CountedQty     float64 `gorm:"type:numeric(18,4)" json:"counted_qty"`
	DiffQty        float64 `gorm:"type:numeric(18,4)" json:"diff_qty"`
}

const (
	TransferStateDraft     = "draft"
	TransferStateSent      = "sent"
	TransferStateInTransit = "in_transit"
	TransferStateReceived  = "received"
	TransferStateCancelled = "cancelled"
)

type WarehouseTransfer struct {
	model.Base
	OrganizationID      *uint64    `json:"organization_id"`
	Name                *string    `json:"name"`
	SrcWarehouseID      uint64     `gorm:"not null" json:"src_warehouse_id"`
	DstWarehouseID      uint64     `gorm:"not null" json:"dst_warehouse_id"`
	State               string     `gorm:"type:text" json:"state"`
	OutShipmentID       *uint64    `json:"out_shipment_id"`
	InShipmentID        *uint64    `json:"in_shipment_id"`
	IsInterorganization bool       `gorm:"default:false" json:"is_interorganization"`
	ScheduledDate       *time.Time `gorm:"type:date" json:"scheduled_date"`
}

const (
	InboundCostStateDraft     = "draft"
	InboundCostStatePosted    = "posted"
	InboundCostStateCancelled = "cancelled"

	SplitMethodQuantity = "by_quantity"
	SplitMethodWeight   = "by_weight"
	SplitMethodVolume   = "by_volume"
	SplitMethodValue    = "by_value"
	SplitMethodEqual    = "equal"
)

type InboundCost struct {
	model.Base
	OrganizationID    *uint64           `json:"organization_id"`
	Name              string            `json:"name"`
	Date              *time.Time        `gorm:"type:date" json:"date"`
	State             string            `json:"state"`
	TargetShipmentIDs helper.Int64Array `gorm:"type:bigint[]" json:"target_shipment_ids"`
	MovementID        *uint64           `json:"movement_id"`
}

func (InboundCost) TableName() string { return "inbound_costs" }

type InboundCostLine struct {
	model.Base
	InboundCostID      uint64  `json:"inbound_cost_id"`
	ItemID             uint64  `json:"item_id"`
	Description        *string `json:"description"`
	Amount             float64 `gorm:"type:numeric(18,4)" json:"amount"`
	SupplierBillLineID *uint64 `json:"supplier_bill_line_id"`
	SplitMethod        string  `json:"split_method"`
	AccountID          *uint64 `json:"account_id"`
}

func (InboundCostLine) TableName() string { return "inbound_cost_lines" }

type InboundCostAdjustment struct {
	model.Base
	InboundCostID   uint64  `json:"inbound_cost_id"`
	StockMovementID uint64  `json:"stock_movement_id"`
	ItemID          uint64  `json:"item_id"`
	AdditionalCost  float64 `gorm:"type:numeric(18,4)" json:"additional_cost"`
	CostLayerID     uint64  `json:"cost_layer_id"`
}

func (InboundCostAdjustment) TableName() string { return "inbound_cost_adjustments" }
