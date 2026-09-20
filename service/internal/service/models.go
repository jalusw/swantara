package service

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	OrderTypeRepair       = "repair"
	OrderTypeMaintenance  = "maintenance"
	OrderTypeInstallation = "installation"
	OrderTypeInspection   = "inspection"

	OrderStateNew        = "new"
	OrderStateScheduled  = "scheduled"
	OrderStateInProgress = "in_progress"
	OrderStateDone       = "done"
	OrderStateInvoiced   = "invoiced"
	OrderStateCancelled  = "cancelled"

	LineTypePart    = "part"
	LineTypeLabor   = "labor"
	LineTypeExpense = "expense"

	ContractStateDraft     = "draft"
	ContractStateActive    = "active"
	ContractStateCancelled = "cancelled"
)

type Equipment struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name"`
	ItemID         *uint64    `json:"item_id"`
	SerialBatchID  *uint64    `json:"serial_batch_id"`
	OwnerContactID *uint64    `json:"owner_contact_id"`
	FixedAssetID   *uint64    `json:"fixed_asset_id"`
	Location       string     `json:"location"`
	InstallDate    *time.Time `gorm:"type:date" json:"install_date"`
	WarrantyEnd    *time.Time `gorm:"type:date" json:"warranty_end"`
	Category       string     `json:"category"`
}

func (Equipment) TableName() string {
	return "equipments"
}

type ServiceContract struct {
	model.Base
	OrganizationID   *uint64    `json:"organization_id"`
	Name             string     `json:"name"`
	ContactID        *uint64    `json:"contact_id"`
	EquipmentID      *uint64    `json:"equipment_id"`
	SubscriptionID   *uint64    `json:"subscription_id"`
	Coverage         string     `json:"coverage"`
	SLAResponseHours *int       `json:"sla_response_hours"`
	DateStart        *time.Time `gorm:"type:date" json:"date_start"`
	DateEnd          *time.Time `gorm:"type:date" json:"date_end"`
	State            string     `gorm:"type:text" json:"state"`
}

func (ServiceContract) TableName() string {
	return "service_contracts"
}

type MaintenancePlan struct {
	model.Base
	EquipmentID  *uint64    `json:"equipment_id"`
	Name         string     `json:"name"`
	IntervalDays int        `json:"interval_days"`
	NextDue      *time.Time `gorm:"type:date" json:"next_due"`
	Active       bool       `gorm:"default:true" json:"active"`
}

func (MaintenancePlan) TableName() string {
	return "maintenance_plans"
}

type ServiceOrder struct {
	model.Base
	OrganizationID *uint64    `json:"organization_id"`
	Name           string     `json:"name"`
	ContactID      *uint64    `json:"contact_id"`
	EquipmentID    *uint64    `json:"equipment_id"`
	ContractID     *uint64    `json:"contract_id"`
	Type           string     `gorm:"type:text" json:"type"`
	Priority       int        `json:"priority"`
	State          string     `gorm:"type:text" json:"state"`
	ScheduledDate  *time.Time `json:"scheduled_date"`
	TechnicianID   *uint64    `json:"technician_id"`
	InvoiceID      *uint64    `json:"invoice_id"`
	DimensionID    *uint64    `json:"dimension_id"`
	ReportedIssue  string     `json:"reported_issue"`
	Resolution     string     `json:"resolution"`
}

func (ServiceOrder) TableName() string {
	return "service_orders"
}

type ServiceOrderLine struct {
	model.Base
	ServiceOrderID    uint64  `json:"service_order_id"`
	Type              string  `gorm:"type:text" json:"type"`
	ItemID            *uint64 `json:"item_id"`
	Description       string  `json:"description"`
	Qty               float64 `gorm:"type:numeric(18,4)" json:"qty"`
	UnitID            *uint64 `json:"unit_id"`
	UnitCost          float64 `gorm:"type:numeric(18,4)" json:"unit_cost"`
	UnitPrice         float64 `gorm:"type:numeric(18,4)" json:"unit_price"`
	StockMovementID   *uint64 `json:"stock_movement_id"`
	Billable          bool    `json:"billable"`
	CoveredByWarranty bool    `json:"covered_by_warranty"`
}

func (ServiceOrderLine) TableName() string {
	return "service_order_lines"
}
