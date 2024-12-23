package main

import (
	"github.com/go-redis/redis/v8"
)

func InitRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	return client
}

var redisClient = InitRedis()
