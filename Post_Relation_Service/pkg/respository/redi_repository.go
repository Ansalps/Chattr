package repository

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/respository/interfacesRepository"
	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(rdb *redis.Client) interfacesRepository.RedisRepository {
	return &redisRepository{
		rdb: rdb,
	}
}
func (r *redisRepository) Incr(ctx context.Context, key string) (string, error) {
	newVal, err := r.rdb.Incr(context.Background(), key).Result()
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(newVal, 10), nil
}

func (r *redisRepository) Pipeline() redis.Pipeliner {
	return r.rdb.Pipeline()
}

func (r *redisRepository) CacheGet(ctx context.Context, cacheKey string) (string, error) {
	cachedData, err := r.rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		return "", err
	}
	return cachedData, nil
}

func (r *redisRepository) CacheSet(ctx context.Context, cacheKey string, dataToCache []byte, expiration time.Duration) error {
	// r.Client is your *redis.Client (go-redis)
	return r.rdb.Set(ctx, cacheKey, dataToCache, expiration).Err()
}
func (r *redisRepository) PullNormalUserPostIDsFromRedis(ctx context.Context, userID uint64, lastID uint64, limit int) ([]uint64, error) {
    // The key is specific to the user viewing the feed
    key := fmt.Sprintf("feed:user:%d", userID)

    // Define pagination range
    // We want posts with scores (timestamps) lower than the last post seen
    maxScore := "+inf"
    if lastID > 0 {
        // Fetch the score of the lastID to use as the pivot for the next page
        score, err := r.rdb.ZScore(ctx, key, strconv.FormatUint(lastID, 10)).Result()
        if err == nil {
            maxScore = fmt.Sprintf("(%f", score)
        }
    }

    // Fetch from the Sorted Set in descending order (newest first)
    // ZRevRangeByScore gives us the high scores first
    ids, err := r.rdb.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
        Max:    maxScore,
        Min:    "-inf",
        Offset: 0,
        Count:  int64(limit),
    }).Result()

    if err != nil {
        if err == redis.Nil {
            return []uint64{}, nil
        }
        return nil, fmt.Errorf("failed to fetch normal feed from redis: %v", err)
    }

    // Convert string slice to uint64 slice
    resultIDs := make([]uint64, 0, len(ids))
    for _, idStr := range ids {
        id, err := strconv.ParseUint(idStr, 10, 64)
        if err == nil {
            resultIDs = append(resultIDs, id)
        }
    }

    return resultIDs, nil
}
func (r *redisRepository) PullCelebPostIDsFromRedis(ctx context.Context, celebIDs []uint64, lastID uint64, limit int) ([]uint64, error) {
	pipe := r.rdb.Pipeline()

	// Define the range for pagination
	maxScore := "+inf"
	if lastID > 0 {
		// Use '(' to make it "less than" rather than "less than or equal to"
		maxScore = fmt.Sprintf("(%d", lastID)
	}

	// 1. Queue up a request for every celebrity followed
	for _, cID := range celebIDs {
		key := fmt.Sprintf("celeb:posts:%d", cID)
		pipe.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
			Max:    maxScore,
			Min:    "-inf",
			Offset: 0,
			Count:  int64(limit),
		})
	}

	// 2. Execute all queries in one network trip
	cmds, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// 3. Collect and deduplicate IDs
	// Using a map to ensure we don't have duplicate post IDs if logic overlaps
	uniqueIDs := make(map[uint64]struct{})
	var resultIDs []uint64

	for _, cmd := range cmds {
		if zCmd, ok := cmd.(*redis.StringSliceCmd); ok {
			ids, _ := zCmd.Result()
			for _, idStr := range ids {
				id, _ := strconv.ParseUint(idStr, 10, 64)
				if _, exists := uniqueIDs[id]; !exists {
					uniqueIDs[id] = struct{}{}
					resultIDs = append(resultIDs, id)
				}
			}
		}
	}
	sort.Slice(resultIDs, func(i, j int) bool {
		return resultIDs[i] > resultIDs[j] // Sort descending (newest first)
	})

	// After sorting, trim to the limit again
	if len(resultIDs) > limit {
		resultIDs = resultIDs[:limit]
	}
	return resultIDs, nil
}
func (r *redisRepository) ExtendTTL(ctx context.Context, key string, ttl time.Duration) error {
	return r.rdb.Expire(ctx, key, ttl).Err()
}

// func (r *redisRepository) ZAdd(ctx context.Context, key string, score float64, member interface{}) error {
// 	return r.rdb.ZAdd(ctx, key, redis.Z{Score: score, Member: member}).Err()
// }

func (r *redisRepository) ZRem(ctx context.Context, key string, member interface{}) error {
	return r.rdb.ZRem(ctx, key, member).Err()
}

// You will need this for the "Pull" part of the Newsfeed
func (r *redisRepository) ZRevRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return r.rdb.ZRevRangeByScore(ctx, key, opt).Result()
}
