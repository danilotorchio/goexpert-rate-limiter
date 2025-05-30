package ratelimiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisStorage implements Storage interface using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(host, port, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{client: client}, nil
}

// Increment increments the request count for the given key
func (r *RedisStorage) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	
	countKey := fmt.Sprintf("rate_limit:%s", key)
	
	incr := pipe.Incr(ctx, countKey)
	pipe.Expire(ctx, countKey, window)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment rate limit: %w", err)
	}
	
	return incr.Val(), nil
}

// Get retrieves the current count for the given key
func (r *RedisStorage) Get(ctx context.Context, key string) (int64, error) {
	countKey := fmt.Sprintf("rate_limit:%s", key)
	
	val, err := r.client.Get(ctx, countKey).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get rate limit: %w", err)
	}
	
	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse rate limit count: %w", err)
	}
	
	return count, nil
}

// SetBlock sets a block for the given key with the specified duration
func (r *RedisStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	blockKey := fmt.Sprintf("blocked:%s", key)
	
	err := r.client.Set(ctx, blockKey, "1", duration).Err()
	if err != nil {
		return fmt.Errorf("failed to set block: %w", err)
	}
	
	return nil
}

// IsBlocked checks if the given key is currently blocked
func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := fmt.Sprintf("blocked:%s", key)
	
	exists, err := r.client.Exists(ctx, blockKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check block status: %w", err)
	}
	
	return exists > 0, nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
} 