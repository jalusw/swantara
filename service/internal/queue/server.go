package queue

import (
	"time"

	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/config"
)

func NewQueueServer(config *config.Config) *asynq.Server {
	server := asynq.NewServer(
		redisClientOpt(config),
		asynq.Config{
			Concurrency:     10,
			Logger:          NewAsynqLogger(),
			LogLevel:        asynq.InfoLevel,
			ShutdownTimeout: 10 * time.Second,
		},
	)
	return server
}
