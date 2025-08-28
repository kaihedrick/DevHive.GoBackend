package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Requests int           `json:"requests"`
	Period   time.Duration `json:"period"`
}

// Default rate limiting configurations
var (
	DefaultRateLimit = RateLimitConfig{
		Requests: 100,
		Period:   time.Minute,
	}

	StrictRateLimitConfig = RateLimitConfig{
		Requests: 20,
		Period:   time.Minute,
	}
)

// RateLimiter creates a rate limiting middleware
func RateLimiter(config RateLimitConfig) gin.HandlerFunc {

	// Create a new rate limiter
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: config.Period,
		Limit:  int64(config.Requests),
	}

	// Create the limiter instance
	limiterInstance := limiter.New(store, rate)

	// Return the gin middleware
	return func(c *gin.Context) {
		// Get client identifier (IP address or user ID)
		clientID := getClientIdentifier(c)

		// Create a custom handler that integrates with gin
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if rate limit is exceeded
			context, err := limiterInstance.Get(c.Request.Context(), clientID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
				c.Abort()
				return
			}

			// Check if limit is exceeded
			if context.Reached {
				c.Header("X-RateLimit-Limit", string(rune(context.Limit)))
				c.Header("X-RateLimit-Remaining", string(rune(context.Remaining)))
				c.Header("X-RateLimit-Reset", string(rune(context.Reset)))
				c.Header("Retry-After", string(rune(context.Reset)))

				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":     "Rate limit exceeded",
					"limit":     context.Limit,
					"remaining": context.Remaining,
					"reset":     context.Reset,
				})
				c.Abort()
				return
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", string(rune(context.Limit)))
			c.Header("X-RateLimit-Remaining", string(rune(context.Remaining)))
			c.Header("X-RateLimit-Reset", string(rune(context.Reset)))

			c.Next()
		})

		// Execute the handler directly since we're already in gin context
		handler(c.Writer, c.Request)
	}
}

// getClientIdentifier returns a unique identifier for the client
func getClientIdentifier(c *gin.Context) string {
	// Try to get user ID from context first (for authenticated users)
	if userID, exists := c.Get("userID"); exists {
		if id, ok := userID.(string); ok {
			return "user:" + id
		}
	}

	// Fallback to IP address
	clientIP := c.ClientIP()
	if clientIP == "" {
		clientIP = "unknown"
	}
	return "ip:" + clientIP
}

// RateLimitByUser creates a rate limiting middleware that limits by user ID
func RateLimitByUser(config RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Create a new rate limiter for this user
		store := memory.NewStore()
		rate := limiter.Rate{
			Period: config.Period,
			Limit:  int64(config.Requests),
		}

		limiterInstance := limiter.New(store, rate)
		clientID := "user:" + userID.(string)

		// Check rate limit
		context, err := limiterInstance.Get(c.Request.Context(), clientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit check failed"})
			c.Abort()
			return
		}

		// Check if limit is exceeded
		if context.Reached {
			c.Header("X-RateLimit-Limit", string(rune(context.Limit)))
			c.Header("X-RateLimit-Remaining", string(rune(context.Remaining)))
			c.Header("X-RateLimit-Reset", string(rune(context.Reset)))
			c.Header("Retry-After", string(rune(context.Reset)))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Rate limit exceeded",
				"limit":     context.Limit,
				"remaining": context.Remaining,
				"reset":     context.Reset,
			})
			c.Abort()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(context.Limit)))
		c.Header("X-RateLimit-Remaining", string(rune(context.Remaining)))
		c.Header("X-RateLimit-Reset", string(rune(context.Reset)))

		c.Next()
	}
}

// StrictRateLimit creates a strict rate limiting middleware for sensitive operations
func StrictRateLimit() gin.HandlerFunc {
	return RateLimiter(StrictRateLimitConfig)
}
