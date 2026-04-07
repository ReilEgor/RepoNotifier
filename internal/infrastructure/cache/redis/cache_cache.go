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
	panic("implement me")
}
func (c *Cache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	panic("implement me")
}
