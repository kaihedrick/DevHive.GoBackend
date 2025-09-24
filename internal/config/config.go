package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port        string
	GRPCPort    string
	DatabaseURL string
	JWT         JWTConfig
	CORS        CORSConfig
	Mail        MailConfig
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SigningKey string
	Issuer     string
	Audience   string
	Expiration time.Duration
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

// MailConfig holds mail service configuration
type MailConfig struct {
	APIKey string
	Domain string
	Sender string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		GRPCPort:    getEnv("GRPC_PORT", "8081"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://devhive:devhive@localhost:5432/devhive?sslmode=disable"),
		JWT: JWTConfig{
			SigningKey: getEnv("JWT_SIGNING_KEY", "your-super-secret-jwt-key-change-in-production"),
			Issuer:     getEnv("JWT_ISSUER", "https://api.devhive.it.com"),
			Audience:   getEnv("JWT_AUDIENCE", "devhive-clients"),
			Expiration: time.Duration(getEnvAsInt("JWT_EXPIRATION_HOURS", 24)) * time.Hour,
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080", "https://d35scdhidypl44.cloudfront.net", "https://devhive.it.com"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
		Mail: MailConfig{
			APIKey: getEnv("MAILGUN_API_KEY", ""),
			Domain: getEnv("MAILGUN_DOMAIN", ""),
			Sender: getEnv("MAILGUN_SENDER", ""),
		},
	}

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated values parsing
		// Split by comma and trim spaces
		values := strings.Split(value, ",")
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
		return values
	}
	return defaultValue
}
