package audit

import (
	"encoding/json"
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

type Action string

const (
	ActionInsert Action = "insert"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Log struct {
	model.Base
	EntityTable string          `gorm:"column:table_name;not null" json:"table_name"`
	RecordID    uint64          `json:"record_id"`
	Action      Action          `gorm:"type:text;not null" json:"action"`
	ChangedBy   uint64          `json:"changed_by"`
	ChangedAt   time.Time       `json:"changed_at"`
	Diff        json.RawMessage `gorm:"type:jsonb" json:"diff" swaggertype:"object"`
}

func (Log) TableName() string {
	return "audit_logs"
}
