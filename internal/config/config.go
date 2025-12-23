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
	Port          string
	GRPCPort      string
	DatabaseURL   string
	JWT           JWTConfig
	CORS          CORSConfig
	Mail          MailConfig
	GoogleOAuth   GoogleOAuthConfig
	AdminPassword string
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	SigningKey                        string
	Issuer                            string
	Audience                          string
	Expiration                        time.Duration // Access token expiration (default: 15 minutes)
	RefreshTokenExpiration            time.Duration // Refresh token expiration (default: 7 days)
	RefreshTokenPersistentExpiration  time.Duration // Persistent refresh token expiration for "Remember Me" (default: 30 days)
	RefreshTokenSessionExpiration     time.Duration // Session refresh token expiration (default: 0 = browser session)
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

// GoogleOAuthConfig holds Google OAuth 2.0 configuration
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
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
			SigningKey:                       getEnv("JWT_SIGNING_KEY", "your-super-secret-jwt-key-change-in-production"),
			Issuer:                           getEnv("JWT_ISSUER", "https://api.devhive.it.com"),
			Audience:                         getEnv("JWT_AUDIENCE", "devhive-clients"),
			Expiration:                       time.Duration(getEnvAsInt("JWT_EXPIRATION_MINUTES", 15)) * time.Minute,                                    // Access token: 15 minutes
			RefreshTokenExpiration:           time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRATION_DAYS", 7)) * 24 * time.Hour,                             // Refresh token: 7 days (default/backward compat)
			RefreshTokenPersistentExpiration: time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRATION_PERSISTENT_DAYS", 30)) * 24 * time.Hour,                 // Persistent: 30 days
			RefreshTokenSessionExpiration:    time.Duration(getEnvAsInt("JWT_REFRESH_EXPIRATION_SESSION_HOURS", 0)) * time.Hour,                         // Session: 0 = browser session
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080", "https://d35scdhidypl44.cloudfront.net", "https://devhive.it.com"}),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
		AdminPassword: getEnv("ADMIN_CERTIFICATES_PASSWORD", "jtAppmine2021"),
		Mail: MailConfig{
			APIKey: getEnv("MAILGUN_API_KEY", ""),
			Domain: getEnv("MAILGUN_DOMAIN", ""),
			Sender: getEnv("MAILGUN_SENDER", ""),
		},
		GoogleOAuth: GoogleOAuthConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
				"openid",
			},
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
