package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	DBHost        string
	DBPort        int
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	JWTSecret     string
	JWTIssuer     string
	JWTAudience   string
	Port          string
	MailgunAPIKey string
	MailgunDomain string
	MailgunSender string
}

// Global config instance
var AppConfig Config

// LoadEnv loads environment variables from .env file
func LoadEnv() error {
	// Try to load .env file, but don't fail if it doesn't exist
	_ = godotenv.Load()

	// Set default values
	AppConfig = Config{
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnvAsInt("DB_PORT", 5432),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", ""),
		DBName:        getEnv("DB_NAME", "devhive"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		JWTSecret:     getEnv("Jwt__Key", getEnv("JwtKey", getEnv("JWT_SECRET", "your-secret-key"))),
		JWTIssuer:     getEnv("Jwt__Issuer", getEnv("JwtIssuer", "devhive-backend")),
		JWTAudience:   getEnv("Jwt__Audience", getEnv("JwtAudience", "devhive-users")),
		Port:          getEnv("PORT", "8080"),
		MailgunAPIKey: getEnv("Mailgun__ApiKey", getEnv("MailgunApiKey", "")),
		MailgunDomain: getEnv("Mailgun__Domain", getEnv("MailgunDomain", "")),
		MailgunSender: getEnv("Mailgun__SenderEmail", ""),
	}

	return nil
}

// GetDatabaseURL returns the PostgreSQL connection string
func GetDatabaseURL() string {
	return "host=" + AppConfig.DBHost +
		" port=" + strconv.Itoa(AppConfig.DBPort) +
		" user=" + AppConfig.DBUser +
		" password=" + AppConfig.DBPassword +
		" dbname=" + AppConfig.DBName +
		" sslmode=" + AppConfig.DBSSLMode
}

// Helper function to get environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as int with default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
