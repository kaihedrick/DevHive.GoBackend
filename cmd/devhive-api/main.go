package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devhive-backend/db"
	"devhive-backend/internal/config"
	"devhive-backend/internal/http/router"
	"devhive-backend/internal/repo"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Setup structured logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting DevHive API server on port %s", cfg.Port)

	// Connect to database
	database, err := initDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Run database migrations
	if err := db.RunMigrations(database); err != nil {
		log.Printf("Warning: Migration failed: %v", err)
	}

	// Create indexes for performance
	if err := db.CreateIndexes(database); err != nil {
		log.Printf("Warning: Index creation failed: %v", err)
	}

	// Create pgx pool for sqlc
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to create pgx pool:", err)
	}
	defer pool.Close()

	// Initialize repository
	queries := repo.New(pool)

	// Setup router
	r := router.Setup(cfg, queries, database)

	// Start server with graceful shutdown
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handler
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-shutdown
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown failed: %v", err)
		}
	}()

	log.Printf("Server started successfully on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
}

// initDB initializes database connection with proper error handling
func initDB(databaseURL string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	database := stdlib.OpenDB(*config)
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	return database, nil
}
