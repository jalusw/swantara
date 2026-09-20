package jobs

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jalusw/swantara/apps/service/internal/kernel/model"
)

func JobRunFixture(opts ...func(*JobRun) *JobRun) *JobRun {
	j := &JobRun{
		Base:      model.Base{ID: uint64(gofakeit.Number(1, 10000))},
		JobType:   TypeSubscriptionBilling,
		Status:    StatusRunning,
		Processed: gofakeit.Number(0, 100),
		StartedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(j)
	}
	return j
}
