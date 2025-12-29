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

	// New methods for Celebrity Hot Lists
    ZAdd(ctx context.Context, key string, score float64, member interface{}) error
    ZRem(ctx context.Context, key string, member interface{}) error
    ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error)

	PullCelebPostIDsFromRedis(ctx context.Context, celebIDs []uint64, lastID uint64, limit int) ([]uint64, error)
}
