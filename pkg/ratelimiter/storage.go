package ratelimiter

import (
	"context"
	"time"
)

// Storage defines the interface for storing rate limiter data
type Storage interface {
	// Increment increments the request count for the given key
	// Returns the current count and whether the key existed before
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)
	
	// Get retrieves the current count for the given key
	Get(ctx context.Context, key string) (int64, error)
	
	// SetBlock sets a block for the given key with the specified duration
	SetBlock(ctx context.Context, key string, duration time.Duration) error
	
	// IsBlocked checks if the given key is currently blocked
	IsBlocked(ctx context.Context, key string) (bool, error)
	
	// Close closes the storage connection
	Close() error
} 