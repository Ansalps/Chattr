package client

import (
	"context"
	"log"

	"github.com/Ansalps/Chattr_Post_Relation_Service/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	// 1. Use ParseURL to automatically handle the 'rediss://' protocol and TLS
    opt, err := redis.ParseURL(cfg.Redis.Address)
    if err != nil {
        log.Fatalf("Invalid Redis URL: %v", err)
    }

    rdb := redis.NewClient(opt)
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr:     cfg.Redis.Address, // or redis:6379 in docker
	// 	Password: "",                // no password by default
	// 	DB:       0,                 // default DB
	// })

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	return rdb
}
