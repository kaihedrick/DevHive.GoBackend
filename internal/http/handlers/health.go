package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"devhive-backend/internal/http/response"
)

// HealthCheck handles basic health check
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "DevHive API",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// ReadinessCheck creates a readiness check handler that verifies database connectivity
func ReadinessCheck(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			response.JSON(w, http.StatusServiceUnavailable, map[string]interface{}{
				"status": "not_ready",
				"error":  "Database connection failed",
				"checks": map[string]string{
					"database": "failed",
				},
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

// LivenessCheck handles liveness probe
func LivenessCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "alive",
	})
}
