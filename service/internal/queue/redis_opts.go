package queue

import (
	"github.com/hibiken/asynq"
	"github.com/jalusw/swantara/apps/service/internal/config"
)

func redisClientOpt(cfg *config.Config) asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:     cfg.QueueRedisAddress(),
		Password: cfg.QueueRedisPassword,
		DB:       cfg.QueueRedisDB,
	}
}
