package ratelimiter

import (
	"context"
	"fmt"
	"time"
)

// RateLimiter handles rate limiting logic
type RateLimiter struct {
	storage       Storage
	defaultIPLimit    int
	defaultTokenLimit int
	blockDuration     time.Duration
	tokenLimits       map[string]int
}

// LimitResult represents the result of a rate limit check
type LimitResult struct {
	Allowed   bool
	Remaining int
	ResetTime time.Time
	Blocked   bool
}

// Config represents rate limiter configuration
type Config struct {
	DefaultIPLimit    int
	DefaultTokenLimit int
	BlockDuration     time.Duration
	TokenLimits       map[string]int
}

// New creates a new RateLimiter instance
func New(storage Storage, config Config) *RateLimiter {
	return &RateLimiter{
		storage:           storage,
		defaultIPLimit:    config.DefaultIPLimit,
		defaultTokenLimit: config.DefaultTokenLimit,
		blockDuration:     config.BlockDuration,
		tokenLimits:       config.TokenLimits,
	}
}

// CheckLimit checks if a request should be allowed based on IP or token
func (rl *RateLimiter) CheckLimit(ctx context.Context, ip, token string) (*LimitResult, error) {
	// Determine which key and limit to use
	key, limit := rl.getKeyAndLimit(ip, token)
	
	// Check if the key is currently blocked
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to check block status: %w", err)
	}
	
	if blocked {
		return &LimitResult{
			Allowed:   false,
			Remaining: 0,
			ResetTime: time.Now().Add(rl.blockDuration),
			Blocked:   true,
		}, nil
	}
	
	// Increment the request count
	count, err := rl.storage.Increment(ctx, key, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to increment counter: %w", err)
	}
	
	// Check if limit is exceeded
	if count > int64(limit) {
		// Block the key
		if err := rl.storage.SetBlock(ctx, key, rl.blockDuration); err != nil {
			return nil, fmt.Errorf("failed to set block: %w", err)
		}
		
		return &LimitResult{
			Allowed:   false,
			Remaining: 0,
			ResetTime: time.Now().Add(rl.blockDuration),
			Blocked:   true,
		}, nil
	}
	
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}
	
	return &LimitResult{
		Allowed:   true,
		Remaining: remaining,
		ResetTime: time.Now().Add(time.Second),
		Blocked:   false,
	}, nil
}

// getKeyAndLimit determines which key and limit to use
// Token limits have priority over IP limits
func (rl *RateLimiter) getKeyAndLimit(ip, token string) (string, int) {
	if token != "" {
		// Check if there's a specific limit for this token
		if limit, exists := rl.tokenLimits[token]; exists {
			return fmt.Sprintf("token:%s", token), limit
		}
		// Use default token limit
		return fmt.Sprintf("token:%s", token), rl.defaultTokenLimit
	}
	
	// Use IP-based limiting
	return fmt.Sprintf("ip:%s", ip), rl.defaultIPLimit
}

// Close closes the rate limiter and its storage
func (rl *RateLimiter) Close() error {
	return rl.storage.Close()
} 