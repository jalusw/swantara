package tasks

import (
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	maxRetries  = 5
	taskTimeout = 30 * time.Second
)

func newTask(typ string, payload any) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(typ, data, asynq.MaxRetry(maxRetries), asynq.Timeout(taskTimeout)), nil
}
