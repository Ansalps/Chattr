package client

import (
	"context"
	"log"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:     cfg.Redis.Address, // or redis:6379 in docker
        Password: "",               // no password by default
        DB:       0,                // default DB
    })

    if err := rdb.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("Redis connection failed: %v", err)
    }

    return rdb
}


// func NewRedisClient(cfg *config.Config) *RedisClient {
// 	rdb := redis.NewClient(&redis.Options{
// 		Addr:     cfg.Redis.Address,
// 		Password: "",
// 		DB:       0,
// 	})
// 	if err := rdb.Ping(context.Background()).Err(); err != nil {
// 		log.Fatalf("Redis connection failed: %v", err)
// 	}

// 	return &RedisClient{
// 		Client: rdb,
// 	}
// }
