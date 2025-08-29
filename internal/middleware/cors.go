package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS middleware for Gin
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Allow specific origins or use wildcard for development
		allowedOrigins := []string{
			"http://localhost:3000",   // React dev server
			"http://localhost:8080",   // Alternative dev port
			"https://devhive.app",     // Production domain
			"https://www.devhive.app", // Production domain with www
		}

		// Check if origin is allowed
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		// Set Access-Control-Allow-Origin header
		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// For development, allow all origins
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// Set other CORS headers
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-API-Key")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
