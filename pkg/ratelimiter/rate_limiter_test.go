package ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage is a mock implementation of Storage interface for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	args := m.Called(ctx, key, window)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) Get(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	args := m.Called(ctx, key, duration)
	return args.Error(0)
}

func (m *MockStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestRateLimiter_CheckLimit_IPBasedAllowed(t *testing.T) {
	mockStorage := new(MockStorage)
	config := Config{
		DefaultIPLimit:    10,
		DefaultTokenLimit: 100,
		BlockDuration:     5 * time.Minute,
		TokenLimits:       make(map[string]int),
	}

	rl := New(mockStorage, config)
	ctx := context.Background()

	// Mock expectations
	mockStorage.On("IsBlocked", ctx, "ip:192.168.1.1").Return(false, nil)
	mockStorage.On("Increment", ctx, "ip:192.168.1.1", time.Second).Return(int64(5), nil)

	result, err := rl.CheckLimit(ctx, "192.168.1.1", "")

	assert.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, 5, result.Remaining)
	assert.False(t, result.Blocked)

	mockStorage.AssertExpectations(t)
}

func TestRateLimiter_CheckLimit_IPBasedExceeded(t *testing.T) {
	mockStorage := new(MockStorage)
	config := Config{
		DefaultIPLimit:    10,
		DefaultTokenLimit: 100,
		BlockDuration:     5 * time.Minute,
		TokenLimits:       make(map[string]int),
	}

	rl := New(mockStorage, config)
	ctx := context.Background()

	// Mock expectations
	mockStorage.On("IsBlocked", ctx, "ip:192.168.1.1").Return(false, nil)
	mockStorage.On("Increment", ctx, "ip:192.168.1.1", time.Second).Return(int64(11), nil)
	mockStorage.On("SetBlock", ctx, "ip:192.168.1.1", 5*time.Minute).Return(nil)

	result, err := rl.CheckLimit(ctx, "192.168.1.1", "")

	assert.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, 0, result.Remaining)
	assert.True(t, result.Blocked)

	mockStorage.AssertExpectations(t)
}

func TestRateLimiter_CheckLimit_TokenBasedAllowed(t *testing.T) {
	mockStorage := new(MockStorage)
	config := Config{
		DefaultIPLimit:    10,
		DefaultTokenLimit: 100,
		BlockDuration:     5 * time.Minute,
		TokenLimits:       map[string]int{"abc123": 50},
	}

	rl := New(mockStorage, config)
	ctx := context.Background()

	// Mock expectations
	mockStorage.On("IsBlocked", ctx, "token:abc123").Return(false, nil)
	mockStorage.On("Increment", ctx, "token:abc123", time.Second).Return(int64(25), nil)

	result, err := rl.CheckLimit(ctx, "192.168.1.1", "abc123")

	assert.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, 25, result.Remaining)
	assert.False(t, result.Blocked)

	mockStorage.AssertExpectations(t)
}

func TestRateLimiter_CheckLimit_TokenBasedExceeded(t *testing.T) {
	mockStorage := new(MockStorage)
	config := Config{
		DefaultIPLimit:    10,
		DefaultTokenLimit: 100,
		BlockDuration:     5 * time.Minute,
		TokenLimits:       map[string]int{"abc123": 50},
	}

	rl := New(mockStorage, config)
	ctx := context.Background()

	// Mock expectations
	mockStorage.On("IsBlocked", ctx, "token:abc123").Return(false, nil)
	mockStorage.On("Increment", ctx, "token:abc123", time.Second).Return(int64(51), nil)
	mockStorage.On("SetBlock", ctx, "token:abc123", 5*time.Minute).Return(nil)

	result, err := rl.CheckLimit(ctx, "192.168.1.1", "abc123")

	assert.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, 0, result.Remaining)
	assert.True(t, result.Blocked)

	mockStorage.AssertExpectations(t)
}

func TestRateLimiter_CheckLimit_AlreadyBlocked(t *testing.T) {
	mockStorage := new(MockStorage)
	config := Config{
		DefaultIPLimit:    10,
		DefaultTokenLimit: 100,
		BlockDuration:     5 * time.Minute,
		TokenLimits:       make(map[string]int),
	}

	rl := New(mockStorage, config)
	ctx := context.Background()

	// Mock expectations
	mockStorage.On("IsBlocked", ctx, "ip:192.168.1.1").Return(true, nil)

	result, err := rl.CheckLimit(ctx, "192.168.1.1", "")

	assert.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, 0, result.Remaining)
	assert.True(t, result.Blocked)

	mockStorage.AssertExpectations(t)
}

func TestRateLimiter_GetKeyAndLimit(t *testing.T) {
	config := Config{
		DefaultIPLimit:    10,
		DefaultTokenLimit: 100,
		BlockDuration:     5 * time.Minute,
		TokenLimits:       map[string]int{"abc123": 50},
	}

	rl := New(nil, config)

	// Test IP-based key and limit
	key, limit := rl.getKeyAndLimit("192.168.1.1", "")
	assert.Equal(t, "ip:192.168.1.1", key)
	assert.Equal(t, 10, limit)

	// Test token-based key and limit with specific limit
	key, limit = rl.getKeyAndLimit("192.168.1.1", "abc123")
	assert.Equal(t, "token:abc123", key)
	assert.Equal(t, 50, limit)

	// Test token-based key and limit with default limit
	key, limit = rl.getKeyAndLimit("192.168.1.1", "xyz789")
	assert.Equal(t, "token:xyz789", key)
	assert.Equal(t, 100, limit)
} 