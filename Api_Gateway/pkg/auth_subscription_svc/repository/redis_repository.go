package repository

import (
	"context"
	"time"

	interfacesrepository "github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/repository/interfacesRepository"
	"github.com/redis/go-redis/v9"
)

type redisRepository struct {
	rdb *redis.Client
}

func NewRedisRepository(rdb *redis.Client) interfacesrepository.RedisRepository {
	return &redisRepository{
		rdb: rdb,
	}
}
func (r *redisRepository) BlacklistToken(jti string, exp time.Time) error {
	// If token already expired, no need to blacklist
	ttl := time.Until(exp)
	if ttl <= 0 {
		return nil
	}

	key := "gateway:blacklist:jti:" + jti

	// Value is irrelevant — existence matters
	return r.rdb.Set(
		context.Background(),
		key,
		"1",
		ttl,
	).Err()
}
func (r *redisRepository) IsTokenBlacklisted(jti string) (bool, error) {
	key := "gateway:blacklist:jti:" + jti

	exists, err := r.rdb.Exists(context.Background(), key).Result()
	if err != nil {
		return false, err
	}

	return exists == 1, nil
}
