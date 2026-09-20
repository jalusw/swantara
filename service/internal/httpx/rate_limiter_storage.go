package httpx

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jalusw/swantara/apps/service/internal/config"
	"github.com/redis/go-redis/v9"
)

type redisRateLimiterStorage struct {
	client *redis.Client
}

func NewRateLimiterStorage(cfg *config.Config) (fiber.Storage, error) {
	if cfg.RateLimiterStorage != "redis" {
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RateLimiterRedisAddress(),
		Password: cfg.RateLimiterRedisPassword,
		DB:       cfg.RateLimiterRedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &redisRateLimiterStorage{client: client}, nil
}

func (s *redisRateLimiterStorage) Get(key string) ([]byte, error) {
	return s.GetWithContext(context.Background(), key)
}

func (s *redisRateLimiterStorage) GetWithContext(ctx context.Context, key string) ([]byte, error) {
	val, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (s *redisRateLimiterStorage) Set(key string, val []byte, exp time.Duration) error {
	return s.SetWithContext(context.Background(), key, val, exp)
}

func (s *redisRateLimiterStorage) SetWithContext(ctx context.Context, key string, val []byte, exp time.Duration) error {
	return s.client.Set(ctx, key, val, exp).Err()
}

func (s *redisRateLimiterStorage) Delete(key string) error {
	return s.DeleteWithContext(context.Background(), key)
}

func (s *redisRateLimiterStorage) DeleteWithContext(ctx context.Context, key string) error {
	return s.client.Del(ctx, key).Err()
}

func (s *redisRateLimiterStorage) Reset() error {
	return s.ResetWithContext(context.Background())
}

func (s *redisRateLimiterStorage) ResetWithContext(ctx context.Context) error {
	return s.client.FlushDB(ctx).Err()
}

func (s *redisRateLimiterStorage) Close() error {
	return s.client.Close()
}
