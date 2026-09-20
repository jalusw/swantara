package jobs

import (
	"time"

	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

const (
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"

	TypeSubscriptionBilling   = "subscription-billing"
	TypeDeferralRecognition   = "deferral-recognition"
	TypeIntegrationEventSweep = "integration-event-sweep"
)

type JobRun struct {
	model.Base
	JobType      string     `gorm:"type:text" json:"job_type"`
	Status       string     `gorm:"type:text" json:"status"`
	Processed    int        `json:"processed"`
	Cursor       *uint64    `json:"cursor"`
	ErrorMessage *string    `json:"error_message"`
	StartedAt    time.Time  `gorm:"type:timestamptz" json:"started_at"`
	FinishedAt   *time.Time `gorm:"type:timestamptz" json:"finished_at"`
}

func (JobRun) TableName() string {
	return "job_runs"
}
