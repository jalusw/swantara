package queue

import (
	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/config"
)

type HealthChecker struct {
	inspector *asynq.Inspector
}

func NewHealthChecker(config *config.Config) *HealthChecker {
	return &HealthChecker{
		inspector: asynq.NewInspector(redisClientOpt(config)),
	}
}

func (h *HealthChecker) WorkerRunning() error {
	servers, err := h.inspector.Servers()
	if err != nil {
		return err
	}
	if len(servers) == 0 {
		return ErrNoQueueWorkerRunning
	}
	return nil
}
