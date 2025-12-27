package repository

import (
	"context"
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

func (r *redisRepository) Incr(ctx context.Context, key string) (string,error) {
	newVal,err:=r.rdb.Incr(context.Background(),key).Result()
	if err!=nil{
		return "",err
	}

    return strconv.FormatInt(newVal,10),nil
}
func(r *redisRepository)ExtendTTL(ctx context.Context,key string,ttl time.Duration)error{
	return r.rdb.Expire(ctx,key,ttl).Err()
}
