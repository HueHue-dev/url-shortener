package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

func RedisClient() *redis.Client {
	fmt.Println("Connecting to redis server on:", os.Getenv("REDIS_HOST"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	fmt.Println("Connected to redis server on:", os.Getenv("REDIS_HOST"))

	return rdb
}

func SetKey(ctx *context.Context, rdb *redis.Client, key string, value string, expiration time.Duration) {
	fmt.Println("Setting key", key, "to", value, "in Redis")
	rdb.Set(*ctx, key, value, expiration)
	fmt.Println("The key", key, "has been set to", value, " successfully")
}

func GetLongURL(ctx *context.Context, rdb *redis.Client, shortURL string) (string, error) {
	longURL, err := rdb.Get(*ctx, shortURL).Result()

	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("short URL not found")
	} else if err != nil {
		return "", fmt.Errorf("failed to retrieve from Redis: %v", err)
	}

	return longURL, nil
}
