package returns

import (
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	TypeCustomerReturn = "customer_return"
	TypeVendorReturn   = "vendor_return"

	StateDraft     = "draft"
	StateConfirmed = "confirmed"
	StateReceived  = "received"
	StateRefunded  = "refunded"
	StateDone      = "done"
	StateCancelled = "cancelled"

	DispositionRestock = "restock"
	DispositionScrap   = "scrap"
	DispositionRepair  = "repair"
	DispositionReplace = "replace"
)

const SequenceRMACode = "rma"

type RMA struct {
	model.Base
	OrganizationID  *uint64 `json:"organization_id"`
	Name            *string `json:"name"`
	Type            string  `gorm:"type:text" json:"type"`
	ContactID       uint64  `gorm:"not null" json:"contact_id"`
	OriginOrderType *string `json:"origin_order_type"`
	OriginOrderID   *uint64 `json:"origin_order_id"`
	Reason          *string `json:"reason"`
	State           string  `gorm:"type:text" json:"state"`
}

type RMALine struct {
	model.Base
	RMAID           uint64  `gorm:"not null" json:"rma_id"`
	ItemID          uint64  `gorm:"not null" json:"item_id"`
	Qty             float64 `gorm:"type:numeric(18,4)" json:"qty"`
	BatchID         *uint64 `json:"batch_id"`
	Disposition     string  `gorm:"type:text" json:"disposition"`
	StockMovementID *uint64 `json:"stock_movement_id"`
	CreditNoteID    *uint64 `json:"credit_note_id"`
}
