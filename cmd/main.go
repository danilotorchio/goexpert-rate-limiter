package main

import (
	"log"
	"fmt"

	"github.com/danilotorchio/go-expert-rate-limiter/internal/config"
	"github.com/danilotorchio/go-expert-rate-limiter/internal/middleware"
	"github.com/danilotorchio/go-expert-rate-limiter/pkg/ratelimiter"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Redis storage
	storage, err := ratelimiter.NewRedisStorage(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Redis storage: %v", err)
	}
	defer storage.Close()

	// Initialize rate limiter
	limiterConfig := ratelimiter.Config{
		DefaultIPLimit:    cfg.RateLimit.DefaultIPLimit,
		DefaultTokenLimit: cfg.RateLimit.DefaultTokenLimit,
		BlockDuration:     cfg.RateLimit.BlockDuration,
		TokenLimits:       cfg.Tokens,
	}

	rateLimiter := ratelimiter.New(storage, limiterConfig)
	defer rateLimiter.Close()

	// Initialize Gin router
	router := gin.Default()

	// Apply rate limiter middleware
	router.Use(middleware.RateLimiterMiddleware(rateLimiter))

	// Add some example routes
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World! Rate limiter is working.",
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	router.POST("/api/data", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Data received successfully",
			"data":    "This is a protected endpoint",
		})
	})

	// Start server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	log.Printf("Rate limiter configuration:")
	log.Printf("- Default IP limit: %d req/s", cfg.RateLimit.DefaultIPLimit)
	log.Printf("- Default token limit: %d req/s", cfg.RateLimit.DefaultTokenLimit)
	log.Printf("- Block duration: %v", cfg.RateLimit.BlockDuration)
	
	if len(cfg.Tokens) > 0 {
		log.Printf("- Token-specific limits:")
		for token, limit := range cfg.Tokens {
			log.Printf("  - %s: %d req/s", token, limit)
		}
	}

	if err := router.Run(fmt.Sprintf(":%s", cfg.Server.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 