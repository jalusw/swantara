package queue

import (
	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/config"
)

func NewQueueClient(config *config.Config) *asynq.Client {
	return asynq.NewClient(redisClientOpt(config))
}
