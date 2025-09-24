package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

// maskPassword masks the password in connection strings for logging
func maskPassword(connStr string) string {
	if connStr == "" {
		return ""
	}
	// Simple masking - replace password=xxx with password=***
	if strings.Contains(connStr, "password=") {
		parts := strings.Split(connStr, "password=")
		if len(parts) > 1 {
			subParts := strings.Split(parts[1], " ")
			if len(subParts) > 0 {
				parts[1] = "*** " + strings.Join(subParts[1:], " ")
			}
		}
		return strings.Join(parts, "password=")
	}
	return connStr
}

// InitDB initializes the PostgreSQL database connection using pgx
func InitDB() (*sql.DB, error) {
	// First try to use the Fly.io DATABASE_URL (standard format)
	dbURL := os.Getenv("DATABASE_URL")
	log.Printf("DEBUG: DATABASE_URL from env: %s", maskPassword(dbURL))

	// Fallback to Fly.io connection string (with double underscores)
	if dbURL == "" {
		dbURL = os.Getenv("ConnectionStrings__DbConnection")
		log.Printf("DEBUG: ConnectionStrings__DbConnection from env: %s", maskPassword(dbURL))
	}

	// If still empty, try individual environment variables
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		if host != "" && user != "" && dbname != "" {
			if port == "" {
				port = "5432"
			}
			if sslmode == "" {
				sslmode = "require"
			}
			dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				user, password, host, port, dbname, sslmode)
		}
	}

	if dbURL == "" {
		return nil, fmt.Errorf("no database connection string found")
	}

	log.Printf("DEBUG: Using database URL: %s", maskPassword(dbURL))

	// Parse the connection string
	config, err := pgx.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Create database connection using stdlib
	database := stdlib.OpenDB(*config)
	
	// Test the connection
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection established successfully")
	return database, nil
}

// CreatePool creates a pgx pool for sqlc operations
func CreatePool(databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Configure pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 5 * time.Minute
	config.MaxConnIdleTime = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the pool
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping pool: %w", err)
	}

	log.Println("Database pool created successfully")
	return pool, nil
}
