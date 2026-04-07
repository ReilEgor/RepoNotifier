package service

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCacheMiss = errors.New("key not found in cache")
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}
