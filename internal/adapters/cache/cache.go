package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	once     sync.Once
	instance *redis.Client
)

// GetRedisClient returns the singleton instance of the Redis client
func GetRedisClient() *redis.Client {
	once.Do(func() {
		instance = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
	})
	return instance
}

// Cache is an interface that defines caching operations
type Cache interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// redisCache implements the Cache interface using Redis
type redisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new instance of redisCache
func NewRedisCache() Cache {
	return &redisCache{
		client: GetRedisClient(),
	}
}

// Get retrieves data from Redis and deserializes it into the provided value
func (r *redisCache) Get(ctx context.Context, key string, value interface{}) error {
	cachedData, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Key does not exist
			return nil
		}
		return err
	}

	if cachedData == "" {
		return nil
	}

	return json.Unmarshal([]byte(cachedData), value)
}

// Set serializes the value and sets it in Redis
func (r *redisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, expiration).Err()
}

// Delete removes data from Redis
func (r *redisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
