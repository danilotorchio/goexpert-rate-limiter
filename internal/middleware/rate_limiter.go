package middleware

import (
	"net"
	"net/http"
	"strconv"

	"github.com/danilotorchio/go-expert-rate-limiter/pkg/ratelimiter"
	"github.com/gin-gonic/gin"
)

// RateLimiterMiddleware creates a Gin middleware for rate limiting
func RateLimiterMiddleware(limiter *ratelimiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract IP address
		ip := getClientIP(c)
		
		// Extract API key token from header
		token := c.GetHeader("API_KEY")
		
		// Check rate limit
		result, err := limiter.CheckLimit(c.Request.Context(), ip, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
			c.Abort()
			return
		}
		
		// Set rate limit headers
		c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
		c.Header("X-RateLimit-Reset", result.ResetTime.Format("2006-01-02T15:04:05Z"))
		
		if !result.Allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// getClientIP extracts the real client IP address
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ip := net.ParseIP(xff); ip != nil {
			return ip.String()
		}
	}
	
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}
	
	// Fall back to remote address
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	
	return ip
} 