package asset

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	MethodLinear              = "linear"
	MethodDeclining           = "declining"
	MethodDecliningThenLinear = "declining_then_linear"

	AssetStateDraft    = "draft"
	AssetStateRunning  = "running"
	AssetStateDisposed = "disposed"
	AssetStateSold     = "sold"

	OriginDepreciation = "depreciation"
	OriginDisposal     = "asset_disposal"
)

type FixedAsset struct {
	model.Base
	OrganizationID  uint64     `json:"organization_id"`
	Name            string     `json:"name"`
	CategoryID      uint64     `json:"category_id"`
	PurchaseValue   float64    `gorm:"type:numeric(18,4)" json:"purchase_value"`
	SalvageValue    float64    `gorm:"type:numeric(18,4);default:0" json:"salvage_value"`
	AcquisitionDate *time.Time `gorm:"type:date" json:"acquisition_date"`
	InServiceDate   *time.Time `gorm:"type:date" json:"in_service_date"`
	OriginalEntryID *uint64    `json:"original_entry_id"`
	InvoiceLineID   *uint64    `json:"invoice_line_id"`
	State           string     `json:"state"`
	DisposalDate    *time.Time `gorm:"type:date" json:"disposal_date"`
}

func (FixedAsset) TableName() string {
	return "fixed_assets"
}

type AssetDepreciationLine struct {
	model.Base
	AssetID          uint64    `json:"asset_id"`
	Sequence         int       `json:"sequence"`
	DepreciationDate time.Time `gorm:"type:date" json:"depreciation_date"`
	Amount           float64   `gorm:"type:numeric(18,4)" json:"amount"`
	Accumulated      float64   `gorm:"type:numeric(18,4)" json:"accumulated"`
	RemainingValue   float64   `gorm:"type:numeric(18,4)" json:"remaining_value"`
	EntryID          *uint64   `json:"entry_id"`
	Posted           bool      `gorm:"default:false" json:"posted"`
}

func (AssetDepreciationLine) TableName() string {
	return "asset_depreciation_lines"
}
