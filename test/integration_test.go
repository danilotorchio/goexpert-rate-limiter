package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danilotorchio/go-expert-rate-limiter/internal/middleware"
	"github.com/danilotorchio/go-expert-rate-limiter/pkg/ratelimiter"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// InMemoryStorage is a simple in-memory implementation for testing
type InMemoryStorage struct {
	counts  map[string]int64
	blocked map[string]time.Time
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		counts:  make(map[string]int64),
		blocked: make(map[string]time.Time),
	}
}

func (i *InMemoryStorage) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	i.counts[key]++
	return i.counts[key], nil
}

func (i *InMemoryStorage) Get(ctx context.Context, key string) (int64, error) {
	return i.counts[key], nil
}

func (i *InMemoryStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	i.blocked[key] = time.Now().Add(duration)
	return nil
}

func (i *InMemoryStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	if blockedUntil, exists := i.blocked[key]; exists {
		if time.Now().Before(blockedUntil) {
			return true, nil
		}
		delete(i.blocked, key)
	}
	return false, nil
}

func (i *InMemoryStorage) Close() error {
	return nil
}

func setupTestRouter() (*gin.Engine, *InMemoryStorage) {
	gin.SetMode(gin.TestMode)
	
	storage := NewInMemoryStorage()
	config := ratelimiter.Config{
		DefaultIPLimit:    3, // Low limit for testing
		DefaultTokenLimit: 5,
		BlockDuration:     10 * time.Second,
		TokenLimits:       map[string]int{"test_token": 2},
	}
	
	rateLimiter := ratelimiter.New(storage, config)
	
	router := gin.New()
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	return router, storage
}

func TestIntegration_IPBasedRateLimit(t *testing.T) {
	router, _ := setupTestRouter()
	
	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	}
	
	// 4th request should be blocked
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "you have reached the maximum number of requests")
}

func TestIntegration_TokenBasedRateLimit(t *testing.T) {
	router, _ := setupTestRouter()
	
	// First 2 requests with token should be allowed (token limit is 2)
	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("API_KEY", "test_token")
		req.RemoteAddr = "192.168.1.1:12345"
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	}
	
	// 3rd request should be blocked
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("API_KEY", "test_token")
	req.RemoteAddr = "192.168.1.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestIntegration_TokenPriorityOverIP(t *testing.T) {
	router, storage := setupTestRouter()
	
	// Exhaust IP limit first
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
	}
	
	// IP should be blocked now
	req, _ := http.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	
	// But token should still work (priority over IP)
	req, _ = http.NewRequest("GET", "/test", nil)
	req.Header.Set("API_KEY", "test_token")
	req.RemoteAddr = "192.168.1.1:12345"
	
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Cleanup
	storage.Close()
}

func TestIntegration_DifferentIPs(t *testing.T) {
	router, _ := setupTestRouter()
	
	// Each IP should have its own limit
	ips := []string{"192.168.1.1:12345", "192.168.1.2:12345", "192.168.1.3:12345"}
	
	for _, ip := range ips {
		// Each IP can make 3 requests
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.RemoteAddr = ip
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusOK, w.Code)
		}
		
		// 4th request should be blocked for each IP
		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = ip
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	}
} 