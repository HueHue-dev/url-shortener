package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type Store struct {
	redisClient *redis.Client
}

func NewRedisClient() *redis.Client {
	fmt.Println("Connecting to redis server on:", os.Getenv("REDIS_HOST"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}

	fmt.Println("Connected to redis server on:", os.Getenv("REDIS_HOST"))

	return rdb
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{redisClient: rdb}
}

func (s *Store) SetKey(ctx context.Context, key string, value string, expiration time.Duration) {
	fmt.Println("Setting key", key, "to", value, "in Redis")
	s.redisClient.Set(ctx, key, value, expiration)
	fmt.Println("The key", key, "has been set to", value, " successfully")
}

func (s *Store) GetLongURL(ctx context.Context, shortURL string) (string, error) {
	longURL, err := s.redisClient.Get(ctx, shortURL).Result()

	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("short URL not found")
	} else if err != nil {
		return "", fmt.Errorf("failed to retrieve from Redis: %v", err)
	}

	return longURL, nil
}

func (s *Store) IncrementMetric(ctx context.Context, shortURL string) {
	key := "metrics:" + shortURL
	s.redisClient.Incr(ctx, key)
}

func (s *Store) GetMetric(ctx context.Context, shortURL string) (int64, error) {
	key := "metrics:" + shortURL
	val, err := s.redisClient.Get(ctx, key).Int64()

	if errors.Is(err, redis.Nil) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return val, nil
}
