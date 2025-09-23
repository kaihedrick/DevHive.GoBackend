package handlers

import (
	"context"
	"net/http"
	"time"

	"devhive-backend/internal/http/response"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// HealthCheck handles basic health check
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "DevHive API",
		"time":    time.Now().UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// ReadinessCheck creates a readiness check handler that verifies database connectivity
func ReadinessCheck(databaseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Test database connection
		config, err := pgx.ParseConfig(databaseURL)
		if err != nil {
			response.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "not_ready",
				"error":  "Failed to parse database URL",
			})
			return
		}

		db := stdlib.OpenDB(*config)
		defer db.Close()

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			response.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "not_ready",
				"error":  "Database connection failed",
			})
			return
		}

		response.JSON(w, http.StatusOK, map[string]interface{}{
			"status": "ready",
			"checks": map[string]string{
				"database": "ok",
			},
		})
	}
}
