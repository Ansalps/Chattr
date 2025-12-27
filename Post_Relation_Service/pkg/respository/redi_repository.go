package repository

import (
	"context"
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
func (r *redisRepository)CacheGet(ctx context.Context,cacheKey string)(string,error){
	cachedData, err := r.rdb.Get(ctx, cacheKey).Result()
	if err!=nil{
		return "",err
	}
	return cachedData,nil
}
func (r *redisRepository) CacheSet(ctx context.Context, cacheKey string, dataToCache []byte, expiration time.Duration) error {
	// r.Client is your *redis.Client (go-redis)
	return r.rdb.Set(ctx, cacheKey, dataToCache, expiration).Err()
}
