package quality

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	CheckResultPending = "pending"
	CheckResultPass    = "pass"
	CheckResultFail    = "fail"

	AlertStateOpen       = "open"
	AlertStateInProgress = "in_progress"
	AlertStateSolved     = "solved"
	AlertStateCancelled  = "cancelled"

	TestTypePassFail    = "pass_fail"
	TestTypeMeasure     = "measure"
	TestTypeInstruction = "instruction"
)

type QualityCheck struct {
	model.Base
	PointID           *uint64    `json:"point_id"`
	ItemID            *uint64    `json:"item_id"`
	BatchID           *uint64    `json:"batch_id"`
	ShipmentID        *uint64    `json:"shipment_id"`
	ProductionOrderID *uint64    `json:"production_order_id"`
	MeasuredValue     *float64   `gorm:"type:numeric(18,4)" json:"measured_value"`
	Result            string     `gorm:"type:text" json:"result"`
	CheckedBy         *uint64    `json:"checked_by"`
	CheckedAt         *time.Time `json:"checked_at"`
}

func (QualityCheck) TableName() string {
	return "quality_checks"
}

type QualityAlert struct {
	model.Base
	ItemID      *uint64 `json:"item_id"`
	BatchID     *uint64 `json:"batch_id"`
	CheckID     *uint64 `json:"check_id"`
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Severity    *string `json:"severity"`
	State       string  `gorm:"type:text" json:"state"`
	AssignedTo  *uint64 `json:"assigned_to"`
}

func (QualityAlert) TableName() string {
	return "quality_alerts"
}
