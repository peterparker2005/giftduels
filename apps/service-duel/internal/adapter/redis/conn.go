package redis

import (
	"github.com/peterparker2005/giftduels/apps/service-duel/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		Username: cfg.Redis.Username,
	})
}
