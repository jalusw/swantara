package interorganization

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	TransactionStateDone      = "done"
	TransactionStateCancelled = "cancelled"

	RunStateDraft = "draft"
	RunStateDone  = "done"
)

type DropshipLink struct {
	model.Base
	SaleOrderLineID     uint64  `json:"sale_order_line_id"`
	PurchaseOrderLineID uint64  `json:"purchase_order_line_id"`
	StockMovementID     *uint64 `json:"stock_movement_id"`
}

func (DropshipLink) TableName() string {
	return "dropship_links"
}

type InterorganizationRule struct {
	model.Base
	FromOrganizationID *uint64 `json:"from_organization_id"`
	ToOrganizationID   *uint64 `json:"to_organization_id"`
	AutoMirror         bool    `gorm:"default:true" json:"auto_mirror"`
	SupplierContactID  *uint64 `json:"supplier_contact_id"`
	CustomerContactID  *uint64 `json:"customer_contact_id"`
}

func (InterorganizationRule) TableName() string {
	return "interorganization_rules"
}

type InterorganizationTransaction struct {
	model.Base
	SourceOrganizationID *uint64 `json:"source_organization_id"`
	SourceType           string  `gorm:"type:text" json:"source_type"`
	SourceID             *uint64 `json:"source_id"`
	MirrorOrganizationID *uint64 `json:"mirror_organization_id"`
	MirrorType           string  `gorm:"type:text" json:"mirror_type"`
	MirrorID             *uint64 `json:"mirror_id"`
	Amount               float64 `gorm:"type:numeric(18,4)" json:"amount"`
	State                string  `gorm:"type:text" json:"state"`
}

func (InterorganizationTransaction) TableName() string {
	return "interorganization_transactions"
}

type ConsolidationRun struct {
	model.Base
	GroupOrganizationID *uint64 `json:"group_organization_id"`
	PeriodID            *uint64 `json:"period_id"`
	ReportingCurrency   string  `gorm:"type:char(3)" json:"reporting_currency"`
	State               string  `gorm:"type:text" json:"state"`
}

func (ConsolidationRun) TableName() string {
	return "consolidation_runs"
}

type ConsolidationElimination struct {
	model.Base
	ConsolidationRunID         uint64  `json:"consolidation_run_id"`
	AccountID                  uint64  `json:"account_id"`
	CounterpartyOrganizationID *uint64 `json:"counterparty_organization_id"`
	Amount                     float64 `gorm:"type:numeric(18,4)" json:"amount"`
	Description                string  `json:"description"`
}

func (ConsolidationElimination) TableName() string {
	return "consolidation_eliminations"
}
