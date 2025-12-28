package interfacesRepository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository interface {
	CacheGet(ctx context.Context, cacheKey string) (string, error)
	CacheSet(ctx context.Context, cacheKey string, dataToCache []byte, expiration time.Duration) error
	Incr(ctx context.Context, key string) (string, error)
	ExtendTTL(ctx context.Context, key string, ttl time.Duration) error
	Pipeline() redis.Pipeliner
}
