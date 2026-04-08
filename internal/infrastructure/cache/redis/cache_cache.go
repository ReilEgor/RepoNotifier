package redis

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	logger *slog.Logger
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{
		client: client,
		logger: slog.With(slog.String("component", "Cache")),
	}
}
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	if err := c.client.Get(ctx, key).Err(); err != nil {
		if err == redis.Nil {
			return "", nil
		}
		c.logger.Error("cache get error", "key", key, "error", err)
		return "", err
	}
	value, err := c.client.Get(ctx, key).Result()
	if err != nil {
		c.logger.Error("cache get error", "key", key, "error", err)
		return "", err
	}
	return value, nil
}
func (c *Cache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		c.logger.Error("cache set error", "key", key, "error", err)
		return err
	}
	return nil
}
